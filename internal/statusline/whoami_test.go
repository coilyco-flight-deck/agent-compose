package statusline

import (
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/agentid"
)

func TestWhoamiNamesTheSeatAndSession(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "uz86")
	got, err := Whoami(Options{Target: shortIDFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Angie [she] uz86" {
		t.Errorf("Whoami = %q, want Angie [she] uz86", got)
	}
}

// Whoami and the status-line row must agree, or two surfaces name the same
// session differently and the hook stops being authoritative.
func TestWhoamiAgreesWithTheStatusLineRow(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "uz86")
	target := shortIDFixture(t)
	name, err := Whoami(Options{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	row, err := Render(Options{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if name == "" || !strings.Contains(row, name) {
		t.Errorf("row %q does not carry the whoami name %q", row, name)
	}
}

// Silence rather than a guess. A session with no composition has no composed
// name, and inventing one is what the retired local computation did wrong.
func TestWhoamiSuppressesWithoutAProjection(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "uz86")
	got, err := Whoami(Options{Target: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("Whoami = %q, want empty without a projection", got)
	}
}

func TestWhoamiWithoutASessionStillNamesTheSeat(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "")
	got, err := Whoami(Options{Target: shortIDFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Angie [she]" {
		t.Errorf("Whoami = %q, want Angie [she]", got)
	}
}
