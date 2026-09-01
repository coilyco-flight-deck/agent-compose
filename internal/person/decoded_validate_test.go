package person

import (
	"strings"
	"testing"
	"testing/fstest"
)

// These replace the parse-time table tests the KDL walker carried. Every case
// names the rule it protects, so a deleted rule fails a named test. #408.

func validRole() Role {
	return Role{
		Purpose:       "Do the bounded thing.",
		Skill:         "role-reader",
		Personalities: []string{"bright"},
	}
}

func validPersonality() Personality {
	return Personality{
		Skill: "personality-bright", Color: "#d98e48",
		Motif: "sunbeam", Geometry: "open-rays",
		Emblem:    Emblem{Names: []string{"lantern"}, Emoji: "🏮"},
		Body:      Body{Archetype: "upright", Attachment: "a lantern"},
		SoundMark: SoundMark{Timbre: "bell", Contour: "rising", Pulse: "triplet"},
	}
}

func validBoundary() Boundary {
	return Boundary{
		Skill: "boundary-shared-thing", Owner: "sysadmin",
		Summary: "Systems Administrator owns it, other roles hand it over",
	}
}

func TestValidateDecodedRoleRules(t *testing.T) {
	cases := []struct {
		name string
		want string
		edit func(*Role)
	}{
		{"briefing and skill", "cannot define both briefing and skill",
			func(r *Role) { r.Briefing = "prose" }},
		{"unknown element", "unknown element",
			func(r *Role) { r.Element = "plasma" }},
		{"no personality", "needs at least one personality",
			func(r *Role) { r.Personalities = nil }},
		{"repeated personality", "repeats personality",
			func(r *Role) { r.Personalities = []string{"bright", "bright"} }},
		{"repeated boundary", "repeats boundary",
			func(r *Role) { r.Boundaries = []string{"shared-thing", "shared-thing"} }},
		{"unstable boundary id", "stable doctrine id",
			func(r *Role) { r.Boundaries = []string{"Shared Thing"} }},
		{"repeated scoped boundary", "repeats scoped boundary",
			func(r *Role) {
				r.ScopedBoundaries = []ScopedBoundary{
					{Name: "shared-thing", Scope: "here"}, {Name: "shared-thing", Scope: "there"},
				}
			}},
		{"defers and scopes", "both defers and scopes",
			func(r *Role) {
				r.Boundaries = []string{"shared-thing"}
				r.ScopedBoundaries = []ScopedBoundary{{Name: "shared-thing", Scope: "here"}}
			}},
		{"self adjacent", "cannot name itself",
			func(r *Role) { r.Adjacents = []Adjacent{{Role: "reader", Reason: "why"}} }},
		{"repeated adjacent", "repeats adjacent",
			func(r *Role) {
				r.Adjacents = []Adjacent{{Role: "writer", Reason: "a"}, {Role: "writer", Reason: "b"}}
			}},
		{"unsupported tier", "unsupported model tier",
			func(r *Role) { r.SupportedModelTiers = []string{"artisanal"} }},
		{"repeated tier", "repeats model tier",
			func(r *Role) { r.SupportedModelTiers = []string{"frontier", "frontier"} }},
		{"method duplicates role skill", "duplicates its role skill",
			func(r *Role) { r.Methods = []string{"role-reader"} }},
		{"identity without name", "identity needs a name",
			func(r *Role) { r.Identity = &AgentIdentity{Pronouns: "they"} }},
		{"identity without pronouns", "identity needs a pronouns",
			func(r *Role) { r.Identity = &AgentIdentity{Name: "Reed"} }},
		{"duplicate seat", "duplicate seat",
			func(r *Role) {
				r.Seats = []Seat{
					{Key: "codex", Harness: "codex", Name: "one"},
					{Key: "codex", Harness: "codex", Name: "two"},
				}
			}},
		{"seat without a name", "needs a name property",
			func(r *Role) { r.Seats = []Seat{{Key: "codex", Harness: "codex"}} }},
		{"seat redefines role identity", "cannot redefine role identity",
			func(r *Role) {
				r.Identity = &AgentIdentity{Name: "Reed", Pronouns: "they"}
				r.Seats = []Seat{{Key: "codex", Harness: "codex", Name: "Other"}}
			}},
		{"seat tier outside the role set", "outside the role compatibility set",
			func(r *Role) {
				r.SupportedModelTiers = []string{"frontier"}
				r.Seats = []Seat{{Key: "codex", Harness: "codex", Name: "Reed", Tier: "commodity"}}
			}},
		{"copy-contract scope", "supported scope tool-response",
			func(r *Role) { r.CopyContract = &CopyContract{Scope: "prose"} }},
		{"copy-contract without rules", "needs forbid rules",
			func(r *Role) { r.CopyContract = &CopyContract{Scope: "tool-response"} }},
		{"copy-contract without prefer", "needs prefer",
			func(r *Role) {
				r.CopyContract = &CopyContract{
					Scope: "tool-response", Rules: []CopyRule{{Forbid: "upload"}},
				}
			}},
		{"copy-contract repeats forbid", "repeats forbid",
			func(r *Role) {
				r.CopyContract = &CopyContract{Scope: "tool-response", Rules: []CopyRule{
					{Forbid: "upload", Prefer: "add"}, {Forbid: "upload", Prefer: "attach"},
				}}
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			role := validRole()
			tc.edit(&role)
			err := validateDecodedRole("reader", &role)
			if err == nil {
				t.Fatalf("%s must fail", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
	valid := validRole()
	if err := validateDecodedRole("reader", &valid); err != nil {
		t.Fatalf("valid role rejected: %v", err)
	}
}

func TestValidateDecodedPersonalityRules(t *testing.T) {
	cases := []struct {
		name string
		want string
		edit func(*Personality)
	}{
		{"missing skill", "needs a skill", func(p *Personality) { p.Skill = "" }},
		{"missing color", "needs a color", func(p *Personality) { p.Color = "" }},
		{"illegible color", "personality", func(p *Personality) { p.Color = "#fffffe" }},
		{"unsemantic motif", "semantic motif", func(p *Personality) { p.Motif = "Sun Beam" }},
		{"unsemantic geometry", "semantic geometry", func(p *Personality) { p.Geometry = "Open Rays" }},
		{"missing emblem", "needs an emblem", func(p *Personality) { p.Emblem = Emblem{} }},
		{"missing body", "needs a body", func(p *Personality) { p.Body = Body{} }},
		{"missing sound-mark", "needs a sound-mark", func(p *Personality) { p.SoundMark = SoundMark{} }},
		{"empty verb", "empty verb", func(p *Personality) { p.Verbs = []string{" "} }},
		{"repeated verb", "repeats verb", func(p *Personality) { p.Verbs = []string{"Measuring", "Measuring"} }},
		{"repeated alias", "repeats alias", func(p *Personality) { p.Aliases = []string{"careful", "careful"} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			personality := validPersonality()
			tc.edit(&personality)
			err := validateDecodedPersonality("bright", personality)
			if err == nil {
				t.Fatalf("%s must fail", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
	if err := validateDecodedPersonality("bright", validPersonality()); err != nil {
		t.Fatalf("valid personality rejected: %v", err)
	}
}

func TestValidateDecodedBoundaryRules(t *testing.T) {
	cases := []struct {
		name string
		want string
		edit func(*Boundary)
	}{
		{"missing skill", "needs a skill", func(b *Boundary) { b.Skill = "" }},
		{"missing summary", "needs a summary", func(b *Boundary) { b.Summary = "" }},
		{"missing owner", "needs an owner", func(b *Boundary) { b.Owner = "" }},
		{"unstable owner", "stable role id", func(b *Boundary) { b.Owner = "Systems Admin" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			boundary := validBoundary()
			tc.edit(&boundary)
			err := validateDecodedBoundary("shared-thing", boundary)
			if err == nil {
				t.Fatalf("%s must fail", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
	if err := validateDecodedBoundary("shared-thing", validBoundary()); err != nil {
		t.Fatalf("valid boundary rejected: %v", err)
	}
}

// decodeTestPackage builds a person from in-memory fragments, which is the YAML
// equivalent of the KDL document the retired parse() took.
func decodeTestPackage(t *testing.T, files map[string]string) (*Person, error) {
	t.Helper()
	source := fstest.MapFS{}
	for name, body := range files {
		source[name] = &fstest.MapFile{Data: []byte(body), Mode: 0o644}
	}
	return decodePersonSource(source, "test package")
}

func TestDecodePreservesAuthoredRoleFields(t *testing.T) {
	p, err := decodeTestPackage(t, map[string]string{
		"person.yaml": "person: workbench\n",
		"roles/01-reader.yaml": "role: reader\n" +
			"purpose: Read the thing.\n" +
			"skill: role-reader\n" +
			"personalities: [bright]\n" +
			"identity:\n  name: Reed\n  pronouns: they\n" +
			"agents:\n  - harness: codex\n",
		"personalities/01-bright.yaml": "personality: bright\n" +
			"skill: personality-bright\ncolor: \"#d98e48\"\n" +
			"motif: sunbeam\ngeometry: open-rays\n" +
			"emblem:\n  names: [lantern]\n  emoji: 🏮\n" +
			"body:\n  archetype: upright\n  attachment: a lantern\n" +
			"sound_mark:\n  timbre: bell\n  contour: rising\n  pulse: triplet\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	role := p.Roles["reader"]
	if got := role.Personalities; len(got) != 1 || got[0] != "bright" {
		t.Fatalf("role personalities = %q", got)
	}
	if role.Identity == nil || role.Identity.Name != "Reed" || role.Identity.Pronouns != "they" {
		t.Fatalf("role identity = %+v", role.Identity)
	}
	if len(role.Seats) != 1 || role.Seats[0].Harness != "codex" {
		t.Fatalf("role seats = %+v", role.Seats)
	}
	if p.Name != "workbench" || p.ProviderKind != "person" {
		t.Fatalf("provider = %s %s", p.ProviderKind, p.Name)
	}
}

func TestDecodeRejectsUnknownKeyAndFilenameMismatch(t *testing.T) {
	base := map[string]string{
		"person.yaml": "person: workbench\n",
		"personalities/01-bright.yaml": "personality: bright\n" +
			"skill: personality-bright\ncolor: \"#d98e48\"\n" +
			"motif: sunbeam\ngeometry: open-rays\n" +
			"emblem:\n  names: [lantern]\n  emoji: 🏮\n" +
			"body:\n  archetype: upright\n  attachment: a lantern\n" +
			"sound_mark:\n  timbre: bell\n  contour: rising\n  pulse: triplet\n",
	}

	unknown := map[string]string{}
	for k, v := range base {
		unknown[k] = v
	}
	unknown["roles/01-reader.yaml"] = "role: reader\npurpose: Read.\nskill: role-reader\n" +
		"personalities: [bright]\nnotafield: yes\n"
	if _, err := decodeTestPackage(t, unknown); err == nil ||
		!strings.Contains(err.Error(), "field notafield not found") {
		t.Fatalf("unknown key error = %v", err)
	}

	mismatch := map[string]string{}
	for k, v := range base {
		mismatch[k] = v
	}
	mismatch["roles/01-reader.yaml"] = "role: writer\npurpose: Read.\nskill: role-writer\n" +
		"personalities: [bright]\n"
	if _, err := decodeTestPackage(t, mismatch); err == nil ||
		!strings.Contains(err.Error(), "filename does not match role") {
		t.Fatalf("filename mismatch error = %v", err)
	}
}
