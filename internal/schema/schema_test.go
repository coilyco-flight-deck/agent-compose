package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/personpolicy"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "contracts", name)
}

func TestParseRequestFixture(t *testing.T) {
	req, err := ParseRequest(fixture(t, "native.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	if req.Role != "engineer" {
		t.Fatalf("unexpected identity: %+v", req)
	}
	if req.Delivery != DeliveryNativeSkills {
		t.Fatalf("unexpected delivery: %+v", req)
	}
	if req.ModelClass != ModelClassFrontier {
		t.Fatalf("unexpected default model class: %+v", req)
	}
	if len(req.Sources) != 1 || req.Sources[0].ID != "aos-public" || !req.Sources[0].Required {
		t.Fatalf("unexpected sources: %+v", req.Sources)
	}
}

func TestParseRequestAcceptsLowContextModelClass(t *testing.T) {
	path := writeRequest(t, `compose {
    role "engineer"
    delivery "compiled"
    model-class "low-context"
}`)
	req, err := ParseRequest(path)
	if err != nil {
		t.Fatal(err)
	}
	if req.ModelClass != ModelClassLowContext {
		t.Fatalf("model class = %q", req.ModelClass)
	}
}

func TestParseRequestAcceptsRelativePersonSource(t *testing.T) {
	req, err := ParseRequest(fixture(t, "custom-person.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	if req.PersonSource != "person-independent" {
		t.Fatalf("person source = %q", req.PersonSource)
	}
	if req.PersonPolicy != personpolicy.ExternalOnly {
		t.Fatalf("person policy = %q", req.PersonPolicy)
	}
}

func writeRequest(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "request.kdl")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseRequestAcceptsLegacyFullDensity(t *testing.T) {
	path := writeRequest(t, `compose {
    role "engineer"
    delivery "compiled"
    density "full"
}`)
	req, err := ParseRequest(path)
	if err != nil {
		t.Fatal(err)
	}
	if req.Role != "engineer" || req.Delivery != DeliveryCompiled {
		t.Fatalf("unexpected legacy request: %+v", req)
	}
}

func TestParseRequestFailsClosed(t *testing.T) {
	cases := map[string]string{
		"unknown node": `compose {
    role "engineer"
    delivery "native-skills"
    privacy-scope "public"
}`,
		"duplicate scalar": `compose {
    role "engineer"
    role "designer"
    delivery "native-skills"
}`,
		"bad delivery": `compose {
    role "engineer"
    delivery "carrier-pigeon"
}`,
		"bad model class": `compose {
    role "engineer"
    delivery "native-skills"
    model-class "tiny"
}`,
		"retired brief density": `compose {
    role "engineer"
    delivery "compiled"
    density "brief"
}`,
		"retired personality selector": `compose {
    role "engineer"
    personality "curious"
    delivery "native-skills"
}`,
		"source without declaration": `compose {
    role "engineer"
    delivery "native-skills"
    source "aos-public"
}`,
		"source with declaration and root": `compose {
    role "engineer"
    delivery "native-skills"
    source "aos-public" declaration="source.kdl" root="."
}`,
		"source with empty root": `compose {
    role "engineer"
    delivery "native-skills"
    source "aos-public" root=""
}`,
		"absolute person source": `compose {
    person-source "/tmp/person"
    role "engineer"
    delivery "native-skills"
}`,
		"external-only without person source": `compose {
    person-policy "external-only"
    role "engineer"
    delivery "native-skills"
}`,
		"unknown person policy": `compose {
    person-policy "prefer-external"
    person-source "person"
    role "engineer"
    delivery "native-skills"
}`,
		"invalid kdl": `compose { role "engineer`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseRequest(writeRequest(t, body)); err == nil {
				t.Fatal("expected parse failure")
			}
		})
	}
}

func TestLoadSourcesRequiredVersusOptional(t *testing.T) {
	required := writeRequest(t, `compose {
    role "engineer"
    delivery "native-skills"
    source "ghost" declaration="ghost.kdl" required=#true
}`)
	req, err := ParseRequest(required)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadSources(req, required); err == nil {
		t.Fatal("expected a required missing source to fail")
	}

	optional := writeRequest(t, `compose {
    role "engineer"
    delivery "native-skills"
    source "ghost" declaration="ghost.kdl"
}`)
	req, err = ParseRequest(optional)
	if err != nil {
		t.Fatal(err)
	}
	sources, missing, err := LoadSources(req, optional)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 0 || len(missing) != 1 || missing[0].ID != "ghost" {
		t.Fatalf("expected one missing optional source, got %+v / %+v", sources, missing)
	}
}

func TestLoadSourcesRejectsEscapingPaths(t *testing.T) {
	path := writeRequest(t, `compose {
    role "engineer"
    delivery "native-skills"
    source "evil" declaration="../evil.kdl" required=#true
}`)
	req, err := ParseRequest(path)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = LoadSources(req, path)
	if err == nil || !strings.Contains(err.Error(), "relative and clean") {
		t.Fatalf("expected escaping declaration to fail, got %v", err)
	}
}

func TestParseSourceFixture(t *testing.T) {
	req, err := ParseRequest(fixture(t, "native.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	sources, missing, err := LoadSources(req, fixture(t, "native.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 0 || len(sources) != 1 {
		t.Fatalf("unexpected load result: %+v / %+v", sources, missing)
	}
	src := sources[0]
	if len(src.Instructions) != 1 || src.Instructions[0].ID != "foundation" {
		t.Fatalf("unexpected instructions: %+v", src.Instructions)
	}
	if len(src.Skills) != 1 || src.Skills[0].ID != "fixture-review" {
		t.Fatalf("expected only the ordinary fixture skill, got %+v", src.Skills)
	}
}

func TestLoadInferredProviderRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "public-provider")
	for _, name := range []string{"personality-bright", "personality-calm", "coding-go"} {
		dir := filepath.Join(root, ".agents", "skills", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	shared := filepath.Join(root, ".agents", "skills", "personality-shared")
	if err := os.MkdirAll(shared, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shared, "INVARIANT.md"), []byte("# Invariant\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"coding-shape-cli", "design-system"} {
		dir := filepath.Join(root, ".agents", "composed", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "COMPOSED.md"), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".agents", "roles.kdl"), []byte(`roles {
    role engineer {
        composed-skill coding-shape-cli
        intent autonomous-coding {
            harness openhands
        }
    }
    role designer {
        composed-skill design-system
    }
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	request := filepath.Join(root, "request.kdl")
	if err := os.WriteFile(request, []byte(`compose {
    role "engineer"
    delivery "native-skills"
    source "aos-public" root="." required=#true
}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	req, err := ParseRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	sources, missing, err := LoadSources(req, request)
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 0 || len(sources) != 1 {
		t.Fatalf("unexpected inferred load result: %+v / %+v", sources, missing)
	}
	src := sources[0]
	if src.ID != "aos-public" ||
		len(src.Instructions) != 1 ||
		src.Instructions[0].ID != "personality-invariant" {
		t.Fatalf("unexpected inferred source: %+v", src)
	}
	if len(src.Skills) != 3 ||
		src.Skills[0].ID != "coding-go" ||
		src.Skills[1].ID != "personality-bright" ||
		src.Skills[2].ID != "personality-calm" {
		t.Fatalf("inferred ordinary skills must be sorted: %+v", src.Skills)
	}
	if len(src.RoleSkills["engineer"]) != 1 ||
		src.RoleSkills["engineer"][0].ID != "coding-shape-cli" ||
		src.RoleSkills["engineer"][0].EntryPoint != "COMPOSED.md" ||
		len(src.RoleSkills["designer"]) != 1 ||
		src.RoleSkills["designer"][0].ID != "design-system" {
		t.Fatalf("unexpected composed role skills: %+v", src.RoleSkills)
	}
	if len(src.RoleIntents["engineer"]) != 1 ||
		src.RoleIntents["engineer"][0].Intent != "autonomous-coding" ||
		src.RoleIntents["engineer"][0].Harness != "openhands" {
		t.Fatalf("unexpected role intent routes: %+v", src.RoleIntents)
	}

	direct, err := LoadSource(root)
	if err != nil {
		t.Fatal(err)
	}
	if direct.ID != filepath.Base(root) ||
		len(direct.Skills) != len(src.Skills) ||
		len(direct.RoleSkills) != len(src.RoleSkills) ||
		len(direct.RoleIntents) != len(src.RoleIntents) {
		t.Fatalf("direct provider load differs: %+v", direct)
	}
}

func TestLoadInferredProviderRejectsUnsafeComposedLayouts(t *testing.T) {
	makeProvider := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		shared := filepath.Join(root, ".agents", "skills", "personality-shared")
		ordinary := filepath.Join(root, ".agents", "skills", "coding-go")
		if err := os.MkdirAll(shared, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(ordinary, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(shared, "INVARIANT.md"), []byte("# Invariant\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(ordinary, "SKILL.md"), []byte("# Go\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return root
	}

	t.Run("role bindings without composed root", func(t *testing.T) {
		root := makeProvider(t)
		if err := os.WriteFile(
			filepath.Join(root, ".agents", "roles.kdl"),
			[]byte("roles {}\n"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadSource(root); err == nil ||
			!strings.Contains(err.Error(), "exist without composed skills") {
			t.Fatalf("orphan role bindings must fail closed, got %v", err)
		}
	})

	t.Run("SKILL.md under composed", func(t *testing.T) {
		root := makeProvider(t)
		composed := filepath.Join(root, ".agents", "composed", "coding-shape-cli")
		if err := os.MkdirAll(composed, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(composed, "COMPOSED.md"), []byte("# CLI\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(composed, "SKILL.md"), []byte("# Leaked\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadSource(root); err == nil || !strings.Contains(err.Error(), "contains SKILL.md") {
			t.Fatalf("composed SKILL.md needs an actionable failure, got %v", err)
		}
	})

	t.Run("ordinary name collision", func(t *testing.T) {
		root := makeProvider(t)
		composed := filepath.Join(root, ".agents", "composed", "coding-go")
		if err := os.MkdirAll(composed, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(composed, "COMPOSED.md"), []byte("# Go\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadSource(root); err == nil || !strings.Contains(err.Error(), "exists in both") {
			t.Fatalf("ordinary/composed collisions must fail closed, got %v", err)
		}
	})

	t.Run("missing role binding target", func(t *testing.T) {
		root := makeProvider(t)
		if err := os.MkdirAll(filepath.Join(root, ".agents", "composed"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".agents", "roles.kdl"), []byte(`roles {
    role "engineer" {
        composed-skill "missing"
    }
}
`), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadSource(root); err == nil || !strings.Contains(err.Error(), "missing composed skill") {
			t.Fatalf("missing composed binding targets must fail closed, got %v", err)
		}
	})

	t.Run("malformed intent route", func(t *testing.T) {
		root := makeProvider(t)
		if err := os.MkdirAll(filepath.Join(root, ".agents", "composed"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".agents", "roles.kdl"), []byte(`roles {
    role engineer {
        intent autonomous-coding {
            model hidden-policy
        }
    }
}
`), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadSource(root); err == nil ||
			!strings.Contains(err.Error(), "expects exactly one harness child") {
			t.Fatalf("malformed intent route must fail closed, got %v", err)
		}
	})
}

func TestLoadInferredProviderAllowsMissingInvariant(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ordinary-provider")
	skill := filepath.Join(root, ".agents", "skills", "personality-bright")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("# Bright\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src, err := LoadSource(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(src.Instructions) != 0 || len(src.Skills) != 1 {
		t.Fatalf("provider without invariant loaded incorrectly: %+v", src)
	}
}
