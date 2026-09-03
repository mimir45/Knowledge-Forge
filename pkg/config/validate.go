package config

import (
	"fmt"
	"strings"
)

// LockedStages are the pipeline stages that accept engine `none` and nothing else.
var LockedStages = []string{"recall", "write", "index"}

// LockedStageError names the stage, the value, and the file the value came from.
type LockedStageError struct {
	Stage, Engine, Source string
}

func (e *LockedStageError) Error() string {
	return fmt.Sprintf("pipeline.%s: engine %q is not allowed — stages %s are locked to "+
		"engine \"none\" (the T0 static core makes zero model calls); set by %s",
		e.Stage, e.Engine, strings.Join(LockedStages, ", "), e.Source)
}

func validate(c *Config, layers []layer) error {
	if err := validateLockedStages(c, layers); err != nil {
		return err
	}
	if err := validateEnums(c); err != nil {
		return err
	}
	return validateThresholds(c)
}

// validateLockedStages runs on the *merged* config, not per layer.
func validateLockedStages(c *Config, layers []layer) error {
	for _, name := range LockedStages {
		st, ok := c.Pipeline[name]
		if !ok || st.Engine == "" || st.Engine == "none" {
			continue
		}
		return &LockedStageError{name, st.Engine, blame(layers, name)}
	}
	return nil
}

// blame reports the highest-precedence layer that set pipeline.<stage>.engine — the file
// whose line the user has to delete to start the binary.
func blame(layers []layer, stage string) string {
	for i := len(layers) - 1; i >= 0; i-- {
		if _, ok := stageEngine(layers[i].data, stage); ok {
			return layers[i].name
		}
	}
	return "an unknown layer"
}

func stageEngine(data map[string]any, stage string) (string, bool) {
	pipe, ok := data["pipeline"].(map[string]any)
	if !ok {
		return "", false
	}
	st, ok := pipe[stage].(map[string]any)
	if !ok {
		return "", false
	}
	e, ok := st["engine"].(string)
	return e, ok
}

// validateEnums rejects a value outside its documented set. An empty value passes: it
// means the key was never set, and the caller's own zero-value handling applies.
func validateEnums(c *Config) error {
	checks := []struct {
		key, got string
		allowed  []string
	}{
		{"trigger.mode", c.Trigger.Mode, []string{"ask", "auto", "manual"}},
		{"recall.strategy", c.Recall.Strategy, []string{"lexical", "hybrid"}},
		{"engines.budget.on_exhausted", c.Engines.Budget.OnExhausted, []string{"queue", "degrade", "stop"}},
		{"static.drift.trigger", c.Static.Drift.Trigger, []string{"git"}},
		{"verify.run_code", c.Verify.RunCode, []string{"auto", "never", "ask"}},
		{"write.diagrams", c.Write.Diagrams, []string{"mermaid", "ascii", "none"}},
		{"telemetry.scope", c.Telemetry.Scope, []string{"local", "team"}},
	}
	for _, ch := range checks {
		if err := oneOf(ch.key, ch.got, ch.allowed); err != nil {
			return err
		}
	}
	return nil
}

func oneOf(key, got string, allowed []string) error {
	if got == "" {
		return nil
	}
	for _, a := range allowed {
		if got == a {
			return nil
		}
	}
	return fmt.Errorf("%s: %q is not one of %s", key, got, strings.Join(allowed, ", "))
}

// validateThresholds enforces the ordering that the decision tree assumes.
func validateThresholds(c *Config) error {
	r := c.Recall
	if r.AnswerThreshold < r.UpdateThreshold {
		return fmt.Errorf("recall: answer_threshold (%.2f) is below update_threshold (%.2f) — "+
			"no score can ever resolve ANSWER_FROM_VAULT", r.AnswerThreshold, r.UpdateThreshold)
	}
	if r.UpdateThreshold < r.NeighbourMinScore {
		return fmt.Errorf("recall: update_threshold (%.2f) is below neighbour_min_score (%.2f) — "+
			"every neighbour would already be an extend target", r.UpdateThreshold, r.NeighbourMinScore)
	}
	return nil
}
