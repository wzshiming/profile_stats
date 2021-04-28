package stats

import (
	"context"
	"fmt"
	"io"

	"github.com/wzshiming/profile_stats"
	"github.com/wzshiming/profile_stats/render"
	"github.com/wzshiming/profile_stats/source"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Stats struct {
	source *source.Source
}

func NewStats(src *source.Source) *Stats {
	return &Stats{
		source: src,
	}
}

func (s *Stats) Generate(ctx context.Context, w io.Writer, args profile_stats.Args) error {
	username, ok := args.Lookup("username")
	if !ok || username == "" {
		return fmt.Errorf("no username")
	}
	return s.Get(ctx, w, username)
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
	return render.StatsRender(w, data)
}

type HandleStatsData func(s *render.StatsData)

func formatSourceStats(stat *source.Stat) []render.StatsItem {
	return []render.StatsItem{
		{
			Id:    "stars",
			Key:   "Total stars",
			Value: formatInt(stat.Stars),
		},
		{
			Id:    "forks",
			Key:   "Total forks",
			Value: formatInt(stat.Forks),
		},
		{
			Id:    "contributeds",
			Key:   "Contributed to",
			Value: formatInt(stat.ContributedTo),
		},
		{
			Id:    "issues",
			Key:   "Issues in a year",
			Value: formatInt(stat.Issues),
		},
		{
			Id:    "commits",
			Key:   "Commits in a year",
			Value: formatInt(stat.Commits),
		},
		{
			Id:    "reviews",
			Key:   "Reviews in a year",
			Value: formatInt(stat.Reviews),
		},
		{
			Id:    "prs",
			Key:   "PRs in a year",
			Value: formatInt(stat.PullRequests),
		},
	}
}

var fmtEn = message.NewPrinter(language.English)

func formatInt(i int) string {
	return fmtEn.Sprint(i)
}
