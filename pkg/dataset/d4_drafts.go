package dataset

import (
	"os"
	"path/filepath"
)

// draftsDir holds one quarantined draft (plus its gate error) per gate failure, waiting
// on an explicit --previous-draft retry. There is deliberately no slug-based auto-pairing
// and no time window: the schema gate can itself fail on the slug, so a plausible fix can
// change the join key, and a stale draft from days ago must never silently pair with an
// unrelated retry of the same slug. The caller — the skill, which just watched its own
// write fail — holds the one honest join key: the exact file it was handed back.
const draftsDir = ".forge/drafts"

// SaveFailingDraft persists a quarantined draft and the gate error that failed it,
// returning the path forge gate prints for the caller to pass back as --previous-draft.
func SaveFailingDraft(vaultRoot, slug string, draft, gateError []byte) (string, error) {
	dir := filepath.Join(vaultRoot, draftsDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, slug+"-"+hash(string(draft))+".md")
	if err := os.WriteFile(path, draft, 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(path+".err", gateError, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// TakePreviousDraft reads back a draft and its paired gate error, then deletes both — a
// retry consumes the pairing exactly once, so a second --previous-draft pointing at the
// same path fails loudly (file not found) rather than double-emitting a D4 pair.
func TakePreviousDraft(path string) (draft, gateError []byte, err error) {
	draft, err = os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	gateError, err = os.ReadFile(path + ".err")
	if err != nil {
		return nil, nil, err
	}
	os.Remove(path)
	os.Remove(path + ".err")
	return draft, gateError, nil
}
