package render

import (
	"embed"
	"io"
	"log"
	"text/template"
	"time"

	"github.com/wzshiming/profile_stats/generator/common"
	"github.com/wzshiming/profile_stats/render"
)

var (
	activitiesTemplate *template.Template
)

//go:embed layouts
var resource embed.FS

func init() {
	var err error
	activitiesTemplate, err = template.New("_").
		Funcs(render.Funcs).
		ParseFS(resource, "layouts/*.md")
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
	Username     string
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
			f, err := common.Resource.ReadFile("icons/" + item.Status + ".svg")
			if err != nil {
				return err
			}
			data.Items[i].IconData = string(f)
		}
	}

	w = render.NewCompressedSpacesWriter(w)
	return activitiesTemplate.ExecuteTemplate(w, data.Layout+".md", data)
}
