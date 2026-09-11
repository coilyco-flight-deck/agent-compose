package person

import (
	"path/filepath"
	"testing"
)

// The case that actually bites: a roster mounted through another session's
// shadow, which that session may edit under every other seat (#7363).
func TestVolatileRootCatchesASessionShadow(t *testing.T) {
	volatile := []string{
		"/private/var/folders/gt/T/aos/native/xh89/projects/x/seed/roster",
		"/tmp/aos/native/ab12/projects/x/seed/roster",
		"/tmp/scratch/roster",
		filepath.Join("/private/tmp", "roster"),
	}
	for _, path := range volatile {
		if !isVolatileRoot(path) {
			t.Errorf("volatile path not caught: %s", path)
		}
	}
}

// The control the check above needs: a roster in a durable checkout must stay
// silent, or the warning fires on every well-configured host and gets ignored.
func TestVolatileRootIsSilentOnADurableCheckout(t *testing.T) {
	durable := []string{
		"/Users/kai/projects/coilyco-flight-deck/agent-compose/seed/roster",
		"/opt/homebrew/share/agent-compose/roster",
		"/usr/local/share/agent-compose/roster",
		"/home/kai/.agent-compose/roster",
	}
	for _, path := range durable {
		if isVolatileRoot(path) {
			t.Errorf("durable path flagged as volatile: %s", path)
		}
	}
}

// A session shadow has a live owner, a temp dir only a lifetime. Saying the
// first about the second is the false alarm that teaches readers to skip it.
func TestVolatileKindSeparatesAnOwnerFromALifetime(t *testing.T) {
	if got := volatileKind("/tmp/aos/native/xh89/projects/x/seed/roster"); got != SessionShadow {
		t.Errorf("session shadow kind = %q, want %q", got, SessionShadow)
	}
	if got := volatileKind("/tmp/scratch/roster"); got != "temporary" {
		t.Errorf("plain temp kind = %q, want %q", got, "temporary")
	}
	if got := volatileKind("/Users/kai/projects/x/seed/roster"); got != "" {
		t.Errorf("durable checkout kind = %q, want empty", got)
	}
}

// The false negative the frontend seat found: volatility is a property of
// where the roster lands, not of whether a symlink was crossed reaching it.
func TestProvenanceReportsAVolatileRootReachedWithoutASymlink(t *testing.T) {
	dir := t.TempDir()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	// t.TempDir may hand back an already-canonical path, which is exactly the
	// shape that used to return no finding at all.
	if kind := volatileKind(resolved); kind == "" {
		t.Fatalf("a temp dir is not volatile: %s", resolved)
	}
	t.Setenv(RosterEnv, resolved)
	root, target, kind := RosterProvenanceKind()
	if kind == "" {
		t.Fatalf("no finding for a roster in %s", resolved)
	}
	if root != resolved {
		t.Fatalf("root = %q, want %q", root, resolved)
	}
	if target != "" {
		t.Fatalf("target = %q, want empty when no symlink was crossed", target)
	}
}
