package engine

import (
	"fmt"
	"time"

	"knowledge-forge/pkg/config"
)

// LockedStages mirrors pkg/config.LockedStages so this package can refuse a tampered
// config without importing pkg/config's validate.go — defense in depth, not a shortcut
// around it: config.Load already refuses to start on the same violation.
var LockedStages = []string{"recall", "write", "index"}

// Resolve walks stage's Engine/Fallback/Then chain and returns the winning engine name —
// "none", "host", "api", "advisor", or the "local" alias — plus a human reason a caller
// can surface (this is what makes `forge engine select --json` show *why* offline fell
// back to none instead of just that it did). ledger may be nil, which fails budget checks
// open; callers that care about an honest answer must pass a real one.
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
// implementation to construct — "local" maps to TierAPI, since it is api.go under a
// different base_url, not a fifth Engine.
func Select(cfg *config.Config, ledger Ledger, clock func() time.Time, stage string) (Tier, string, error) {
	name, reason, err := Resolve(cfg, ledger, clock, stage)
	if err != nil {
		return "", "", err
	}
	return tierOf(name), reason, nil
}

// Exhausted reports whether stage's chain names a metered tier (api/advisor) that is
// currently out of budget — the distinction Resolve's single winning name loses once it
// has already degraded to "none". `on_exhausted: queue` needs this to tell "no budget
// today" apart from "nothing metered was ever configured here" before it queues a note.
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
// enforces at load time. It inspects Engine, Fallback, *and* Then — a tamper that hides
// behind pipeline.write.fallback rather than .engine must be caught here too, or this
// layer is decorative.
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

// chain lists the stage's candidates in priority order, falling back to engines.default
// when the stage sets no Engine of its own — an unset stage is not a locked-to-none claim,
// it is silence, and cfg.Engines.Default is what fills it.
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
