package render

import (
	"embed"
	"io"
	"log"
	"text/template"
)

var (
	//go:embed stats
	statsResource embed.FS
	statsTemplate *template.Template
)

func init() {
	var err error
	statsTemplate, err = template.New("_").
		Funcs(funcs).
		ParseFS(statsResource, "stats/layouts/*.svg", "stats/themes/*.css")
	if err != nil {
		log.Fatal(err)
	}
}

type StatsData struct {
	Layout              string
	Theme               string
	CSS                 string
	Title               string
	TitleColor          string
	DarkTitleColor      string
	KeyColor            string
	DarkKeyColor        string
	ValueColor          string
	DarkValueColor      string
	IconColor           string
	DarkIconColor       string
	BackgroundColor     string
	DarkBackgroundColor string
	StrokeColor         string
	DarkStrokeColor     string
	Items               []StatsItem
}

type StatsItem struct {
	Id       string
	IconData string
	Key      string
	Value    string
}

func StatsRender(w io.Writer, data StatsData) error {
	if data.Theme == "" {
		data.Theme = "default"
	}
	if data.Layout == "" {
		data.Layout = "default"
	}
	if len(data.Items) == 0 {
		data.Items = itemsDefault
	}
	for i, item := range data.Items {
		if item.IconData == "" && item.Id != "" {
			f, err := statsResource.ReadFile("stats/icons/" + item.Id + ".svg")
			if err != nil {
				return err
			}
			data.Items[i].IconData = string(f)
		}
	}

	if data.CSS == "" {
		buf := getBuffer()
		err := statsTemplate.ExecuteTemplate(buf, data.Theme+".css", data)
		if err != nil {
			putBuffer(buf)
			return err
		}
		data.CSS = buf.String()
		putBuffer(buf)
	}

	w = &compressedSpacesWriter{
		writer: w,
	}
	return statsTemplate.ExecuteTemplate(w, data.Layout+".svg", data)
}

var itemsDefault = []StatsItem{
	{
		Id:  "stars",
		Key: "Total Stars",
	},
	{
		Id:  "forks",
		Key: "Total Forks",
	},
	{
		Id:  "issues",
		Key: "Total Issues",
	},
	{
		Id:  "commits",
		Key: "Total Commits",
	},
	{
		Id:  "reviews",
		Key: "Total Reviews",
	},
	{
		Id:  "prs",
		Key: "Total PRs",
	},
	{
		Id:  "contributeds",
		Key: "Contributed to",
	},
}
