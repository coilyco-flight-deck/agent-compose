package person

import "testing"

// An unauthored legal name must stay absent rather than defaulting to the
// model, which would let identity assert a selection it does not grant (#396).
func TestIdentitySentenceOmitsAnUnauthoredLegalName(t *testing.T) {
	got := IdentitySentence("Evie", "Applied Scientist", "", "Moss-Toad")
	want := "I am Evie, an agent-compose persona. My role is Applied Scientist. " +
		"My creature is the Moss-Toad."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdentitySentenceStatesAllFourParts(t *testing.T) {
	got := IdentitySentence("Gem", "Developer Advocate", "Mixpost", "Dragon-Butterfly")
	want := "I am Gem, an agent-compose persona. My role is Developer Advocate. " +
		"On this seat my legal name is Mixpost, and my creature is the Dragon-Butterfly."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Outside a projection there is no composed name, so nothing is fabricated.
func TestIdentitySentenceIsEmptyWithoutAName(t *testing.T) {
	if got := IdentitySentence("", "Applied Scientist", "Claude", "Moss-Toad"); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
