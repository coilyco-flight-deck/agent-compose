// Package compose orchestrates one composition end to end so the CLI stays a
// thin rendering layer over the same path the tests exercise.
package compose

import (
	"errors"
	"path/filepath"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/personpolicy"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

type Result struct {
	Resolution   *resolver.Resolution
	Bundle       *bundle.Result
	ExternalOnly bool
}

type Options struct {
	PersonPolicy string
	PersonSource string
}

type externalOnlyError struct {
	err error
}

func (e *externalOnlyError) Error() string {
	return e.err.Error()
}

func (e *externalOnlyError) Unwrap() error {
	return e.err
}

// IsExternalOnlyError reports whether composition failed after resolving an
// external-only policy. Launchers use it to prohibit unsafe fallback.
func IsExternalOnlyError(err error) bool {
	var target *externalOnlyError
	return errors.As(err, &target)
}

func Run(requestPath, outDir string) (*Result, error) {
	return RunWithOptions(requestPath, outDir, Options{})
}

// RunWithOptions applies a host person selection beneath any explicit
// request-local selection.
func RunWithOptions(requestPath, outDir string, opts Options) (*Result, error) {
	req, err := schema.ParseRequest(requestPath)
	if err != nil {
		return nil, err
	}
	hostExternalOnly := opts.PersonPolicy == personpolicy.ExternalOnly
	if err := personpolicy.Validate(opts.PersonPolicy, opts.PersonSource); err != nil {
		return nil, wrapPolicyError(err, hostExternalOnly)
	}
	externalOnly := hostExternalOnly ||
		req.PersonPolicy == personpolicy.ExternalOnly
	personSource := opts.PersonSource
	if req.PersonSource != "" {
		personSource = filepath.Join(filepath.Dir(requestPath), req.PersonSource)
	}
	if err := personpolicy.Validate(effectivePolicy(externalOnly), personSource); err != nil {
		return nil, wrapPolicyError(err, externalOnly)
	}

	p, err := person.Load()
	if personSource != "" {
		p, err = person.LoadDirectory(personSource)
	}
	if err != nil {
		return nil, wrapPolicyError(err, externalOnly)
	}
	sources, missing, err := schema.LoadSources(req, requestPath)
	if err != nil {
		return nil, wrapPolicyError(err, externalOnly)
	}
	res, err := resolver.Resolve(req, p, sources, missing)
	if err != nil {
		return nil, wrapPolicyError(err, externalOnly)
	}
	b, err := bundle.Materialize(res, outDir)
	if err != nil {
		return nil, wrapPolicyError(err, externalOnly)
	}
	return &Result{
		Resolution:   res,
		Bundle:       b,
		ExternalOnly: externalOnly,
	}, nil
}

func effectivePolicy(externalOnly bool) string {
	if externalOnly {
		return personpolicy.ExternalOnly
	}
	return ""
}

func wrapPolicyError(err error, externalOnly bool) error {
	if externalOnly {
		return &externalOnlyError{err: err}
	}
	return err
}
