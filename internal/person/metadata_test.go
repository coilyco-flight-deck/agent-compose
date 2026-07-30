package person

import (
	"fmt"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
)

func TestRenderRoleMetadataIncludesCompleteSelectedFacts(t *testing.T) {
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
		"* Provider: `person:fixture`",
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
		"Role `builder`: `Builder Credit` (`builder-credit`)",
		"* Fit: long role fit stays out",
		"Personality `bright`: `Bright Credit` (`bright-credit`)",
		"* Fit: long personality fit stays out",
		"* Achievement: long achievement stays out",
		"* Impact mode: `bright-impact`",
		"* Impact fit: long impact fit stays out",
		"* Profile citation: `profile-citation-stays-out`",
		"* Renderer expressions: `available`, `listening`, `thinking`",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("metadata missing %q:\n%s", want, got)
		}
	}
	for _, excluded := range []string{
		"personality-inactive",
		"Bright Talk",
		"fixture-talk",
		"Fixture Conference",
		"keynote",
		"long appearance summary stays out",
		"appearance-citation-stays-out",
	} {
		if strings.Contains(got, excluded) {
			t.Errorf("metadata contains excluded field %q:\n%s", excluded, got)
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
	got, err := p.RenderRoleTranscript("engineer", "#90a66a", RoleTranscriptOptions{})
	if err != nil {
		t.Fatal(err)
	}
	role := p.Roles["engineer"]
	for _, want := range []string{
		"role metadata",
		"person: kai // provided by: person:kai",
		"role: engineer",
		"purpose: " + role.Purpose,
		"personalities: " + strings.Join(role.Personalities, " // "),
		"melded color: #90a66a",
		"role inspiration fit: " + role.Inspiration.Fit,
		"briefing:",
		"seats:",
		"personality metadata",
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
		inspiration := p.Inspirations[binding.Inspiration.ID]
		for _, want := range []string{
			"personality: " + name,
			"skill: " + binding.Skill,
			"color: " + binding.Color,
			"motif: " + binding.Motif,
			"emblem: " + binding.Emblem.Name,
			"form: silhouette " + binding.Form.Silhouette,
			"sound mark: timbre " + binding.SoundMark.Timbre,
			"inspiration fit: " + binding.Inspiration.Fit,
			"inspiration achievement: " + inspiration.Achievement,
			"inspiration impact mode: " + inspiration.ImpactMode,
			"inspiration impact fit: " + inspiration.ImpactFit,
			"inspiration profile citation: " + inspiration.ProfileCitation,
		} {
			if !strings.Contains(got, want) {
				t.Errorf("transcript missing personality field %q:\n%s", want, got)
			}
		}
	}
	roleInspiration := p.Inspirations[role.Inspiration.ID]
	for _, want := range []string{
		"role inspiration achievement: " + roleInspiration.Achievement,
		"role inspiration impact mode: " + roleInspiration.ImpactMode,
		"role inspiration impact fit: " + roleInspiration.ImpactFit,
		"role inspiration profile citation: " + roleInspiration.ProfileCitation,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("transcript missing role inspiration field %q:\n%s", want, got)
		}
	}
	for _, excluded := range []string{
		"additional linked metadata",
		"inspiration appearance:",
		"inspiration appearance event:",
		"inspiration appearance citations:",
	} {
		if strings.Contains(got, excluded) {
			t.Errorf("transcript retained excluded appearance key %q:\n%s", excluded, got)
		}
	}
	identityRefs := []InspirationRef{role.Inspiration}
	for _, name := range role.Personalities {
		identityRefs = append(identityRefs, p.Personalities[name].Inspiration)
	}
	for _, ref := range identityRefs {
		appearance := p.Inspirations[ref.ID].Appearance
		for _, excluded := range []string{
			appearance.Title,
			appearance.ID,
			appearance.Event,
			strings.Join(strings.Fields(appearance.Summary), " "),
			strings.Join(appearance.Citations, " // "),
		} {
			if strings.Contains(got, excluded) {
				t.Errorf("transcript retained excluded appearance field %q:\n%s", excluded, got)
			}
		}
	}
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if strings.HasPrefix(line, " ") {
			t.Fatalf("transcript line is not flush-left: %q", line)
		}
	}
}

func TestRenderRoleTranscriptUsesCanonicalColors(t *testing.T) {
	t.Parallel()
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	trueColor, err := p.RenderRoleTranscript("engineer", "#90a66a", RoleTranscriptOptions{
		Color: true, TrueColor: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"\x1b[38;2;144;166;106mrole metadata",
		"\x1b[38;2;217;142;72mpersonality: curious",
		"\x1b[38;2;95;168;122mpersonality: grounded",
		"\x1b[38;2;125;159;211mpersonality: meticulous",
	} {
		if !strings.Contains(trueColor, want) {
			t.Errorf("truecolor transcript missing %q:\n%q", want, trueColor)
		}
	}
	fallback, err := p.RenderRoleTranscript("engineer", "#90a66a", RoleTranscriptOptions{
		Color: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fallback, "\x1b[38;5;") {
		t.Fatalf("fallback transcript omitted 256-color ANSI:\n%q", fallback)
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
