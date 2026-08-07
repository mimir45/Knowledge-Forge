# Backlog

Work that came up outside the current phase's scope. Per `CLAUDE.md`, things land here
instead of getting built. Nothing in this file is scheduled; each entry names the phase
that should own it.

---

## B-001 — The design docs have never been audited for internal coherence

**Owner: Phase 0 (Audit).** Status: open.

The five docs in `docs/` were written incrementally and only ever checked against each
other where they *self-flag* a conflict. Exactly three such flags exist and all three
are already recorded in `CLAUDE.md`:

1. STACK/ADR-001 supersedes ADDENDUM §B (which specified Python — "that was wrong").
2. STACK/ADR-002 supersedes B2B §8's Spring Boot assumption (now an open decision).
3. DESIGN's rev-2 note reinterprets every `scripts/*.py` reference as a `forge`
   subcommand.

**Nothing beyond those three has been verified.** No pass has been done for
contradictions the docs do *not* announce — conflicting phase numbering, a report named
in ADDENDUM §B.4 that no phase's deliverables produce, config keys in ADDENDUM §E that
no stage reads, latency budgets stated differently in two places, gate criteria that
disagree. Any of those would surface first as a Phase 2b implementation that cannot
satisfy two docs at once.

Phase 0's brief is already "establish the factual baseline before changing anything"
and its deliverable is `docs/AUDIT.md`. Add a section to it: **doc-vs-doc coherence**,
listing every contradiction found and which doc wins under the precedence rule in
`CLAUDE.md`. Resolve precedence there rather than editing the docs mid-flight.

---

## B-002 — Fixture vault exists at `testdata/vault/`; it is not `examples/vault/`

**Owner: Phase 6 must not confuse the two.** Status: done (the fixture), open (the
distinction is a trap).

`testdata/vault/` is a 13-note fixture reproducing the real vault's **pre-migration**
topology plus ten deliberate defects (catalogue in `testdata/README.md`). Its purpose is
to let Phase 1's topology migration — the one irreversible step in the plan — be
rehearsed without touching `/Users/mimir45/Documents/Base`. It carries **no `.git`** by
design — the harness copies it into a temp dir and `git init`s the copy, which is what
exercises the migration's "refuses to run on a dirty tree" precondition and drift's
`--since-commit`. Nesting a repo here would become a stray gitlink the moment this repo
is `git init`-ed. Never `git init` `testdata/vault/` in place.

`examples/vault/` is a **separate, still-unbuilt Phase 6 deliverable** (ROADMAP row 6):
a clean, exemplary vault shipped with the plugin. Do not point Phase 6 at `testdata/`,
and do not "clean up" the fixture's defects — they are the test surface.

---

## B-003 — Repo directory is still named `TIL`

**Owner: whenever, low cost. Status: open.**

The Go module was renamed to `knowledge-forge` on 2026-08-08, but the directory is still
`/Users/mimir45/TIL`, and the docs call the artifact `knowledge-forge/`. Purely cosmetic
now; it becomes mildly annoying once tooling, README paths, and the goreleaser config in
Phase 6 assume the artifact name. Renaming the directory is a user decision (it breaks
any shell aliases, IDE projects, and `.idea/` state pointing at the old path).

---

## B-004 — Module path has no VCS host prefix

**Owner: Phase 6 (Package & release). Status: deferred by decision.**

`go.mod` reads `module knowledge-forge` — a bare path, chosen deliberately on
2026-08-08 ("no need github for now"). This is legal and fine for a project distributed
as a goreleaser binary rather than `go get`. If the module ever needs to be importable
by others, it must become `github.com/<user>/knowledge-forge`, which rewrites every
import line. Cheapest to decide before `pkg/` has many files; effectively free today.
