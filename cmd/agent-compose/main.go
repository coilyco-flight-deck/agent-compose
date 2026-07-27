package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/converge"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/describe"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/evaluation"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/home"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/launch"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/nativemcp"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/overlay"
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
					&cli.BoolFlag{
						Name:  "reapply",
						Usage: "rewrite the host compose layout even when it is current",
					},
					&cli.BoolFlag{
						Name:  "verbose",
						Usage: "print every host compose source => destination mapping",
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
				Name:  "mcp",
				Usage: "project one mcporter inventory into Claude and Codex native MCP registries",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "inventory",
						Usage:    "canonical mcporter.json source",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "home",
						Usage: "home receiving mcporter and native harness configs",
					},
					&cli.BoolFlag{
						Name:  "check",
						Usage: "report drift without writing",
					},
				},
				Action: runNativeMCP,
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
				Name:  "evaluation",
				Usage: "emit the four-case frontier and OSS human-review matrix",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "role",
						Value: "engineer",
						Usage: "canonical role under evaluation",
					},
					&cli.StringFlag{
						Name:  "seat",
						Value: "codex",
						Usage: "harness seat held constant across the comparison",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "markdown",
						Usage: "output format: markdown or yaml",
					},
					&cli.StringFlag{
						Name:  "person-source",
						Usage: "external person-package root (defaults to embedded person:kai)",
					},
				},
				Action: runEvaluation,
			},
			{
				Name:  "overlay",
				Usage: "project one member identity and caller-supplied expression",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "role",
						Usage:    "canonical role",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "seat",
						Usage:    "harness seat within the role",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "expression",
						Value: "available",
						Usage: "renderer expression supplied by the caller",
					},
					&cli.IntFlag{
						Name:  "width",
						Value: 40,
						Usage: "maximum text width",
					},
					&cli.BoolFlag{
						Name:  "json",
						Usage: "emit the versioned renderer document",
					},
					&cli.StringFlag{
						Name:  "person-source",
						Usage: "external person-package root (defaults to embedded person:kai)",
					},
				},
				Action: runOverlay,
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
				ArgsUsage: "[source-path...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out",
						Usage: "artifact directory (defaults to ~/.agent-compose/sources/personality)",
					},
					&cli.StringFlag{
						Name:  "person-source",
						Usage: "external person-package root (defaults to embedded person:kai)",
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
					&cli.StringFlag{
						Name:  "person-source",
						Usage: "external person-package root (defaults to embedded person:kai)",
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

func runEvaluation(_ context.Context, cmd *cli.Command) error {
	p, external, err := loadSelectedPerson(cmd.String("person-source"))
	if err != nil {
		return err
	}
	var pack *evaluation.Pack
	if external {
		pack, err = evaluation.BuildFor(p, cmd.String("role"), cmd.String("seat"))
	} else {
		pack, err = evaluation.Build(cmd.String("role"), cmd.String("seat"))
	}
	if err != nil {
		return err
	}
	raw, err := evaluationOutput(pack, cmd.String("format"))
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(raw)
	return err
}

func evaluationOutput(pack *evaluation.Pack, format string) ([]byte, error) {
	switch format {
	case "markdown":
		return evaluation.Markdown(pack), nil
	case "yaml":
		return evaluation.MarshalYAML(pack)
	default:
		return nil, fmt.Errorf("evaluation --format must be markdown or yaml, got %q", format)
	}
}

func runNativeMCP(_ context.Context, cmd *cli.Command) error {
	result, err := nativemcp.Project(nativemcp.Options{
		Inventory: cmd.String("inventory"),
		Home:      cmd.String("home"),
		Check:     cmd.Bool("check"),
	})
	if err != nil {
		return err
	}
	state := "unchanged"
	if result.Changed {
		state = "changed"
	}
	fmt.Printf("native-mcp servers=%d state=%s\n", result.Servers, state)
	if cmd.Bool("check") && result.Changed {
		return cli.Exit("native MCP projection drift", 1)
	}
	return nil
}

func runOverlay(_ context.Context, cmd *cli.Command) error {
	p, _, err := loadSelectedPerson(cmd.String("person-source"))
	if err != nil {
		return err
	}
	doc, err := overlay.Build(
		p,
		cmd.String("role"),
		cmd.String("seat"),
		cmd.String("expression"),
	)
	if err != nil {
		return err
	}
	if cmd.Bool("json") {
		raw, err := overlay.Marshal(doc)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(raw)
		return err
	}
	rendered, err := overlay.RenderText(doc, cmd.Int("width"))
	if err != nil {
		return err
	}
	fmt.Print(rendered)
	return nil
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
	if requestPath != "" && (cmd.Bool("reapply") || cmd.Bool("verbose")) {
		return fmt.Errorf("--reapply and --verbose apply only to host convergence without a request")
	}

	if len(command) > 0 {
		return refreshThenExec(cmd, requestPath, command)
	}
	if cmd.String("layout") != "" {
		return fmt.Errorf("--layout drives the exec path; add `-- <command>` (or use the project verb)")
	}
	if requestPath == "" {
		if code := converge.Run(cascade.DefaultPaths(), converge.Options{
			Reapply: cmd.Bool("reapply"),
			Verbose: cmd.Bool("verbose"),
		}, os.Stdout, os.Stderr); code != 0 {
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
	summaryOpts := person.RoleTranscriptOptions{
		Color:     colorEnabled(),
		TrueColor: trueColorTerminal(),
	}
	if err := printSummary(os.Stdout, result, summaryOpts); err != nil {
		return err
	}
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
	printVerification(os.Stdout, verification)
	return nil
}

func printVerification(w io.Writer, verification *bundle.Verification) {
	fmt.Fprintf(w, "bundle verified: %d skills // %d files\n",
		len(verification.Identities), verification.Files)
}

// refreshThenExec is the absorbed launch verb: refresh context, then hand
// the process to the real command, sentinel-guarded against recursion.
func refreshThenExec(cmd *cli.Command, requestPath string, command []string) error {
	if os.Getenv(launch.EnvSentinel) != "" {
		fmt.Fprintln(os.Stderr, "agent-compose: nested launch detected; skipping refresh")
		return execReal(command)
	}
	if requestPath == "" {
		if code := converge.Run(cascade.DefaultPaths(), converge.Options{
			Reapply: cmd.Bool("reapply"),
			Verbose: cmd.Bool("verbose"),
		}, os.Stdout, os.Stderr); code != 0 {
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
		code = cascade.Run(paths, cascade.RunOptions{DryRun: cmd.Bool("dry-run")}, os.Stdout, os.Stderr)
	}
	if code != 0 {
		return cli.Exit("", code)
	}
	return nil
}

func runRoster(_ context.Context, cmd *cli.Command) error {
	p, _, err := loadSelectedPerson(cmd.String("person-source"))
	if err != nil {
		return err
	}
	var sources []*schema.Source
	for _, sourcePath := range cmd.Args().Slice() {
		src, err := schema.LoadSource(sourcePath)
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
	p, _, err := loadSelectedPerson(cmd.String("person-source"))
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

func loadSelectedPerson(source string) (*person.Person, bool, error) {
	if source == "" {
		p, err := person.Load()
		return p, false, err
	}
	p, err := person.LoadDirectory(source)
	return p, true, err
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

// printSummary renders the complete selected-role identity transcript followed
// by bounded composition audit counts.
func printSummary(
	w io.Writer,
	r *compose.Result,
	opts person.RoleTranscriptOptions,
) error {
	state := "new"
	if r.Bundle.Reused {
		state = "reused"
	}
	counts := map[string]int{}
	for _, d := range r.Resolution.Decisions {
		counts[d.Outcome]++
	}
	req := r.Resolution.Request
	intro := fmt.Sprintf(
		"bundle %s (%s)\nrequest: model class %s // delivery %s\n",
		r.Bundle.Key, state, req.ModelClass, req.Delivery,
	)
	if opts.Color {
		intro = color.ANSI(r.Resolution.FavoriteColor, intro, opts.TrueColor)
	}
	fmt.Fprintln(w, intro)
	metadata, err := r.Resolution.Person.RenderRoleTranscript(
		req.Role,
		r.Resolution.FavoriteColor,
		opts,
	)
	if err != nil {
		return err
	}
	fmt.Fprint(w, metadata)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "sources: %s\n", strings.Join(r.Resolution.SourceIDs, ", "))
	fmt.Fprintf(w, "decisions: %d selected // %d excluded // %d shadowed // %d delivered\n",
		counts[resolver.OutcomeSelected], counts[resolver.OutcomeExcluded],
		counts[resolver.OutcomeShadowed], counts[resolver.OutcomeDelivered])
	fmt.Fprintf(w, "path: %s\n", r.Bundle.Dir)
	fmt.Fprintf(w, "trace: %s\n", filepath.Join(r.Bundle.Dir, "trace.json"))
	return nil
}
