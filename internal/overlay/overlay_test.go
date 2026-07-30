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
	doc, err := Build(p, "engineer", "codex", "acting")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format != Format || doc.SchemaVersion != SchemaVersion ||
		doc.Person != p.Name || doc.Role != "engineer" ||
		doc.Seat.Harness != "codex" || doc.Seat.Name == "" ||
		doc.Expression != "acting" || doc.FavoriteColor == "" {
		t.Fatalf("overlay identity is incomplete: %+v", doc)
	}
	if len(doc.Personalities) != len(p.Roles["engineer"].Personalities) {
		t.Fatalf("overlay personalities = %d", len(doc.Personalities))
	}
	for _, personality := range doc.Personalities {
		if personality.Emblem.Name == "" || personality.Motif == "" ||
			personality.Form.Silhouette == "" || personality.SoundMark.Timbre == "" {
			t.Fatalf("overlay personality is incomplete: %+v", personality)
		}
	}
}

func TestBuildRejectsUnknownSelectionFacts(t *testing.T) {
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	for name, selection := range map[string][3]string{
		"role":       {"missing", "codex", "acting"},
		"seat":       {"engineer", "missing", "acting"},
		"expression": {"engineer", "codex", "invented"},
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
	doc, err := Build(p, "community", "codex", "waiting-for-human")
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
	doc, err := Build(p, "ops", "claude", "blocked")
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
