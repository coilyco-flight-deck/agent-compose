package bundle_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func composeBundle(t *testing.T, name string) string {
	t.Helper()
	result, err := compose.Run(filepath.Join("..", "..", "testdata", "contracts", name), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return result.Bundle.Dir
}

func TestVerifyNativeAndCompiledBundles(t *testing.T) {
	cases := []struct {
		request string
		mode    string
	}{
		{"native-full.kdl", schema.DeliveryNativeSkills},
		{"compiled-full.kdl", schema.DeliveryCompiled},
	}
	for _, tc := range cases {
		t.Run(tc.mode, func(t *testing.T) {
			verified, err := bundle.Verify(composeBundle(t, tc.request))
			if err != nil {
				t.Fatal(err)
			}
			if verified.Manifest.Delivery.Mode != tc.mode {
				t.Fatalf("delivery = %q, want %q", verified.Manifest.Delivery.Mode, tc.mode)
			}
			if len(verified.Identities) != len(verified.Manifest.Personalities)+1 {
				t.Fatalf("identities = %+v, personalities = %v", verified.Identities, verified.Manifest.Personalities)
			}
			var foundOrdinary bool
			for _, identity := range verified.Identities {
				if identity.Source != "aos-public" {
					t.Fatalf("identity source = %q, want aos-public", identity.Source)
				}
				if identity.Skill == "fixture-review" {
					foundOrdinary = true
				}
			}
			if !foundOrdinary {
				t.Fatalf("ordinary skill missing from verified identities: %+v", verified.Identities)
			}
			if verified.Files == 0 {
				t.Fatal("verified bundle reported no files")
			}
		})
	}
}

func TestVerifyRejectsUnsafeIncompleteAndAmbiguousBundles(t *testing.T) {
	t.Run("unsafe manifest path", func(t *testing.T) {
		dir := copyBundle(t, composeBundle(t, "native-full.kdl"))
		manifest, err := bundle.ReadManifest(dir)
		if err != nil {
			t.Fatal(err)
		}
		manifest.Delivery.Instructions = "../outside"
		writeJSON(t, filepath.Join(dir, "manifest.json"), manifest)
		if _, err := bundle.Verify(dir); err == nil || !strings.Contains(err.Error(), "safe relative path") {
			t.Fatalf("expected unsafe-path failure, got %v", err)
		}
	})

	t.Run("missing entry point", func(t *testing.T) {
		dir := copyBundle(t, composeBundle(t, "compiled-full.kdl"))
		if err := os.Remove(filepath.Join(dir, "delivery", "compiled.md")); err != nil {
			t.Fatal(err)
		}
		if _, err := bundle.Verify(dir); err == nil || !strings.Contains(err.Error(), "compiled_context") {
			t.Fatalf("expected missing-entry failure, got %v", err)
		}
	})

	t.Run("extra identity", func(t *testing.T) {
		dir := copyBundle(t, composeBundle(t, "native-full.kdl"))
		extra := filepath.Join(dir, "content", "skills", "aos-public", "personality-extra")
		if err := os.MkdirAll(extra, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(extra, "SKILL.md"), []byte("# Grounded\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := bundle.Verify(dir); err == nil || !strings.Contains(err.Error(), "does not match selected identities") {
			t.Fatalf("expected identity-cardinality failure, got %v", err)
		}
	})

	t.Run("manifest trace mismatch", func(t *testing.T) {
		dir := copyBundle(t, composeBundle(t, "native-full.kdl"))
		manifest, err := bundle.ReadManifest(dir)
		if err != nil {
			t.Fatal(err)
		}
		manifest.Personalities[0] = "other"
		writeJSON(t, filepath.Join(dir, "manifest.json"), manifest)
		if _, err := bundle.Verify(dir); err == nil || !strings.Contains(err.Error(), "manifest profile") {
			t.Fatalf("expected profile-trace failure, got %v", err)
		}
	})

	t.Run("missing identity document", func(t *testing.T) {
		dir := copyBundle(t, composeBundle(t, "native-full.kdl"))
		skillDoc := filepath.Join(dir, "content", "skills", "aos-public", "personality-curious", "SKILL.md")
		if err := os.Remove(skillDoc); err != nil {
			t.Fatal(err)
		}
		if _, err := bundle.Verify(dir); err == nil || !strings.Contains(err.Error(), "has no SKILL.md") {
			t.Fatalf("expected missing-identity-document failure, got %v", err)
		}
	})

	t.Run("symlink", func(t *testing.T) {
		dir := copyBundle(t, composeBundle(t, "native-full.kdl"))
		link := filepath.Join(dir, "content", "linked")
		if err := os.Symlink("instructions.md", link); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		if _, err := bundle.Verify(dir); err == nil || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("expected symlink failure, got %v", err)
		}
	})
}

func TestMaterializeVerifiesCacheReuse(t *testing.T) {
	out := t.TempDir()
	request := filepath.Join("..", "..", "testdata", "contracts", "native-full.kdl")
	first, err := compose.Run(request, out)
	if err != nil {
		t.Fatal(err)
	}
	extra := filepath.Join(first.Bundle.Dir, "content", "skills", "aos-public", "personality-extra")
	if err := os.MkdirAll(extra, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := compose.Run(request, out); err == nil || !strings.Contains(err.Error(), "cached bundle") {
		t.Fatalf("expected invalid-cache failure, got %v", err)
	}
}

func TestMaterializeKeysRenderedInstructions(t *testing.T) {
	out := t.TempDir()
	request := filepath.Join("..", "..", "testdata", "contracts", "native-full.kdl")
	first, err := compose.Run(request, out)
	if err != nil {
		t.Fatal(err)
	}

	first.Resolution.RoleBriefing += "\n\nThe agent follows the revised role contract."
	second, err := bundle.Materialize(first.Resolution, out)
	if err != nil {
		t.Fatal(err)
	}
	if second.Key == first.Bundle.Key || second.Reused {
		t.Fatalf("rendered instruction change reused stale bundle: %+v then %+v", first.Bundle, second)
	}
}

func TestMaterializeRejectsUnsafeIdentitySegments(t *testing.T) {
	resolution := &resolver.Resolution{
		Request: &schema.Request{
			Role:     "engineer",
			Delivery: schema.DeliveryNativeSkills, Density: schema.DensityFull,
		},
		Person: &person.Person{Raw: []byte("fixture")},
		Skills: []resolver.Selected{{
			Source: "aos-public", ID: "../outside", Path: t.TempDir(),
		}},
	}
	if _, err := bundle.Materialize(resolution, t.TempDir()); err == nil || !strings.Contains(err.Error(), "unsafe") {
		t.Fatalf("expected unsafe identity failure, got %v", err)
	}
}

func copyBundle(t *testing.T, source string) string {
	t.Helper()
	target := t.TempDir()
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		out := filepath.Join(target, rel)
		if entry.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(out, raw, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return target
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
