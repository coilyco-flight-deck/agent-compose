package remoteskills

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func gitFixture(t *testing.T, dir string, args ...string) string {
	t.Helper()
	argv := append([]string{
		"-C", dir,
		"-c", "user.email=test@example.test",
		"-c", "user.name=Test",
		"-c", "commit.gpgSign=false",
		"-c", "tag.gpgSign=false",
	}, args...)
	raw, err := exec.Command("git", argv...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, raw)
	}
	return strings.TrimSpace(string(raw))
}

func newOrigin(t *testing.T) string {
	t.Helper()
	origin := filepath.Join(t.TempDir(), "origin")
	if err := os.MkdirAll(filepath.Join(origin, ".agents", "skills", "example"), 0o755); err != nil {
		t.Fatal(err)
	}
	gitFixture(t, origin, "init", "-b", "main", ".")
	writeSkill(t, origin, "v1")
	gitFixture(t, origin, "add", ".")
	gitFixture(t, origin, "commit", "-m", "v1")
	gitFixture(t, origin, "tag", "v1")
	return origin
}

func writeSkill(t *testing.T, origin, body string) {
	t.Helper()
	path := filepath.Join(origin, ".agents", "skills", "example", "SKILL.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readSkill(t *testing.T, catalog string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(catalog, "example", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func testOptions(t *testing.T, ttl time.Duration) Options {
	t.Helper()
	return Options{StateDir: t.TempDir(), TTL: ttl, Now: time.Now}
}

func TestHydrateClonesThenReusesFreshCheckout(t *testing.T) {
	source := Source{URL: newOrigin(t)}
	opts := testOptions(t, time.Hour)

	first, err := Hydrate(context.Background(), []Source{source}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || first[0].State != StateHydrated {
		t.Fatalf("first hydration = %+v", first)
	}
	if got := readSkill(t, first[0].Catalog.Path); got != "v1" {
		t.Fatalf("skill = %q, want v1", got)
	}
	if first[0].Source.Ref != "HEAD:.agents/skills" {
		t.Fatalf("canonical ref = %q", first[0].Source.Ref)
	}

	sentinel := filepath.Join(filepath.Dir(filepath.Dir(first[0].Catalog.Path)), "sentinel")
	if err := os.WriteFile(sentinel, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := Hydrate(context.Background(), []Source{source}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if second[0].State != StateCached {
		t.Fatalf("second hydration state = %s, want cached", second[0].State)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatal("fresh cache recreated its working checkout")
	}
}

func TestHydrateUsesGitRevisionPath(t *testing.T) {
	origin := newOrigin(t)
	if err := os.MkdirAll(filepath.Join(origin, "catalog"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(
		filepath.Join(origin, ".agents", "skills"),
		filepath.Join(origin, "catalog", "skills"),
	); err != nil {
		t.Fatal(err)
	}
	gitFixture(t, origin, "add", ".")
	gitFixture(t, origin, "commit", "-m", "move catalog")

	results, err := Hydrate(
		context.Background(),
		[]Source{{URL: origin, Ref: "main:catalog/skills"}},
		testOptions(t, time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := readSkill(t, results[0].Catalog.Path); got != "v1" {
		t.Fatalf("skill = %q, want v1", got)
	}
}

func TestHydrateRefreshesStaleMirrorAndCheckout(t *testing.T) {
	origin := newOrigin(t)
	source := Source{URL: origin}
	opts := testOptions(t, time.Hour)
	if _, err := Hydrate(context.Background(), []Source{source}, opts); err != nil {
		t.Fatal(err)
	}

	writeSkill(t, origin, "v2")
	gitFixture(t, origin, "commit", "-am", "v2")
	opts.TTL = 0
	results, err := Hydrate(context.Background(), []Source{source}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if results[0].State != StateRefreshed {
		t.Fatalf("state = %s, want refreshed", results[0].State)
	}
	if got := readSkill(t, results[0].Catalog.Path); got != "v2" {
		t.Fatalf("skill after refresh = %q, want v2", got)
	}
}

func TestHydrateFallsBackOffline(t *testing.T) {
	origin := newOrigin(t)
	source := Source{URL: origin}
	opts := testOptions(t, time.Hour)
	first, err := Hydrate(context.Background(), []Source{source}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(origin); err != nil {
		t.Fatal(err)
	}

	opts.TTL = 0
	results, err := Hydrate(context.Background(), []Source{source}, opts)
	if err != nil {
		t.Fatalf("offline hydration: %v", err)
	}
	if results[0].State != StateFallback || results[0].Warning == "" {
		t.Fatalf("offline result = %+v", results[0])
	}
	if got := readSkill(t, results[0].Catalog.Path); got != "v1" {
		t.Fatalf("offline skill = %q, want v1", got)
	}
	if results[0].Catalog.Path != first[0].Catalog.Path {
		t.Fatal("offline fallback changed the cached catalog path")
	}
}

func TestHydrateChecksOutTagsAndSkipsNetworkForPinnedCommit(t *testing.T) {
	origin := newOrigin(t)
	writeSkill(t, origin, "v2")
	gitFixture(t, origin, "commit", "-am", "v2")

	tagged := Source{URL: origin, Ref: "v1"}
	taggedResult, err := Hydrate(context.Background(), []Source{tagged}, testOptions(t, time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if got := readSkill(t, taggedResult[0].Catalog.Path); got != "v1" {
		t.Fatalf("tagged skill = %q, want v1", got)
	}

	sha := gitFixture(t, origin, "rev-parse", "main")
	pinned := Source{URL: origin, Ref: sha}
	opts := testOptions(t, time.Hour)
	first, err := Hydrate(context.Background(), []Source{pinned}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(origin); err != nil {
		t.Fatal(err)
	}
	opts.TTL = 0
	second, err := Hydrate(context.Background(), []Source{pinned}, opts)
	if err != nil {
		t.Fatalf("pinned offline hydration: %v", err)
	}
	if second[0].State != StateCached || second[0].Warning != "" {
		t.Fatalf("pinned result = %+v", second[0])
	}
	if readSkill(t, first[0].Catalog.Path) != readSkill(t, second[0].Catalog.Path) {
		t.Fatal("pinned checkout changed")
	}
}

func TestCacheKeyCoversCompleteLocator(t *testing.T) {
	base := Source{URL: "https://example.test/catalog.git", Ref: "main:.agents/skills"}
	keys := map[string]bool{
		cacheKey(base): true,
		cacheKey(Source{URL: base.URL + "-other", Ref: base.Ref}):   true,
		cacheKey(Source{URL: base.URL, Ref: "v1:.agents/skills"}):   true,
		cacheKey(Source{URL: base.URL, Ref: "main:catalog/skills"}): true,
	}
	if len(keys) != 4 {
		t.Fatal("remote locator fields collided in the cache key")
	}
	for key := range keys {
		if len(key) != sha256HexLength {
			t.Fatalf("cache key length = %d, want %d", len(key), sha256HexLength)
		}
	}
}

const sha256HexLength = 64

func TestFirstHydrationFailureLeavesNoWorkingCheckout(t *testing.T) {
	source := Source{URL: newOrigin(t), Ref: "main:missing/skills"}
	opts := testOptions(t, time.Hour)
	if _, err := Hydrate(context.Background(), []Source{source}, opts); err == nil {
		t.Fatal("missing skill path passed")
	}
	work := filepath.Join(opts.StateDir, "cache", "remote-skills", cacheKey(source), "work")
	if _, err := os.Stat(work); !os.IsNotExist(err) {
		t.Fatalf("failed first hydration left working checkout: %v", err)
	}
}

func TestValidateSourceRejectsUnsafeRevisionPaths(t *testing.T) {
	for name, source := range map[string]Source{
		"missing url":      {},
		"missing revision": {URL: "origin", Ref: ":skills"},
		"absolute path":    {URL: "origin", Ref: "main:/skills"},
		"escaping path":    {URL: "origin", Ref: "main:../skills"},
		"empty path":       {URL: "origin", Ref: "main:"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateSource(source); err == nil {
				t.Fatal("invalid source passed")
			}
		})
	}
}
