package main

import (
	"fmt"
	"io"
	"strings"
)

// startupClass is what a failed startup step does to a native launch: behavior
// warns, reach refuses. See docs/launch.md.
type startupClass string

const (
	startupWarn   startupClass = "warn"
	startupRefuse startupClass = "refuse"
)

type startupStep struct {
	Name  string
	Class startupClass
}

var (
	stepLaunchDepth     = startupStep{"launch-depth", startupRefuse}
	stepHostConverge    = startupStep{"host-converge", startupWarn}
	stepStateDirectory  = startupStep{"state-directory", startupRefuse}
	stepPerson          = startupStep{"person", startupWarn}
	stepOperatingBase   = startupStep{"operating-base", startupWarn}
	stepProjectionGuard = startupStep{"projection-guard", startupRefuse}
	stepRoleComposition = startupStep{"role-composition", startupWarn}
	stepRepoComposition = startupStep{"role-composition-repo-scope", startupRefuse}
	stepCard            = startupStep{"card", startupWarn}
	stepLaunchPause     = startupStep{"launch-pause", startupWarn}
	stepSelectorEnv     = startupStep{"selector-environment", startupWarn}
	stepRuntimeHome     = startupStep{"runtime-home", startupWarn}
	stepTelemetry       = startupStep{"telemetry", startupWarn}
	stepMCPScope        = startupStep{"mcp-scope", startupRefuse}
)

// startupSteps is every step in launch order, the list the doc must match.
var startupSteps = []startupStep{
	stepLaunchDepth,
	stepHostConverge,
	stepStateDirectory,
	stepPerson,
	stepOperatingBase,
	stepProjectionGuard,
	stepRoleComposition,
	stepRepoComposition,
	stepCard,
	stepLaunchPause,
	stepSelectorEnv,
	stepRuntimeHome,
	stepTelemetry,
	stepMCPScope,
}

type degradedStep struct {
	Step   string
	Reason string
}

// startupRun applies the policy to one launch and keeps what did not load.
type startupRun struct {
	w        io.Writer
	degraded []degradedStep
}

func (r *startupRun) do(step startupStep, fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}
	if step.Class == startupRefuse {
		return fmt.Errorf("%s refused the launch: %w", step.Name, err)
	}
	reason := strings.TrimSpace(strings.SplitN(err.Error(), "\n", 2)[0])
	r.degraded = append(r.degraded, degradedStep{Step: step.Name, Reason: reason})
	fmt.Fprintf(r.w, "agent-compose: warning: %s did not load, launching without it: %s\n", step.Name, reason)
	return nil
}

func (r *startupRun) names() []string {
	names := make([]string, 0, len(r.degraded))
	for _, d := range r.degraded {
		names = append(names, d.Step)
	}
	return names
}

// degradedOSC tells aterm which steps did not load, since the harness repaints
// over any warning. See docs/launch.md.
func degradedOSC(names []string) string {
	return "\x1b]7750;agent-compose;degraded=" + strings.Join(names, ",") + "\x07"
}

// degradedNote is the agent's own account of what it launched without.
func degradedNote(degraded []degradedStep) string {
	var b strings.Builder
	b.WriteString("agent-compose launched this session degraded. These startup steps did not load, ")
	b.WriteString("so the session runs without what they provide:\n")
	for _, d := range degraded {
		fmt.Fprintf(&b, "- %s: %s\n", d.Step, d.Reason)
	}
	b.WriteString("Do not assume role doctrine, identity, or skills that these steps would have supplied. ")
	b.WriteString("Say so when it bears on the task.")
	return b.String()
}

// degradedHarnessArgs hands the note to a harness that accepts extra system
// context as a launch argument. A caller's own flag wins, as elsewhere.
func degradedHarnessArgs(harness string, args []string, degraded []degradedStep) []string {
	if len(degraded) == 0 {
		return nil
	}
	note := degradedNote(degraded)
	switch harness {
	case "claude":
		if nativeArgsCarry(args, "--append-system-prompt") {
			return nil
		}
		return []string{"--append-system-prompt", note}
	case "codex":
		return []string{"-c", "developer_instructions=" + tomlString(note)}
	}
	return nil
}

// tomlString quotes a value for a codex -c override, which parses it as TOML.
func tomlString(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)
	return `"` + replacer.Replace(value) + `"`
}
