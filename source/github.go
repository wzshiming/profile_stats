package source

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	ghv3 "github.com/google/go-github/v33/github"
	ghv4 "github.com/shurcooL/githubv4"
	"github.com/wzshiming/httpcache"
	"golang.org/x/oauth2"
)

const MaxPageSize = 100

type Stat struct {
	Name          string
	Stars         int
	Forks         int
	Issues        int
	Commits       int
	Reviews       int
	PullRequests  int
	ContributedTo int
}

type OrgStat struct {
	Name         string
	LogoURL      string
	Issues       int
	Commits      int
	Reviews      int
	PullRequests int
}

type intervalRequest struct {
	interval      time.Duration
	last          time.Time
	roundTripperr http.RoundTripper
}

func newIntervalRequest(roundTripperr http.RoundTripper, interval time.Duration) http.RoundTripper {
	if interval <= 0 {
		return roundTripperr
	}
	return &intervalRequest{
		roundTripperr: roundTripperr,
		interval:      interval,
	}
}

func (l *intervalRequest) RoundTrip(r *http.Request) (*http.Response, error) {
	defer func() {
		l.last = time.Now()
	}()
	now := time.Now()
	sub := now.Sub(l.last)
	if s := l.interval - sub; s > 0 {
		time.Sleep(s)
	}
	return l.roundTripperr.RoundTrip(r)
}

func NewSource(token string, cache string, interval time.Duration) *Source {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	ctx := context.Background()
	transport := oauth2.NewClient(ctx, src).Transport

	opts := []httpcache.Option{}
	if cache != "" {
		opts = append(opts, httpcache.WithStorer(httpcache.DirectoryStorer(cache)))
	}

	transport = newIntervalRequest(transport, interval)
	return &Source{
		cliv3: ghv3.NewClient(&http.Client{
			Transport: httpcache.NewRoundTripper(transport,
				append([]httpcache.Option{
					httpcache.WithFilterer(
						httpcache.MethodFilterer(http.MethodGet),
					),
					httpcache.WithKeyer(
						httpcache.JointKeyer(
							httpcache.MethodKeyer(),
							httpcache.PathKeyer(),
						),
					),
				}, opts...)...,
			),
		}),
		cliv4: ghv4.NewClient(&http.Client{
			Transport: httpcache.NewRoundTripper(transport,
				append([]httpcache.Option{
					httpcache.WithFilterer(
						httpcache.MethodFilterer(http.MethodPost),
					),
					httpcache.WithKeyer(
						httpcache.JointKeyer(
							httpcache.MethodKeyer(),
							httpcache.PathKeyer(),
							httpcache.BodyKeyer(),
						),
					),
				}, opts...)...,
			),
		}),
	}
}

type Source struct {
	cliv3 *ghv3.Client
	cliv4 *ghv4.Client
}

func (s *Source) Stat(ctx context.Context, username string) (*Stat, error) {
	var query struct {
		User struct {
			Repositories struct {
				Nodes []struct {
					StargazerCount ghv4.Int
					ForkCount      ghv4.Int
					IsFork         ghv4.Boolean
				}
			} `graphql:"repositories(first: 100, ownerAffiliations: OWNER, orderBy: {direction: DESC, field: STARGAZERS})"`
			Contributions struct {
				TotalCommitContributions            ghv4.Int
				TotalPullRequestReviewContributions ghv4.Int
				TotalPullRequestContributions       ghv4.Int
				TotalIssueContributions             ghv4.Int
			} `graphql:"contributionsCollection(from: $from)"`
			ContributedTo struct {
				TotalCount ghv4.Int
			} `graphql:"repositoriesContributedTo(first: 0)"`
			Name ghv4.String
		} `graphql:"user(login: $username)"`
	}
	now := time.Now().UTC()
	variables := map[string]interface{}{
		"from":     ghv4.DateTime{now.AddDate(-1, 0, 0)},
		"username": ghv4.String(username),
	}

	err := s.cliv4.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}
	stat := Stat{}
	stat.Name = string(query.User.Name)
	for _, repo := range query.User.Repositories.Nodes {
		stat.Stars += int(repo.StargazerCount)
		if !repo.IsFork {
			stat.Forks += int(repo.ForkCount)
		}
	}

	stat.Commits = int(query.User.Contributions.TotalCommitContributions)
	stat.Reviews = int(query.User.Contributions.TotalPullRequestReviewContributions)
	stat.PullRequests = int(query.User.Contributions.TotalPullRequestContributions)
	stat.Issues = int(query.User.Contributions.TotalIssueContributions)
	stat.ContributedTo = int(query.User.ContributedTo.TotalCount)
	return &stat, nil
}

func (s *Source) OrgStat(ctx context.Context, username string, org string) (*OrgStat, error) {
	// Can't got Organization ID in API v4
	o, _, err := s.cliv3.Organizations.Get(ctx, org)
	if err != nil {
		return nil, err
	}
	var query struct {
		User struct {
			Contributions struct {
				TotalCommitContributions            ghv4.Int
				TotalPullRequestReviewContributions ghv4.Int
				TotalPullRequestContributions       ghv4.Int
				TotalIssueContributions             ghv4.Int
			} `graphql:"contributionsCollection(from: $from, organizationID: $orgID)"`
		} `graphql:"user(login: $username)"`
	}

	now := time.Now().UTC()
	variables := map[string]interface{}{
		"from":     ghv4.DateTime{now.AddDate(-1, 0, 0)},
		"orgID":    ghv4.ID(*o.NodeID),
		"username": ghv4.String(username),
	}

	err = s.cliv4.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	stat := OrgStat{}
	stat.LogoURL = *o.AvatarURL
	stat.Name = *o.Name
	stat.Commits = int(query.User.Contributions.TotalCommitContributions)
	stat.Reviews = int(query.User.Contributions.TotalPullRequestReviewContributions)
	stat.PullRequests = int(query.User.Contributions.TotalPullRequestContributions)
	stat.Issues = int(query.User.Contributions.TotalIssueContributions)
	return &stat, nil
}

func (s *Source) CommitCounter(ctx context.Context, username string) (int, error) {
	result, _, err := s.cliv3.Search.Commits(ctx, fmt.Sprintf("author:%q", username), &ghv3.SearchOptions{
		ListOptions: ghv3.ListOptions{PerPage: 1},
	})
	if err != nil {
		return 0, err
	}
	return *result.Total, nil
}

type PullRequest struct {
	Username     string
	Title        string
	URL          *url.URL
	BaseRef      string
	State        string
	Additions    int
	Deletions    int
	Commits      int
	ChangedFiles int
	ChangeSize   string
	CreatedAt    time.Time
	ClosedAt     time.Time
	MergedAt     time.Time
	UpdatedAt    time.Time
	Labels       []string
	SortTime     time.Time
}

type PullRequestCallback func(pr *PullRequest) bool

func (s *Source) PullRequests(ctx context.Context, username string, states []PullRequestState, orderField IssueOrderField, orderDirection OrderDirection, size int, cbs ...PullRequestCallback) ([]*PullRequest, error) {
	var pageSize = MaxPageSize
	if size >= 0 && pageSize > size {
		pageSize = size
	}
	if len(states) == 0 {
		states = []PullRequestState{PullRequestStateOpen}
	}

	orderBy := ghv4.IssueOrder{
		Field:     orderField,
		Direction: orderDirection,
	}
	type pr struct {
		Author struct {
			Login ghv4.String
		}
		Title        ghv4.String
		URL          ghv4.URI
		BaseRefName  ghv4.String
		State        ghv4.PullRequestState
		Additions    ghv4.Int
		Deletions    ghv4.Int
		ChangedFiles ghv4.Int
		CreatedAt    ghv4.DateTime
		ClosedAt     ghv4.DateTime
		MergedAt     ghv4.DateTime
		UpdatedAt    ghv4.DateTime
		Commits      struct {
			TotalCount ghv4.Int
			Nodes      []struct {
				Commit struct {
					CommitUrl ghv4.URI
				}
			}
		} `graphql:"commits(first: 1)"`
		MergeCommit struct {
			Parents struct {
				TotalCount ghv4.Int
				Nodes      []struct {
					CommitUrl ghv4.URI
				}
			} `graphql:"parents(first: 2)"`
		}
		Labels struct {
			TotalCount ghv4.Int
			Nodes      []struct {
				Name ghv4.String
			}
		} `graphql:"labels(first: 100)"`
	}

	conv := func(r *pr) *PullRequest {
		p := PullRequest{
			Username:     string(r.Author.Login),
			Title:        string(r.Title),
			URL:          r.URL.URL,
			BaseRef:      string(r.BaseRefName),
			State:        string(r.State),
			Additions:    int(r.Additions),
			Deletions:    int(r.Deletions),
			ChangedFiles: int(r.ChangedFiles),
			ChangeSize:   changeSize(int(r.Additions + r.Deletions)),
			Commits:      int(r.Commits.TotalCount),
			CreatedAt:    r.CreatedAt.Time,
			ClosedAt:     r.ClosedAt.Time,
			MergedAt:     r.MergedAt.Time,
			UpdatedAt:    r.UpdatedAt.Time,
			SortTime:     r.UpdatedAt.Time,
		}
		if len(r.Labels.Nodes) != 0 {
			labels := make([]string, 0, len(r.Labels.Nodes))
			for _, label := range r.Labels.Nodes {
				labels = append(labels, string(label.Name))
				if string(label.Name) == "tide/merge-method-squash" {
					p.Commits = 1
				}
			}
			p.Labels = labels
		}
		if len(states) == 1 {
			switch states[0] {
			case PullRequestStateMerged:
				p.SortTime = p.MergedAt
			case PullRequestStateClosed:
				p.SortTime = p.ClosedAt
			case PullRequestStateOpen:
				p.SortTime = p.CreatedAt
			}
		}
		return &p
	}
	var query struct {
		User struct {
			PullRequests struct {
				TotalCount ghv4.Int
				PageInfo   struct {
					HasNextPage ghv4.Boolean
					EndCursor   ghv4.String
				}
				Nodes []pr
			} `graphql:"pullRequests(first: $size, states: $states, after: $after, orderBy: $orderBy)"`
		} `graphql:"user(login: $username)"`
	}
	variables := map[string]interface{}{
		"username": ghv4.String(username),
		"states":   states,
		"size":     ghv4.Int(pageSize),
		"after":    (*ghv4.String)(nil),
		"orderBy":  orderBy,
	}
	err := s.cliv4.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	count := int(query.User.PullRequests.TotalCount)
	next := bool(query.User.PullRequests.PageInfo.HasNextPage)
	cursor := string(query.User.PullRequests.PageInfo.EndCursor)
	prs := make([]*PullRequest, 0, count)
	for _, r := range query.User.PullRequests.Nodes {
		pr := conv(&r)
		if len(cbs) != 0 {
			for _, cb := range cbs {
				if !cb(pr) {
					return prs, nil
				}
			}
		}
		prs = append(prs, pr)
	}

	for next && cursor != "" &&
		(size < 0 || len(prs) < size) {
		pageSize := pageSize
		if size >= 0 && pageSize+len(prs) > size {
			pageSize = size - len(prs)
		}
		variables := map[string]interface{}{
			"username": ghv4.String(username),
			"states":   states,
			"size":     ghv4.Int(pageSize),
			"after":    ghv4.String(cursor),
			"orderBy":  orderBy,
		}
		err := s.cliv4.Query(ctx, &query, variables)
		if err != nil {
			return nil, err
		}
		for _, r := range query.User.PullRequests.Nodes {
			pr := conv(&r)
			if len(cbs) != 0 {
				for _, cb := range cbs {
					if !cb(pr) {
						return prs, nil
					}
				}
			}
			prs = append(prs, pr)
		}
		next = bool(query.User.PullRequests.PageInfo.HasNextPage)
		cursor = string(query.User.PullRequests.PageInfo.EndCursor)
	}
	return prs, nil
}

func changeSize(i int) string {
	switch {
	case i < 10:
		return "XS"
	case i < 30:
		return "S"
	case i < 100:
		return "M"
	case i < 500:
		return "L"
	case i < 1000:
		return "XL"
	default:
		return "XXL"
	}
}
