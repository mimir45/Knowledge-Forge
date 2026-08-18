package scrub

import (
	"fmt"
	"path/filepath"

	"knowledge-forge/pkg/vault"
)

// scrubOne redacts one note. A note with no frontmatter block at all (the pre-migration
// shape a few real notes are still in) is redacted body-only; anything else with
// unparseable frontmatter fails closed rather than guessing at its shape.
func scrubOne(srcDir, rel string, schema *vault.Schema) (data []byte, noFM bool, redactions int, err error) {
	n, err := vault.Load(filepath.Join(srcDir, filepath.FromSlash(rel)), rel)
	if err != nil {
		return nil, false, 0, err
	}
	if n.FMErr == vault.ErrNoFrontmatter {
		body, count := redactBytes(n.Body)
		return body, true, count, nil
	}
	if n.FMErr != nil {
		return nil, false, 0, fmt.Errorf("unparseable frontmatter: %w", n.FMErr)
	}
	return scrubFrontmatterNote(n, schema)
}

// scrubFrontmatterNote redacts frontmatter values and body text, re-renders, and — only
// for a note that was schema-valid before scrubbing — checks it still is. A note that
// was already invalid is scrubbed best-effort; scrub is not the tool that fixes it.
func scrubFrontmatterNote(n *vault.Note, schema *vault.Schema) ([]byte, bool, int, error) {
	wasValid := len(vault.Validate(n, schema)) == 0
	count := scrubFields(n.FM)
	body, bodyCount := redactBytes(n.Body)
	out, err := vault.RenderNote(n.FM, schema, body)
	if err != nil {
		return nil, false, 0, fmt.Errorf("re-render: %w", err)
	}
	if wasValid {
		if err := reValidate(out, n.Rel, schema); err != nil {
			return nil, false, 0, err
		}
	}
	return out, false, count + bodyCount, nil
}

// reValidate parses a freshly rendered note back and checks it against the schema — the
// only way to know redaction didn't break a value the schema constrains.
func reValidate(rendered []byte, rel string, schema *vault.Schema) error {
	yamlSrc, body, err := vault.SplitFrontmatter(rendered)
	if err != nil {
		return fmt.Errorf("scrubbed note lost its frontmatter: %w", err)
	}
	fm, err := vault.ParseFrontmatter(yamlSrc)
	if err != nil {
		return fmt.Errorf("scrubbed note has invalid yaml: %w", err)
	}
	n := &vault.Note{Rel: rel, FM: fm, Body: body}
	if issues := vault.Validate(n, schema); len(issues) > 0 {
		return fmt.Errorf("scrubbed note fails schema: %s", issues[0])
	}
	return nil
}
