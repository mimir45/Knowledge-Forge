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
a rejection (wrong denominator), corrected same day: the hypothesis is confirmed, item
stays open.** A short-lived B-039, opened on the same wrong pass, was retracted outright —
see BACKLOG.md's B-036 closing note for the full correction, including why the "no code
change" and "B-039" text both briefly existed here before being reverted. Anchors were
verified against the tree at that date; re-grep before trusting a line number.
**Closed and recorded items are dropped from this file
entirely, not just their plan sections** — `docs/BACKLOG.md` is the full census (every ID,
closed or open, with the reasoning) and the durable record of *why*; this file exists only
to track *how to close* whatever is still open. If an ID isn't below, it's done — look it
up in BACKLOG.

## Two status classes

Not every open backlog item can take a plan, and the absence of steps below is a decision,
not an oversight:

- **PLANNED** — workable, has a full six-field section in this file. **None written yet.**
  B-034 (2026-08-25) was the last of the eight: B-029, B-027, B-033, B-032, B-023, B-031,
  B-015 and B-035 closed 2026-08-23/24/25 before it. B-033's closure opened B-036 and
  B-032's opened B-037. **B-036's own unblock condition was run 2026-08-26 and now has a
  real answer** (see its BACKLOG entry) — it is the one item here closest to PLANNED, but
  the design step is a `pkg/recall` scoring change and deliberately wasn't written in the
  same pass that corrected the measurement. B-037/B-038 are still NO STEPS by their own
  argument, each naming a measurement to run before any design is chosen.
- **NO STEPS** — open but not actionable by an implementation session: blocked on external
  observation, or a user decision, or "record, don't fix" by standing rule. Listed with
  its unblock condition instead of steps.

The execution-order chain this file tracked through B-034 (`B-029 → B-033 → B-032 →
B-023 → B-031 → B-015 → B-035 → B-034 → B-027`) is fully closed and dropped along with
the plan sections below, per this file's own rule — the sequencing rationale (why B-033
had to land before B-032, why B-029 went first) is recorded in each item's BACKLOG
closing note, not restated here.

---

## Index — 7 open items (32 closed/recorded IDs live in BACKLOG only)

| ID | Subject | Class | Section |
|---|---|---|---|
| B-003 | Repo directory still named `TIL` | NO STEPS — user decision | [below](#no-steps) |
| B-004 | Module path has no VCS host prefix | NO STEPS — deferred by decision | [below](#no-steps) |
| B-012 | `code_refs` has no live producer | NO STEPS — blocked on packaging | [below](#no-steps) |
| B-025 | `PostToolUse`/WebFetch payload shape | NO STEPS — **BLOCKED** | [below](#no-steps) |
| B-036 | Two general Spring notes admitted on every Spring-flavored query | NO STEPS — measured, needs a design pass | [below](#no-steps) |
| B-037 | Intent gate FIRE/QUIET margin now negative | NO STEPS — measure first | [below](#no-steps) |
| B-038 | `bodyPass` window allocated by path, not relevance | NO STEPS — measure first | [below](#b-038--bodypasss-top-20-window-is-allocated-by-path-not-by-relevance) |

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

## B-036 — a broad query links ten neighbours, and no floor can separate them

**Measured 2026-08-26 against its own unblock condition — see BACKLOG.md's B-036 closing
note for the full number and the correction of a wrong first reading. Still open: the
hypothesis holds, and what's missing now is a design, not more measurement.**

`cmd/forge/neighbour_frequency_test.go` (`TestNeighbourDocumentFrequency`) ran the "appears
in N of M query results" column this entry originally asked for, over
`testdata/neighbour-labels.txt`'s fifteen queries. Read against the right denominator —
the four queries that literally contain "Spring" among the fifteen, not all fifteen —
`spring-cli-and-maven-commands-for-spring-boot` and `meterreadingsservice-spring-boot-4-x-
project` both appear in **every one of them**. That is this entry's original claim,
confirmed: a note scoring on every query in an ecosystem is being admitted as a neighbour
regardless of what the question specifically asks.

**Unblock condition (updated): none — the measurement is done and it confirms the
hypothesis.** What remains is a design decision this entry always deferred to "after
reading it," now actually reachable: `Thresholds.Neighbours` (`pkg/recall/rank.go:102-
110`) only ever sees `recall.Rank`'s already-truncated top-10 candidate list, so excluding
a universal note there cannot surface an 11th candidate `Rank` never computed in the first
place — any real fix has to touch `Rank`'s internal window, not just `Neighbours`'
filtering, which is a bigger, more careful `pkg/recall` change than a one-line filter. That
design (and its own PLANNED write-up) is the next step, not attempted in the same pass
that corrected this entry's own measurement.

**Do not respond to this by raising the floor.** Unchanged from the original entry —
every floor B-033's sweep tried that drops these two notes also drops the Storybook
family B-033 was opened to fix.

## B-037 — `forge intent`'s FIRE/QUIET margin went negative under B-032's scale

**Measure before touching the gate — it is not broken today, and that is the reason there
is no plan below rather than a threshold nudge.** `cmd/forge/testdata/intent-gate.golden`:
before B-032, the lowest gate-admitted FIRE prompt and the highest QUIET prompt separated
by +0.005 (0.407 vs 0.402). After, they overlap: lowest FIRE 0.407, highest QUIET **0.443**,
margin **-0.036**. `TestIntentGateSeparation`'s two pinned invariants both still hold
mechanically — the gate (0.50) admits zero QUIET prompts and exactly `minFireAdmitted` (8)
FIRE prompts, both unchanged counts — because the gate sits above the whole overlapping
band (0.407–0.443), not inside it.

What changed is the *reason* 0.50 is safe. B-033's derivation argued it as "the lowest
value still a clear step above the QUIET ceiling" — a margin argument. That margin, read
literally across the full FIRE/QUIET split rather than just at the gate, is now negative:
somewhere in [0.407, 0.443] a FIRE prompt and a QUIET prompt trade places by score, so no
single threshold separates the two labelled sets everywhere, only above 0.443 specifically.
0.50 clears 0.443 with room (0.057), so nothing is mis-admitted today.

**Unblock condition:** either more labelled prompts on both sides of the overlap band, to
tell whether -0.036 is this 25-prompt sample's noise or a real, growing overlap as more of
the vault's tags/stack channels shift under B-032 — or a wider sweep of `examples/vault`
prompts the way `neighbour-labels.txt` widened B-033's evidence past nine queries. Either
answers whether 0.50 keeps clearing the band as the corpus grows, which this measurement
alone cannot say.

**Do not respond to this by moving the gate.** 0.50 is still measured safe against every
labelled prompt on file; a margin turning negative in a data slice the gate doesn't
actually sit inside is a reason to get more data, not a reason to re-derive a number that
isn't failing.

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
