package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	packaged "github.com/mimir45/Knowledge-Forge/config"
)

// PackagedName is what the embedded base layer is called in errors and in Layers. It is
// not a path on disk: the example is compiled into the binary so a stranger who runs
// `forge` from a tarball still gets a complete config (D-2 — the packaged layer is
// never edited by users, so there is no reason for it to be a file they can lose).
const PackagedName = "<packaged>/config/forge.config.example.md"

// EnvVar is the highest-precedence layer: an explicit, ad hoc, one-run override.
const EnvVar = "FORGE_CONFIG"

// Options are the chain's inputs. Every zero value resolves to the real environment, so
// Load(Options{}) is the production call and tests override one field at a time.
type Options struct {
	ProjectDir string // holds .forge.config.md; default "."
	HomeDir    string // holds .forge/forge.config.md; default os.UserHomeDir()
	EnvPath    string // default os.Getenv(EnvVar)
	Packaged   []byte // default the embedded example; tests only
}

type layer struct {
	name string
	data map[string]any
}

// Load resolves the four-layer config chain and returns the merged, validated config.
//
// Precedence, lowest first: packaged example, ~/.forge/forge.config.md (written by
// forge init), <project>/.forge.config.md, $FORGE_CONFIG. A missing optional layer is
// skipped; a missing $FORGE_CONFIG is an error, because the user named it explicitly
// and silently ignoring it would run them on settings they think they replaced.
func Load(opts Options) (*Config, error) {
	opts = opts.withDefaults()
	layers, err := readLayers(opts)
	if err != nil {
		return nil, err
	}
	merged := map[string]any{}
	names := make([]string, 0, len(layers))
	for _, l := range layers {
		merged = merge(merged, l.data)
		names = append(names, l.name)
	}
	c, err := decode(merged)
	if err != nil {
		return nil, err
	}
	c.Layers = names
	return c, validate(c, layers)
}

func (o Options) withDefaults() Options {
	if o.ProjectDir == "" {
		o.ProjectDir = "."
	}
	if o.HomeDir == "" {
		o.HomeDir, _ = os.UserHomeDir() // an unresolvable home skips that layer, not fails
	}
	if o.EnvPath == "" {
		o.EnvPath = os.Getenv(EnvVar)
	}
	if o.Packaged == nil {
		o.Packaged = packaged.Example
	}
	return o
}

// readLayers returns the layers that exist, lowest precedence first.
func readLayers(o Options) ([]layer, error) {
	base, err := parse(o.Packaged)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", PackagedName, err)
	}
	layers := []layer{{PackagedName, base}}
	for _, p := range optionalPaths(o) {
		l, ok, err := readFile(p)
		if err != nil {
			return nil, err
		}
		if ok {
			layers = append(layers, l)
		}
	}
	return appendEnvLayer(layers, o.EnvPath)
}

func optionalPaths(o Options) []string {
	var out []string
	if o.HomeDir != "" {
		out = append(out, filepath.Join(o.HomeDir, ".forge", "forge.config.md"))
	}
	return append(out, filepath.Join(o.ProjectDir, ".forge.config.md"))
}

// appendEnvLayer adds $FORGE_CONFIG, which unlike the others must exist if it is set.
func appendEnvLayer(layers []layer, path string) ([]layer, error) {
	if path == "" {
		return layers, nil
	}
	l, ok, err := readFile(path)
	switch {
	case err != nil:
		return nil, err
	case !ok:
		return nil, fmt.Errorf("%s=%s: no such file", EnvVar, path)
	}
	return append(layers, l), nil
}

func readFile(path string) (layer, bool, error) {
	src, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return layer{}, false, nil
	}
	if err != nil {
		return layer{}, false, err
	}
	data, err := parse(src)
	if err != nil {
		return layer{}, false, fmt.Errorf("%s: %w", path, err)
	}
	return layer{path, data}, true, nil
}

// parse reads one layer. The file is frontmatter-only markdown (DESIGN §10: "so it's
// readable in Obsidian too"), but a bare YAML document is accepted as well — nothing is
// gained by rejecting a file whose content is unambiguous.
func parse(src []byte) (map[string]any, error) {
	y := frontmatter(src)
	var out map[string]any
	if err := yaml.Unmarshal(y, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

// frontmatter returns the YAML between the leading `---` fence and the next one, or the
// whole input when there is no fence. CRLF is normalized first so a config edited on
// Windows and committed without .gitattributes still parses.
func frontmatter(src []byte) []byte {
	s := strings.ReplaceAll(string(src), "\r\n", "\n")
	s = strings.TrimPrefix(s, "\ufeff")
	if !strings.HasPrefix(s, "---\n") {
		return []byte(s)
	}
	rest := s[len("---\n"):]
	if i := strings.Index(rest, "\n---"); i >= 0 {
		return []byte(rest[:i+1])
	}
	return []byte(rest)
}

// decode turns the merged map into the typed Config by round-tripping it through YAML.
// Going back through the marshaller rather than reflecting over the map by hand means
// the struct tags stay the single description of the schema.
func decode(merged map[string]any) (*Config, error) {
	buf, err := yaml.Marshal(merged)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(buf, &c); err != nil {
		return nil, fmt.Errorf("merged config: %w", err)
	}
	c.VaultPath = expandHome(c.VaultPath)
	c.RepoPath = expandHome(c.RepoPath)
	return &c, nil
}

// expandHome resolves a leading ~/ because DESIGN §10's own example uses one.
func expandHome(p string) string {
	if !strings.HasPrefix(p, "~/") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	return filepath.Join(home, p[2:])
}
