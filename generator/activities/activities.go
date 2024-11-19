package activities

import (
	"context"
	"fmt"
	"io"
	"log"
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
	usernames, ok := args.StringSlice("username")
	if !ok {
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

	repository, _ := args.StringSlice("repository")
	branch, _ := args.StringSlice("branch")
	labels, _ := args.StringSlice("labels")
	labelsFilter, _ := args.StringSlice("labels_filter")

	var states []source.PullRequestState
	statesSlice, ok := args.StringSlice("states")
	if ok {
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

	return a.Get(ctx, w, usernames, size, states, repository, branch, labels, labelsFilter, last)
}

func (a *Activities) Get(ctx context.Context, w io.Writer, usernames []string, size int, states []source.PullRequestState, repository, branch, labels, labelsFilter []string, last time.Time) error {
	items := []*source.PullRequest{}

	cbs := []source.PullRequestCallback{}
	if !last.IsZero() {
		cbs = append(cbs, func(pr *source.PullRequest) bool {
			return pr.MergedAt.After(last)
		})
	}

	usernames, attrs := utils.KeyAttribute(usernames)

	for _, username := range usernames {
		prs, err := a.source.PullRequests(ctx, username,
			states,
			source.IssueOrderFieldCreatedAt, source.OrderDirectionDesc, size,
			cbs...)
		if err != nil {
			return fmt.Errorf("list PullRequests %q: %w", username, err)
		}

		var before, after time.Time
		if v := attrs[username]["before"]; v != "" {
			before, err = utils.ParseTime(v, time.Local)
			if err != nil {
				log.Printf("error ParseTime: %s", err)
			}
		}
		if v := attrs[username]["after"]; v != "" {
			after, err = utils.ParseTime(v, time.Local)
			if err != nil {
				log.Printf("error ParseTime: %s", err)
			}
		}

		for _, pr := range prs {
			if pr.SortTime.Before(last) {
				continue
			}
			if len(branch) != 0 && !utils.Match(branch, pr.BaseRef) {
				continue
			}
			if len(labels) != 0 {
				match := false
				for _, label := range pr.Labels {
					if utils.Match(labels, label) {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}
			if len(repository) != 0 {
				repo := strings.TrimPrefix(strings.Split(pr.URL.Path, "/pull/")[0], "/")
				if !utils.Match(repository, repo) {
					continue
				}
			}
			if !before.IsZero() && !pr.SortTime.Before(before) {
				continue
			}
			if !after.IsZero() && !pr.SortTime.After(after) {
				continue
			}

			if n := attrs[username]["name"]; n != "" {
				pr.Username = n
			}
			if len(labelsFilter) != 0 {
				list := make([]string, 0, len(pr.Labels))
				for _, label := range pr.Labels {
					if utils.Match(labelsFilter, label) {
						list = append(list, label)
					}
				}
				pr.Labels = list
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
		item := render.ActivitiesItem{
			URL:          pr.URL.String(),
			Username:     pr.Username,
			Link:         fmt.Sprintf("%s#%s", ref, index),
			Title:        pr.Title,
			BaseRef:      pr.BaseRef,
			State:        pr.State,
			Additions:    pr.Additions,
			Deletions:    pr.Deletions,
			Commits:      pr.Commits,
			ChangedFiles: pr.ChangedFiles,
			ChangeSize:   pr.ChangeSize,
			CreatedAt:    pr.CreatedAt,
			ClosedAt:     pr.ClosedAt,
			MergedAt:     pr.MergedAt,
			UpdatedAt:    pr.UpdatedAt,
			Labels:       pr.Labels,
		}
		items = append(items, item)
	}
	return items
}
