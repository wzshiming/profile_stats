package profile_stats

import (
	"context"
	"io"
)

type Args interface {
	Lookup(name string) (string, bool)
	Get(name string) string
}

type Generator interface {
	Generate(ctx context.Context, w io.Writer, args Args) error
}
