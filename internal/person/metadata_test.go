package person

import (
	"fmt"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
)

func TestRenderRoleMetadataIncludesCompactSelectedFacts(t *testing.T) {
	p := &Person{
		Name: "fixture",
		Roles: map[string]Role{
			"builder": {
				Purpose:       "Build the fixture.",
				Personalities: []string{"bright", "steady"},
				Seats: []Seat{
					{Harness: "alpha", Name: "bright builder", Pronouns: "she"},
					{Harness: "beta", Name: "steady builder", Pronouns: "he"},
				},
				Inspiration: InspirationRef{ID: "builder-credit", Fit: "long role fit stays out"},
			},
		},
		Personalities: map[string]Personality{
			"bright": {
				Skill:       "personality-bright",
				Color:       "#d98e48",
				Inspiration: InspirationRef{ID: "bright-credit", Fit: "long personality fit stays out"},
			},
			"steady": {
				Skill:       "personality-steady",
				Color:       "#5fa87a",
				Inspiration: InspirationRef{ID: "steady-credit", Fit: "another long fit stays out"},
			},
			"inactive": {
				Skill: "personality-inactive",
				Color: "#7d9fd3",
			},
		},
		Inspirations: map[string]Inspiration{
			"builder-credit": fixtureInspiration("Builder Credit", "builder-impact", "Builder Talk"),
			"bright-credit":  fixtureInspiration("Bright Credit", "bright-impact", "Bright Talk"),
			"steady-credit":  fixtureInspiration("Steady Credit", "steady-impact", "Steady Talk"),
		},
	}

	got, err := p.RenderRoleMetadata("builder", "#90a66a")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## Active role metadata",
		"* Person: `fixture`",
		"* Role: `builder`",
		"* Purpose: Build the fixture.",
		"`bright`: skill `personality-bright`, favorite color `#d98e48`",
		"`steady`: skill `personality-steady`, favorite color `#5fa87a`",
		"* Melded favorite color: `#90a66a`",
		"`alpha`: `bright builder` (pronouns: `she`)",
		"`beta`: `steady builder` (pronouns: `he`)",
		"Role `builder`: `Builder Credit` (`builder-credit`), impact mode `builder-impact`",
		"Personality `bright`: `Bright Credit` (`bright-credit`), impact mode `bright-impact`",
		"appearance `Bright Talk` (`fixture-talk`) at Fixture Conference (2026, keynote)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("metadata missing %q:\n%s", want, got)
		}
	}
	for _, excluded := range []string{
		"personality-inactive",
		"long role fit stays out",
		"long personality fit stays out",
		"long achievement stays out",
		"long impact fit stays out",
		"long appearance summary stays out",
		"profile-citation-stays-out",
		"appearance-citation-stays-out",
	} {
		if strings.Contains(got, excluded) {
			t.Errorf("metadata contains excluded long-form field %q:\n%s", excluded, got)
		}
	}
}

func TestRenderRoleMetadataRejectsIncompleteRelationships(t *testing.T) {
	p := &Person{
		Name: "fixture",
		Roles: map[string]Role{
			"builder": {
				Personalities: []string{"missing"},
				Inspiration:   InspirationRef{ID: "missing"},
			},
		},
	}
	if _, err := p.RenderRoleMetadata("unknown", "#90a66a"); err == nil {
		t.Fatal("unknown role must fail")
	}
	if _, err := p.RenderRoleMetadata("builder", ""); err == nil {
		t.Fatal("missing melded color must fail")
	}
	if _, err := p.RenderRoleMetadata("builder", "#90a66a"); err == nil {
		t.Fatal("missing personality must fail")
	}
}

func TestEmbeddedRoleMetadataCarriesEverySeat(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, roleName := range p.RoleOrder {
		role := p.Roles[roleName]
		var colors []string
		for _, name := range role.Personalities {
			colors = append(colors, p.Personalities[name].Color)
		}
		favorite, err := color.Favorite(colors)
		if err != nil {
			t.Fatal(err)
		}
		metadata, err := p.RenderRoleMetadata(roleName, favorite)
		if err != nil {
			t.Fatal(err)
		}
		for _, seat := range role.Seats {
			want := fmt.Sprintf("`%s`: `%s` (pronouns: `%s`)",
				seat.Harness, seat.Name, seat.Pronouns)
			if !strings.Contains(metadata, want) {
				t.Errorf("role %q metadata missing seat %q:\n%s", roleName, want, metadata)
			}
		}
	}
}

func fixtureInspiration(name, impactMode, title string) Inspiration {
	return Inspiration{
		Name:            name,
		Achievement:     "long achievement stays out",
		ImpactMode:      impactMode,
		ImpactFit:       "long impact fit stays out",
		ProfileCitation: "profile-citation-stays-out",
		Appearance: Appearance{
			ID:        "fixture-talk",
			Title:     title,
			Event:     "Fixture Conference",
			Year:      "2026",
			Format:    "keynote",
			Summary:   "long appearance summary stays out",
			Citations: []string{"appearance-citation-stays-out"},
		},
	}
}
