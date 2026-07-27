package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/evaluation"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/personpolicy"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func TestEvaluationOutputUsesYAML(t *testing.T) {
	t.Parallel()
	pack, err := evaluation.Build("engineer", "codex")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := evaluationOutput(pack, "yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "format: agent-compose.evaluation-pack.v1\n") {
		t.Fatalf("evaluation output is not YAML:\n%s", raw)
	}
	if _, err := evaluationOutput(pack, "json"); err == nil {
		t.Fatal("legacy JSON evaluation output remains accepted")
	}
}

func TestDirectPersonSelectionInheritsExternalOnlyHost(t *testing.T) {
	dir := t.TempDir()
	personRoot, err := filepath.Abs(filepath.Join(
		"..", "..", "testdata", "contracts", "person-independent",
	))
	if err != nil {
		t.Fatal(err)
	}
	paths := cascade.Paths{
		Config: filepath.Join(dir, "agent-compose.yaml"),
		Home:   filepath.Join(dir, "home"),
	}
	config := "person_policy: " + personpolicy.ExternalOnly + "\n" +
		"person_source: " + personRoot + "\n"
	if err := os.WriteFile(paths.Config, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	p, external, err := loadSelectedPersonAt("", paths)
	if err != nil {
		t.Fatal(err)
	}
	if !external || p.Name != "workbench" {
		t.Fatalf("direct selection = external:%t person:%q", external, p.Name)
	}
}

func TestDirectPersonSelectionFailsClosedWithBrokenHostGuard(t *testing.T) {
	dir := t.TempDir()
	paths := cascade.Paths{
		Config: filepath.Join(dir, "agent-compose.yaml"),
		Home:   filepath.Join(dir, "home"),
	}
	if err := os.WriteFile(
		paths.Config,
		[]byte("person_policy: "+personpolicy.ExternalOnly+"\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadSelectedPersonAt("", paths); err == nil ||
		!strings.Contains(err.Error(), "requires person_source") {
		t.Fatalf("broken host guard error = %v", err)
	}
}

func TestProjectPersonPolicyRejectsEmbeddedBundle(t *testing.T) {
	result, err := compose.Run(
		filepath.Join("..", "..", "testdata", "contracts", "native.kdl"),
		t.TempDir(),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = validateProjectPersonPolicy(result.Bundle.Dir, compose.Options{
		PersonPolicy: personpolicy.ExternalOnly,
		PersonSource: "/person",
	})
	if err == nil || !strings.Contains(err.Error(), "person:kai") {
		t.Fatalf("embedded bundle policy error = %v", err)
	}
}

func TestProjectPersonPolicyAcceptsExternalBundle(t *testing.T) {
	result, err := compose.Run(
		filepath.Join("..", "..", "testdata", "contracts", "custom-person.kdl"),
		t.TempDir(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateProjectPersonPolicy(result.Bundle.Dir, compose.Options{
		PersonPolicy: personpolicy.ExternalOnly,
		PersonSource: "/person",
	}); err != nil {
		t.Fatalf("external bundle rejected: %v", err)
	}
}

func TestDispatchArgs(t *testing.T) {
	cases := map[string]struct{ in, want []string }{
		"canonical name untouched": {
			[]string{"/usr/local/bin/agent-compose", "describe", "x"},
			[]string{"/usr/local/bin/agent-compose", "describe", "x"},
		},
		"acompose injects compose": {
			[]string{"/opt/homebrew/bin/acompose"},
			[]string{"/opt/homebrew/bin/acompose", "compose"},
		},
		"acompose keeps trailing args": {
			[]string{"acompose", "--", "claude"},
			[]string{"acompose", "compose", "--", "claude"},
		},
		"windows exe suffix": {
			[]string{`C:\shims\acompose.exe`, "req.kdl"},
			[]string{`C:\shims\acompose.exe`, "compose", "req.kdl"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := dispatchArgs(tc.in); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestPrintSummaryUsesSlashSeparators(t *testing.T) {
	t.Parallel()
	p, err := person.Load()
	if err != nil {
		t.Fatal(err)
	}
	result := &compose.Result{
		Bundle: &bundle.Result{Key: "abc123", Dir: "/tmp/bundle", Reused: true},
		Resolution: &resolver.Resolution{
			Request: &schema.Request{
				Role:       "engineer",
				ModelClass: schema.ModelClassFrontier,
				Delivery:   "native-skills",
			},
			Personalities: []string{"curious", "grounded", "meticulous"},
			FavoriteColor: "#90a66a",
			SourceIDs:     []string{"person:kai", "aos-public"},
			Person:        p,
			Decisions: []resolver.Decision{
				{Outcome: resolver.OutcomeSelected},
				{Outcome: resolver.OutcomeExcluded},
				{Outcome: resolver.OutcomeShadowed},
				{Outcome: resolver.OutcomeDelivered},
			},
		},
	}

	var output strings.Builder
	if err := printSummary(&output, result, person.RoleTranscriptOptions{}); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	for _, want := range []string{
		"request: model class frontier // delivery native-skills",
		"person: kai // provided by: person:kai",
		"role: engineer",
		"personalities: curious // grounded // meticulous",
		"melded color: #90a66a",
		"personality: curious",
		"inspiration achievement:",
		"renderer expressions: available // listening // thinking",
		"decisions: 1 selected // 1 excluded // 1 shadowed // 1 delivered",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("summary missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "·") {
		t.Fatalf("summary retained middle-dot separator:\n%s", got)
	}
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if len(line) > 0 && line[0] == ' ' {
			t.Fatalf("summary line is not flush-left: %q", line)
		}
	}

	var colored strings.Builder
	if err := printSummary(&colored, result, person.RoleTranscriptOptions{
		Color: true, TrueColor: true,
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(colored.String(), "\x1b[38;2;144;166;106mbundle") {
		t.Fatalf("summary intro did not use melded truecolor:\n%q", colored.String())
	}
}

func TestPrintVerificationUsesBoundedCounts(t *testing.T) {
	t.Parallel()
	verification := &bundle.Verification{
		Files: 128,
		Identities: []bundle.Identity{
			{Source: "person:kai", Skill: "personality-curious"},
			{Source: "aos-public", Skill: "coding-go"},
		},
	}

	var output strings.Builder
	printVerification(&output, verification)
	if got, want := output.String(), "bundle verified: 2 skills // 128 files\n"; got != want {
		t.Fatalf("verification output = %q, want %q", got, want)
	}
}
