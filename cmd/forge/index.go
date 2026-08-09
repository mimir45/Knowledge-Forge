package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"knowledge-forge/pkg/graph"
	"knowledge-forge/pkg/report"
	"knowledge-forge/pkg/store"
	"knowledge-forge/pkg/vault"
)

func cmdIndex(args []string) int {
	fs := flag.NewFlagSet("forge index", flag.ContinueOnError)
	vaultDir := fs.String("vault", ".", "vault root")
	out := fs.String("out", "_index.md", "output file, relative to the vault root")
	budget := fs.Int("max-bytes", 4096, "byte budget for the generated index")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "usage: forge index [--vault DIR] [--out FILE] [--max-bytes N]\n\n"+
			"Rebuilds the vault index from markdown. Idempotent within a day: the header\n"+
			"carries a date, not a timestamp, so a second run leaves the file untouched.\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	return runIndex(*vaultDir, *out, *budget, false)
}

func cmdReindex(args []string) int {
	fs := flag.NewFlagSet("forge reindex", flag.ContinueOnError)
	vaultDir := fs.String("vault", ".", "vault root")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "usage: forge reindex [--vault DIR]\n\n"+
			"Discards the derived SQLite cache and rebuilds it entirely from markdown.\n"+
			"Markdown is the only source of truth; nothing is lost by running this.\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	return runIndex(*vaultDir, "_index.md", 4096, true)
}

func runIndex(vaultDir, outName string, budget int, reset bool) int {
	root, err := filepath.Abs(vaultDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge index: %v\n", err)
		return 2
	}
	st, err := store.Open(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge index: cache: %v\n", err)
		return 1
	}
	defer st.Close()
	if reset {
		if err := st.Reset(); err != nil {
			fmt.Fprintf(os.Stderr, "forge reindex: %v\n", err)
			return 1
		}
	}
	return buildAndWrite(root, outName, budget, st)
}

func buildAndWrite(root, outName string, budget int, st *store.Store) int {
	notes, err := loadNotes(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge index: %v\n", err)
		return 1
	}
	g, ix := buildGraph(root, notes)
	if err := persist(st, notes, ix); err != nil {
		fmt.Fprintf(os.Stderr, "forge index: cache write: %v\n", err)
		return 1
	}
	md := report.RenderIndex(report.IndexInput{
		Entries: entries(notes, g), Now: time.Now(), MaxSize: budget,
	})
	return writeIfChanged(filepath.Join(root, outName), md, len(notes))
}

func loadNotes(root string) ([]*vault.Note, error) {
	rels, err := vault.Walk(root)
	if err != nil {
		return nil, err
	}
	var out []*vault.Note
	for _, rel := range rels {
		if !vault.IsContentNote(rel) {
			continue
		}
		if n, err := vault.Load(filepath.Join(root, rel), rel); err == nil {
			out = append(out, n)
		}
	}
	return out, nil
}

func buildGraph(root string, notes []*vault.Note) (*graph.Graph, *vault.Index) {
	rels := make([]string, 0, len(notes))
	for _, n := range notes {
		rels = append(rels, n.Rel)
	}
	ix := vault.NewIndex(rels)
	return graph.Build(nodesOf(ix, notes)), ix
}

// nodesOf is split out because forge check needs the node slice itself: graph.Components
// and Graph.Orphans both take it, and rebuilding it from the Graph is not possible — a
// Graph keeps counts, not edges.
func nodesOf(ix *vault.Index, notes []*vault.Note) []graph.Node {
	nodes := make([]graph.Node, 0, len(notes))
	for _, n := range notes {
		nodes = append(nodes, graph.Node{Rel: n.Rel, Outbound: resolveAll(ix, n)})
	}
	return nodes
}

func resolveAll(ix *vault.Index, n *vault.Note) []string {
	var out []string
	for _, t := range vault.Wikilinks(n.Body) {
		if target, ok := ix.Resolve(t); ok && target != n.Rel {
			out = append(out, target)
		}
	}
	return out
}

func entries(notes []*vault.Note, g *graph.Graph) []report.Entry {
	s, _ := vault.LoadSchema()
	out := make([]report.Entry, 0, len(notes))
	for _, n := range notes {
		// Hubs stay in the graph but out of the listing: they have no frontmatter to
		// render and counting them as "failing" would misreport the contract number.
		if !vault.IsContractNote(n.Rel) {
			continue
		}
		out = append(out, entryOf(n, g, s))
	}
	return out
}

func entryOf(n *vault.Note, g *graph.Graph, s *vault.Schema) report.Entry {
	e := report.Entry{Rel: n.Rel, Title: n.Title(), Slug: slugOf(n), Orphan: g.IsOrphan(n.Rel)}
	if n.FM == nil {
		return e
	}
	e.Type, e.Stack = n.FM.Str("type"), n.FM.List("stack")
	e.Updated, _ = time.Parse("2006-01-02", n.FM.Str("updated"))
	e.Verified, _ = time.Parse("2006-01-02", n.FM.Str("verified"))
	e.FreshnessDays = atoiOr(n.FM.Str("freshness_days"), s.FreshnessDefault(e.Type))
	e.Valid = len(vault.Validate(n, s)) == 0
	return e
}

func atoiOr(s string, fallback int) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil || n <= 0 {
		return fallback
	}
	return n
}

func persist(st *store.Store, notes []*vault.Note, ix *vault.Index) error {
	tx, err := st.DB.Begin()
	if err != nil {
		return err
	}
	live := make(map[string]bool, len(notes))
	for _, n := range notes {
		live[n.Rel] = true
		if err := store.Put(tx, rowOf(n, ix)); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := commit(tx); err != nil {
		return err
	}
	return st.Prune(live)
}

func commit(tx *sql.Tx) error { return tx.Commit() }

func rowOf(n *vault.Note, ix *vault.Index) store.Row {
	r := store.Row{Rel: n.Rel, Title: n.Title(), Slug: slugOf(n),
		MTime: n.ModTime, Size: n.Size, Links: linksOf(n, ix)}
	if n.FM == nil {
		return r
	}
	r.Type, r.Confidence = n.FM.Str("type"), n.FM.Str("confidence")
	r.Updated, r.Verified = n.FM.Str("updated"), n.FM.Str("verified")
	r.Stack, r.Tags = n.FM.List("stack"), n.FM.List("tags")
	r.FreshnessDays = atoiOr(n.FM.Str("freshness_days"), 0)
	return r
}

func linksOf(n *vault.Note, ix *vault.Index) []store.Link {
	var out []store.Link
	for _, t := range vault.Wikilinks(n.Body) {
		resolved, _ := ix.Resolve(t)
		out = append(out, store.Link{Target: t, Resolved: resolved})
	}
	return out
}

// writeIfChanged keeps the index idempotent on disk, not merely in content: rewriting an
// identical file would bump its mtime and invalidate every downstream mtime cache.
func writeIfChanged(path string, md []byte, n int) int {
	if old, err := os.ReadFile(path); err == nil && string(old) == string(md) {
		fmt.Printf("%s unchanged (%d notes, %d bytes)\n", filepath.Base(path), n, len(md))
		return 0
	}
	if err := os.WriteFile(path, md, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "forge index: %v\n", err)
		return 1
	}
	fmt.Printf("%s written (%d notes, %d bytes)\n", filepath.Base(path), n, len(md))
	return 0
}
