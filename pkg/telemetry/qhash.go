package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
)

// QHash one-way hashes a question so the log never carries raw question text, following
// the security invariant.
func QHash(question string) string {
	sum := sha256.Sum256([]byte(question))
	return hex.EncodeToString(sum[:])[:12]
}
