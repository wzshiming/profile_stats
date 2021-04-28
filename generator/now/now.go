package now

import (
	"context"
	"io"
	"time"

	"github.com/wzshiming/profile_stats"
)

type Now struct {
}

func NewNow() *Now {
	return &Now{}
}

func (p *Now) Generate(ctx context.Context, w io.Writer, args profile_stats.Args) error {
	w.Write([]byte(time.Now().Format(time.RFC3339)))
	return nil
}
