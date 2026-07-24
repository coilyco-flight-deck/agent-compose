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
				Motif:       "sunbeam",
				Emblem:      Emblem{Name: "lantern", Emoji: "🏮", Glyph: "✦"},
				Form:        Form{Silhouette: "beacon", Geometry: "open-rays", Motion: "glowing"},
				SoundMark:   SoundMark{Timbre: "bell", Contour: "rising", Pulse: "triplet"},
				Inspiration: InspirationRef{ID: "bright-credit", Fit: "long personality fit stays out"},
			},
			"steady": {
				Skill:       "personality-steady",
				Color:       "#5fa87a",
				Motif:       "stone",
				Emblem:      Emblem{Name: "anchor", Emoji: "⚓", Glyph: "◆"},
				Form:        Form{Silhouette: "cairn", Geometry: "stacked-rounds", Motion: "settling"},
				SoundMark:   SoundMark{Timbre: "wood-block", Contour: "returning", Pulse: "steady-pair"},
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
		"emblem `lantern` / `🏮` / `✦`, motif `sunbeam`",
		"form `beacon` / `open-rays` / `glowing`",
		"sound mark `bell` / `rising` / `triplet`",
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

func TestRenderRoleTranscriptIncludesCompleteSelectedMetadata(t *testing.T) {
	t.Parallel()
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.RenderRoleTranscript("engineer", "#90a66a")
	if err != nil {
		t.Fatal(err)
	}
	role := p.Roles["engineer"]
	for _, want := range []string{
		"role metadata",
		"person: kai // provided by: person:kai",
		"role: engineer",
		"purpose: " + role.Purpose,
		"personalities: curious // grounded // meticulous",
		"melded color: #90a66a",
		"role inspiration fit: " + role.Inspiration.Fit,
		"briefing:",
		"seats:",
		"personality metadata",
		"additional linked metadata",
		"renderer expressions: " + strings.Join(ExpressionVocabulary(), " // "),
	} {
		if !strings.Contains(got, want) {
			t.Errorf("transcript missing %q:\n%s", want, got)
		}
	}
	for _, seat := range role.Seats {
		want := fmt.Sprintf("seat %s: %s // pronouns: %s", seat.Harness, seat.Name, seat.Pronouns)
		if !strings.Contains(got, want) {
			t.Errorf("transcript missing seat %q:\n%s", want, got)
		}
	}
	for _, name := range role.Personalities {
		binding := p.Personalities[name]
		for _, want := range []string{
			"personality: " + name,
			"skill: " + binding.Skill,
			"color: " + binding.Color,
			"motif: " + binding.Motif,
			"emblem: " + binding.Emblem.Name,
			"form: silhouette " + binding.Form.Silhouette,
			"sound mark: timbre " + binding.SoundMark.Timbre,
			"inspiration fit: " + binding.Inspiration.Fit,
		} {
			if !strings.Contains(got, want) {
				t.Errorf("transcript missing personality field %q:\n%s", want, got)
			}
		}
	}
	for _, id := range []string{"marian-croak", "helen-jackson", "mellody-hobson"} {
		inspiration := p.Inspirations[id]
		for _, want := range []string{
			"linked inspiration: " + inspiration.Name + " (" + id + ")",
			"achievement: " + inspiration.Achievement,
			"impact mode: " + inspiration.ImpactMode,
			"impact fit: " + inspiration.ImpactFit,
			"profile citation: " + inspiration.ProfileCitation,
			"appearance: " + inspiration.Appearance.Title + " (" + inspiration.Appearance.ID + ")",
			"appearance citations: " + strings.Join(inspiration.Appearance.Citations, " // "),
		} {
			if !strings.Contains(got, want) {
				t.Errorf("transcript missing inspiration field %q:\n%s", want, got)
			}
		}
		if count := strings.Count(got, "linked inspiration: "+inspiration.Name+" ("+id+")"); count != 1 {
			t.Errorf("linked inspiration %q rendered %d times:\n%s", id, count, got)
		}
	}
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if strings.HasPrefix(line, " ") {
			t.Fatalf("transcript line is not flush-left: %q", line)
		}
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
