package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/coilyco-flight-deck/agent-compose/v2/internal/mcpscope"
	"github.com/coilyco-flight-deck/agent-compose/v2/internal/person"
	"github.com/urfave/cli/v3"
)

// mcpLaunch is the per-harness form of one role's MCP selection.
type mcpLaunch struct {
	ClaudeConfig string
	CodexArgs    []string
}

// mcpInventoryHome is the home the host projection expands ${HOME} against, so a
// scoped server launches exactly as the host registry would launch it.
func mcpInventoryHome() (string, error) {
	if h := strings.TrimSpace(os.Getenv("AOS_NATIVE_CANONICAL_HOME")); h != "" {
		return h, nil
	}
	return os.UserHomeDir()
}

func loadMCPInventory(p *person.Person, path string) (*mcpscope.Inventory, error) {
	entries, err := p.RoleCatalog()
	if err != nil {
		return nil, err
	}
	roles := make([]string, 0, len(entries))
	for _, e := range entries {
		roles = append(roles, e.Slug)
	}
	return mcpscope.Load(path, roles)
}

// nativeMCPScope narrows the host MCP inventory to role for one launch. Any gap
// leaves it unscoped rather than leaving a seat with no servers.
func nativeMCPScope(w io.Writer, p *person.Person, role, harness, stateDir string, args []string) (mcpLaunch, error) {
	if p == nil || (harness != "claude" && harness != "codex") {
		return mcpLaunch{}, nil
	}
	if harness == "claude" && (nativeArgsCarry(args, "--mcp-config") || nativeArgsCarry(args, "--strict-mcp-config")) {
		return mcpLaunch{}, nil
	}
	home, err := mcpInventoryHome()
	if err != nil {
		return mcpLaunch{}, fmt.Errorf("resolve MCP inventory home: %w", err)
	}
	inv, err := loadMCPInventory(p, filepath.Join(home, ".mcporter", "mcporter.json"))
	if errors.Is(err, mcpscope.ErrNoInventory) {
		return mcpLaunch{}, nil
	}
	if err != nil {
		return mcpLaunch{}, err
	}
	sel := inv.Select(role)
	fmt.Fprintln(w, sel.Summary())
	if harness == "codex" {
		return mcpLaunch{CodexArgs: sel.CodexOverrides()}, nil
	}
	path, err := sel.WriteClaude(filepath.Join(stateDir, "mcp"), home)
	if err != nil {
		return mcpLaunch{}, err
	}
	return mcpLaunch{ClaudeConfig: path}, nil
}

// runMCP reports one role's resolved MCP set without launching anything.
func runMCP(_ context.Context, cmd *cli.Command) error {
	role := strings.TrimSpace(cmd.String("role"))
	p, _, err := loadSelectedPersonWithLibraries(cmd.String("person-source"), nil)
	if err != nil {
		return err
	}
	path := cmd.String("inventory")
	if path == "" {
		home, err := mcpInventoryHome()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".mcporter", "mcporter.json")
	}
	inv, err := loadMCPInventory(p, path)
	if err != nil {
		return err
	}
	sel := inv.Select(role)
	w := cmd.Root().Writer
	fmt.Fprintln(w, strings.TrimPrefix(sel.Summary(), "agent-compose: "))
	for _, name := range sel.Selected {
		fmt.Fprintf(w, "+ %s\n", name)
	}
	for _, name := range sel.Omitted {
		fmt.Fprintf(w, "- %s\n", name)
	}
	return nil
}
