// Package describe renders stored decision evidence for humans. It only
// reads trace.json and manifest.json; reasons are never reconstructed.
package describe

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/color"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/person"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
)

type Options struct {
	All       bool
	Color     bool
	TrueColor bool
}

// collapseAt is the excluded-skill count where describe folds repeats into
// one line; --all expands them.
const collapseAt = 4

func Bundle(dir string, opts Options) (string, error) {
	manifest, trace, err := load(dir)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "bundle %s · %s/%s · %s%s\n",
		filepath.Base(dir), manifest.Role, strings.Join(manifest.Personalities, "+"),
		manifest.Delivery.Mode, favoriteSuffix(manifest, opts))

	sections := []struct {
		title string
		kinds []string
	}{
		{"profile", []string{"profile"}},
		{"sources", []string{"source"}},
		{"selection", []string{"instruction", "skill"}},
		{"delivery", []string{"delivery"}},
	}
	for _, section := range sections {
		var lines []string
		var excludedSkills []resolver.Decision
		if section.title == "sources" {
			for _, id := range manifest.Sources {
				lines = append(lines, fmt.Sprintf("  %s %-36s %-14s %s",
					symbol(resolver.OutcomeSelected, opts), id, "", "composed"))
			}
		}
		for _, d := range trace.Decisions {
			if !contains(section.kinds, d.Kind) {
				continue
			}
			if section.title == "selection" && d.Kind == "skill" && d.Outcome == resolver.OutcomeExcluded && !opts.All {
				excludedSkills = append(excludedSkills, d)
				continue
			}
			lines = append(lines, renderLine(d, opts))
		}
		lines = append(lines, collapse(excludedSkills, opts)...)
		if len(lines) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n%s\n", section.title)
		for _, line := range lines {
			fmt.Fprintln(&b, line)
		}
	}

	fmt.Fprintf(&b, "\nmachine-readable trace: trace.json\n")
	return b.String(), nil
}

// Why follows one item through consideration to its outcome, adding what
// would have selected an excluded personality skill.
func Why(dir, subject string, opts Options) (string, error) {
	manifest, trace, err := load(dir)
	if err != nil {
		return "", err
	}
	var matched []resolver.Decision
	for _, d := range trace.Decisions {
		if d.Subject == subject || strings.TrimPrefix(d.Subject, d.Kind+":") == subject {
			matched = append(matched, d)
		}
	}
	if len(matched) == 0 {
		return "", fmt.Errorf("no decision mentions %q; try `agent-compose describe %s` for the full list", subject, dir)
	}
	var b strings.Builder
	for _, d := range matched {
		fmt.Fprintf(&b, "%s", subjectLabel(d))
		if d.Source != "" {
			fmt.Fprintf(&b, " (%s)", d.Source)
		}
		fmt.Fprintln(&b)
		if d.Source != "" {
			fmt.Fprintf(&b, "  considered: declared by %s\n", d.Source)
		}
		fmt.Fprintf(&b, "  outcome: %s\n", d.Outcome)
		fmt.Fprintf(&b, "  reason: %s\n", d.Reason)
		if hint := selectionHint(d, manifest); hint != "" {
			fmt.Fprintf(&b, "  %s\n", hint)
		}
	}
	return b.String(), nil
}

// Diff reports semantic decision changes between two bundles, keyed by
// subject rather than by file bytes.
func Diff(leftDir, rightDir string) (string, error) {
	leftManifest, leftTrace, err := load(leftDir)
	if err != nil {
		return "", err
	}
	rightManifest, rightTrace, err := load(rightDir)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s/%s (%s) → %s/%s (%s)\n",
		leftManifest.Role, strings.Join(leftManifest.Personalities, "+"), leftManifest.Delivery.Mode,
		rightManifest.Role, strings.Join(rightManifest.Personalities, "+"), rightManifest.Delivery.Mode)

	left, right := bySubject(leftTrace), bySubject(rightTrace)
	var subjects []string
	seen := map[string]bool{}
	for _, key := range append(keys(left), keys(right)...) {
		if !seen[key] {
			seen[key] = true
			subjects = append(subjects, key)
		}
	}
	sort.Strings(subjects)

	unchanged, changed := 0, 0
	for _, subject := range subjects {
		l, inLeft := left[subject]
		r, inRight := right[subject]
		switch {
		case inLeft && inRight && l.Outcome == r.Outcome && l.Reason == r.Reason:
			unchanged++
		case inLeft && inRight && l.Outcome == r.Outcome:
			changed++
			fmt.Fprintf(&b, "  ~ %-36s still %s, reason changed\n", subjectLabel(l), l.Outcome)
		case inLeft && inRight:
			changed++
			fmt.Fprintf(&b, "  ~ %-36s %s → %s\n", subjectLabel(l), l.Outcome, r.Outcome)
		case inLeft:
			changed++
			fmt.Fprintf(&b, "  - %-36s (%s)\n", subjectLabel(l), l.Outcome)
		default:
			changed++
			fmt.Fprintf(&b, "  + %-36s (%s)\n", subjectLabel(r), r.Outcome)
		}
	}
	if changed == 0 {
		fmt.Fprintln(&b, "  no semantic differences")
	}
	fmt.Fprintf(&b, "  %d decisions unchanged\n", unchanged)
	return b.String(), nil
}

func load(dir string) (*bundle.Manifest, *bundle.Trace, error) {
	manifest, err := bundle.ReadManifest(dir)
	if err != nil {
		return nil, nil, err
	}
	trace, err := bundle.ReadTrace(dir)
	if err != nil {
		return nil, nil, err
	}
	return manifest, trace, nil
}

func renderLine(d resolver.Decision, opts Options) string {
	label := subjectLabel(d)
	source := ""
	if d.Source != "" {
		source = "(" + d.Source + ")"
	}
	return fmt.Sprintf("  %s %-36s %-14s %s", symbol(d.Outcome, opts), label, source, d.Reason)
}

func collapse(excluded []resolver.Decision, opts Options) []string {
	if len(excluded) == 0 {
		return nil
	}
	if len(excluded) < collapseAt {
		lines := make([]string, 0, len(excluded))
		for _, d := range excluded {
			lines = append(lines, renderLine(d, opts))
		}
		return lines
	}
	return []string{fmt.Sprintf("  %s %d skills excluded — %s (--all to expand)",
		symbol(resolver.OutcomeExcluded, opts), len(excluded), excluded[0].Reason)}
}

func selectionHint(d resolver.Decision, manifest *bundle.Manifest) string {
	if d.Kind != "skill" || d.Outcome != resolver.OutcomeExcluded {
		return ""
	}
	p, err := person.Load()
	if err != nil {
		return ""
	}
	skillID := strings.TrimPrefix(d.Subject, "skill:")
	for personality, bound := range p.Personalities {
		if bound.Skill != skillID {
			continue
		}
		return fmt.Sprintf("personality %q binds it, but role %q does not activate that personality", personality, manifest.Role)
	}
	return "no personality binds this skill"
}

// favoriteSuffix appends the role's melded favorite color: hex always, a tinted
// swatch only where a terminal will render it.
func favoriteSuffix(manifest *bundle.Manifest, opts Options) string {
	if manifest.Color == "" {
		return ""
	}
	if !opts.Color {
		return " · melded " + manifest.Color
	}
	return " · melded " + manifest.Color + " " + color.ANSI(manifest.Color, "■", opts.TrueColor)
}

func symbol(outcome string, opts Options) string {
	symbols := map[string][2]string{
		resolver.OutcomeSelected:  {"✓", "\x1b[32m"},
		resolver.OutcomeExcluded:  {"✗", "\x1b[31m"},
		resolver.OutcomeShadowed:  {"~", "\x1b[33m"},
		resolver.OutcomeDelivered: {"→", "\x1b[36m"},
	}
	entry, ok := symbols[outcome]
	if !ok {
		entry = [2]string{"?", ""}
	}
	if !opts.Color {
		return entry[0]
	}
	return entry[1] + entry[0] + "\x1b[0m"
}

func subjectLabel(d resolver.Decision) string {
	return strings.Replace(d.Subject, ":", " ", 1)
}

func bySubject(tr *bundle.Trace) map[string]resolver.Decision {
	out := map[string]resolver.Decision{}
	for _, d := range tr.Decisions {
		out[d.Subject] = d
	}
	return out
}

func keys(m map[string]resolver.Decision) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func contains(list []string, item string) bool {
	for _, entry := range list {
		if entry == item {
			return true
		}
	}
	return false
}
