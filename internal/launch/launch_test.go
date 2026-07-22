package launch

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "contracts", name)
}

func TestRefreshAcrossAllLayouts(t *testing.T) {
	cases := map[string]string{
		"claude":   "native-full.kdl",
		"codex":    "native-full.kdl",
		"goose":    "compiled-full.kdl",
		"opencode": "compiled-brief.kdl",
	}
	for layout, request := range cases {
		t.Run(layout, func(t *testing.T) {
			target := t.TempDir()
			result, err := Refresh(Options{
				RequestPath: fixture(t, request),
				Layout:      layout,
				TargetDir:   target,
				OutDir:      t.TempDir(),
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.Fallback || result.Projected == 0 {
				t.Fatalf("expected a fresh projection, got %+v", result)
			}
			if err := project.Validate(target); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWarmRefreshIsFastAndReuses(t *testing.T) {
	target, out := t.TempDir(), t.TempDir()
	opts := Options{RequestPath: fixture(t, "native-full.kdl"), Layout: "claude", TargetDir: target, OutDir: out}
	if _, err := Refresh(opts); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	result, err := Refresh(opts)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if !result.BundleReused {
		t.Fatalf("warm refresh must reuse the bundle, got %+v", result)
	}
	if elapsed > 250*time.Millisecond {
		t.Fatalf("warm refresh took %s, budget is 250ms", elapsed)
	}
}

func TestRefreshFallsBackToLastKnownGood(t *testing.T) {
	target, out := t.TempDir(), t.TempDir()
	good := Options{RequestPath: fixture(t, "native-full.kdl"), Layout: "claude", TargetDir: target, OutDir: out}
	if _, err := Refresh(good); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(t.TempDir(), "broken.kdl")
	if err := os.WriteFile(broken, []byte("compose { role \"engineer\""), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Refresh(Options{RequestPath: broken, Layout: "claude", TargetDir: target, OutDir: out})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Fallback || result.Warning == "" {
		t.Fatalf("expected a loud fallback, got %+v", result)
	}
	if !strings.Contains(readFile(t, target, "CLAUDE.md"), "Fixture foundation") {
		t.Fatal("last-known-good projection must remain intact")
	}
}

func TestRefreshFailsWithoutLastKnownGood(t *testing.T) {
	broken := filepath.Join(t.TempDir(), "broken.kdl")
	if err := os.WriteFile(broken, []byte("not kdl at all {"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Refresh(Options{RequestPath: broken, Layout: "claude", TargetDir: t.TempDir(), OutDir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "no last-known-good") {
		t.Fatalf("expected hard failure, got %v", err)
	}
}

func TestConcurrentIdenticalLaunchesConverge(t *testing.T) {
	target, out := t.TempDir(), t.TempDir()
	opts := Options{RequestPath: fixture(t, "native-full.kdl"), Layout: "claude", TargetDir: target, OutDir: out}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Refresh(opts); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	var bundles int
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			bundles++
		}
	}
	if bundles != 1 {
		t.Fatalf("identical launches must converge on one cache entry, found %d", bundles)
	}
	if err := project.Validate(target); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentDistinctRequestsStayIsolated(t *testing.T) {
	out := t.TempDir()
	targetA, targetB := t.TempDir(), t.TempDir()
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	launches := []Options{
		{RequestPath: fixture(t, "native-full.kdl"), Layout: "claude", TargetDir: targetA, OutDir: out},
		{RequestPath: fixture(t, "native-brief.kdl"), Layout: "claude", TargetDir: targetB, OutDir: out},
	}
	for _, opts := range launches {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Refresh(opts); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	if !strings.Contains(readFile(t, targetA, ".claude/skills/personality-curious/SKILL.md"), "# Curious") {
		t.Fatal("target A got the wrong personality")
	}
	if !strings.Contains(readFile(t, targetB, ".claude/skills/personality-meticulous/SKILL.md"), "# Meticulous") {
		t.Fatal("target B got the wrong personality")
	}
}

func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
