package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
)

// QHash one-way hashes a question so the log never carries raw question text — DESIGN
// §14's invariant. Truncated to 12 hex chars: enough to dedupe repeats across a single
// vault's ask volume without keeping more of the question around than a topic needs.
func QHash(question string) string {
	sum := sha256.Sum256([]byte(question))
	return hex.EncodeToString(sum[:])[:12]
}
