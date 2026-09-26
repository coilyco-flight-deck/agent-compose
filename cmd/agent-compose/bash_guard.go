package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/coilyco-flight-deck/agent-compose/v2/internal/bashguard"
	"github.com/urfave/cli/v3"
)

// hookCommand holds the harness hooks a launch wires into its settings fragment.
var hookCommand = &cli.Command{
	Name:  "hook",
	Usage: "harness hooks a native launch wires into its settings",
	Commands: []*cli.Command{{
		Name:  "bash-guard",
		Usage: "PreToolUse hook: refuse bare kubectl and helm outside the role's allowed verbs",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "allow-kubectl",
				Usage: "a kubectl verb this role may still run bare",
			},
		},
		Action: runBashGuard,
	}},
}

// runBashGuard exits 2 to block, which Claude Code shows the model. Unreadable
// input blocks too, since a guard that cannot see the command cannot pass it.
func runBashGuard(_ context.Context, cmd *cli.Command) error {
	var input struct {
		ToolName  string `json:"tool_name"`
		ToolInput struct {
			Command string `json:"command"`
		} `json:"tool_input"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		return cli.Exit(fmt.Sprintf("agent-compose bash-guard: unreadable hook input: %v", err), 2)
	}
	if input.ToolName != "Bash" {
		return nil
	}
	policy := bashguard.Policy{KubectlVerbs: cmd.StringSlice("allow-kubectl")}
	if reason := bashguard.Check(input.ToolInput.Command, policy); reason != "" {
		return cli.Exit(reason, 2)
	}
	return nil
}
