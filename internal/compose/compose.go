// Package compose orchestrates one composition end to end so the CLI stays a
// thin rendering layer over the same path the tests exercise.
package compose

import (
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

type Result struct {
	Resolution *resolver.Resolution
	Bundle     *bundle.Result
}

func Run(requestPath, outDir string) (*Result, error) {
	req, err := schema.ParseRequest(requestPath)
	if err != nil {
		return nil, err
	}
	p, err := person.Load()
	if err != nil {
		return nil, err
	}
	sources, missing, err := schema.LoadSources(req, requestPath)
	if err != nil {
		return nil, err
	}
	res, err := resolver.Resolve(req, p, sources, missing)
	if err != nil {
		return nil, err
	}
	b, err := bundle.Materialize(res, outDir)
	if err != nil {
		return nil, err
	}
	return &Result{Resolution: res, Bundle: b}, nil
}
