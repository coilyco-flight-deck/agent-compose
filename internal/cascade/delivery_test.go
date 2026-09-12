package cascade

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Delivery coverage for the import mode, whose guard matters more than its
// saving. Rationale: docs/cascade.md.

func writeSource(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

const sourceBody = `# Doctrine

A substantive rule that would be inlined in the default mode.

## See also

- [justfile](justfile) - repo-local navigation the inline path strips.
`

func TestImportDeliveryEmitsPointerNotBody(t *testing.T) {
	dir := t.TempDir()
	src := writeSource(t, dir, "AGENTS.md", sourceBody)

	composed, _, err := ComposePartsDelivered([]string{src}, nil, nil, "", DeliveryImport)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}

	if !strings.Contains(composed, "\n@"+src) {
		t.Errorf("composed output has no import for %s:\n%s", src, composed)
	}
	if strings.Contains(composed, "A substantive rule") {
		t.Errorf("import mode inlined the body, which defeats the point:\n%s", composed)
	}
}

func TestInlineDeliveryStaysTheDefault(t *testing.T) {
	dir := t.TempDir()
	src := writeSource(t, dir, "AGENTS.md", sourceBody)

	viaDefault, _, err := ComposeParts([]string{src}, nil, nil, "")
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	viaEmpty, _, err := ComposePartsDelivered([]string{src}, nil, nil, "", "")
	if err != nil {
		t.Fatalf("compose: %v", err)
	}

	if viaDefault != viaEmpty {
		t.Errorf("an unset source_delivery changed the output; it must mean inline")
	}
	if !strings.Contains(viaDefault, "A substantive rule") {
		t.Errorf("inline mode stopped inlining the body:\n%s", viaDefault)
	}
}

// The control the whole change rests on: a pointer to nothing ships silently
// and every rule in that source stops firing.
func TestImportDeliveryFailsOnAMissingTarget(t *testing.T) {
	absent := filepath.Join(t.TempDir(), "AGENTS.md")

	_, _, err := ComposePartsDelivered([]string{absent}, nil, nil, "", DeliveryImport)

	if err == nil {
		t.Fatalf("composing an import of a missing file succeeded; a dead pointer would ship")
	}
	if !strings.Contains(err.Error(), absent) {
		t.Errorf("error does not name the missing source: %v", err)
	}
}

func TestImportDeliveryFailsOnAnEmptyTarget(t *testing.T) {
	dir := t.TempDir()
	src := writeSource(t, dir, "AGENTS.md", "")

	if _, _, err := ComposePartsDelivered([]string{src}, nil, nil, "", DeliveryImport); err == nil {
		t.Fatalf("composing an import of an empty file succeeded")
	}
}

func TestImportDeliveryFailsOnADirectory(t *testing.T) {
	dir := t.TempDir()

	if _, _, err := ComposePartsDelivered([]string{dir}, nil, nil, "", DeliveryImport); err == nil {
		t.Fatalf("composing an import of a directory succeeded")
	}
}

func TestImportDeliveryFailsOnARelativeSource(t *testing.T) {
	if _, _, err := ComposePartsDelivered([]string{"AGENTS.md"}, nil, nil, "", DeliveryImport); err == nil {
		t.Fatalf("composing a relative import succeeded; the harness would resolve it elsewhere")
	}
}

// An override rewrites a body import mode never reads, so refuse it rather than
// dropping it silently.
func TestImportDeliveryRefusesAnOverride(t *testing.T) {
	dir := t.TempDir()
	src := writeSource(t, dir, "AGENTS.md", sourceBody)
	override := writeSource(t, dir, "AGENTS.claude.md", "# Doctrine\n\nReplaced.\n")

	_, _, err := ComposePartsDelivered(
		[]string{src}, map[string]string{src: override}, nil, "", DeliveryImport)

	if err == nil {
		t.Fatalf("an override on an imported source was accepted and would be dropped")
	}
}

func TestUnknownDeliveryIsRejected(t *testing.T) {
	dir := t.TempDir()
	src := writeSource(t, dir, "AGENTS.md", sourceBody)

	if _, _, err := ComposePartsDelivered([]string{src}, nil, nil, "", "pointer"); err == nil {
		t.Fatalf("an unknown source_delivery was accepted")
	}
}

func TestImportDeliveryStillComposesTheAppendix(t *testing.T) {
	dir := t.TempDir()
	src := writeSource(t, dir, "AGENTS.md", sourceBody)
	appendix := []AppendixBlock{{Fence: "<!-- appendix -->", Body: "Tail rule."}}

	_, tail, err := ComposePartsDelivered([]string{src}, nil, appendix, "", DeliveryImport)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}

	if !strings.Contains(tail, "Tail rule.") {
		t.Errorf("import mode dropped the appendix: %q", tail)
	}
}

func TestImportDeliveryShrinksTheComposedFile(t *testing.T) {
	dir := t.TempDir()
	src := writeSource(t, dir, "AGENTS.md", strings.Repeat(sourceBody, 40))

	inline, _, err := ComposePartsDelivered([]string{src}, nil, nil, "", DeliveryInline)
	if err != nil {
		t.Fatalf("inline: %v", err)
	}
	imported, _, err := ComposePartsDelivered([]string{src}, nil, nil, "", DeliveryImport)
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if len(imported) >= len(inline) {
		t.Errorf("import mode did not shrink the composed file: %d >= %d",
			len(imported), len(inline))
	}
}
