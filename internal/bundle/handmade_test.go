package bundle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Hand-authored, never composed. A failure here means the bundle protocol
// stopped being producer-agnostic. See docs/bundle-protocol.md.
const handmadeFixture = "../../testdata/handmade-bundle"

func TestVerifyAcceptsHandAuthoredBundle(t *testing.T) {
	verification, err := Verify(filepath.FromSlash(handmadeFixture))
	if err != nil {
		t.Fatalf("verify hand-authored bundle: %v", err)
	}
	if got := len(verification.Identities); got != 2 {
		t.Fatalf("hand-authored bundle selects %d identities, want 2", got)
	}
	if verification.Manifest.Role != "courier" {
		t.Fatalf("hand-authored bundle role is %q, want courier", verification.Manifest.Role)
	}
	for _, identity := range verification.Identities {
		if identity.Source != "fixture:handmade" {
			t.Fatalf("identity %q came from %q, want fixture:handmade", identity.Skill, identity.Source)
		}
	}
}

// The provider budget is the one field a foreign producer cannot guess, so it
// earns the negative control beside the positive case above.
func TestVerifyRejectsWrongProviderBudget(t *testing.T) {
	dir := copyHandmade(t)
	tracePath := filepath.Join(dir, "trace.json")
	var trace Trace
	raw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read copied trace: %v", err)
	}
	if err := json.Unmarshal(raw, &trace); err != nil {
		t.Fatalf("parse copied trace: %v", err)
	}
	trace.Providers[0].ContextBytes++
	trace.Providers[0].ApproximateTokens = (trace.Providers[0].ContextBytes + 3) / 4
	rewritten, err := json.Marshal(&trace)
	if err != nil {
		t.Fatalf("encode mutated trace: %v", err)
	}
	if err := os.WriteFile(tracePath, rewritten, 0o644); err != nil {
		t.Fatalf("write mutated trace: %v", err)
	}
	_, err = Verify(dir)
	if err == nil {
		t.Fatal("verify accepted a provider budget that disagrees with the skill tree")
	}
	if !strings.Contains(err.Error(), "budget") {
		t.Fatalf("verify rejected for the wrong reason: %v", err)
	}
}

func copyHandmade(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "bundle")
	if err := os.CopyFS(dir, os.DirFS(filepath.FromSlash(handmadeFixture))); err != nil {
		t.Fatalf("copy hand-authored fixture: %v", err)
	}
	return dir
}
