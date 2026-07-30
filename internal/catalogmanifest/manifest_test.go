package catalogmanifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeManifestTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadPreservesVerifiedLocalCatalogueOrder(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second")
	for _, path := range []string{first, second} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	manifest := filepath.Join(root, "catalogues.json")
	writeManifestTestFile(t, manifest, `{
  "format": "aos.catalogues.v1",
  "catalogues": [
    {"source": "one/catalogue@main", "path": "`+filepath.ToSlash(first)+`", "commit": "1111111111111111111111111111111111111111"},
    {"source": "two/catalogue@main", "path": "`+filepath.ToSlash(second)+`", "commit": "2222222222222222222222222222222222222222"}
  ]
}
`)
	catalogs, err := Load(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogs) != 2 ||
		catalogs[0].Path != first ||
		catalogs[1].Path != second {
		t.Fatalf("catalogue order = %+v", catalogs)
	}
}

func TestLoadRejectsUntrustedManifestInputs(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "catalogue")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	regular := filepath.Join(root, "regular")
	writeManifestTestFile(t, regular, "not a directory\n")
	commit := "1111111111111111111111111111111111111111"
	for name, body := range map[string]string{
		"unsupported format": `{"format":"aos.catalogues.v2","catalogues":[]}`,
		"unknown field":      `{"format":"aos.catalogues.v1","catalogues":[],"remote":"https://example.test"}`,
		"relative path": `{"format":"aos.catalogues.v1","catalogues":[` +
			`{"source":"one","path":"relative","commit":"` + commit + `"}]}`,
		"missing path": `{"format":"aos.catalogues.v1","catalogues":[` +
			`{"source":"one","path":"` + filepath.ToSlash(filepath.Join(root, "missing")) +
			`","commit":"` + commit + `"}]}`,
		"regular file": `{"format":"aos.catalogues.v1","catalogues":[` +
			`{"source":"one","path":"` + filepath.ToSlash(regular) +
			`","commit":"` + commit + `"}]}`,
		"short commit": `{"format":"aos.catalogues.v1","catalogues":[` +
			`{"source":"one","path":"` + filepath.ToSlash(directory) +
			`","commit":"1234"}]}`,
		"trailing JSON": `{"format":"aos.catalogues.v1","catalogues":[]}` +
			`{"format":"aos.catalogues.v1","catalogues":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			manifest := filepath.Join(root, strings.ReplaceAll(name, " ", "-")+".json")
			writeManifestTestFile(t, manifest, body)
			if _, err := Load(manifest); err == nil {
				t.Fatal("invalid manifest passed")
			}
		})
	}
}
