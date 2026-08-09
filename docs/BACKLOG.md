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

---

## B-009 — `pkg/gitsig` shells out to `git`; STACK specifies go-git

**Owner: Phase 6 (packaging). Status: deviation taken knowingly in 2b.**

STACK names `go-git` for history analysis. `pkg/gitsig` runs the `git` CLI instead: go-git's
log walk over these repos was slower than the subprocess and its rename detection is weaker,
and the CLI gives `--follow` and `--numstat` for free. The cost is a runtime dependency on
`git` being on `PATH`, which matters at packaging time — a goreleaser binary that assumes it
will fail on a machine that has none. `gitsig.withStderr` already turns `exit status 128`
into a message naming the repo, so the failure is legible; Phase 6 should decide whether it
is *acceptable* and say so in the README's requirements.

---

## B-010 — `AUDIT.md` §7 says `food-ordering-system` has no git history; it does

**Owner: whoever next re-measures the baseline. Status: correction recorded, doc unedited.**

§7 recorded the repo as having no `.git`. Measured 2026-08-09: it is a git repo at
`19290d78`, 7 commits, 270 files. Everything §7 derived from "no history" for that repo —
churn, ownership, co-change — was therefore reported as unavailable when it was merely
unmeasured. Per the standing rule the doc was not edited. The three repo pins as measured in
2b: `meter` `7c1c8bfb` (41 files), `leprecoin` `72990ab2` (183), `food` `19290d78` (270).
Trap worth keeping: `/Users/mimir45/Code/BE/UtilMeter` is a *different* repo (`967b7d08`) and
is not `meter`.

---

## B-011 — `reports/` and `moc/` are graph nodes but not contract notes

**Owner: Phase 3 (config) if the split ever needs to be tunable. Status: implemented in 2b.**

`forge check` writes ten files into the vault, and they are notes: they carry wikilinks, they
de-orphan what they link, and the graph must count them or `graph-health.md` describes a
vault that does not exist. They are *not* contract notes — no frontmatter, no schema, and
`coverage.md`/`staleness.md` dividing by them would understate every percentage.

Two populations, both correct, and the difference is exactly the non-contract files: the real
vault measures **94 graph notes vs 91 contract entries** (`index`, `log`, `codebase`).
`vault.IsContractNote` draws the line and `TestNotesAndEntriesDifferByExactlyTheNonContractNotes`
pins it. Recorded because the numbers look like a bug when read side by side.

Second, subtler half: a report is a graph node, so writing reports changes the graph the next
run measures. `orphans.md` and `graph-health.md` both moved for two runs before settling
(1628 → 1509 → 1437 B and 670 → 374 B). The fixed point is reached in **two to three runs**,
after which every output is byte-identical. `forge check` is convergent, not idempotent from
the first run, and `TestCheckIsIdempotentOnDisk` does not catch this because it runs without
`--repo` and so never writes the two files that move.

---

## B-012 — `code_refs` is in the schema and nothing writes it yet

**Owner: Phase 4 (subagents write notes). Status: field added in 2b, no producer.**

AUDIT NF-4 found the vault cites code as prose shorthand: 14 of 19 path-shaped refs resolved
to no file, and none carried a symbol. `references/schema.yaml` now defines an optional
`code_refs: list<string>` in the canonical `repo:path[:line][#symbol]` form — the only form
that names its repository instead of leaving `pkg/coderef` to guess.

`pkg/coderef` still recovers refs from inline code spans, so the 91 existing notes keep
working. But **every note in the vault today reaches drift through the recovery path**, which
is why the ambiguity in B-018 exists at all. Phase 4's `forge-librarian` should write
`code_refs` on every note it authors; until something does, the field is documentation.

---

## B-013 — the code-index cache has no format version

**Owner: Phase 6 (release). Status: open, currently harmless.**

`codeindex.Save`/`Load` persist `<vault>/.forge/code-index-<repo>.json` and `pkg/drift`
patches it forward from the cached commit. Nothing in the file records the extractor's
version. A future change to what counts as a symbol — the arrow-const fix in 2b was exactly
such a change — leaves every existing cache silently mixed: old entries under the old rules,
patched entries under the new. The symptom is a drift verdict that disagrees with a clean
rebuild, which is the hardest kind of bug to see.

Shape: stamp a `SchemaVersion` constant into the saved struct and treat a mismatch as a cache
miss. Cheap now, and it must land before the first released binary, because that is the point
at which caches start outliving the code that wrote them.

---

## B-014 — the code index parses TypeScript, not Kotlin

**Owner: recorded, not scheduled. Status: deliberate swap.**

`CLAUDE.md` says "start with Java + Kotlin only". 2b ships **Java + TypeScript/TSX**, because
the vault's actual code citations are Java (`food`, `meter`) and React/TypeScript
(`leprecoin`) — there is no Kotlin in any of the three repos, and a Kotlin grammar would have
been dead weight while `SignUpPage` went unparsed. Consequence to know: `coverage.md` lists
`kotlin` as one of its two uncovered stacks, and that gap is now partly a property of the
index rather than of the vault.

---

## B-015 — `CodeGroup.DependsOn` is declared and never populated

**Owner: whichever phase makes `moc/codebase.md` a dependency map. Status: open.**

ADDENDUM §B.5 asks the codebase map to show what depends on what. The struct field exists;
nothing fills it, because `codeindex.File` captures declarations only and no import edges. So
the MOC currently groups by directory and ranks by churn but draws no arrows.

Adding imports to the extractor is the real work; the grouping is the smaller problem. Note
also that "module = directory" is an honest limitation, not a placeholder: nothing in the
index knows about Maven modules or Go packages, and inventing a grouping the code does not
declare would file code under modules its authors never wrote.

---

## B-016 — the vault carries both `sources:` and `source:`

**Owner: a future vault migration. Status: read-both workaround shipped in 2b.**

The schema key is plural. Pre-migration notes wrote singular, and Phase 1's migration did not
rename it in notes it could not otherwise fix. `sourcesOf` reads both, which is why
`deadlinks.md` stopped reporting "0 of 0 URLs" over a fully cited vault. The workaround is
correct and should stay for old notes, but the split itself is unresolved: two keys meaning
one thing will eventually be written inconsistently by something that only knows one of them.

Measured while fixing it: **63 citations in the vault are first-party** — a vault-relative
path, which an HTTP checker cannot and should not judge. `DeadlinksInput.FirstParty` reports
them separately so the summary does not read as an uncited vault.

---

## B-017 — §B.5's 90-day window shows nothing on these repos

**Owner: Phase 3 (config chain owns the default). Status: measured, not decided.**

`moc/codebase.md`'s "undocumented and moving" section is empty at the default 90 days for all
three repos, because their recent churn is genuinely low — leprecoin's per-file ceiling over
the last 90 days is 3, and `minSymbolCommits` is 2. At `--days 365` the same section reports
**7 / 51 / 0** (meter / leprecoin / food). The report is working; the default window is simply
longer than these repos' activity.

This is a defaults question, not a code question, which is why it belongs to the phase that
owns the config chain. Worth stating plainly in the report itself either way: a section that
says "0" for a 90-day window reads as "nothing to do" when it means "nothing moved recently".

---

## B-018 — a bare symbol citation is credited to one arbitrary declaration

**Owner: Phase 4 (`code_refs` producers) — or a decision to refuse. Status: known asymmetry.**

`locate` resolves a symbol-only citation through drift's symbol table so `coverage` and
`drift` name the same file. Where a name is declared once that is exactly right. Where it is
declared many times, `Find` returns `hits[0]` with `ok=true` and the file is credited as
documented — while an ambiguous *path* citation resolves to `Ambiguous` with no `RepoPath`
and is dropped. Two shapes of the same uncertainty, answered two different ways.

Measured on the real vault: **4 of 51 resolved symbol citations are ambiguous** —
`BeanConfiguration` (4 declarations), `Builder` (**44**), `OrderItem` (2), `Product` (3). All
four land in `food`, whose undocumented-and-moving count is 0 at both 90 and 365 days, so none
of the false positives the fix removed from `meter` and `leprecoin` rest on an arbitrary pick.
The exposure is bounded and currently invisible in the output.

Kept rather than reverted, because agreeing with drift's arbitration is worth more than a
second, differently-guessed answer — but it is a decision, not a side effect. Two ways out,
both better than tuning: `Find` could expose the hit count so an ambiguous credit can be
reported as such, or B-012's `code_refs` could make the citation unambiguous at the source.
Bluntly: one note citing bare `` `Builder` `` is not a claim anything can adjudicate, and the
ordering fix made that answer *stable*, not *right*.

---

## B-019 — duplicate detection ships three deviations from DESIGN §8

**Owner: Phase 6 if the numbers ever go in a README. Status: deliberate, measured.**

`duplicates.md` reports at **0.40**, not §5.3's 0.85; shingles at **one word**, not three;
and compares **same-type, body-only**. Each was forced by measurement, not preference: over
the whole real vault the same-type similarity ceiling is **0.504** and the cross-type ceiling
**0.609**, so a 0.85 report is empty on this corpus and says nothing. Three-word shingles put
the fixture's planted near-duplicate pair (F7, `soft-delete` ↔ `soft-deletion`) below every
threshold worth reporting; at one word it scores **0.575** against a next-best of 0.196.

The real vault's top pairs are 0.48 `rag-provider` ↔ `rag-server-port`, 0.48
`continue-config-json` ↔ `dev-tools-continue-dev`, 0.42 `config-precedence` ↔
`continue-config-json` — 3 pairs over 1547 candidates. **No pair reaches 0.85.** The report
header states the deviation rather than hiding it, and `TestDuplicatesHeaderAdmitsTheSpecThresholdIsUnmet`
pins that. The trap to avoid later is quoting "3 near-duplicate pairs" as if it were a §5.3
number.

---

## B-020 — `sort.Slice` comparators need a tiebreak unique in their collection

**Owner: any phase adding a ranked report. Status: four instances found and fixed in 2b.**

`sort.Slice` is not stable, so a comparator that reports two distinct items equal hands the
order to Go's map iteration. That is not a cosmetic flaw here: it breaks the invariant that a
drift verdict is a pure function of (note refs, tree state).

Found by md5-ing consecutive runs on an unchanged tree. `drift.md`'s headline flipped between
**9 and 10 notes** because `nameMap.sort` ordered symbol hits by `(repo, path)` only, and one
Java file declares both `Order.Builder` and `OrderItem.Builder` — tied under the short name
`Builder`. Auditing every `sort.Slice` in the tree found three more: `staleEntries` (on
`verified`, a *date*, so ties are the common case — and the list truncates to 15, so the tie
decided *membership*), `sortUncovered` (no path behind the symbol name), and `groupByNote`
(no reason behind the ref). The other sixteen end on a key unique in their collection.

The rule for later phases: **the last term of a comparator must be unique in the collection
being sorted.** A file path, a slug, a URL, a note pair. A symbol name is not. Verify the way
this was verified — hash the whole output set across consecutive runs, not one file, since
stability on today's data is not a total order.
