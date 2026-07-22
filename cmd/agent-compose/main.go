package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/urfave/cli/v3"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/launch"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
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
			{
				Name:      "project",
				Usage:     "place bundle content at a harness layout's load points",
				ArgsUsage: "<bundle-dir>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "layout",
						Usage:    "load-point layout name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "target",
						Value: ".",
						Usage: "directory receiving the load points",
					},
				},
				Action: runProject,
			},
			{
				Name:      "launch",
				Usage:     "refresh context, then exec the real command",
				ArgsUsage: "-- <command> [args...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "request",
						Usage:    "compose request KDL path",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "layout",
						Usage:    "load-point layout name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "target",
						Value: ".",
						Usage: "directory receiving the load points",
					},
					&cli.StringFlag{
						Name:  "out",
						Usage: "bundle output directory (defaults to the user cache)",
					},
				},
				Action: runLaunch,
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

func runLaunch(_ context.Context, cmd *cli.Command) error {
	argv := cmd.Args().Slice()
	if len(argv) == 0 {
		return fmt.Errorf("launch needs a command after --")
	}
	if os.Getenv(launch.EnvSentinel) != "" {
		fmt.Fprintln(os.Stderr, "agent-compose: nested launch detected; skipping refresh")
		return execReal(argv)
	}
	outDir := cmd.String("out")
	if outDir == "" {
		cache, err := os.UserCacheDir()
		if err != nil {
			return fmt.Errorf("no --out given and no user cache dir: %w", err)
		}
		outDir = filepath.Join(cache, "agent-compose", "bundles")
	}
	result, err := launch.Refresh(launch.Options{
		RequestPath: cmd.String("request"),
		Layout:      cmd.String("layout"),
		TargetDir:   cmd.String("target"),
		OutDir:      outDir,
	})
	if err != nil {
		return err
	}
	if result.Fallback {
		fmt.Fprintf(os.Stderr, "agent-compose: WARNING: refresh failed (%s); launching with the last-known-good projection\n", result.Warning)
	} else {
		state := "new"
		if result.BundleReused {
			state = "reused"
		}
		fmt.Fprintf(os.Stderr, "agent-compose: context refreshed (%s bundle, %d files) for %s\n",
			state, result.Projected, cmd.String("target"))
	}
	return execReal(argv)
}

// execReal replaces this process with the target command; the sentinel in
// the child environment is the wrapper-recursion guard.
func execReal(argv []string) error {
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return fmt.Errorf("launch target: %w", err)
	}
	env := append(os.Environ(), launch.EnvSentinel+"=1")
	return syscall.Exec(path, argv, env)
}

func runProject(_ context.Context, cmd *cli.Command) error {
	bundleDir := cmd.Args().First()
	if bundleDir == "" {
		return fmt.Errorf("project needs a bundle directory")
	}
	result, err := project.Project(bundleDir, cmd.String("layout"), cmd.String("target"))
	if err != nil {
		return err
	}
	fmt.Printf("projected %d files into layout %s under %s\n",
		len(result.Files), result.Layout, cmd.String("target"))
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
