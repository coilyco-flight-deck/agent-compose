package person

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
)

const (
	SnapshotFormat        = "agent-compose.person-snapshot.v3"
	SnapshotSchemaVersion = 3
)

// Snapshot is the complete public person boundary emitted during convergence.
// Model selection, authority, and runtime routing stay outside this contract.
type Snapshot struct {
	Format           string                  `json:"format"`
	SchemaVersion    int                     `json:"schema_version"`
	Source           string                  `json:"source"`
	Person           string                  `json:"person"`
	RoleOrder        []string                `json:"role_order"`
	Roles            map[string]SnapshotRole `json:"roles"`
	Personalities    map[string]Personality  `json:"personalities"`
	Expressions      []string                `json:"expressions"`
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
		Expressions:      ExpressionVocabulary(),
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

// SnapshotV4 is an additive inspection projection. The v3 snapshot remains
// available for consumers that have not adopted aliases and affinities.
type SnapshotV4 struct {
	Format        string                         `json:"format"`
	SchemaVersion int                            `json:"schema_version"`
	Person        string                         `json:"person"`
	Roles         map[string]SnapshotRole        `json:"roles"`
	Personalities map[string]SnapshotPersonality `json:"personalities"`
	Expressions   []string                       `json:"expressions"`
}

type SnapshotPersonality struct {
	Personality
	SourceLibrary string            `json:"source_library"`
	Digest        string            `json:"digest"`
	Affinities    []PersonalityMeld `json:"affinities"`
}

type PersonalityMeld struct {
	Role          string   `json:"role"`
	Personalities []string `json:"personalities"`
}

// BuildSnapshotV4 derives aliases and role affinity from the effective
// profile graph. Legacy complete packages are represented as one local source.
func BuildSnapshotV4(p *Person) (*SnapshotV4, error) {
	v3, err := BuildSnapshot(p)
	if err != nil {
		return nil, err
	}
	personalities := make(map[string]SnapshotPersonality, len(p.Personalities))
	for _, name := range p.personalityOrder() {
		binding := p.Personalities[name]
		raw, err := json.Marshal(binding)
		if err != nil {
			return nil, fmt.Errorf("marshal personality %q for digest: %w", name, err)
		}
		digest := sha256.Sum256(raw)
		entry := SnapshotPersonality{
			Personality:   binding,
			SourceLibrary: "person:" + p.Name + ":local",
			Digest:        fmt.Sprintf("sha256:%x", digest),
		}
		for _, roleName := range p.RoleOrder {
			role := p.Roles[roleName]
			for _, member := range role.Personalities {
				if member == name {
					entry.Affinities = append(entry.Affinities, PersonalityMeld{
						Role:          roleName,
						Personalities: append([]string(nil), role.Personalities...),
					})
					break
				}
			}
		}
		personalities[name] = entry
	}
	return &SnapshotV4{
		Format:        "agent-compose.person-snapshot.v4",
		SchemaVersion: 4,
		Person:        p.Name,
		Roles:         v3.Roles,
		Personalities: personalities,
		Expressions:   v3.Expressions,
	}, nil
}

func MarshalSnapshotV4(p *Person) ([]byte, error) {
	snapshot, err := BuildSnapshotV4(p)
	if err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal person snapshot v4: %w", err)
	}
	return append(raw, '\n'), nil
}
