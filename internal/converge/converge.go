// Package converge is bare `agent-compose compose`: refresh the roster
// artifact, then run the cascade - the whole host in one verb.
package converge

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/cascade"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/nativemcp"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/remoteskills"
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
	ttl, err := cascade.RemoteSkillTTL(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "agent-compose: %v\n", err)
		return 1
	}
	remote, err := remoteskills.Hydrate(
		context.Background(),
		activeRemoteSources(cfg.RemoteSkillSources, cfg.SkillLoadPoints),
		remoteskills.Options{
			StateDir: filepath.Dir(paths.Config),
			TTL:      ttl,
		},
	)
	if err != nil {
		fmt.Fprintf(stderr, "agent-compose: %v\n", err)
		return 1
	}
	catalogs := make([]skillmount.Catalog, 0, len(remote))
	states := map[remoteskills.State]int{}
	for _, result := range remote {
		catalogs = append(catalogs, skillmount.Catalog{
			Path:      result.Catalog.Path,
			Harnesses: result.Catalog.Harnesses,
		})
		states[result.State]++
		if result.Warning != "" {
			fmt.Fprintf(stderr, "agent-compose: warning: %s\n", result.Warning)
		}
		if opts.Verbose {
			label := result.Source.URL
			if result.Source.Ref != "" {
				label += "@" + result.Source.Ref
			}
			fmt.Fprintf(stdout, "remote  %s => %s [%s]\n",
				label, result.Catalog.Path, result.State)
		}
	}
	if len(remote) > 0 {
		fmt.Fprintf(stdout, "remote  sources=%d cached=%d hydrated=%d refreshed=%d fallback=%d\n",
			len(remote),
			states[remoteskills.StateCached],
			states[remoteskills.StateHydrated],
			states[remoteskills.StateRefreshed],
			states[remoteskills.StateFallback],
		)
	}
	manifestPath := filepath.Join(filepath.Dir(paths.Composed), "mount-eligibility.json")
	skills, err := skillmount.ApplyWithCatalogs(
		manifestPath,
		cfg.SkillLoadPoints,
		filepath.Dir(paths.Config),
		catalogs,
	)
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

func activeRemoteSources(
	sources []remoteskills.Source,
	loadPoints map[string]string,
) []remoteskills.Source {
	if len(loadPoints) == 0 {
		return nil
	}
	active := make([]remoteskills.Source, 0, len(sources))
	for _, source := range sources {
		if len(source.Harnesses) == 0 {
			active = append(active, source)
			continue
		}
		for _, harness := range source.Harnesses {
			if _, exists := loadPoints[strings.TrimSpace(harness)]; exists {
				active = append(active, source)
				break
			}
		}
	}
	return active
}
