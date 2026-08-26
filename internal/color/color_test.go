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

// Seven accents tinted into a near-black land inside the side-by-side JND of
// each other. Solving the set is what separates them. agentic-os#358
func TestBackgroundsSeparateTheWholeSet(t *testing.T) {
	accents := []string{"#9c8b31", "#009895", "#3fd7a9", "#e583f7", "#637ffd", "#e36966", "#f7ab5d"}
	solved, err := Backgrounds(append([]string(nil), accents...))
	if err != nil {
		t.Fatalf("solve: %v", err)
	}
	if len(solved) != len(accents) {
		t.Fatalf("solved %d of %d", len(solved), len(accents))
	}
	separation, err := MinSeparation(solved)
	if err != nil {
		t.Fatalf("measure: %v", err)
	}
	if separation < 0.030 {
		t.Fatalf("closest pair %.4f, want at least 0.030", separation)
	}
	// Every background has to stay dark and quiet, or it stops being one.
	for index, hex := range solved {
		lab, err := toOKLab(hex)
		if err != nil {
			t.Fatalf("parse %s: %v", hex, err)
		}
		if lab.L > 0.32 {
			t.Fatalf("%s is lightness %.3f, too bright to read against", hex, lab.L)
		}
		if lab.chroma() > 0.08 {
			t.Fatalf("%s is chroma %.3f, too saturated for a background", hex, lab.chroma())
		}
		if hex == accents[index] {
			t.Fatalf("role %d kept its accent as a background", index)
		}
	}
}

// A role that turns over should move its own background and nobody else's more
// than the solve requires, so the set stays recognisable between rosters.
func TestBackgroundsKeepHueOrderAndTurnEachRoleTheLeast(t *testing.T) {
	accents := []string{"#9c8b31", "#009895", "#3fd7a9", "#e583f7", "#637ffd", "#e36966", "#f7ab5d"}
	solved, err := Backgrounds(append([]string(nil), accents...))
	if err != nil {
		t.Fatalf("solve: %v", err)
	}
	worst, total := 0.0, 0.0
	for index, hex := range solved {
		accent, err := toOKLab(accents[index])
		if err != nil {
			t.Fatal(err)
		}
		background, err := toOKLab(hex)
		if err != nil {
			t.Fatal(err)
		}
		turn := math.Abs(angleBetween(
			math.Atan2(accent.B, accent.A),
			math.Atan2(background.B, background.A),
		)) * 180 / math.Pi
		total += turn
		if turn > worst {
			worst = turn
		}
	}
	// Two roles sit at nearly the same hue, so one of them has to move a long
	// way. Everything else should barely move.
	if mean := total / float64(len(accents)); mean > 20 {
		t.Fatalf("mean rotation %.1f degrees, want the solve to stay near each role's own hue", mean)
	}
	if worst > 45 {
		t.Fatalf("worst rotation %.1f degrees, want no role thrown across the wheel", worst)
	}
}

func angleBetween(first, second float64) float64 {
	delta := second - first
	for delta > math.Pi {
		delta -= 2 * math.Pi
	}
	for delta < -math.Pi {
		delta += 2 * math.Pi
	}
	return delta
}

func TestBackgroundsRefuseAnEmptySet(t *testing.T) {
	if _, err := Backgrounds(nil); err == nil {
		t.Fatal("an empty set should be refused")
	}
	if _, err := Backgrounds([]string{"not a color"}); err == nil {
		t.Fatal("an unparseable accent should be refused")
	}
}
