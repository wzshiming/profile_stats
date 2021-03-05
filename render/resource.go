package render

import (
	"bytes"
	"embed"
	"io"
	"log"
	"strings"
	"sync"
	"text/template"
	"unicode"
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
	data.Theme = strings.TrimSuffix(data.Theme, ".css")
	if data.Theme == "" {
		data.Theme = "default"
	}
	data.Layout = strings.TrimSuffix(data.Layout, ".svg")
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

var poolBuffer = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

func getBuffer() *strings.Builder {
	buf := poolBuffer.Get().(*strings.Builder)
	buf.Reset()
	return buf
}

func putBuffer(buf *strings.Builder) {
	poolBuffer.Put(buf)
}

type compressedSpacesWriter struct {
	writer io.Writer
}

func (c *compressedSpacesWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	p = bytes.TrimLeftFunc(p, unicode.IsSpace)
	for len(p) != 0 {
		i := bytes.IndexFunc(p, unicode.IsSpace)
		if i == -1 {
			return c.writer.Write(p)
		}
		_, err = c.writer.Write(p[:i+1])
		if err != nil {
			return 0, err
		}
		p = bytes.TrimLeftFunc(p[i+1:], unicode.IsSpace)
	}
	return n, nil
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
