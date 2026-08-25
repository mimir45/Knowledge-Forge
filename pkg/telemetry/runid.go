package telemetry

import (
	"crypto/rand"
	"encoding/hex"
)

// NewRunID mints the correlation key BACKLOG B-035 needed: an opaque, collision-resistant
// identifier joining one `forge recall` call to the note write (`forge gate --run-id`)
// that may follow it minutes later, in a different process. 16 random bytes rather than a
// counter or a timestamp — a run_id must carry no ordering or wall-clock semantics a
// reader could infer meaning from, only identity.
//
// A read failure from crypto/rand means the OS entropy source is broken, not a condition
// `forge recall` should degrade an already-correctly-scored answer over — it returns ""
// instead, which every caller already treats as "no run_id", the same degradation an
// omitted --run-id on the gate side produces.
func NewRunID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
