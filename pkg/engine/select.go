package engine

import (
	"fmt"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// LockedStages re-exports pkg/config's list so this package can refuse a tampered
// config by checking the stage name directly — defense in depth.
var LockedStages = config.LockedStages

// Resolve walks stage's Engine/Fallback/Then chain and returns the winning engine name
// — "none", "host", "api", "advisor", or the "local" alias.
func Resolve(cfg *config.Config, ledger Ledger, clock func() time.Time, stage string) (name, reason string, err error) {
	st := cfg.Pipeline[stage]
	if err := checkLocked(stage, st); err != nil {
		return "", "", err
	}
	for _, cand := range chain(cfg, stage, st) {
		if ok, why := available(cfg, ledger, clock, cand); ok {
			return cand, why, nil
		}
	}
	return "none", "no candidate in the chain was available; degrading to none", nil
}

// Select is Resolve narrowed to a Tier for callers that only need to know which
// implementation to construct — "local" maps to TierAPI.
func Select(cfg *config.Config, ledger Ledger, clock func() time.Time, stage string) (Tier, string, error) {
	name, reason, err := Resolve(cfg, ledger, clock, stage)
	if err != nil {
		return "", "", err
	}
	return tierOf(name), reason, nil
}

// Exhausted reports whether stage's chain names a metered tier (api/advisor) that is
// currently out of budget.
func Exhausted(cfg *config.Config, ledger Ledger, clock func() time.Time, stage string) bool {
	st := cfg.Pipeline[stage]
	for _, cand := range chain(cfg, stage, st) {
		if cand != "api" && cand != "advisor" {
			continue
		}
		if ok, _ := available(cfg, ledger, clock, cand); !ok {
			return true
		}
	}
	return false
}

func tierOf(name string) Tier {
	if name == "local" {
		return TierAPI
	}
	return Tier(name)
}

// checkLocked is the defense-in-depth pkg/config/validate.go's LockedStageError already
// enforces at load time.
func checkLocked(stage string, st config.Stage) error {
	if !isLocked(stage) {
		return nil
	}
	for _, v := range []string{st.Engine, st.Fallback, st.Then} {
		if v != "" && v != "none" {
			return fmt.Errorf("engine: pipeline.%s: %q is not allowed — %s are locked to "+
				"\"none\" (T0 static core)", stage, v, LockedStages)
		}
	}
	return nil
}

func isLocked(stage string) bool {
	for _, s := range LockedStages {
		if s == stage {
			return true
		}
	}
	return false
}

// chain lists the stage's candidates in priority order.
func chain(cfg *config.Config, stage string, st config.Stage) []string {
	var out []string
	if st.Engine != "" {
		out = append(out, st.Engine)
	} else if cfg.Engines.Default != "" {
		out = append(out, cfg.Engines.Default)
	}
	if st.Fallback != "" {
		out = append(out, st.Fallback)
	}
	if st.Then != "" {
		out = append(out, st.Then)
	}
	if len(out) == 0 {
		out = []string{"none"}
	}
	return out
}
