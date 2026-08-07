package nativeui

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// update rewrites the vendored harness data from the installed binary. See
// docs/harness-vendoring.md for the refresh procedure.
var update = flag.Bool("update", false, "rewrite vendored harness data from the installed Claude Code binary")

const (
	vendoredVerbsFile   = "harness-default-verbs.txt"
	vendoredTokensFile  = "harness-theme-tokens.txt"
	vendoredVersionFile = "harness-version.txt"
)

// verbAnchor is the first element of the default spinner verb array. The array
// is one JS literal, so the anchor plus the closing bracket bounds it.
const verbAnchor = `["Accomplishing"`

// tokenAnchor appears once per base theme object. Every base is keyed alike, so
// the union across all of them is the set of token names the harness knows.
const tokenAnchor = `promptBorderShimmer:"`

var (
	quotedString = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)
	objectKey    = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*):"`)
	hexEscape    = regexp.MustCompile(`\\x([0-9A-Fa-f]{2})`)
	unicodeEsc   = regexp.MustCompile(`\\u([0-9A-Fa-f]{4})`)
)

// The vendored lists are the whole coupling to an unpublished contract, so the
// check compares extracted content rather than a version string.
func TestVendoredHarnessDataMatchesTheInstalledBinary(t *testing.T) {
	path := installedHarness(t)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read harness binary %s: %v", path, err)
	}
	source := string(raw)

	verbs, err := extractDefaultVerbs(source)
	if err != nil {
		t.Fatalf("extract default verbs from %s: %v", path, err)
	}
	tokens, err := extractThemeTokens(source)
	if err != nil {
		t.Fatalf("extract theme tokens from %s: %v", path, err)
	}

	if *update {
		writeVendoredLines(t, vendoredVerbsFile, verbs)
		writeVendoredLines(t, vendoredTokensFile, tokens)
		writeVendoredLines(t, vendoredVersionFile, []string{harnessVersion(t, path)})
		t.Logf("refreshed vendored harness data from %s", path)
		return
	}

	compareVendored(t, vendoredVerbsFile, "default spinner verb", verbs)
	compareVendored(t, vendoredTokensFile, "theme token", tokens)
}

// installedHarness resolves the binary, or skips. A machine without Claude Code
// installed cannot check this coupling and must not fail for it.
func installedHarness(t *testing.T) string {
	t.Helper()
	if override := strings.TrimSpace(os.Getenv("AGENT_COMPOSE_CLAUDE_BINARY")); override != "" {
		return override
	}
	path, err := exec.LookPath("claude")
	if err != nil {
		t.Skip("claude is not on PATH, so the vendored harness coupling cannot be checked")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

func harnessVersion(t *testing.T, path string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "--version").Output()
	if err != nil {
		t.Fatalf("read harness version from %s: %v", path, err)
	}
	return strings.TrimSpace(string(out))
}

func extractDefaultVerbs(source string) ([]string, error) {
	start := strings.Index(source, verbAnchor)
	if start < 0 {
		return nil, fmt.Errorf("verb anchor %s is absent", verbAnchor)
	}
	end := strings.Index(source[start:], "]")
	if end < 0 {
		return nil, fmt.Errorf("verb array is unterminated")
	}
	matches := quotedString.FindAllStringSubmatch(source[start:start+end+1], -1)
	verbs := make([]string, 0, len(matches))
	for _, match := range matches {
		verbs = append(verbs, unescapeJS(match[1]))
	}
	if len(verbs) < 100 {
		return nil, fmt.Errorf("verb array yielded only %d entries", len(verbs))
	}
	slices.Sort(verbs)
	return slices.Compact(verbs), nil
}

func extractThemeTokens(source string) ([]string, error) {
	seen := map[string]bool{}
	bases := 0
	for offset := 0; ; {
		index := strings.Index(source[offset:], tokenAnchor)
		if index < 0 {
			break
		}
		at := offset + index
		offset = at + len(tokenAnchor)
		open := strings.LastIndex(source[:at], "{")
		closing := strings.Index(source[at:], "}")
		if open < 0 || closing < 0 || at-open > 20000 {
			continue
		}
		keys := objectKey.FindAllStringSubmatch(source[open:at+closing], -1)
		if len(keys) < 20 {
			continue
		}
		bases++
		for _, key := range keys {
			seen[key[1]] = true
		}
	}
	if bases == 0 {
		return nil, fmt.Errorf("no base theme object carries %s", tokenAnchor)
	}
	tokens := make([]string, 0, len(seen))
	for token := range seen {
		tokens = append(tokens, token)
	}
	slices.Sort(tokens)
	return tokens, nil
}

// unescapeJS handles the escapes a bundler emits inside string literals. Two
// default verbs carry \xE9, which is why this exists.
func unescapeJS(value string) string {
	value = hexEscape.ReplaceAllStringFunc(value, func(match string) string {
		code, err := strconv.ParseInt(match[2:], 16, 32)
		if err != nil {
			return match
		}
		return string(rune(code))
	})
	value = unicodeEsc.ReplaceAllStringFunc(value, func(match string) string {
		code, err := strconv.ParseInt(match[2:], 16, 32)
		if err != nil {
			return match
		}
		return string(rune(code))
	})
	return strings.NewReplacer(`\'`, "'", `\"`, `"`, `\\`, `\`).Replace(value)
}

func readVendoredLines(t *testing.T, name string) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read vendored %s: %v", name, err)
	}
	var lines []string
	for _, line := range strings.Split(string(raw), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func writeVendoredLines(t *testing.T, name string, lines []string) {
	t.Helper()
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join("testdata", name), []byte(body), 0o644); err != nil {
		t.Fatalf("write vendored %s: %v", name, err)
	}
}

func compareVendored(t *testing.T, name, label string, extracted []string) {
	t.Helper()
	vendored := readVendoredLines(t, name)
	if slices.Equal(vendored, extracted) {
		return
	}
	added, removed := difference(extracted, vendored), difference(vendored, extracted)
	t.Errorf(
		"vendored %s list is stale: %d added %v, %d removed %v. "+
			"Refresh with `ward exec harness-refresh` and review what changed.",
		label, len(added), added, len(removed), removed,
	)
}

func difference(from, without []string) []string {
	var only []string
	for _, value := range from {
		if !slices.Contains(without, value) {
			only = append(only, value)
		}
	}
	return only
}
