package skillselector

import (
	"slices"
	"strings"
	"testing"
)

func TestSelectExactAndGlobInCatalogueOrder(t *testing.T) {
	catalogue := []string{"compute-stack", "home-power-strip", "machine-alpha", "machine-beta"}
	selected, excluded, err := Select([]string{"compute-stack", "machine-*"}, catalogue)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(selected, []string{"compute-stack", "machine-alpha", "machine-beta"}) {
		t.Fatalf("selected = %v", selected)
	}
	if !slices.Equal(excluded, []string{"home-power-strip"}) {
		t.Fatalf("excluded = %v", excluded)
	}

	all, excluded, err := Select(nil, catalogue)
	if err != nil || !slices.Equal(all, catalogue) || len(excluded) != 0 {
		t.Fatalf("omitted selector changed catalogue: selected=%v excluded=%v err=%v", all, excluded, err)
	}
}

func TestSelectFailsClosed(t *testing.T) {
	catalogue := []string{"compute-stack", "machine-alpha"}
	for name, patterns := range map[string][]string{
		"explicit empty": {},
		"empty pattern":  {""},
		"invalid glob":   {"["},
		"unmatched":      {"machine-z*"},
		"overlap":        {"machine-*", "*-alpha"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := Select(patterns, catalogue); err == nil {
				t.Fatal("invalid selector passed")
			}
		})
	}

	_, _, err := Select([]string{"machine-*", "*-alpha"}, catalogue)
	if err == nil || !strings.Contains(err.Error(), "overlap on skill \"machine-alpha\"") {
		t.Fatalf("overlap error = %v", err)
	}
}
