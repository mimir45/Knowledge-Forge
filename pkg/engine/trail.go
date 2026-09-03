package engine

// stampable lists the five pipeline stages this package actually audits by tier.
// cfg.Pipeline has nine keys; intake, plan.
var stampable = map[string]bool{
	"recall": true, "research": true, "write": true,
	"verify": true, "index": true,
}

// TrailEntry renders one engine_trail list item for stage=tier, and reports whether the
// pair is stampable at all.
func TrailEntry(stage, tier string) (entry string, ok bool) {
	if !stampable[stage] {
		return "", false
	}
	return stage + "=" + tier, true
}
