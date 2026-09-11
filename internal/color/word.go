package color

import (
	"math"
	"strings"
)

// Below this saturation the hue carries no information, so a word cannot
// disagree with it. HSL saturation, not OKLab chroma.
const flatSaturation = 0.10

// Hue bands for the basic colour terms, in HSL degrees. These describe colour
// language rather than roster policy, which is why they live here.
var colorTerms = map[string][2]float64{
	"red": {350, 372}, "rose": {320, 360}, "pink": {315, 350},
	"coral": {5, 25}, "orange": {15, 42}, "amber": {35, 55},
	"gold": {38, 62}, "olive": {45, 75}, "yellow": {48, 68},
	"pear": {58, 82}, "chartreuse": {65, 90}, "lime": {70, 95},
	"green": {80, 165}, "sage": {90, 160}, "moss": {90, 160},
	"emerald": {130, 165}, "jade": {135, 175}, "teal": {160, 195},
	"cyan": {175, 200}, "azure": {195, 220}, "blue": {200, 250},
	"periwinkle": {215, 248}, "indigo": {240, 270}, "lavender": {255, 290},
	"violet": {255, 292}, "lilac": {262, 298}, "purple": {265, 305},
	"orchid": {288, 322}, "magenta": {290, 330},
}

// WordAgrees reports whether word names the hue hex actually carries. The
// second result is false when nothing about the pair can be checked.
func WordAgrees(hex, word string) (agrees bool, checkable bool) {
	hue, saturation, ok := hueSaturation(hex)
	if !ok || saturation < flatSaturation {
		return false, false
	}
	recognised := false
	for _, segment := range strings.Split(strings.ToLower(strings.TrimSpace(word)), "-") {
		band, known := colorTerms[segment]
		if !known {
			continue
		}
		recognised = true
		if inBand(hue, band) {
			return true, true
		}
	}
	return false, recognised
}

// A band may run past 360 so the reds stay one range rather than two.
func inBand(hue float64, band [2]float64) bool {
	return (hue >= band[0] && hue <= band[1]) ||
		(hue+360 >= band[0] && hue+360 <= band[1])
}

func hueSaturation(hex string) (hue float64, saturation float64, ok bool) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) != 6 {
		return 0, 0, false
	}
	channels := make([]float64, 3)
	for index := range channels {
		var value int
		for _, digit := range hex[index*2 : index*2+2] {
			position := strings.IndexRune("0123456789abcdef", unicodeLower(digit))
			if position < 0 {
				return 0, 0, false
			}
			value = value*16 + position
		}
		channels[index] = float64(value) / 255
	}
	red, green, blue := channels[0], channels[1], channels[2]
	high := math.Max(red, math.Max(green, blue))
	low := math.Min(red, math.Min(green, blue))
	span := high - low
	if span == 0 {
		return 0, 0, true
	}
	switch high {
	case red:
		hue = math.Mod((green-blue)/span, 6)
	case green:
		hue = (blue-red)/span + 2
	default:
		hue = (red-green)/span + 4
	}
	hue *= 60
	if hue < 0 {
		hue += 360
	}
	lightness := (high + low) / 2
	saturation = span / (1 - math.Abs(2*lightness-1))
	return hue, saturation, true
}

func unicodeLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
