package statusline

import (
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/agentid"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
)

// Whoami returns what this session calls itself: `Angie [she] uz86`, or "" with
// no composition rather than inventing one. See docs/whoami.md.
func Whoami(opts Options) (string, error) {
	projection := resolveProjection(opts.Target)
	if projection.Bundle == "" {
		return "", nil
	}
	manifest, err := bundle.ReadManifest(projection.Bundle)
	if err != nil {
		return "", err
	}
	return person.WithShortID(
		seatDisplayName(manifest, projection.Layout),
		agentid.FromEnv(),
	), nil
}
