package telemetry

import (
	"crypto/rand"
	"encoding/hex"
)

// NewRunID mints the correlation key D1 outcome-joining needed: an opaque,
// collision-resistant identifier joining one `forge recall` call to the note write.
func NewRunID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
