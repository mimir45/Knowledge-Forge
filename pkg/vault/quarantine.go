package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteToInbox stamps confidence: low, appends an ## Open questions section naming each
// failed gate, and writes the note under _inbox/. This is the one path a gate failure may
// take: DESIGN and the PROMPT both forbid a silent drop, and forbid publishing a failing
// note anywhere confidence: low doesn't already say "don't trust this yet." Callers own
// any supersedes-style back-pointer (an UPDATE-mode retry) by setting it on n.FM before
// calling — WriteToInbox itself does not know CREATE from UPDATE.
func WriteToInbox(root string, n *Note, s *Schema, openQuestions []string) error {
	if n.FM == nil {
		return ErrNoFM
	}
	setScalar(n.FM, "confidence", "low")
	body := appendOpenQuestions(n.Body, openQuestions)
	out, err := RenderNote(n.FM, s, body)
	if err != nil {
		return err
	}
	path, err := inboxPath(root, n.FM.Str("slug"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	n.Path = path
	return os.WriteFile(path, out, 0o644)
}

// appendOpenQuestions adds the section if there are questions to record, and skips
// re-adding it if the draft already carries one (a retried WriteToInbox call).
func appendOpenQuestions(body []byte, qs []string) []byte {
	if len(qs) == 0 || strings.Contains(string(body), "## Open questions") {
		return body
	}
	var b strings.Builder
	b.Write(body)
	if len(body) > 0 && body[len(body)-1] != '\n' {
		b.WriteByte('\n')
	}
	b.WriteString("\n## Open questions\n\n")
	for _, q := range qs {
		fmt.Fprintf(&b, "- %s\n", q)
	}
	return []byte(b.String())
}

// inboxPath is _inbox/<slug>.md, required non-empty since a quarantined draft with no
// slug can't be found again by anyone reviewing _inbox/ later.
func inboxPath(root, slug string) (string, error) {
	if slug == "" {
		return "", fmt.Errorf("quarantine: note has no slug")
	}
	return filepath.Join(root, "_inbox", slug+".md"), nil
}
