package color

import (
	"math"
	"strings"
	"testing"
)

var palette = []string{"#d98e48", "#5fa87a", "#7d9fd3"}

func TestLegibleAcceptsPaletteAndRejectsExtremes(t *testing.T) {
	for _, hex := range palette {
		if err := Legible(hex); err != nil {
			t.Fatalf("palette color %s must be legible: %v", hex, err)
		}
	}
	for _, hex := range []string{"#000000", "#ffffff", "#101010", "#f8f8f8", "#9a9a9a"} {
		if err := Legible(hex); err == nil {
			t.Fatalf("%s must fail the legibility gate", hex)
		}
	}
	if err := Legible("not-a-color"); err == nil {
		t.Fatal("malformed hex must fail")
	}
}

func TestFavoriteOfOneIsItself(t *testing.T) {
	got, err := Favorite([]string{"#d98e48"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "#d98e48" {
		t.Fatalf("single-component favorite must be the component, got %s", got)
	}
}

func TestFavoriteBlendsLegiblyAndDeterministically(t *testing.T) {
	first, err := Favorite(palette)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Favorite(palette)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("favorite must be deterministic: %s vs %s", first, second)
	}
	if err := Legible(first); err != nil {
		t.Fatalf("derived favorite %s must stay in the legible band: %v", first, err)
	}
	for _, component := range palette {
		if first == component {
			t.Fatalf("three-way blend should not collapse onto component %s", component)
		}
	}
}

func TestFavoriteResistsGrayingOut(t *testing.T) {
	got, err := Favorite([]string{"#d98e48", "#7d9fd3"})
	if err != nil {
		t.Fatal(err)
	}
	lab, err := toOKLab(got)
	if err != nil {
		t.Fatal(err)
	}
	if lab.chroma() < minChroma {
		t.Fatalf("complementary blend %s grayed out (chroma %.3f)", got, lab.chroma())
	}
}

func TestANSIRendering(t *testing.T) {
	if got := ANSI("#d98e48", "x", true); !strings.Contains(got, "38;2;217;142;72") {
		t.Fatalf("unexpected truecolor sequence: %q", got)
	}
	if got := ANSI("#d98e48", "x", false); !strings.Contains(got, "38;5;") {
		t.Fatalf("unexpected 256-color sequence: %q", got)
	}
	if got := ANSI("bogus", "x", true); got != "x" {
		t.Fatalf("bad hex must degrade to plain text, got %q", got)
	}
}

func TestShimmerLightensWithoutChangingHue(t *testing.T) {
	base := "#ac8fd7"
	shimmer, err := Shimmer(base)
	if err != nil {
		t.Fatalf("shimmer: %v", err)
	}
	baseLab, _ := toOKLab(base)
	shimmerLab, _ := toOKLab(shimmer)
	if shimmerLab.L <= baseLab.L {
		t.Errorf("shimmer lightness %.3f is not above base %.3f", shimmerLab.L, baseLab.L)
	}
	baseHue := math.Atan2(baseLab.B, baseLab.A)
	shimmerHue := math.Atan2(shimmerLab.B, shimmerLab.A)
	if math.Abs(baseHue-shimmerHue) > 0.01 {
		t.Errorf("shimmer shifted hue from %.3f to %.3f", baseHue, shimmerHue)
	}
}

func TestShimmerRejectsBadHex(t *testing.T) {
	if _, err := Shimmer("purple"); err == nil {
		t.Fatal("expected an error for a non-hex color")
	}
}

func TestNearestPicksTheClosestHue(t *testing.T) {
	slots := []string{"#dc2626", "#6a9bcc", "#16a34a", "#827dbd"}
	got, err := Nearest("#ac8fd7", slots)
	if err != nil {
		t.Fatalf("nearest: %v", err)
	}
	if got != "#827dbd" {
		t.Errorf("nearest slot is %q, want the purple slot", got)
	}
}

func TestNearestNeedsCandidates(t *testing.T) {
	if _, err := Nearest("#ac8fd7", nil); err == nil {
		t.Fatal("expected an error with no candidates")
	}
}
