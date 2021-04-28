package activities

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/wzshiming/profile_stats"
	"github.com/wzshiming/profile_stats/render"
	"github.com/wzshiming/profile_stats/source"
)

type Activities struct {
	source *source.Source
}

func NewActivities(src *source.Source) *Activities {
	return &Activities{
		source: src,
	}
}

func (a *Activities) Generate(ctx context.Context, w io.Writer, args profile_stats.Args) error {
	username, ok := args.Lookup("username")
	if !ok || username == "" {
		return fmt.Errorf("no username")
	}
	return a.Get(ctx, w, username)
}

func (a *Activities) Get(ctx context.Context, w io.Writer, username string, handles ...HandleActivitiesData) error {
	stat, err := a.source.PullRequests(ctx, username,
		[]source.PullRequestState{source.PullRequestStateOpen, source.PullRequestStateClosed, source.PullRequestStateMerged},
		source.IssueOrderFieldUpdatedAt, source.OrderDirectionDesc, 50)
	if err != nil {
		return err
	}
	data := render.ActivitiesData{
		Title: username + "'s Activities",
		Items: formatSourceActivities(stat),
	}
	for _, handle := range handles {
		handle(&data)
	}
	return render.ActivitiesRender(w, data)
}

type HandleActivitiesData func(s *render.ActivitiesData)

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
