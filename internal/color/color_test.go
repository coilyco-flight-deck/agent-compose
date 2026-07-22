package color

import (
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
