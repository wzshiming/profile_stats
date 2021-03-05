package stats

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/wzshiming/profile_stats/render"
	"github.com/wzshiming/profile_stats/source"
)

type Stats struct {
	source *source.Source
}

func NewStats(src *source.Source) *Stats {
	return &Stats{
		source: src,
	}
}

func (s *Stats) Get(ctx context.Context, w io.Writer, username string, handles ...HandleStatsData) error {
	stat, err := s.source.Stat(ctx, username)
	if err != nil {
		return err
	}
	data := render.StatsData{
		Title: stat.Name + "'s Stats",
		Items: formatSourceStats(stat),
	}
	for _, handle := range handles {
		handle(&data)
	}
	err = render.StatsRender(w, data)
	if err != nil {
		return err
	}
	return err
}

type HandleStatsData func(s *render.StatsData)

func formatSourceStats(stat *source.Stat) []render.StatsItem {
	return []render.StatsItem{
		{
			Id:    "stars",
			Key:   "Total Stars",
			Value: formatInt(stat.Stars),
		},
		{
			Id:    "forks",
			Key:   "Total Forks",
			Value: formatInt(stat.Forks),
		},
		{
			Id:    "issues",
			Key:   "Total Issues",
			Value: formatInt(stat.Issues),
		},
		{
			Id:    "commits",
			Key:   fmt.Sprintf("Total Commits (%d)", time.Now().Year()),
			Value: formatInt(stat.Commits),
		},
		{
			Id:    "reviews",
			Key:   "Total Reviews",
			Value: formatInt(stat.Reviews),
		},
		{
			Id:    "prs",
			Key:   "Total PRs",
			Value: formatInt(stat.PullRequests),
		},
		{
			Id:    "contributeds",
			Key:   "Contributed to",
			Value: formatInt(stat.ContributedTo),
		},
	}
}

func formatInt(i int) string {
	return strconv.FormatInt(int64(i), 10)
}
