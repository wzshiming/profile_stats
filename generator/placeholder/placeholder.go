package placeholder

import (
	"context"
	"fmt"
	"io"

	"github.com/wzshiming/profile_stats"
	"github.com/wzshiming/profile_stats/generator/placeholder/render"
)

type PlaceHolder struct {
}

func NewPlaceHolder() *PlaceHolder {
	return &PlaceHolder{}
}

func (p *PlaceHolder) Generate(ctx context.Context, w io.Writer, args profile_stats.Args) error {
	text, ok := args.String("text")
	if !ok || text == "" {
		return fmt.Errorf("no text")
	}
	return p.Get(ctx, w, text)
}

func (p *PlaceHolder) Get(ctx context.Context, w io.Writer, text string) error {
	data := render.PlaceHolderData{
		Text: text,
	}

	return render.PlaceHolderRender(w, data)
}
