package render

import (
	"embed"
	"io"
	"log"
	"text/template"

	"github.com/wzshiming/profile_stats/render"
)

var (
	placeHolderTemplate *template.Template
)

//go:embed layouts
//go:embed themes
var resource embed.FS

func init() {
	var err error
	placeHolderTemplate, err = template.New("_").
		Funcs(render.Funcs).
		ParseFS(resource, "layouts/*.svg", "themes/*.css")
	if err != nil {
		log.Fatal(err)
	}
}

type PlaceHolderData struct {
	Layout              string
	Theme               string
	CSS                 string
	Text                string
	TextColor           string
	DarkTextColor       string
	BackgroundColor     string
	DarkBackgroundColor string
	StrokeColor         string
	DarkStrokeColor     string
	Width               int
	Height              int
}

func PlaceHolderRender(w io.Writer, data PlaceHolderData) error {
	if data.Theme == "" {
		data.Theme = "default"
	}
	if data.Layout == "" {
		data.Layout = "default"
	}
	if data.Height <= 0 {
		data.Height = 200
	}
	if data.Width <= 0 {
		data.Width = 200
	}

	if data.CSS == "" {
		buf := render.GetBuffer()
		err := placeHolderTemplate.ExecuteTemplate(buf, data.Theme+".css", data)
		if err != nil {
			render.PutBuffer(buf)
			return err
		}
		data.CSS = buf.String()
		render.PutBuffer(buf)
	}

	w = render.NewCompressedSpacesWriter(w)
	return placeHolderTemplate.ExecuteTemplate(w, data.Layout+".svg", data)
}
