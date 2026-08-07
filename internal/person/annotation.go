package person

import "strings"

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

// SubjectPronoun narrows an authored pronoun value to the subject form, since a
// package writes either the bare subject or the pair.
func SubjectPronoun(pronouns string) string {
	subject, _, _ := strings.Cut(strings.TrimSpace(pronouns), "/")
	return strings.TrimSpace(subject)
}
