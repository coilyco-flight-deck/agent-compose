package main

import (
	"io"
	"reflect"
	"testing"
)

func TestNativeHarnessCommandScopesClaudeMCP(t *testing.T) {
	t.Parallel()
	got := nativeHarnessCommand("claude", []string{"--model", "opus"}, nativeIdentity{
		MCP: mcpLaunch{ClaudeConfig: "/state/mcp/platform-eng-abc.json"},
	})
	want := []string{
		"claude",
		"--strict-mcp-config", "--mcp-config", "/state/mcp/platform-eng-abc.json",
		"--model", "opus",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}

func TestNativeHarnessCommandDisablesOmittedCodexServers(t *testing.T) {
	t.Parallel()
	overrides := []string{"-c", "mcp_servers.pw_x.enabled=false"}
	got := nativeHarnessCommand("codex", []string{"resume"}, nativeIdentity{MCP: mcpLaunch{CodexArgs: overrides}})
	want := []string{"codex", "-c", "mcp_servers.pw_x.enabled=false", "resume"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}

// No person or an unscoped harness leaves the full registry in place.
func TestNativeMCPScopeUnscopedWithoutPerson(t *testing.T) {
	t.Parallel()
	for _, harness := range []string{"claude", "codex", "goose"} {
		got, err := nativeMCPScope(io.Discard, nil, "platform-eng", harness, t.TempDir(), nil)
		if err != nil || !reflect.DeepEqual(got, mcpLaunch{}) {
			t.Fatalf("%s: scope = %#v, %v; want unscoped", harness, got, err)
		}
	}
}
