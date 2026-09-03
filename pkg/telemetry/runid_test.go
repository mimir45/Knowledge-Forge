package telemetry

import "testing"

func TestNewRunIDLength(t *testing.T) {
	if got := len(NewRunID()); got != 32 { // 16 random bytes, hex-encoded
		t.Fatalf("want 32 hex chars, got %d", got)
	}
}

// TestNewRunIDIsUnique pins the one property NewRunID actually promises: identity, not
// a counter or a timestamp.
func TestNewRunIDIsUnique(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		id := NewRunID()
		if seen[id] {
			t.Fatalf("collision after %d draws: %s", i, id)
		}
		seen[id] = true
	}
}
