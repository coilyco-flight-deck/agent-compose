package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "contracts", name)
}

func TestParseRequestFixture(t *testing.T) {
	req, err := ParseRequest(fixture(t, "native-full.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	if req.Role != "engineer" || req.Personality != "curious" {
		t.Fatalf("unexpected identity: %+v", req)
	}
	if req.Delivery != DeliveryNativeSkills || req.Density != DensityFull {
		t.Fatalf("unexpected delivery/density: %+v", req)
	}
	if len(req.Sources) != 1 || req.Sources[0].ID != "aos-public" || !req.Sources[0].Required {
		t.Fatalf("unexpected sources: %+v", req.Sources)
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

func TestParseRequestFailsClosed(t *testing.T) {
	cases := map[string]string{
		"unknown node": `compose {
    role "engineer"
    personality "curious"
    delivery "native-skills"
    density "full"
    privacy-scope "public"
}`,
		"duplicate scalar": `compose {
    role "engineer"
    role "designer"
    personality "curious"
    delivery "native-skills"
    density "full"
}`,
		"bad delivery": `compose {
    role "engineer"
    personality "curious"
    delivery "carrier-pigeon"
    density "full"
}`,
		"bad density": `compose {
    role "engineer"
    personality "curious"
    delivery "compiled"
    density "verbose"
}`,
		"missing personality": `compose {
    role "engineer"
    delivery "native-skills"
    density "full"
}`,
		"source without declaration": `compose {
    role "engineer"
    personality "curious"
    delivery "native-skills"
    density "full"
    source "aos-public"
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
    personality "curious"
    delivery "native-skills"
    density "full"
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
    personality "curious"
    delivery "native-skills"
    density "full"
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
    personality "curious"
    delivery "native-skills"
    density "full"
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
	req, err := ParseRequest(fixture(t, "native-full.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	sources, missing, err := LoadSources(req, fixture(t, "native-full.kdl"))
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
	if len(src.Skills) != 4 {
		t.Fatalf("expected four skills, got %+v", src.Skills)
	}
}
