package person

import "testing"

func TestSeatAnnotation(t *testing.T) {
	cases := map[string]struct {
		name        string
		pronouns    string
		displayName string
		want        string
	}{
		"complete":            {"Angie", "she", "Engineer", "Angie [she] (Engineer)"},
		"authored label wins": {"Quail", "they", "QA", "Quail [they] (QA)"},
		"pronoun pair narrows": {
			"Darren", "he/him", "Director", "Darren [he] (Director)",
		},
		"missing pronouns":     {"Angie", "", "Engineer", "Angie (Engineer)"},
		"missing display name": {"Angie", "she", "", "Angie [she]"},
		"missing name":         {"", "she", "Engineer", ""},
		"padded input":         {" Angie ", " she ", " Engineer ", "Angie [she] (Engineer)"},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			got := SeatAnnotation(test.name, test.pronouns, test.displayName)
			if got != test.want {
				t.Errorf("SeatAnnotation = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSeatLabelStopsShortOfTheRole(t *testing.T) {
	if got := SeatLabel("Angie", "she"); got != "Angie [she]" {
		t.Errorf("SeatLabel = %q, want Angie [she]", got)
	}
	if got := SeatLabel("Angie", ""); got != "Angie" {
		t.Errorf("SeatLabel without pronouns = %q, want Angie", got)
	}
	if got := SeatLabel("", "she"); got != "" {
		t.Errorf("SeatLabel without a name = %q, want empty", got)
	}
}

func TestSubjectPronoun(t *testing.T) {
	for input, want := range map[string]string{
		"she":       "she",
		"they/them": "they",
		"he / him":  "he",
		"":          "",
	} {
		if got := SubjectPronoun(input); got != want {
			t.Errorf("SubjectPronoun(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestShippedRolesAnnotate keeps the roster honest: every seat the shipped
// person package offers has to render a complete annotation.
func TestShippedRolesAnnotate(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for name, role := range p.Roles {
		displayName := p.RoleDisplayName(name)
		for _, seat := range role.Seats {
			annotation := SeatAnnotation(seat.Name, seat.Pronouns, displayName)
			if annotation == "" {
				t.Errorf("role %q seat %q renders no annotation", name, seat.Selector())
				continue
			}
			if SubjectPronoun(seat.Pronouns) == "" {
				t.Errorf("role %q seat %q has no pronouns", name, seat.Selector())
			}
			if displayName == "" {
				t.Errorf("role %q has no display name", name)
			}
		}
	}
}
