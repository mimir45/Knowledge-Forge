package engine

import (
	"testing"
	"time"

	"knowledge-forge/pkg/config"
)

// fakeLedger is a minimal in-memory Ledger, so select_test.go never needs pkg/store —
// pkg/engine stays a leaf; only cmd/forge wires the real *store.Store in.
type fakeLedger struct{ spent map[string]float64 }

func (f *fakeLedger) Spend(tier string, usd float64, _ func() time.Time) error {
	f.spent[tier] += usd
	return nil
}

func (f *fakeLedger) Remaining(tier string, capUSD float64, _ func() time.Time) (float64, error) {
	return capUSD - f.spent[tier], nil
}

func clock() time.Time { return time.Now() }

func TestResolveLockedStageEngineTamper(t *testing.T) {
	cfg := &config.Config{Pipeline: map[string]config.Stage{"write": {Engine: "api"}}}
	if _, _, err := Resolve(cfg, nil, clock, "write"); err == nil {
		t.Fatal("want an error for pipeline.write.engine: api")
	}
}

// TestResolveLockedStageFallbackTamper is the defense-in-depth case pkg/config/validate.go
// alone does not cover: a tamper hiding behind .fallback rather than .engine.
func TestResolveLockedStageFallbackTamper(t *testing.T) {
	cfg := &config.Config{Pipeline: map[string]config.Stage{
		"index": {Engine: "none", Fallback: "api"},
	}}
	if _, _, err := Resolve(cfg, nil, clock, "index"); err == nil {
		t.Fatal("want an error for pipeline.index.fallback: api")
	}
}

func TestResolveOffline(t *testing.T) {
	cfg := &config.Config{Pipeline: map[string]config.Stage{"research": {Engine: "none"}}}
	name, _, err := Resolve(cfg, nil, clock, "research")
	if err != nil || name != "none" {
		t.Fatalf("Resolve = %q, %v, want none", name, err)
	}
}

func TestResolveFallsThroughToDefaultWhenStageUnset(t *testing.T) {
	cfg := &config.Config{Engines: config.Engines{Default: "host"}}
	name, _, err := Resolve(cfg, nil, clock, "research")
	if err != nil || name != "host" {
		t.Fatalf("Resolve = %q, %v, want host", name, err)
	}
}

func TestResolveFallsBackWhenAPIBudgetExhausted(t *testing.T) {
	cfg := &config.Config{
		Engines:  config.Engines{Budget: config.Budget{APIUSDPerDay: 1.00}},
		Pipeline: map[string]config.Stage{"research": {Engine: "api", Fallback: "host"}},
	}
	l := &fakeLedger{spent: map[string]float64{"api": 1.00}}
	name, _, err := Resolve(cfg, l, clock, "research")
	if err != nil || name != "host" {
		t.Fatalf("Resolve = %q, %v, want host (api exhausted)", name, err)
	}
}

func TestResolveLocalRequiresBaseURL(t *testing.T) {
	cfg := &config.Config{Pipeline: map[string]config.Stage{
		"research": {Engine: "local", Then: "host"},
	}}
	cfg.Engines.Local.Enabled = true // BaseURL left empty
	name, _, err := Resolve(cfg, nil, clock, "research")
	if err != nil || name != "host" {
		t.Fatalf("Resolve = %q, %v, want host (local has no base_url)", name, err)
	}
}

func TestSelectMapsLocalToTierAPI(t *testing.T) {
	cfg := &config.Config{Pipeline: map[string]config.Stage{"research": {Engine: "local"}}}
	cfg.Engines.Local = config.Local{Enabled: true, BaseURL: "http://x"}
	tier, _, err := Select(cfg, nil, clock, "research")
	if err != nil || tier != TierAPI {
		t.Fatalf("Select = %q, %v, want api", tier, err)
	}
}

func TestExhaustedTrueWhenAPIBudgetIsSpent(t *testing.T) {
	cfg := &config.Config{
		Engines:  config.Engines{Budget: config.Budget{APIUSDPerDay: 1.00}},
		Pipeline: map[string]config.Stage{"research": {Engine: "api"}},
	}
	l := &fakeLedger{spent: map[string]float64{"api": 1.00}}
	if !Exhausted(cfg, l, clock, "research") {
		t.Error("Exhausted = false, want true (api budget fully spent)")
	}
}

func TestExhaustedFalseWhenStageIsNotMetered(t *testing.T) {
	cfg := &config.Config{Pipeline: map[string]config.Stage{"research": {Engine: "host"}}}
	if Exhausted(cfg, nil, clock, "research") {
		t.Error("Exhausted = true, want false (host is never metered)")
	}
}
