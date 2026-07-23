package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/converge"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/describe"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/home"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/launch"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/palette"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/roster"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

// version is stamped by the release build via -ldflags; dev builds say dev.
var version = "dev"

// dispatchArgs makes the acompose install name behave as the compose verb,
// so the daily command is dash-free and stutter-free.
func dispatchArgs(args []string) []string {
	base := args[0]
	if i := strings.LastIndexAny(base, `/\`); i >= 0 {
		base = base[i+1:]
	}
	if strings.TrimSuffix(base, ".exe") != "acompose" {
		return args
	}
	return append([]string{args[0], "compose"}, args[1:]...)
}

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
				Usage:     "converge the host, compose a bundle, or refresh then exec after --",
				ArgsUsage: "[request.kdl] [-- <command> [args...]]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out",
						Usage: "bundle output directory (defaults to ~/.agent-compose/bundles)",
					},
					&cli.BoolFlag{
						Name:  "explain",
						Usage: "print the full decision tree after the summary",
					},
					&cli.StringFlag{
						Name:  "layout",
						Usage: "load-point layout for the exec path (with a request)",
					},
					&cli.StringFlag{
						Name:  "target",
						Value: ".",
						Usage: "directory receiving the load points on the exec path",
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
				Name:      "verify",
				Usage:     "verify that a bundle is complete and safe to consume",
				ArgsUsage: "<bundle-dir>",
				Action:    runVerify,
			},
			{
				Name:   "cascade",
				Hidden: true,
				Usage:  "compose doctrine sources into harness global load points",
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
				Hidden:    true,
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
				Name:   "palette-data",
				Hidden: true,
				Usage:  "render canonical personality and role colors for the local palette explorer",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "out",
						Usage:    "write JSON to this path instead of stdout",
						Required: false,
					},
				},
				Action: runPaletteData,
			},
			{
				Name:      "project",
				Usage:     "transactionally place a verified bundle at harness load points",
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
		},
	}
	if err := cmd.Run(context.Background(), dispatchArgs(os.Args)); err != nil {
		fmt.Fprintf(os.Stderr, "agent-compose: %v\n", err)
		os.Exit(1)
	}
}

func runCompose(_ context.Context, cmd *cli.Command) error {
	command := argvAfterDash()
	args := cmd.Args().Slice()
	if len(command) > 0 && len(args) >= len(command) {
		args = args[:len(args)-len(command)]
	}
	requestPath := ""
	if len(args) > 0 {
		requestPath = args[0]
	}

	if len(command) > 0 {
		return refreshThenExec(cmd, requestPath, command)
	}
	if cmd.String("layout") != "" {
		return fmt.Errorf("--layout drives the exec path; add `-- <command>` (or use the project verb)")
	}
	if requestPath == "" {
		if code := converge.Run(cascade.DefaultPaths(), os.Stdout, os.Stderr); code != 0 {
			return cli.Exit("", code)
		}
		return nil
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

func runVerify(_ context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 1 {
		return fmt.Errorf("verify needs exactly one bundle directory")
	}
	verification, err := bundle.Verify(cmd.Args().First())
	if err != nil {
		return err
	}
	manifest := verification.Manifest
	identities := make([]string, 0, len(verification.Identities))
	for _, identity := range verification.Identities {
		identities = append(identities, identity.Source+"/"+identity.Skill)
	}
	fmt.Printf("bundle verified: role=%s personalities=%s delivery=%s identities=%s files=%d\n",
		manifest.Role, strings.Join(manifest.Personalities, ","), manifest.Delivery.Mode,
		strings.Join(identities, ","), verification.Files)
	return nil
}

// refreshThenExec is the absorbed launch verb: refresh context, then hand
// the process to the real command, sentinel-guarded against recursion.
func refreshThenExec(cmd *cli.Command, requestPath string, command []string) error {
	if os.Getenv(launch.EnvSentinel) != "" {
		fmt.Fprintln(os.Stderr, "agent-compose: nested launch detected; skipping refresh")
		return execReal(command)
	}
	if requestPath == "" {
		if code := converge.Run(cascade.DefaultPaths(), os.Stdout, os.Stderr); code != 0 {
			return cli.Exit("", code)
		}
		return execReal(command)
	}
	layout := cmd.String("layout")
	if layout == "" {
		return fmt.Errorf("a request with `--` needs --layout to place its load points")
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
		RequestPath: requestPath,
		Layout:      layout,
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
	return execReal(command)
}

// argvAfterDash finds the raw command after the `--` terminator, independent
// of how the flag parser folds those tokens into Args.
func argvAfterDash() []string {
	for i, arg := range os.Args {
		if arg == "--" {
			return os.Args[i+1:]
		}
	}
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

func runPaletteData(_ context.Context, cmd *cli.Command) error {
	p, err := person.Load()
	if err != nil {
		return err
	}
	raw, err := palette.Marshal(p)
	if err != nil {
		return err
	}
	if out := cmd.String("out"); out != "" {
		return palette.Write(out, raw)
	}
	_, err = os.Stdout.Write(raw)
	return err
}

func runProject(_ context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 1 {
		return fmt.Errorf("project needs exactly one bundle directory")
	}
	bundleDir := cmd.Args().First()
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
	fmt.Printf("  role %s · personalities %s · melded %s · delivery %s · density %s\n",
		req.Role, strings.Join(r.Resolution.Personalities, ", "), r.Resolution.FavoriteColor,
		req.Delivery, req.Density)
	fmt.Printf("  sources: %s\n", strings.Join(r.Resolution.SourceIDs, ", "))
	fmt.Printf("  decisions: %d selected · %d excluded · %d shadowed · %d fallback · %d delivered\n",
		counts[resolver.OutcomeSelected], counts[resolver.OutcomeExcluded],
		counts[resolver.OutcomeShadowed], counts[resolver.OutcomeFallback],
		counts[resolver.OutcomeDelivered])
	fmt.Printf("  path: %s\n", r.Bundle.Dir)
	fmt.Printf("  trace: %s\n", filepath.Join(r.Bundle.Dir, "trace.json"))
}
