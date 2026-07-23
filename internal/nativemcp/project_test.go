package nativemcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestProjectConvergesBothNativeRegistries(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	inventoryPath := filepath.Join(dir, "mcporter.json")
	write(t, inventoryPath, `{
  "imports": [],
  "mcpServers": {
    "reader": {"baseUrl": "https://mcp.example.test/mcp", "headers": {"X-Test": "${HOME}/token"}, "x-codex": {"defaultToolsApprovalMode": "approve"}},
    "local": {"command": "${HOME}/bin/server", "args": ["--stdio"], "env": {"CACHE": "${HOME}/cache"}}
  }
}
`)
	write(t, filepath.Join(home, ".claude.json"), `{"theme":"dark","mcpServers":{"stale":{"command":"old"}}}`+"\n")
	write(t, filepath.Join(home, ".codex", "config.toml"), `[features]
js_repl = false

[mcp_servers."computer-use"]
enabled = false
`)

	result, err := Project(Options{Inventory: inventoryPath, Home: home})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || result.Servers != 2 {
		t.Fatalf("result = %+v, want two changed servers", result)
	}

	var claude map[string]any
	raw, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &claude); err != nil {
		t.Fatal(err)
	}
	if claude["theme"] != "dark" {
		t.Fatalf("Claude foreign state was not preserved: %v", claude)
	}
	servers := claude["mcpServers"].(map[string]any)
	if _, ok := servers["stale"]; ok {
		t.Fatalf("stale Claude server survived hard overwrite: %v", servers)
	}
	if got := servers["reader"].(map[string]any)["type"]; got != "http" {
		t.Fatalf("Claude HTTP transport = %v, want http", got)
	}
	if _, ok := servers["reader"].(map[string]any)["x-codex"]; ok {
		t.Fatalf("Claude projection contains Codex-only metadata: %v", servers["reader"])
	}
	if got := servers["local"].(map[string]any)["command"]; got != filepath.Join(home, "bin", "server") {
		t.Fatalf("Claude home expansion = %v", got)
	}

	codexRaw, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	codex := string(codexRaw)
	for _, want := range []string{
		`[mcp_servers."computer-use"]`,
		"enabled = false",
		blockBegin,
		`[mcp_servers."reader"]`,
		`[mcp_servers."local"]`,
		`default_tools_approval_mode = "approve"`,
		filepath.Join(home, "bin", "server"),
	} {
		if !strings.Contains(codex, want) {
			t.Fatalf("Codex projection missing %q:\n%s", want, codex)
		}
	}

	homeInventory, err := os.ReadFile(filepath.Join(home, ".mcporter", "mcporter.json"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(inventoryPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(homeInventory) != string(source) {
		t.Fatalf("home inventory differs from canonical source")
	}

	result, err = Project(Options{Inventory: inventoryPath, Home: home, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Changed {
		t.Fatalf("second projection drifted: %+v", result)
	}
}

func TestProjectCheckReportsDriftWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	inventoryPath := filepath.Join(dir, "mcporter.json")
	write(t, inventoryPath, `{"imports":[],"mcpServers":{"reader":{"baseUrl":"https://mcp.example.test/mcp"}}}`+"\n")

	result, err := Project(Options{Inventory: inventoryPath, Home: home, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed {
		t.Fatal("check must report missing projections as drift")
	}
	if _, err := os.Stat(filepath.Join(home, ".claude.json")); !os.IsNotExist(err) {
		t.Fatalf("check wrote Claude config: %v", err)
	}
}

func TestProjectRejectsReverseImports(t *testing.T) {
	dir := t.TempDir()
	inventoryPath := filepath.Join(dir, "mcporter.json")
	write(t, inventoryPath, `{"mcpServers":{"reader":{"baseUrl":"https://mcp.example.test/mcp"}}}`+"\n")

	_, err := Project(Options{Inventory: inventoryPath, Home: filepath.Join(dir, "home")})
	if err == nil || !strings.Contains(err.Error(), "imports") {
		t.Fatalf("missing imports=[] error = %v", err)
	}
}

func TestProjectRejectsUnsupportedCodexApprovalMode(t *testing.T) {
	dir := t.TempDir()
	inventoryPath := filepath.Join(dir, "mcporter.json")
	write(t, inventoryPath, `{"imports":[],"mcpServers":{"reader":{"baseUrl":"https://mcp.example.test/mcp","x-codex":{"defaultToolsApprovalMode":"always"}}}}`+"\n")

	_, err := Project(Options{Inventory: inventoryPath, Home: filepath.Join(dir, "home")})
	if err == nil || !strings.Contains(err.Error(), `unsupported Codex approval mode "always"`) {
		t.Fatalf("unsupported approval mode error = %v", err)
	}
}
