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
re-grep before trusting a line number. **B-037 got a PLANNED section 2026-08-27, ran the
same day, and was dropped back to NO STEPS the same day** — same shape as B-036's own
measure-then-done arc, just without a design at the end of it. The wide-sweep half of its
two named unblock paths ran (widen `intent-gate-labels.txt`, per user decision); the margin
came back unchanged at -0.036, which rules out sample noise but doesn't resolve the item —
see BACKLOG's B-037 measurement note and its NO STEPS section below for what's left.
**Closed and recorded items are dropped from this file
entirely, not just their plan sections** — `docs/BACKLOG.md` is the full census (every ID,
closed or open, with the reasoning) and the durable record of *why*; this file exists only
to track *how to close* whatever is still open. If an ID isn't below, it's done — look it
up in BACKLOG.

## Two status classes

Not every open backlog item can take a plan, and the absence of steps below is a decision,
not an oversight:

- **PLANNED** — workable, has a full six-field section in this file. **None right now.**
  B-034 (2026-08-25) was the last of the prior eight: B-029, B-027, B-033, B-032, B-023,
  B-031, B-015 and B-035 closed 2026-08-23/24/25 before it. B-033's closure opened B-036
  and B-032's opened B-037. **B-036 closed 2026-08-26** (measured, then designed and
  shipped the same day — see its BACKLOG entry) without ever getting a PLANNED write-up
  here: the design was scoped and low-risk enough (widen `Rank`'s internal window, one new
  constant, two new functions, three golden regenerations) to go straight from measurement
  to implementation once its ecosystem-label prerequisite landed. **B-037 got a PLANNED
  section 2026-08-27, ran it the same day, and dropped back to NO STEPS** — its plan was
  scoped to measurement only (widen the labelled corpus, re-measure the margin), never to
  moving `intentGate`, and the margin came back unchanged, so there was no design to do.
  B-038 is still NO STEPS by its own argument, naming a measurement to run before any
  design is chosen.
- **NO STEPS** — open but not actionable by an implementation session: blocked on external
  observation, or a user decision, or "record, don't fix" by standing rule. Listed with
  its unblock condition instead of steps.

The execution-order chain this file tracked through B-034 (`B-029 → B-033 → B-032 →
B-023 → B-031 → B-015 → B-035 → B-034 → B-027`) is fully closed and dropped along with
the plan sections below, per this file's own rule — the sequencing rationale (why B-033
had to land before B-032, why B-029 went first) is recorded in each item's BACKLOG
closing note, not restated here.

---

## Index — 5 open items (34 closed/recorded IDs live in BACKLOG only)

| ID | Subject | Class | Section |
|---|---|---|---|
| B-003 | Repo directory still named `TIL` | NO STEPS — user decision | [below](#no-steps) |
| B-004 | Module path has no VCS host prefix | NO STEPS — deferred by decision | [below](#no-steps) |
| B-012 | `code_refs` has no live producer | NO STEPS — blocked on packaging | [below](#no-steps) |
| B-037 | Intent gate FIRE/QUIET margin now negative | NO STEPS — measure further (targeted band) | [below](#no-steps) |
| B-038 | `bodyPass` window allocated by path, not relevance | NO STEPS — measure first | [below](#b-038--bodypasss-top-20-window-is-allocated-by-path-not-by-relevance) |

---

<a name="no-steps"></a>

# NO STEPS — open, but not implementation work

These have no plan on purpose. Each names the condition that would change that.

## B-012 — `code_refs` has no live producer

`agents/forge-librarian.md` already carries the instruction (`:38-39`, `:63`, `:75`) and
`pkg/coderef.FromFrontmatter` already reads the canonical block
(`cmd/forge/drift.go:147`). The gap is that **root-level `agents/` is not live** — Claude
Code loads `.claude/agents/`, and the plugin manifest that would close this shipped in
Phase 6 but has never been verified from a clean machine.

**Unblock condition:** a verified plugin install where `forge-librarian` actually
dispatches. Until then the field is documentation, every note reaches drift through
`pkg/coderef`'s recovery path, and B-018's ambiguity is the direct consequence.

## B-037 — `forge intent`'s FIRE/QUIET margin went negative under B-032's scale

**One of the two original unblock paths already ran — see BACKLOG's B-037 measurement
note (2026-08-27).** `intent-gate-labels.txt` was widened 25 → 50 prompts (10 → 20 FIRE,
15 → 30 QUIET), covering nine ecosystems the original set under-represented, and the
margin came back byte-identical: lowest FIRE 0.407, highest QUIET 0.443, margin **-0.036**
— the same two prompts still define both edges of the band at twice the sample. `intentGate`
(0.50) was not moved; `minFireAdmitted` was rescaled 8 → 16 to keep pinning the measured
80% admission rate rather than silently loosening.

**What that answers and what it doesn't.** It rules out "the -0.036 finding is this
25-prompt sample's noise" — a real, stable overlap band survives 2x the sample and nine
more ecosystems. It does not answer whether a FIRE prompt and a QUIET prompt actually trade
places *inside* [0.407, 0.443] on some as-yet-unwritten prompt, only that the band's outer
edges haven't moved. Nothing is mis-admitted today either way — the gate (0.50) still
clears the whole band with room (0.057).

**Unblock condition, narrowed to what's left:** the other original path — more labelled
prompts written specifically to land inside [0.407, 0.443] — has not run. Same discipline
required: prompts written from the corpus's own topics before any score is measured, not
chosen by guessing which score they'd get, which would make the result unfalsifiable.

**Do not respond to this by moving the gate.** Still true after the wide sweep: 0.50 is
measured safe against every labelled prompt on file, including the widened 50, and a stable
overlap band the gate sits above is a reason to gather the remaining evidence, not to
re-derive a number that isn't failing.

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
