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

**Owner: Phase 4 (Subagents). Status: done, 2026-08-12.** Both halves landed: the
fixture test (below) and the producer — `agents/forge-librarian.md`'s prompt instructs
`git commit --trailer "Forge-Write: true" -m "<summary>"` on the note-write commit and
any index-rebuild commit it makes. The packaging gap (`agents/` isn't loaded by Claude
Code today — see `CLAUDE.md`'s Status section) means this is verified spec, not yet a
live enforced guarantee; re-check once `agents/` is actually wired into a dispatchable
agent.

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

**Done this session:** `pkg/dataset/d3_forge_write_test.go` — stages `testdata/vault`
into a fresh temp-dir git repo (never in place, per B-002), commits an edit *with* the
trailer (asserts zero pairs), then an otherwise-identical edit *without* it (asserts
exactly one pair, pinning that step one's zero result is the guard firing, not a setup
bug). No Go write path changed; `forge capture`/`dataset.Capture` already had the guard.
Still open: nothing in this repo yet issues `git commit --trailer "Forge-Write: true"` —
that has to live in `agents/forge-librarian.md`'s prompt, not in Go.

---

## B-008 — `tags` and `stack` recall channels have no IDF weighting

**Owner: Phase 3. Status: open — the prescribed weighting shipped in 2b and neither case
moved past its threshold; the cause stated below is wrong, see the measurement at the end.**

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

### The weighting shipped in 2b. It did not fix either case, and the cause above is wrong.

`pkg/recall/score.go` now weights both channels by smoothed IDF — `log(1 + N/df)`, capped
at 3.5, document frequency counted in the pass that already walks every note. It is the
prescribed shape, it is tested, and `--explain` prints the per-term weights so the number
stays auditable. Re-measured against the real vault:

| case | before | after | verdict |
|---|---|---|---|
| "Redis caching in Spring Boot" → `spring-cli-and-maven-commands-for-spring-boot` | 0.740 | **0.740** | unchanged — still UPDATE(extend) into the wrong note |
| "Kafka consumers with Testcontainers" → `testcontainers-docker-based-integration-testing` | 0.469 | **0.501** | moved, still under 0.55 — still CREATE |

Document frequency over the 91 notes that carry frontmatter explains why:

| term | tag df | stack df |
|---|---|---|
| spring | 1 | 30 |
| boot | 0 | 30 |
| **redis** | **0** | **0** |
| **caching** | **0** | **0** |
| **kafka** | **0** | **0** |
| testcontainers | 0 | 7 |

Two mechanisms, neither of them flat weighting:

1. **The denominator is restricted to the vault's vocabulary.** `redis` and `caching` are
   carried by no note at all, so they are filtered out before any weight is computed. The
   only surviving tag term is `spring` and the only stack terms are `spring`/`boot` — all
   of which the wrong note carries. Both channels still read 1.000. IDF cannot discriminate
   on a term it never sees.
2. **`spring` is a hapax *tag*** — tag df 1 of 91. A weighted ratio over a single surviving
   term is 1.0 on any hit, whatever that term weighs. Capping helps a crowded denominator;
   it does nothing to a denominator of one.

**Candidate fix, deliberately not taken in 2b.** Admit question terms the vault carries
nowhere into the channel denominator: they are evidence the vault has no such note, and
`spring-cli-and-maven-commands-for-spring-boot` would fall to ≈0.373 for missing
`redis`/`caching`. An explicit `--stack` hint the vault has never seen must stay out — it
is a user filter, not evidence — so `newScope` would have to stop merging the two sources
into `stackTerms`, and `TestStackChannelIgnoresTermsNoNoteCarries` would need revisiting.

Rejected for now on blast radius, not difficulty. It re-creates the effect spec §2.5
rejected **from measurement** — an active-zero tags channel dragging down every tagged note
whose tags miss the question — and 31 of 91 notes are under-tagged after the Phase 1
migration. Verifying it honestly means re-deriving the whole §3.1 calibration table, not
re-running two queries. The tags-only variant lands case 1 at 0.515, which is 0.035 under
the threshold: too thin a margin on the single case that motivated the change to be
distinguishable from tuning, which this entry forbids by name.

**Still open. The thresholds still do not move.** The next attempt owns the §3.1
recalibration as part of its scope.

### Sized 2026-08-21. Read this before scheduling a session for it.

Not attempted this pass. Three facts that change what "own the §3.1 recalibration" costs,
none of them stated above:

**1. There is no harness that produces the §3.1 table, and there never was.** The nine
queries exist only as prose in `references/recall-spec.md:271-318`. Checked: `evals/run.sh`
is a *determinism* check — it builds `forge`, runs one query twice against the two-note
`evals/fixtures/vault/`, and diffs for byte-identity; it prints no score and never touches
the real vault. The `Makefile`'s `bench` target does not include `pkg/recall`. There is one
bench file in the repo and it is `pkg/vault/bench_test.go`. Nothing under `scripts/` runs
recall. So re-deriving the table today means nine hand-run `forge recall --explain`
invocations and manual transcription. **The first deliverable of that session is the
harness, not the fix.**

**2. The "before" column is stale too.** §3.1 says 91 notes; the vault has drifted since.
An "after" column compared against numbers measured on a different corpus proves nothing,
so the before column has to be re-derived on the same run — which is another argument for
building the harness first rather than reproducing nine numbers by hand twice.

**3. The candidate fix is two changes, not one, and the second is the actual open
decision.** Admitting df-0 terms requires relaxing `inVocab` (`pkg/recall/score.go:60`,
`if df[t] > 0`) **and** giving absent terms a non-zero weight — because `idf` returns 0 for
`df <= 0` (`score.go:94-96`), and a 0 weight adds nothing to `weighted`'s denominator
(`score.go:189-202`), which deactivates the channel anyway. Relaxing only `inVocab` moves
no number at all. **What that weight should be is unresolved**; this entry's "≈0.373"
estimate presumes some unstated choice, and the choice directly sets how far the false
positive falls. Note also that `score_test.go:202-204` (`idf(0, 91) != 0` — *"a term no
note carries must weigh nothing"*) has to be **inverted**, not re-pinned: it is the
assertion that defines today's behavior.

Blast radius, measured rather than guessed: **9 assertions across 6 tests** in
`pkg/recall/score_test.go`. `rank_test.go` and `result_test.go` are unaffected — their
fixture docs carry no `tags:`/`stack:`, so both channels are note-side inactive regardless
of the denominator rule. Splitting the `--stack` hint from question-derived terms
(`score.go:32-33`) changes the `scope` struct (`:20-26`) and `stackChannel` (`:167-176`)
but **not** `newScope`'s signature, and its only non-test caller is `pkg/recall/rank.go:21`.

One thing the session will want that does not exist yet: `--explain` prints IDF *weights*
but not document frequency, and `scope` discards `docFreq`'s counts after building the
weight map. Raw df is exactly the table this entry's second half needed, so plumbing it
through is additional work. While there, `references/recall-spec.md:372-385`'s sample
`--explain` output is stale — it predates the `idf ...` line the code has printed since 2b.

**Do not respond to any of this by moving the thresholds.**

---

## B-009 — `pkg/gitsig` shells out to `git`; STACK specifies go-git

**Owner: Phase 6 (packaging). Status: closed 2026-08-21 — Phase 6 made the call and wrote
it down; nobody had updated this entry.**

STACK names `go-git` for history analysis. `pkg/gitsig` runs the `git` CLI instead: go-git's
log walk over these repos was slower than the subprocess and its rename detection is weaker,
and the CLI gives `--follow` and `--numstat` for free. The cost is a runtime dependency on
`git` being on `PATH`, which matters at packaging time — a goreleaser binary that assumes it
will fail on a machine that has none. `gitsig.withStderr` already turns `exit status 128`
into a message naming the repo, so the failure is legible; Phase 6 should decide whether it
is *acceptable* and say so in the README's requirements.

**Closure note.** That is exactly what `README.md:25-31` does — a "Requirements" section
whose first bullet names `git` on `PATH`, attributes it to `pkg/gitsig`, calls it "a
deliberate, documented trade-off, not an oversight," and links back here. The deviation is
therefore *accepted*, not merely tolerated, and no code change follows: `pkg/gitsig` keeps
shelling out. Found still marked open during the 2026-08-21 defect sweep — the work landed
in Phase 6 and the status line was never updated, which is the only thing this note fixes.

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

**Owner: Phase 6 (release). Status: closed 2026-08-18 — mechanism predates this entry;
one gap in its contract widened, no new field added.**

Re-checking this at Phase 6 found the fix already in the tree: `pkg/codeindex/index.go`'s
`Index.Extractor int` field, stamped at every construction site (`build.go`'s `Build`,
`store.go`'s `Patch`), and `store.go`'s `Load` already rejects any stamp but the current
`Extractor` constant as a cache miss — precisely the "stamp a constant, treat a mismatch
as a cache miss" shape this entry asked for. Git history confirms `Extractor` was not
added in response to this entry: commit `1ba62dc` (2026-08-09 14:09) introduced it, and
commit `9d58a7a` (2026-08-09 14:39) — thirty minutes later — is what wrote this entry into
BACKLOG.md. The entry's own "nothing in the file records the extractor's version" line was
already stale the moment it was committed.

What was a real, if narrow, gap: `Extractor`'s doc comment scoped the bump rule to
"whenever `declKinds` or `kindOf` changes" — extraction-rule drift only. A future change to
`Symbol` or `File`'s serialized *shape* (a renamed or added field) wouldn't move that
number, and Go's `json.Unmarshal` silently ignores unknown/missing fields rather than
erroring, so an old cache would keep loading successfully under a struct it was never
written for — the same silently-mixed-cache symptom this entry warned about, just from a
different trigger. Fixed by widening the comment's contract to cover both triggers
explicitly; no new field, since a second version integer alongside `Extractor` would only
create two numbers with overlapping jobs.

Doesn't cover: nothing enforces the widened contract mechanically — a future shape change
that forgets to bump `Extractor` is a code-review miss, not a compile or test failure. If
that turns out to matter in practice, revisit with a struct-hash-based check instead of a
hand-bumped constant.

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

---

## B-021 — B2B is now a fully separate project, not a phase of this repo

**Owner: none — this is a scope decision, not implementation work. Status: decided,
2026-08-09.**

Every earlier doc — `docs/ROADMAP.md`'s old Sequencing notes, `docs/KNOWLEDGE-FORGE-B2B.md`
itself, and the "What it is" paragraph at the top of ROADMAP — treated B2B
(Slack/GitHub ingestion → org wiki → MCP server) as this repo's own Phase 7 gate: "OSS
v2.0 shipped, 30 days of real usage, ≥3 outside users reporting value." The user decided
otherwise: B2B becomes a **completely separate project**, not phase-gated inside this
one. Two sub-decisions, both taken via `AskUserQuestion` and both the recommended option:

1. **Timing:** update roadmap documentation now; do not create a second repo/project
   skeleton yet. That still waits for the same readiness condition above — informally,
   for the separate project, not as a gate this repo enforces.
2. **Where this repo's roadmap ends:** at **Phase 6b**. There is no Phase 7 row in
   `docs/ROADMAP.md`'s phase table anymore. Phase 7's actual content — `docs/RESULTS.md`,
   hit-rate/dedup/drift tracking, a launch post, a LoRA experiment gate — was **not**
   B2B content; it was OSS-lifecycle polish. It survives as ROADMAP's new, unnumbered
   "After 6b — run it and measure" section rather than being deleted.

**What was and was not edited.** `docs/ROADMAP.md` and `CLAUDE.md` (the two docs this
project already updates routinely as phases complete) were edited directly. The five
protected design docs were **not** touched, per the project's own "record, don't fix"
rule — they still say the old thing:

- `docs/KNOWLEDGE-FORGE-DESIGN.md:726` — `### Phase 7 — Polish on real usage (ongoing, 1
  month)`. Stale: read as OSS-lifecycle content now living in ROADMAP's "After 6b"
  section, not a numbered phase and not B2B.
- `docs/CLAUDE-CODE-PROMPT.md:563` — `## Phase 7 — After a month of real use`. Same.
- `docs/KNOWLEDGE-FORGE-B2B.md` — the whole file. Stale in framing only: it still reads
  as if it were a phase of this project. Read it as the separate project's spec, kept
  here for reference/history.

`.claude/agents/doc-auditor.md` was also updated (it is a workflow agent, not one of the
five protected docs): B2B no longer participates in the STACK → DESIGN → ADDENDUM
precedence order at all — a conflict against it is `OUT_OF_SCOPE_B2B`, not
`UNRESOLVED`.

Nothing about the *content* of `KNOWLEDGE-FORGE-B2B.md` changed — ADR-002 (Spring Boot
vs. an open decision) and everything else in it stands, unexamined, for whenever that
separate project actually starts.

---

## B-022 — `engine_trail`'s schema pattern doesn't cover four real pipeline stages

**Owner: Phase 4 (`code_refs`/`engine_trail` producers) or whoever next touches
`references/schema.yaml`. Status: done, 2026-08-10.** Fixed exactly as this entry's own
shape suggested: `references/schema.yaml`'s `item_pattern` now regenerates the alternation
from `cfg.Pipeline`'s nine keys minus `critique` — a fixed enumerated pattern, not
config-chain-driven validation, because `pkg/config` deliberately doesn't import
`pkg/vault` and note validity must not depend on which config happens to be loaded.
`pkg/engine/trail.go`'s `stampable` map was **not** touched — confirmed by the existing
`pkg/engine/trail_entry_test.go` (`TestTrailEntryUnstampedStages`, predating this fix) that
excluding `intake`/`plan`/`synthesize`/`link` from `stampable` was already a deliberate 3b
decision, not this gap; the schema now merely *accepts* those four stage names for
whichever future producer records them.

`references/schema.yaml`'s `engine_trail` field restricts each entry to
`^(recall|research|write|verify|critique|index)=(none|host|api|advisor)$`. The packaged
`config/forge.config.example.md` pipeline names nine stages, four of which the pattern
has no case for: `intake`, `plan`, `synthesize`, `link`. A note whose trail records what
actually ran one of those stages fails schema validation on a field that exists to
document exactly that.

`critique` is in the pattern and is not a `cfg.Pipeline` key at all — `pkg/engine`
deliberately excludes it from `stampable` (an earlier 3b decision, not this gap) because
the advisor tier's critique pass has no engine of its own to record. So the pattern is
both under- and over-inclusive relative to the real pipeline. Shape: regenerate the
alternation from `cfg.Pipeline`'s nine keys minus `critique`, or drop the enumerated list
and validate stage names against the config chain instead of a fixed regex.

---

## B-023 — code's `on_exhausted` value is `stop`; every doc still says `fail`

**Owner: whoever decides whether `stop` and `degrade` should diverge. Status: half done,
2026-08-21 — the four doc lines now say `stop`; the behavior question below is untouched
and still open. Do not close this entry on the doc edit alone.**

`pkg/config/validate.go` accepts `on_exhausted: queue | degrade | stop`, and that is what
`config/forge.config.example.md`'s comment and every preset's comment say too — `stop` is
the code's third value and it is consistent everywhere in `pkg/` and `config/`. Every doc
mention (`ADDENDUM.md:117`, `:485`, `:671`; `CLAUDE-CODE-PROMPT.md:339`) instead says
`degrade | queue | fail`. AUDIT §8.4 D-5 settled the *default* (`queue`, resolving C-13)
but did not touch this third value's name, so the docs' `fail` was carried forward
unexamined into 3b's implementation, which independently chose `stop`.

Checked this session by grepping every `OnExhausted` read site
(`cmd/forge/check_collect.go`, `cmd/forge/engine_run.go`): only `"queue"` is ever branched
on, at `engine_run.go:77`, to stamp `pending_advisor: true`. `degrade` and `stop` are
accepted by the validator and displayed verbatim in `cost.md`'s summary line, but nothing
in `pkg/engine` or `cmd/forge` reads either value — `engine.Resolve` already degrades a
metered stage to `"none"` regardless of `on_exhausted`, so the two configured values
produce byte-identical behavior today. So this is not just a naming mismatch: `stop` does
not halt anything, and `degrade` is not a distinct code path from the default silent
fallthrough. Left for whoever next touches those two files to reconcile the wording *and*
decide whether `stop`/`degrade` should diverge in behavior before renaming either one.

### Doc half resolved, 2026-08-21. Behavior half deliberately not.

All four doc sites now read `degrade | queue | stop`, matching `pkg/config/validate.go:89`
and `config/forge.config.example.md:82`: `ADDENDUM.md:117`, `:485`, `:671` and
`CLAUDE-CODE-PROMPT.md:339`. The `docs/tr/` mirror needed nothing — it never carried the
enum line, only the resolved `queue` value.

Re-confirmed while editing, so the next owner does not have to re-grep: `pkg/engine`
contains **zero** reads of `OnExhausted`; the degrade at `pkg/engine/select.go:30` is
unconditional. The three read sites are `cmd/forge/engine_run.go:77` and
`cmd/forge/check_drain.go:22` (both branch only on `"queue"`) and
`cmd/forge/check_collect.go:181` (passthrough into `cost.md`). No test exercises `degrade`
or `stop`.

**Why the behavior question was not settled here.** Making `stop` actually halt is a
behavior change to a budget-exhaustion path, and this was a documentation-sync pass; a
change that can start failing commands belongs in a session that owns it and can write
tests for it. Three options for that session, none chosen: give `stop` a real non-zero-exit
path and leave `degrade` as today's silent fallthrough; collapse to `queue | degrade` and
drop `stop` (backward-incompatible for any config that already sets it); or keep all three
and document explicitly that `stop` and `degrade` are synonyms today. The doc now names the
value the code accepts, which was the strictly-wrong part; the value's *meaning* is still
unimplemented.

---

## B-024 — `D2Tag` is spelled `"d2_advisor"`; the packaged config says `"d2"`

**Owner: whoever next touches `pkg/dataset/d2.go` or `config/forge.config.example.md`.
Status: closed 2026-08-21 — `D2Tag` renamed to `"d2"`, with a regression test that pins
the packaged config against the code. Originally found while building Phase 4's D4
dataset, out of that task's owned scope (only B-007 and B-022 were).**

`pkg/dataset/d2.go:17` defines `D2Tag = "d2_advisor"`, and `Enabled(capture []string)`
does an exact-string match against it. `config/forge.config.example.md:170` ships
`dataset.capture: [d1, d2, d3, d4, d5]` — the packaged config says `"d2"`, the code checks
for `"d2_advisor"`. They never match, so `dataset.Enabled` returns `false` and
`captureD2` (`cmd/forge/engine_run.go`) never writes a `d2.jsonl` line under the packaged
config, silently, today — confirmed by reading both files directly, not by running
anything (D2 capture requires a live advisor call this session didn't make).
(This entry originally cited `:169`; the line is `:170`.)

Two equally defensible fixes, not chosen here: rename `D2Tag` to `"d2"` to match the
config, or change the packaged config's list entry to `"d2_advisor"` to match the code.
Either is a one-line change; the only reason to pick one is whichever string is easier
for `forge init`'s wizard prompt to explain to a user. Don't touch `D2Kind`/`D2Path` — this
is a `capture:`-list membership bug, not a dataset-identity one.

`pkg/dataset/d4.go`'s `D4Tag = "d4"` was deliberately spelled to match the packaged
config exactly, so it does not repeat this mismatch — D4 fires under
`dataset.capture: [d1, d2, d3, d4, d5]` today where D2 silently does not. See that file's
doc comment.

### Closure note, 2026-08-21

Took the first fix: **`D2Tag` is now `"d2"`**, matching `d1`/`d3`/`d4`/`d5` and the
packaged list. `d4.go`'s doc comment recorded the asymmetry as "intentional, not a bug to
fix here," which is a marker for resolving in this direction; both it and
`cmd/forge/check_drain.go:79`'s "inert under the shipped config" comment were rewritten.
`D2Kind`/`D2Path` untouched, per the entry's own instruction.

**The reason this shipped green is that no test asserted config and code agree** — the tag
tests used hand-written lists, and `pkg/config`'s tests only checked that the packaged
layer parses. New guard: `TestPackagedCaptureListGates`
(`pkg/dataset/capture_gate_test.go`) loads the packaged base layer with no optional layers
and asserts both `Enabled` and `D4Enabled` return true against it. Verified to actually
bite — restoring `"d2_advisor"` fails it with `packaged dataset.capture [d1 d2 d3 d4 d5]
does not enable D2`. `pkg/dataset/d2_test.go` also gained a negative case pinning that the
old spelling no longer works, so the mismatch cannot return by a partial revert.

It lives in `pkg/dataset`, not `pkg/config`, because `pkg/config` cannot import
`pkg/dataset`: `dataset → vault → config` is a real edge (`pkg/vault/validate.go`), so the
reverse is a cycle.

**Two limits on what this closes, stated so nobody over-reads it:**

1. **Not verified end to end.** Both `captureD2` call sites (`cmd/forge/engine_run.go:117`,
   `cmd/forge/check_drain.go:95`) sit behind a live, metered advisor call — the second only
   runs after `buildEngine(cfg,"advisor").Run(req)` and `st.Spend(...)` both succeed. No
   advisor call was made, so no `d2.jsonl` line was observed. The claim is "the packaged
   config now matches the code, and a test pins it," not "D2 capture was seen working."
2. **Six presets still ship no `dataset.capture` key at all** and inherit the embedded
   `Example` layer, so they are fixed by this too; `config/presets/minimal.md:30` ships
   `[]` on purpose and stays inert, correctly.

A third thing surfaced while scoping this and is filed separately as **B-030**: three of
the five entries in that `capture:` list are read by no code at all.

---

## B-025 — `forge cache-source`'s `PostToolUse`/WebFetch `tool_response` JSON shape is unconfirmed

**Owner: whoever next touches `cmd/forge/cache_source.go`. Status: BLOCKED on external
confirmation, not open work — re-triaged 2026-08-21. Originally found while building
Phase 5's five hook subcommands.**

**Do not re-attempt the WebFetch.** Three tries against two official doc pages already
failed (below); a fourth is not new evidence. The trigger that unblocks this is
*observational*: a live `PostToolUse` hook firing on a real `WebFetch` call, whose payload
can be captured and read. Until someone has that payload in hand there is no work to do —
`cacheBody` already handles both shapes and both branches are tested
(`cache_source_test.go:12,34,49,59`), so the current code is the correct response to not
knowing, not a placeholder.

Three separate `WebFetch` calls this phase, against both
`https://code.claude.com/docs/en/hooks` and `https://code.claude.com/docs/en/tools-
reference`, failed to surface a literal field-by-field schema for `PostToolUse`'s
`tool_response` when `tool_name` is `"WebFetch"`. `tool_input.url`/`tool_input.prompt` are
known with confidence — they're WebFetch's own published tool parameters — but whether
`tool_response` is a plain string, or an object with a `content`/`result`/`text` key, was
never confirmed from official docs (checked 2026-08-13).

`cacheFetch`/`cacheBody` (`cmd/forge/cache_source.go`) deliberately does not guess a
field name: it unmarshals `tool_response` as a JSON string and uses it verbatim if that
succeeds, otherwise caches the raw JSON bytes Claude Code sent, unmodified. This means an
object-shaped response is cached as its full JSON text (e.g. `{"content":"...","ok":
true}`) rather than as just the extracted content — correct today only in the sense that
it never silently drops data or asserts a false schema. If a future session confirms the
real shape (e.g. by inspecting a live `.claude/settings.json` hook firing against a real
WebFetch call), `cacheBody` should be updated to extract the actual text field instead of
caching the wrapper JSON.

---

## B-026 — a citation to a fully deleted file can never verdict BROKEN

**Owner: whoever next touches `pkg/coderef/resolve.go` or `pkg/drift/check.go`. Status:
done for a full sweep with a verified date, 2026-08-16 — see the closure note below for
the five conditions it does not cover. Found while smoke-testing Phase 5's git-anchored
drift hooks against a real citation-and-deletion scenario; pre-existing in `pkg/drift`
(Phase 2b), out of Phase 5's scope to fix.**

**Closure note.** Fixed by generalizing `absentSymbol`'s existing shallow/deep split
(`pkg/drift/check.go`) from symbols to paths: `checkRef`'s `Unresolved` case now calls
`unresolvedPath`, which on a full deep sweep with a verified date calls the new
`Source.ResolveAt(ref, asOf)` — `GitSource.ResolveAt` builds a `coderef.Registry` at the
note's verified-era revision via `coderef.ScanRepo` (memoised per `asOf` in
`GitSource.registryAt` — the underlying `git ls-tree` cost is one per (repo, date) pair,
not per citation) and, if the citation resolved there and is `Unresolved` at HEAD,
verdicts `Broken`. Five things this does *not* close, so a reader does not assume more
than what shipped:

1. **Full sweep only.** The fix fires only when `opts.Deep` is true *and* `changed ==
   nil` (a true full sweep — `forge check`'s weekly run, or `forge drift --deep` with no
   `--since-commit`) *and* the note carries a `verified:` date. The hook path
   (`forge drift`'s default, `opts.Deep == false`) gains no immediacy — a same-commit
   deletion is only caught on the next full sweep. That gap is **B-028**, filed rather
   than built.
2. **Deletion and rename verdict identically.** The fallback cannot tell a file that was
   deleted from one that was renamed or moved: both resolve at the verified date and
   `Unresolved` at HEAD, and `Broken` is correct for both, because the citation as written
   is unopenable at HEAD either way — repairing a stale citation and retiring a dead claim
   are different fixes for the same verdict, but that distinction is for the human reading
   `drift.md`, not something the verdict can carry.
3. **A basename collision still hides the deletion.** A deleted file whose basename
   matches something else at HEAD resolves `Ambiguous`, not `Unresolved`, and `checkRef`'s
   `Ambiguous` branch already `skip`s it, unchanged by this fix — not newly caught.
4. **Cost bound, corrected.** This entry's original text called the trigger condition
   rare. It is not: `pkg/coderef/resolve.go`'s own comment names the underlying case "NF-4's
   14-of-19", i.e. most path citations in the real vault fail to resolve at HEAD as
   written. What actually bounds the cost is that `registryAt` is memoised per distinct
   `(repo, verified-date)` pair across the whole vault, not once per citation — one extra
   `git ls-tree` per pair, not per citation.
5. **No shipped invocation applies the `Broken` verdict.** `forge check`'s full sweep
   (`cmd/forge/check_codebase.go`) passes `opts.Deep: true`, satisfying the fallback's
   gate, but only renders `drift.md` — it never calls `drift.Apply`. `hooks/code-
   post-commit` does call `--apply`, but not `--deep`, so `changed != nil` closes the
   gate before `unresolvedPath` ever runs. The only combination that both triggers the
   fallback and demotes a note today is a hand-typed `forge drift --deep --apply` with no
   `--since-commit`. So this fix makes `drift.md` accurate on a full sweep; it does not
   yet demote a note through any automated path. Distinct from B-028 below (the hook
   path's lack of same-commit immediacy) — this is the deep-sweep path having no writer
   at all.

`cmd/forge/drift.go`'s `registryOf` builds the `coderef.Registry` from
`coderef.ScanRepo(r.Name, r.Root, "HEAD")` — the literal string `"HEAD"`, i.e. the
*current* tree, every run. `checkRef` (`pkg/drift/drift.go`) resolves every path-kind
citation (canonical `repo:path[:line][#symbol]` refs included, not just body-inline
shorthand) through that registry *before* it ever reaches `checkPath`'s file-existence
ladder. A file that has been fully deleted is, by construction, absent from a registry
built off current HEAD — `Registry.Resolve` returns `Unresolved` (`byBase` has no entry
for that basename), and `checkRef` reports `Skipped: "no registered repository contains
this path"` rather than reaching `checkPath`'s `Broken: "file no longer exists at HEAD"`
branch, which is dead code for this exact case. `Skipped` does not demote
(`Finding.Demoting()` is `Broken`-only), so **a note whose cited file was deleted outright
keeps its confidence forever** — it only gets caught if some other declaration happens to
share the same basename in the registry.

Confirmed empirically, not just read: a scratch note citing `repo:App.java` via
`code_refs` kept `confidence: high` and `drift_checked_at` advancing normally across a
commit that `git rm`'d `App.java` entirely — `forge drift --deep --json` on that state
returns `"verdict": "skipped", "reason": "no registered repository contains this path"`,
never `"broken"`. The addendum's ladder ("file gone, then symbol gone, then line moved,
then body changed") is written as if `checkPath` always runs; it silently doesn't for
this input.

This is orthogonal to `checkSymbol`'s and `absentSymbol`'s handling of a *symbol* going
missing while its file survives (or a bare symbol-only citation, which uses
`src.Find`/`--deep`'s verified-era lookup instead of the registry and correctly reports
`Broken`) — those paths are unaffected and already covered by
`TestCheckLadder`. The gap is specifically: path-kind citation + the entire file gone
from the current tree. A fix likely means giving `checkRef` a fallback when
`Registry.Resolve` returns `Unresolved` for a *canonical* (not body-inline) ref: probe
`src.At(ref.Repo, ref.Path, head)` directly using the ref's own literal repo/path before
giving up, since a canonical ref already claims to know the exact path and doesn't need
the registry's basename fuzzy-matching to find it.

**This sketch is not what shipped.** Probing `src.At` with the ref's literal path at HEAD
cannot distinguish "never existed" (a typo'd citation) from "was deleted" (a real drift
finding) — both return `!ok`. The closure note above resolves against the *verified-era*
tree instead, mirroring `absentSymbol`'s existing shallow/deep pattern for symbols, so
only a citation proven to have once resolved verdicts `Broken`.

---

## B-027 — `.forge/code-index-<repo>.json` is plural-per-repo; ADDENDUM/DESIGN say the singular `.forge/code-index.json`

**Owner: whoever next edits ADDENDUM §B.6 or DESIGN §15. Status: half done, 2026-08-21 —
the code side and one wrong agent instruction are fixed; the design docs still say the
singular name, deliberately. Originally found during Phase 5b's explore pass while
confirming `forge logback`'s requirement 3 ("keep the code index fresh") was already
satisfied by existing machinery; the naming predates Phase 5b, live since Phase 2b's
`pkg/drift`.**

`pkg/drift/gitindex.go`'s `GitSource.build` caches each repo's symbol table at
`.forge/code-index-<repo>.json` — one file per configured `--repo name=path`, keyed by
name. ADDENDUM §B.6 and DESIGN §15 both describe a single `.forge/code-index.json`, no
per-repo suffix. The code's shape is the correct one: `forge drift`/`forge check`/`forge
logback` all accept repeatable `--repo`, so a single shared filename would let a second
repo's index overwrite the first's on the very next run. Nothing is broken — every
caller that reads the cache already goes through `GitSource`, which knows its own
suffix — but a reader following the docs literally would look for a file that never
exists under that name. Fix is documentation, not code: update ADDENDUM §B.6 and DESIGN
§15 to show the `-<repo>` suffix, or update `gitindex.go`'s doc comment for `build` to
say explicitly why it deviates, whichever the next phase touching either file finds it
easier to keep in sync.

### Took the second option, 2026-08-21 — plus one site this entry missed

Chose the doc-comment route: `pkg/drift/gitindex.go`'s `persist` now states the filename,
says the suffix is required rather than cosmetic (repeatable `--repo` means a shared name
would let repo two overwrite repo one), and points here. ADDENDUM §B.6 and DESIGN §15 are
**deliberately left saying the singular name**, per the standing "record, don't fix" rule —
which is exactly why the code comment had to carry the explanation instead.

**This entry scoped the problem as docs-only, and that was wrong in one place.** Three
sites inside the tree also claimed the singular name:

- `agents/forge-codebase-scout.md:33` told that agent to *"seed your search from
  `.forge/code-index.json` when it exists."* That path never exists on disk, so the
  instruction was not stale prose but a wrong instruction to a live component — the only
  site here with operational consequence. Now names the `-<repo>` pattern and says to glob
  for it.
- `pkg/codeindex/index.go:52` and `store.go:9` both asserted the singular path in doc
  comments. Fixed the honest way rather than by substituting a different hardcoded name:
  `pkg/codeindex` does not choose the filename — its caller does — so the comments now say
  that. `cachePath` in `pkg/drift` remains the single place a name is constructed.

Still open: ADDENDUM §B.6 (`:247`, `:318`) and DESIGN §15 (`:714`, `:954`), plus
`CLAUDE-CODE-PROMPT.md:208,365,458` and `ROADMAP.md:53`. `examples/vault/` is out of scope
by construction — it is scrubbed vault content, a historical artifact, not documentation.

---

## B-028 — the drift hook path gains no immediacy on a deleted-file citation

**Owner: whoever next touches `pkg/drift/check.go`'s `checkRef`/`unresolvedPath`. Status:
done, 2026-08-17 — see the closure note below for the gate-ordering correction and the
residual limitation it does not close.**

**Closure note.** Fixed largely as this entry's own sketch describes, plus one correction
the sketch did not anticipate. `coderef.ChangedFilesStatus` (`pkg/coderef/scan.go`) is a
new, additive sibling to `ChangedFiles`: same `git diff --no-renames` cheap-gate call,
`--name-status` instead of `--name-only`, so it can tell a same-commit deletion from an
edit without a second subprocess. `cmd/forge/drift.go`'s `gateOf` now returns a
`*drift.Changed{Touched, Deleted}` instead of a flat `map[string]bool` — `Touched` keeps
its existing over-inclusive-across-repos shape, `Deleted` maps each deleted path to the
repo that reported it, because a `Broken` finding built from it needs a real repo to
resolve against. `checkRef`'s `Unresolved` case now delegates to a new
`checkUnresolvedPath`, which on the hook path (`changed != nil`) checks the citation's
basename against `changed.Deleted` via a new `deletedInGate` (basename-equality only, not
`pkg/coderef`'s unexported subsequence matcher — by the time this path is reached the
registry has already failed to resolve the citation, so there is no candidate set left to
disambiguate among) and verdicts `Broken` on a match, with no historical `git ls-tree`
scan and no `--deep` required.

**The gate-ordering correction.** A literal reading of this entry's sketch — verdict
`Broken` on a match, otherwise fall through to the existing `Skipped` — is self-defeating.
Trace it: a file deleted in commit C demotes the note to `low`; an unrelated commit D then
fires the hook with `--since-commit C..D`; the citation is still `Unresolved`, but its
basename no longer appears in *this* narrower range's `Deleted` set, so a naive fix still
emits `Finding{Skipped}` — and `apply.go`'s own invariant (a note with no `Broken` finding
falls through to `restore`) would then flip the note straight back to `high` while the
file is still gone. The actual fix: on the hook path, an `Unresolved` citation with no
deletion match produces **no finding at all**, mirroring the resolved-path gate `checkRef`
already applies a few lines up. `unresolvedPath` itself shrank back to being purely the
full-sweep branch it was already documented as — `checkUnresolvedPath` only delegates to
it when `changed == nil`. `TestRollbackSymmetryOnDeletion`
(`pkg/drift/rollback_test.go`) is the regression test for exactly this: same-commit
deletion demotes, an unrelated later commit does *not* restore it, and reverting the
deletion does restore it — real repo, real `Apply`, real `.forge/` store, not a unit test
against `Check` alone.

**What this does not close**, same shape as B-026's own caveat: a basename collision (an
unrelated same-basename file deleted elsewhere in the same commit range) can still match
the wrong path in `deletedInGate` — narrow, deterministic (lexicographically first path
wins, so two runs on the same tree never disagree), and not engineered away, matching the
class of imprecision B-026 already accepted for the full-sweep case. Distinct from a
citation whose basename still exists elsewhere at HEAD — that resolves `Ambiguous`
upstream in `Registry.Resolve` and never reaches this code at all, unaffected by this fix.
Verified via `CGO_ENABLED=0` and `CGO_ENABLED=1` `go test ./...`, both green, plus a
hand-built smoke test against a real temp repo confirming `forge drift --since-commit
<sha> --apply` with no `--deep` demotes a note to `low` immediately on a same-commit file
deletion — the case that was previously impossible on any automated path.

B-026's fix only fires on a full sweep (`opts.Deep == true`, `changed == nil`): `forge
drift`'s default hook-path run (`opts.Deep == false`) still leaves a same-commit
deletion `Skipped` until the next weekly `forge check`. The hook already computes
`changed` — the set of paths the triggering commit range touched — cheaply, as the gate
`checkRef` applies after resolution. In principle that same set could prove existence:
if an `Unresolved` citation's basename appears among `changed`'s deleted paths, that is
strong same-commit evidence of a deletion, no historical `git ls-tree` scan required.

Not worth building now: `pkg/coderef`'s subsequence-over-basename matching
(`Registry.lookup`, `matching`, `isSubsequence`) is unexported, so proving a citation
*would have* matched a since-deleted path means either exporting that matching logic for
`checkRef` to call standalone, or building a second registry from the pre-commit tree
just for this — both cut across `checkRef`'s separation of concerns (resolve, then
verdict) for a gap `absentSymbol` already accepts for symbols. Revisit only if the
hook-path latency gap (deletions caught up to a week late) turns out to matter in
practice.

---

## B-029 — `errcheck` is disabled tree-wide in `.golangci.yml`

**Owner: recorded during Phase 6's CI delta. Status: open, scoped deliberately.**

Phase 6 added `golangci-lint` to `ci.yml` (`docs/CLAUDE-CODE-PROMPT.md`'s original Phase 6
intent). Its default linter set (`errcheck`, `gosimple`, `govet`, `ineffassign`,
`staticcheck`, `unused`) surfaced four `staticcheck`/`gosimple` findings, all confirmed
false positives against code that already documents the intent inline — two determinism
tests calling the same function twice and comparing the results (`Sign(noteA) !=
Sign(noteA)` in `pkg/similarity/similarity_test.go`, `QHash(...) != QHash(...)` in
`pkg/telemetry/qhash_test.go`), one deliberate nil-map panic under test
(`cmd/forge/check_test.go`'s `TestOneBadRendererCostsOneFile`), and one struct literal
that `pkg/engine/host.go` keeps explicit on purpose, per its own comment, so `Request`
growing fields later can't silently leak into `instruction`'s JSON contract. All four are marked
with a scoped `//nolint` at the specific line, not a blanket exclusion.

`errcheck` was a different matter: ~20 findings spread across `cmd/forge` and `pkg/drift`,
`pkg/report`, `pkg/store`, `pkg/engine`, `pkg/sentinel`, `pkg/telemetry`, `pkg/codeindex`.
Some are already-documented deliberate ignores (`pkg/drift/demotions.go`'s `json.Unmarshal`
next to the comment "a corrupt store loses restore targets, never verdicts"); others are
test-setup fire-and-forget calls; a few — `cmd/forge/recall_load.go`'s `tx.Commit()`/
`tx.Rollback()`, `pkg/drift/gitindex.go`'s `codeindex.Save` on the hook path — are the kind
where silently swallowing the error deserves individual judgment, not a rushed sweep under
a CI-delta task. Sorting all ~20 into "already-fine, needs a nolint" vs. "needs a real fix"
is real work, out of scope for landing the lint step itself, so `.golangci.yml` disables
`errcheck` tree-wide for now with a comment pointing here.

Shape of the fix when someone owns it: go file by file, add `//nolint:errcheck` with a
reason comment where ignoring is correct (matching this entry's four `staticcheck`
precedents), fix the rest to actually check the error, then remove `errcheck` from
`.golangci.yml`'s `disable` list.

### Re-measured 2026-08-21. The "~20" above is an undercount and the package list is short.

Not fixed this pass — deliberately, it needs its own session — but sized honestly so that
session does not start on a wrong estimate. Three numbers, provenance kept separate
because they are not equally trustworthy:

| Count | Provenance |
|---|---|
| **95** | **Measured.** `errcheck v1.9.0 ./...`, byte-identical under `CGO_ENABLED=0` and `=1` |
| **~37** | **Derived, not measured.** Hand-application of golangci-lint's default `EXC0001` exclusion (`.*Close`, `.*Flush`, `os.Remove(All)?`, `.*print(f\|ln)?`, `os.(Un)?Setenv`) to the 95. golangci-lint itself was never run |
| ~20 | This entry's original claim, above |

Even taking the conservative ~37, the estimate was low by roughly half. `.golangci.yml` has
no `issues:` block, so `exclude-use-default: true` applies under v1.64.8 — that is
reasoning about the tool's defaults, not an observation, hence the middle row's caveat.

**Three packages this entry never named have findings:** `pkg/dataset` (10),
`pkg/qualitygate` (2), `pkg/linkcheck` (1).

One scope worry can be closed: the two build lanes produce identical output. The only
tag-gated files are `pkg/codeindex/parse_cgo.go` / `parse_nocgo.go` and neither has a
finding, so there is no CI blind spot hiding behind `CGO_ENABLED`.

**Of the ~37, only 10 are production code; the other 27 are `_test.go` fire-and-forget.**
Three of the ten are not lint compliance at all and should be triaged before the sweep:

1. **`cmd/forge/recall_load.go`'s `refresh()` is a correctness bug that errcheck only
   partly sees.** errcheck flags `tx.Rollback()` (`:108`) and `tx.Commit()` (`:112`), but
   `:104` swallows `st.DB.Begin()`'s error on a line errcheck cannot flag, and *every*
   path in the function returns `nil` — so a failed commit is reported as success and the
   caller (`:58`, which does propagate) is told nothing went wrong. There is already an
   exemplar 100 lines away: `cmd/forge/index.go:217` is
   `func commit(tx *sql.Tx) error { return tx.Commit() }`, checked at `:211`. The fix is
   "use the helper that exists," not "add an error check." **Highest-value item in B-029;
   arguably worth pulling out ahead of the lint work.**
2. **`pkg/drift/apply.go:109` and `pkg/drift/gitindex.go:48` need signature changes, not
   `//nolint`s.** `stamp()` returns nothing and both callers (`apply.go:74`, `:91`) ignore
   it; `persist()` returns nothing and its caller (`gitindex.go:38`) ignores it. Checking
   those errors means changing signatures and call sites. Size accordingly.
3. **`pkg/codeindex/catfile.go:47`** ignores `cmd.Wait()`, so a `git cat-file` that exits
   non-zero after partial output reads as success. `:55`/`:57` compound it — the write and
   its flush are both unchecked, making a write error unrecoverable regardless.

Two smaller notes for that session: `cmd/forge/index.go:207` is a *different* class from
`recall_load.go:108` — it does `tx.Rollback(); return err`, so the caller does learn of the
failure and Rollback's own error is genuinely secondary; that one is a `//nolint`. And
there is currently **no `//nolint:errcheck` anywhere in the tree**, so this introduces the
first; match the four existing precedents' style exactly — `//nolint:<linter> // <lowercase
reason>`, no trailing period, on the offending line (`pkg/engine/host.go:22`,
`cmd/forge/check_test.go:131`, `pkg/similarity/similarity_test.go:134`,
`pkg/telemetry/qhash_test.go:6`).

Closing it is a one-file edit: delete `.golangci.yml`'s `disable:` block. No Makefile or
workflow change — `make lint` is gofmt + `go vet` only and never ran errcheck.

---

## B-030 — `dataset.capture` accepts five tiers; only two of them gate anything

**Owner: whoever next touches `pkg/dataset` or `config/forge.config.example.md`. Status:
recorded, not fixed — found 2026-08-21 while scoping B-024, out of that fix's scope.**

`config/forge.config.example.md:170` ships `dataset.capture: [d1, d2, d3, d4, d5]`, and
`pkg/config/types.go:210-216`'s doc comment explains the list form as future-proofing:
*"Capture names the tiers d1…d5, not booleans, so a new tier is a list entry rather than a
schema change."* Reasonable. But `cfg.Dataset.Capture` has exactly **two** readers in the
whole tree:

- `cmd/forge/engine_run.go:127` → `dataset.Enabled` → `D2Tag`
- `cmd/forge/gate.go:157` → `dataset.D4Enabled` → `D4Tag`

There is no `D1Tag`, no `D3Tag`, no `D5Tag`. So:

- **`d1` and `d5` are inert because nothing implements them.** Harmless — the list is
  ahead of the code, which is what the type's doc comment describes.
- **`d3` is the misleading one.** D3 *is* implemented (`pkg/dataset/d3.go`, driven by the
  vault's post-commit hook via `cmd/forge/capture.go:48`) and it never consults
  `cfg.Dataset.Capture` at all. So removing `d3` from the list **does not stop D3 capture**
  — the hook keeps writing. A user who edits the list to opt out of human-edit capture
  gets no error, no warning, and no change in behavior.

That last case is the reason this is filed rather than shrugged off: an opt-out control
that silently does nothing is worse than an absent one, and D3 is the tier that records
human edits to notes, i.e. the one a privacy-minded user is most likely to try to turn off.

Not a naming bug and **not the same class as B-024**, which was a real string mismatch on a
gate that did exist. Do not "fix" this by renaming anything.

Three shapes, none chosen: make `cmd/forge/capture.go` honour the list (a behavior change
to the hook path — it would need to stay silent-and-never-fail like the rest of that path);
or document in `config/forge.config.example.md` and `pkg/config/types.go` exactly which
entries are live today; or drop the unimplemented entries from the packaged list and let
them return when their tiers land. The middle option is the smallest honest one; the first
is the only one that makes the control real.

`pkg/dataset/capture_gate_test.go`'s `TestPackagedCaptureListGates` is deliberately scoped
to D2 and D4 with a comment pointing here, so it pins the gates that exist rather than
inventing assertions about three that do not.
