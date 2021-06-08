package charts

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/wzshiming/profile_stats"
	"github.com/wzshiming/profile_stats/generator/charts/render"
	"github.com/wzshiming/profile_stats/source"
	"github.com/wzshiming/profile_stats/utils"
)

type Charts struct {
	source *source.Source
}

func NewCharts(src *source.Source) *Charts {
	return &Charts{
		source: src,
	}
}

const (
	KindCommits = "commits"
	KindPRs     = "prs"
)

func (a *Charts) Generate(ctx context.Context, w io.Writer, args profile_stats.Args) (err error) {
	usernames, ok := args.String("username")
	if !ok || usernames == "" {
		return fmt.Errorf("no usernames")
	}

	size, ok := args.Int("size")
	if !ok {
		size = -1
	}

	kind, ok := args.String("kind")
	if !ok {
		kind = KindCommits
	}
	kind = strings.ToLower(kind)

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

	title, ok := args.String("title")
	if !ok {
		title = kind + " " + strings.ReplaceAll(statesRaw, ",", "/") + " in the last " + span + " in the " + repository
	}

	width, _ := args.Int("width")
	if width == 0 {
		width = 1200
	}

	height, _ := args.Int("height")
	if height == 0 {
		height = 800
	}

	maxVal, _ := args.Int("max_value")
	if maxVal == 0 {
		maxVal = 49
	}

	return a.Get(ctx, w, title, strings.Split(usernames, ","), size, states, repository, branch, last, kind, width, height, maxVal)
}

func (a *Charts) Get(ctx context.Context, w io.Writer, title string, usernames []string, size int, states []source.PullRequestState, repository, branch string, last time.Time, kind string, width, height, maxVal int) error {
	data := render.ChartData{
		Title:        title,
		ValueMessage: kind,
		Width:        width,
		Height:       height,
		MaxValue:     maxVal,
	}

	cbs := []source.PullRequestCallback{}
	if !last.IsZero() {
		cbs = append(cbs, func(pr *source.PullRequest) bool {
			return pr.CreatedAt.After(last)
		})
	}

	usernames, attrs := utils.KeyAttribute(usernames)

	for i, username := range usernames {
		prs, err := a.source.PullRequests(ctx, username,
			states,
			source.IssueOrderFieldCreatedAt, source.OrderDirectionDesc, size,
			cbs...)
		if err != nil {
			return fmt.Errorf("list PullRequests %q: %w", username, err)
		}
		if len(prs) == 0 {
			continue
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

		days := map[string]int{}
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
			if !before.IsZero() && !pr.SortTime.Before(before) {
				continue
			}
			if !after.IsZero() && !pr.SortTime.After(after) {
				continue
			}

			key := pr.SortTime.Format(render.DateFmt)
			switch kind {
			case KindCommits:
				days[key] = days[key] + pr.Commits
			case KindPRs:
				days[key] = days[key] + 1
			}
		}
		if len(days) == 0 {
			continue
		}
		points := make(render.Points, 0, len(days))
		for date, val := range days {
			t, _ := time.Parse(render.DateFmt, date)
			points = append(points, render.Point{
				Value: val,
				Time:  t,
			})
		}

		name := username
		if n := attrs[username]["name"]; n != "" {
			name = n
		}
		data.Series = append(data.Series, render.Series{
			Name:   name,
			Points: points,
			Index:  i,
		})
	}

	return render.ChartRender(w, data)
}
