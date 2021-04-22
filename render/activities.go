package render

import (
	"io"
	"log"
	"text/template"
	"time"
)

var (
	activitiesTemplate *template.Template
)

func init() {
	var err error
	activitiesTemplate, err = template.New("_").
		Funcs(funcs).
		ParseFS(resource, "activities/layouts/*.md")
	if err != nil {
		log.Fatal(err)
	}
}

type ActivitiesData struct {
	Layout string
	Title  string
	Items  []ActivitiesItem
}

type ActivitiesItem struct {
	Status       string
	IconData     string
	URL          string
	Link         string
	Title        string
	BaseRef      string
	State        string
	Additions    int
	Deletions    int
	Commits      int
	ChangedFiles int
	ChangeSize   string
	CreatedAt    time.Time
	ClosedAt     time.Time
	MergedAt     time.Time
	UpdatedAt    time.Time
}

func ActivitiesRender(w io.Writer, data ActivitiesData) error {
	if data.Layout == "" {
		data.Layout = "default"
	}

	for i, item := range data.Items {
		if item.IconData == "" && item.Status != "" {
			f, err := resource.ReadFile("icons/" + item.Status + ".svg")
			if err != nil {
				return err
			}
			data.Items[i].IconData = string(f)
		}
	}

	w = &compressedSpacesWriter{
		writer: w,
	}
	return activitiesTemplate.ExecuteTemplate(w, data.Layout+".md", data)
}
