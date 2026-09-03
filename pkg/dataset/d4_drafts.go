package dataset

import (
	"os"
	"path/filepath"
)

// draftsDir holds one quarantined draft (plus its gate error) per gate failure.
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
// retry consumes the pairing exactly once.
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
