package render

import (
	"io"
	"log"
	"text/template"
)

var (
	placeHolderTemplate *template.Template
)

func init() {
	var err error
	placeHolderTemplate, err = template.New("_").
		Funcs(funcs).
		ParseFS(resource, "placeholder/layouts/*.svg", "placeholder/themes/*.css")
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
		buf := getBuffer()
		err := placeHolderTemplate.ExecuteTemplate(buf, data.Theme+".css", data)
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
	return placeHolderTemplate.ExecuteTemplate(w, data.Layout+".svg", data)
}
