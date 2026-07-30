package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/converge"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/describe"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/evaluation"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/home"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/launch"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/nativelaunch"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/nativemcp"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/overlay"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/palette"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/personpolicy"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/roster"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

// version is stamped by the release build via -ldflags; dev builds say dev.
var version = "dev"

const nativeCodexIntroductionPrompt = "Introduce yourself now as the active Codex seat for your assigned role. Use your loaded identity card and personality meld, keep the introduction warm and concise, then ask what the user would like to work on."

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
	if len(args) >= 3 && !strings.HasPrefix(args[1], "-") &&
		nativeHarness(args[2]) {
		return append([]string{args[0], "launch"}, args[1:]...)
	}
	return append([]string{args[0], "compose"}, args[1:]...)
}

func nativeHarness(value string) bool {
	switch value {
	case "claude", "codex", "goose", "opencode":
		return true
	default:
		return false
	}
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
				Name:            "launch",
				Usage:           "launch one native harness with a caller-assigned role bundle",
				ArgsUsage:       "<role> <harness> [harness arguments...]",
				SkipFlagParsing: true,
				Action:          runNativeLaunch,
			},
			{
				Name: "bundle", Usage: "inspect or export verified bundles",
				Commands: []*cli.Command{{
					Name: "export", Usage: "write a deterministic verified .tar.gz archive", ArgsUsage: "<bundle-dir>",
					Flags:  []cli.Flag{&cli.StringFlag{Name: "out", Required: true, Usage: "archive output path"}},
					Action: runBundleExport,
				}},
			},
			{
				Name:  "catalog",
				Usage: "inspect the selected local profile catalogue",
				Commands: []*cli.Command{
					{
						Name: "personalities", Usage: "list personalities or resolve one cue",
						Description: "JSON items: slug, skill, description, aliases, identity primitives, source_library, digest, and affinities.",
						Flags:       personCatalogFlags(true), Action: runCatalogPersonalities,
					},
					{
						Name: "roles", Usage: "list profile roles",
						Description: "JSON items: slug, purpose, role skill provenance, seats, ordered personalities, and favorite_color.",
						Flags:       personCatalogFlags(false), Action: runCatalogRoles,
					},
					{
						Name: "seats", Usage: "list profile seats",
						Description: "JSON items: role plus the complete stable seat key, name, pronouns, channel, and tier.",
						Flags:       append(personCatalogFlags(false), &cli.StringFlag{Name: "role", Usage: "limit to one role"}), Action: runCatalogSeats,
					},
					{
						Name: "expressions", Usage: "list stable expression vocabulary",
						Description: "JSON items are the stable expression strings.",
						Flags:       []cli.Flag{&cli.BoolFlag{Name: "json", Usage: "emit agent-compose.catalog.v1 JSON"}}, Action: runCatalogExpressions,
					},
				},
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
					&cli.StringSliceFlag{
						Name:  "personality-library",
						Usage: "additional local personality-library root (repeatable)",
					},
				},
				Action: runEvaluation,
			},
			{
				Name:  "scorecard",
				Usage: "render a compact Markdown page from scored evaluation records",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "results",
						Value: "evaluations/latest",
						Usage: "directory containing scored evaluation YAML",
					},
					&cli.StringFlag{
						Name:  "seat",
						Value: "codex",
						Usage: "include records for this harness seat",
					},
					&cli.StringFlag{
						Name:  "out",
						Usage: "write the page to this path instead of standard output",
					},
					&cli.BoolFlag{
						Name:  "check",
						Usage: "fail when the output path does not match the generated page",
					},
				},
				Action: runScorecard,
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
					&cli.StringSliceFlag{
						Name:  "personality-library",
						Usage: "additional local personality-library root (repeatable)",
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
					&cli.StringSliceFlag{
						Name:  "personality-library",
						Usage: "additional local personality-library root (repeatable)",
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
					&cli.StringSliceFlag{
						Name:  "personality-library",
						Usage: "additional local personality-library root (repeatable)",
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

func personCatalogFlags(includeQuery bool) []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{Name: "person-source", Usage: "external person-package root (defaults to embedded person:kai)"},
		&cli.StringSliceFlag{Name: "personality-library", Usage: "additional local personality-library root (repeatable)"},
		&cli.BoolFlag{Name: "json", Usage: "emit agent-compose.catalog.v1 JSON"},
	}
	if includeQuery {
		flags = append(flags, &cli.StringFlag{Name: "query", Usage: "personality slug or declared cue"})
	}
	return flags
}

func runBundleExport(_ context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 1 {
		return fmt.Errorf("bundle export needs exactly one bundle directory")
	}
	return bundle.Export(cmd.Args().First(), cmd.String("out"))
}

func writeCatalog(value any, asJSON bool, text string) error {
	if !asJSON {
		_, err := fmt.Fprint(os.Stdout, text)
		return err
	}
	raw, err := json.MarshalIndent(map[string]any{"format": "agent-compose.catalog.v1", "items": value}, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, string(raw))
	return err
}

func runCatalogPersonalities(_ context.Context, cmd *cli.Command) error {
	p, _, err := loadSelectedPersonWithLibraries(cmd.String("person-source"), cmd.StringSlice("personality-library"))
	if err != nil {
		return err
	}
	var names []string
	if query := cmd.String("query"); query != "" {
		names, err = p.LookupCue(query)
		if err != nil {
			return err
		}
		if names == nil {
			names = []string{}
		}
	}
	entries, err := p.PersonalityCatalog(names)
	if err != nil {
		return err
	}
	var text strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(
			&text,
			"%s // %s // %s // aliases: %s // roles: %s\n",
			entry.Slug,
			entry.Skill,
			entry.Description,
			strings.Join(entry.Aliases, ", "),
			catalogAffinityRoles(entry.Affinities),
		)
	}
	return writeCatalog(entries, cmd.Bool("json"), text.String())
}

func runCatalogRoles(_ context.Context, cmd *cli.Command) error {
	p, _, err := loadSelectedPersonWithLibraries(cmd.String("person-source"), cmd.StringSlice("personality-library"))
	if err != nil {
		return err
	}
	entries, err := p.RoleCatalog()
	if err != nil {
		return err
	}
	var text strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(
			&text,
			"%s // %s // %s // meld: %s // color: %s\n",
			entry.Slug,
			entry.Skill,
			entry.Purpose,
			strings.Join(entry.Personalities, ", "),
			entry.FavoriteColor,
		)
	}
	return writeCatalog(entries, cmd.Bool("json"), text.String())
}

func runCatalogSeats(_ context.Context, cmd *cli.Command) error {
	p, _, err := loadSelectedPersonWithLibraries(cmd.String("person-source"), cmd.StringSlice("personality-library"))
	if err != nil {
		return err
	}
	seats, err := p.SeatCatalog(cmd.String("role"))
	if err != nil {
		return err
	}
	var text strings.Builder
	for _, entry := range seats {
		fmt.Fprintf(
			&text,
			"%s // %s // %s // %s // channel: %s // tier: %s\n",
			entry.Role,
			entry.Seat.Selector(),
			entry.Seat.Name,
			entry.Seat.Pronouns,
			entry.Seat.Channel,
			entry.Seat.Tier,
		)
	}
	return writeCatalog(seats, cmd.Bool("json"), text.String())
}

func catalogAffinityRoles(affinities []person.PersonalityMeld) string {
	roles := make([]string, 0, len(affinities))
	for _, affinity := range affinities {
		roles = append(roles, affinity.Role)
	}
	return strings.Join(roles, ", ")
}

func runCatalogExpressions(_ context.Context, cmd *cli.Command) error {
	expressions := person.ExpressionVocabulary()
	return writeCatalog(expressions, cmd.Bool("json"), strings.Join(expressions, "\n")+"\n")
}

func runEvaluation(_ context.Context, cmd *cli.Command) error {
	p, external, err := loadSelectedPersonWithLibraries(cmd.String("person-source"), cmd.StringSlice("personality-library"))
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

func writeScorecard(stdout io.Writer, raw []byte, output string, check bool) error {
	if output == "" {
		if check {
			return fmt.Errorf("scorecard --check requires --out")
		}
		_, err := stdout.Write(raw)
		return err
	}
	if check {
		current, err := os.ReadFile(output)
		if err != nil {
			return fmt.Errorf("read committed scorecard: %w", err)
		}
		if !bytes.Equal(current, raw) {
			return fmt.Errorf(
				"scorecard is stale: rerun ward exec evaluation-scorecard",
			)
		}
		return nil
	}
	if err := os.WriteFile(output, raw, 0o644); err != nil {
		return fmt.Errorf("write scorecard: %w", err)
	}
	return nil
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

func runScorecard(_ context.Context, cmd *cli.Command) error {
	raw, err := evaluation.MarkdownScorecard(
		cmd.String("results"),
		cmd.String("seat"),
	)
	if err != nil {
		return err
	}
	return writeScorecard(os.Stdout, raw, cmd.String("out"), cmd.Bool("check"))
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
	p, _, err := loadSelectedPersonWithLibraries(cmd.String("person-source"), cmd.StringSlice("personality-library"))
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
	hostPerson, err := loadHostPersonOptions(cascade.DefaultPaths())
	if err != nil {
		return err
	}
	result, err := compose.RunWithOptions(requestPath, outDir, hostPerson)
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

func runNativeLaunch(_ context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) < 2 {
		return fmt.Errorf("launch needs <role> <harness> [harness arguments...]")
	}
	role := strings.TrimSpace(args[0])
	harness := strings.TrimSpace(args[1])
	if !nativeHarness(harness) {
		return fmt.Errorf(
			"unsupported native harness %q: want claude, codex, goose, or opencode",
			harness,
		)
	}
	if os.Getenv(launch.EnvSentinel) != "" {
		return fmt.Errorf("native role launch cannot start inside another agent-compose launch")
	}
	paths := cascade.DefaultPaths()
	if code := converge.Run(
		paths,
		converge.Options{},
		os.Stdout,
		os.Stderr,
	); code != 0 {
		return cli.Exit("", code)
	}
	stateDir, err := home.Dir()
	if err != nil {
		return fmt.Errorf("resolve agent-compose state: %w", err)
	}
	cwd, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("resolve native launch directory: %w", err)
	}
	personSelection, err := loadHostPersonOptions(paths)
	if err != nil {
		return err
	}
	result, err := nativelaunch.Refresh(nativelaunch.Options{
		Role:            role,
		Harness:         harness,
		ModelClass:      os.Getenv(nativelaunch.EnvModelClass),
		CWD:             cwd,
		TargetDir:       cwd,
		ManifestPath:    filepath.Join(filepath.Dir(paths.Composed), "mount-eligibility.json"),
		OutDir:          filepath.Join(stateDir, "bundles"),
		PersonSelection: personSelection,
	})
	if err != nil {
		return err
	}
	interactive := nativeLaunchInteractive(os.Stdin, os.Stdout)
	summaryOpts := person.RoleTranscriptOptions{
		Color:     colorEnabled(),
		TrueColor: trueColorTerminal(),
	}
	state := "new"
	if result.BundleReused {
		state = "reused"
	}
	if interactive {
		printNativeLaunchStatus(os.Stderr, role, harness, result, state)
	}
	if err := printNativeLaunchSummary(
		os.Stdout,
		result.Composition,
		summaryOpts,
		interactive,
	); err != nil {
		return err
	}
	if err := acknowledgeNativeLaunch(
		os.Stdin,
		os.Stdout,
		interactive,
	); err != nil {
		return err
	}
	runtimeHome := strings.TrimSpace(os.Getenv(nativelaunch.EnvRuntimeHome))
	if err := clearNativeLaunchEnvironment(); err != nil {
		return err
	}
	if runtimeHome != "" {
		if err := activateNativeRuntimeHome(runtimeHome, harness); err != nil {
			return err
		}
	}
	if !interactive {
		printNativeLaunchStatus(os.Stderr, role, harness, result, state)
	}
	return execReal(nativeHarnessCommand(harness, args[2:]))
}

func nativeLaunchInteractive(input, output *os.File) bool {
	return term.IsTerminal(int(input.Fd())) && term.IsTerminal(int(output.Fd()))
}

func printNativeLaunchStatus(
	w io.Writer,
	role, harness string,
	result *nativelaunch.Result,
	state string,
) {
	fmt.Fprintf(
		w,
		"agent-compose: assigned %s to %s (%s %s bundle, %d files)\n",
		role,
		harness,
		result.ModelClass,
		state,
		result.Projected,
	)
}

func acknowledgeNativeLaunch(
	input io.Reader,
	output io.Writer,
	interactive bool,
) error {
	if !interactive {
		return nil
	}
	if _, err := fmt.Fprint(output, "Press Enter to continue"); err != nil {
		return fmt.Errorf("write native launch acknowledgement: %w", err)
	}
	var next [1]byte
	for {
		count, err := input.Read(next[:])
		if count > 0 && next[0] == '\n' {
			if _, writeErr := fmt.Fprintln(output); writeErr != nil {
				return fmt.Errorf("finish native launch acknowledgement: %w", writeErr)
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("read native launch acknowledgement: %w", err)
		}
	}
}

func nativeHarnessCommand(harness string, args []string) []string {
	command := append([]string{harness}, args...)
	if harness == "codex" && codexAcceptsInitialPrompt(args) {
		command = append(command, nativeCodexIntroductionPrompt)
	}
	return command
}

func codexAcceptsInitialPrompt(args []string) bool {
	valueOptions := map[string]bool{
		"-a": true, "--add-dir": true, "--ask-for-approval": true,
		"-c": true, "--cd": true, "--config": true,
		"-i": true, "--image": true,
		"-m": true, "--model": true,
		"-p": true, "--profile": true,
		"-s": true, "--sandbox": true,
		"--disable": true, "--enable": true,
		"--local-provider": true, "--remote": true, "--remote-auth-token-env": true,
	}
	booleanOptions := map[string]bool{
		"--dangerously-bypass-approvals-and-sandbox": true,
		"--dangerously-bypass-hook-trust":            true,
		"--no-alt-screen":                            true,
		"--oss":                                      true,
		"--search":                                   true,
		"--strict-config":                            true,
	}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--" || !strings.HasPrefix(arg, "-") {
			return false
		}
		name := arg
		if before, _, found := strings.Cut(arg, "="); found {
			name = before
		}
		if valueOptions[name] {
			if strings.Contains(arg, "=") {
				continue
			}
			index++
			if index >= len(args) {
				return false
			}
			continue
		}
		if !booleanOptions[name] {
			return false
		}
	}
	return true
}

func clearNativeLaunchEnvironment() error {
	for _, name := range []string{
		nativelaunch.EnvModelClass,
		nativelaunch.EnvRuntimeHome,
	} {
		if err := os.Unsetenv(name); err != nil {
			return fmt.Errorf("clear native launch environment %s: %w", name, err)
		}
	}
	return nil
}

func activateNativeRuntimeHome(home, harness string) error {
	absolute, err := filepath.Abs(home)
	if err != nil {
		return fmt.Errorf("resolve native runtime home: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return fmt.Errorf("inspect native runtime home %s: %w", absolute, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("native runtime home %s is not a directory", absolute)
	}
	codexHome := filepath.Join(absolute, ".codex")
	if harness == "codex" {
		if resolved, err := filepath.EvalSymlinks(codexHome); err == nil {
			codexHome = resolved
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("resolve native Codex home %s: %w", codexHome, err)
		}
	}
	environment := map[string]string{
		"HOME":            absolute,
		"USERPROFILE":     absolute,
		"CODEX_HOME":      codexHome,
		"XDG_CONFIG_HOME": filepath.Join(absolute, ".config"),
	}
	if harness == "claude" {
		environment["CLAUDE_CONFIG_DIR"] = filepath.Join(absolute, ".claude")
	}
	for name, value := range environment {
		if err := os.Setenv(name, value); err != nil {
			return fmt.Errorf("set native runtime environment %s: %w", name, err)
		}
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
	hostPerson, err := loadHostPersonOptions(cascade.DefaultPaths())
	if err != nil {
		return err
	}
	result, err := launch.Refresh(launch.Options{
		RequestPath:  requestPath,
		Layout:       layout,
		TargetDir:    cmd.String("target"),
		OutDir:       outDir,
		PersonPolicy: hostPerson.PersonPolicy,
		PersonSource: hostPerson.PersonSource,
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
	p, _, err := loadSelectedPersonWithLibraries(cmd.String("person-source"), cmd.StringSlice("personality-library"))
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
	p, _, err := loadSelectedPersonWithLibraries(cmd.String("person-source"), cmd.StringSlice("personality-library"))
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
	return loadSelectedPersonAt(source, cascade.DefaultPaths())
}

func loadSelectedPersonAt(source string, paths cascade.Paths) (*person.Person, bool, error) {
	return loadSelectedPersonWithLibrariesAt(source, nil, paths)
}

func loadSelectedPersonWithLibraries(source string, libraries []string) (*person.Person, bool, error) {
	return loadSelectedPersonWithLibrariesAt(source, libraries, cascade.DefaultPaths())
}

func loadSelectedPersonWithLibrariesAt(source string, libraries []string, paths cascade.Paths) (*person.Person, bool, error) {
	if source == "" {
		hostPerson, err := loadHostPersonOptions(paths)
		if err != nil {
			return nil, false, err
		}
		source = hostPerson.PersonSource
		if source == "" {
			p, err := person.Load()
			return p, false, err
		}
	}
	p, err := person.LoadDirectoryWithLibraries(source, libraries...)
	return p, true, err
}

func loadHostPersonOptions(paths cascade.Paths) (compose.Options, error) {
	if _, err := os.Stat(paths.Config); err != nil {
		if os.IsNotExist(err) {
			return compose.Options{}, nil
		}
		return compose.Options{}, fmt.Errorf("inspect host config %s: %w", paths.Config, err)
	}
	cfg, err := cascade.LoadConfig(paths.Config)
	if err != nil {
		return compose.Options{}, err
	}
	if cfg.PersonPolicy == "" {
		return compose.Options{}, nil
	}
	return compose.Options{
		PersonPolicy: cfg.PersonPolicy,
		PersonSource: cascade.ResolveConfiguredPath(
			cfg.PersonSource,
			paths.Config,
			paths.Home,
		),
		PersonalityLibraries: resolveConfiguredPaths(cfg.PersonalityLibraries, paths.Config, paths.Home),
	}, nil
}

func resolveConfiguredPaths(values []string, configPath, home string) []string {
	paths := make([]string, 0, len(values))
	for _, value := range values {
		paths = append(paths, cascade.ResolveConfiguredPath(value, configPath, home))
	}
	return paths
}

func runProject(_ context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 1 {
		return fmt.Errorf("project needs exactly one bundle directory")
	}
	bundleDir := cmd.Args().First()
	hostPerson, err := loadHostPersonOptions(cascade.DefaultPaths())
	if err != nil {
		return err
	}
	if err := validateProjectPersonPolicy(bundleDir, hostPerson); err != nil {
		return err
	}
	result, err := project.ProjectScoped(bundleDir, cmd.String("layout"), cmd.String("target"), cmd.String("scope"))
	if err != nil {
		return err
	}
	fmt.Printf("projected %d files into layout %s (%s scope) under %s\n",
		len(result.Files), result.Layout, cmd.String("scope"), cmd.String("target"))
	return nil
}

func validateProjectPersonPolicy(bundleDir string, opts compose.Options) error {
	if opts.PersonPolicy != personpolicy.ExternalOnly {
		return nil
	}
	verification, err := bundle.Verify(bundleDir)
	if err != nil {
		return err
	}
	external := false
	for _, identity := range verification.Identities {
		if identity.Source == "person:kai" {
			return fmt.Errorf("external-only person policy rejects bundle source %q", identity.Source)
		}
		if strings.HasPrefix(identity.Source, "person:") {
			external = true
		}
	}
	if external {
		return nil
	}
	return fmt.Errorf("external-only person policy requires an external person bundle")
}

// printSummary renders the complete selected-role identity transcript followed
// by bounded composition audit counts.
func printSummary(
	w io.Writer,
	r *compose.Result,
	opts person.RoleTranscriptOptions,
) error {
	return printNativeLaunchSummary(w, r, opts, false)
}

func printNativeLaunchSummary(
	w io.Writer,
	r *compose.Result,
	opts person.RoleTranscriptOptions,
	roleLast bool,
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
	var audit strings.Builder
	fmt.Fprintf(&audit, "sources: %s\n", strings.Join(r.Resolution.SourceIDs, ", "))
	fmt.Fprintf(&audit, "decisions: %d selected // %d excluded // %d shadowed // %d delivered\n",
		counts[resolver.OutcomeSelected], counts[resolver.OutcomeExcluded],
		counts[resolver.OutcomeShadowed], counts[resolver.OutcomeDelivered])
	fmt.Fprintf(&audit, "path: %s\n", r.Bundle.Dir)
	fmt.Fprintf(&audit, "trace: %s\n", filepath.Join(r.Bundle.Dir, "trace.json"))
	if roleLast {
		fmt.Fprint(w, audit.String())
		fmt.Fprintln(w)
		fmt.Fprint(w, metadata)
		return nil
	}
	fmt.Fprint(w, metadata)
	fmt.Fprintln(w)
	fmt.Fprint(w, audit.String())
	return nil
}
