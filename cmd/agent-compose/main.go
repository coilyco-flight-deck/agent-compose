package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/describe"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/home"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/launch"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/roster"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

// version is stamped by the release build via -ldflags; dev builds say dev.
var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:    "agent-compose",
		Usage:   "compose personality context into an immutable bundle",
		Version: version,
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "print the build version",
				Action: func(_ context.Context, _ *cli.Command) error {
					fmt.Println(version)
					return nil
				},
			},
			{
				Name:      "compose",
				Usage:     "compose a request KDL file into a bundle",
				ArgsUsage: "<request.kdl>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out",
						Usage: "bundle output directory (defaults to ~/.agent-compose/bundles)",
					},
					&cli.BoolFlag{
						Name:  "explain",
						Usage: "print the full decision tree after the summary",
					},
				},
				Action: runCompose,
			},
			{
				Name:      "describe",
				Usage:     "render a bundle's stored decision tree",
				ArgsUsage: "<bundle-dir>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "why",
						Usage: "follow one item (e.g. skill:personality-curious) to its outcome",
					},
					&cli.BoolFlag{
						Name:  "all",
						Usage: "expand collapsed exclusion groups",
					},
				},
				Action: runDescribe,
			},
			{
				Name:      "diff",
				Usage:     "report semantic decision changes between two bundles",
				ArgsUsage: "<left-bundle> <right-bundle>",
				Action:    runDiff,
			},
			{
				Name:  "cascade",
				Usage: "compose doctrine sources into harness global load points",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "print the plan, change nothing",
					},
					&cli.BoolFlag{
						Name:  "check",
						Usage: "verify composed outputs are in sync; exit 1 on drift",
					},
				},
				Action: runCascade,
			},
			{
				Name:      "roster",
				Usage:     "render the seat dispatch table as a v1-cascade source",
				ArgsUsage: "[source.kdl...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out",
						Usage: "artifact directory (defaults to ~/.agent-compose/sources/personality)",
					},
				},
				Action: runRoster,
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
					&cli.StringFlag{
						Name:  "scope",
						Value: project.ScopeRepo,
						Usage: "load-point scope: repo (project files) or home (global files)",
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
						Usage: "bundle output directory (defaults to ~/.agent-compose/bundles)",
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
		stateDir, err := home.Dir()
		if err != nil {
			return fmt.Errorf("no --out given and no state dir: %w", err)
		}
		outDir = filepath.Join(stateDir, "bundles")
	}
	result, err := compose.Run(requestPath, outDir)
	if err != nil {
		return err
	}
	printSummary(result)
	if cmd.Bool("explain") {
		rendered, err := describe.Bundle(result.Bundle.Dir, describe.Options{All: true, Color: colorEnabled(), TrueColor: trueColorTerminal()})
		if err != nil {
			return err
		}
		fmt.Print(rendered)
	}
	return nil
}

func runDescribe(_ context.Context, cmd *cli.Command) error {
	dir := cmd.Args().First()
	if dir == "" {
		return fmt.Errorf("describe needs a bundle directory")
	}
	opts := describe.Options{All: cmd.Bool("all"), Color: colorEnabled(), TrueColor: trueColorTerminal()}
	var rendered string
	var err error
	if why := cmd.String("why"); why != "" {
		rendered, err = describe.Why(dir, why, opts)
	} else {
		rendered, err = describe.Bundle(dir, opts)
	}
	if err != nil {
		return err
	}
	fmt.Print(rendered)
	return nil
}

func runDiff(_ context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 2 {
		return fmt.Errorf("diff needs exactly two bundle directories")
	}
	rendered, err := describe.Diff(cmd.Args().Get(0), cmd.Args().Get(1))
	if err != nil {
		return err
	}
	fmt.Print(rendered)
	return nil
}

// colorEnabled keeps redirected output plain and deterministic; color is a
// TTY-only affordance and NO_COLOR always wins.
func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func trueColorTerminal() bool {
	ct := os.Getenv("COLORTERM")
	return strings.Contains(ct, "truecolor") || strings.Contains(ct, "24bit")
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
		stateDir, err := home.Dir()
		if err != nil {
			return fmt.Errorf("no --out given and no state dir: %w", err)
		}
		outDir = filepath.Join(stateDir, "bundles")
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

func runCascade(_ context.Context, cmd *cli.Command) error {
	if cmd.Bool("dry-run") && cmd.Bool("check") {
		return fmt.Errorf("--dry-run and --check are mutually exclusive")
	}
	paths := cascade.DefaultPaths()
	var code int
	if cmd.Bool("check") {
		code = cascade.Check(paths, os.Stdout, os.Stderr)
	} else {
		code = cascade.Run(paths, cmd.Bool("dry-run"), os.Stdout, os.Stderr)
	}
	if code != 0 {
		return cli.Exit("", code)
	}
	return nil
}

func runRoster(_ context.Context, cmd *cli.Command) error {
	p, err := person.Load()
	if err != nil {
		return err
	}
	var sources []*schema.Source
	for _, declPath := range cmd.Args().Slice() {
		src, err := schema.LoadSource(declPath)
		if err != nil {
			return err
		}
		sources = append(sources, src)
	}
	outFlag := cmd.String("out")
	if outFlag == "" {
		stateDir, err := home.Dir()
		if err != nil {
			return fmt.Errorf("no --out given and no state dir: %w", err)
		}
		outFlag = filepath.Join(stateDir, "sources", "personality")
	}
	outDir, err := filepath.Abs(outFlag)
	if err != nil {
		return err
	}
	files, err := roster.Render(p, sources, outDir)
	if err != nil {
		return err
	}
	result, err := project.ApplyOwned(outDir, files, "roster", "person:"+p.Name)
	if err != nil {
		return err
	}
	fmt.Printf("roster artifact: %d files under %s\n", len(result.Files), outDir)
	return nil
}

func runProject(_ context.Context, cmd *cli.Command) error {
	bundleDir := cmd.Args().First()
	if bundleDir == "" {
		return fmt.Errorf("project needs a bundle directory")
	}
	result, err := project.ProjectScoped(bundleDir, cmd.String("layout"), cmd.String("target"), cmd.String("scope"))
	if err != nil {
		return err
	}
	fmt.Printf("projected %d files into layout %s (%s scope) under %s\n",
		len(result.Files), result.Layout, cmd.String("scope"), cmd.String("target"))
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
