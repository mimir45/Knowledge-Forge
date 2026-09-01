package engine

import (
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// loadPreset mirrors pkg/config/chain_test.go's idiom (Preset → layer → Load), pointed at
// the packaged preset file directly via EnvPath rather than pkg/config's unexported
// marshalForTest — pkg/engine has no access to that helper and should not need one.
func loadPreset(t *testing.T, name string) *config.Config {
	t.Helper()
	t.Setenv(config.EnvVar, "")
	cfg, err := config.Load(config.Options{
		ProjectDir: t.TempDir(), HomeDir: t.TempDir(),
		EnvPath: "../../config/presets/" + name + ".md",
	})
	if err != nil {
		t.Fatalf("Load(%s): %v", name, err)
	}
	return cfg
}

// TestTrailEntriesMatchSchema asserts every (stage,tier) pair the four engine presets can
// produce — across Engine, Fallback and Then, since a fallback firing is a real trail
// entry too — matches references/schema.yaml's engine_trail item_pattern. A copy of the
// regex here could drift from the schema silently; this test reads the schema's own.
func TestTrailEntriesMatchSchema(t *testing.T) {
	schema, err := vault.LoadSchema()
	if err != nil {
		t.Fatal(err)
	}
	re := schema.EngineTrailPattern()
	if re == nil {
		t.Fatal("schema has no engine_trail field")
	}
	for _, preset := range []string{"offline", "claude-only", "byo-api", "max"} {
		cfg := loadPreset(t, preset)
		checkPresetTrail(t, preset, cfg, re)
	}
}

func checkPresetTrail(t *testing.T, preset string, cfg *config.Config, re interface{ MatchString(string) bool }) {
	t.Helper()
	for stage, st := range cfg.Pipeline {
		for _, name := range []string{st.Engine, st.Fallback, st.Then} {
			if name == "" {
				continue
			}
			entry, ok := TrailEntry(stage, string(tierOf(name)))
			if !ok {
				continue // one of the four unstamped stages — nothing to check
			}
			if !re.MatchString(entry) {
				t.Errorf("%s: %q does not match engine_trail's item_pattern", preset, entry)
			}
		}
	}
}
