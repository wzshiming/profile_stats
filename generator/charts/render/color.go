package render

import (
	"fmt"
	"image/color"
	"math"

	"github.com/vdobler/chart"
)

var (
	StandardColors = []color.Color{}
)

func init() {
	standardColors := []color.Color{
		color.NRGBA{0x43, 0x85, 0xf4, 0xff},
		color.NRGBA{0xea, 0x44, 0x35, 0xff},
		color.NRGBA{0xa3, 0x7f, 0xe2, 0xff},
		color.NRGBA{0xfb, 0xbc, 0x03, 0xff},
		color.NRGBA{0x33, 0xa8, 0x54, 0xff},
		color.NRGBA{0xff, 0x6d, 0x04, 0xff},
		color.NRGBA{0x47, 0xbd, 0xc6, 0xff},
	}

	StandardColors = append(StandardColors, standardColors...)
	for _, color := range standardColors {
		StandardColors = append(StandardColors, lighter(color, 0.3))
	}
	for _, color := range standardColors {
		StandardColors = append(StandardColors, lighter(color, 0.7))
	}
}

func AutoStyle(i int) (style chart.Style) {
	nc := len(StandardColors)
	ci := i % nc
	fillColor := StandardColors[ci]
	lineColor := fillColor
	if n := i / nc; n > 0 {
		lineColor = darker(lineColor, 0.01*float64(n))
		style.LineColor = lineColor
		style.LineStyle = chart.SolidLine
		style.LineWidth = 3
	}
	style.FillColor = fillColor

	return
}

func hsv2rgb(h, s, v int) (r, g, b int) {
	H := int(math.Floor(float64(h) / 60))
	S, V := float64(s)/100, float64(v)/100
	f := float64(h)/60 - float64(H)
	p := V * (1 - S)
	q := V * (1 - S*f)
	t := V * (1 - S*(1-f))

	switch H {
	case 0, 6:
		r, g, b = int(255*V), int(255*t), int(255*p)
	case 1:
		r, g, b = int(255*q), int(255*V), int(255*p)
	case 2:
		r, g, b = int(255*p), int(255*V), int(255*t)
	case 3:
		r, g, b = int(255*p), int(255*q), int(255*V)
	case 4:
		r, g, b = int(255*t), int(255*p), int(255*V)
	case 5:
		r, g, b = int(255*V), int(255*p), int(255*q)
	default:
		panic(fmt.Sprintf("Ooops: Strange H value %d in hsv2rgb(%d,%d,%d).", H, h, s, v))
	}

	return
}

func f3max(a, b, c float64) float64 {
	switch true {
	case a > b && a >= c:
		return a
	case b > c && b >= a:
		return b
	case c > a && c >= b:
		return c
	}
	return a
}

func f3min(a, b, c float64) float64 {
	switch true {
	case a < b && a <= c:
		return a
	case b < c && b <= a:
		return b
	case c < a && c <= b:
		return c
	}
	return a
}

func rgb2hsv(r, g, b int) (h, s, v int) {
	R, G, B := float64(r)/255, float64(g)/255, float64(b)/255

	if R == G && G == B {
		h, s = 0, 0
		v = int(r * 255)
	} else {
		max, min := f3max(R, G, B), f3min(R, G, B)
		if max == R {
			h = int(60 * (G - B) / (max - min))
		} else if max == G {
			h = int(60 * (2 + (B-R)/(max-min)))
		} else {
			h = int(60 * (4 + (R-G)/(max-min)))
		}
		if max == 0 {
			s = 0
		} else {
			s = int(100 * (max - min) / max)
		}
		v = int(100 * max)
	}
	if h < 0 {
		h += 360
	}
	return
}

func lighter(col color.Color, f float64) color.Color {
	r, g, b, a := col.RGBA()
	h, s, v := rgb2hsv(int(r/256), int(g/256), int(b/256))
	f = 1 - f
	s = int(float64(s) * f)
	v += int((100 - float64(v)) * f)
	if v > 100 {
		v = 100
	}
	rr, gg, bb := hsv2rgb(h, s, v)

	return color.NRGBA{uint8(rr), uint8(gg), uint8(bb), uint8(a / 256)}
}

func darker(col color.Color, f float64) color.Color {
	r, g, b, a := col.RGBA()
	h, s, v := rgb2hsv(int(r), int(g), int(b))
	f = 1 - f
	v = int(float64(v) * f)
	s += int((100 - float64(s)) * f)
	if s > 100 {
		s = 100
	}
	rr, gg, bb := hsv2rgb(h, s, v)

	return color.NRGBA{uint8(rr), uint8(gg), uint8(bb), uint8(a / 256)}
}
