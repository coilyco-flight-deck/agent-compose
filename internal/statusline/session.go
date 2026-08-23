package statusline

import (
	"os"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/launch"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
)

// resolveProjection answers which composition this call describes. The launch
// binding wins, an explicit target keeps the walk. See docs/whoami.md.
func resolveProjection(target string) project.Projection {
	target = strings.TrimSpace(target)
	if target == "" {
		if bound, ok := sessionProjection(); ok {
			return bound
		}
		target = "."
	}
	projection, _ := project.FindProjection(target)
	return projection
}

func sessionProjection() (project.Projection, bool) {
	dir := strings.TrimSpace(os.Getenv(launch.SessionBundleEnv))
	layout := strings.TrimSpace(os.Getenv(launch.SessionLayoutEnv))
	if dir == "" || layout == "" {
		return project.Projection{}, false
	}
	return project.Projection{Bundle: dir, Layout: layout}, true
}
