package generator

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/wzshiming/profile_stats"
)

func NewArgs(tag string, env bool) profile_stats.Args {
	tag = strings.ReplaceAll(tag, "\n", " ")
	return &args{
		tag: reflect.StructTag(tag),
		env: env,
	}
}

type args struct {
	tag reflect.StructTag
	env bool
}

func (a args) String(name string) (string, bool) {
	val, ok := a.tag.Lookup(name)
	if !ok {
		return "", false
	}
	if a.env {
		val = os.Expand(val, os.Getenv)
	}
	return val, true
}

func (a args) StringSlice(name string) ([]string, bool) {
	val, ok := a.String(name)
	if !ok {
		return []string{}, false
	}
	vals := []string{}
	for _, v := range strings.Split(val, "\n") {
		for _, v := range strings.Split(v, ",") {
			s := strings.TrimSpace(v)
			if s != "" {
				vals = append(vals, s)
			}
		}
	}
	return vals, len(vals) != 0
}

func (a args) Int(name string) (int, bool) {
	raw, ok := a.String(name)
	if !ok {
		return 0, false
	}
	n, _ := strconv.ParseInt(raw, 0, 0)
	return int(n), true
}
