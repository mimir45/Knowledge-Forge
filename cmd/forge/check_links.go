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
		http, firstParty := sourcesOf(n.FM)
		for _, u := range http {
			byURL[u] = append(byURL[u], n.Rel)
		}
		d.firstParty += firstParty
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

// sourceURLs reads the url out of every citation, under both key names the vault holds.
// The schema's key is `sources`; the pre-migration notes wrote a singular `source`, and
// the migration did not rename it in the notes it could not otherwise fix. Reading only
// one of the two is how the first run of this report checked 0 URLs across 91 notes.
//
// Both value shapes are handled too — the schema's list of mappings and the bare scalar.
// Only http(s) is returned: the schema also admits a vault-relative path for a first-party
// source, and there is nothing for an HTTP checker to do with one.
func sourceURLs(fm *vault.Frontmatter) []string {
	http, _ := sourcesOf(fm)
	return http
}

// sourcesOf splits a note's citations into the ones an HTTP checker can judge and the
// first-party ones it cannot. Both counts are reported: dropping the second would let
// deadlinks.md print "0 of 0 URLs" over a vault whose every note is cited.
func sourcesOf(fm *vault.Frontmatter) (http []string, firstParty int) {
	for _, key := range []string{"sources", "source"} {
		n, ok := fm.Vals[key]
		if !ok {
			continue
		}
		for _, raw := range rawURLs(n) {
			if u := strings.Trim(raw, `"' `); isHTTP(u) {
				http = append(http, u)
			} else if u != "" {
				firstParty++
			}
		}
	}
	return http, firstParty
}

func rawURLs(n *yaml.Node) []string {
	if n.Kind == yaml.ScalarNode {
		return []string{n.Value}
	}
	out := make([]string, 0, len(n.Content))
	for _, item := range n.Content {
		out = append(out, mapValue(item, "url"))
	}
	return out
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

func isHTTP(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}
