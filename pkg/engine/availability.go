package engine

import (
	"fmt"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// available reports whether candidate name can run right now, and why (or why not) — the
// reason string is what select.go hands back for `forge engine select --json`'s "reason".
func available(cfg *config.Config, ledger Ledger, clock func() time.Time, name string) (bool, string) {
	switch name {
	case "none":
		return true, "static core, zero model calls"
	case "host":
		return true, "runs in the Claude Code session, not metered here"
	case "local":
		return availableLocal(cfg)
	case "api":
		return availableMetered(ledger, clock, "api", cfg.Engines.Budget.APIUSDPerDay)
	case "advisor":
		return availableMetered(ledger, clock, "advisor", cfg.Engines.Budget.AdvisorUSDPerDay)
	default:
		return false, fmt.Sprintf("unknown engine %q", name)
	}
}

// availableLocal requires both Enabled and a non-empty BaseURL — Local currently has no
// BaseURL default, so a bare `enabled: true` must not resolve to an API call against "".
func availableLocal(cfg *config.Config) (bool, string) {
	l := cfg.Engines.Local
	if !l.Enabled {
		return false, "engines.local.enabled is false"
	}
	if l.BaseURL == "" {
		return false, "engines.local.base_url is not set"
	}
	return true, "local model server configured"
}

// availableMetered fails open (true) when ledger is nil — cmd/forge/engine_cmd.go always
// passes a real one; a nil ledger only reaches here from a caller that has not opened a
// budget store, and pkg/engine cannot itself tell the difference between "budget already
// checked upstream" and "forgot to open the store", so it does not pretend to.
func availableMetered(ledger Ledger, clock func() time.Time, tier string, capUSD float64) (bool, string) {
	if ledger == nil {
		return true, tier + " budget not checked (no ledger)"
	}
	remaining, err := ledger.Remaining(tier, capUSD, clock)
	if err != nil {
		return false, fmt.Sprintf("%s budget lookup failed: %v", tier, err)
	}
	if remaining <= 0 {
		return false, fmt.Sprintf("%s budget exhausted for today (cap $%.2f)", tier, capUSD)
	}
	return true, fmt.Sprintf("%s budget: $%.2f remaining", tier, remaining)
}
