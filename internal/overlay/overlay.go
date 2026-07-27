// Package overlay projects the canonical person model into a compact,
// caller-stateful identity surface.
package overlay

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
)

const (
	Format        = "agent-compose.overlay.v1"
	SchemaVersion = 1
)

type Personality struct {
	Name      string           `json:"name"`
	Color     string           `json:"color"`
	Motif     string           `json:"motif"`
	Emblem    person.Emblem    `json:"emblem"`
	Form      person.Form      `json:"form"`
	SoundMark person.SoundMark `json:"sound_mark"`
}

type Document struct {
	Format        string        `json:"format"`
	SchemaVersion int           `json:"schema_version"`
	Person        string        `json:"person"`
	Role          string        `json:"role"`
	Purpose       string        `json:"purpose"`
	Seat          person.Seat   `json:"seat"`
	Expression    string        `json:"expression"`
	FavoriteColor string        `json:"favorite_color"`
	Personalities []Personality `json:"personalities"`
}

func Build(p *person.Person, roleName, harness, expression string) (*Document, error) {
	if p == nil {
		return nil, fmt.Errorf("build overlay: person is nil")
	}
	role, ok := p.Roles[roleName]
	if !ok {
		return nil, fmt.Errorf("build overlay: role %q is not defined", roleName)
	}
	var seat person.Seat
	for _, candidate := range role.Seats {
		if candidate.Selector() == harness {
			seat = candidate
			break
		}
	}
	if seat.Selector() == "" {
		return nil, fmt.Errorf("build overlay: role %q has no %q seat", roleName, harness)
	}
	if !validExpression(expression) {
		return nil, fmt.Errorf("build overlay: expression %q is not defined", expression)
	}

	doc := &Document{
		Format:        Format,
		SchemaVersion: SchemaVersion,
		Person:        p.Name,
		Role:          roleName,
		Purpose:       role.Purpose,
		Seat:          seat,
		Expression:    expression,
	}
	colors := make([]string, 0, len(role.Personalities))
	for _, name := range role.Personalities {
		binding, exists := p.Personalities[name]
		if !exists {
			return nil, fmt.Errorf("build overlay: role %q names missing personality %q", roleName, name)
		}
		colors = append(colors, binding.Color)
		doc.Personalities = append(doc.Personalities, Personality{
			Name:      name,
			Color:     binding.Color,
			Motif:     binding.Motif,
			Emblem:    binding.Emblem,
			Form:      binding.Form,
			SoundMark: binding.SoundMark,
		})
	}
	favorite, err := color.Favorite(colors)
	if err != nil {
		return nil, fmt.Errorf("build overlay: role %q favorite color: %w", roleName, err)
	}
	doc.FavoriteColor = favorite
	return doc, nil
}

func Marshal(doc *Document) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("marshal overlay: document is nil")
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal overlay: %w", err)
	}
	return append(raw, '\n'), nil
}

func RenderText(doc *Document, width int) (string, error) {
	if doc == nil {
		return "", fmt.Errorf("render overlay: document is nil")
	}
	if width < 20 {
		return "", fmt.Errorf("render overlay: width must be at least 20")
	}
	marks := make([]string, 0, len(doc.Personalities))
	names := make([]string, 0, len(doc.Personalities))
	for _, personality := range doc.Personalities {
		marks = append(marks, personality.Emblem.Glyph)
		names = append(names, personality.Name)
	}
	role := strings.ReplaceAll(doc.Role, "-", " ")
	wide := fmt.Sprintf(
		"%s  %s  ·  %s / %s  ·  %s  ·  %s",
		strings.Join(marks, " "),
		doc.Seat.Name,
		role,
		doc.Expression,
		strings.Join(names, " + "),
		doc.FavoriteColor,
	)
	if utf8.RuneCountInString(wide) <= width {
		return wide + "\n", nil
	}
	lines := []string{
		strings.Join(marks, " ") + "  " + doc.Seat.Name,
		role + " / " + doc.Expression,
		strings.Join(names, " + "),
		doc.FavoriteColor,
	}
	for _, line := range lines {
		if utf8.RuneCountInString(line) > width {
			return "", fmt.Errorf("render overlay: width %d cannot fit %q", width, line)
		}
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func validExpression(value string) bool {
	for _, expression := range person.ExpressionVocabulary() {
		if value == expression {
			return true
		}
	}
	return false
}
