package coderef

import (
	"sort"
	"strings"
)

// Repo is one code repository the vault cites. Name is what a canonical ref says;
// Files are repo-relative paths with forward slashes.
type Repo struct {
	Name  string
	Root  string
	Files []string
}

type hit struct{ repo, path string }

// Registry resolves shorthand citations to repo-relative paths.
//
// The matching rule is an ordered *subsequence* over path segments, not a suffix. That
// is not a refinement, it is the whole point: AUDIT NF-4's failing citations look like
// "common-domain/valueobject/Money.java", and the file is
// "common-domain/src/main/java/com/food/domain/valueobject/Money.java". The note drops
// the build-layout segments in the middle, so no suffix of the citation is a suffix of
// the path and suffix matching resolves it to nothing. Requiring only that the cited
// segments appear in order, anchored at the filename, is what turns the shorthand the
// vault actually contains into something drift can open.
type Registry struct {
	repos  []Repo
	byBase map[string][]hit // "Money.java" -> every file with that basename
	byName map[string]bool
}

// NewRegistry indexes files by basename. Basename is the only cheap anchor: it is the
// one segment every citation shape carries, and it cuts the subsequence check down to
// the handful of files that could possibly match.
func NewRegistry(repos []Repo) *Registry {
	rg := &Registry{repos: repos, byBase: map[string][]hit{}, byName: map[string]bool{}}
	for _, r := range repos {
		rg.byName[r.Name] = true
		for _, f := range r.Files {
			seg := Segments(f)
			if len(seg) == 0 {
				continue
			}
			base := seg[len(seg)-1]
			rg.byBase[base] = append(rg.byBase[base], hit{r.Name, f})
		}
	}
	return rg
}

// Resolve reports what the repositories say about one citation. A symbol-kind ref is
// never Unresolved here: it carries no path, so pkg/drift resolves it through the
// symbol table instead, and calling it broken at this layer would be wrong.
func (rg *Registry) Resolve(r Ref) Resolution {
	if r.Kind == KindSymbol {
		return Resolution{Ref: r, Status: Resolved}
	}
	hits := rg.lookup(r)
	switch {
	case len(hits) == 0:
		return Resolution{Ref: r, Status: Unresolved}
	case len(hits) == 1:
		r.Repo = hits[0].repo
		return Resolution{Ref: r, Status: Resolved, RepoPath: hits[0].path}
	}
	return Resolution{Ref: r, Status: Ambiguous, Ambiguity: qualify(hits)}
}

// lookup narrows by basename, then by repo, then keeps only subsequence matches, then
// keeps only the tightest ones. Each step can leave several candidates; only the last
// step's survivors are reported, and more than one of them is a genuine ambiguity.
func (rg *Registry) lookup(r Ref) []hit {
	seg := Segments(rg.stripRoot(&r))
	if len(seg) == 0 {
		return nil
	}
	cands := rg.filterRepo(rg.byBase[seg[len(seg)-1]], r.Repo)
	m := matching(cands, seg)
	if len(seg) == 1 {
		// A bare filename named no directory context, so there is nothing to break the
		// tie *with*. Preferring the shortest path here would invent a preference the
		// note never expressed; two modules defining Money.java is a real ambiguity and
		// the report should say so.
		return plain(m)
	}
	return tightest(m)
}

// stripRoot turns an absolute citation into a repo-relative one when it points inside a
// registered repository, and pins the repo while it is at it. Some notes paste the full
// path their editor showed them; that is a *more* precise citation than the shorthand,
// and treating it as unresolvable because it starts with a slash throws the precision
// away.
func (rg *Registry) stripRoot(r *Ref) string {
	for _, repo := range rg.repos {
		if repo.Root == "" {
			continue
		}
		if rest, ok := strings.CutPrefix(r.Path, strings.TrimSuffix(repo.Root, "/")+"/"); ok {
			r.Repo = repo.Name
			return rest
		}
	}
	return r.Path
}

func plain(in []scored) []hit {
	out := make([]hit, 0, len(in))
	for _, s := range in {
		out = append(out, s.hit)
	}
	return out
}

type scored struct {
	hit
	extra int // segments in the file path that the citation did not name
}

func matching(cands []hit, seg []string) []scored {
	var out []scored
	for _, c := range cands {
		fileSeg := Segments(c.path)
		if isSubsequence(seg, fileSeg) {
			out = append(out, scored{c, len(fileSeg) - len(seg)})
		}
	}
	return out
}

// tightest keeps the candidates that named the fewest unstated segments. Between
// "common-domain/.../domain/valueobject/Money.java" and
// "order-service/.../order/domain/valueobject/Money.java", a citation beginning
// "common-domain" only subsequence-matches the first, but where both match the shorter
// path is the one the note more plausibly meant.
func tightest(in []scored) []hit {
	best := -1
	for _, s := range in {
		if best < 0 || s.extra < best {
			best = s.extra
		}
	}
	var out []hit
	for _, s := range in {
		if s.extra == best {
			out = append(out, s.hit)
		}
	}
	return out
}

// isSubsequence reports whether every segment of want appears in have, in order. The
// last segment of want is the basename and is already known to match, so the walk only
// has to place the segments before it.
func isSubsequence(want, have []string) bool {
	i := 0
	for _, h := range have {
		if i < len(want) && strings.EqualFold(h, want[i]) {
			i++
		}
	}
	return i == len(want)
}

// filterRepo honours an explicitly named repository. A canonical ref that names a repo
// the registry does not know is not silently searched everywhere — that would turn a
// typo into a confident wrong answer.
func (rg *Registry) filterRepo(hits []hit, repo string) []hit {
	if repo == "" {
		return hits
	}
	var out []hit
	for _, h := range hits {
		if h.repo == repo {
			out = append(out, h)
		}
	}
	return out
}

func qualify(hits []hit) []string {
	out := make([]string, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.repo+":"+h.path)
	}
	sort.Strings(out)
	return out
}

// Known reports whether a repository name is registered, so a canonical ref pointing at
// a repo that is simply not on this machine can be skipped rather than called broken.
func (rg *Registry) Known(name string) bool { return rg.byName[name] }

// Repos returns the registered repositories in registration order.
func (rg *Registry) Repos() []Repo { return rg.repos }
