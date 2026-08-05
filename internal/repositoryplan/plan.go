// Package repositoryplan owns the strict compiled repository-availability
// contract. Agent Compose is the only roles.kdl reader. Native launchers, AOS,
// Ward, and fleet convergence consume this machine-readable projection.
package repositoryplan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const Format = "agent-compose.repositories.v1"

type Selection struct {
	Identity   string   `json:"identity"`
	Path       string   `json:"path"`
	Source     string   `json:"source"`
	Scope      string   `json:"scope"`
	Reason     string   `json:"reason"`
	Required   bool     `json:"required,omitempty"`
	Skills     []string `json:"skills,omitempty"`
	Name       string   `json:"name,omitempty"`
	DeclaredBy string   `json:"declared_by,omitempty"`
}

type Plan struct {
	Format       string                 `json:"format"`
	ProjectsRoot string                 `json:"projects_root"`
	Roles        map[string][]Selection `json:"roles"`
	Residency    []Selection            `json:"residency"`
}

func Load(filename string) (Plan, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return Plan{}, fmt.Errorf("read repository plan %s: %w", filename, err)
	}
	var plan Plan
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return Plan{}, fmt.Errorf("parse repository plan %s: %w", filename, err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err != nil {
			return Plan{}, fmt.Errorf("parse repository plan %s: %w", filename, err)
		}
		return Plan{}, fmt.Errorf("parse repository plan %s: trailing JSON value", filename)
	}
	if err := plan.Validate(); err != nil {
		return Plan{}, fmt.Errorf("repository plan %s: %w", filename, err)
	}
	return plan, nil
}

func (p Plan) Validate() error {
	if p.Format != Format {
		return fmt.Errorf("format is %q, want %q", p.Format, Format)
	}
	root := strings.TrimSpace(p.ProjectsRoot)
	if root == "" || !filepath.IsAbs(root) {
		return fmt.Errorf("projects_root must be an absolute path")
	}
	for role, selections := range p.Roles {
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("roles contains an empty role")
		}
		if err := validateSelections(root, "role "+fmt.Sprintf("%q", role), selections); err != nil {
			return err
		}
	}
	return validateSelections(root, "residency", p.Residency)
}

func validateSelections(root, owner string, selections []Selection) error {
	prior := ""
	for index, selection := range selections {
		if !validIdentity(selection.Identity) {
			return fmt.Errorf("%s entry %d has invalid identity %q", owner, index, selection.Identity)
		}
		if selection.Identity <= prior {
			return fmt.Errorf("%s repository identities must be strictly sorted and deduplicated", owner)
		}
		prior = selection.Identity
		if strings.TrimSpace(selection.Source) == "" || strings.TrimSpace(selection.Scope) == "" || strings.TrimSpace(selection.Reason) == "" {
			return fmt.Errorf("%s repository %q needs source, scope, and reason provenance", owner, selection.Identity)
		}
		if !filepath.IsAbs(selection.Path) {
			return fmt.Errorf("%s repository %q path must be absolute", owner, selection.Identity)
		}
		rel, err := filepath.Rel(root, filepath.Clean(selection.Path))
		if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("%s repository %q path %s is outside projects_root %s", owner, selection.Identity, selection.Path, root)
		}
		if (selection.Name == "") != (selection.DeclaredBy == "") {
			return fmt.Errorf("%s repository %q needs both name and declared_by provider provenance", owner, selection.Identity)
		}
	}
	return nil
}

func validIdentity(value string) bool {
	if value == "" || strings.Contains(value, `\`) || path.IsAbs(value) || path.Clean(value) != value {
		return false
	}
	parts := strings.Split(value, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func (p Plan) ForRole(role string) ([]Selection, error) {
	selections, ok := p.Roles[strings.TrimSpace(role)]
	if !ok {
		roles := make([]string, 0, len(p.Roles))
		for candidate := range p.Roles {
			roles = append(roles, candidate)
		}
		sort.Strings(roles)
		return nil, fmt.Errorf("repository plan has no role %q, available roles: %s", role, strings.Join(roles, ", "))
	}
	return append([]Selection(nil), selections...), nil
}
