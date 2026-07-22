package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
)

func main() {
	cmd := &cli.Command{
		Name:  "agent-compose",
		Usage: "compose personality context into an immutable bundle",
		Commands: []*cli.Command{
			{
				Name:      "compose",
				Usage:     "compose a request KDL file into a bundle",
				ArgsUsage: "<request.kdl>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out",
						Usage: "bundle output directory (defaults to the user cache)",
					},
				},
				Action: runCompose,
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "agent-compose: %v\n", err)
		os.Exit(1)
	}
}

func runCompose(_ context.Context, cmd *cli.Command) error {
	requestPath := cmd.Args().First()
	if requestPath == "" {
		return fmt.Errorf("compose needs a request KDL path")
	}
	outDir := cmd.String("out")
	if outDir == "" {
		cache, err := os.UserCacheDir()
		if err != nil {
			return fmt.Errorf("no --out given and no user cache dir: %w", err)
		}
		outDir = filepath.Join(cache, "agent-compose", "bundles")
	}
	result, err := compose.Run(requestPath, outDir)
	if err != nil {
		return err
	}
	printSummary(result)
	return nil
}

// printSummary renders the bounded one-screen composition summary: identity
// on one line, decision counts instead of per-item listings.
func printSummary(r *compose.Result) {
	state := "new"
	if r.Bundle.Reused {
		state = "reused"
	}
	counts := map[string]int{}
	for _, d := range r.Resolution.Decisions {
		counts[d.Outcome]++
	}
	req := r.Resolution.Request
	fmt.Printf("bundle %s (%s)\n", r.Bundle.Key, state)
	fmt.Printf("  role %s · personality %s · delivery %s · density %s\n",
		req.Role, req.Personality, req.Delivery, req.Density)
	fmt.Printf("  sources: %s\n", strings.Join(r.Resolution.SourceIDs, ", "))
	fmt.Printf("  decisions: %d selected · %d excluded · %d shadowed · %d fallback · %d delivered\n",
		counts[resolver.OutcomeSelected], counts[resolver.OutcomeExcluded],
		counts[resolver.OutcomeShadowed], counts[resolver.OutcomeFallback],
		counts[resolver.OutcomeDelivered])
	fmt.Printf("  path: %s\n", r.Bundle.Dir)
	fmt.Printf("  trace: %s\n", filepath.Join(r.Bundle.Dir, "trace.json"))
}
