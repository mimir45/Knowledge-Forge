package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"knowledge-forge/pkg/config"
)

func runInit(o initOpts) int {
	delta, err := configDelta(o)
	if err != nil {
		return initFail(err)
	}
	cfgBytes, profBytes, err := render(o, delta)
	if err != nil {
		return initFail(err)
	}
	cfgPath, profPath, err := initTargets(o.vault)
	if err != nil {
		return initFail(err)
	}
	if o.dryRun {
		return printDryRun(cfgPath, cfgBytes, profPath, profBytes)
	}
	if err := writeInit(o, cfgPath, cfgBytes, profPath, profBytes); err != nil {
		return initFail(err)
	}
	return summarize(o, cfgPath, profPath)
}

func initFail(err error) int {
	fmt.Fprintf(os.Stderr, "forge init: %v\n", err)
	return 1
}

func initTargets(vault string) (cfgPath, profPath string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("no home directory: %w", err)
	}
	return filepath.Join(home, ".forge", "forge.config.md"),
		filepath.Join(vault, "profiles", "me.md"), nil
}

// configDelta composes the two preset overlays and then the five answers on top. Only
// the delta is written: the packaged layer stays underneath and keeps supplying
// everything the user did not decide, so upgrading the binary delivers new defaults
// instead of being shadowed by a full copy of the old ones.
func configDelta(o initOpts) (map[string]any, error) {
	d := map[string]any{}
	for _, name := range []string{o.enginePreset, o.stackPreset} {
		p, err := config.Preset(name)
		if err != nil {
			return nil, err
		}
		d = config.Merge(d, p)
	}
	return config.Merge(d, answers(o)), nil
}

func answers(o initOpts) map[string]any {
	m := map[string]any{
		"vault_path": o.vault,
		"trigger":    map[string]any{"mode": o.trigger},
		"write":      map[string]any{"language": o.noteLanguage},
	}
	// A stack preset already decided static.languages, and it decided it better than a
	// single --language answer can. Only fill the gap when no preset was chosen.
	if o.stackPreset == "" {
		if langs := indexableLanguages(o.language); len(langs) > 0 {
			m["static"] = map[string]any{"languages": langs}
		}
	}
	return m
}

// indexableLanguages maps the primary-language answer onto what pkg/codeindex can
// actually parse. Listing a language with no grammar would promise AST-level drift the
// binary cannot deliver, so an unknown answer yields nothing rather than itself.
func indexableLanguages(lang string) []string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "java", "kotlin":
		return []string{"java", "kotlin"}
	case "typescript", "javascript", "ts", "js":
		return []string{"typescript", "javascript"}
	}
	return nil
}

func render(o initOpts, delta map[string]any) (cfg, prof []byte, err error) {
	cfg, err = renderConfig(delta)
	if err != nil {
		return nil, nil, err
	}
	prof, err = renderProfile(o)
	return cfg, prof, err
}

// renderConfig wraps the delta as frontmatter-only markdown, the format DESIGN §10 fixes
// so the file is readable in Obsidian. yaml.Marshal sorts map keys, so re-running init
// with the same answers produces a byte-identical file.
func renderConfig(delta map[string]any) ([]byte, error) {
	body, err := yaml.Marshal(delta)
	if err != nil {
		return nil, err
	}
	return []byte("---\n" + configHeader + string(body) + "---\n" + configFooter), nil
}

const configHeader = `# Knowledge Forge — your settings. Written by ` + "`forge init`" + `; edit it freely.
#
# These keys override the packaged defaults compiled into the binary. Anything absent
# here keeps following those defaults, which is why this file is short: an upgrade
# should be able to improve settings you never decided on.
#
# A .forge.config.md at the root of a project overrides this file for that project, and
# $FORGE_CONFIG overrides everything for one run. ` + "`forge config --layers`" + ` shows which
# files are in play right now.

`

const configFooter = `
Run ` + "`forge config`" + ` to see the merged result, or ` + "`forge config --layers`" + ` to see which
files produced it.
`

func writeInit(o initOpts, cfgPath string, cfg []byte, profPath string, prof []byte) error {
	files := []struct {
		path string
		data []byte
	}{{cfgPath, cfg}, {profPath, prof}}
	// Both guards run before either write. Interleaving them means a run that stops on
	// the profile still leaves a new config behind, so the retry the error asks for
	// starts from a half-configured machine.
	for _, f := range files {
		if err := guardExisting(f.path, o.force); err != nil {
			return err
		}
	}
	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(f.path, f.data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// guardExisting is why --force exists. Both targets are files a user edits by hand after
// the first run; silently replacing an edited profile would destroy the one part of this
// system that is genuinely theirs.
func guardExisting(path string, force bool) error {
	if force {
		return nil
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists — pass --force to replace it", path)
	}
	return nil
}

func printDryRun(cfgPath string, cfg []byte, profPath string, prof []byte) int {
	fmt.Printf("would write %s:\n\n%s\n", cfgPath, cfg)
	fmt.Printf("would write %s:\n\n%s\n", profPath, prof)
	return 0
}

func summarize(o initOpts, cfgPath, profPath string) int {
	fmt.Printf("wrote %s\nwrote %s\n\n", cfgPath, profPath)
	fmt.Printf("vault      %s\nengine     %s\nstack      %s\ntrigger    %s\nprofile    %s, depth %d\n",
		o.vault, o.enginePreset, orNone(o.stackPreset), o.trigger, o.seniority, o.depth)
	fmt.Printf("\nnext: forge index --vault %q, then ask a question.\n", o.vault)
	return 0
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
