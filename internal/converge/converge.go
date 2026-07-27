// Package converge is bare `agent-compose compose`: refresh the roster
// artifact, then run the cascade - the whole host in one verb.
package converge

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/nativemcp"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/roster"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/skillmount"
)

// Options controls host compose-layout reporting and forced application.
type Options struct {
	Reapply bool
	Verbose bool
}

// Run refreshes the roster from the selected person plus configured overlays,
// then cascades. Absent config stays the documented no-op.
func Run(paths cascade.Paths, opts Options, stdout, stderr io.Writer) int {
	cascadeOpts := cascade.RunOptions{
		Reapply: opts.Reapply,
		Verbose: opts.Verbose,
	}
	if _, err := os.Stat(paths.Config); err != nil {
		return cascade.Run(paths, cascadeOpts, stdout, stderr)
	}
	cfg, err := cascade.LoadConfig(paths.Config)
	if err != nil {
		fmt.Fprintf(stderr, "agent-compose: %v\n", err)
		return 1
	}

	p, err := person.Load()
	if err != nil {
		fmt.Fprintf(stderr, "agent-compose: %v\n", err)
		return 1
	}
	if cfg.PersonSource != "" {
		sourcePath := cascade.ResolveConfiguredPath(cfg.PersonSource, paths.Config, paths.Home)
		p, err = person.LoadDirectory(sourcePath)
		if err != nil {
			fmt.Fprintf(stderr, "agent-compose: %v\n", err)
			return 1
		}
	}
	var sources []*schema.Source
	for _, sourcePath := range cfg.RosterSources {
		src, err := schema.LoadSource(sourcePath)
		if err != nil {
			fmt.Fprintf(stderr, "agent-compose: warning: roster source %s: %v (skipped)\n", sourcePath, err)
			continue
		}
		sources = append(sources, src)
	}
	outDir := filepath.Join(filepath.Dir(paths.Config), "sources", "personality")
	files, err := roster.Render(p, sources, outDir)
	if err != nil {
		fmt.Fprintf(stderr, "agent-compose: %v\n", err)
		return 1
	}
	result, err := project.ApplyOwned(outDir, files, "roster", "person:"+p.Name)
	if err != nil {
		fmt.Fprintf(stderr, "agent-compose: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "roster  %s (%d files)\n", outDir, len(result.Files))

	if code := cascade.Run(paths, cascadeOpts, stdout, stderr); code != 0 {
		return code
	}
	manifestPath := filepath.Join(filepath.Dir(paths.Composed), "mount-eligibility.json")
	skills, err := skillmount.Apply(manifestPath, cfg.SkillLoadPoints, filepath.Dir(paths.Config))
	if err != nil {
		fmt.Fprintf(stderr, "agent-compose: %v\n", err)
		return 1
	}
	if skills.LoadPoints > 0 {
		fmt.Fprintf(stdout, "skills  managed=%d load-points=%d verified=%d linked=%d removed=%d preserved=%d\n",
			skills.Managed, skills.LoadPoints, skills.Verified, skills.Linked, skills.Removed, skills.Skipped)
	}
	if cfg.MCPInventory != "" {
		inventory := cascade.ResolveConfiguredPath(cfg.MCPInventory, paths.Config, paths.Home)
		result, err := nativemcp.Project(nativemcp.Options{
			Inventory: inventory,
			Home:      paths.Home,
		})
		if err != nil {
			fmt.Fprintf(stderr, "agent-compose: %v\n", err)
			return 1
		}
		state := "unchanged"
		if result.Changed {
			state = "changed"
		}
		fmt.Fprintf(stdout, "mcp     servers=%d state=%s\n", result.Servers, state)
	}

	return 0
}
