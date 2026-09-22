package person

import (
	"fmt"
	"strings"
)

// SeatAnnotation renders `Moss-Toad (Engineer)`, the identity string every
// terminal surface shows. See docs/overlay.md.
func SeatAnnotation(name, roleDisplayName string) string {
	annotation := SeatLabel(name)
	if annotation == "" {
		return ""
	}
	// Each part is optional, so a package without a display name still renders.
	if label := strings.TrimSpace(roleDisplayName); label != "" {
		annotation += " (" + label + ")"
	}
	return annotation
}

// SeatLabel renders `Moss-Toad`, for surfaces that already print the role
// beside the seat and would otherwise state it twice.
func SeatLabel(name string) string {
	return strings.TrimSpace(name)
}

// WithShortID appends the session's dictatable short id: `Moss-Toad` becomes
// `Moss-Toad uz86`. Ephemeral surfaces only. See docs/identity.md.
func WithShortID(display, shortID string) string {
	display = strings.TrimSpace(display)
	shortID = strings.TrimSpace(shortID)
	if display == "" || shortID == "" {
		return display
	}
	return display + " " + shortID
}

// IdentitySentence states the parts a seat answers "who are you" with. See
// #396 and docs/identity.md.
func IdentitySentence(name, roleDisplayName, legalName, creature string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "I am %s, an agent-compose persona.", name)
	if role := strings.TrimSpace(roleDisplayName); role != "" {
		fmt.Fprintf(&b, " My role is %s.", role)
	}
	// An unauthored legal name stays absent rather than guessed, so a seat that
	// is not a model simply does not answer that part.
	legal, beast := strings.TrimSpace(legalName), strings.TrimSpace(creature)
	if beast == name {
		beast = ""
	}
	switch {
	case legal != "" && beast != "":
		fmt.Fprintf(&b, " On this seat my legal name is %s, and my creature is the %s.", legal, beast)
	case legal != "":
		fmt.Fprintf(&b, " On this seat my legal name is %s.", legal)
	case beast != "":
		fmt.Fprintf(&b, " My creature is the %s.", beast)
	}
	return b.String()
}
