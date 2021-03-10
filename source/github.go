package source

import (
	"context"
	"fmt"

	ghv3 "github.com/google/go-github/v33/github"
	ghv4 "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

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

func NewSource(token string) *Source {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return &Source{
		cliv3: ghv3.NewClient(httpClient),
		cliv4: ghv4.NewClient(httpClient),
	}
}

type Source struct {
	cliv3 *ghv3.Client
	cliv4 *ghv4.Client
}

func (s *Source) Stat(ctc context.Context, username string) (*Stat, error) {
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
			} `graphql:"contributionsCollection"`
			ContributedTo struct {
				TotalCount ghv4.Int
			} `graphql:"repositoriesContributedTo(first: 0)"`
			PullRequests struct {
				TotalCount ghv4.Int
			} `graphql:"pullRequests(first: 0)"`
			Issues struct {
				TotalCount ghv4.Int
			} `graphql:"issues(first: 0)"`
			// Login ghv4.String
			Name ghv4.String
		} `graphql:"user(login: $username)"`
	}

	variables := map[string]interface{}{
		"username": ghv4.String(username),
	}

	err := s.cliv4.Query(ctc, &query, variables)
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
	stat.ContributedTo = int(query.User.ContributedTo.TotalCount)
	stat.PullRequests = int(query.User.PullRequests.TotalCount)
	stat.Issues = int(query.User.Issues.TotalCount)
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
