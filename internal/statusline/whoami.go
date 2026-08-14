package statusline

import (
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/agentid"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
)

// Whoami returns what this session calls itself: `Angie [she] uz86`, or "" with
// no projection rather than inventing one. See docs/whoami.md.
func Whoami(opts Options) (string, error) {
	target := strings.TrimSpace(opts.Target)
	if target == "" {
		target = "."
	}
	projection, _ := project.FindProjection(target)
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
