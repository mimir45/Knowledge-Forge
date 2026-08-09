package main

import (
	"context"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"knowledge-forge/pkg/linkcheck"
	"knowledge-forge/pkg/report"
	"knowledge-forge/pkg/vault"
)

// links gathers every cited URL and the notes citing it, then checks each once.
//
// The map is keyed by URL rather than by note on purpose: a doc page cited by nine notes
// is one request, not nine, and the report is only actionable if a dead URL names all
// nine at once.
func (d *checkData) links() {
	byURL := map[string][]string{}
	for _, n := range d.notes {
		if n.FM == nil {
			continue
		}
		for _, u := range sourceURLs(n.FM) {
			byURL[u] = append(byURL[u], n.Rel)
		}
	}
	d.citations = d.statuses(byURL)
}

func (d *checkData) statuses(byURL map[string][]string) []report.Citation {
	urls := sortedStrings(byURL)
	dir := filepath.Join(d.root, ".forge")
	var got []linkcheck.Status
	if d.cfg.offline {
		got = cachedOnly(dir, urls)
	} else {
		got = linkcheck.New(dir, linkcheck.DefaultOptions()).Check(context.Background(), urls)
	}
	out := make([]report.Citation, 0, len(got))
	for _, s := range got {
		out = append(out, report.Citation{Status: s, Notes: byURL[s.URL]})
	}
	return out
}

// cachedOnly answers from the cache and reports everything else as unreachable, which is
// exactly what an offline run learned: nothing. The alternative — omitting the unknowns —
// would shrink the denominator and let deadlinks.md quietly claim a cleaner vault than it
// checked.
func cachedOnly(dir string, urls []string) []linkcheck.Status {
	c := linkcheck.LoadCache(dir)
	ttl := linkcheck.DefaultOptions().TTL
	out := make([]linkcheck.Status, 0, len(urls))
	for _, u := range urls {
		s, ok := c.Get(u, ttl)
		if !ok {
			s = linkcheck.Status{URL: u, Verdict: linkcheck.Unreachable, Detail: "offline, not cached"}
		}
		out = append(out, s)
	}
	return out
}

// sourceURLs reads the url out of every `source:` entry, in both shapes the vault holds:
// the schema's list of mappings, and the bare scalar the pre-migration notes used. Only
// http(s) is returned — the schema also admits a vault-relative path for a first-party
// source, and there is nothing for an HTTP checker to do with one.
func sourceURLs(fm *vault.Frontmatter) []string {
	n, ok := fm.Vals["source"]
	if !ok {
		return nil
	}
	if n.Kind == yaml.ScalarNode {
		return httpOnly([]string{strings.Trim(n.Value, `"' `)})
	}
	raw := make([]string, 0, len(n.Content))
	for _, item := range n.Content {
		raw = append(raw, mapValue(item, "url"))
	}
	return httpOnly(raw)
}

func mapValue(n *yaml.Node, key string) string {
	if n.Kind == yaml.ScalarNode {
		return n.Value
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1].Value
		}
	}
	return ""
}

func httpOnly(in []string) []string {
	var out []string
	for _, u := range in {
		if u = strings.Trim(u, `"' `); strings.HasPrefix(u, "http://") ||
			strings.HasPrefix(u, "https://") {
			out = append(out, u)
		}
	}
	return out
}
