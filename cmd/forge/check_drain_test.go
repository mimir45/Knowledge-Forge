package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/store"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// TestDrainAdvisorQueueClearsFlagAndBooksSpend is drain's happy path: a real (httptest)
// advisor call succeeds, so the flag comes off and the spend lands in the same store
// spentToday reads — mirroring TestEngineRunHitsRealHTTPAndBooksSpend's approach for the
// api tier.
func TestDrainAdvisorQueueClearsFlagAndBooksSpend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(critiqueHandler))
	defer srv.Close()

	root := fixtureCopy(t)
	rel := "concepts/hibernate.md"
	queueFixtureNote(t, root, rel)

	d := drainTestData(t, root, drainConfig(srv.URL))
	drainAdvisorQueue(d)

	assertNotQueued(t, root, rel)
	drainAssertSpent(t, root, 0.02)
}

// TestDrainAdvisorQueueSkipsOffline confirms --offline never dials the network: the flag
// must survive untouched, the same way a skipped deadlinks.md probe leaves its input alone.
func TestDrainAdvisorQueueSkipsOffline(t *testing.T) {
	root := fixtureCopy(t)
	rel := "concepts/hibernate.md"
	queueFixtureNote(t, root, rel)

	d := drainTestData(t, root, drainConfig("http://unused.invalid"))
	d.cfg.offline = true
	drainAdvisorQueue(d)

	n, err := vault.Load(filepath.Join(root, rel), rel)
	if err != nil {
		t.Fatal(err)
	}
	if !isQueued(n) {
		t.Error("offline drain cleared pending_advisor; it must leave the queue untouched")
	}
}

// TestDrainAdvisorQueueRequiresBothGates confirms the flag being true, alone, dispatches
// nothing without both cfg.Check.DrainAdvisorQueue and on_exhausted:queue.
func TestDrainAdvisorQueueRequiresBothGates(t *testing.T) {
	root := fixtureCopy(t)
	rel := "concepts/hibernate.md"
	queueFixtureNote(t, root, rel)

	cfg := drainConfig("http://unused.invalid")
	cfg.Check.DrainAdvisorQueue = false
	d := drainTestData(t, root, cfg)
	drainAdvisorQueue(d)

	n, err := vault.Load(filepath.Join(root, rel), rel)
	if err != nil {
		t.Fatal(err)
	}
	if !isQueued(n) {
		t.Error("drain ran with DrainAdvisorQueue false; both gates must be required")
	}
}

// TestDrainAdvisorQueueStopsWhenAlreadyExhausted is the regression test for the bug the
// advisor review caught: drainOne always dispatches to the advisor tier regardless of
// pipeline.synthesize's configured engine, so the budget guard must check the advisor
// ledger directly. This config mirrors the packaged default's pipeline.synthesize:{engine:
// host} — under the old engine.Exhausted(cfg, st, clock, "synthesize") guard, "host" isn't
// api/advisor, so the chain walk would have found nothing metered and never stopped.
func TestDrainAdvisorQueueStopsWhenAlreadyExhausted(t *testing.T) {
	root := fixtureCopy(t)
	rel := "concepts/hibernate.md"
	queueFixtureNote(t, root, rel)

	cfg := drainConfig("http://unused.invalid")
	cfg.Pipeline["synthesize"] = config.Stage{Engine: "host"}

	st, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Spend("advisor", 1.00, time.Now); err != nil {
		t.Fatal(err)
	}
	st.Close()

	d := drainTestData(t, root, cfg)
	drainAdvisorQueue(d)

	n, err := vault.Load(filepath.Join(root, rel), rel)
	if err != nil {
		t.Fatal(err)
	}
	if !isQueued(n) {
		t.Error("drain dispatched despite the advisor tier already at its daily cap")
	}
}

func critiqueHandler(w http.ResponseWriter, r *http.Request) {
	critique := `{"disputed":[],"missing":[],"confidence":"medium","patch":""}`
	_ = json.NewEncoder(w).Encode(map[string]any{
		"output": critique, "tokens": 10, "cost_usd": 0.02,
	})
}

func drainConfig(baseURL string) *config.Config {
	return &config.Config{
		Pipeline: map[string]config.Stage{"synthesize": {Engine: "advisor"}},
		Check:    config.Check{DrainAdvisorQueue: true},
		Engines: config.Engines{
			API:     config.API{Provider: "openai", Model: "test", BaseURL: baseURL},
			Advisor: config.Advisor{Model: "test-advisor"},
			Budget:  config.Budget{AdvisorUSDPerDay: 1.00, OnExhausted: "queue"},
		},
	}
}

// queueFixtureNote stamps pending_advisor: true directly into the fixture copy's raw
// bytes, ahead of any vault.Load — the same on-disk state queueNote would have left.
func queueFixtureNote(t *testing.T, root, rel string) {
	t.Helper()
	p := filepath.Join(root, rel)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := strings.Replace(string(b), "status: active\n", "status: active\npending_advisor: \"true\"\n", 1)
	if s == string(b) {
		t.Fatalf("queueFixtureNote: %s has no 'status: active' anchor line to stamp after", rel)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func drainTestData(t *testing.T, root string, cfg *config.Config) *checkData {
	t.Helper()
	rels, err := vault.Walk(root)
	if err != nil {
		t.Fatal(err)
	}
	notes := make([]*vault.Note, 0, len(rels))
	for _, rel := range rels {
		n, err := vault.Load(filepath.Join(root, rel), rel)
		if err != nil {
			t.Fatal(err)
		}
		notes = append(notes, n)
	}
	return &checkData{cfg: checkCfg{config: cfg}, root: root, notes: notes}
}

func assertNotQueued(t *testing.T, root, rel string) {
	t.Helper()
	n, err := vault.Load(filepath.Join(root, rel), rel)
	if err != nil {
		t.Fatal(err)
	}
	if isQueued(n) {
		t.Error("drain left pending_advisor: true after a successful advisor call")
	}
}

func drainAssertSpent(t *testing.T, root string, want float64) {
	t.Helper()
	st, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	remaining, err := st.Remaining("advisor", 1.00, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if got := 1.00 - remaining; got < want-0.001 || got > want+0.001 {
		t.Errorf("spent = %.2f, want %.2f", got, want)
	}
}
