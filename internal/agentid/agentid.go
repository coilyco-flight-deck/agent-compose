// Package agentid reads the dictatable short id naming one running agent
// session, so a terminal surface can show which agent it is looking at.
// Contract and rationale: docs/identity.md.
package agentid

import (
	"os"
	"strings"
)

// Duplicated rather than imported: agentic-os owns the canonical definition and
// cross-language vectors, and this module only ever recognises an id.
const (
	idLetters   = "abcdefghjkmpqrstuvwxyz"
	idDigits    = "456789"
	idLen       = 4
	idLetterLen = 2
)

// SessionEnv is the variable `aos` exports for the active native session.
const SessionEnv = "AOS_NATIVE_SESSION"

// Valid reports whether raw is canonical: two dictatable letters, two digits.
func Valid(raw string) bool {
	if len(raw) != idLen {
		return false
	}
	for _, r := range raw[:idLetterLen] {
		if !strings.ContainsRune(idLetters, r) {
			return false
		}
	}
	for _, r := range raw[idLetterLen:] {
		if !strings.ContainsRune(idDigits, r) {
			return false
		}
	}
	return true
}

// FromEnv returns the session's short id, or "" when there is none. It reads
// rather than mints, because the status line re-renders on every tick.
func FromEnv() string {
	return Normalize(os.Getenv(SessionEnv))
}

// Normalize trims and lowercases a dictated id, returning "" when the result is
// not canonical. A malformed value is dropped rather than displayed.
func Normalize(raw string) string {
	id := strings.ToLower(strings.TrimSpace(raw))
	if !Valid(id) {
		return ""
	}
	return id
}
