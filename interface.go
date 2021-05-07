package profile_stats

import (
	"context"
	"io"
)

type Args interface {
	String(name string) (string, bool)
	Int(name string) (int, bool)
}

type Generator interface {
	Generate(ctx context.Context, w io.Writer, args Args) error
}
