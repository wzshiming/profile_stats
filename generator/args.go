package generator

import (
	"reflect"
	"strconv"

	"github.com/wzshiming/profile_stats"
)

func NewArgs(tag reflect.StructTag) profile_stats.Args {
	return &args{tag: tag}
}

type args struct {
	tag reflect.StructTag
}

func (a args) String(name string) (string, bool) {
	return a.tag.Lookup(name)
}

func (a args) Int(name string) (int, bool) {
	raw, ok := a.tag.Lookup(name)
	if !ok {
		return 0, false
	}
	n, _ := strconv.ParseInt(raw, 0, 0)
	return int(n), true
}
