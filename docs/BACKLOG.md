# Backlog

Work that came up outside the current phase's scope. Per `CLAUDE.md`, things land here
instead of getting built. Nothing in this file is scheduled; each entry names the phase
that should own it.

**Plans live in `docs/TODO.md`.** This file is *why* an item exists; TODO is *how to close
it* — step-by-step, with verified `file:line` anchors, a verification command, and the
prohibitions each entry carries. Nine items have a full plan; the rest have an index row
or an unblock condition. Written 2026-08-23.

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

**Owner: whenever, low cost. Status: open — user confirmed keep-as-is, 2026-08-24.**

The Go module was renamed to `knowledge-forge` on 2026-08-08, but the directory is still
`/Users/mimir45/TIL`, and the docs call the artifact `knowledge-forge/`. Purely cosmetic
now; it becomes mildly annoying once tooling, README paths, and the goreleaser config in
Phase 6 assume the artifact name. Renaming the directory is a user decision (it breaks
any shell aliases, IDE projects, and `.idea/` state pointing at the old path).

**Asked directly, 2026-08-24.** The user asked to "change repository name from TIL"; on
finding both the GitHub remote (`Knowledge-Forge`) and `go.mod`'s module path
(`knowledge-forge`) already renamed, and the local directory being the only thing left
answering to `TIL`, they were asked whether to rename the directory too and chose not to
— "leave this as TIL" — as long as it stays cosmetic. Reaffirms this entry's existing
"do not rename unasked" answer rather than changing it; nothing renamed.

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

**Owner: Phase 3. Status: closed 2026-08-22 — absent-term admission shipped on top of 2b's
weighting, and §3.1 is now generated rather than transcribed. Read the closure note at the
end first if you only read one part: the two earlier passes below diagnosed correctly and
fixed nothing, and their reasoning is kept in place rather than rewritten.**

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

### Closed 2026-08-22. What shipped, what it cost, and what was split out.

The harness came first, as the sizing note demanded. `cmd/forge/calibration_test.go` runs
§3.1's nine queries against `examples/vault` (92 scored docs, git-tracked) and diffs a
golden table; `-update` re-records it. The corpus is staged into a temp dir per run,
because `loadDocs` opens a SQLite cache under `<root>/.forge` and writes rows back —
scoring in place would have mutated a tracked directory and made the golden depend on
whether the cache happened to be warm. The "before" column was measured against the
unmodified scorer and committed before a line of the fix was written.

**Two of this entry's own sizing-pass findings were wrong, corrected by that measurement:**
`docker-compose-init-container-pattern` *is* in `examples/vault` — the file is
`…-with-health-gated-sequencing.md` and the entry's shorthand slug is simply shorter — so
the argument against lowering the threshold could be held against the note it actually
names. "Java virtual threads" likewise returns a candidate. And the worry that the corpus
had drifted too far to compare does not bite: eight of the nine "before" scores reproduce
§3.1's original numbers exactly, the ninth being the 0.501 this entry itself re-measured.

**The fix is two changes, as predicted, and the open decision was the weight.**

1. The vocabulary filter changed sides: it applied to question terms and not to `--stack`
   hints, and now applies to hints and not to questions. A hint is a user filter — `--stack
   kotlin` against a Kotlin-less vault must not dilute every note — and was harmless
   unfiltered only while unknown terms weighed zero. A question term is evidence.
2. An absent term weighs **the mean of the present ones**, assigned in the weight-map
   builder rather than inside `idf()`. That placement is why `score_test.go`'s
   `idf(0, 91) == 0` assertion was **preserved rather than inverted**, contrary to what
   this entry predicted: inverse document frequency is defined inside the corpus and is
   honestly zero outside it; what an absent term is worth as *evidence* is a policy one
   layer up. The mean is parameterless, keeps the k-of-m reading of the ratio, and stays
   under `idfCap` by construction. Flooring `df` at 1 — the alternative sketched above —
   would hand an absent term the maximum weight and invert the cap's purpose.

**Measured, same run, `examples/vault`:**

| | before | after |
|---|---|---|
| `spring-cli-and-maven-commands-for-spring-boot` ← "Redis caching in Spring Boot" | 0.740, UPDATE(extend), 1st | **0.415, CREATE, 2nd** |
| `docker-compose-init-container-pattern…` ← "Docker multi-stage build cache optimization" | 0.429 | **0.163** |
| `storybook-isolated-component-development…` ← "Storybook interaction testing with play functions" | 0.617, UPDATE(extend) | **0.217, CREATE** |

Seven of nine top-1 winners are unchanged. The two that moved are the same note losing the
same inflated channels in two queries.

**The pre-registered criteria, and the one that failed.** Criterion 1 ("top-1 unchanged for
eight of nine, case 1 excepted") **failed as written** — two winners moved. Read from
`--explain`, the second was case 1's note losing tags 1.000→0.200 and stack 1.000→0.400 in
a second query, letting the note already sitting second at 0.581 through. That is the
fix's own subject appearing twice, not the ranking scramble the criterion was written to
catch, and the reinterpretation is recorded here rather than made silently. Criteria 2–5
passed; the thresholds did not move.

**The cost, stated plainly.** Every UPDATE verdict in the "before" column came from one
artifact: a frontmatter channel reading 1.000 off a denominator of a single surviving
term. The Storybook row is the case against the fix — that note genuinely is the one to
extend — but its 0.617 was the same artifact (`tags` 1.000 off `{testing}` alone, `stack`
1.000 off `{storybook}` alone). The distinction between "artifact that was wrong" and
"artifact that was right by luck" does not exist in the data. Four narrower admission
rules were measured against that row: 0.217 as shipped, 0.242, 0.377, 0.392. None restores
it; only a threshold change would, and this entry forbids that by name three times. A
term-level distinction is not available either — `redis` and `play` each appear in exactly
one note corpus-wide.

**Nine-of-nine CREATE is not a dead decision tree**, which was the other thing worth
checking before shipping. Those queries are adjacent-topic by construction. Queries naming
a note's actual subject still clear the bands: "what is the transactional outbox pattern"
and "hexagonal architecture ports and adapters" hold at 1.000 ANSWER_FROM_VAULT,
"Storybook decorator pattern for Redux providers" 0.823→0.652 UPDATE, "Testcontainers
Docker based integration testing" 0.814→0.626 UPDATE. One demotion: "how does keyset
pagination work" 0.917 ANSWER → 0.729 UPDATE(extend), against a note narrower than the
question — arguably the better reading, but recorded as a demotion rather than argued away.

**Shipped on the user's decision**, taken with the above on the table: the damaging
direction is UPDATE(extend), because extending **writes** into an existing note, while a
miss produces a new one. Three items were opened rather than folded in — **B-031** (the
Kafka/Testcontainers miss is a coverage defect and is split out of this entry), **B-032**
(an untagged note escapes the absent-term penalty entirely, §2.5's asymmetry running the
other way), **B-033** (adjacent-topic queries now fall under the 0.30 neighbour floor and
get linked to nothing).

Spec updated in the same pass: §2.3 and §2.4's formulas were stale by *two* changes, since
2b's IDF weighting was never documented at all; there is now a §2.3.1 covering weights and
admission, a §2.5 note distinguishing query-side admission from note-side activation, a
generated §3.1, and a §4.1 example that matches what the binary prints.

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

**Owner: whichever phase makes `moc/codebase.md` a dependency map. Status: closed
2026-08-24 — see the closure note below.**

ADDENDUM §B.5 asks the codebase map to show what depends on what. The struct field exists;
nothing fills it, because `codeindex.File` captures declarations only and no import edges. So
the MOC currently groups by directory and ranks by churn but draws no arrows.

Adding imports to the extractor is the real work; the grouping is the smaller problem. Note
also that "module = directory" is an honest limitation, not a placeholder: nothing in the
index knows about Maven modules or Go packages, and inventing a grouping the code does not
declare would file code under modules its authors never wrote.

### Closed 2026-08-24, out-of-phase, on `feat/b-015-codegroup-dependson` — TODO.md's plan
followed as written, with two verified assumptions the plan flagged as unchecked.

`codeindex.File` gained `Imports []string` (Extractor bumped **2 → 3**, in the same commit
as the shape change per B-013); Java and TypeScript both extract it in `parse_cgo.go`'s
existing `walk`. Resolution and folding-to-directory happen together in the new
`cmd/forge/check_codebase_deps.go`'s `dependsOn`, called from both `check_codebase.go`'s
`oneCodebase` and `logback_map.go`'s `buildGroups` — the two existing places that build a
`[]report.CodeGroup`, unchanged otherwise.

Two things TODO.md's plan named as risks and asked to be checked before writing the
resolver, both verified rather than assumed: **(a)** the tree-sitter node shapes — a
throwaway probe (deleted before commit) confirmed Java's `import_declaration` exposes the
qualified name as a `scoped_identifier`/`identifier` child with `static` and a trailing
`.*` already excluded by the grammar itself (a wildcard import and a class import are
structurally identical strings, which `resolveJavaImport`'s right-to-left trim-and-retry
is built around), and TypeScript's `import_statement`/`export_statement` both expose the
specifier as field `source`, absent exactly on a plain declaration (`export function X`).
**(b)** `coderef.ScanRepo`'s file list — confirmed filtered to `sourceExt` (six code
extensions), not a full tree listing, which is why `dependsOn` matches TypeScript imports
against an exact `fileSet` (extension-tried candidates) rather than a directory-existence
guess: the advisor's flagged false-edge case (`../other/Nonexistent` resolving to `src/
other` merely because the parent directory exists) does not reproduce, because the
resolver requires a real matching file, not just a real matching directory.

Java resolution has no source root to anchor from (Maven's `src/main/java`, or nothing at
all), so it matches a file path by **suffix** — the plan's own words, "resolve then fold to
directory" — trying the full dotted import as a class file first, then one segment
shorter, to cover a static member import (`import static a.b.C.FIELD`) the same way as a
plain class import; a wildcard import (star already stripped) falls through to matching
the untouched string as a directory suffix. Suffix matches are deterministic by
construction — the lexicographically first candidate wins (B-020's rule) — for the case
TODO.md called out as genuinely ambiguous: two modules sharing a package fragment under
different source roots.

Verified: both build lanes (`CGO_ENABLED=0/1 go build ./... && go test ./...`) green, `go
vet` clean on both. `TestGitSourceRebuildsFromScratchOnStaleExtractor` (`pkg/drift`) is the
plan's own stronger ask honored — not `codeindex.Load` in isolation, but the real hook path
(`GitSource.build`) proven to take the full-rebuild branch, not `Patch` a bogus stale entry
forward, against a cache file stamped by `Extractor - 1`.
`TestGroupsOfPopulatesDependsOnEndToEnd` drives the real pipeline (a temp git repo, two
Java packages and a TypeScript pages/widgets pair, `codeindex.Build` then `dependsOn` then
`groupsOf`) rather than a hand-built `Index`, per the advisor's point that the unit tests
alone would pass even if the grammar assumption were wrong.
`TestDependsOnIsDeterministicAcrossRuns` guards the map-iteration-order bug B-020 named
four times over in 2b, run 20 times against a directory with four ambiguous-looking
imports — that is what was actually checked; the plan's own verification line asked for
`moc/codebase.md` rendered six times and diffed, which was not run. The substitution is
sound (`sortGroups` predates this item and `RenderCodebase` is a pure function of its
input, so a deterministic `dependsOn` is sufficient), but it is a substitution, not the
literal check. Not re-measured, not claimed: `forge drift`'s <100ms hook budget — this item
adds no work to that path at all (`DependsOn` is `forge check`/`forge logback` only), so
there is nothing to re-measure, but B-029's precedent is to say so rather than imply it.

`Patch` (`store.go`) copies a persisted index's untouched `Files` forward wholesale and
re-stamps `Extractor` unconditionally on its output — traced rather than assumed: its one
caller, `GitSource.build`, only reaches it after a `Load` that already rejected any
`ix.Extractor != Extractor`, so every entry `Patch` carries forward was itself written by
the current extractor. `Patch` does not re-check this itself; `store.go`'s doc comment now
says so, for whichever future caller might skip that gate.

One honest limitation this item did not close, named rather than silently accepted: a
TypeScript barrel file (an `index.ts` that only re-exports, no declarations of its own)
has imports but zero symbols, and `build.go`'s `parseAll` drops any file with zero symbols
from `ix.Files` entirely — pre-existing behavior, out of this item's scope to change. Such
a file is still *resolvable as a dependency target* (`resolveTSImport` matches
`joined+"/index"+ext` without needing the barrel itself indexed), but it is invisible as a
*dependent* — its own re-exported imports never appear in any `DependsOn`, because the
file that would carry them was never parsed into the index in the first place.

One real, one-time cost worth naming because nothing else records it: the Extractor bump
forces a full re-parse on the first hook run after upgrade, on every repo with a persisted
`.forge/code-index-<repo>.json` — `store.go`'s own comment says a full build "takes
seconds" against the very budget that cache exists to protect. That is a one-time cost at
upgrade, not a standing one, but it is real and unmeasured on any of the three repos the
vault's citations name (none are on this machine, per B-029's closure).

`pkg/report/knowledgemap.go`'s stale "nothing populates it" comment was replaced, not
deleted per TODO.md's own instruction — `RenderKnowledgeMap` still omits the column, now
for a surviving reason (`moc/codebase.md` already renders it; this map is deliberately that
report's notes-join minus everything else), not because the column would always be empty.

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

**Owner: whoever decides whether `stop` and `degrade` should diverge. Status: closed
2026-08-24 — see the closure note below. Doc half was 2026-08-21; behavior half is this
close.**

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

### Behavior half closed 2026-08-24 — option (a): `stop` gets a real exit, `degrade` unchanged.

Chosen over the other two named above: `degrade` staying today's silent fallthrough is
already the honest reading of the word, so nothing there needed to change; `stop` gets a
distinct, tested non-zero exit instead. Collapsing to `queue | degrade` (dropping `stop`)
was rejected as backward-incompatible for no real gain, and documenting `stop`/`degrade`
as synonyms was rejected because TODO.md's own entry called that the option that "leaves
the entry half-open forever."

`cmd/forge/engine_run.go`'s exhaustion check (previously gated on `rel != "" &&
OnExhausted == "queue"` alone) now runs whenever `engine.Resolve` has degraded the stage
to `"none"` for lack of budget, and dispatches on `OnExhausted` in a new `onExhausted`
helper: `"queue"` keeps its existing `--rel`-gated `pending_advisor: true` stamp and falls
through to `none` exactly as before; `"stop"` prints to stderr and returns exit 1 without
ever calling the tier; `"degrade"` (or any other validator-accepted value) is the
unmodified silent fallthrough. `pkg/engine/select.go:30`'s unconditional degrade is
untouched — `on_exhausted` is still read only in `cmd/forge`, one layer up, which is where
Exhausted() already lived for exactly this reason.

New tests: `TestOnExhaustedBehaviorDiverges` (`cmd/forge/engine_run_test.go`) forces the
exhausted path (api capped at $0.00, no fallback) for all three values and asserts both
the exit code and whether `pending_advisor` got stamped — `stop` exits 1 and stamps
nothing, `queue` exits 0 and stamps, `degrade` exits 0 and stamps nothing.
`TestOnExhaustedStopDoesNotFireWhenBudgetAvailable`
(`cmd/forge/engine_run_httptest_test.go`) is the other direction: with budget available,
`on_exhausted: stop` must not fire — the real HTTP call still runs and spend is still
booked. Verified: `go test ./pkg/engine/... ./cmd/forge/...` and the full `go test ./...`
green under both `CGO_ENABLED=0` and `CGO_ENABLED=1`; `go vet ./...` clean. The four doc
sites (`ADDENDUM.md:117`, `:485`, `:671`; `CLAUDE-CODE-PROMPT.md:339`) were not touched —
they only ever enumerated the three value names, never claimed a behavior for `stop`, so
there was no wrong claim to correct.

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

**Owner: whoever next edits ADDENDUM §B.6 or DESIGN §15. Status: closed 2026-08-23 — the
code side and one wrong agent instruction were fixed 2026-08-21, the eight doc sites on
2026-08-23. Originally found during Phase 5b's explore pass while
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

### Closed 2026-08-23. The docs were edited, and here is why that is not a rule change.

The eight sites listed above now show the `-<repo>` suffix: ADDENDUM `:247` (§B.6), `:318`
and the phase table at `:563`; DESIGN `:714` and `:954` (§15); `ROADMAP.md:53`;
`CLAUDE-CODE-PROMPT.md:208,365,458`. Three Turkish mirrors that described this entry as
half-open were corrected with it (`docs/tr/02-MIMARI.md:749`, `03-KULLANIM-KILAVUZU.md:761`;
`04-DOSYA-DOSYA.md:349` was already accurate and is unchanged).

**The standing "record, don't fix" rule is untouched, on two grounds.** First, it exists so
a doc does not shift under a phase that is mid-flight against it; the roadmap ended at 6b,
so there is no such phase. Second and more important: that rule and AUDIT §8.4's mechanism
govern **decisions** — a line where the doc says one thing, a later decision says another,
and §8.4 is what a reader follows. B-027 is not one of those. Nobody disagrees about the
design; the docs simply name a file that has never existed on disk under that name. A
reader following §8.4's mechanism here would be told to follow a doc that is factually
wrong about a path. Correcting a filename is not overriding a decision, and no §8.4 entry
was added, because there was no decision to record.

The two normative sites (ADDENDUM §B.6, DESIGN §15) carry a one-line marker saying they
said the singular name until 2026-08-23 and pointing here, so the edit is traceable rather
than a silent rewrite. The reason itself — repeatable `--repo` means a shared name lets
repo two overwrite repo one — is now stated at §B.6 rather than living only in
`pkg/drift/gitindex.go`'s `persist` doc comment.

**Unchanged, deliberately:** the filename on disk. `cachePath` in `pkg/drift` stays the
single place a name is constructed. `examples/vault/` is scrubbed vault content and a
historical artifact, out of scope by construction.

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

**Owner: recorded during Phase 6's CI delta. Status: closed 2026-08-23.** Triage item 1
landed 2026-08-22; the sweep and the `disable:` block closed 2026-08-23. **Every count in
this entry and in its 2026-08-21 re-measure is wrong — read the last section first.**

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

`errcheck` was a different matter: ~20 findings — the real figure is **35**, see the
closing section — spread across `cmd/forge` and `pkg/drift`,
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

### Triage item 1 landed 2026-08-22. The prescribed fix was wrong; two claims corrected.

Pulled out ahead of the sweep, as the section above suggested. **The entry's own
prescription — "use the helper that exists", i.e. propagate through `commit` — does not
survive a trace of the callers and was not followed.**

`loadDocs`' error reaches `runRecall` (`cmd/forge/recall.go:77`), which prints it and
returns 1 **without emitting the candidates it has already scored correctly**. A concurrent
`forge intent` on the UserPromptSubmit hook holding the SQLite write lock is enough to
produce that. So propagating would trade a stale cache — self-healing, since `rowFor`
re-parses whatever the cache does not match — for a discarded correct answer, on the
command whose entire job is to emit that answer. This entry was sized from errcheck output,
not from `recall.go:77`.

**What was actually wrong is the signature, and that is what changed.** Every path returned
`nil`, so `return docs, refresh(...)` read as propagation while guaranteeing the opposite,
and a later edit that started returning a real error would have turned a transient lock into
an exit 1 with nothing in review to catch it. `refresh` no longer returns `error`, so the
promise matches the behaviour; its body moved to `writeRows`, which checks all three errors
the old one dropped — including `st.DB.Begin()`, on an assignment errcheck cannot flag, which
is why this function's finding count was understated. The swallow is now a single
`_ = writeRows(...)` with the reasoning above beside it. `commit` **is** reused, as the entry
asked — just inside `writeRows` rather than as the propagation path.

Two tests pin it (`cmd/forge/recall_load_test.go`): `writeRows` reports a failed write, and
`refresh` survives one. The point of the split is that the error must exist before ignoring
it can be a decision rather than an accident. A third, asserting the empty case opens no
transaction, was written and then deleted: it could not fail, because a `refresh` that did
open a transaction on a closed store would swallow the error and pass anyway. A comment
promising a check the test does not make is the defect this item exists to fix.

**No observable behaviour change**, and the commit says so. A failed cache write was
swallowed before and is swallowed now.

**Item 3's claim is weaker than stated and should be re-sized before that session.** This
entry says `pkg/codeindex/catfile.go:47`'s unchecked `cmd.Wait()` means "a `git cat-file`
that exits non-zero after partial output reads as success". Traced 2026-08-22: it does not.
`drainBlobs` (`:62-78`) reads exactly one reply per requested file and returns the error
from `ReadString`/`io.ReadFull`, so a `git` that dies after partial output surfaces as EOF
and `Build` (`pkg/codeindex/build.go:22`) returns it. The same EOF covers a failed
`w.Flush()` in `feedRequests`, since fewer requests means fewer replies. What `cmd.Wait()`
actually hides is the narrow case of **all** replies delivered followed by a non-zero exit.
So item 3 is not a bigger correctness risk than item 1 was, and it should not be scheduled
as though it were.

One genuinely unguarded path was noticed while tracing and is **not** an errcheck finding,
so it belongs to whoever owns `catfile.go` rather than to the sweep: `blobSize` returns
`!ok` for any header that is not three fields ending in a blob size, and `drainBlobs`
treats every such line as the documented `"<name> missing"` case and `continue`s. Any other
unexpected line — a `git` diagnostic, a truncated header — therefore desynchronises the
request/reply stream, after which blob bodies are parsed as headers. Pre-existing, not
introduced here, and worth its own look because the index feeds drift's verdicts.

### Closed 2026-08-23. Every number above was wrong, in both directions.

`errcheck` is off `.golangci.yml`'s `disable:` list; `golangci-lint run ./...` under the
CI-pinned **v1.64.8** returns zero findings, both build lanes green.

**The sizing table is superseded.** It compared an `errcheck` binary's raw output against a
hand-derived exclusion estimate, and never ran the linter CI actually runs. Measured this
session, all four numbers:

| Count | What it is |
|---|---|
| **105** | `errcheck v1.20.0 ./...`, raw. Byte-identical across `CGO_ENABLED=0` and `=1` — the entry's "no CGO blind spot" finding reproduces |
| 50 | `golangci-lint` **v2** (`--default=none --enable errcheck`). A different tool with a different exclusion set; recorded only so the number is not mistaken for the one below |
| **35** | **The worklist.** `golangci-lint` v1.64.8 — the version `ci.yml` pins — with the repo's own config and `--max-same-issues=0 --max-issues-per-linter=0` |
| 22 | The same run at **stock truncation limits**. See below |

The 105 vs. 95 gap is `errcheck` v1.20.0 vs. v1.9.0, not code drift. The derived "~37"
landed near the truthful 35 by luck: it was applied to the wrong raw count with the wrong
tool's exclusions.

**35 splits 26 test / 9 production, not 27 / 10.** The entry's "three packages this entry
never named" — `pkg/dataset` (4, all test), `pkg/qualitygate` (0), `pkg/linkcheck` (0) —
mostly vanish under v1.64.8's default exclusions; the two `pkg/qualitygate` and one
`pkg/linkcheck` findings were `EXC0001` cases all along.

**A truncation trap was found and closed, and it is the part of this item worth
remembering.** golangci-lint v1 defaults to `max-issues-per-linter: 50` and
`max-same-issues: 3`. With errcheck enabled and stock limits the same tree reports **22**
findings, not 35 — thirteen repeats silently dropped. A gate that under-reports is worse
than one that is noisy, so `.golangci.yml` now sets both limits to `0` and says why. This
was never an `errcheck` question; it applies to every linter in the default set and was
in force for all of Phase 6.

**Item 2's prescription was traced and not followed, for the same reason item 1's was
not.** The entry says `pkg/drift/apply.go:109`'s `stamp()` and `pkg/drift/gitindex.go`'s
`persist()` "need signature changes, not `//nolint`s." They do not, and the distinction
against `refresh()` is the point: `refresh` promised propagation it never performed, and
its body hid three swallowed errors including one errcheck could not see. `stamp` and
`persist` each contain exactly **one** call, their failures are self-healing, and neither
caller has an error channel to receive one — `applyNote` returns `(Result, bool)` and
`GitSource.full` returns a bare `codeindex.Index`. Giving them an `error` return would move
the ignore from one site to two and add a return value that every caller must discard.
A failed `stamp` leaves `drift_checked_at` stale, so the next run re-evaluates the note —
the same non-event `demote`'s own write failure at `apply.go:77` already is. A failed
`persist` costs the next run a rebuild of a derived cache. Both are `//nolint:errcheck`
with those reasons on the line.

**Item 3 was the one real fix.** `cmd.Wait()` is now checked, and only when `drainBlobs`
succeeded: `if werr := cmd.Wait(); err == nil { err = werr }`. That is exactly the narrow
case the 2026-08-22 re-trace identified — a full transcript followed by a non-zero exit —
and the guard is load-bearing, because on a truncated stream `Wait` reports a broken pipe
that would bury the read error the caller actually needs. `:55`'s `w.WriteString` stays
ignored: `feedRequests` runs in a goroutine with nowhere to report, and its failure ends
git's output, which `drainBlobs` sees as a short read.

**Everything else was mechanical.** Nine production ignores now carry a reason in the four
precedents' style (`//nolint:<linter> // <lowercase reason>`, no trailing period). The 26
test findings split three ways by what the failure would cost: setup writes whose failure
would make the assertions below them meaningless are now checked and `t.Fatal` (a `seed`
helper in `pkg/sentinel/sentinel_test.go`, the paired `Append`/`AppendD2`/`AppendD4` calls
whose whole assertion is "two lines survived"); `httptest` handler writes are `_, _ =` with
one explanation on `testServer`, because `t.Fatal` is illegal off the test goroutine and
the client-side error is what each test already asserts; the rest are `_ =` at the line.
No blanket `issues:` exclusion was added — every ignore in this tree names its own reason.

**Not re-measured:** `forge drift`'s <100ms hook-path budget. `catfile.go` is the only
`pkg/codeindex`/`pkg/drift` change with any runtime effect and it adds no work — `cmd.Wait()`
was already called at the same point, unconditionally; only its return value is now read.
The three repos the vault's cached indexes name (`food`, `leprecoin`, `meter`) are not
present on this machine at any locatable path, so an end-to-end timing run was not possible.
`CGO_ENABLED=1 go test ./pkg/codeindex/... ./pkg/drift/...` passes, which exercises
`streamBlobs` but is not a latency measurement.

---

## B-030 — `dataset.capture` accepts five tiers; only two of them gate anything

**Owner: Phase 6b. Status: closed 2026-08-22.** Found 2026-08-21 while scoping B-024, out
of that fix's scope. Closed with the *first* of the three shapes below — the only one that
makes the control real — plus a fourth gap the entry did not know about. See the closure
note at the end of this item. The original brief follows.

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

### Closure note — 2026-08-22, Phase 6b

Closed by making the control real, not by documenting it. `cmd/forge/capture.go` now calls
`captureConsented` before harvesting, so removing `d3` from `dataset.capture` stops D3
capture. The hook contract shaped every branch of that function rather than the other way
round: `hooks/vault-post-commit` binds this command to *never fail a commit* and *never
print to the terminal*, so the gate returns exit 0 on every path and speaks only on stderr,
which the hook's own redirect turns into `.forge/capture.log`.

One decision inside it is worth naming. **An unreadable config skips capture rather than
proceeding.** Fail-open is the wrong default for a consent check — the capture list is how
a user says no, and a config that will not parse is not a yes. The commit still succeeds.

D1 and D5 stopped being inert in the same phase, so all five tags now gate a real write
path and the entry's "the list is ahead of the code" framing no longer applies to any of
them. `TestPackagedCaptureListGates` widened from two tiers to five accordingly, and still
asserts against the packaged config layer rather than a hand-written list — that pairing is
the whole reason B-024's guard exists and widening it must not quietly drop it.

**The entry missed a fourth gap, and it was the same defect one level up.** Neither reader
it names checked `dataset.enabled` at all, so `{enabled: false, capture: [d2]}` captured
anyway: a master switch that switched nothing. The fix moved the gate into
`Tier.Enabled(config.Dataset)`, which takes the whole struct rather than the bare list
precisely so a call site cannot forget the outer switch. Packaged config sets
`enabled: true`, so no default behaviour changed.

The entry's "do not fix this by renaming anything" instruction was honoured. `Enabled` and
`D4Enabled` were removed, but what replaced them is a per-tier gate mechanism
(`pkg/dataset/tier.go`), not a rename: `Enabled` read as general while hardcoding `D2Tag`,
which is the trap a third tier would have fallen into, and `d4.go`'s own comment said
`D4Enabled` existed only because of that.

---

## B-031 — the Kafka/Testcontainers miss is a coverage defect, not a precision one

**Owner: unassigned. Status: closed 2026-08-24, `worktree-b-008-recall-recalibration` —
out-of-phase work, not a phase. Recorded 2026-08-22 while closing B-008.**

B-008 carried two cases in one item: a false positive that had to fall, and a miss that had
to rise. Absent-term admission is **strictly decreasing** for every positive weight, so one
knob cannot move them in opposite directions. Measured on `examples/vault`:

| | before | after |
|---|---|---|
| "Kafka consumers with Testcontainers" → `testcontainers-docker-based-integration-testing` | 0.501, CREATE | **0.311, CREATE** |

The note is the right one. Its frontmatter is `stack: [testcontainers, spring-boot, docker,
java]`, `tags: [testing]` — and `kafka`, the term carrying the question, appears in its
title, tags and stack **not at all**. There is nothing for a scoring change to find. Pushing
this case up with the same knob that pushes case 1 down would be tuning, which B-008 forbids
by name, which is why it was split rather than solved.

Two shapes, neither chosen. **Fix the corpus:** the note is under-curated and a `kafka` tag
would be honest — but a fix that edits the vault to make a query score is not a recall fix,
and it does not generalise. **Fix the coverage signal:** the body channel is the only one
that sees `kafka` here, and it carries 0.1 of the blend and only runs for the top 20
candidates (`recall.BodyPassSize`). Whether a term the body carries strongly and the
frontmatter carries nowhere should lift a candidate is a real open question and the more
interesting one — but it re-opens DESIGN §8's weight ratios, so it needs its own session and
its own argument, not a coefficient nudged until this row passes.

Do not respond to this by moving the thresholds either. The same argument B-008 makes
applies: 0.311 and 0.315 sit next to notes that should not be admitted.

### Closed 2026-08-24: neither shape survives contact with the corpus

TODO.md's own step 1 said write the choice down before coding. Doing that first surfaced
that **shape 1 was mischaracterized** and **shape 2 was measured, not assumed** — and both
findings point the same way.

**Shape 1 is not "under-curated," it's wrong.** `grep -i kafka` against
`testcontainers-docker-based-integration-testing.md` — title, frontmatter, and body — comes
back empty. So does `consumer`. The note does not mention Kafka anywhere, not just in its
frontmatter. Tagging it `kafka` would not be honest curation of an under-tagged note; it
would assert the note covers something it does not discuss. Shape 1 is off the table for a
different reason than "does not generalise."

**Shape 2's own premise — "the body channel is the only one that sees `kafka` here" — is
true of the *system*, not of this row, and that distinction is what makes the fix a
no-op for this query.** `kafka` does appear heavily in the corpus: `cqrs-and-event-driven-
messaging.md` (44 hits) and `transactional-outbox-pattern.md` (28 hits), both
`notes/howto/`. Neither ever reached a body pass, for a structural reason worth recording
on its own: `bodyPass` opens `cands[:20]` after sorting by frontmatter score then **path
ascending** (`rank.go`'s `sortByScore`), and on this query ~84 of 92 docs tie at 0.000
frontmatter — `notes/concept/*` sorts before `notes/howto/*`, so the tie-break alone fills
the 20-wide window with `concept/` notes before any kafka-heavy `howto/` note is reached.

Measured directly (`BodyPassSize` bumped 20 → 200 locally, rebuilt, re-run against
`examples/vault`, reverted — not committed):

| Note | Body match | Score | ≥ 0.150 floor? | ≥ winner's 0.311? |
|---|---|---|---|---|
| `testcontainers-docker-based-integration-testing` (today's winner) | `testcontainers` only | 0.311 | — | — |
| `transactional-outbox-pattern` | `kafka`, `consumers` (no `testcontainers`) | 0.111 | no | no |
| `cqrs-and-event-driven-messaging` | `kafka`, `consumers` (no `testcontainers`) | 0.089 | no | no |

Both kafka-bearing notes score *below the neighbour floor* even with the `BodyPassSize` cap
removed entirely — the most permissive version of shape 2 imaginable. They have zero
title/tags/stack signal (they are architecture notes that use Kafka as an example service,
not testing-infrastructure notes), so `wBody = 0.1` cannot lift them past a winner that
carries partial frontmatter signal too. This is TODO.md step 2's "measure which one is
binding" answered directly: **neither `wBody` nor `BodyPassSize` is binding for this row** —
raising either would not change the top-1, the verdict, or the neighbour set, so there is no
DESIGN §8 argument to build here against this row.

**Conclusion for this row: today's behavior is correct, not a miss.** `testcontainers-
docker-based-integration-testing` — a testing-infrastructure note — is the closer topical
match to "Kafka consumers with Testcontainers" than an architecture note that mentions
Kafka in passing; CREATE with four testcontainers-family neighbours (populated since
B-033) is the right answer for a question the vault has no dedicated note for. No code
change, no corpus edit, no golden diff. `go test ./cmd/forge -run TestCalibration` was
re-run unmodified and still passes.

**This closes the row, not the general question.** TODO.md's own "done when" offered two
outs: a corpus fix, or "the body-channel question is answered with an argument that stands
without reference to this one query." Neither happened here — this entry only establishes
that neither shape moves *this* row. The general question the row surfaced — should a term
the body carries strongly and the frontmatter carries nowhere ever lift a candidate, when
the reason it doesn't today is which of the seven type directories the note happens to
live in — stands on its own and generalizes past Kafka: `transactional-outbox-pattern`
(kafka + consumers in body, zero frontmatter) is invisible to `bodyPass` for a reason that
has nothing to do with `wBody`'s value. Filed separately as **B-038**, because it is a
defect in the ranker, not a coverage gap in one note, and closing B-031 must not be read
as having settled it.

---

## B-032 — an untagged note escapes B-008's absent-term penalty entirely

**Owner: unassigned. Status: closed 2026-08-23, `worktree-b-008-recall-recalibration` —
out-of-phase work, not a phase. Recorded 2026-08-22 while closing B-008.**

Two rules interact in a way neither anticipates alone. §2.5's activation is two-sided: a
note with no `tags:` leaves the tags channel *inactive*, dropping out of the blend's
denominator rather than scoring zero — decided from measurement, because zeroing untagged
notes ranked a correct under-curated note below a well-tagged irrelevant one. B-008's
admission then charges the tags channel for query terms the vault tags nowhere. A **tagged**
note pays that charge in full; an **untagged** note never sees it.

Measured, `examples/vault`, "Redis caching in Spring Boot":

```
meterreadingsservice-spring-boot-4-x-project   0.500   tags: []        <- tags inactive
spring-cli-and-maven-commands-for-spring-boot  0.415   tags: [spring-cli]
```

So the note that wins the row that motivated B-008 wins it partly by carrying no tags. That
is §2.5's own effect running in the opposite direction from the one it was written to
prevent. 9 of `examples/vault`'s 91 notes are untagged; CLAUDE.md records 31 of 91 in the
live vault as missing `tags:` or `stack:` after the Phase 1 migration, so the exposure is
larger there than the example corpus suggests.

**It did not change B-008's verdict** — 0.500 is CREATE, and the note that had to stop
winning did stop winning — which is why this is filed rather than treated as a regression.

The tension is genuine and both halves are argued from measurement, so do not "fix" it by
deleting either rule. The shape worth exploring: activation currently asks whether the note
carries the field at all; it could instead ask whether the note carries the field *and* the
query has something the field could answer, so that an untagged note is neither penalised
nor advantaged. That changes `blend`'s denominator for a large fraction of the vault and
must be measured against `cmd/forge/testdata/calibration.golden`, which now exists.

### Closed 2026-08-23 — activation moved from field-presence to hit-presence

**Two candidate fixes were computed by hand against the cited row before either was
written, and one of them is disqualified by the row itself.** Dropping the note-side
conjunct entirely (`c.Active = ok`, letting any note with tags evaluate at 0.000 when
nothing overlaps) does fix the row — `meterreadingsservice` drops to 0.350, below
`spring-cli`'s unchanged 0.415 — but it does so by deleting §2.5's untagged-exemption rule
outright, which this entry and `references/recall-spec.md` both name as not-to-be-deleted,
and it independently fails the row check: 0.415 becomes first place, which is exactly
B-008's false positive returning. That path was rejected on both grounds, not just one.

**The shipped fix:** `tagsChannel` and `stackChannel` activate on `len(hits) > 0` rather
than `len(tags) > 0` / `len(stack) > 0` (`pkg/recall/score.go`). A note whose tags don't
overlap the query is now inactive exactly like a note with no tags — parity, not a new
exemption, and the note-side rule is preserved rather than deleted. The principle is
already in the codebase: `weighted`'s own comment argues that scoring 0.0 on an active
channel "drags every note down uniformly" for a query outside the vault's vocabulary
entirely, and rejects it for that corpus-wide case. This fix is the same argument applied
per note rather than per corpus.

**The cited row does not move.** `meterreadingsservice-spring-boot-4-x-project` stays
0.500 (still untagged, still inactive on tags); `spring-cli-and-maven-commands-for-spring-
boot` stays 0.415 (its one real hit, `spring`, is unaffected). This matches the entry's own
observation that the row "is not a regression" — B-032's defect is the *mechanism*, not
this specific ranking, and the fix targets the mechanism.

**What does move, measured:** across the nine calibration queries, active tags/stack
channels drop from 128 to 84, and all 43 that scored a hard 0.000 are eliminated. One
calibration winner changes: the Docker query's top-1 moves from `docker-compose-init-
container-pattern…` (0.163, one real `docker` tag hit) to `docker-compose-local-yaml…`
(0.170, no relevant tags). That is the same trade recurring one level down — a channel
capped low by vault-wide term absence (three of six query terms have df 0) still drags an
active note below where exclusion would leave it — and it is `weighted`'s documented
behavior, not a defect this fix introduces. No verdict crosses a threshold anywhere in the
nine-query set; every row stays CREATE. `cmd/forge/testdata/calibration.golden` carries
the full diff.

**B-033's two derivations were re-run against the new scale, per this entry's own
prerequisite, not skipped.** The neighbour floor's F1 peak moved off 0.125 — the new
maximum is 0.150 (F1 0.578), re-derived against the same unedited
`neighbour-labels.txt`. `pkg/recall/doc.go` and `config/forge.config.example.md` moved
together. `forge intent`'s gate derivation still holds *mechanically* — gate stays 0.50,
still admits 8 of 10 FIRE prompts, still admits zero QUIET — but the measured FIRE/QUIET
separation margin went from +0.005 to **-0.036**: the lowest gate-admitted FIRE prompt and
the highest QUIET prompt now overlap between 0.407 and 0.443. The gate sits above both, so
nothing is broken today, but the classes are no longer separable everywhere, which is a
finding rather than something this fix corrects — filed as **B-037**, not nudged.

Verified: `go test ./pkg/recall/...` and `go test ./cmd/forge -run TestCalibration` green
against the re-recorded goldens; `idf(0, n) == 0` unchanged; no existing test in either
package pinned the old tagged-but-zero-overlap behavior, so nothing was inverted to make
this pass. `go test ./...` green on both build lanes.

---

## B-033 — the 0.30 neighbour floor was calibrated against the pre-B-008 scale

**Owner: unassigned. Status: open — recorded 2026-08-22 while closing B-008.**

B-008 changed the scale of two of the four channels; the neighbour band's floor did not
move with it. DESIGN §5.3's band exists to answer "what should this new note link to" on a
CREATE verdict, and on adjacent-topic queries it now answers "nothing".

Measured, `examples/vault`, "Storybook interaction testing with play functions":

| | verdict | top-3 | neighbours |
|---|---|---|---|
| before | UPDATE(extend) 0.617 | 0.617 / 0.601 / 0.540 | 0 (none emitted on UPDATE, by design) |
| after | CREATE 0.217 | 0.217 / 0.201 / 0.160 | **0 — both Storybook notes are under 0.30** |

So the new note would be written unlinked to the two Storybook notes obviously related to
it. That is an orphan-creation path in a vault whose own graph report already tracks 21
orphans of 94, and it is the concrete residual cost of B-008's fix.

**Why it was not fixed in the same pass.** Re-deriving the floor against the same nine
queries used to validate B-008 is circular — the number would be chosen to make those rows
produce links. An honest re-derivation needs its own query set, and specifically one where
the right neighbour set is known independently of what the scorer says.

**A second instance, same root cause, found while closing B-008.** `cmd/forge/intent.go`'s
`printIntent` gates on a hardcoded `0.7` — chosen when the false positive it guards against
read 0.740. That note now reads 0.415, so the gate is stricter in effect than it was written
to be, and a legitimate hit like "how does keyset pagination work" clears it at 0.729 with
little room. Its doc comment now says so. Re-derive both numbers in the same session; they
are one question, not two.

The harness is the reason this is cheap now: `cmd/forge/calibration_test.go` and its golden
already stage the corpus and diff a table, and the neighbour column is an addition to
`calibrationRow`, not new machinery. **The answer/update thresholds still do not move** —
B-008 forbids that and this entry does not reopen it. `neighbour_min_score` is a different
knob, named nowhere in that prohibition, and it is the only one in scope here.

### Closed 2026-08-23 — floor 0.30 → 0.125, and the intent gate 0.7 → `Update`

**The floor is 0.125.** Derived, not tuned: `cmd/forge/testdata/neighbour-labels.txt` is
fifteen adjacent-topic queries with 58 expected neighbours, written from the corpus file
list **before any score was measured** and committed one commit ahead of the sweep that
reads them, so the ordering is checkable in git rather than asserted here.
`TestNeighbourFloorSweep` records `testdata/neighbour-sweep.golden`.

| Floor | Precision | Recall | F1 | Median links/query | Queries with none |
|---|---|---|---|---|---|
| 0.100 | 0.478 | 0.741 | 0.581 | 6 | 0 |
| **0.125** | **0.548** | **0.690** | **0.611** | **4** | **0** |
| 0.150 | 0.600 | 0.569 | 0.584 | 3 | 0 |
| 0.175 | 0.643 | 0.466 | 0.540 | 2 | 0 |
| 0.300 (old) | 0.900 | 0.155 | 0.265 | 1 | **6** |

The old floor's numbers are the finding: precision 0.900 looks excellent and is an
artifact of emitting ten links across fifteen queries, six of which got none at all.

Three things chose 0.125, in order. **It is the only swept value that fixes the case this
entry was opened for** — the five Storybook notes cluster at 0.131–0.323, and every floor
at 0.150 and up emits the top one and nothing else, no better than the status quo on that
row. **F1 peaks there** (0.611), within 0.011 of F₂'s peak, and F₂ is the right weighting
given the cost asymmetry: a missed neighbour orphans a note silently, a spurious one is a
wrong wikilink a reviewer deletes. §2.2 already set the precedent of preferring F₂ in this
scorer. **Volume stays reviewable:** median 4 links per query, none empty until 0.200.

Verified end to end — `forge recall "Storybook interaction testing with play functions"`
against `examples/vault` now emits seven neighbours, five of them the Storybook family,
where before it emitted zero.

**Precision 0.548 is a lower bound and was deliberately not improved.** Several counted
false positives are defensible links the labels did not name (`food-ordering-system-course-
architecture-index` on the saga and DDD queries). Re-labelling after seeing the scores is
how a derivation becomes a fit, so the labels were left exactly as written.

**Two of this entry's own assumptions did not survive.** The "before" state is not
uniformly zero neighbours — three of §3.1's nine emitted none, the other six emitted one
to five; the entry generalised from the Storybook row. And `pkg/recall`'s
`TestNeighbourBandEdges` spelled `0.30` into its fixture, so it failed on the change as if
the band had broken; it now expresses both edges in terms of `DefaultThresholds`, because
what belongs in a unit test is which side of an edge is included, not what number the edge
sits at.

**The intent gate got the opposite answer, and that is the point.** This entry said the
two numbers "are one question, not two." True of the root cause, false of the answer:
`printIntent` interrupts a live session on a hook contracted never to disturb it, so it is
derived for precision first, and for recall only inside what precision leaves free.
`testdata/intent-gate-labels.txt` labels 25 prompts FIRE/QUIET (ten of the QUIET ones
adjacent-topic hard negatives) and `TestIntentGateSeparation` pins both directions.
Measured: **0.7 admitted 3 of 10 FIRE prompts**, dropping one at 0.652 that matches a note
title almost verbatim. But the classes separate at 0.402/0.407 — a **0.005** margin — so
every value from 0.405 to 0.7 is false-positive-free on that set and *the labels cannot
choose the replacement*. The gate is **0.50**: the lowest value still a clear step above
the QUIET ceiling (~24% headroom) that admits every FIRE prompt whose phrasing tracks a
note title, recovering three the old gate dropped at 0.546, 0.533 and 0.517. 8 of 10, and
`minFireAdmitted` is set to exactly 8 rather than to something comfortably below it — a
tripwire with slack would have let 0.7's silent decay happen again, more slowly.

**`DefaultThresholds.Update` was tried first and rejected, which is worth recording
because it is the more elegant-looking answer.** The argument for it was that below Update
the verdict is CREATE, so `emitIntentHit`'s "may already answer this" would contradict the
scorer. Checked against the code, that argument is false: `printIntent` computes no
verdict — it reads `cands[0].Score` — and the message hedges with "may". Binding to Update
would have cost 3 of 10 FIRE prompts, all near-verbatim title matches, for an alignment
nothing in the function asserts, and coupled hook behaviour to a config key this path
never reads. It is the same failure as `TestNeighbourBandEdges` above, inverted: that test
pinned a *number* while claiming to test a *rule*; this would have pinned the gate to a
*rule* that does not govern it. It stays a plain constant — `printIntent` runs under a
50ms budget and loads no config, and that half of the argument does survive.

**`DESIGN:257` still says "0.3–0.55" and was deliberately not edited** — a decision
superseded by a later ruling is exactly what AUDIT §8.4 governs. The ruling is **D-9**,
the first §8.4 entry with no C-number. `config/presets/` restates neither threshold, so
the two default sites (`pkg/recall/doc.go`, `config/forge.config.example.md`) are the whole
surface. Spec §3, §3.1 and §3.2 were updated; §4's `"neighbours": []` example is on an
ANSWER verdict and is still correct.

**Opened in its place: B-036** — §3.1's broadest queries now emit ten neighbours, because
two general Spring notes score on every Spring question. No floor separates them.

---

## B-034 — D6 (code↔knowledge) is specified but not built

**Owner: unassigned. Status: closed 2026-08-25 — see the closure note below.**

`ADDENDUM §D.1`'s table lists **six** datasets. Phase 6b built five. The sixth, D6
"Code↔knowledge" — pairs of (repo symbol or module → the note explaining it), volume
"= note count", intended use "retrieval / RAG eval" — was scoped out by explicit decision
and this entry is the record of that decision, not an oversight.

**Why five and not six.** Every other source agrees on five: `docs/ROADMAP.md` says five,
and both phase prompts in `docs/CLAUDE-CODE-PROMPT.md` say five. Only §D.1's own table says
six, and `docs/AUDIT.md` never flagged the disagreement — it is not among the thirteen
contradictions §8.1 catalogues, so precedence gives no ruling and there is no §8.4 decision
to follow. Five was chosen because it is what four of the five sources say.

**Why it is genuinely different from the other five.** D1–D5 are *capture* tiers: each has
a write path on a live command, and the data only accumulates forward, which is the whole
argument for building capture early. D6 has no capture path and needs none. It is a
**derivation** over state that already exists — `forge logback` (Phase 5b) already builds
`docs/knowledge-map.md` from `pkg/coderef`'s citations plus `.forge/code-index-<repo>.json`,
which is exactly the (symbol → note) mapping D6 wants. Nothing is lost by deriving it later;
nothing accumulates in the meantime that a late start would miss.

**Shape when someone picks it up.** An export *view* over `pkg/logback`'s map, not a sixth
`.forge/datasets/d6.jsonl`. Concretely: a `--set d6` case in `pkg/dataset/export.go` whose
`loadTier` reads the code index and citation registry instead of a JSONL file. That means
`Tier` gains a tier with no `Path`, which the current struct does not model — the one real
design question in the item.

Two consequences to keep in view. `--since` has no meaning for a derived set (there is no
per-record timestamp), so it should be refused rather than silently ignored. And anonymizing
a D6 pair is a harder problem than anonymizing any of D1–D5: the *symbol and module names
are the feature*, and they are also the most employer-identifying strings in the whole
system. `pkg/dataset/anonymize.go`'s current answer for note paths — hash the slug, keep the
type — has no equivalent that leaves D6 useful. Do not assume the existing scrubber covers
it.

### Closed 2026-08-25, out-of-phase, on `feat/b-035-run-id` — built to TODO.md's plan; every step landed, and the five-vs-six question resolved the other way from how this entry framed it.

**§D.1's "six" was right all along — nothing is superseded, no AUDIT §8.4 entry is owed.**
This entry's own "why five and not six" section reasoned from ROADMAP and the phase
prompts outvoting §D.1 three sources to one. That was the wrong frame: those three are a
phase-planning snapshot and historical execution prompts, not living specs, and §D.1 is
the one that was actually correct about the dataset shape from the start. `docs/ROADMAP.md`
and `docs/CLAUDE-CODE-PROMPT.md` were deliberately **not edited** for the same reason B-027
left `DESIGN:257` alone before D-9 superseded it — they are dated snapshots of what a past
phase planned and asked for, not doc sites this closure updates. `docs/BACKLOG.md`,
`docs/TODO.md`, `CLAUDE.md`, and the two dataset skills are the live surfaces, and all four
are updated here.

**Shipped as this entry's own plan specified**, with one structural decision the plan
flagged but didn't resolve: `Tier` gained `Derived bool` (not a loader-function field —
`Path` stays meaningful for D1–D5 and empty for D6, which is what `loadTier`'s existing
type-switch pattern wanted) and **D6 is a full member of `Tiers()`**, not a side list —
`resolve()`'s `--set` matching, `dataset-stats`' iteration, and the tag-uniqueness test all
see it for free. The cost this decision has: `forge dataset-stats` now re-derives D6 (glob
`.forge/code-index-*.json`, walk every note, resolve every citation) on **every** call, not
just on export — previously five file reads. Not measured against a real multi-repo vault
on this machine, per B-029's and B-015's precedent for repos this machine doesn't have;
`writeTierCounts`' header wording was changed to say so rather than claiming a uniform
"captured pairs" table.

**The derivation itself needs no live repo access and no new `--repo` flag anywhere.**
`.forge/code-index-<repo>.json` already carries a repo's full file list and symbol table
(`codeindex.Index.Files`), so `pkg/dataset/d6.go`'s `loadD6` builds a `coderef.Registry`
straight from the cached indexes and resolves each note's citations against it — the same
`locate()` branch (`cmd/forge/check_codebase.go`) `forge check`'s own coverage numbers use,
mirrored rather than loosened, so a citation `forge check` already treats as unresolved
contributes no D6 pair either. Symbol lookup uses `codeindex.File.Lookup` (exact name, then
trailing member) over each index's files in sorted order — deliberately not
`codeindex.Index.FindSymbol`, which ranges over a Go map and would pick a different file on
different runs when two files in one repo declare the same trailing member; sorted iteration
is BACKLOG B-020's determinism rule applied one layer further out.

**Fail-closed on a cache Load can't read, not silently smaller.** `loadIndexes` treats an
absent `.forge/code-index-*.json` as "nobody has run `forge logback` here yet" (not an
error) but a *present and unreadable* one (stale `Extractor`, corrupt JSON) as a hard
failure naming the file — the same reasoning `pkg/dataset/read.go` already applied to a
torn capture line, carried to a cache instead of a log. Silently continuing over the
unreadable file was the design this closure's own first advisor pass caught before it
shipped: it would have reported export success over a corpus with an unannounced hole.

**Anonymization: refused, not attempted**, exactly as this entry's own "shape when someone
picks it up" section concluded no acceptable answer exists. `refuseDerivedOptions`
(`pkg/dataset/export.go`) rejects both `--since` and `--anonymize` on any `Tier.Derived`
before a record is read — general on the field rather than hardcoded to `d6`, so a future
derived tier inherits both refusals — and since `dataset.anonymize_on_export` defaults
`true` in the packaged config, exporting `d6` at all requires `--no-anonymize` explicitly.
Both refusals are `UsageError` (exit 2, `--out` untouched, no `exports.jsonl` line), pinned
by `pkg/dataset/d6_test.go` and, at the CLI layer, `cmd/forge/export_dataset_test.go`.

**Two "do not" items from the plan were both honoured.** No sixth capture path — there is
no `AppendD6`, on purpose. No `d6` entry in `config/forge.config.example.md`'s
`dataset.capture` list — `Tier.Enabled` is never called for a derived tier by any real call
site, and `TestPackagedCaptureListGates` now skips `Derived` tiers explicitly rather than
being satisfied by accident.

**Records ≠ notes**, stated in the D6 datasheet and in both dataset skills: a note citing
five symbols yields five D6 records, so its count answers "how many citations resolve",
not §D.1's literal "= note count" line.

Verified: both build lanes (`CGO_ENABLED=0` and `=1`) green, `go vet` clean, new tests in
`pkg/dataset` (positive derivation over a real fixture code index + note, dedup, both
refusals, fail-closed on a stale-Extractor cache) and `cmd/forge` (exit-code-2 coverage for
both refusals). Nothing here touches `pkg/recall`'s scoring or `pkg/report`'s codebase
rendering — `TestCalibration` and the neighbour/intent-gate goldens are unchanged, no
`-update` run.

---

## B-035 — D1 has no outcome label, because nothing correlates a recall call to what followed

**Owner: unassigned. Status: closed 2026-08-25 — see the closure note below.**

`ADDENDUM §D.1` describes D1's pair as "question → `ANSWER`/`UPDATE`/`CREATE` + topic +
stack", sourced from "every run, **auto-labelled by recall + outcome**". Phase 6b built the
first half. There is no outcome.

What ships is `(question features → the routing decision)`: `pkg/dataset/d1.go`'s `D1Pair`
carries the hash, topic, stack, top score, candidate count and the verdict `forge recall`
returned. Nothing records whether that verdict turned out to be right — whether the
`ANSWER_FROM_VAULT` the user got was actually the note they wanted, whether the `CREATE_NEW`
produced a note that duplicated an existing one.

**The blocker is structural, not effort.** There is no correlation key anywhere in the
system. `forge recall` prints JSON and exits; the note write that may follow happens minutes
later through `forge gate` and `forge-librarian`, in a different process, with nothing
linking the two. `telemetry.Event` has no run id either, so the ask log cannot be joined
back to a note write for the same purpose. Adding an outcome field to `D1Pair` without a key
would just be a column nothing can ever populate.

**The consequence, stated plainly in every D1 datasheet:** the corpus is supervision on the
router's own output. A model trained on it learns to reproduce `pkg/recall`'s decisions,
including its mistakes — which is a legitimate distillation target (a 20ms local classifier
replacing a scoring pass) and is *not* evidence the routing rule is correct. Anyone using D1
to evaluate recall quality is measuring agreement with the thing being evaluated.

**Shape when someone picks it up.** A `run_id` minted in `runRecall`, emitted in the JSON
envelope, carried on both the telemetry event and the D1 pair, and accepted by `forge gate`
as an optional flag so the write path can stamp it. That last hop is the one that decides
the item's real size: `forge gate` is invoked by `skills/forge/SKILL.md`, so the id has to
survive a skill hand-off, and a skill that forgets to pass it degrades to today's behaviour
rather than failing — which is the right degradation, but means the join will be partial and
the datasheet will have to say so.

Related: **B-032** is the other place D1's features are thinner than they look (an untagged
note escapes the absent-term penalty), and **B-031** is the coverage side of the same
scoring surface.

### Closed 2026-08-25, out-of-phase, on `feat/b-035-run-id` — built to TODO.md's plan as
written; every step landed, nothing dropped.

`telemetry.NewRunID` mints the key (16 random bytes, hex — no counter, no timestamp
semantics), called once in `runRecall`. It rides on three surfaces: `telemetry.Event.RunID`
(new, `omitempty`, additive to DESIGN §14's schema), `D1Pair.RunID` (new, `omitempty`), and
the JSON envelope `forge recall` prints — `recall.Result` itself stays untouched (a
zero-model-call scoring package has no business knowing about dataset capture), so
`cmd/forge/recall.go` wraps it in a local `recallEnvelope{recall.Result; RunID}` purely at
the emit layer. `forge gate` takes an optional `--run-id`; when set, and only when D1's own
tier is enabled — an outcome with no corresponding pair is dead weight no export can ever
join — `captureD1Outcome` appends a `D1Outcome{RunID, Published}` record to a **second
file**, `.forge/datasets/d1-outcomes.jsonl`, not a rewrite of the D1 line: that line is
already on disk and immutable by the time the gate call happens, sometimes minutes later in
a different process. `Published` is `true` on the accept branch and `false` on the
quarantine branch — both are join-worthy outcomes, not just the success case.

**The join happens at export, not at capture**, per the plan's own steps 4 and 6.
`pkg/dataset/export_records.go`'s `loadD1` reads both files and matches by `RunID` before
boxing, so `since`/`anonymizeAll`/`roundTripAll` all see the already-joined shape and needed
no changes. The join result lives on `D1Pair` itself as an export-time-only `Outcome *bool`
field (`omitempty`, never written by `AppendD1`) — a pointer so "never joined" (nil) and
"joined, quarantined" (non-nil `false`) stay distinguishable, which is what let the SFT and
CSV renderers add one field/column each (`outcome`, or `published`/`quarantined`/empty in
CSV) without inventing a second record shape. `ExportReport.D1Joined` and the D1 datasheet's
first Limitations bullet now state the actual join fraction (`N%, M of K records`) rather
than repeating "no outcome label" verbatim — the plan's own instruction not to imply a
census.

**`--run-id` is optional everywhere and the degradation is the tested case, not an
afterthought**: `TestCaptureD1OutcomeSkipsOnEmptyRunID` pins that an empty id writes
nothing, and `TestRecallEmitsRunID` pins that two calls never collide. `skills/forge/
SKILL.md` was updated at both hops — Stage 1 tells the caller to hold onto `run_id`, Stage 4
shows `--run-id` on the gate invocation — which is TODO.md step 5, "thread it through the
skill", done as a doc edit since the flag itself is the mechanism; nothing in Go enforces a
skill actually passes it, same posture as D5's gate-is-a-skill-invariant note already on
file. `references/recall-spec.md` §4 got the new field in its worked example plus a
paragraph, per step 1's "that is an addition to a documented contract."

Verified: both build lanes (`CGO_ENABLED=0` and `=1`) green, `go vet` clean. New tests:
`pkg/telemetry`'s `TestNewRunIDLength`/`TestNewRunIDIsUnique`; `cmd/forge`'s
`TestRecallEmitsRunID` (envelope, non-colliding) and four `captureD1Outcome` cases
(published, quarantined, empty-run-id no-op, D1-disabled no-op); `pkg/dataset`'s
`TestD1ExportJoinsOutcomeByRunID`, `TestD1ExportOutcomeInCSV`, and
`TestD1ExportSurvivesAnUnreadableOutcomeFile` (the strict reader refuses a torn outcome
line exactly like every other capture file, naming file and line). `TestCalibration` and
the neighbour/intent-gate goldens are untouched by this item — nothing here reaches
`pkg/recall`'s scoring — confirmed by an unchanged `go test ./...` result, no `-update` run.

---

## B-036 — a broad query links ten neighbours, and no floor can separate them

**Owner: unassigned. Status: open — opened 2026-08-23 while closing B-033.**

Closing B-033 lowered the neighbour floor to 0.125 and the calibration golden now shows
what that costs at the broad end. Three of §3.1's nine queries emit **ten** neighbours —
the maximum `forge recall` returns at all — and the tenth is not a near miss but a cliff:
the list is truncated, so a broader query would emit more.

Measured, `examples/vault`, at floor 0.125:

| Query | Neighbours | Of which are the two general Spring notes |
|---|---|---|
| Redis caching in Spring Boot | 10 | 2 |
| Spring Boot 4 configuration properties binding | 10 | 2 |
| Java virtual threads with Spring Boot | 10 | 2 |
| Storybook interaction testing with play functions | 7 | 0 |
| JPA entity graph to avoid N+1 | 0 | 0 |

`meterreadingsservice-spring-boot-4-x-project` and `spring-cli-and-maven-commands-for-
spring-boot` appear on every Spring question regardless of what it asks. They are broad
notes about a broad ecosystem and they score legitimately; nothing in the ranking is
wrong. **That is why the floor cannot fix it** — admission is a single scalar cut, and
these notes sit above every note the same query genuinely needs. B-033 measured four
narrower floors and each one that removed them also removed the Storybook family.

**Why it matters rather than being cosmetic.** DESIGN §5.3's band feeds a new note's link
list. Ten links in a 91-note vault attaches a note to 11% of the graph, and `pkg/graph`'s
hub and centrality reports are where that shows up — as a hub that is an artifact of the
linking rule, not of the knowledge. The failure is the mirror of the orphan B-033 fixed,
and both come from treating one scalar as the whole answer.

**Shape when someone picks it up.** A cap on neighbour count is the obvious move and the
least interesting one; a cap alone keeps the same ten and truncates arbitrarily, since
they are already score-ordered. The question worth answering first is whether a note that
scores on *every* query in an ecosystem should be admitted as a neighbour at all — which
is a document-frequency property the scorer already computes for terms (§2.3.1) and does
not compute for notes. Measure before building: `TestNeighbourFloorSweep`'s harness
already stages the corpus, and a per-note "appears in N of M query results" column is an
addition to it, not new machinery.

**Do not respond to this by raising the floor.** B-033's sweep is in its closure note and
every floor that drops these two notes also drops the case B-033 was opened to fix.

Related: **B-031** is the coverage side of the same scoring surface, and **B-032** moves
`blend`'s denominator, which will change every number above — re-measure after it lands.

### Measured 2026-08-26, out-of-phase, on `dev` — first pass got the denominator wrong
and reversed the verdict; corrected same day. **Still open — the hypothesis is
confirmed, not rejected.**

TODO.md's own unblock condition was built as specified: a per-note "appears in N of M
query results" column, added as `cmd/forge/neighbour_frequency_test.go`
(`TestNeighbourDocumentFrequency`), run over `testdata/neighbour-labels.txt`'s fifteen
adjacent-topic queries at the shipped floor (0.150). The first read of that table —
`spring-cli-and-maven-commands-for-spring-boot` at 5/15, `meterreadingsservice-spring-
boot-4-x-project` at 4/15, "no note comes close to universal" — was posted as a
rejection and briefly committed as one. **That reading used the wrong denominator.**
This entry's own hypothesis was never "on every query" — it was "on every query in an
**ecosystem**" (TODO.md's unblock condition uses that word for a reason). The fifteen
labelled queries span several unrelated ecosystems (Spring, Keycloak, React, Docker,
Kafka, Liquibase, ...); diluting a Spring-specific count by fourteen unrelated queries
was always going to look small.

**Corrected reading, same golden — and this reading's own subset boundary is disclosed
rather than presented as settled, because picking it *after* seeing which queries the two
notes hit is the same failure shape as the first wrong reading, one layer up.** No
ecosystem label existed before this measurement ran; "which of the fifteen queries count
as Spring" was decided by reading the result, not fixed in advance the way
`neighbour-labels.txt` itself was written before any score existed. Two honest readings,
not one:

- **Narrow — literal word "Spring" in the query text (4 of 15):** `Spring Data JPA
  pagination...`, `Testcontainers reuse ... in Spring Boot`, `Spring Security method
  security...`, `Binding a nested YAML list into a Spring Boot configuration record`.
  `spring-cli-and-maven-commands-for-spring-boot` and `meterreadingsservice-spring-boot-4-
  x-project` both appear in **4 of 4**.
- **Wide — the above plus Maven (this vault's Spring build tool) and Hibernate/JPA
  (Spring Data's persistence layer), 6 of 15:** adds `Maven multi module build with a
  shared parent pom` and `Soft delete with Hibernate entity filters`.
  `spring-cli-and-maven-commands-for-spring-boot` reaches **5 of 6** (misses the Hibernate
  query); `meterreadingsservice-spring-boot-4-x-project` reaches **4 of 6** (misses both
  Maven and Hibernate).

The verdict does not depend on which boundary is right: under either reading both notes
clear a strong majority of an ecosystem-scoped subset while sitting at 33%/27% of the
full fifteen. That is this entry's original claim, holding. What the boundary choice
*does* change is the precise number to quote, which is why neither "4 of 4" nor "5 of 6"
belongs in prose as if it were derived rather than chosen. `docs/TODO.md`'s per-note table
(now carrying the query list per slug, not just the count) is the artifact to read this
from directly.

**A second miscount rode along with the first, in the same commit, and is retracted
outright rather than corrected: there is no B-039.** The initial pass also reported "14
of 15 queries hit `recall.Rank`'s `TopN=10` cap" as evidence of a general window-
saturation mechanism distinct from this entry, and opened B-039 on it. That number
counted `len(ranked) == recall.TopN` — `Rank` returns `nonZero(cands)` truncated to 10,
so this only means ten notes had *some* nonzero overlap with the query, which a tagged
92-note corpus will do almost everywhere; it says nothing about the floor. The number
that actually says the *neighbour* window saturated is `len(Neighbours(ranked)) ==
recall.TopN` — measured (temporarily, not committed) at **2 of 15**, and both of those
two queries are the same literal-"Spring" queries above (`Testcontainers reuse ... in
Spring Boot`, `Binding a nested YAML list into a Spring Boot configuration record`).
There is no general window-saturation mechanism separate from this entry's own finding
— B-039 was built on a metric that didn't measure what its prose claimed, and its real,
corrected signal is fully explained by the two universal notes already named here, not
a second cause. B-039's BACKLOG section and its `docs/TODO.md` entry are removed
outright, not just closed.

**Status: still open. Updated unblock condition: a pre-committed ecosystem label, written
before any design and before re-reading scores — the same discipline
`neighbour-labels.txt` itself got.** The hypothesis is confirmed under both readings above,
so a design pass is now justified — but "which queries count as the same ecosystem as
which notes" needs to be decided and committed to a file *before* it is used to justify a
scoring change, exactly the way this entry's own boundary should have been fixed before
being read off the result. Once that label exists, "design after reading it," per this
entry's original text, is not attempted in this pass regardless: it is a `pkg/recall`
scoring-surface change
(admission would have to interact with `Rank`'s `TopN=10` truncation, since `Neighbours`
only ever sees `Rank`'s already-truncated top 10 — excluding a universal note there
cannot surface an 11th candidate `Rank` never computed), and this session already
produced one wrong verdict under time pressure; writing the fix in the same sitting as
correcting that mistake is how a second one happens. A `docs/superpowers/plans/` PLANNED
write-up for the design is the right next step, not a hurried diff here.

**Do not respond to this by raising the floor.** Unchanged from the original entry — every
floor B-033's sweep tried that drops these two notes also drops the Storybook family
B-033 was opened to fix.

### Closed 2026-08-26, same day as the measurement pass, on `worktree-b-036-neighbour-window`
— out-of-phase work, not a phase.

TODO.md's updated unblock condition was satisfied first, in its own commit, before any
`pkg/recall` file was touched: `cmd/forge/testdata/query-ecosystems.txt` labels all
fifteen `neighbour-labels.txt` queries by a mechanical, documented rule (Spring/Maven/
Hibernate-JPA → `spring`; a single named technology; `unclustered` if nothing matches).
`spring` = 6/15, matching this entry's own already-disclosed "wide" boundary exactly — a
fixed artifact, not re-derived. `git log` shows that commit landing before the scoring
change, the same discipline `neighbour-labels.txt` itself got.

**The fix widens `Rank`'s internal window rather than filtering `Neighbours` more
cleverly, per this entry's own framing.** `Thresholds.Neighbours` only ever saw
`recall.Rank`'s `TopN=10`-truncated list — measured directly against `examples/vault`
before writing any code: all three offending queries return exactly 10 candidates, every
one already clearing the 0.150 floor by 0.06–0.10 margin, confirming the truncation, not
the floor, is what's binding. `pkg/recall/rank.go` gained `RankPool` (the same computation
`Rank` always did, minus the final `TopN` truncation) and a new `NeighbourPool =
truncate(pool, NeighbourWindow=20)` — a separate constant from `BodyPassSize`, even though
equal today, pinned by `TestNeighbourWindowMatchesBodyPassSizeToday` so a future B-038
change to one doesn't silently move the other. `Rank` itself is now `truncate(RankPool(...),
TopN)` — provably byte-identical output to before. `Thresholds.Result` → `ResultFrom(q,
pool)`: `Candidates` still truncates to `TopN` (recall-spec.md §4's contract, unchanged),
`Neighbours` now band-filters the wider `NeighbourPool`.

**A candidate alternative was tried and rejected before the window-widening design was
written, not instead of measuring it.** This entry's own "shape when picked up" section
asked whether a note-level document-frequency signal — extending §2.3.1's term-level IDF
to notes — should exclude a note from the neighbour band outright. Computed by hand over
the real 91-note corpus: it does not separate the two universal notes from labelled-wanted
neighbours. Several notes `neighbour-labels.txt` calls wanted share the identical min-idf
value (1.395) with the two universal notes, and `idfCap` saturates almost every tagged
note in a given ecosystem at the same max value (3.5) regardless of whether it's universal
or specific. This rules out a note-level admission filter as the mechanism; the shipped
fix widens the window instead and accepts the resulting neighbour-volume increase as a
measured, disclosed consequence rather than engineering it away with an unverified filter.

**Golden diffs, checked against a stated prediction before accepting `-update`'s output, not
after** — `calibration.golden`: all 9 §3.1 rows' Top-1 slug / Top score / Verdict stayed
byte-identical; only the 3 previously-10-capped rows' Neighbours column changed, each now
up to 20 entries, the original 10 present as a leading subset, plus genuinely Spring/Java-
ecosystem additions — two of which (`jpaspecificationexecutor-dynamic-queries-with-
specification-pattern`, `keyset-pagination-compound-or-predicate`) are
`neighbour-labels.txt`'s own labelled-wanted neighbours for an adjacent query, direct
evidence the added links are real recall, not noise. `neighbour-sweep.golden`: F1's peak
stayed at floor 0.150 (0.578 → 0.587, still the argmax across every swept floor) — **the
floor did not move**, checked before accepting the regenerated golden rather than after,
exactly the one place this entry's standing "do not raise the floor" rule could have been
violated by accident. `neighbour-frequency.golden`: both predicted effects present, not
just one — the two universal notes' overall frequency rose slightly (they were previously
crowded out of some query's top 10 too) **and** a previously crowded-out specific note
(`preauthorize-spring-security-method-level-access-control`) gained admissions on
`spring`-cluster queries (now 3/6). `TestNeighbourBandEdges`, `TestDecideAtThresholdBoundaries`
and `TestIntentGateSeparation` all pass unmodified, no `-update` — nothing outside the
neighbour band moved. Full suite green under both `CGO_ENABLED=0` and `CGO_ENABLED=1`.

**Binding invariants held**: `Answer`/`Update` thresholds untouched (0.85/0.55), `score.go`
untouched entirely (`idf(0,n)==0` intact), `neighbour-labels.txt` unedited, `candidates`'
at-most-10 contract unchanged (`Rank` provably byte-identical), the 0.150 floor not raised,
`BodyPassSize` not moved.

---

## B-037 — `forge intent`'s FIRE/QUIET margin went negative under B-032's scale

**Owner: unassigned. Status: open — opened 2026-08-23 while closing B-032.**

B-032 moved `blend`'s denominator, which shifted every score in `cmd/forge/testdata/
intent-gate.golden`. `TestIntentGateSeparation`'s two pinned invariants still hold
mechanically — the gate (0.50) admits zero QUIET prompts and at least `minFireAdmitted`
FIRE prompts, both unchanged counts — because the gate sits above the whole overlapping
score band, not inside it. What changed is the *reason* 0.50 is safe: B-033's derivation
argued it as "the lowest value still a clear step above the QUIET ceiling," a margin
argument, and that margin — the lowest gate-admitted-relevant FIRE prompt's score minus the
highest QUIET prompt's score, read literally across the full labelled set rather than just
at the gate — went from +0.005 (pre-B-032) to **-0.036**: lowest FIRE 0.407, highest QUIET
0.443. Somewhere in [0.407, 0.443] a FIRE prompt and a QUIET prompt trade places by score,
so no single threshold separates the two labelled sets everywhere, only above 0.443
specifically. 0.50 clears 0.443 with room (0.057), so nothing is mis-admitted today.

**Do not respond to this by moving the gate.** 0.50 is still measured safe against every
labelled prompt on file; a margin turning negative in a data slice the gate doesn't
actually sit inside is a reason to get more data, not a reason to re-derive a number that
isn't failing.

`docs/TODO.md` named two unblock paths and chose neither at the time: more labelled prompts
targeted at the [0.407, 0.443] overlap band, or a wider sweep of `examples/vault` prompts
the way `neighbour-labels.txt` widened B-033's evidence past nine queries.

### Measured 2026-08-27, out-of-phase, on `worktree-b-037-intent-gate-plan` — the
wide-sweep path, chosen by user decision. **Still open — the -0.036 finding replicates,
unchanged, at twice the sample.**

`docs/TODO.md`'s PLANNED plan for this item covered only the wide-sweep half of the two
named unblock paths. `cmd/forge/testdata/intent-gate-labels.txt` widened 25 → 50 prompts
(10 → 20 FIRE, 15 → 30 QUIET), written from `examples/vault`'s note titles **before any
score on the new batch was measured** — the same discipline `query-ecosystems.txt` used
ahead of B-036's rescoring, in its own commit ahead of the golden regeneration so the
ordering is checkable in git history. The ten new FIRE prompts and fifteen new QUIET
prompts deliberately target nine ecosystems the original 25 under-represented: Keycloak+JWT,
Liquibase, DDD/hexagonal/CQRS, MapStruct, Spring Security beyond `@Value`, Docker
init-container sequencing, Spring Boot 4 breaking changes, Storybook+Next.js, and
Continue.dev config precedence.

**The margin did not move: lowest FIRE 0.407, highest QUIET 0.443, margin -0.036 —
byte-identical to the pre-widening measurement.** None of the ten new FIRE prompts scored
below the original lowest FIRE (`why does @Value not work for a YAML list in Spring Boot`,
0.407); none of the fifteen new QUIET prompts scored above the original highest QUIET
(`Testcontainers reuse across test classes in Spring Boot`, 0.443) — across nine additional
ecosystems and twice the sample. This is the answer the original unblock condition asked
for: it rules out "this is this 25-prompt sample's noise," because the same two prompts
still define both edges of the band at 50. It does not show the overlap *growing* either —
the band's width is stable, not expanding, as the corpus grows.

`minFireAdmitted` was rescaled 8 → 16 (`cmd/forge/intent_gate_test.go`), proportionally, not
loosened: the widened set measures 16/20 FIRE admitted, the same 80% the original 8/10
measured, and 0/30 QUIET admitted — the never-disturb invariant holds exactly, at twice the
sample. Leaving the floor at 8 would have silently tolerated a regression down to 8/20 (40%)
without failing the build, which is the exact failure shape B-033's original derivation
built the floor to catch.

**Verified:** `go test ./cmd/forge -run TestIntentGate` green under the rescaled floor;
`go test ./...` green on both `CGO_ENABLED=0` and `CGO_ENABLED=1`; `pkg/recall` untouched;
`calibration.golden`, `neighbour-sweep.golden` and `neighbour-frequency.golden` untouched
(no `-update` run on any of them); `intentGate` (0.50) untouched.

**This closes the wide-sweep half of the original unblock condition, not the item.** The
finding is now measured at 2x scale rather than assumed to be sample noise, which is new
information — but B-037's own standing rule is unchanged by it: nothing is mis-admitted
today, so this is not a reason to move the gate. The targeted-boundary-sampling alternative
(more labelled prompts specifically inside [0.407, 0.443]) is still untouched and is the
next thing to run if the question is "does a real FIRE/QUIET pair exist inside the band
itself, not just at its edges" — the wide sweep answered "does the band's width hold at
scale," not that.

## B-038 — `bodyPass`'s top-20 window is allocated by path, not by relevance

**Owner: unassigned. Status: open — split out of B-031 on 2026-08-24, deliberately.**

Closing B-031 measured, rather than assumed, whether the 0.1 `wBody` weight or the
`BodyPassSize = 20` cap was binding for the "Kafka consumers with Testcontainers" row —
and found neither was, for a reason with nothing to do with either constant's *value*.
`bodyPass` (`pkg/recall/rank.go`) opens `cands[:20]` after `sortByScore` orders by score
descending, tie-broken by **path ascending**. On this query ~84 of `examples/vault`'s 92
docs tie at 0.000 frontmatter score, so the tie-break alone decides which zero-scoring
notes get a body pass at all — and `notes/concept/*` sorts before `notes/howto/*`
lexically, so a `concept/` note with an incidental one-word body hit gets read before an
`howto/` note that discusses the query's actual subject at length.

Measured directly (`BodyPassSize` bumped 20 → 200 locally, rebuilt, re-run, not
committed): `transactional-outbox-pattern.md` (kafka × ~28, "consumers" present, zero
title/tags/stack signal) and `cqrs-and-event-driven-messaging.md` (kafka × 44, same) both
sit in `notes/howto/` and are never opened at `BodyPassSize = 20`, because the window's
20 slots are consumed by `notes/concept/*` first. This is not specific to Kafka or to this
one query — it reproduces on **any** query where more candidates tie at 0.000 frontmatter
than fit in the window, which the calibration table's other adjacent-topic rows suggest is
common, not rare.

**Why this is a defect and not just a cap that's too small.** A wider window (or no window
at all) would fix it by brute force, but the actual bug is that *which* notes lose the
tie-break is decided by an accident of `pkg/vault`'s directory layout (B-005's seven note
types) rather than by anything about the note's content. A vault that renamed
`notes/concept/` to `notes/aardvark/` would silently change which notes get read, with the
scoring logic untouched.

**Owner note:** for B-031's own row, this defect turns out not to matter — even a
`BodyPassSize` of 200 leaves both kafka-bearing notes below the 0.150 neighbour floor and
below B-031's winning candidate, because they carry zero frontmatter signal. So fixing
this item would not have changed B-031's calibration row. It matters for whatever query
*does* have a genuinely-relevant, frontmatter-silent note sitting past slot 20 in
path order — which this investigation did not go looking for and has not found.

**Shape when someone picks it up — measure before designing.** Two questions, not one:
(a) does raising or removing `BodyPassSize` move `calibration.golden` on the current
nine-query set at all (this entry's own measurement says no, but that is nine queries,
not the corpus); (b) if a tie-break has to exist, should it favor something about the
note (recency, verified date, a document-frequency signal per §2.3.1) over an arbitrary
path sort. `TestCalibration`'s harness already stages the corpus and is the cheapest place
to run (a) at scale before touching (b).

**Do not respond to this by simply raising `BodyPassSize`.** Cost is real and unmeasured
at corpus scale — `bodyPass` opens a file per candidate in the window, and DESIGN §8's
"top 20" was sized for latency, not correctness; widening it moves a performance
constraint before anyone has shown the wider window changes an actual verdict.

Related: split from **B-031**, which established the row that surfaced this but is closed
on its own terms — see its BACKLOG closing section.

---

## B-039 — Windows release targets removed, 2026-08-27, by user decision (`şimdilik`, "for now")

**Owner: unassigned. Status: closed at open — this is a record of a reversible cut, not
a defect.** Out-of-phase, direct user request (not a doc-coherence finding, so no
`AUDIT.md` §8.4 entry — §8.4 is for a decision permanently superseded by a later ruling;
this is a scope cut the user may reverse).

`windows` dropped from the actual build/release surface:
`Makefile:16`'s `PLATFORMS` (was `darwin/amd64 darwin/arm64 linux/amd64 linux/arm64
windows/amd64 windows/arm64`, now the first four only), the now-dead `.exe`-suffix branch
in the `dist` target, and `.goreleaser.yml`'s `goos` list plus its now-unreachable
`format_overrides: [{goos: windows, formats: [zip]}]` block. `make dist` verified: exactly
four binaries in `dist/`, none `.exe`, `checksums.txt` lists four. Two Turkish
user-guide lines describing `make dist`'s live behavior were corrected to match
(`docs/tr/03-KULLANIM-KILAVUZU.md:49`, `docs/tr/01-FIKIR.md:291`) — same shape as B-027
(a doc naming something that isn't there), not B-033's (a superseded decision).

**Deliberately left alone**, per the discriminator this repo already uses for design docs
vs. live behavior: `docs/KNOWLEDGE-FORGE-STACK.md:72` (Go's cross-compile *capability* as
a reason to pick Go — stays true whether an artifact ships) and `:133` (one of three cgo
options considered, already superseded by the shipped portable-only goreleaser config);
`docs/CLAUDE-CODE-PROMPT.md:255` and `STACK.md:323` (the historical phase-6 execution
prompt — B-034's precedent: a dated planning snapshot isn't a living spec). Also left
alone, because it isn't Windows-*support* at all: `pkg/config/load.go:154` and
`pkg/config/chain_test.go:215`'s CRLF/BOM normalization (a markdown config can arrive CRLF
from any editor, not specifically a shipped Windows binary) and `docs/tr/02-MIMARI.md:177`,
the Turkish description of the same. `bin/forge` is POSIX `sh` and never had a Windows
branch to remove.

**To restore Windows as a release target:** re-add `windows/amd64 windows/arm64` to
`Makefile:16`'s `PLATFORMS`, restore the `ext=".exe"` branch and `$$ext` suffix in `dist`,
add `windows` back to `.goreleaser.yml`'s `goos` list, and restore the `format_overrides`
block (`goos: windows` → `formats: [zip]`, since a bare `.exe` in a `.tar.gz` is not the
convention). No other file needs touching.

