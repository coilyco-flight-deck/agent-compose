// Package launch is the refresh-then-exec path: compose, project, and hand
// off to the real harness with a recursion guard and a last-known-good net.
package launch

import (
	"fmt"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
)

// EnvSentinel marks a process launched by agent-compose. A nested launch
// sees it and skips refresh instead of recursing.
const EnvSentinel = "AGENT_COMPOSE_LAUNCH"

type Options struct {
	RequestPath string
	Layout      string
	TargetDir   string
	OutDir      string
}

type Result struct {
	BundleDir    string
	BundleReused bool
	Projected    int
	Fallback     bool
	Warning      string
}

// Refresh composes and projects, falling back to a validated last-known-good
// projection on failure with the cause surfaced as a warning.
func Refresh(opts Options) (*Result, error) {
	composed, err := compose.Run(opts.RequestPath, opts.OutDir)
	if err != nil {
		return fallbackOr(opts, fmt.Errorf("compose: %w", err))
	}
	projected, err := project.Project(composed.Bundle.Dir, opts.Layout, opts.TargetDir)
	if err != nil {
		return fallbackOr(opts, fmt.Errorf("project: %w", err))
	}
	return &Result{
		BundleDir:    composed.Bundle.Dir,
		BundleReused: composed.Bundle.Reused,
		Projected:    len(projected.Files),
	}, nil
}

func fallbackOr(opts Options, cause error) (*Result, error) {
	if err := project.Validate(opts.TargetDir); err != nil {
		return nil, fmt.Errorf("refresh failed and no last-known-good projection is usable (%v): %w", err, cause)
	}
	return &Result{Fallback: true, Warning: cause.Error()}, nil
}
