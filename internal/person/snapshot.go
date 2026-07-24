package person

import (
	"encoding/json"
	"fmt"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
)

const (
	SnapshotFormat        = "agent-compose.person-snapshot.v1"
	SnapshotSchemaVersion = 1
)

// Snapshot is the complete public, machine-readable person boundary emitted
// during host convergence. Authority and model-routing data deliberately stay
// outside this contract.
type Snapshot struct {
	Format           string                  `json:"format"`
	SchemaVersion    int                     `json:"schema_version"`
	Source           string                  `json:"source"`
	Person           string                  `json:"person"`
	RoleOrder        []string                `json:"role_order"`
	Roles            map[string]SnapshotRole `json:"roles"`
	Personalities    map[string]Personality  `json:"personalities"`
	InspirationOrder []string                `json:"inspiration_order"`
	Inspirations     map[string]Inspiration  `json:"inspirations"`
}

// SnapshotRole embeds the canonical role so future role fields enter the
// export automatically, then adds its deterministic derived favorite color.
type SnapshotRole struct {
	Role
	FavoriteColor string `json:"favorite_color"`
}

// BuildSnapshot converts the loaded person model without maintaining a second
// policy copy.
func BuildSnapshot(p *Person) (*Snapshot, error) {
	if p == nil {
		return nil, fmt.Errorf("build person snapshot: person is nil")
	}
	roles := make(map[string]SnapshotRole, len(p.Roles))
	for _, name := range p.RoleOrder {
		role, ok := p.Roles[name]
		if !ok {
			return nil, fmt.Errorf("build person snapshot: role order names missing role %q", name)
		}
		colors := make([]string, 0, len(role.Personalities))
		for _, personalityName := range role.Personalities {
			binding, ok := p.Personalities[personalityName]
			if !ok {
				return nil, fmt.Errorf(
					"build person snapshot: role %q names missing personality %q",
					name, personalityName,
				)
			}
			colors = append(colors, binding.Color)
		}
		favorite, err := color.Favorite(colors)
		if err != nil {
			return nil, fmt.Errorf("build person snapshot: role %q favorite color: %w", name, err)
		}
		roles[name] = SnapshotRole{Role: role, FavoriteColor: favorite}
	}
	if len(roles) != len(p.Roles) {
		return nil, fmt.Errorf(
			"build person snapshot: role order covers %d of %d roles",
			len(roles), len(p.Roles),
		)
	}
	return &Snapshot{
		Format:           SnapshotFormat,
		SchemaVersion:    SnapshotSchemaVersion,
		Source:           "person:" + p.Name,
		Person:           p.Name,
		RoleOrder:        append([]string(nil), p.RoleOrder...),
		Roles:            roles,
		Personalities:    p.Personalities,
		InspirationOrder: append([]string(nil), p.InspirationOrder...),
		Inspirations:     p.Inspirations,
	}, nil
}

func MarshalSnapshot(p *Person) ([]byte, error) {
	snapshot, err := BuildSnapshot(p)
	if err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal person snapshot: %w", err)
	}
	return append(raw, '\n'), nil
}
