package placeholder

import (
	"context"
	"io"

	"github.com/wzshiming/profile_stats/render"
)

type PlaceHolder struct {
}

func NewPlaceHolder() *PlaceHolder {
	return &PlaceHolder{}
}

func (s *PlaceHolder) Get(ctx context.Context, w io.Writer, text string, handles ...HandlePlaceHolderData) error {
	data := render.PlaceHolderData{
		Text: text,
	}
	for _, handle := range handles {
		handle(&data)
	}
	return render.PlaceHolderRender(w, data)
}

type HandlePlaceHolderData func(s *render.PlaceHolderData)
