package vault

import (
	"fmt"
	"regexp"
	"sync"

	"gopkg.in/yaml.v3"

	"knowledge-forge/references"
)

// Field is one entry under `fields:` in references/schema.yaml. Every constraint the
// schema can express has a home here; a nil pointer means "not constrained".
type Field struct {
	Type           string            `yaml:"type"`
	Required       bool              `yaml:"required"`
	Immutable      bool              `yaml:"immutable"`
	Vocabulary     string            `yaml:"vocabulary"`
	Format         string            `yaml:"format"`
	Constraint     string            `yaml:"constraint"`
	Pattern        string            `yaml:"pattern"`
	ItemPattern    string            `yaml:"item_pattern"`
	Values         []string          `yaml:"values"`
	Min            *int              `yaml:"min"`
	Max            *int              `yaml:"max"`
	MinLength      *int              `yaml:"min_length"`
	MaxLength      *int              `yaml:"max_length"`
	MinItems       *int              `yaml:"min_items"`
	MaxItems       *int              `yaml:"max_items"`
	Default        any               `yaml:"default"`
	DefaultsByType map[string]int    `yaml:"defaults_by_type"`
	MinItemsByType map[string]int    `yaml:"min_items_by_type"`
	ItemFields     map[string]*Field `yaml:"item_fields"`
	Notes          string            `yaml:"notes"`

	rePattern *regexp.Regexp
	reItem    *regexp.Regexp
	allowed   map[string]bool
}

// Schema is the parsed note contract.
type Schema struct {
	Version  int                          `yaml:"version"`
	ForgeVer string                       `yaml:"forge_schema_for"`
	Fields   map[string]*Field            `yaml:"fields"`
	Aliases  map[string]map[string]string `yaml:"aliases"`
	KeyOrder []string                     `yaml:"key_order"`
}

var (
	loadOnce sync.Once
	loaded   *Schema
	loadErr  error
)

// LoadSchema parses the embedded contract. The result is cached: `forge validate --all`
// walks a hundred notes and re-parsing the schema per note would dominate the runtime.
func LoadSchema() (*Schema, error) {
	loadOnce.Do(func() { loaded, loadErr = ParseSchema(references.SchemaYAML) })
	return loaded, loadErr
}

// ParseSchema parses schema YAML from bytes. Exported so tests can feed a variant.
func ParseSchema(src []byte) (*Schema, error) {
	var s Schema
	if err := yaml.Unmarshal(src, &s); err != nil {
		return nil, fmt.Errorf("schema.yaml: %w", err)
	}
	for name, f := range s.Fields {
		if err := f.compile(); err != nil {
			return nil, fmt.Errorf("schema.yaml: field %q: %w", name, err)
		}
	}
	if len(s.KeyOrder) != len(s.Fields) {
		return nil, fmt.Errorf("schema.yaml: key_order has %d entries, fields has %d",
			len(s.KeyOrder), len(s.Fields))
	}
	return &s, nil
}

// compile precomputes the regexps and the allowed-value set once per field.
func (f *Field) compile() error {
	var err error
	if f.Pattern != "" {
		if f.rePattern, err = regexp.Compile(f.Pattern); err != nil {
			return err
		}
	}
	if f.ItemPattern != "" {
		if f.reItem, err = regexp.Compile(f.ItemPattern); err != nil {
			return err
		}
	}
	f.allowed = make(map[string]bool, len(f.Values))
	for _, v := range f.Values {
		f.allowed[v] = true
	}
	for _, sub := range f.ItemFields {
		if err := sub.compile(); err != nil {
			return err
		}
	}
	return nil
}

// Canonical returns the canonical form of a value for a field with an alias map, and
// whether it was an alias.
func (s *Schema) Canonical(field, value string) (string, bool) {
	if c, ok := s.Aliases[field][value]; ok {
		return c, true
	}
	return value, false
}

// FreshnessDefault returns the freshness_days default for a note type, or 0.
func (s *Schema) FreshnessDefault(noteType string) int {
	f, ok := s.Fields["freshness_days"]
	if !ok {
		return 0
	}
	return f.DefaultsByType[noteType]
}
