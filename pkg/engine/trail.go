package engine

// stampable lists the five pipeline stages this package actually audits by tier.
// cfg.Pipeline has nine keys; intake, plan, synthesize and link are host-orchestration
// bookkeeping the skill does regardless of tier, not audited model-call decisions, so
// they stay deliberately absent here — pkg/engine/trail_entry_test.go pins this. A
// later schema change only widened what references/schema.yaml's engine_trail item_pattern *accepts* (all
// nine real stage names, and dropped the schema's old "critique" alternative, which was
// never a real stage — only the advisor engine's operating mode, engines.advisor.mode);
// it did not change which stages this package stamps.
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
