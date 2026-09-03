package dataset

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// D3 is the human-correction dataset: (model note, your edited note) pairs.
const (
	D3Kind = "d3-human-edit"
	D3Path = ".forge/datasets/d3.jsonl"

	// D3Tag is the cfg.Dataset.Capture entry that gates this tier.
	// cmd/forge/capture.go checks it before every capture.
	D3Tag = "d3"

	D3Window  = 7 * 24 * time.Hour
	notesRoot = "notes/"

	// ForgeTrailer marks a commit forge made itself.
	ForgeTrailer = "Forge-Write"
)

// generatedOrigins are the schema.yaml `origin` values that mean forge produced the
// note.
var generatedOrigins = map[string]bool{"ask": true, "session-capture": true, "garden": true}

// Pair is one JSONL record.
type Pair struct {
	Kind        string    `json:"kind"`
	Note        string    `json:"note"`
	Topic       string    `json:"topic"`
	ContentHash string    `json:"content_hash"`
	Origin      string    `json:"origin"`
	Stack       []string  `json:"stack,omitempty"`
	EngineTrail []string  `json:"engine_trail,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
	EditedAt    time.Time `json:"edited_at"`
	GenCommit   string    `json:"gen_commit"`
	EditCommit  string    `json:"edit_commit"`
	Generated   string    `json:"generated"`
	Preferred   string    `json:"preferred"`
}

// Key identifies a pair for deduplication. A hook can fire twice on the same commit —
// `commit --amend`, a rebase, a re-run by hand — and the file is append-only.
func (p Pair) Key() string { return p.EditCommit + "\x00" + p.Note }

type Options struct {
	Commit string
	Window time.Duration
}

// Capture returns the pairs one commit contributes.
func Capture(g Git, opt Options) ([]Pair, error) {
	sha, edited, err := g.CommitMeta(opt.Commit)
	if err != nil {
		return nil, err
	}
	own, err := forgeAuthored(g, sha)
	if err != nil || own {
		return nil, err
	}
	paths, err := g.ModifiedFiles(sha, "MR")
	if err != nil {
		return nil, err
	}
	return pairsFor(g, sha, edited, paths, window(opt)), nil
}

func window(opt Options) time.Duration {
	if opt.Window > 0 {
		return opt.Window
	}
	return D3Window
}

func forgeAuthored(g Git, sha string) (bool, error) {
	tr, err := g.Trailers(sha)
	if err != nil {
		return false, err
	}
	return strings.Contains(tr, ForgeTrailer+":"), nil
}

func pairsFor(g Git, sha string, edited time.Time, paths []string, w time.Duration) []Pair {
	var out []Pair
	for _, rel := range paths {
		if !isNote(rel) {
			continue
		}
		if p, ok := pairFor(g, sha, edited, rel, w); ok {
			out = append(out, p)
		}
	}
	return out
}

// isNote restricts D3 to the contract notes under notes/. raw/ and sources/ are inputs.
func isNote(rel string) bool {
	return strings.HasPrefix(rel, notesRoot) && strings.HasSuffix(rel, ".md")
}

// pairFor decides whether one modified note is a human correction of a forge-written one.
// Every gate here is a reason a pair would be noise rather than signal.
func pairFor(g Git, sha string, edited time.Time, rel string, w time.Duration) (Pair, bool) {
	org, err := g.FirstAdded(sha, rel)
	if err != nil || org.SHA == "" || org.SHA == sha {
		return Pair{}, false // unknown history, or born in this very commit
	}
	if d := edited.Sub(org.When); d < 0 || d > w {
		return Pair{}, false // outside the window: a later revisit, not a correction
	}
	gen, err := g.Show(org.SHA, org.Path)
	if err != nil {
		return Pair{}, false
	}
	cur, err := g.Show(sha, rel)
	if err != nil || cur == gen {
		return Pair{}, false // a rename or a mode change, with the text untouched
	}
	return build(rel, org, sha, edited, gen, cur)
}

func build(rel string, org Origin, sha string, edited time.Time, gen, cur string) (Pair, bool) {
	n := parse(org.Path, gen)
	if n.FM == nil || !generatedOrigins[n.FM.Str("origin")] {
		return Pair{}, false // imported or unlabelled: forge did not write it
	}
	return Pair{
		Kind: D3Kind, Note: rel, Topic: n.Title(), ContentHash: hash(gen),
		Origin: n.FM.Str("origin"), Stack: n.FM.List("stack"),
		EngineTrail: n.FM.List("engine_trail"),
		GeneratedAt: org.When, EditedAt: edited,
		GenCommit: org.SHA, EditCommit: sha,
		Generated: gen, Preferred: cur,
	}, true
}

// parse reads a note out of a git blob rather than off disk; pkg/vault.Load is
// filesystem-bound and the generated side of a pair no longer exists in the working tree.
func parse(rel, text string) *vault.Note {
	n := &vault.Note{Rel: rel}
	yamlSrc, body, err := vault.SplitFrontmatter([]byte(text))
	if err != nil {
		return n
	}
	n.Body = body
	if fm, err := vault.ParseFrontmatter(yamlSrc); err == nil {
		n.FM = fm
	}
	return n
}

func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:12])
}
