package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeLayer drops a frontmatter-only config at path, creating its directory. The format
// is the one DESIGN §10 fixes, so the tests exercise the same parser production does
// rather than a YAML shortcut.
func writeLayer(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("---\n"+body+"---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func homeCfg(dir string) string    { return filepath.Join(dir, ".forge", "forge.config.md") }
func projectCfg(dir string) string { return filepath.Join(dir, ".forge.config.md") }

// isolated returns Options pointing every optional layer at empty temp dirs, so a real
// ~/.forge/forge.config.md on the machine running the tests cannot reach them. Clearing
// $FORGE_CONFIG matters as much: an empty EnvPath falls through to the environment, and
// a developer who has one exported would otherwise see failures nobody else can.
func isolated(t *testing.T) Options {
	t.Helper()
	t.Setenv(EnvVar, "")
	return Options{ProjectDir: t.TempDir(), HomeDir: t.TempDir()}
}

// TestPresetsRoundTrip is the highest-value test in this package: it asserts that what
// `forge init` writes is something `config.Load` accepts. Every engine preset assigns
// model engines to pipeline stages, and three stages are locked to "none" — if a preset
// ever reached one of them, the default install would produce a config the binary
// refuses on the very next command.
func TestPresetsRoundTrip(t *testing.T) {
	for _, name := range PresetNames() {
		t.Run(name, func(t *testing.T) {
			o := isolated(t)
			p, err := Preset(name)
			if err != nil {
				t.Fatalf("Preset(%q): %v", name, err)
			}
			writeLayer(t, homeCfg(o.HomeDir), marshalForTest(t, p))
			assertLockedNone(t, mustLoad(t, o))
		})
	}
}

// TestPresetEmptyName is the default `forge init` path: --stack-preset defaults to "" and
// configDelta asks for it unconditionally, so an error here would break every install
// that did not name a stack.
func TestPresetEmptyName(t *testing.T) {
	p, err := Preset("")
	if err != nil {
		t.Fatalf("Preset(\"\"): %v", err)
	}
	if len(p) != 0 {
		t.Fatalf("Preset(\"\") = %v, want an empty overlay", p)
	}
}

func TestPresetUnknownName(t *testing.T) {
	if _, err := Preset("no-such-preset"); err == nil {
		t.Fatal("Preset(\"no-such-preset\") = nil error, want one")
	}
}

// TestPrecedence walks all four layers with the same key. Each higher layer must win,
// and Layers must report what was actually read — `forge config --layers` is the only
// way a user diagnoses "why is this setting not taking effect".
func TestPrecedence(t *testing.T) {
	o := isolated(t)
	env := filepath.Join(t.TempDir(), "env.md")
	writeLayer(t, homeCfg(o.HomeDir), "vault_path: /from/home\n")
	c := mustLoad(t, o)
	if c.VaultPath != "/from/home" {
		t.Fatalf("home layer: vault_path = %q", c.VaultPath)
	}
	writeLayer(t, projectCfg(o.ProjectDir), "vault_path: /from/project\n")
	if c = mustLoad(t, o); c.VaultPath != "/from/project" {
		t.Fatalf("project layer: vault_path = %q", c.VaultPath)
	}
	writeLayer(t, env, "vault_path: /from/env\n")
	o.EnvPath = env
	if c = mustLoad(t, o); c.VaultPath != "/from/env" {
		t.Fatalf("env layer: vault_path = %q", c.VaultPath)
	}
	if len(c.Layers) != 4 {
		t.Fatalf("Layers = %v, want all four", c.Layers)
	}
}

// TestPackagedAlwaysLowest guards the reason init writes only a delta: keys the user
// never decided must keep coming from the packaged layer, so a binary upgrade delivers
// new defaults instead of being shadowed by a stale full copy.
func TestPackagedAlwaysLowest(t *testing.T) {
	o := isolated(t)
	writeLayer(t, homeCfg(o.HomeDir), "vault_path: /somewhere\n")
	c := mustLoad(t, o)
	if c.Recall.AnswerThreshold != 0.85 {
		t.Fatalf("answer_threshold = %v, want the packaged 0.85", c.Recall.AnswerThreshold)
	}
	if c.Paths.Notes == "" || c.Paths.Index == "" {
		t.Fatalf("paths came back empty: %+v", c.Paths)
	}
	if c.Layers[0] != PackagedName {
		t.Fatalf("Layers[0] = %q, want %q", c.Layers[0], PackagedName)
	}
}

// TestMergeNarrowsLists is the rule that would be wrong under union semantics: a Go repo
// setting static.languages: [go] must not silently inherit java and typescript.
func TestMergeNarrowsLists(t *testing.T) {
	o := isolated(t)
	writeLayer(t, projectCfg(o.ProjectDir), "static:\n  languages: [go]\n")
	c := mustLoad(t, o)
	if len(c.Static.Languages) != 1 || c.Static.Languages[0] != "go" {
		t.Fatalf("static.languages = %v, want exactly [go]", c.Static.Languages)
	}
}

// TestMergeKeepsSiblingKeys is the same rule's other half. A layer that sets one key of
// a map must not delete the map's other keys — this is what keeps `pipeline.recall:
// {engine: none}` alive when a user layer only mentions `pipeline.verify`.
func TestMergeKeepsSiblingKeys(t *testing.T) {
	o := isolated(t)
	writeLayer(t, projectCfg(o.ProjectDir), "paths:\n  notes: written/\n")
	c := mustLoad(t, o)
	if c.Paths.Notes != "written/" {
		t.Fatalf("paths.notes = %q", c.Paths.Notes)
	}
	if c.Paths.Inbox == "" || c.Paths.Archive == "" {
		t.Fatalf("sibling paths were dropped: %+v", c.Paths)
	}
}

// TestLockedStageRefused is the T0 invariant's enforcement point. The error must name
// the file that set it: a chain has four layers and "something set write: host" is not
// a diagnosis a user can act on.
func TestLockedStageRefused(t *testing.T) {
	for _, stage := range LockedStages {
		t.Run(stage, func(t *testing.T) {
			o := isolated(t)
			path := projectCfg(o.ProjectDir)
			writeLayer(t, path, "pipeline:\n  "+stage+": {engine: host}\n")
			_, err := Load(o)
			if err == nil {
				t.Fatalf("pipeline.%s: engine host was accepted", stage)
			}
			assertBlames(t, err, path, stage)
		})
	}
}

// TestLockedStageMergedNotPerLayer: a packaged "none" plus a project "host" must fail.
// Validating layers in isolation would pass both — the packaged one says none, and the
// project one is only a fragment — so validation has to run on the merged result.
func TestLockedStageMergedNotPerLayer(t *testing.T) {
	o := isolated(t)
	writeLayer(t, homeCfg(o.HomeDir), "pipeline:\n  recall: {engine: none}\n")
	writeLayer(t, projectCfg(o.ProjectDir), "pipeline:\n  recall: {engine: api}\n")
	if _, err := Load(o); err == nil {
		t.Fatal("a project layer overriding a locked stage was accepted")
	}
}

// TestLockedStageAbsenceIsFine: absence is not a violation. Only an explicit non-none
// value is a claim, or every partial config would be rejected.
func TestLockedStageAbsenceIsFine(t *testing.T) {
	o := isolated(t)
	writeLayer(t, projectCfg(o.ProjectDir), "pipeline:\n  verify: {engine: host}\n")
	mustLoad(t, o)
}

// TestEnvSetButMissingIsError: the user named this file explicitly. Skipping it silently
// would run them on settings they believe they replaced.
func TestEnvSetButMissingIsError(t *testing.T) {
	o := isolated(t)
	o.EnvPath = filepath.Join(t.TempDir(), "does-not-exist.md")
	if _, err := Load(o); err == nil {
		t.Fatal("a missing $FORGE_CONFIG was ignored")
	}
}

// TestOptionalLayersMayBeMissing is the other side: with no files anywhere, the packaged
// layer alone must produce a working config. This is the tarball case.
func TestOptionalLayersMayBeMissing(t *testing.T) {
	c := mustLoad(t, isolated(t))
	if len(c.Layers) != 1 || c.Layers[0] != PackagedName {
		t.Fatalf("Layers = %v, want only the packaged layer", c.Layers)
	}
}

// TestThresholdOrdering enforces answer >= update >= neighbour. It deliberately checks
// ordering rather than values: moving the thresholds to paper over
// a recall scoring defect is a different matter from a user tuning their own vault, and
// only the latter is this validation's job.
func TestThresholdOrdering(t *testing.T) {
	o := isolated(t)
	writeLayer(t, projectCfg(o.ProjectDir),
		"recall:\n  answer_threshold: 0.40\n  update_threshold: 0.70\n")
	if _, err := Load(o); err == nil {
		t.Fatal("update_threshold above answer_threshold was accepted")
	}
	writeLayer(t, projectCfg(o.ProjectDir),
		"recall:\n  answer_threshold: 0.90\n  update_threshold: 0.60\n")
	mustLoad(t, o)
}

// TestFrontmatterVariants: the file is markdown, so it arrives with prose after the
// fence, occasionally CRLF from a Windows editor, and occasionally a BOM from one that
// means well. All three must parse to the same thing.
func TestFrontmatterVariants(t *testing.T) {
	body := "vault_path: /v\n"
	cases := map[string]string{
		"plain":      "---\n" + body + "---\n",
		"with prose": "---\n" + body + "---\n\n# Notes\n\nsome prose\n",
		"crlf":       "---\r\n" + "vault_path: /v\r\n" + "---\r\n",
		"bom":        "\ufeff---\n" + body + "---\n",
		"no fence":   body,
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			m, err := Parse([]byte(src))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if m["vault_path"] != "/v" {
				t.Fatalf("vault_path = %v, want /v", m["vault_path"])
			}
		})
	}
}

func TestInvalidYAMLReportsTheFile(t *testing.T) {
	o := isolated(t)
	path := projectCfg(o.ProjectDir)
	writeLayer(t, path, "paths:\n  notes: [unclosed\n")
	_, err := Load(o)
	if err == nil {
		t.Fatal("malformed YAML was accepted")
	}
	if !contains(err.Error(), path) {
		t.Fatalf("error %q does not name %q", err, path)
	}
}
