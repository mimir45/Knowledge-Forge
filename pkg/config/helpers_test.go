package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func mustLoad(t *testing.T, o Options) *Config {
	t.Helper()
	c, err := Load(o)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return c
}

// marshalForTest renders a preset overlay back to YAML so it can be written as a layer
// file.
func marshalForTest(t *testing.T, m map[string]any) string {
	t.Helper()
	b, err := yaml.Marshal(m)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}
	return string(b)
}

// assertLockedNone checks the invariant directly on the merged result rather than
// trusting that validate ran.
func assertLockedNone(t *testing.T, c *Config) {
	t.Helper()
	for _, stage := range LockedStages {
		s, ok := c.Pipeline[stage]
		if !ok {
			t.Fatalf("pipeline.%s is missing from the merged config", stage)
		}
		if s.Engine != "none" {
			t.Fatalf("pipeline.%s.engine = %q, want none", stage, s.Engine)
		}
	}
}

// assertBlames requires the error to name both the offending file and the stage. Naming
// only the stage leaves the user grepping four layers for who set it.
func assertBlames(t *testing.T, err error, path, stage string) {
	t.Helper()
	msg := err.Error()
	if !contains(msg, stage) {
		t.Fatalf("error %q does not name stage %q", msg, stage)
	}
	if !contains(msg, path) {
		t.Fatalf("error %q does not name the file %q that set it", msg, path)
	}
}

func contains(haystack, needle string) bool { return strings.Contains(haystack, needle) }
