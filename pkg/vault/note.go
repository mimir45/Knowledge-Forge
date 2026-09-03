package vault

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Note is one markdown file on disk, parsed far enough to validate and index it.
type Note struct {
	Path    string       // absolute
	Rel     string       // vault-relative, forward slashes
	FM      *Frontmatter // nil when the file has no frontmatter block
	Body    []byte
	FMErr   error // ErrNoFrontmatter, or a YAML parse error
	ModTime int64
	Size    int64
}

// Load reads and parses one note. A missing or malformed frontmatter block is recorded
// in FMErr rather than returned: `forge validate` must report on such notes, not stop.
func Load(abs, rel string) (*Note, error) {
	src, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	n := &Note{Path: abs, Rel: filepath.ToSlash(rel), Size: int64(len(src))}
	if fi, err := os.Stat(abs); err == nil {
		n.ModTime = fi.ModTime().Unix()
	}
	yamlSrc, body, err := SplitFrontmatter(src)
	n.Body = body
	if err != nil {
		n.FMErr = err
		return n, nil
	}
	n.FM, n.FMErr = ParseFrontmatter(yamlSrc)
	return n, nil
}

// Title returns the frontmatter title, falling back to the first ATX heading and then
// to the filename. Used by the migration, which meets notes that have none of the three.
func (n *Note) Title() string {
	if n.FM != nil {
		if t := strings.Trim(n.FM.Str("title"), `"' `); t != "" {
			return t
		}
	}
	if h := firstHeading(n.Body); h != "" {
		return h
	}
	return strings.TrimSuffix(filepath.Base(n.Rel), ".md")
}

func firstHeading(body []byte) string {
	for _, line := range strings.Split(string(body), "\n") {
		if after, ok := strings.CutPrefix(strings.TrimSpace(line), "# "); ok {
			return strings.TrimSpace(after)
		}
	}
	return ""
}

// skipDirs are never walked. `.obsidian` and `.trash` are Obsidian's own state;
// `.forge` is our derived cache and must never be mistaken for vault content.
var skipDirs = map[string]bool{
	".git": true, ".obsidian": true, ".trash": true, ".forge": true,
	"node_modules": true, ".idea": true,
}

// Walk returns every markdown file under root that is a candidate note, vault-relative
// paths sorted by filepath.WalkDir's lexical order so results are reproducible.
func Walk(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".md") {
			rel, _ := filepath.Rel(root, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	return out, err
}

// excludedPrefixes are regions of the vault the note contract deliberately does not
// cover: symlinked ingest input and its compiled digests.
var excludedPrefixes = []string{
	"raw/", "sources/", "_archive/", "archive/", "reports/", "moc/weekly/",
}

// excludedNames are root-level files that live in the vault but are not notes.
var excludedNames = map[string]bool{
	"CLAUDE.md": true, "README.md": true, "lint-report.md": true,
}

// hubNames are hand-maintained root-level hubs.
var hubNames = map[string]bool{"index.md": true, "log.md": true}

// IsContentNote reports whether a vault-relative path is a node in the link graph.
func IsContentNote(rel string) bool {
	if excludedNames[rel] || strings.HasPrefix(filepath.Base(rel), "_index") {
		return false
	}
	for _, p := range excludedPrefixes {
		if strings.HasPrefix(rel, p) {
			return false
		}
	}
	return true
}

// IsContractNote reports whether a path is subject to the note contract.
func IsContractNote(rel string) bool {
	return IsContentNote(rel) && !hubNames[rel] && !strings.HasPrefix(rel, "moc/")
}
