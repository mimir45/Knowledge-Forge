# Backlog

Work that came up outside the current phase's scope. Per `CLAUDE.md`, things land here
instead of getting built. Nothing in this file is scheduled; each entry names the phase
that should own it.

---

## B-001 — The design docs have never been audited for internal coherence

**Owner: Phase 0 (Audit).** Status: **done, 2026-08-09.** The pass ran and found thirteen
contradictions the docs do not self-flag; they are `AUDIT.md` §8.1. Seven resolved without
escalation. The six that could not be — C-5, C-7, C-8, C-9, C-10, C-13 — were
escalated and decided by the user the same day; the decisions are **`AUDIT.md` §8.4
(D-1 … D-8)**, which is binding for later phases. No design doc was edited. The original
brief follows.

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
topology plus twelve deliberate defects (catalogue in `testdata/README.md`). Its purpose is
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

---

## B-005 — Three docs disagree on how many note types there are

**Owner: Phase 1 (decided there, recorded here). Status: resolved by decision, 2026-08-09.**

Not one of the thirteen contradictions `AUDIT.md` §8.1 caught — grep confirms §8.1/§8.4
never mention templates or `incident`. Three arities:

| Source | Says |
|---|---|
| DESIGN §6.1 `type` enum | **7** — `concept howto pattern pitfall decision api incident` |
| DESIGN §6.2 template list | **7** — same set |
| `CLAUDE-CODE-PROMPT.md` item 2 | **6** — omits `incident` |
| DESIGN §7 `notes/` tree | **5** — `concept howto pattern pitfall decision` |

**Decision: the enum is authoritative.** Phase 1 builds seven templates and seven
`notes/<type>/` subdirs. The alternative leaves `incident` schema-valid with no template
and no home — a note that validates but cannot be written. DESIGN §7's tree ends in `…`
and reads as illustrative; the prompt's list is a plausible transcription slip. Recorded
in `references/taxonomy.md` §4. No doc was edited, per the standing precedence rule.

---

## B-006 — The topology migration breaks path-qualified wikilinks; the spec has no step for it

**Owner: Phase 1. Status: in scope, added to `migrate_vault.py`.**

`CLAUDE-CODE-PROMPT.md` item 6 lists seven non-negotiables for the migration and a
link-rewrite pass is not among them. But `AUDIT.md` records that inbound links to
`issues/testcontainers-docker-socket` are **path-qualified** (`[[issues/…]]`), and the
migration deletes `issues/` as a directory. Every such link dangles the moment the note
moves.

"Never deletes or reorders body content; only adds frontmatter" constrains what happens
*inside* a note, not whether the note moves — so move + backfill is the correct reading,
and a rewrite pass is required to keep it non-destructive in effect as well as in letter.

Shape: build the complete `old path → new path` map before writing anything, then rewrite
only `[[dir/name]]` forms to the new directory. Bare `[[name]]` links need no rewrite —
Obsidian resolves them by name, and `pkg/vault`'s resolver must too (`AUDIT.md` §11).
The dry-run summary must count rewritten links separately from moved files.

---

## B-007 — Phase 4's librarian must stamp `Forge-Write: true` or it poisons the D3 dataset

**Owner: Phase 4 (Subagents). Status: open — the guard exists, the producer does not.**

`pkg/dataset` (Phase 1) refuses to harvest a commit whose message carries the trailer
`Forge-Write:` — `dataset.ForgeTrailer`, pinned by `TestSkipsForgeAuthoredCommits`. Today
that check is **inert**: every commit in the vault is a human's, so nothing is ever
skipped by it.

It stops being inert the moment `forge-librarian` starts committing notes it wrote. Such a
commit modifies a note whose `origin` is `ask`/`session-capture`/`garden`, within the
window, in a commit that is not the one that created it — every D3 gate passes, and the
pair enters `d3.jsonl` with model output on **both** sides, labelled as a human preference.
Nothing surfaces it: the file is append-only, local, and not read again until Phase 6b's
export. D.1 calls D3 the most valuable dataset; this is the one way to silently ruin it.

**Requirement for Phase 4:** every commit `forge-librarian` authors must end with
`Forge-Write: true`. Verify with a fixture test that a librarian-authored commit yields
zero pairs — the guard is already written, only the producer side is missing.

---

## B-008 — `tags` and `stack` recall channels have no IDF weighting

**Owner: Phase 2b (drift + the nine reports). Status: open — measured in Phase 2, not fixed.**

`forge recall`'s tag and stack channels score a term carried by every note exactly like a
term carried by three. In a vault that is mostly Spring notes, `spring` and `boot` are
close to zero information, but they fire both channels at 1.000 — half the blend's weight
— leaving only the 0.4 title channel to discriminate.

Measured on the real vault (`references/recall-spec.md` §3.1):

- *"Redis caching in Spring Boot"* → **0.740** for `spring-cli-and-maven-commands-for-spring-boot`,
  a note about CLI invocations. `tags 1.000`, `stack 1.000`, `title` only 0.476. That is
  UPDATE(extend) into the wrong note — the damaging direction, because extend writes.
- *"Kafka consumers with Testcontainers"* → **0.469** for the genuinely correct
  `testcontainers-docker-based-integration-testing`, i.e. CREATE when extend was right.

**Do not fix this by moving the thresholds.** Dropping `update_threshold` to 0.45 admits
the second case but also admits `docker-compose-init-container-pattern` at 0.429 for a
question about build caching, and does nothing about the 0.740 false positive. The
thresholds stay at DESIGN §5.3's 0.85 / 0.55 until the cause is addressed.

Shape: weight each tag/stack hit by `log(N / df(term))` over the vault's own note count,
computed during the frontmatter scan that already walks every note, so it costs no extra
pass. Cap the weight so a hapax tag cannot dominate. It changes every recall number, which
is why it belongs in 2b — that phase re-measures the reports against recall anyway, and
`pkg/recall`'s test suite pins the current values so the delta will be explicit.
