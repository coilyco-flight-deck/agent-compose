package overlay

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/coilyco-flight-deck/agent-compose/v2/internal/person"
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
	if doc.RoleDisplayName != "Platform Engineer" {
		t.Errorf("role display name = %q, want Platform Engineer", doc.RoleDisplayName)
	}
	want := person.SeatAnnotation(doc.Seat.Name, doc.Seat.Pronouns, "Platform Engineer")
	if doc.Annotation != want || !strings.HasSuffix(doc.Annotation, "] (Platform Engineer)") {
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
	doc, err := Build(p, "director", "claude", "acting")
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
	doc, err := Build(p, "advocate", "codex", "waiting-for-human")
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

// Pins agent-compose#7362: these reached consumers only inside the Identity
// sentence, so aosx record.py parsed prose. The sentence still carries them.
func TestBuildProjectsElementAndCreatureAsFields(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"platform", "sysadmin", "science"} {
		doc, err := Build(p, role, "claude", "acting")
		if err != nil {
			t.Fatal(err)
		}
		want := p.Roles[role]
		if doc.Element != want.Element || doc.Element == "" {
			t.Fatalf("role %q element field = %q, role carries %q", role, doc.Element, want.Element)
		}
		if doc.Creature != want.Creature || doc.Creature == "" {
			t.Fatalf("role %q creature field = %q, role carries %q", role, doc.Creature, want.Creature)
		}
		// The regression this replaces: lineage read back out of prose.
		if !strings.Contains(doc.Identity, want.Creature) {
			t.Fatalf("role %q identity dropped the creature: %q", role, doc.Identity)
		}
	}
}

// The fields must survive the JSON `overlay --json` emits, which is the
// surface the consumer actually reads.
func TestMarshalCarriesElementAndCreature(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(p, "sysadmin", "claude", "acting")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"element", "creature"} {
		value, present := decoded[key]
		if !present {
			t.Fatalf("overlay JSON has no %q key: %v", key, keysOf(decoded))
		}
		if text, _ := value.(string); text == "" {
			t.Fatalf("overlay JSON %q is empty", key)
		}
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
