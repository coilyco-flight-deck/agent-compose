// Package launch is the refresh-then-exec path: compose, project, and hand
// off to the real harness with a recursion guard and a last-known-good net.
package launch

import (
	"fmt"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/personpolicy"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
)

// EnvSentinel marks a process launched by agent-compose. A nested launch
// sees it and skips refresh instead of recursing.
const EnvSentinel = "AGENT_COMPOSE_LAUNCH"

// AttributionRoleEnv carries the composed role to the git attribution shim.
// Per-session: the host projection is global. See coilysiren/inbox#362.
const AttributionRoleEnv = "AGENT_GIT_ATTRIBUTION_ROLE"

type Options struct {
	RequestPath  string
	Layout       string
	TargetDir    string
	OutDir       string
	PersonPolicy string
	PersonSource string
}

type Result struct {
	BundleDir    string
	BundleReused bool
	Projected    int
	Fallback     bool
	Warning      string
	Warnings     []string
}

// Refresh composes and projects, falling back to a validated last-known-good
// projection on failure with the cause surfaced as a warning.
func Refresh(opts Options) (*Result, error) {
	composed, err := compose.RunWithOptions(opts.RequestPath, opts.OutDir, compose.Options{
		PersonPolicy: opts.PersonPolicy,
		PersonSource: opts.PersonSource,
	})
	if err != nil {
		return fallbackOr(opts, compose.IsExternalOnlyError(err), fmt.Errorf("compose: %w", err))
	}
	projected, err := project.Project(composed.Bundle.Dir, opts.Layout, opts.TargetDir)
	if err != nil {
		return fallbackOr(opts, composed.ExternalOnly, fmt.Errorf("project: %w", err))
	}
	return &Result{
		BundleDir:    composed.Bundle.Dir,
		BundleReused: composed.Bundle.Reused,
		Projected:    len(projected.Files),
		Warnings:     append([]string(nil), composed.Resolution.Warnings...),
	}, nil
}

func fallbackOr(opts Options, externalOnly bool, cause error) (*Result, error) {
	if externalOnly || opts.PersonPolicy == personpolicy.ExternalOnly {
		return nil, fmt.Errorf("external-only person policy prohibits last-known-good fallback: %w", cause)
	}
	if err := project.Validate(opts.TargetDir); err != nil {
		return nil, fmt.Errorf("refresh failed and no last-known-good projection is usable (%v): %w", err, cause)
	}
	return &Result{Fallback: true, Warning: cause.Error()}, nil
}
