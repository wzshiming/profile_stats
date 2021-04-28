package placeholder

import (
	"context"
	"fmt"
	"io"

	"github.com/wzshiming/profile_stats"
	"github.com/wzshiming/profile_stats/render"
)

type PlaceHolder struct {
}

func NewPlaceHolder() *PlaceHolder {
	return &PlaceHolder{}
}

func (p *PlaceHolder) Generate(ctx context.Context, w io.Writer, args profile_stats.Args) error {
	text, ok := args.Lookup("text")
	if !ok || text == "" {
		return fmt.Errorf("no text")
	}
	return p.Get(ctx, w, text)
}

func (p *PlaceHolder) Get(ctx context.Context, w io.Writer, text string, handles ...HandlePlaceHolderData) error {
	data := render.PlaceHolderData{
		Text: text,
	}
	for _, handle := range handles {
		handle(&data)
	}
	return render.PlaceHolderRender(w, data)
}

type HandlePlaceHolderData func(s *render.PlaceHolderData)
