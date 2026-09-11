package person

import "testing"

// A bare substring matches "rope" inside "roped down", a verb rather than a
// material. The frontend seat found that against the shipped six.
func TestContainsWordRejectsAMatchInsideALongerWord(t *testing.T) {
	cases := []struct {
		noun, clause string
	}{
		{"rope", "a neat stack of square cargo containers roped down on top"},
		{"die", "the creature died in the second cut"},
		{"glass", "a fibreglass hull, seamless"},
		{"collar", "two collarbones showing through the fur"},
	}
	for _, tc := range cases {
		if containsWord(tc.clause, tc.noun) {
			t.Errorf("%q matched inside a longer word in %q", tc.noun, tc.clause)
		}
	}
}

// The control: every noun the roster actually declares must still match, or
// the tightening trades a false positive for a false negative.
func TestContainsWordKeepsEveryRealNoun(t *testing.T) {
	vera := "a seven-spoked iron ship's helm wheel mounted upright on the side of its " +
		"shell, a neat stack of square cargo containers roped down on top of its shell"
	cases := []struct {
		noun, clause string
	}{
		{"helm wheel", vera},
		{"cargo containers", vera},
		{"hammer", "a heavy timber-framing hammer gripped in one forepaw"},
		{"vessel", "standing inside a tall rounded glass vessel that rises past its shoulders"},
		{"collars", "brass collars banding the rim and the foot of the glass"},
		{"die", "a large twenty-sided die held up in one forelimb"},
		{"binoculars", "a pair of weathered brass binoculars raised in both hands"},
		{"lantern", "a weathered brass lantern held up in one raised hand"},
		{"glass", "a live fire burning steady and banked behind its glass"},
	}
	for _, tc := range cases {
		if !containsWord(tc.clause, tc.noun) {
			t.Errorf("real noun %q no longer matches its own clause", tc.noun)
		}
	}
}

// Punctuation and possessives bound a word too, so a noun at a comma or an
// apostrophe still matches rather than needing a space on both sides.
func TestContainsWordTreatsPunctuationAsABoundary(t *testing.T) {
	for _, tc := range []struct{ noun, clause string }{
		{"wheel", "the wheel, mounted upright"},
		{"wheel", "the wheel's seven spokes"},
		{"lantern", "a lantern."},
		{"die", "(a die)"},
	} {
		if !containsWord(tc.clause, tc.noun) {
			t.Errorf("%q did not match across punctuation in %q", tc.noun, tc.clause)
		}
	}
}
