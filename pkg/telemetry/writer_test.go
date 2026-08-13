package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendRoundTrip(t *testing.T) {
	dir := t.TempDir()
	ev := Event{TS: time.Now().UTC().Truncate(time.Second), Event: "ask", QHash: QHash("q"),
		Topic: "kafka-consumer-rebalancing", Stack: []string{"kafka"}, Decision: "CREATE",
		RecallTopScore: 0.31, Sources: 6, Project: "order-service"}
	if err := Append(dir, ev); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".forge", "log.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var got Event
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(b))), &got); err != nil {
		t.Fatal(err)
	}
	if got.Topic != ev.Topic || got.QHash != ev.QHash || !got.TS.Equal(ev.TS) {
		t.Fatalf("round trip mismatch: got %+v want %+v", got, ev)
	}
}

func TestAppendAppendsRatherThanOverwrites(t *testing.T) {
	dir := t.TempDir()
	Append(dir, Event{Topic: "a"})
	Append(dir, Event{Topic: "b"})
	b, err := os.ReadFile(filepath.Join(dir, ".forge", "log.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %q", len(lines), string(b))
	}
}
