package engine

// stampable lists the six pipeline stages references/schema.yaml's engine_trail
// item_pattern accepts. cfg.Pipeline has nine keys; intake, plan, synthesize and link are
// host-orchestration bookkeeping the skill does regardless of tier, not audited model-call
// decisions, so they are deliberately absent here (see docs/BACKLOG.md for the widening).
var stampable = map[string]bool{
	"recall": true, "research": true, "write": true,
	"verify": true, "index": true,
}

// TrailEntry renders one engine_trail list item for stage=tier, and reports whether the
// pair is stampable at all. verify:{engine:advisor} stamps as "verify=advisor" — verify is
// the pipeline stage that actually exists; the schema's "critique" alternative names the
// tier's older label, not a ninth stage, and this package never emits it.
func TrailEntry(stage, tier string) (entry string, ok bool) {
	if !stampable[stage] {
		return "", false
	}
	return stage + "=" + tier, true
}
