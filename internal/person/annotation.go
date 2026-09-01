package person

import (
	"fmt"
	"strings"
)

// SeatAnnotation renders `Angie [she] (Engineer)`, the identity string every
// terminal surface shows. See docs/overlay.md.
func SeatAnnotation(name, pronouns, roleDisplayName string) string {
	annotation := SeatLabel(name, pronouns)
	if annotation == "" {
		return ""
	}
	// Each part is optional, so a package without a display name still renders.
	if label := strings.TrimSpace(roleDisplayName); label != "" {
		annotation += " (" + label + ")"
	}
	return annotation
}

// SeatLabel renders `Angie [she]`, for surfaces that already print the role
// beside the seat and would otherwise state it twice.
func SeatLabel(name, pronouns string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if subject := SubjectPronoun(pronouns); subject != "" {
		return name + " [" + subject + "]"
	}
	return name
}

// WithShortID appends the session's dictatable short id: `Angie [she]` becomes
// `Angie [she] uz86`. Ephemeral surfaces only. See docs/identity.md.
func WithShortID(display, shortID string) string {
	display = strings.TrimSpace(display)
	shortID = strings.TrimSpace(shortID)
	if display == "" || shortID == "" {
		return display
	}
	return display + " " + shortID
}

// SubjectPronoun narrows an authored pronoun value to the subject form, since a
// package writes either the bare subject or the pair.
func SubjectPronoun(pronouns string) string {
	subject, _, _ := strings.Cut(strings.TrimSpace(pronouns), "/")
	return strings.TrimSpace(subject)
}

// IdentitySentence states the four parts a seat answers "who are you" with:
// preferred name, role, legal name on this seat, and creature. See #396.
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
