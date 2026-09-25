package main

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestStartupRunWarnsAndContinuesOnABehaviorStep(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	run := &startupRun{w: &out}
	if err := run.do(stepRoleComposition, func() error { return errors.New("compose native role\nsecond line") }); err != nil {
		t.Fatalf("a warn step should not refuse: %v", err)
	}
	want := []degradedStep{{Step: "role-composition", Reason: "compose native role"}}
	if !reflect.DeepEqual(run.degraded, want) {
		t.Fatalf("degraded = %#v, want %#v", run.degraded, want)
	}
	if !strings.Contains(out.String(), "role-composition did not load") {
		t.Fatalf("the warning should name the step: %q", out.String())
	}
}

func TestStartupRunRefusesOnAReachStep(t *testing.T) {
	t.Parallel()
	run := &startupRun{w: &bytes.Buffer{}}
	err := run.do(stepMCPScope, func() error { return errors.New("no inventory") })
	if err == nil || !strings.Contains(err.Error(), "mcp-scope refused the launch") {
		t.Fatalf("a refuse step should stop the launch, got %v", err)
	}
	if len(run.degraded) != 0 {
		t.Fatalf("a refusal is not a degradation: %#v", run.degraded)
	}
}

func TestDegradedNoteReachesEachHarnessThatTakesOne(t *testing.T) {
	t.Parallel()
	degraded := []degradedStep{{Step: "person", Reason: `bad "quote"`}}
	claude := degradedHarnessArgs("claude", nil, degraded)
	if len(claude) != 2 || claude[0] != "--append-system-prompt" || !strings.Contains(claude[1], "- person: ") {
		t.Fatalf("claude args = %#v", claude)
	}
	if got := degradedHarnessArgs("claude", []string{"--append-system-prompt", "mine"}, degraded); got != nil {
		t.Fatalf("a caller's own system prompt should win, got %#v", got)
	}
	codex := degradedHarnessArgs("codex", nil, degraded)
	if len(codex) != 2 || !strings.HasPrefix(codex[1], `developer_instructions="`) || strings.Contains(codex[1], "\n") {
		t.Fatalf("codex args should be one quoted TOML string, got %#v", codex)
	}
	if got := degradedHarnessArgs("goose", nil, degraded); got != nil {
		t.Fatalf("goose takes no note, got %#v", got)
	}
	if got := degradedHarnessArgs("claude", nil, nil); got != nil {
		t.Fatalf("a whole launch adds nothing, got %#v", got)
	}
}

func TestDegradedOSCNamesTheSteps(t *testing.T) {
	t.Parallel()
	if got := degradedOSC([]string{"person", "card"}); got != "\x1b]7750;agent-compose;degraded=person,card\x07" {
		t.Fatalf("osc = %q", got)
	}
}

// The doc is where a reader learns the policy, so it names every step and its
// class exactly as the code does.
func TestStartupPolicyDocClassesEveryStep(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../docs/launch.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	for _, step := range startupSteps {
		if !strings.Contains(doc, "* `"+step.Name+"` - "+string(step.Class)+" - ") {
			t.Errorf("docs/launch.md does not class %s as %s", step.Name, step.Class)
		}
	}
}
