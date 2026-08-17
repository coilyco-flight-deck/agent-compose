package color

import (
	"math"
	"slices"
	"testing"
)

// Roles sharing two of three personalities land on top of each other under a
// per-group mean, which is the collision the roster-wide derivation breaks.
var overlappingGroups = [][]string{
	{"#d98e48", "#7d9fd3", "#8f8c47"}, // curious, meticulous, tenacious
	{"#e1514e", "#19a9b0", "#da6c74"}, // bold, diplomatic, decisive
	{"#7d9fd3", "#3dbfe2", "#5d8fd7"}, // meticulous, candid, skeptical
	{"#009a85", "#5fa87a", "#787bd2"}, // protective, grounded, reflective
	{"#b682ed", "#e882e1", "#4f9eb8"}, // imaginative, playful, editorial
	{"#b3be49", "#5fa87a", "#da6c74"}, // outward, grounded, decisive
	{"#4f9eb8", "#e996a2", "#ff9c79"}, // editorial, nurturing, warm
	{"#2ed1aa", "#7d9fd3", "#5d8fd7"}, // empirical, meticulous, skeptical
}

func uniformCentroids(t *testing.T, groups [][]string) []string {
	t.Helper()
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		hex, err := Favorite(group)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, hex)
	}
	return out
}

func TestFavoritesSeparateOverlappingGroupsBetterThanTheMean(t *testing.T) {
	before, err := MinSeparation(uniformCentroids(t, overlappingGroups))
	if err != nil {
		t.Fatal(err)
	}
	derived, err := Favorites(overlappingGroups)
	if err != nil {
		t.Fatal(err)
	}
	after, err := MinSeparation(derived)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("closest pair did not improve: mean %.4f, derived %.4f", before, after)
	}
	// The mean leaves these under 0.04 apart. Guard the gain rather than the
	// exact figure, which moves whenever the roster does.
	if after < 0.10 {
		t.Fatalf("closest derived pair %.4f is below the floor the roster asserts", after)
	}
}

func TestFavoritesAreDeterministic(t *testing.T) {
	first, err := Favorites(overlappingGroups)
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 3; attempt++ {
		again, err := Favorites(overlappingGroups)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(first, again) {
			t.Fatalf("attempt %d drifted: %v then %v", attempt, first, again)
		}
	}
}

func TestFavoritesStayLegibleAndNearTheirOwnBlend(t *testing.T) {
	derived, err := Favorites(overlappingGroups)
	if err != nil {
		t.Fatal(err)
	}
	if len(derived) != len(overlappingGroups) {
		t.Fatalf("derived %d colors for %d groups", len(derived), len(overlappingGroups))
	}
	for index, hex := range derived {
		if err := Legible(hex); err != nil {
			t.Fatalf("group %d: %v", index, err)
		}
		anchor, err := Favorite(overlappingGroups[index])
		if err != nil {
			t.Fatal(err)
		}
		anchorLab, _ := toOKLab(anchor)
		lab, _ := toOKLab(hex)
		// The weighting moves the anchor too, so allow for both stages.
		if drift := separation(lab, anchorLab); drift > driftCap*4 {
			t.Fatalf("group %d drifted %.4f from its own blend", index, drift)
		}
	}
}

func TestFavoritesCannotRescueIdenticalGroups(t *testing.T) {
	groups := [][]string{
		{"#d98e48", "#7d9fd3", "#8f8c47"},
		{"#d98e48", "#7d9fd3", "#8f8c47"},
		{"#009a85", "#5fa87a", "#787bd2"},
	}
	derived, err := Favorites(groups)
	if err != nil {
		t.Fatal(err)
	}
	first, _ := toOKLab(derived[0])
	second, _ := toOKLab(derived[1])
	apart := separation(first, second)
	if apart == 0 {
		t.Fatal("identical groups produced one indistinguishable color")
	}
	// Repair runs after the drift cap and can add a little, so assert the
	// property the roster needs: identical melds stay under its 0.08 floor.
	if apart >= 0.08 {
		t.Fatalf("identical groups separated by %.4f, which would slip past the roster floor", apart)
	}
}

func TestFavoritesMatchTheMeanForOneGroup(t *testing.T) {
	group := []string{"#d98e48", "#7d9fd3", "#8f8c47"}
	alone, err := Favorites([][]string{group})
	if err != nil {
		t.Fatal(err)
	}
	mean, err := Favorite(group)
	if err != nil {
		t.Fatal(err)
	}
	if alone[0] != mean {
		t.Fatalf("single group derived %s, want the plain mean %s", alone[0], mean)
	}
}

func TestFavoritesRejectsEmptyInput(t *testing.T) {
	if _, err := Favorites(nil); err == nil {
		t.Fatal("empty group list passed")
	}
	if _, err := Favorites([][]string{{}}); err == nil {
		t.Fatal("group with no components passed")
	}
}

func TestMinSeparationOfFewerThanTwoColors(t *testing.T) {
	got, err := MinSeparation([]string{"#d98e48"})
	if err != nil || !math.IsInf(got, 1) {
		t.Fatalf("MinSeparation of one color = %v, %v", got, err)
	}
}
