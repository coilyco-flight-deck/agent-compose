package main

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/coilyco-flight-deck/agent-compose/v2/internal/mcpscope"
	"github.com/coilyco-flight-deck/agent-compose/v2/internal/person"
)

func TestNativeHarnessCommandScopesClaudeMCP(t *testing.T) {
	t.Parallel()
	got := nativeHarnessCommand("claude", []string{"--model", "opus"}, nativeIdentity{
		MCP: mcpLaunch{ClaudeConfig: "/state/mcp/eng-platform-abc.json"},
	})
	want := []string{
		"claude",
		"--strict-mcp-config", "--mcp-config", "/state/mcp/eng-platform-abc.json",
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

// A seat with no scope would load the whole user-level MCP set, so claude and
// codex refuse rather than launch unscoped. Goose has no scope to apply.
func TestNativeMCPScopeFailsClosed(t *testing.T) {
	t.Setenv("AOS_NATIVE_CANONICAL_HOME", t.TempDir())
	noRoster := func() (*person.Person, error) { return nil, errors.New("roster unreadable") }
	for _, harness := range []string{"claude", "codex"} {
		if _, err := nativeMCPScope(io.Discard, noRoster, "eng-platform", harness, t.TempDir(), nil); err == nil {
			t.Fatalf("%s: an unreadable roster should refuse", harness)
		}
		if _, err := nativeMCPScope(io.Discard, person.Load, "eng-platform", harness, t.TempDir(), nil); !errors.Is(err, mcpscope.ErrNoInventory) {
			t.Fatalf("%s: a host with no inventory should refuse, got %v", harness, err)
		}
	}
	got, err := nativeMCPScope(io.Discard, noRoster, "eng-platform", "goose", t.TempDir(), nil)
	if err != nil || !reflect.DeepEqual(got, mcpLaunch{}) {
		t.Fatalf("goose: scope = %#v, %v; want unscoped", got, err)
	}
}

// A caller's own MCP config is a scope of its own, so no roster is read.
func TestNativeMCPScopeKeepsACallerScope(t *testing.T) {
	t.Parallel()
	unread := func() (*person.Person, error) { t.Fatal("the roster should not be read"); return nil, nil }
	got, err := nativeMCPScope(io.Discard, unread, "eng-platform", "claude", t.TempDir(), []string{"--mcp-config", "mine.json"})
	if err != nil || !reflect.DeepEqual(got, mcpLaunch{}) {
		t.Fatalf("scope = %#v, %v; want the caller's own", got, err)
	}
}
