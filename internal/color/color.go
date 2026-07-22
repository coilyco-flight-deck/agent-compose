// Package color owns favorite-color math: hex parsing, the terminal-legible
// OKLab band, and derived favorites for multi-component identities.
package color

import (
	"fmt"
	"math"
	"strings"
)

// The legible band keeps colors readable on dark and light terminals:
// mid lightness, never gray.
const (
	minL      = 0.60
	maxL      = 0.80
	minChroma = 0.05
)

type okLab struct {
	L, A, B float64
}

func (c okLab) chroma() float64 {
	return math.Hypot(c.A, c.B)
}

// Legible validates a hex color against the band; it is the parse-time
// gate for personality colors.
func Legible(hex string) error {
	lab, err := toOKLab(hex)
	if err != nil {
		return err
	}
	if lab.L < minL || lab.L > maxL {
		return fmt.Errorf("color %s has OKLab lightness %.2f outside the terminal-legible band [%.2f, %.2f]", hex, lab.L, minL, maxL)
	}
	if lab.chroma() < minChroma {
		return fmt.Errorf("color %s is too gray (chroma %.3f, floor %.3f)", hex, lab.chroma(), minChroma)
	}
	return nil
}

// Favorite derives one color from component colors: the OKLab centroid with
// chroma restored to the components' minimum, clamped into the band.
func Favorite(hexes []string) (string, error) {
	if len(hexes) == 0 {
		return "", fmt.Errorf("favorite needs at least one component color")
	}
	var sum okLab
	minComponentChroma := math.Inf(1)
	for _, hex := range hexes {
		lab, err := toOKLab(hex)
		if err != nil {
			return "", err
		}
		sum.L += lab.L
		sum.A += lab.A
		sum.B += lab.B
		minComponentChroma = math.Min(minComponentChroma, lab.chroma())
	}
	n := float64(len(hexes))
	blend := okLab{L: sum.L / n, A: sum.A / n, B: sum.B / n}
	if c := blend.chroma(); c > 0 && c < minComponentChroma {
		scale := minComponentChroma / c
		blend.A *= scale
		blend.B *= scale
	}
	blend.L = math.Max(minL, math.Min(maxL, blend.L))
	if blend.chroma() < minChroma {
		return "", fmt.Errorf("derived favorite fell below the gray floor; give the components more chroma")
	}
	return fromOKLab(blend), nil
}

// ANSI renders text in the color: truecolor when supported, else the
// nearest 256-color index.
func ANSI(hex, text string, trueColor bool) string {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return text
	}
	if trueColor {
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, text)
	}
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", nearest256(r, g, b), text)
}

func parseHex(hex string) (int, int, int, error) {
	h := strings.TrimPrefix(hex, "#")
	if len(h) != 6 {
		return 0, 0, 0, fmt.Errorf("color %q must be #rrggbb", hex)
	}
	var r, g, b int
	if _, err := fmt.Sscanf(h, "%02x%02x%02x", &r, &g, &b); err != nil {
		return 0, 0, 0, fmt.Errorf("color %q must be #rrggbb: %w", hex, err)
	}
	return r, g, b, nil
}

func toOKLab(hex string) (okLab, error) {
	ri, gi, bi, err := parseHex(hex)
	if err != nil {
		return okLab{}, err
	}
	r, g, b := linearize(ri), linearize(gi), linearize(bi)
	l := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*b)
	m := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*b)
	s := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*b)
	return okLab{
		L: 0.2104542553*l + 0.7936177850*m - 0.0040720468*s,
		A: 1.9779984951*l - 2.4285922050*m + 0.4505937099*s,
		B: 0.0259040371*l + 0.7827717662*m - 0.8086757660*s,
	}, nil
}

func fromOKLab(lab okLab) string {
	l := cube(lab.L + 0.3963377774*lab.A + 0.2158037573*lab.B)
	m := cube(lab.L - 0.1055613458*lab.A - 0.0638541728*lab.B)
	s := cube(lab.L - 0.0894841775*lab.A - 1.2914855480*lab.B)
	r := 4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	g := -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	b := -0.0041960863*l - 0.7034186147*m + 1.7076147010*s
	return fmt.Sprintf("#%02x%02x%02x", delinearize(r), delinearize(g), delinearize(b))
}

func linearize(channel int) float64 {
	c := float64(channel) / 255
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func delinearize(c float64) int {
	c = math.Max(0, math.Min(1, c))
	if c <= 0.0031308 {
		c *= 12.92
	} else {
		c = 1.055*math.Pow(c, 1/2.4) - 0.055
	}
	return int(math.Round(c * 255))
}

func cube(v float64) float64 {
	return v * v * v
}

func nearest256(r, g, b int) int {
	toCube := func(v int) int {
		if v < 48 {
			return 0
		}
		if v < 115 {
			return 1
		}
		return (v - 35) / 40
	}
	levels := []int{0, 95, 135, 175, 215, 255}
	cr, cg, cb := toCube(r), toCube(g), toCube(b)
	cubeIndex := 16 + 36*cr + 6*cg + cb
	cubeDist := sq(r-levels[cr]) + sq(g-levels[cg]) + sq(b-levels[cb])

	gray := (r + g + b) / 3
	grayIndex := 232 + max(0, min(23, (gray-8)/10))
	grayLevel := 8 + 10*(grayIndex-232)
	grayDist := sq(r-grayLevel) + sq(g-grayLevel) + sq(b-grayLevel)

	if grayDist < cubeDist {
		return grayIndex
	}
	return cubeIndex
}

func sq(v int) int {
	return v * v
}
