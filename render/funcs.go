package render

import (
	"text/template"
	"unicode/utf8"
)

var funcs = template.FuncMap{
	"add":    add,
	"sub":    sub,
	"mul":    mul,
	"div":    div,
	"min":    min,
	"max":    max,
	"strLen": strLen,
}

func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func mul(a, b int) int {
	return a * b
}

func div(a, b int) int {
	return a / b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func strLen(str string) int {
	i := 0
	for _, v := range str {
		i += runeWidth(v)
	}
	return i
}

func runeWidth(r rune) int {
	switch {
	case r == utf8.RuneError || r < '\x20':
		return 0

	case '\x20' <= r && r < '\u2000':
		return 1

	case '\u2000' <= r && r < '\uFF61':
		return 2

	case '\uFF61' <= r && r < '\uFFA0':
		return 1

	case '\uFFA0' <= r:
		return 2
	}

	return 0
}
