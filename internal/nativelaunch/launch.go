// Package nativelaunch builds and projects one caller-assigned role bundle for
// a native harness session.
package nativelaunch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/project"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/skillmount"
)

const (
	// EnvModelClass carries the launch consumer's runtime model-class selection.
	EnvModelClass = "AGENT_COMPOSE_MODEL_CLASS"
	// EnvRuntimeHome selects a session-scoped home for the harness process.
	EnvRuntimeHome = "AGENT_COMPOSE_RUNTIME_HOME"
)

// Options supplies the host-owned selection and filesystem anchors for one
// native role launch.
type Options struct {
	Role            string
	Harness         string
	ModelClass      string
	CWD             string
	TargetDir       string
	ManifestPath    string
	OutDir          string
	PersonSelection compose.Options
}

// Result records the immutable bundle and projected load points selected for
// the native session.
type Result struct {
	Composition  *compose.Result
	BundleDir    string
	BundleReused bool
	Projected    int
	ModelClass   string
	Sources      []compose.RootSource
}

type repository struct {
	relative string
	provider skillmount.Provider
}

// Refresh resolves eligible providers, composes the complete role meld, and
// transactionally projects the result at the selected harness load points.
func Refresh(opts Options) (*Result, error) {
	if err := validateHarness(opts.Harness); err != nil {
		return nil, err
	}
	modelClass, err := normalizeModelClass(opts.ModelClass)
	if err != nil {
		return nil, err
	}
	manifest, err := skillmount.LoadEligibility(opts.ManifestPath)
	if err != nil {
		return nil, err
	}
	roots, missing, err := resolveRoots(manifest, opts.Harness, opts.Role, opts.CWD)
	if err != nil {
		return nil, err
	}
	request := &schema.Request{
		Role:       strings.TrimSpace(opts.Role),
		Delivery:   schema.DeliveryNativeSkills,
		ModelClass: modelClass,
	}
	composed, err := compose.RunRootsWithMissing(
		request,
		roots,
		missing,
		opts.OutDir,
		opts.PersonSelection,
	)
	if err != nil {
		return nil, fmt.Errorf("compose native role %q: %w", opts.Role, err)
	}
	target := opts.TargetDir
	if target == "" {
		target = opts.CWD
	}
	projected, err := project.Project(composed.Bundle.Dir, opts.Harness, target)
	if err != nil {
		return nil, fmt.Errorf(
			"project native role %q for %s into %s: %w",
			opts.Role,
			opts.Harness,
			target,
			err,
		)
	}
	return &Result{
		Composition:  composed,
		BundleDir:    composed.Bundle.Dir,
		BundleReused: composed.Bundle.Reused,
		Projected:    len(projected.Files),
		ModelClass:   modelClass,
		Sources:      roots,
	}, nil
}

func validateHarness(harness string) error {
	switch strings.TrimSpace(harness) {
	case "claude", "codex", "goose", "opencode":
		return nil
	default:
		return fmt.Errorf(
			"unsupported native harness %q: want claude, codex, goose, or opencode",
			harness,
		)
	}
}

func normalizeModelClass(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return schema.ModelClassFrontier, nil
	}
	switch value {
	case schema.ModelClassFrontier, schema.ModelClassLowContext:
		return value, nil
	default:
		return "", fmt.Errorf("unsupported native model class %q", value)
	}
}

func resolveRoots(
	manifest skillmount.Eligibility,
	harness string,
	role string,
	cwd string,
) ([]compose.RootSource, []schema.MissingSource, error) {
	providers := manifest.Providers(harness, strings.TrimSpace(role))
	if len(providers) == 0 {
		return nil, nil, fmt.Errorf("mount eligibility selects no repositories for %s", harness)
	}
	relative := make([]repository, 0, len(providers))
	for _, provider := range providers {
		rel, err := filepath.Rel(manifest.ProjectsRoot, provider.Path)
		if err != nil || rel == "." || rel == ".." ||
			strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, nil, fmt.Errorf(
				"eligible repository %s is outside projects_root %s",
				provider.Path,
				manifest.ProjectsRoot,
			)
		}
		relative = append(relative, repository{relative: rel, provider: provider})
	}
	projectsRoot := resolveProjectsRoot(cwd, manifest.ProjectsRoot, relative)
	roots := make([]compose.RootSource, 0, len(relative))
	var missing []schema.MissingSource
	hasRoleProvider := false
	sourceIDs := map[string]bool{}
	selectedProviderIDs := map[string]bool{}
	for _, repo := range relative {
		root := filepath.Join(projectsRoot, repo.relative)
		id := sourceID(repo.relative)
		selectedProviderIDs[id] = true
		if info, err := os.Stat(filepath.Join(root, ".agents", "skills")); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if repo.provider.Scope == "role" {
					reason := fmt.Sprintf(
						"optional role provider %s for role %q is unavailable",
						repo.relative,
						role,
					)
					if repo.provider.Required {
						return nil, nil, fmt.Errorf(
							"required role provider %s for role %q is unavailable beneath %s",
							repo.relative,
							role,
							projectsRoot,
						)
					}
					missing = append(missing, schema.MissingSource{
						ID: id, Reason: reason, ProviderScope: schema.ProviderScopeRole,
					})
				}
				continue
			}
			return nil, nil, fmt.Errorf("inspect eligible provider %s: %w", root, err)
		} else if !info.IsDir() {
			if repo.provider.Scope == "role" {
				reason := fmt.Sprintf(
					"optional role provider %s for role %q has no .agents/skills directory",
					repo.relative,
					role,
				)
				if repo.provider.Required {
					return nil, nil, fmt.Errorf(
						"required role provider %s for role %q has no .agents/skills directory",
						repo.relative,
						role,
					)
				}
				missing = append(missing, schema.MissingSource{
					ID: id, Reason: reason, ProviderScope: schema.ProviderScopeRole,
				})
			}
			continue
		}
		if info, err := os.Stat(filepath.Join(root, ".agents", "roles.kdl")); err == nil &&
			info.Mode().IsRegular() {
			hasRoleProvider = true
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, nil, fmt.Errorf("inspect role bindings in %s: %w", root, err)
		}
		if sourceIDs[id] {
			return nil, nil, fmt.Errorf(
				"eligible providers produce duplicate source id %q",
				id,
			)
		}
		sourceIDs[id] = true
		roots = append(roots, compose.RootSource{
			ID:     id,
			Root:   root,
			Reason: providerReason(repo.provider.Scope, harness, role),
			Scope:  repo.provider.Scope,
			Skills: append([]string(nil), repo.provider.Skills...),
		})
	}
	if len(roots) == 0 {
		return nil, nil, fmt.Errorf(
			"no eligible skill providers are available beneath %s",
			projectsRoot,
		)
	}
	if !hasRoleProvider {
		return nil, nil, fmt.Errorf(
			"native role launch needs an eligible provider with .agents/roles.kdl beneath %s",
			projectsRoot,
		)
	}
	missing = append(missing, excludedRoleProviders(
		manifest,
		strings.TrimSpace(role),
		projectsRoot,
		selectedProviderIDs,
	)...)
	return roots, missing, nil
}

func providerReason(scope, harness, role string) string {
	switch scope {
	case "harness":
		return fmt.Sprintf("provider selected for harness %q", harness)
	case "role":
		return fmt.Sprintf("role provider selected because role %q requests it", role)
	default:
		return "default provider selected for every assigned role"
	}
}

func excludedRoleProviders(
	manifest skillmount.Eligibility,
	selectedRole string,
	projectsRoot string,
	selectedIDs map[string]bool,
) []schema.MissingSource {
	rolesByPath := map[string][]string{}
	roles := make([]string, 0, len(manifest.RoleProviders))
	for role := range manifest.RoleProviders {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	for _, role := range roles {
		if role == selectedRole {
			continue
		}
		for _, provider := range manifest.RoleProviders[role] {
			rel, err := filepath.Rel(manifest.ProjectsRoot, provider.Path)
			if err != nil || rel == "." || rel == ".." ||
				strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				continue
			}
			id := sourceID(rel)
			if selectedIDs[id] {
				continue
			}
			rolesByPath[rel] = append(rolesByPath[rel], role)
		}
	}
	paths := make([]string, 0, len(rolesByPath))
	for path := range rolesByPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	missing := make([]schema.MissingSource, 0, len(paths))
	for _, rel := range paths {
		providerRoles := rolesByPath[rel]
		reason := fmt.Sprintf(
			"role provider is scoped to role(s) %s, not selected role %q",
			strings.Join(providerRoles, ", "),
			selectedRole,
		)
		entry := schema.MissingSource{
			ID: sourceID(rel), Reason: reason, ProviderScope: schema.ProviderScopeRole,
		}
		if source, err := schema.LoadSource(filepath.Join(projectsRoot, rel)); err == nil {
			seen := map[string]bool{}
			for _, skill := range source.Skills {
				if !seen[skill.ID] {
					entry.Skills = append(entry.Skills, skill.ID)
					seen[skill.ID] = true
				}
			}
			for _, roleSkills := range source.RoleSkills {
				for _, skill := range roleSkills {
					if !seen[skill.ID] {
						entry.Skills = append(entry.Skills, skill.ID)
						seen[skill.ID] = true
					}
				}
			}
			sort.Strings(entry.Skills)
		}
		missing = append(missing, entry)
	}
	return missing
}

func resolveProjectsRoot(cwd, configured string, repositories []repository) string {
	candidate, err := filepath.Abs(cwd)
	if err == nil {
		for {
			if providerCount(candidate, repositories) > 0 {
				return candidate
			}
			parent := filepath.Dir(candidate)
			if parent == candidate {
				break
			}
			candidate = parent
		}
	}
	return configured
}

func providerCount(root string, repositories []repository) int {
	count := 0
	for _, repo := range repositories {
		if info, err := os.Stat(filepath.Join(
			root,
			repo.relative,
			".agents",
			"skills",
		)); err == nil && info.IsDir() {
			count++
		}
	}
	return count
}

func sourceID(relative string) string {
	slash := filepath.ToSlash(relative)
	return strings.NewReplacer("/", "--", "\\", "--").Replace(slash)
}
