package overlay

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
)

func TestBuildProjectsOneCanonicalMember(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(p, "platform", "codex", "acting")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != Format || doc.SchemaVersion != SchemaVersion ||
		doc.Person != p.Name || doc.Role != "platform" ||
		doc.Seat.Harness != "codex" || doc.Seat.Name == "" ||
		doc.Expression != "acting" || doc.FavoriteColor == "" {
		t.Fatalf("overlay identity is incomplete: %+v", doc)
	}
	if len(doc.Personalities) != len(p.Roles["platform"].Personalities) {
		t.Fatalf("overlay personalities = %d", len(doc.Personalities))
	}
	if doc.Stance == "" {
		t.Fatalf("overlay carries no role stance: %+v", doc)
	}
	for _, personality := range doc.Personalities {
		if personality.Emblem.Name() == "" || personality.Motif == "" ||
			personality.Geometry == "" || personality.Body.Archetype == "" ||
			personality.Body.Attachment == "" || personality.SoundMark.Timbre == "" {
			t.Fatalf("overlay personality is incomplete: %+v", personality)
		}
	}
}

// TestBuildComposesTheSeatAnnotation pins the identity string every terminal
// renderer shows verbatim, so a consumer never reassembles it from parts.
func TestBuildComposesTheSeatAnnotation(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(p, "platform", "claude", "acting")
	if err != nil {
		t.Fatal(err)
	}
	if doc.RoleDisplayName != "Agentic Platform Engineer" {
		t.Errorf("role display name = %q, want Agentic Platform Engineer", doc.RoleDisplayName)
	}
	want := person.SeatAnnotation(doc.Seat.Name, doc.Seat.Pronouns, "Agentic Platform Engineer")
	if doc.Annotation != want || !strings.HasSuffix(doc.Annotation, "] (Agentic Platform Engineer)") {
		t.Errorf("annotation = %q, want %q", doc.Annotation, want)
	}
}

// TestRenderTextCarriesPronounsNotTheRoleLabel guards the card's split: it
// already prints the role on its own, so the seat stops at the pronouns.
func TestRenderTextCarriesPronounsNotTheRoleLabel(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(p, "tpm", "claude", "acting")
	if err != nil {
		t.Fatal(err)
	}
	wide, err := RenderText(doc, 200)
	if err != nil {
		t.Fatal(err)
	}
	label := person.SeatLabel(doc.Seat.Name, doc.Seat.Pronouns)
	if !strings.Contains(wide, label) {
		t.Errorf("card %q omits seat label %q", wide, label)
	}
	if strings.Contains(wide, "(Portfolio Director)") {
		t.Errorf("card %q repeats the role label", wide)
	}
}

func TestBuildRejectsUnknownSelectionFacts(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	for name, selection := range map[string][3]string{
		"role":       {"missing", "codex", "acting"},
		"seat":       {"platform", "missing", "acting"},
		"expression": {"platform", "codex", "invented"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Build(p, selection[0], selection[1], selection[2]); err == nil {
				t.Fatal("unknown overlay selection must fail")
			}
		})
	}
}

func TestRenderTextIsWidthResponsiveAndPlain(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(p, "devrel", "codex", "waiting-for-human")
	if err != nil {
		t.Fatal(err)
	}
	narrow, err := RenderText(doc, 40)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSuffix(narrow, "\n"), "\n") {
		if utf8.RuneCountInString(line) > 40 {
			t.Fatalf("narrow line exceeds width: %q", line)
		}
	}
	wide, err := RenderText(doc, 200)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(wide, "\n") != 1 || strings.Contains(wide, "\x1b") {
		t.Fatalf("wide overlay is not one plain line: %q", wide)
	}
}

func TestMarshalProducesVersionedJSON(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(p, "sysadmin", "claude", "blocked")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Document
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Format != Format || decoded.Expression != "blocked" {
		t.Fatalf("unexpected overlay JSON: %+v", decoded)
	}
}
