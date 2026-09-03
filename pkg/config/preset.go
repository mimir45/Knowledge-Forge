package config

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	packaged "github.com/mimir45/Knowledge-Forge/config"
)

// EnginePresets and StackPresets are two independent axes.
var (
	EnginePresets = []string{"offline", "claude-only", "byo-api", "max"}
	StackPresets  = []string{"java-backend", "frontend", "devops", "minimal"}
)

// Preset returns one preset as a merge overlay.
func Preset(name string) (map[string]any, error) {
	if name == "" {
		return map[string]any{}, nil
	}
	src, err := packaged.Presets.ReadFile(path.Join("presets", name+".md"))
	if err != nil {
		return nil, fmt.Errorf("unknown preset %q (available: %s)",
			name, strings.Join(PresetNames(), ", "))
	}
	data, err := parse(src)
	if err != nil {
		return nil, fmt.Errorf("preset %s: %w", name, err)
	}
	return data, nil
}

// PresetNames lists every packaged preset, sorted, for error messages and for the
// wizard to offer.
func PresetNames() []string {
	entries, err := fs.ReadDir(packaged.Presets, "presets")
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, strings.TrimSuffix(e.Name(), ".md"))
	}
	sort.Strings(out)
	return out
}

// Merge is the exported merge rule, for `forge init` composing preset overlays before
// it writes them. Maps merge key by key; scalars and lists replace wholesale.
func Merge(low, high map[string]any) map[string]any { return merge(low, high) }

// Parse reads one frontmatter-only markdown config into a merge overlay.
func Parse(src []byte) (map[string]any, error) { return parse(src) }
