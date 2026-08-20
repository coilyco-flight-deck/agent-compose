package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// coreBindingRequest stages a package whose role binds a core personality it
// never defines, then points a request at the reserved embedded root.
func coreBindingRequest(t *testing.T, libraries ...string) string {
	t.Helper()
	profile := filepath.Join(t.TempDir(), "profile")
	if err := os.CopyFS(profile, os.DirFS(filepath.Join("..", "..", "examples", "person-profile"))); err != nil {
		t.Fatal(err)
	}
	// The staged package depends on the core library alone, so the shared
	// example library stays out of what this test proves.
	rebind(t, filepath.Join(profile, "roles", "01-bulk-captioner.kdl"),
		`personality "local-guide" "shared-care"`, `personality "local-guide"`)
	rebind(t, filepath.Join(profile, "roles", "02-caption-review.kdl"),
		`personality "local-guide"`, `personality "skeptical"`)
	body := "compose {\n    person-source \".\"\n"
	for _, library := range libraries {
		body += fmt.Sprintf("    personality-library %q\n", library)
	}
	body += "    role \"caption-review\"\n    delivery \"compiled\"\n    model-tier \"commodity\"\n}\n"
	request := filepath.Join(profile, "core-request.kdl")
	if err := os.WriteFile(request, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return request
}

func rebind(t *testing.T, path, from, to string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rebound := strings.Replace(string(raw), from, to, 1)
	if rebound == string(raw) {
		t.Fatalf("example profile no longer contains %q, so the fixture needs updating", from)
	}
	if err := os.WriteFile(path, []byte(rebound), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestComposeRequestAdmitsTheEmbeddedCorePersonalityLibrary(t *testing.T) {
	if _, err := Run(coreBindingRequest(t), t.TempDir()); err == nil ||
		!strings.Contains(err.Error(), `personality "skeptical" has no catalog binding`) {
		t.Fatalf("unreferenced core personality must fail, got %v", err)
	}

	result, err := Run(coreBindingRequest(t, "roster:core"), t.TempDir())
	if err != nil {
		t.Fatalf("reserved core root rejected: %v", err)
	}
	manifest := readManifest(t, result.Bundle.Dir)
	if len(manifest.Personalities) != 1 || manifest.Personalities[0] != "skeptical" {
		t.Fatalf("composed personalities = %+v", manifest.Personalities)
	}
	// Package selection stays exclusive: only the disposition axis crossed.
	for _, source := range manifest.Sources {
		if source == "roster:core" {
			t.Fatalf("core personality admission leaked the roster as a source: %+v", manifest.Sources)
		}
	}
	body, err := os.ReadFile(filepath.Join(
		result.Bundle.Dir, "content", "skills", "person%3Aexample", "personality-skeptical", "SKILL.md",
	))
	if err != nil {
		t.Fatalf("core personality body did not materialize: %v", err)
	}
	want, err := os.ReadFile(filepath.Join(
		"..", "person", "data", "personality-skeptical", "SKILL.md",
	))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), strings.TrimSpace(skillProse(string(want)))) {
		t.Fatal("composed core body diverged from the embedded copy")
	}
}

// skillProse drops the frontmatter so the comparison reads the prose the
// bundle carries rather than the header the renderer rewrites.
func skillProse(raw string) string {
	parts := strings.SplitN(raw, "---", 3)
	if len(parts) < 3 {
		return raw
	}
	return parts[2]
}
