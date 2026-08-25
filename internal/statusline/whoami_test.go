package statusline

import (
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/agentid"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/launch"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
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

// seatBundleFixture writes one bundle carrying a single named seat, so a test
// can bind two sessions to two different compositions.
func seatBundleFixture(t *testing.T, role, name, pronouns string) string {
	t.Helper()
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "manifest.json"), bundle.Manifest{
		Format: "agent-compose.bundle", Role: role, ModelTier: "frontier",
		Identity: bundle.RoleIdentity{
			Person: "core",
			Seats:  []person.Seat{{Key: "claude", Name: name, Pronouns: pronouns}},
		},
	})
	return dir
}

// The failure this binding exists to fix: a native session shadow has no
// projection above its cwd, so the walk found nothing and whoami said nothing.
func TestWhoamiAnswersFromTheSessionBundleWithNoProjectionInReach(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "kj58")
	t.Setenv(launch.SessionBundleEnv, seatBundleFixture(t, "platform", "Angie", "she"))
	t.Setenv(launch.SessionLayoutEnv, "claude")
	got, err := Whoami(Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Angie [she] kj58" {
		t.Errorf("Whoami = %q, want Angie [she] kj58", got)
	}
}

// Two roles on one host resolved to whichever projection converged last.
func TestWhoamiKeepsConcurrentSessionsApart(t *testing.T) {
	platform := seatBundleFixture(t, "platform", "Angie", "she")
	tpm := seatBundleFixture(t, "tpm", "Portia", "they")
	for _, tc := range []struct{ bundleDir, want string }{
		{platform, "Angie [she]"},
		{tpm, "Portia [they]"},
	} {
		t.Setenv(agentid.SessionEnv, "")
		t.Setenv(launch.SessionBundleEnv, tc.bundleDir)
		t.Setenv(launch.SessionLayoutEnv, "claude")
		got, err := Whoami(Options{})
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("Whoami = %q, want %q", got, tc.want)
		}
	}
}

// --target is an inspection request, so it still reports what sits at that path.
func TestWhoamiTargetStillInspectsAProjection(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "")
	t.Setenv(launch.SessionBundleEnv, seatBundleFixture(t, "tpm", "Portia", "they"))
	t.Setenv(launch.SessionLayoutEnv, "claude")
	got, err := Whoami(Options{Target: shortIDFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Angie [she]" {
		t.Errorf("Whoami = %q, want the targeted projection's seat Angie [she]", got)
	}
}

// A half-set binding is not a binding, so the walk still runs.
func TestWhoamiIgnoresAnIncompleteBinding(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "")
	t.Setenv(launch.SessionBundleEnv, seatBundleFixture(t, "tpm", "Portia", "they"))
	t.Setenv(launch.SessionLayoutEnv, "")
	got, err := Whoami(Options{Target: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("Whoami = %q, want empty when the binding is incomplete", got)
	}
}

// The row and the hook resolve the same way, so a bound session with no
// projection in reach cannot have one surface name it and the other stay blank.
func TestTheRowAgreesWithWhoamiUnderASessionBinding(t *testing.T) {
	t.Setenv(agentid.SessionEnv, "kj58")
	t.Setenv(launch.SessionBundleEnv, shortIDBundleFixture(t))
	t.Setenv(launch.SessionLayoutEnv, "claude")
	name, err := Whoami(Options{})
	if err != nil {
		t.Fatal(err)
	}
	row, err := Render(Options{})
	if err != nil {
		t.Fatal(err)
	}
	if name == "" || !strings.Contains(row, name) {
		t.Errorf("row %q does not carry the whoami name %q", row, name)
	}
}
