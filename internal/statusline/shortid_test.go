package statusline

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/agentid"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
)

// shortIDFixture writes the same minimal projection the other tests use, so
// these assertions differ only by the session id in scope.
func shortIDFixture(t *testing.T) string {
	target, _ := shortIDFixtureDirs(t)
	return target
}

// shortIDBundleFixture is the same fixture addressed by its bundle, for a test
// that binds a session instead of walking to a projection.
func shortIDBundleFixture(t *testing.T) string {
	_, bundleDir := shortIDFixtureDirs(t)
	return bundleDir
}

func shortIDFixtureDirs(t *testing.T) (string, string) {
	t.Helper()
	target := t.TempDir()
	bundleDir := t.TempDir()
	writeJSON(t, filepath.Join(target, ".agent-compose", "projection.json"), project.Projection{
		Layout: "claude", Bundle: bundleDir, Files: []string{"CLAUDE.md"},
	})
	writeJSON(t, filepath.Join(bundleDir, "manifest.json"), bundle.Manifest{
		Format: "agent-compose.bundle", Role: "platform", ModelTier: "frontier",
		Color: "#b39258",
		Identity: bundle.RoleIdentity{
			Person: "core",
			Seats: []person.Seat{
				{Key: "claude", Name: "Angie", Pronouns: "she"},
			},
			Personalities: []bundle.IdentityPersonality{
				{Name: "tenacious", Color: "#d98e48", Emblem: person.Emblem{Emoji: "🧭"}},
			},
		},
	})
	writeJSON(t, filepath.Join(bundleDir, "trace.json"), bundle.Trace{
		Format: "agent-compose.trace",
		Providers: []resolver.ProviderReport{
			{Source: "aos", Outcome: resolver.OutcomeSelected, Skills: 4, ApproximateTokens: 900},
		},
	})
	return target, bundleDir
}

// The session row names the agent a human would speak to.
func TestSessionRowCarriesTheShortID(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "uz86")
	got, err := Render(Options{Target: shortIDFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Angie [she] uz86") {
		t.Errorf("statusline = %q, want the seat annotated with uz86", got)
	}
}

// Outside a native session the row is exactly what it was before. This is the
// property that lets the id ship without touching any existing surface.
func TestSessionRowWithoutASessionIsUnchanged(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "")
	got, err := Render(Options{Target: shortIDFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Angie [she] //") {
		t.Errorf("statusline = %q, want the unannotated seat", got)
	}
}

// A malformed environment value must not reach the row. The id's contract is
// that it can be dictated; a value that cannot is worse than none.
func TestSessionRowDropsAMalformedSessionID(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "not-an-id")
	got, err := Render(Options{Target: shortIDFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "not-an-id") {
		t.Errorf("statusline = %q, want the malformed id dropped", got)
	}
	if !strings.Contains(got, "Angie [she] //") {
		t.Errorf("statusline = %q, want the unannotated seat", got)
	}
}

// Subagent rows deliberately omit the id: they all share one session, so the
// same four characters on every row would disambiguate nothing.
func TestSubagentRowsOmitTheShortID(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "uz86")
	target := shortIDFixture(t)
	// Marshaled, not concatenated: a Windows cwd carries backslashes that are
	// invalid JSON string escapes.
	request, err := json.Marshal(map[string]any{
		"columns": 72,
		"tasks":   []map[string]any{{"id": "a", "cwd": target, "status": "running"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	out, err := RenderSubagents(strings.NewReader(string(request)), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "uz86") {
		t.Errorf("subagent rows = %q, want no session id: every row would repeat it", out)
	}
	if !strings.Contains(out, "Angie [she]") {
		t.Errorf("subagent rows = %q, want the seat still named", out)
	}
}
