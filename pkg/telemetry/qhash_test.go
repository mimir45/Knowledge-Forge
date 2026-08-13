package telemetry

import "testing"

func TestQHashDeterministic(t *testing.T) {
	if QHash("kafka rebalancing") != QHash("kafka rebalancing") {
		t.Fatal("QHash should be deterministic for the same input")
	}
}

func TestQHashDiffersByInput(t *testing.T) {
	if QHash("a") == QHash("b") {
		t.Fatal("QHash should differ for different input")
	}
}

func TestQHashLength(t *testing.T) {
	if got := len(QHash("x")); got != 12 {
		t.Fatalf("want 12 hex chars, got %d", got)
	}
}
