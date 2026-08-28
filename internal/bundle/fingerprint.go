package bundle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Versioned so a change to the covered set cannot collide with an older hash.
const fingerprintVersion = "agent-compose.bundle-fingerprint.v1"

// Fingerprint names the composition a manifest describes, so a session
// transcript can point at an exact bundle without copying any of it. It is
// derived from manifest.json alone, so a reader holding the bundle recomputes
// it rather than trusting a recorded value. See docs/manifest-schema.md.
//
// It names a composition and does not attest to one. `verify` checks
// content digests for shape and never recomputes them against bytes, so a
// producer writing well-formed but untrue digests mints a well-formed but
// untrue fingerprint. Joining on it is sound, trusting it is not.
func Fingerprint(m Manifest) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00", fingerprintVersion)
	fmt.Fprintf(h, "role\x00%s\x00", m.Role)
	fmt.Fprintf(h, "role-skill\x00%s\x00%s\x00%s\x00",
		m.RoleSkill, m.RoleSkillSource, m.RoleSkillDigest)
	fmt.Fprintf(h, "model-tier\x00%s\x00", m.ModelTier)
	for _, p := range m.Personalities {
		fmt.Fprintf(h, "personality\x00%s\x00", p)
	}
	for _, b := range m.Boundaries {
		fmt.Fprintf(h, "boundary\x00%s\x00", b)
	}
	for _, s := range m.Sources {
		fmt.Fprintf(h, "source\x00%s\x00", s)
	}
	for _, c := range m.Content {
		fmt.Fprintf(h, "content\x00%s\x00%s\x00", c.ID, c.Digest)
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
