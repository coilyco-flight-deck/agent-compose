package home

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func quietWarn(string) (int, error) { return 0, nil }

func TestFreshHostResolvesWithoutCreating(t *testing.T) {
	root := t.TempDir()
	dir, err := dirFrom(root, quietWarn)
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Join(root, ".agent-compose") {
		t.Fatalf("unexpected dir %s", dir)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("resolution must not create the directory")
	}
}

func TestLegacyDirectoryMigrates(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, ".config", "agent-compose")
	if err := os.MkdirAll(filepath.Join(legacy, "sources"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "agent-compose.yaml"), []byte("sources: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var warned strings.Builder
	dir, err := dirFrom(root, func(s string) (int, error) { warned.WriteString(s); return len(s), nil })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "agent-compose.yaml")); err != nil {
		t.Fatal("config must have moved to the canonical home")
	}
	info, err := os.Lstat(legacy)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("legacy path must become a compatibility symlink")
	}
	if resolved, _ := os.Readlink(legacy); resolved != dir {
		t.Fatalf("symlink points at %s", resolved)
	}
	if _, err := os.Stat(filepath.Join(legacy, "agent-compose.yaml")); err != nil {
		t.Fatal("legacy path must keep resolving through the symlink")
	}
	if !strings.Contains(warned.String(), "migrated") {
		t.Fatal("migration must announce itself")
	}

	again, err := dirFrom(root, quietWarn)
	if err != nil || again != dir {
		t.Fatalf("second resolution must be a stable no-op: %s %v", again, err)
	}
}

func TestBothRealDirectoriesPreferCanonicalWithoutDestroying(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, ".config", "agent-compose")
	canonical := filepath.Join(root, ".agent-compose")
	for _, dir := range []string{legacy, canonical} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(legacy, "keep.txt"), []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var warned strings.Builder
	dir, err := dirFrom(root, func(s string) (int, error) { warned.WriteString(s); return len(s), nil })
	if err != nil || dir != canonical {
		t.Fatalf("must prefer canonical: %s %v", dir, err)
	}
	if _, err := os.Stat(filepath.Join(legacy, "keep.txt")); err != nil {
		t.Fatal("ambiguous case must never destroy legacy data")
	}
	if !strings.Contains(warned.String(), "both") {
		t.Fatal("ambiguous case must warn")
	}
}
