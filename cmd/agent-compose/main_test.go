package main

import (
	"reflect"
	"strings"
	"testing"

	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/bundle"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/compose"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/resolver"
	"forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/internal/schema"
)

func TestDispatchArgs(t *testing.T) {
	cases := map[string]struct{ in, want []string }{
		"canonical name untouched": {
			[]string{"/usr/local/bin/agent-compose", "describe", "x"},
			[]string{"/usr/local/bin/agent-compose", "describe", "x"},
		},
		"acompose injects compose": {
			[]string{"/opt/homebrew/bin/acompose"},
			[]string{"/opt/homebrew/bin/acompose", "compose"},
		},
		"acompose keeps trailing args": {
			[]string{"acompose", "--", "claude"},
			[]string{"acompose", "compose", "--", "claude"},
		},
		"windows exe suffix": {
			[]string{`C:\shims\acompose.exe`, "req.kdl"},
			[]string{`C:\shims\acompose.exe`, "compose", "req.kdl"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := dispatchArgs(tc.in); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestPrintSummaryUsesSlashSeparators(t *testing.T) {
	t.Parallel()
	result := &compose.Result{
		Bundle: &bundle.Result{Key: "abc123", Dir: "/tmp/bundle", Reused: true},
		Resolution: &resolver.Resolution{
			Request: &schema.Request{
				Role:       "engineer",
				ModelClass: schema.ModelClassFrontier,
				Delivery:   "native-skills",
			},
			Personalities: []string{"curious", "grounded", "meticulous"},
			FavoriteColor: "#90a66a",
			SourceIDs:     []string{"person:kai", "aos-public"},
			Decisions: []resolver.Decision{
				{Outcome: resolver.OutcomeSelected},
				{Outcome: resolver.OutcomeExcluded},
				{Outcome: resolver.OutcomeShadowed},
				{Outcome: resolver.OutcomeDelivered},
			},
		},
	}

	var output strings.Builder
	printSummary(&output, result)
	got := output.String()
	for _, want := range []string{
		"role engineer // model class frontier // personalities curious, grounded, meticulous // melded #90a66a // delivery native-skills",
		"decisions: 1 selected // 1 excluded // 1 shadowed // 1 delivered",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("summary missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "·") {
		t.Fatalf("summary retained middle-dot separator:\n%s", got)
	}
}

func TestPrintVerificationUsesBoundedCounts(t *testing.T) {
	t.Parallel()
	verification := &bundle.Verification{
		Files: 128,
		Identities: []bundle.Identity{
			{Source: "person:kai", Skill: "personality-curious"},
			{Source: "aos-public", Skill: "coding-go"},
		},
	}

	var output strings.Builder
	printVerification(&output, verification)
	if got, want := output.String(), "bundle verified: 2 skills // 128 files\n"; got != want {
		t.Fatalf("verification output = %q, want %q", got, want)
	}
}
