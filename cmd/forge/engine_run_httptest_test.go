package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-forge/pkg/config"
	"knowledge-forge/pkg/store"
)

// TestEngineRunHitsRealHTTPAndBooksSpend is the api tier's one real-network exercise: an
// httptest.Server stands in for a provider, and runEngineRun's buildEngine wires
// http.DefaultTransport straight to it — no injected RoundTripper needed, because a
// listener on localhost is real HTTP, not the network call pkg/engine's tests must avoid.
func TestEngineRunHitsRealHTTPAndBooksSpend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(costingHandler))
	defer srv.Close()

	root := fixtureCopy(t)
	cfg := researchConfig(srv.URL)
	if code := runEngineRun(root, cfg, "research", "hello", ""); code != 0 {
		t.Fatalf("runEngineRun exit %d", code)
	}
	assertSpent(t, root, 0.05)
}

// costingHandler is the fake provider: the uniform envelope api.go's do() reads,
// regardless of which provider payload shape hit it.
func costingHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]any{
		"output": "ok", "tokens": 10, "cost_usd": 0.05,
	})
}

func researchConfig(baseURL string) *config.Config {
	return &config.Config{
		Pipeline: map[string]config.Stage{"research": {Engine: "api"}},
		Engines: config.Engines{
			API:    config.API{Provider: "openai", Model: "test", BaseURL: baseURL},
			Budget: config.Budget{APIUSDPerDay: 1.00, OnExhausted: "queue"},
		},
	}
}

// assertSpent reads the spend back the way spentToday does — cap minus Remaining — so the
// assertion exercises the same store the collector reads, not a second code path.
func assertSpent(t *testing.T, root string, want float64) {
	t.Helper()
	st, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	remaining, err := st.Remaining("api", 1.00, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if got := 1.00 - remaining; got < want-0.001 || got > want+0.001 {
		t.Errorf("spent = %.2f, want %.2f", got, want)
	}
}
