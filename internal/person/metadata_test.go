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
	got, err := p.RenderRoleTranscript("platform", "#90a66a", RoleTranscriptOptions{Expanded: true})
	if err != nil {
		t.Fatal(err)
	}
	role := p.Roles["platform"]
	personalityMetadata := strings.Index(got, "personality metadata")
	rendererExpressions := strings.Index(got, "renderer expressions:")
	roleMetadata := strings.Index(got, "role metadata")
	if personalityMetadata < 0 ||
		rendererExpressions < personalityMetadata ||
		roleMetadata < rendererExpressions {
		t.Fatalf("transcript does not end with role metadata:\n%s", got)
	}
	for _, want := range []string{
		"role metadata",
		"roster: core // provided by: roster:core",
		"role: platform",
		"purpose: " + role.Purpose,
		"agent identity: " + role.Identity.Name + " // pronouns: " + role.Identity.Pronouns,
		"personalities: " + strings.Join(role.Personalities, " // "),
		"melded color: #90a66a",
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
		want := fmt.Sprintf("seat %s%s", seat.Selector(), seatRoutingSuffix(seat))
		if !strings.Contains(got, want) {
			t.Errorf("transcript missing seat %q:\n%s", want, got)
		}
	}
	for _, name := range role.Personalities {
		binding := p.Personalities[name]
		for _, want := range []string{
			binding.Emblem.Emoji + " " + binding.Emblem.Name + " " + binding.Emblem.Glyph +
				" // " + binding.Motif + " // " + binding.Color,
			"// form: " + binding.Form.Silhouette + ", " + binding.Form.Geometry +
				", " + binding.Form.Motion,
			"// sound: " + binding.SoundMark.Timbre + ", " + binding.SoundMark.Contour +
				", " + binding.SoundMark.Pulse,
			"// skill: " + binding.Skill,
		} {
			if !strings.Contains(got, want) {
				t.Errorf("transcript missing personality field %q:\n%s", want, got)
			}
		}
	}
	for _, excluded := range []string{
		"additional linked metadata",
		"inspiration:",
		"inspiration fit:",
		"inspiration achievement:",
		"inspiration profile citation:",
		"inspiration appearance:",
		"inspiration appearance event:",
		"inspiration appearance citations:",
	} {
		if strings.Contains(got, excluded) {
			t.Errorf("transcript retained excluded appearance key %q:\n%s", excluded, got)
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
	role := p.Roles["platform"]
	favoriteColor := role.FavoriteColor
	trueColor, err := p.RenderRoleTranscript("platform", favoriteColor, RoleTranscriptOptions{
		Color: true, TrueColor: true, Expanded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	wants := []string{strings.TrimSuffix(color.ANSI(favoriteColor, "role metadata", true), "\x1b[0m")}
	for _, personalityName := range role.Personalities {
		wants = append(wants, strings.TrimSuffix(color.ANSI(
			p.Personalities[personalityName].Color,
			"personality: "+personalityName,
			true,
		), "\x1b[0m"))
	}
	for _, want := range wants {
		if !strings.Contains(trueColor, want) {
			t.Errorf("truecolor transcript missing %q:\n%q", want, trueColor)
		}
	}
	fallback, err := p.RenderRoleTranscript("platform", favoriteColor, RoleTranscriptOptions{
		Color: true, Expanded: true,
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
		favorite := role.FavoriteColor
		metadata, err := p.RenderRoleMetadata(roleName, favorite)
		if err != nil {
			t.Fatal(err)
		}
		identity := fmt.Sprintf(
			"* Agent identity: `%s` (pronouns: `%s`)",
			role.Identity.Name,
			role.Identity.Pronouns,
		)
		if strings.Count(metadata, identity) != 1 {
			t.Errorf("role %q metadata identity count = %d:\n%s", roleName, strings.Count(metadata, identity), metadata)
		}
		for _, seat := range role.Seats {
			want := fmt.Sprintf("`%s`%s", seat.Selector(), seatRoutingSuffix(seat))
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

// The default is what a consumer reads on every launch, so the identity texture
// rides --explain and the meld is still named by the role block. See #323.
func TestRenderRoleTranscriptKeepsTheDefaultTerse(t *testing.T) {
	t.Parallel()
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	terse, err := p.RenderRoleTranscript("platform", "#90a66a", RoleTranscriptOptions{})
	if err != nil {
		t.Fatal(err)
	}
	role := p.Roles["platform"]
	for _, unwanted := range []string{
		"personality metadata",
		"renderer expressions:",
		p.Personalities[role.Personalities[0]].Motif,
		p.Personalities[role.Personalities[0]].Emblem.Name,
		p.Personalities[role.Personalities[0]].SoundMark.Timbre,
	} {
		if strings.Contains(terse, unwanted) {
			t.Errorf("terse transcript still carries %q:\n%s", unwanted, terse)
		}
	}
	for _, want := range []string{
		"role metadata",
		"role: platform",
		"personalities: " + strings.Join(role.Personalities, " // "),
	} {
		if !strings.Contains(terse, want) {
			t.Errorf("terse transcript missing %q:\n%s", want, terse)
		}
	}
	expanded, err := p.RenderRoleTranscript("platform", "#90a66a", RoleTranscriptOptions{Expanded: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(expanded) <= len(terse) {
		t.Errorf("expanded transcript (%d) is not longer than terse (%d)", len(expanded), len(terse))
	}
}
