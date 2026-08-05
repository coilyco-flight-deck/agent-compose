package repositoryplan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureDigest = "sha256:0123456789012345678901234567890123456789012345678901234567890123"
const fixtureRevision = "0123456789012345678901234567890123456789"

func TestLoadRejectsUnsafeYAMLAndInvalidProvenance(t *testing.T) {
	validInput := "  - identity: example/aosk\n    revision: " + fixtureRevision + "\n    policy:\n      path: .agents/roles.kdl\n      sha256: " + fixtureDigest + "\n"
	validSelection := "    - identity: example/context\n      path: /tmp/projects/example/context\n      source: example/aosk\n      scope: global\n      reason: selected globally by repository policy\n"
	for name, body := range map[string]string{
		"unknown field":         "format: agent-compose.repositories.v2\nprojects_root: /tmp/projects\ninputs:\n" + validInput + "roles:\n  engineer:\n" + validSelection + "residency: []\nextra: true\n",
		"duplicate key":         "format: agent-compose.repositories.v2\nformat: agent-compose.repositories.v2\nprojects_root: /tmp/projects\ninputs:\n" + validInput + "roles:\n  engineer:\n" + validSelection + "residency: []\n",
		"alias":                 "format: &format agent-compose.repositories.v2\nprojects_root: /tmp/projects\ninputs:\n" + validInput + "roles:\n  engineer:\n" + validSelection + "residency: []\n",
		"custom tag":            "format: !custom agent-compose.repositories.v2\nprojects_root: /tmp/projects\ninputs:\n" + validInput + "roles:\n  engineer:\n" + validSelection + "residency: []\n",
		"ambiguous scalar":      "format: agent-compose.repositories.v2\nprojects_root: /tmp/projects\ninputs:\n" + validInput + "roles:\n  engineer:\n    - identity: example/context\n      path: /tmp/projects/example/context\n      source: example/aosk\n      scope: global\n      reason: true\nresidency: []\n",
		"unsafe path":           "format: agent-compose.repositories.v2\nprojects_root: /tmp/projects\ninputs:\n" + validInput + "roles:\n  engineer:\n    - identity: example/context\n      path: /tmp/other/example/context\n      source: example/aosk\n      scope: global\n      reason: selected\nresidency: []\n",
		"unsorted identities":   "format: agent-compose.repositories.v2\nprojects_root: /tmp/projects\ninputs:\n" + validInput + "roles:\n  engineer:\n    - identity: example/two\n      path: /tmp/projects/example/two\n      source: example/aosk\n      scope: global\n      reason: selected\n    - identity: example/one\n      path: /tmp/projects/example/one\n      source: example/aosk\n      scope: global\n      reason: selected\nresidency: []\n",
		"incomplete provenance": "format: agent-compose.repositories.v2\nprojects_root: /tmp/projects\ninputs:\n" + validInput + "roles:\n  engineer:\n    - identity: example/context\n      path: /tmp/projects/example/context\n      source: example/aosk\n      scope: provider\n      reason: selected\n      name: hardware\nresidency: []\n",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "repository-plan.yaml")
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil {
				t.Fatal("invalid repository plan passed")
			}
		})
	}
}

func TestMarshalProducesDeterministicYAML(t *testing.T) {
	plan := Plan{
		Format:       Format,
		ProjectsRoot: "/tmp/projects",
		Inputs: []Input{{
			Identity: "example/aosk",
			Revision: fixtureRevision,
			Policy:   PolicyInput{Path: PolicyPath, SHA256: fixtureDigest},
		}},
		Roles: map[string][]Selection{
			"engineer": {{
				Identity: "example/context",
				Path:     "/tmp/projects/example/context",
				Source:   "example/aosk",
				Scope:    "global",
				Reason:   "selected globally by repository policy",
			}},
		},
		Residency: []Selection{{
			Identity: "example/context",
			Path:     "/tmp/projects/example/context",
			Source:   "example/aosk",
			Scope:    "role-union",
			Reason:   "repository is selected by at least one canonical role",
		}},
	}
	first, err := Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("marshal changed between runs:\n%s\n---\n%s", first, second)
	}
	for _, want := range []string{
		"format: agent-compose.repositories.v2",
		"inputs:",
		"revision: \"" + fixtureRevision + "\"",
		"sha256: " + fixtureDigest,
		"roles:",
		"residency:",
	} {
		if !strings.Contains(string(first), want) {
			t.Fatalf("rendered plan missing %q:\n%s", want, first)
		}
	}
}
