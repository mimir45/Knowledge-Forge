# TODO — executable plans for the open backlog

`docs/BACKLOG.md` is the record of *why* each item exists. This file is the record of
*how to close it*. One section per workable item, each written so a future session can
open it cold, follow the steps, and stop when "Done when" is satisfied.

**Read `docs/BACKLOG.md`'s own entry before starting any item here.** These plans are
derived from it and deliberately do not restate its argument. Where a backlog entry has a
`###` closing section, that section is more current than its `Status:` line.

Written 2026-08-23 from a full scan of B-001…B-035; B-036 and B-037 added the same day as
B-033 and B-032 closed; B-038 added 2026-08-24 when B-031 closed and split it out. B-036
was measured 2026-08-26 against its own named unblock condition — **first read wrongly as
a rejection (wrong denominator), corrected same day: the hypothesis was confirmed** — and
closed the same day, on `worktree-b-036-neighbour-window`: `Rank`'s internal window
widened (`NeighbourWindow=20`, separate from `BodyPassSize`) so the neighbour band can see
past the old `TopN=10` cliff. A short-lived B-039, opened on the wrong-denominator pass,
was retracted outright — see BACKLOG.md's B-036 closing note for the full correction and
the closure itself, including why the "no code change" and "B-039" text both briefly
existed here before being reverted. Anchors were verified against the tree at that date;
re-grep before trusting a line number. **B-037 got its first PLANNED section 2026-08-27**
— scoped to only one of its two named unblock paths (widen the labelled corpus, per user
decision), leaving the other (targeted boundary sampling) untouched; see its section below
for what the plan does and does not cover.
**Closed and recorded items are dropped from this file
entirely, not just their plan sections** — `docs/BACKLOG.md` is the full census (every ID,
closed or open, with the reasoning) and the durable record of *why*; this file exists only
to track *how to close* whatever is still open. If an ID isn't below, it's done — look it
up in BACKLOG.

## Two status classes

Not every open backlog item can take a plan, and the absence of steps below is a decision,
not an oversight:

- **PLANNED** — workable, has a full six-field section in this file. **One now: B-037**,
  and only for its wide-sweep half. B-034 (2026-08-25) was the last of the prior eight:
  B-029, B-027, B-033, B-032, B-023, B-031, B-015 and B-035 closed 2026-08-23/24/25 before
  it. B-033's closure opened B-036 and B-032's opened B-037. **B-036 closed 2026-08-26**
  (measured, then designed and shipped the same day — see its BACKLOG entry) without ever
  getting a PLANNED write-up here: the design was scoped and low-risk enough (widen
  `Rank`'s internal window, one new constant, two new functions, three golden
  regenerations) to go straight from measurement to implementation once its
  ecosystem-label prerequisite landed. **B-037 got a PLANNED section 2026-08-27**, but it
  is still measurement, not design — it widens the labelled corpus and re-measures the
  margin; it does not move `intentGate`. B-038 is still NO STEPS by its own argument,
  naming a measurement to run before any design is chosen.
- **NO STEPS** — open but not actionable by an implementation session: blocked on external
  observation, or a user decision, or "record, don't fix" by standing rule. Listed with
  its unblock condition instead of steps.

The execution-order chain this file tracked through B-034 (`B-029 → B-033 → B-032 →
B-023 → B-031 → B-015 → B-035 → B-034 → B-027`) is fully closed and dropped along with
the plan sections below, per this file's own rule — the sequencing rationale (why B-033
had to land before B-032, why B-029 went first) is recorded in each item's BACKLOG
closing note, not restated here.

---

## Index — 6 open items (33 closed/recorded IDs live in BACKLOG only)

| ID | Subject | Class | Section |
|---|---|---|---|
| B-003 | Repo directory still named `TIL` | NO STEPS — user decision | [below](#no-steps) |
| B-004 | Module path has no VCS host prefix | NO STEPS — deferred by decision | [below](#no-steps) |
| B-012 | `code_refs` has no live producer | NO STEPS — blocked on packaging | [below](#no-steps) |
| B-025 | `PostToolUse`/WebFetch payload shape | NO STEPS — **BLOCKED** | [below](#no-steps) |
| B-037 | Intent gate FIRE/QUIET margin now negative | PLANNED — widen labelled corpus (measurement only) | [below](#b-037--widen-the-intent-gate-labelled-corpus-wide-sweep-measurement) |
| B-038 | `bodyPass` window allocated by path, not relevance | NO STEPS — measure first | [below](#b-038--bodypasss-top-20-window-is-allocated-by-path-not-by-relevance) |

---

# B-037 — widen the intent-gate labelled corpus (wide-sweep measurement)

**Why it's open.** After B-032 changed `blend`'s denominator, `forge intent`'s gate (0.50)
still admits zero QUIET prompts and exactly `minFireAdmitted` (8) FIRE prompts — both
unchanged counts — because it sits above the whole FIRE/QUIET score range. Read literally
across the full labelled set, the two classes now overlap: lowest gate-admitted-relevant
FIRE prompt 0.407, highest QUIET prompt 0.443, margin **-0.036** (was +0.005 before B-032).
Nothing is mis-admitted today — the finding is that no single threshold separates the two
classes everywhere on this 25-prompt sample, only above 0.443 specifically. The NO-STEPS
version of this entry named two unblock paths and chose neither: more labelled prompts
targeted at the [0.407, 0.443] overlap band, or a wider sweep of `examples/vault` prompts
the way `neighbour-labels.txt` widened B-033's evidence past nine queries. **The
wider-sweep path was chosen 2026-08-27, by user decision** (not derivable from the code),
mirroring B-036's own proven pattern rather than sampling near a score the current scorer
already produced. This plan covers only that path — the targeted-boundary-sampling
alternative is untouched and stays available if this measurement doesn't settle the
question.

**Anchors.**

- `cmd/forge/testdata/intent-gate-labels.txt` — the 25-prompt (10 FIRE / 15 QUIET) ground
  truth. Its own header already documents the "written from the corpus slug list before
  any score was measured" discipline this plan must not violate.
- `cmd/forge/testdata/intent-gate.golden`, `cmd/forge/intent_gate_test.go` —
  `TestIntentGateSeparation`, `minFireAdmitted = 8` (`:39`), `intentMargin` (`:119`).
- `cmd/forge/intent.go` — `intentGate = 0.50`. Untouched by this plan.
- `cmd/forge/testdata/query-ecosystems.txt`, `cmd/forge/neighbour_frequency_test.go` — the
  precedent this plan mirrors: a mechanical, pre-committed labelling rule, committed in its
  own commit ahead of any scoring re-run, so the ordering is checkable in git history
  rather than asserted in prose.
- `examples/vault/notes/**` — the source corpus new prompts must be drawn from.

**Prerequisites.**

- **This is measurement, not design.** Nothing in this plan touches `intentGate` (0.50) or
  `pkg/recall`'s scoring. If the widened margin is still negative, the output is a recorded
  finding for a future design-only follow-up, not a threshold nudge inside this same work.
- **New prompts must be written before any score is looked at**, from the corpus's own note
  titles/topics the same way the existing 25 were, covering ecosystems and phrasings the
  current 25 under-represent — not written because they're expected to land near
  0.407–0.443. Picking prompts after guessing where they'd score is the same failure shape
  as B-036's own two wrong readings (evidence selected after seeing the answer it was meant
  to test), and it would make the resulting margin unfalsifiable in either direction.
- `minFireAdmitted = 8` is pinned against today's 10 FIRE prompts (8/10). Widening the FIRE
  count changes what that absolute floor means as a fraction — resolve this explicitly in
  step 4 rather than letting it silently become a laxer or stricter bar than the comment at
  `intent_gate_test.go:33-38` argues for.

**Steps.**

1. From `examples/vault`'s note corpus, list ecosystems/topics the current 10 FIRE prompts
   don't cover, and adjacent-topic/off-corpus QUIET topics the current 15 don't cover. Aim
   for roughly doubling the set while keeping the FIRE:QUIET ratio close to today's 10:15 —
   matching B-036's own scale of widening (9→15 queries, ~1.67x).
2. Append the new lines to `cmd/forge/testdata/intent-gate-labels.txt`, same `FIRE:`/
   `QUIET:` format, **as its own commit**, before touching any test or golden file — the
   same ordering discipline `query-ecosystems.txt` used.
3. Run `go test ./cmd/forge -run TestIntentGate -update` to regenerate
   `intent-gate.golden` against the widened set. **Read the diff by hand before accepting
   it** — the same discipline B-036's closing note names ("checked before accepting
   `-update`'s output rather than after").
4. Resolve `minFireAdmitted`: either keep it as an absolute floor of 8 (state why the new
   FIRE prompts don't change what "clearing today's known-good set" means), or rescale it
   and derive the new number in the comment the same way 8 is derived today. Don't leave it
   unexamined.
5. Recompute the margin (`intentMargin`'s lowest-FIRE-minus-highest-QUIET) on the widened
   set and compare it to the recorded -0.036. Record whether it stays negative, narrows, or
   widens.
6. Write the result into BACKLOG's B-037 text as a measurement note — not a closure, since
   this plan doesn't decide whether the gate needs to move, only whether the finding
   replicates at a larger sample. If the margin is still negative, name that as the trigger
   for a future design-only follow-up; if it's non-negative, say plainly that the original
   -0.036 looks like this 25-prompt sample's noise.

**Verification.**

- `go test ./cmd/forge -run TestIntentGate` (`TestIntentGateSeparation` included) → green
  under whatever `minFireAdmitted` step 4 lands on.
- `go test ./...` on both build lanes → green; nothing in `pkg/recall` changes, so
  `TestCalibration` and the neighbour goldens are untouched (no `-update` run on them).
- `git log` shows the labels-file commit strictly before the golden-regeneration commit.

**Done when** `intent-gate-labels.txt` is meaningfully widened with pre-committed labels,
`intent-gate.golden` reflects the widened set, and BACKLOG's B-037 text states the
re-measured margin as a fact — either confirming the overlap persists at scale (opening the
real design question as a follow-up) or showing it doesn't (closing B-037 outright as
sample noise, the same way B-031 closed with no code change).

**Do not.** Do not move `intentGate` (0.50) in this same work — that is a separate,
design-shaped decision this plan deliberately doesn't make. Do not write new labelled
prompts by guessing which score they'll land near. Do not treat this as satisfying the
*other* named unblock path (targeted boundary sampling) — it doesn't, and if the wide sweep
alone doesn't settle the question, B-037 should stay open, re-scoped to that alternative,
rather than being called done.

---

<a name="no-steps"></a>

# NO STEPS — open, but not implementation work

These have no plan on purpose. Each names the condition that would change that.

## B-025 — `PostToolUse`/WebFetch payload shape — **BLOCKED**

**Do not re-attempt the WebFetch.** Three tries against two official doc pages already
failed; a fourth is not new evidence. The unblock trigger is **observational**: a live
`PostToolUse` hook firing on a real `WebFetch` call, whose payload can be captured and
read. Until someone has that payload in hand there is no work to do — `cacheBody`
(`cmd/forge/cache_source.go`) already handles both shapes and both branches are tested
(`cache_source_test.go:12,34,49,59`). The current code is the correct response to not
knowing, not a placeholder.

When the payload arrives: update `cacheBody` to extract the real text field instead of
caching the wrapper JSON, and keep both branches — the raw fallback stays correct for any
shape that changes later.

## B-012 — `code_refs` has no live producer

`agents/forge-librarian.md` already carries the instruction (`:38-39`, `:63`, `:75`) and
`pkg/coderef.FromFrontmatter` already reads the canonical block
(`cmd/forge/drift.go:147`). The gap is that **root-level `agents/` is not live** — Claude
Code loads `.claude/agents/`, and the plugin manifest that would close this shipped in
Phase 6 but has never been verified from a clean machine.

**Unblock condition:** a verified plugin install where `forge-librarian` actually
dispatches. Until then the field is documentation, every note reaches drift through
`pkg/coderef`'s recovery path, and B-018's ambiguity is the direct consequence.

## B-038 — `bodyPass`'s top-20 window is allocated by path, not by relevance

**Measure before designing — see BACKLOG's entry for the full derivation, split out of
B-031's closure.** `bodyPass` opens `cands[:20]` after sorting by frontmatter score then
**path ascending**; on any query where more candidates tie at 0.000 frontmatter than fit
the window, which directory a note lives in (not its content) decides whether its body is
ever read. Confirmed for one pair: `transactional-outbox-pattern.md` and
`cqrs-and-event-driven-messaging.md` (`notes/howto/`) carry heavy body signal for a real
query and are never opened, because `notes/concept/*` fills the window first.

**Unblock condition:** run `TestCalibration`'s corpus staging at a widened or removed
`BodyPassSize` across more than the nine-query calibration set (the entry's own check
found no change on those nine) to see whether the window ever changes a real verdict —
that answers question (a) in the BACKLOG entry. If it does, question (b) — what the
tie-break should be instead of path — needs its own design pass, not a default guess.

**Do not respond to this by simply raising `BodyPassSize`.** The cost (a file open per
window slot) is real and unmeasured at corpus scale; DESIGN §8 sized "top 20" for latency,
and nothing here shows a wider window changes an actual verdict yet.

## B-003 — repo directory still named `TIL`

**User decision.** Renaming breaks shell aliases, IDE projects, and `.idea/` state pointing
at the old path. Cosmetic today; mildly annoying as tooling and README paths assume the
artifact name `knowledge-forge`. Do not rename unasked — CLAUDE.md says so explicitly.

## B-004 — module path has no VCS host prefix

**Deferred by decision** (2026-08-08, "no need github for now"). `module knowledge-forge`
is legal and fine for a goreleaser-distributed binary. If the module ever needs to be
importable by others it becomes `github.com/<user>/knowledge-forge`, which rewrites every
import line. Effectively free today; the cost grows with the file count.

---

# Standing rules that apply to more than one item

Collected here so a plan does not have to restate them and a session cannot miss them.

1. **Scoring changes must show the golden diff.** `go test ./cmd/forge -run TestCalibration
   -update` rewrites `cmd/forge/testdata/calibration.golden`, and that diff is the review
   surface. A `pkg/recall` change without it is unreviewed, not harmless. Applies to
   B-031, B-032, B-036.
1b. **Two more goldens are now derivations, not just records.**
   `testdata/neighbour-sweep.golden` and `testdata/intent-gate.golden` are produced by
   `TestNeighbourFloorSweep` and `TestIntentGateSeparation` from committed label files.
   A scoring change re-runs both with `-update` and states the verdict. **Never edit the
   label files to match new scores** — they were written before any score was measured and
   that ordering is the only thing making the derivation honest.
2. **The answer/update thresholds do not move.** 0.85 / 0.55, per DESIGN §5.3 and B-008's
   closure. `neighbour_min_score` is a separate knob and is the only threshold in scope
   anywhere in this file.
3. **`idf(0, n) == 0` is correct** and its test must not be inverted — the absent-term
   policy lives one layer up, in `weightsOver`.
4. **The vocabulary filter applies to `--stack` hints, not to question terms.** The reverse
   looks like the obvious reading and is the bug B-008 fixed.
5. **Two build lanes.** `CGO_ENABLED=0 go build ./... && go test ./...` is the default
   lane; `pkg/codeindex` needs `CGO_ENABLED=1`. Check both before claiming green.
6. **`forge drift` <100ms is the binding latency budget** — it runs on the git-hook path.
   Measured 60–70ms. Anything touching `pkg/drift` or `pkg/codeindex` re-measures.
7. **Reports render deterministically.** Six consecutive runs, md5-identical. Ranked output
   needs a `sort.Slice` tiebreak unique in its collection (B-020).
8. **Rebuild the vault's hook binary after any `pkg/dataset` or `cmd/forge/capture.go`
   change**: `CGO_ENABLED=0 go build -o ~/.forge/bin/forge ./cmd/forge`. The hook never
   fails a commit and never prints, so a stale binary is silent — read
   `<vault>/.forge/capture.log` if pairs stop appearing.
