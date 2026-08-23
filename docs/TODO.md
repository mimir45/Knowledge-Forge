# TODO — executable plans for the open backlog

`docs/BACKLOG.md` is the record of *why* each item exists. This file is the record of
*how to close it*. One section per workable item, each written so a future session can
open it cold, follow the steps, and stop when "Done when" is satisfied.

**Read `docs/BACKLOG.md`'s own entry before starting any item here.** These plans are
derived from it and deliberately do not restate its argument. Where a backlog entry has a
`###` closing section, that section is more current than its `Status:` line.

Written 2026-08-23 from a full scan of B-001…B-035. Anchors were verified against the
tree at that date; re-grep before trusting a line number.

## Three status classes

Not every backlog item can take a plan, and the absence of steps below is a decision, not
an oversight:

- **PLANNED** — open, workable, has a full six-field section in this file. Nine items when
  this was written; **B-029, B-027 and B-033 closed 2026-08-23**, leaving six. B-033's
  closure opened **B-036**, which is NO STEPS by its own argument: it names a measurement
  to run before any design is chosen, not a fix to apply.
- **NO STEPS** — open but not actionable by an implementation session: blocked on external
  observation, or a user decision, or "record, don't fix" by standing rule. Listed with
  its unblock condition instead of steps.
- **CLOSED / RECORDED** — done, or a measured deviation kept on purpose. One row, no
  section. Do not re-litigate these; the backlog carries the reasoning.

## Execution order

```
B-029  →  B-033  →  B-032  →  B-031  →  B-015  →  B-023  →  B-035  →  B-034  →  B-027
 done     done      done      cover     imports   engine    run_id    D6       done
```

**B-029, B-027, B-033 and B-032 all closed 2026-08-23.** **B-031 is now the head** — its
own prerequisite ("B-033 and B-032 should land first") is now satisfied.

**B-033 landed before B-032, which was the load-bearing part of this order.** Both
re-measure against `cmd/forge/testdata/calibration.golden`, and B-032 moves `blend`'s
denominator for a large fraction of the vault, so every score shifts under it. B-033's
floor is now derived against the pre-B-032 scale — **so B-032 must re-run the derivation,
not just re-record the golden.** That is cheap on purpose: `TestNeighbourFloorSweep` reads
the committed labels and rewrites `testdata/neighbour-sweep.golden` with `-update`, and
the same applies to `TestIntentGateSeparation`. Both are steps in §B-032 below.

B-029 was first because it is independent of everything else and was the largest single item.
B-027 was last because it is documentation with no code consequence left.

---

## Index — all 36 items

| ID | Subject | Class | Section |
|---|---|---|---|
| B-001 | Doc coherence audit | CLOSED 2026-08-09 | — |
| B-002 | `testdata/vault/` vs `examples/vault/` | RECORDED (a trap, not work) | — |
| B-003 | Repo directory still named `TIL` | NO STEPS — user decision | [below](#no-steps) |
| B-004 | Module path has no VCS host prefix | NO STEPS — deferred by decision | [below](#no-steps) |
| B-005 | Seven note types | CLOSED 2026-08-09 | — |
| B-006 | Path-qualified wikilink rewrite | CLOSED 2026-08-09 | — |
| B-007 | Librarian stamps `Forge-Write: true` | CLOSED Phase 4 | — |
| B-008 | IDF weighting / absent-term admission | CLOSED 2026-08-22 | — |
| B-009 | `pkg/gitsig` shells out to `git` | CLOSED 2026-08-21 | — |
| B-010 | AUDIT §7 git-history correction | RECORDED, doc unedited | — |
| B-011 | `reports/`+`moc/` as graph nodes | CLOSED in 2b | — |
| B-012 | `code_refs` has no live producer | NO STEPS — blocked on packaging | [below](#no-steps) |
| B-013 | Code-index cache format version | CLOSED 2026-08-18 | — |
| B-014 | Index parses TypeScript, not Kotlin | RECORDED — deliberate swap | — |
| B-015 | `CodeGroup.DependsOn` never populated | **PLANNED** | [§B-015](#b-015--populate-codegroupdependson) |
| B-016 | Vault carries `sources:` and `source:` | RECORDED — read-both shipped | — |
| B-017 | §B.5's 90-day window shows nothing | RECORDED — measured, defaults question | — |
| B-018 | Bare symbol citation arbitration | RECORDED — known asymmetry | — |
| B-019 | Duplicate detection deviations | RECORDED — deliberate, measured | — |
| B-020 | `sort.Slice` tiebreaks | CLOSED in 2b | — |
| B-021 | B2B is a separate project | CLOSED — scope decision | — |
| B-022 | `engine_trail` schema pattern | CLOSED Phase 4 | — |
| B-023 | `on_exhausted`: `stop` halts nothing | **PLANNED** (behaviour half) | [§B-023](#b-023--make-on_exhausted-mean-something) |
| B-024 | `D2Tag` spelling | CLOSED 2026-08-21 | — |
| B-025 | `PostToolUse`/WebFetch payload shape | NO STEPS — **BLOCKED** | [below](#no-steps) |
| B-026 | Deleted-file citation never BROKEN | CLOSED 2026-08-16 | — |
| B-027 | `.forge/code-index-<repo>.json` naming | CLOSED 2026-08-23 | — |
| B-028 | Hook path immediacy on deletion | CLOSED 2026-08-17 | — |
| B-029 | `errcheck` disabled tree-wide | CLOSED 2026-08-23 | — |
| B-030 | `dataset.capture` gates only two tiers | CLOSED Phase 6b | — |
| B-031 | Kafka/Testcontainers coverage miss | **PLANNED** | [§B-031](#b-031--the-coverage-side-of-the-scoring-surface) |
| B-032 | Untagged note escapes absent-term penalty | CLOSED 2026-08-23 | — |
| B-033 | 0.30 neighbour floor on the old scale | CLOSED 2026-08-23 | — |
| B-034 | D6 (code↔knowledge) not built | **PLANNED** | [§B-034](#b-034--build-d6-as-an-export-view) |
| B-035 | D1 has no outcome label | **PLANNED** | [§B-035](#b-035--mint-a-run_id-so-d1-can-carry-an-outcome) |
| B-036 | Broad query links ten neighbours | NO STEPS — measure first | [below](#no-steps) |
| B-037 | Intent gate FIRE/QUIET margin now negative | NO STEPS — measure first | [below](#no-steps) |

---

# B-029 — re-enable `errcheck` — **CLOSED 2026-08-23**

Done. `errcheck` is off `.golangci.yml`'s `disable:` list and `golangci-lint run ./...`
returns zero findings under the version `ci.yml` pins. The plan that stood here is kept
only as the four corrections it earned, because each was a wrong number this file taught a
later session to trust:

1. **The worklist was 35, not 95 / ~37 / ~27+~10.** Those came from a bare `errcheck`
   binary and a hand-applied exclusion estimate. The number that matters is what
   `golangci-lint` **v1.64.8** — `ci.yml`'s pin — reports against the repo's own config: 35,
   splitting 26 test / 9 production.
2. **Measure with the pinned linter, not a fresh one.** `golangci-lint` v2 reports 50 on the
   same tree; it is a different tool with a different exclusion set and its number is not
   this repo's gate.
3. **Stock truncation hides findings.** v1 defaults to `max-issues-per-linter: 50` and
   `max-same-issues: 3`; the same tree reports **22** at those limits. `.golangci.yml` now
   sets both to `0`. This was never an errcheck question — it applied to every linter in
   the default set for all of Phase 6.
4. **Step 5's two signature changes were traced and rejected.** `pkg/drift/apply.go`'s
   `stamp()` and `pkg/drift/gitindex.go`'s `persist()` each hold one call whose failure is
   self-healing, and neither caller has an error channel — unlike `refresh()`, which
   promised propagation it never performed and hid three swallowed errors. Both are
   `//nolint:errcheck` with the reason on the line. Step 6's re-size held: `cmd.Wait()` was
   the one real fix and it is now checked, guarded so a truncated stream still reports the
   read error rather than a broken pipe.

Full closure note: `docs/BACKLOG.md` B-029, final section.

---

# B-033 — re-derive the neighbour floor — **CLOSED 2026-08-23**

Done. `neighbour_min_score` is **0.125** at both default sites, and `cmd/forge/intent.go`'s
gate is `recall.DefaultThresholds.Update` rather than a hardcoded 0.7. The plan that stood
here is kept only as the five corrections it earned.

1. **The "before" state was not uniformly zero neighbours.** B-033's entry generalised from
   the Storybook row; three of §3.1's nine emitted none, the other six emitted one to five.
   The plan's step 2 was right to demand that column exist as its own commit — it is what
   showed the entry was over-general.
2. **The plan missed `references/recall-spec.md`.** B-032's section has a spec-update step
   and this one did not, while §3, §3.1 and §3.2 all described the old floor as current.
   Any threshold item needs that step; it is not specific to B-032.
3. **A unit test spelled the threshold into a fixture.** `pkg/recall`'s
   `TestNeighbourBandEdges` used `0.30` as its "inclusive lower bound" literal and failed
   on the change as if the band had broken. It now expresses both edges in terms of
   `DefaultThresholds`. Worth checking for before any threshold move: a test that pins a
   *number* while claiming to test a *rule* reads as a real regression.
4. **Step 6's "two acceptable outcomes" were both wrong for the intent gate.** Promoting
   to config was rejected on the 50ms hook budget; keeping 0.7 with a better comment was
   ruled out by measurement, since it admitted 3 of 10 FIRE prompts. The answer was a
   third: re-derive the literal (0.50) and defend it with a test rather than a comment.
   Reusing `DefaultThresholds.Update` was tried before that and rejected — it reads as the
   elegant answer and costs three near-verbatim title matches to buy an alignment
   `printIntent` never asserts, since it computes no verdict at all.
5. **Labelled prompts could rule the old number out but not choose the new one.** The FIRE
   and QUIET classes separate at 0.402/0.407, so every value in [0.405, 0.7] scores
   identically on false positives. A derivation that ends in a range still needs an
   argument to land on a number, and pretending the data chose it would have been the
   dishonest half of this item.

The sweep survives as `TestNeighbourFloorSweep` and `TestIntentGateSeparation`, both
reading committed label files and both re-recordable with `-update`. **B-032 must re-run
them**, not just re-record the calibration golden.

Full closure note: `docs/BACKLOG.md` B-033, final section. The ruling that supersedes
`DESIGN:257` is `docs/AUDIT.md` §8.4 **D-9**.

---


# B-032 — activation vs. absent-term penalty — **CLOSED 2026-08-23**

Done, but not by the shape step 2 describes below. That step's English ("carries the field
**and** the query has something the field could answer") reads as the code already shipped
(`ok && len(tags) > 0`) — the two candidates that literal reading actually admits were both
computed by hand first and one is disqualified: dropping `len(tags) > 0` entirely fixes the
cited row but resurrects B-008's false positive at first place (step 5's own named check)
*and* deletes the untagged-exemption rule step 5's prerequisites forbid deleting. The
shipped fix reads "carries the field" as "carries a **hit**" — `len(hits) > 0` in place of
`len(tags) > 0` / `len(stack) > 0` — which needed no change to `weightsOver` at all, so the
"plus whatever weightsOver must expose" half of step 2 turned out to be nothing.
Step 4b was not optional and both derivations moved: the neighbour floor's F1 peak shifted
0.125 → 0.150 (§B-031's prerequisite below is now satisfied), and the intent gate's
FIRE/QUIET margin went negative (+0.005 → -0.036) — mechanically still safe (gate 0.50,
8/10 FIRE admitted, 0 QUIET admitted) but no longer a clean separation, filed as **B-037**
rather than nudged, exactly as step 4b's own text anticipated. Full closure note:
`docs/BACKLOG.md` B-032, final section.

---

# B-032 (original plan, kept for the corrections above) — activation vs. absent-term penalty

**Why it's open.** §2.5's activation is two-sided: a note with no `tags:` leaves the tags
channel *inactive*, dropping out of the blend's denominator rather than scoring zero.
B-008's admission then charges the tags channel for query terms the vault tags nowhere. A
tagged note pays that charge in full; an untagged note never sees it. So a note can win a
row partly by carrying no tags — §2.5's own effect running opposite to its intent.

**Anchors.**

- `pkg/recall/score.go:218` — `if c.Active = ok && len(tags) > 0; c.Active {`. The tags
  activation rule.
- `pkg/recall/score.go:231` — the same for stack.
- `pkg/recall/score.go:287-293` — `blend`, the weighted mean over active channels; `:293`
  is the `if !c.Active { continue }` that removes the channel from the denominator.
- `pkg/recall/score.go:110-124` — `weightsOver`, where an absent term takes the mean of the
  present ones. `:149` `idf`.
- `pkg/recall/score.go:208`, `:245` — the doc comments arguing both halves from measurement.

**Prerequisites.**

- Both rules are argued from measured vault behaviour. **Do not "fix" this by deleting
  either one.**
- **B-033 landed first and its floor is derived against the pre-B-032 scale.** Re-running
  `TestNeighbourFloorSweep` and `TestIntentGateSeparation` is a step below, not an
  optional check: this change moves every score, so the floor's derivation and the intent
  gate's separation margin are both measured against a scale that no longer exists.
- Exposure: 9 of `examples/vault`'s 91 notes are untagged. CLAUDE.md records 31 of 91 in
  the live vault as missing `tags:` or `stack:` after the Phase 1 migration — so the real
  corpus is worse than the example corpus suggests.
- B-033 should land first (see Execution order).

**Steps.**

1. Reproduce the measurement in the entry before changing anything: "Redis caching in
   Spring Boot" against `examples/vault` should give
   `meterreadingsservice-spring-boot-4-x-project` 0.500 with `tags: []` above
   `spring-cli-and-maven-commands-for-spring-boot` 0.415 with `tags: [spring-cli]`.
   If it does not reproduce, stop and re-open the entry — the corpus or the scorer moved.
2. Implement the shape the entry proposes: activation currently asks *whether the note
   carries the field at all*; make it ask whether the note carries the field **and** the
   query has something that field could answer. An untagged note is then neither penalised
   nor advantaged. Keep the change inside `score.go`'s two activation lines plus whatever
   `weightsOver` must expose for the second half of the condition.
3. Measure the blast radius explicitly: how many of the 91 notes change activation state,
   and how many change rank. The denominator moves for a large fraction of the vault, so
   **every score moves** — that is expected and must be stated, not discovered in review.
4. Re-record: `go test ./cmd/forge -run TestCalibration -update`. Paste the diff into the
   commit message.
4b. **Re-run B-033's two derivations against the new scale**:
   `go test ./cmd/forge -run 'TestNeighbourFloorSweep|TestIntentGateSeparation' -update`.
   The label files are committed and must not be edited — re-labelling after seeing new
   scores is how a derivation becomes a fit. Read the re-recorded
   `neighbour-sweep.golden`: if F1's peak has moved off 0.125, the floor moves with it and
   `pkg/recall/doc.go` + `config/forge.config.example.md` change together. Read
   `intent-gate.golden`: if the FIRE/QUIET margin has gone negative the classes overlap and
   no gate works, which is a finding, not a number to nudge.
5. Re-read the golden diff against B-008's closure note. If B-008's false positive
   (0.415 after the fix) returns to first place, the change is wrong regardless of what
   this row does.
6. Update `references/recall-spec.md` §2.5 — the asymmetry note there describes the old
   behaviour.

**Verification.**

- `go test ./cmd/forge -run TestCalibration` → green against the re-recorded golden.
- `go test ./pkg/recall/...` → green. `idf(0, n) == 0` must still hold; the absent-term
  policy lives one layer up in `weightsOver` and its test must not be inverted.
- `go test ./...`, both lanes.

**Done when** an untagged note is neither charged nor exempted, the golden diff is in the
commit, spec §2.5 matches the code, B-008's false positive has not returned, and B-033's
two derivations have been re-run with their verdicts stated — including "unchanged", if
that is what they say.

**Do not.** Do not move the answer/update thresholds. Note carefully: this change moves
every score *without* touching a threshold, so "thresholds didn't move" is **not** on its
own a justification — the commit must argue the denominator change on its merits.

---

# B-031 — the coverage side of the scoring surface

**Why it's open.** "Kafka consumers with Testcontainers" ranks
`testcontainers-docker-based-integration-testing` at 0.311 (CREATE). The note is the right
one, but `kafka` — the term carrying the question — appears in its title, tags and stack
**not at all**. Admission is strictly decreasing for every positive weight, so the same
knob that pushed B-008's false positive down cannot push this up. Split from B-008
deliberately.

**Anchors.**

- `pkg/recall/rank.go:12` — `const BodyPassSize = 20`. The body channel runs only for the
  top 20 candidates.
- `pkg/recall/rank.go:79` — where that cap is applied.
- `pkg/recall/score.go:268` — `Channel{Name: "body", Weight: wBody, Active: true}`. The body
  channel is unconditionally active and carries 0.1 of the blend.
- `docs/KNOWLEDGE-FORGE-DESIGN.md` §8 — the weight ratios this item reopens.

**Prerequisites.**

- **This item needs its own session and its own argument.** It reopens DESIGN §8's weight
  ratios. Do not fold it into a scoring pass that is doing something else.
- **B-033 and B-032 have both landed (2026-08-23)** — measure this item against the
  current `calibration.golden`, not the pre-B-032 numbers in earlier drafts of this file.

**Steps.**

1. Choose between the two shapes the entry names, and write the choice down before coding:
   - **Fix the corpus** — add a `kafka` tag to the note. Honest, and the note *is*
     under-curated. But a fix that edits the vault to make a query score is not a recall
     fix and does not generalise. If chosen, close this item as a corpus fix and say so.
   - **Fix the coverage signal** — the body channel is the only one that sees `kafka` here.
     Whether a term the body carries strongly and the frontmatter carries nowhere should
     lift a candidate is the real question and the more interesting one.
2. If shape 2: the 0.1 body weight and the `BodyPassSize = 20` cap are two separate
   constraints and a candidate at rank 40 never gets a body pass at all. Measure which one
   is binding for this row *before* changing either.
3. Build the argument against DESIGN §8, not against this row. A coefficient nudged until
   one query passes is tuning; B-008 forbids it by name.
4. Re-record: `go test ./cmd/forge -run TestCalibration -update`. Paste the diff.

**Verification.**

- `go test ./cmd/forge -run TestCalibration` → green.
- The B-008 false positive stays out of first place. Check this explicitly — it is the
  regression this whole area exists to prevent.
- `go test ./...`, both lanes.

**Done when** either the corpus fix is made and recorded, or the body-channel question is
answered with an argument that stands without reference to this one query.

**Do not.** Do not move the thresholds. 0.311 and 0.315 sit next to notes that should not
be admitted; lowering the bar admits them too.

---

# B-015 — populate `CodeGroup.DependsOn`

**Why it's open.** ADDENDUM §B.5 asks the codebase map to show what depends on what. The
struct field exists and the renderer handles it; nothing fills it, because
`codeindex.File` captures declarations only and no import edges. So `moc/codebase.md`
groups by directory and ranks by churn but draws no arrows.

**Anchors.**

- `pkg/report/codebase.go:16` — `DependsOn []string`, declared.
- `pkg/report/codebase.go:89-90` — the renderer, already handling a non-empty value.
- `pkg/report/knowledgemap.go:18` — the doc comment recording that nothing populates it.
- `pkg/codeindex/index.go:28` `Symbol`, `:37` `File` — `File` has `Path`, `Lang`, `Symbols`
  and no imports field.
- `pkg/codeindex/index.go:52` — `const Extractor = 2`.

**Prerequisites.**

- `pkg/codeindex` is the **only cgo package** and is build-tag gated. Work on it needs
  `CGO_ENABLED=1` and a host toolchain, and every change must be checked in both lanes.
- "Module = directory" is an **honest limitation, not a placeholder** — nothing in the
  index knows about Maven modules or Go packages, and inventing a grouping the code does
  not declare would file code under modules its authors never wrote. Do not "fix" it as
  part of this item.

**Steps.**

1. Add an `Imports []string` field to `codeindex.File`. Extract it in the tree-sitter pass
   for both supported languages — Java `import` declarations, TypeScript `import` /
   `export … from`.
2. **Bump `codeindex.Extractor` to 3.** This is the step nobody guesses and the one that
   costs a wrong answer if skipped: its doc comment (`index.go:44-51`) says explicitly to
   bump it whenever `Symbol` or `File`'s *serialized shape* changes, not only when
   extraction logic changes, because `json.Unmarshal` would otherwise load an old cache
   "successfully" into a struct it was never written for. B-013 exists for this.
3. Resolve import strings to repo-relative paths, then fold paths up to their directory —
   that is the grouping `CodeGroup` already uses. Unresolvable imports (third-party,
   stdlib) drop out; do not invent nodes for them.
4. Populate `DependsOn` in whatever builds `CodeGroup`, deduped and **sorted** — the
   codebase report must render deterministically, and a `sort.Slice` comparator here needs
   a tiebreak unique in its collection (B-020's rule, four instances of the same bug were
   fixed in 2b).
5. Remove the "nothing populates it" note at `pkg/report/knowledgemap.go:18`.

**Verification.**

- `CGO_ENABLED=1 go test ./pkg/codeindex/...` → green. `CGO_ENABLED=0 go build ./...` →
  still builds (the nocgo lane must not break).
- Determinism: render `moc/codebase.md` six consecutive times, md5-identical. That is the
  standard 2b set for every report and this one is no exception.
- Cache invalidation: run against a repo with a pre-existing
  `.forge/code-index-<repo>.json` written by Extractor 2, confirm it is treated as a **miss**
  and rebuilt, not unmarshalled.
- `forge drift --since-commit` still under its 100ms budget — it is the binding latency
  constraint and it reads this index on the git-hook path.

**Done when** `moc/codebase.md` draws arrows, the report is byte-stable across runs, and
the Extractor bump is in the same commit as the shape change.

**Do not.** Do not introduce Maven/Go module awareness. Do not let the extractor change
land without the version bump — a stale cache silently missing symbols is exactly what
drift reads as BROKEN.

---

# B-023 — make `on_exhausted` mean something

**Why it's open.** The doc half closed 2026-08-21 (four sites now say `stop`). The
behaviour half is untouched: `stop` halts nothing and `degrade` is not a distinct code
path from the default silent fallthrough, so two of the three configured values produce
byte-identical behaviour.

**Anchors.**

- `pkg/config/validate.go:89` — accepts `queue | degrade | stop`.
- `config/forge.config.example.md:82` — the packaged comment.
- `pkg/engine/select.go:30` — the degrade to `"none"`, **unconditional**; `pkg/engine`
  contains zero reads of `OnExhausted`.
- `cmd/forge/engine_run.go:77` — branches on `"queue"` only, to stamp `pending_advisor: true`.
- `cmd/forge/check_drain.go:22` — branches on `"queue"` only.
- `cmd/forge/check_collect.go:181` — passthrough into `cost.md`'s summary line.
- No test exercises `degrade` or `stop`.

**Prerequisites.**

- This is a behaviour change to a **budget-exhaustion path** — it can start failing
  commands that today succeed. It needs a session that owns it and can write the tests.
- AUDIT §8.4 D-5 settled the *default* (`queue`). This item does not reopen that.

**Steps.**

1. Choose one of three, and record the choice in B-023 before coding:
   - **(a)** Give `stop` a real non-zero-exit path; leave `degrade` as today's silent
     fallthrough. Most faithful to the names.
   - **(b)** Collapse to `queue | degrade`, drop `stop`. Backward-incompatible for any
     config that already sets it — needs a validator error message that says what to
     change to, not just "invalid".
   - **(c)** Keep all three and document explicitly that `stop` and `degrade` are synonyms
     today. Cheapest, and honest, but leaves the entry half-open forever — if chosen, say
     so and close the item rather than leaving it to be rediscovered.
2. Implement in `pkg/engine` if the value should influence resolution
   (`select.go:30`'s degrade becomes conditional), or in `cmd/forge` if it should only
   influence the exit path. **Not both** — one read site per meaning.
3. Write the tests. All three values, both the exhausted and non-exhausted branch. The
   absence of any test today is why this drifted.
4. Sync the four doc sites again if any wording changes: `ADDENDUM.md:117`, `:485`, `:671`;
   `CLAUDE-CODE-PROMPT.md:339`.

**Verification.**

- `go test ./pkg/engine/... ./cmd/forge/...` → green, with the new cases present.
- Manual: a config with `on_exhausted: stop` and an exhausted budget behaves as chosen, and
  `cost.md` still renders the configured value verbatim.

**Done when** each accepted value has a distinct, tested behaviour — or the doc states
plainly that two of them do not, and the entry is closed on that statement.

**Do not.** Do not close this on a documentation edit. The entry says so in its own status
line, because that is exactly how the first half was closed.

---

# B-035 — mint a `run_id` so D1 can carry an outcome

**Why it's open.** ADDENDUM §D.1 describes D1's pair as "question → verdict + topic +
stack, auto-labelled by recall **+ outcome**". Phase 6b built the first half. There is no
outcome, because nothing in the system correlates a recall call to the note write that may
follow it minutes later in a different process.

**Anchors.**

- `pkg/dataset/d1.go` — `D1Pair` (Kind, QHash, Topic, Decision, Stack, RecallTopScore,
  Candidates, CapturedAt) and the doc comment already naming this item.
- `pkg/telemetry/event.go:11-21` — `Event`, which has no run id either.
- `cmd/forge/recall.go` — `runRecall`, where D1 captures beside `logAsk`.
- `cmd/forge/recall.go:145` — the JSON envelope's contract comment.
- `cmd/forge/gate.go:45` `cmdGate`, `:62` `runGate` — the write path that would stamp it.
- `skills/forge/SKILL.md` — invokes `forge gate`; the id has to survive this hand-off.

**Prerequisites.**

- **The blocker is structural, not effort.** Adding an outcome field without a key produces
  a column nothing can ever populate. The key comes first.
- The join **will be partial** — a skill that forgets to pass the id degrades to today's
  behaviour rather than failing. That is the right degradation, and the datasheet must say
  so. Decide this before building, not after measuring.

**Steps.**

1. Mint a `run_id` in `runRecall`. Opaque, collision-resistant, no timestamp semantics
   leaking into it. Emit it in the JSON envelope — that is an addition to a documented
   contract (`recall.go:145`), so update the comment in the same edit.
2. Carry it on `telemetry.Event` (new field, `json:"run_id"`) and on `D1Pair`. Both are
   append-only JSONL; a new field is backward-compatible for readers, but check
   `pkg/dataset/read.go`'s strict reader — it refuses lines it cannot parse, by design.
3. Accept `--run-id` as an **optional** flag on `forge gate`. Absent → today's behaviour,
   no error.
4. Add the outcome record. Decide whether it is a new field on `D1Pair` (requires a
   rewrite of an append-only file — probably wrong) or a **separate outcome record** keyed
   by `run_id` that the export path joins. The second is almost certainly right; state the
   reason either way.
5. Thread the id through `skills/forge/SKILL.md`'s dispatch. This hop decides the item's
   real size.
6. Update every D1 datasheet: the corpus stops being purely supervision-on-its-own-output,
   *for the subset that joined*. Say what fraction joined; do not imply it is a census.

**Verification.**

- `go test ./pkg/dataset/... ./pkg/telemetry/... ./cmd/forge/...` → green.
- `forge recall` JSON carries `run_id`; `forge gate` without `--run-id` still exits 0 and
  behaves identically to today (assert this with a test — it is the degradation contract).
- `forge export-dataset --set d1` renders the outcome for joined records and omits it for
  unjoined ones, without a shape error from the strict reader.

**Done when** a recall call and the note write that followed it can be joined, and the
datasheet states the join rate honestly.

**Do not.** Do not add an outcome field before the key exists. Do not make `--run-id`
required — it would turn a skill omission into a failed write.

---

# B-034 — build D6 as an export view

**Why it's open.** ADDENDUM §D.1's table lists **six** datasets; Phase 6b built five. D6
"Code↔knowledge" — (repo symbol or module → the note explaining it) — was scoped out by
explicit decision, because ROADMAP and both phase prompts say five and only §D.1 says six,
and AUDIT never flagged the disagreement so precedence gives no ruling.

**Anchors.**

- `pkg/dataset/tier.go` — `Tier{Tag, Kind, Path}`, the registry `D1…D5`, `Tiers()`,
  `Enabled()`, `Append()`.
- `pkg/dataset/export.go:111` `resolve`, `:120` `checkFormat`, `:131` `prepare`,
  `:149` `since`, `:162` `anonymizeAll`, `:179` `commit`.
- `pkg/dataset/anonymize.go` — the note-path answer (hash the slug, keep the type).
- `pkg/report/knowledgemap.go` `RenderKnowledgeMap`, `cmd/forge/logback_map.go` — the
  existing (symbol → note) mapping.
- `pkg/coderef` — the citation registry. `.forge/code-index-<repo>.json` — the symbol table.

**Prerequisites.**

- **D6 is a derivation, not a capture tier.** D1–D5 each have a write path on a live
  command and accumulate forward, which is the whole argument for building capture early.
  D6 has no capture path and needs none — `forge logback` already builds exactly the
  mapping D6 wants. Nothing is lost by deriving it late.
- Ship it as `.forge/datasets/d6.jsonl`? **No.** An export *view*, not a sixth capture file.

**Steps.**

1. Resolve the struct problem first — it is the one real design question. `Tier` has a
   `Path` field and D6 has no file. Either make `Path` optional with an explicit
   `Derived bool`, or give `Tier` a loader function instead of a path. Whichever: `Tiers()`
   is iterated by both export and `dataset-stats`, so a derived tier must not break
   `dataset-stats`' per-file counting.
2. Add a `--set d6` case whose `loadTier` reads the code index and citation registry
   instead of a JSONL file.
3. **Refuse `--since` for d6** with a clear error. There is no per-record timestamp on a
   derived set, so silently ignoring the flag would report a filtered export that never
   filtered. Per Phase 6b's own precedent, an undefined `(set, format)` combination exits
   **2, not 3** — 3 promises "a real attempt was made"; this is rejected before a record
   is read.
4. Solve anonymisation, and expect it to be harder than D1–D5. **The symbol and module
   names are the feature, and they are also the most employer-identifying strings in the
   system.** `anonymize.go`'s note-path answer (hash the slug, keep the type) has no
   equivalent that leaves D6 useful. Do not assume the existing scrubber covers it. If no
   acceptable answer exists, the honest outcome is `--anonymize` refusing d6 rather than
   producing a corpus that looks scrubbed and is not.
5. Write the datasheet. State the derivation source and the anonymisation limit.
6. Record the five-vs-six decision's resolution in B-034 and, if D6 ships, in ROADMAP.

**Verification.**

- `go test ./pkg/dataset/...` → green, including a case asserting `--since` on d6 exits 2.
- `forge dataset-stats` still reports D1–D5 correctly with a derived tier in the registry.
- `TestAnonymizeRemovesEverySeededSecret`-style coverage for whatever D6's redaction is —
  that test is D-6's regression guard and the only thing that proves no secret escaped.
  Neither buffer-then-commit nor the per-record re-decode proves that.

**Done when** `--set d6` exports a (symbol → note) corpus with a datasheet that states its
anonymisation limit plainly, or the item closes with a written decision not to build it.

**Do not.** Do not add a sixth capture path. Do not let `--since` through silently.

---

# B-027 — decide the design-doc half — **CLOSED 2026-08-23**

Decided: **the docs were editable, and were edited.** All eight sites now show the
`-<repo>` suffix, plus three Turkish mirrors that described the entry as half-open.

The reasoning, because the decision is the deliverable and not the find-and-replace: the
"record, don't fix" rule and AUDIT §8.4's mechanism govern **decisions** — a doc line
superseded by a later ruling, where §8.4 is what a reader follows. B-027 is not one of
those. Nobody disagrees about the design; the docs name a file that has never existed on
disk under that name, so a reader following the mechanism here is sent to a doc that is
factually wrong about a path. That is a correction, not an override, and no §8.4 entry was
added because there was no decision to record. The roadmap ending at 6b removes the other
half of the rule's purpose — there is no in-flight phase to destabilise.

The two normative sites (ADDENDUM §B.6, DESIGN §15) carry a dated one-line marker so the
edit is traceable rather than a silent rewrite. The filename on disk is unchanged;
`cachePath` in `pkg/drift` stays the single place a name is constructed.

Full closure note: `docs/BACKLOG.md` B-027, final section.

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

## B-036 — a broad query links ten neighbours

**Measure before designing, and the measurement is the first real step — which is why
there are no implementation steps here rather than a plan with a guess at the top.** After
B-033's floor landed, three of §3.1's nine queries emit ten neighbours, the maximum
`forge recall` returns; two general Spring notes appear on every Spring question. The
obvious move — cap the count — truncates a score-ordered list arbitrarily and keeps the
same two notes at the top of it.

**Unblock condition:** a per-note "appears in N of M query results" column added to
`TestNeighbourFloorSweep`'s harness, which already stages the corpus. That answers whether
a note scoring on *every* query in an ecosystem should be admitted as a neighbour at all —
a document-frequency property the scorer computes for terms (§2.3.1) and not for notes.
Design after reading it.

**Do not respond to this by raising the floor.** Every floor in B-033's sweep that drops
those two notes also drops the Storybook family B-033 was opened to fix.

**Re-measured after B-032, 2026-08-23: unchanged in kind.** Still three of nine
calibration queries at the ten-neighbour cap, same two general Spring notes. B-032 moved
the underlying scores but not this shape — the unblock condition above is still what to
build.

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
