package main

import (
	"bytes"
	"strings"
	"text/template"
	"time"

	"github.com/mimir45/Knowledge-Forge/profiles"
)

// profileData is what profiles/me.template.md is rendered against. Five of these come
// from the wizard's five questions; the rest are derived here so the wizard does not
// grow a sixth. Every derived value is a starting point the template tells the user to
// edit — assume_known in particular is worth thirty seconds of their attention and
// cannot be guessed from a seniority label.
type profileData struct {
	PrimaryLanguage string
	Frameworks      []string
	Infra           []string
	Seniority       string
	DefaultDepth    int
	NoteLanguage    string
	ExplainStyle    string
	AssumeKnown     []string
	NeverAssume     []string
	CodeStyle       map[string]string
	Avoid           []string
	Generated       string
}

// defaultAvoid is the default "avoid" list for a generated profile. It is
// seniority-independent on purpose:
// nobody at any level wants the note to open with "in this article we will".
var defaultAvoid = []string{
	"marketing language", "history lessons", "'in this article we will'",
}

func renderProfile(o initOpts) ([]byte, error) {
	tpl, err := template.New("profile").
		Funcs(template.FuncMap{"yamlList": yamlList}).
		Parse(string(profiles.Template))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, profileFrom(o)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func profileFrom(o initOpts) profileData {
	known, never := assumptionsFor(o.seniority)
	return profileData{
		PrimaryLanguage: o.PrimaryLanguageOrAgnostic(),
		Frameworks:      splitList(o.frameworks),
		Infra:           splitList(o.infra),
		Seniority:       o.seniority,
		DefaultDepth:    o.depth,
		NoteLanguage:    o.noteLanguage,
		ExplainStyle:    o.explainStyle,
		AssumeKnown:     known,
		NeverAssume:     never,
		CodeStyle:       map[string]string{o.PrimaryLanguageOrAgnostic(): "idiomatic; replace this with your actual conventions"},
		Avoid:           defaultAvoid,
		Generated:       time.Now().Format("2006-01-02"),
	}
}

func (o initOpts) PrimaryLanguageOrAgnostic() string {
	if o.language == "" {
		return "en-agnostic"
	}
	return o.language
}

// assumptionsFor turns one seniority answer into the two lists that do the most work at
// synthesis time. The junior case is the interesting one: it is the only level with a
// non-empty never_assume, because the fields are asymmetric — assume_known removes
// explanation, never_assume adds it, and a junior needs the second more than the first.
func assumptionsFor(seniority string) (known, never []string) {
	switch seniority {
	case "junior":
		return []string{"git"},
			[]string{"concurrency", "caching", "database-indexing", "memory-model"}
	case "senior":
		return []string{"oop", "rest", "sql", "git", "basic-concurrency", "caching",
			"database-indexing", "http-semantics"}, nil
	}
	return []string{"oop", "rest", "sql", "git", "basic-concurrency"}, nil
}

func splitList(csv string) []string {
	var out []string
	for _, p := range strings.Split(csv, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// yamlList renders a flow sequence. An empty list must still print `[]` rather than
// nothing, or the rendered profile has a key with no value and stops parsing.
func yamlList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	out := make([]string, len(items))
	for i, v := range items {
		out[i] = yamlScalar(v)
	}
	return "[" + strings.Join(out, ", ") + "]"
}

// yamlScalar quotes anything that is not a bare identifier. defaultAvoid above
// is the forcing case: it contains spaces, and one entry is already single-quoted, so
// emitting these raw produces frontmatter that stops parsing at the first apostrophe.
// Simple values stay unquoted so the common lists still read as prose in Obsidian.
func yamlScalar(v string) string {
	if v != "" && strings.IndexFunc(v, func(r rune) bool {
		return !(r == '-' || r == '_' || r == '.' || r == '/' ||
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	}) < 0 {
		return v
	}
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v) + `"`
}
