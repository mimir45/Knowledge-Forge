package main

import (
	"strings"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// queueNote is on_exhausted:queue's whole job: stamp pending_advisor:true and nothing
// else about the note. This is the write path SetScalars' doc-comment now names.
func TestQueueNoteStampsPendingAdvisor(t *testing.T) {
	root := fixtureCopy(t)
	rel := "concepts/hibernate.md"
	if err := queueNote(root, rel); err != nil {
		t.Fatal(err)
	}
	n, err := vault.Load(root+"/"+rel, rel)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(n.FM.Str("pending_advisor"), "true") {
		t.Errorf("pending_advisor = %q, want true", n.FM.Str("pending_advisor"))
	}
}

// TestOnExhaustedBehaviorDiverges is B-023's behaviour half: before this, all three
// on_exhausted values reached the same silent none-fallthrough. Each case forces the
// same exhausted-budget path (cap $0.00, no fallback in the chain) and checks the exit
// code and, for queue, the note write that used to be the only distinguishable effect.
func TestOnExhaustedBehaviorDiverges(t *testing.T) {
	cases := []struct {
		name        string
		onExhausted string
		wantCode    int
		wantQueued  bool
	}{
		{"stop halts non-zero", "stop", 1, false},
		{"queue stamps and falls through", "queue", 0, true},
		{"degrade falls through silently", "degrade", 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := fixtureCopy(t)
			rel := "concepts/hibernate.md"
			cfg := exhaustedConfig(c.onExhausted)
			if code := runEngineRun(root, cfg, "research", "hello", rel); code != c.wantCode {
				t.Errorf("exit = %d, want %d", code, c.wantCode)
			}
			n, err := vault.Load(root+"/"+rel, rel)
			if err != nil {
				t.Fatal(err)
			}
			if queued := strings.EqualFold(n.FM.Str("pending_advisor"), "true"); queued != c.wantQueued {
				t.Errorf("pending_advisor queued = %v, want %v", queued, c.wantQueued)
			}
		})
	}
}

// exhaustedConfig points "research" at api with a $0.00 cap and no fallback, so
// engine.Resolve always degrades to none and engine.Exhausted is always true — the
// precondition every onExhausted case needs, independent of the value under test.
func exhaustedConfig(onExhausted string) *config.Config {
	return &config.Config{
		Pipeline: map[string]config.Stage{"research": {Engine: "api"}},
		Engines: config.Engines{
			API:    config.API{Provider: "openai", Model: "test", BaseURL: "http://127.0.0.1:1"},
			Budget: config.Budget{APIUSDPerDay: 0.00, OnExhausted: onExhausted},
		},
	}
}
