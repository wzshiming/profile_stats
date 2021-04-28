package generator

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"github.com/wzshiming/profile_stats"
	"github.com/wzshiming/profile_stats/generator/activities"
	"github.com/wzshiming/profile_stats/generator/now"
	"github.com/wzshiming/profile_stats/generator/placeholder"
	"github.com/wzshiming/profile_stats/generator/stats"
	"github.com/wzshiming/profile_stats/source"
	"github.com/wzshiming/xmlinjector"
)

var (
	key = []byte("PROFILE_STATS")
)

type Handler struct {
	registry map[string]profile_stats.Generator
}

func NewHandler(src *source.Source) *Handler {
	r := &Handler{
		registry: map[string]profile_stats.Generator{},
	}

	r.register("now", now.NewNow())
	r.register("updatedat", now.NewNow())
	r.register("stats", placeholder.NewPlaceHolder())
	r.register("activities", activities.NewActivities(src))
	r.register("stats", stats.NewStats(src))
	return r
}

func (r *Handler) register(name string, generator profile_stats.Generator) {
	r.registry[name] = generator
}

func (r *Handler) Handle(ctx context.Context, origin []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	return xmlinjector.Inject(key, origin, func(args, origin []byte) []byte {
		tag := reflect.StructTag(args)
		template, ok := tag.Lookup("template")
		if !ok || template == "" {
			return errInfo("no template")
		}

		generator, ok := r.registry[template]
		if !ok {
			return errInfo(fmt.Sprintf("not support template %q", template))
		}
		buf.Reset()
		buf.Write([]byte("\n"))
		err := generator.Generate(ctx, buf, tag)
		if err != nil {
			return errInfo(err.Error())
		}
		buf.Write([]byte("\n"))
		return buf.Bytes()
	})
}

func errInfo(msg string) []byte {
	return []byte(fmt.Sprintf("\n<!-- error: %q -->\n", msg))
}
