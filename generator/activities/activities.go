package activities

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/wzshiming/profile_stats"
	"github.com/wzshiming/profile_stats/generator/activities/render"
	"github.com/wzshiming/profile_stats/source"
	"github.com/wzshiming/profile_stats/utils"
)

type Activities struct {
	source *source.Source
}

func NewActivities(src *source.Source) *Activities {
	return &Activities{
		source: src,
	}
}

func (a *Activities) Generate(ctx context.Context, w io.Writer, args profile_stats.Args) (err error) {
	usernames, ok := args.String("username")
	if !ok || usernames == "" {
		return fmt.Errorf("no usernames")
	}

	size, ok := args.Int("size")
	if !ok {
		size = -1
	}

	var last time.Time
	span, ok := args.String("span")
	if !ok {
		span = "1years"
	}
	if span != "" {
		now := time.Now()
		last, err = utils.ParseTimeSpan(span, now)
		if err != nil {
			return err
		}
	}

	repository, _ := args.String("repository")
	branch, _ := args.String("branch")

	var states []source.PullRequestState
	statesRaw, ok := args.String("states")
	if ok {
		statesSlice := strings.Split(statesRaw, ",")
		states = make([]source.PullRequestState, 0, len(statesSlice))
		for _, state := range statesSlice {
			s := source.PullRequestState(strings.ToUpper(state))
			switch s {
			default:
				return fmt.Errorf("can't support %q", state)
			case source.PullRequestStateOpen, source.PullRequestStateClosed, source.PullRequestStateMerged:
			}
			states = append(states, s)
		}
	}
	if len(states) == 0 {
		states = []source.PullRequestState{source.PullRequestStateOpen, source.PullRequestStateClosed, source.PullRequestStateMerged}
	}

	return a.Get(ctx, w, strings.Split(usernames, ","), size, states, repository, branch, last)
}

func (a *Activities) Get(ctx context.Context, w io.Writer, usernames []string, size int, states []source.PullRequestState, repository, branch string, last time.Time) error {
	items := []*source.PullRequest{}

	cbs := []source.PullRequestCallback{}
	if !last.IsZero() {
		cbs = append(cbs, func(pr *source.PullRequest) bool {
			return pr.CreatedAt.After(last)
		})
	}

	sort.Strings(usernames)
	for _, username := range usernames {
		prs, err := a.source.PullRequests(ctx, username,
			states,
			source.IssueOrderFieldCreatedAt, source.OrderDirectionDesc, size,
			cbs...)
		if err != nil {
			return fmt.Errorf("list PullRequests %q: %w", username, err)
		}

		for _, pr := range prs {
			if pr.SortTime.Before(last) {
				continue
			}
			if !utils.Match(branch, pr.BaseRef) {
				continue
			}
			repo := strings.TrimPrefix(strings.Split(pr.URL.Path, "/pull/")[0], "/")
			if !utils.Match(repository, repo) {
				continue
			}
			items = append(items, pr)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].SortTime.After(items[j].SortTime)
	})
	data := render.ActivitiesData{
		Items: formatSourceActivities(items),
	}

	return render.ActivitiesRender(w, data)
}

func formatSourceActivities(prs []*source.PullRequest) []render.ActivitiesItem {
	items := make([]render.ActivitiesItem, 0, len(prs))
	for _, pr := range prs {
		refAndIndex := strings.SplitN(pr.URL.Path, "/pull/", 2)
		ref := strings.TrimPrefix(refAndIndex[0], "/")
		index := refAndIndex[1]
		state := pr.State
		switch state {
		case string(source.PullRequestStateMerged):
			mergedAt := formatTime(pr.MergedAt)
			state = fmt.Sprintf("Merged (%s)", mergedAt)
		case string(source.PullRequestStateOpen):
			createdAt := formatTime(pr.CreatedAt)
			updatedAt := formatTime(pr.UpdatedAt)
			if createdAt == updatedAt {
				state = fmt.Sprintf("Open (%s)", createdAt)
			} else {
				state = fmt.Sprintf("Open (%s, %s)", createdAt, updatedAt)
			}
		case string(source.PullRequestStateClosed):
			closedAt := formatTime(pr.ClosedAt)
			state = fmt.Sprintf("Closed (%s)", closedAt)
		}

		item := render.ActivitiesItem{
			URL:          pr.URL.String(),
			Username:     pr.Username,
			Link:         fmt.Sprintf("%s#%s", ref, index),
			Title:        pr.Title,
			BaseRef:      pr.BaseRef,
			State:        state,
			Additions:    pr.Additions,
			Deletions:    pr.Deletions,
			Commits:      pr.Commits,
			ChangedFiles: pr.ChangedFiles,
			ChangeSize:   pr.ChangeSize,
			CreatedAt:    pr.CreatedAt,
			ClosedAt:     pr.ClosedAt,
			MergedAt:     pr.MergedAt,
			UpdatedAt:    pr.UpdatedAt,
		}
		items = append(items, item)
	}
	return items
}

func formatTime(t time.Time) string {
	return t.Local().Format("2006-01-02")
}
