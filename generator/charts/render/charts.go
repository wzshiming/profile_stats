package render

import (
	"fmt"
	"image/color"
	"io"
	"sort"
	"strconv"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/vdobler/chart"
	"github.com/vdobler/chart/svgg"
)

const (
	DateFmt  = "2006-01"
	baseYear = 1990
)

func EncodeYearMonth(t time.Time) int {
	t = t.Local()
	return (t.Year()-baseYear)*12 + int(t.Month())
}

func DecodeYearMonth(i int) time.Time {
	return time.Date(i/12+baseYear, time.Month(i%12), 1, 0, 0, 0, 0, time.Local)
}

type Point struct {
	Time  time.Time
	Value int
}

func (p Point) String() string {
	return fmt.Sprintf("<%s: %d>", p.Time.Format(DateFmt), p.Value)
}

type Points []Point

func (p Points) sort() {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Time.Before(p[j].Time)
	})
}

func (p Points) Values() (points []chart.Point) {
	p.sort()
	points = make([]chart.Point, 0, len(p))
	last := -1
	for _, p := range p {
		m := EncodeYearMonth(p.Time)
		if last == -1 {
			last = m - 1
		}
		for m-last > 1 {
			points = append(points, chart.Point{
				X: float64(last + 1),
				Y: 0,
			})
			last++
		}
		points = append(points, chart.Point{
			X: float64(EncodeYearMonth(p.Time)),
			Y: float64(p.Value),
		})
		last = m
	}
	return points
}

type Series struct {
	Name   string
	Points Points
	Index  int
}

type ChartData struct {
	Title        string
	TimeMessage  string
	ValueMessage string
	Width        int
	Height       int
	Series       []Series
}

func ChartRender(w io.Writer, data ChartData) error {
	graph := chart.BarChart{
		Stacked:      true,
		SameBarWidth: true,
		Title:        data.Title,
		ShowVal:      3,
	}

	for _, series := range data.Series {
		if series.Name == "" {
			continue
		}
		style := chart.AutoStyle(series.Index, true)
		points := series.Points.Values()
		graph.AddData(series.Name, points, style)
	}

	graph.Key.Pos = "ort"
	graph.XRange.TicSetting.Delta = 1
	graph.XRange.TicSetting.Format = func(f float64) string {
		d := int(f)
		return DecodeYearMonth(d).Format(DateFmt)
	}

	if graph.YRange.TicSetting.Delta < 1 {
		graph.YRange.TicSetting.Delta = 1
	}
	graph.YRange.Min = 1
	graph.YRange.TicSetting.Format = func(f float64) string {
		return strconv.FormatInt(int64(f), 10)
	}

	width := data.Width
	height := data.Height
	w.Write([]byte(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height)))
	defer w.Write([]byte(`</svg>`))
	render := svg.New(w)
	graphics := svgg.New(render, width, height, "", 0, color.RGBA{})
	graph.Plot(graphics)
	return nil
}
