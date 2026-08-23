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

- **PLANNED** — open, workable, has a full six-field section in this file. Nine items.
- **NO STEPS** — open but not actionable by an implementation session: blocked on external
  observation, or a user decision, or "record, don't fix" by standing rule. Listed with
  its unblock condition instead of steps.
- **CLOSED / RECORDED** — done, or a measured deviation kept on purpose. One row, no
  section. Do not re-litigate these; the backlog carries the reasoning.

## Execution order

```
B-029  →  B-033  →  B-032  →  B-031  →  B-015  →  B-023  →  B-035  →  B-034  →  B-027
 lint     floor     denom     cover     imports   engine    run_id    D6       docs
```

**B-033 must land before B-032, and the ordering is load-bearing.** Both re-measure against
`cmd/forge/testdata/calibration.golden`. B-032 changes `blend`'s denominator for a large
fraction of the vault — every score moves. If B-032 lands first, B-033's re-derivation is
against a scale that just shifted again and has to be redone. If the order is reversed
anyway, B-032's plan must add a step re-deriving the floor a second time.

B-029 is first because it is independent of everything else and is the largest single item.
B-027 is last because it is documentation with no code consequence left.

---

## Index — all 35 items

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
| B-027 | `.forge/code-index-<repo>.json` naming | **PLANNED** (doc half) | [§B-027](#b-027--decide-the-design-doc-half) |
| B-028 | Hook path immediacy on deletion | CLOSED 2026-08-17 | — |
| B-029 | `errcheck` disabled tree-wide | **PLANNED** | [§B-029](#b-029--re-enable-errcheck) |
| B-030 | `dataset.capture` gates only two tiers | CLOSED Phase 6b | — |
| B-031 | Kafka/Testcontainers coverage miss | **PLANNED** | [§B-031](#b-031--the-coverage-side-of-the-scoring-surface) |
| B-032 | Untagged note escapes absent-term penalty | **PLANNED** | [§B-032](#b-032--activation-vs-absent-term-penalty) |
| B-033 | 0.30 neighbour floor on the old scale | **PLANNED** | [§B-033](#b-033--re-derive-the-neighbour-floor) |
| B-034 | D6 (code↔knowledge) not built | **PLANNED** | [§B-034](#b-034--build-d6-as-an-export-view) |
| B-035 | D1 has no outcome label | **PLANNED** | [§B-035](#b-035--mint-a-run_id-so-d1-can-carry-an-outcome) |

---

# B-029 — re-enable `errcheck`

**Why it's open.** `.golangci.yml` disables `errcheck` tree-wide so Phase 6 could land the
lint step without sorting ~95 findings. Triage item 1 landed 2026-08-22; the sweep and the
`disable:` block are untouched.

**Anchors.**

- `.golangci.yml` — the `linters.disable:` block, currently `- errcheck`, with a comment
  pointing at B-029. Its "~20 findings" claim is the stale undercount.
- Four `//nolint` precedents to match exactly — `pkg/engine/host.go:22`,
  `cmd/forge/check_test.go:131`, `pkg/similarity/similarity_test.go:134`,
  `pkg/telemetry/qhash_test.go:6`. Verified 2026-08-23: still exactly four in the tree.
- Signature-change sites — `pkg/drift/apply.go:109` (`stamp()`), callers at `:74` and `:91`;
  `pkg/drift/gitindex.go:48` (`persist()`), caller at `:38`.
- The `//nolint`-not-fix exemplar — `cmd/forge/index.go:207` (`tx.Rollback(); return err`).
- Already-documented deliberate ignore — `pkg/drift/demotions.go`'s `json.Unmarshal`.

**Prerequisites.**

- `errcheck` and `golangci-lint` are **not on PATH** in this environment (checked
  2026-08-23). Install both before starting, or the sizing below cannot be reproduced.
- Read B-029's two `###` sections first. The entry's original prescriptions are wrong in
  two places and the entry says so — see "Do not" below.

**Steps.**

1. Install the tools and re-measure, both lanes: `errcheck ./...` under `CGO_ENABLED=0`
   and `=1`. Expect ~95 raw and byte-identical output across lanes (the only tag-gated
   files, `pkg/codeindex/parse_{cgo,nocgo}.go`, have no findings). If the count has moved,
   record the new number in B-029 before proceeding — the plan below is sized against 95.
2. Run `golangci-lint run --no-config --enable errcheck ./...` once to get the *actual*
   post-exclusion count. B-029's "~37" is derived by hand from `EXC0001`, not observed;
   this step replaces a guess with a measurement.
3. Split the list into `_test.go` and production. Expect ~27 test / ~10 production. Do the
   **test files first** — they are fire-and-forget setup calls, mechanical, and finishing
   them shrinks the review surface for the interesting half.
4. Production, file by file. For each finding choose one of three, and write the reason:
   - ignoring is correct → `//nolint:errcheck // <lowercase reason>`, no trailing period,
     on the offending line, matching the four precedents' style exactly.
   - the error should be checked → check it.
   - the signature is the defect → change it (see step 5).
5. `pkg/drift/apply.go:109` and `pkg/drift/gitindex.go:48` need signature changes, not
   `//nolint`s. `stamp()` and `persist()` return nothing and their callers cannot check
   what does not exist. Size these two separately from the rest of the sweep.
6. **Re-size item 3 before touching it.** `pkg/codeindex/catfile.go:47`'s unchecked
   `cmd.Wait()` does *not* mean "non-zero exit after partial output reads as success" —
   that was re-traced 2026-08-22 and is wrong. `drainBlobs` (`:62-78`) returns the
   `ReadString`/`io.ReadFull` error, so partial output surfaces as EOF through
   `Build` (`pkg/codeindex/build.go:22`). What `cmd.Wait()` actually hides is the narrow
   case of **all replies delivered, then a non-zero exit**. Schedule accordingly; it is
   smaller than the entry originally implied.
7. Delete the `disable:` block from `.golangci.yml`. This is the whole close — no Makefile
   or workflow change, because `make lint` is gofmt + `go vet` and never ran errcheck.
8. Update B-029's status to closed, and correct its "~20" line so the number does not
   outlive the work.

**Verification.**

- `golangci-lint run ./...` → zero findings.
- `CGO_ENABLED=0 go build ./...` and `CGO_ENABLED=1 go build ./...` → both clean.
- `go test ./...` → 18 packages `ok`, both lanes.
- `go vet ./...` → clean.

**Done when** `errcheck` is off the `disable:` list and CI is green in both lanes.

**Do not.** Do not follow the entry's own item-1 prescription ("use the helper that
exists" / propagate through `commit`) — it was traced and rejected on 2026-08-22, because
propagating reaches `runRecall` (`cmd/forge/recall.go:77`), which exits 1 without emitting
candidates it already scored correctly. That item is closed; leave `refresh`/`writeRows`
alone. Do not add a blanket exclusion in an `issues:` block — the point of the sweep is
that each ignore carries its own reason.

**Bonus, not part of the sweep.** `pkg/codeindex/catfile.go`'s `blobSize` returns `!ok` for
any header that is not three fields ending in a blob size, and `drainBlobs` treats every
such line as the documented `"<name> missing"` case and `continue`s. Any other unexpected
line desynchronises the request/reply stream, after which blob bodies are parsed as
headers. Not an errcheck finding; belongs to whoever owns `catfile.go`.

---

# B-033 — re-derive the neighbour floor

**Why it's open.** B-008 changed the scale of two of four channels; the 0.30 neighbour
floor did not move with it, so an adjacent-topic query now verdicts CREATE with **zero**
neighbours — an orphan-creation path in a vault whose graph report already tracks 21
orphans of 94.

**Anchors.**

- `pkg/recall/doc.go:30` — `var DefaultThresholds = Thresholds{Answer: 0.85, Update: 0.55,
  Neighbour: 0.30}`. The Go-side default.
- `config/forge.config.example.md:47` — `neighbour_min_score: 0.30`. The packaged default.
- `pkg/config/types.go:61` — `NeighbourMinScore float64`.
- `cmd/forge/recall.go:57-58` — where the config value overrides `t.Neighbour`.
- `pkg/config/validate.go:125-127` — rejects `update_threshold < neighbour_min_score`.
- `pkg/recall/rank.go:102-105` — `Thresholds.Neighbours`, the band `>= Neighbour && < Update`.
- `pkg/recall/result.go:18,33` — neighbours populated on CREATE only.
- `cmd/forge/intent.go:48-52,62` — the second instance: a hardcoded `0.7` gate, with a doc
  comment already pointing here.
- `cmd/forge/calibration_test.go:93` `calibrationRow`, `:52` `goldenPath`, `:49` `-update`.

**Prerequisites — read these before choosing a number.**

- **The knob already exists.** This item is *not* "add a config value and pick a number";
  it is "re-derive the default." Confirmed 2026-08-23.
- **The default lives in two places** — `pkg/recall/doc.go:30` and
  `config/forge.config.example.md:47`. They must move together or they diverge silently,
  and the Go default is what every test and every un-configured install sees.
- **The new floor has a hard ceiling of 0.55.** `validate.go:125` enforces
  `update_threshold >= neighbour_min_score`, and B-008 forbids moving `update_threshold`.
- **Re-deriving against B-008's own nine queries is circular** — the number would be chosen
  to make those rows produce links. An honest derivation needs its own query set, one where
  the right neighbour set is known independently of what the scorer says.

**Steps.**

1. Add a neighbour column to `calibrationRow` (`cmd/forge/calibration_test.go:93`) — count,
   and the slugs. This is an addition to an existing row builder, not new machinery.
2. Re-record the golden: `go test ./cmd/forge -run TestCalibration -update`. Commit that
   diff **on its own**, before any threshold changes. It is the "before" measurement and
   it must exist as a separate commit or the later diff proves nothing.
3. Build a **separate** query set — not §3.1's nine. For each query, write down the notes
   that *should* be neighbours before running the scorer. Ten to fifteen queries over
   `examples/vault` is enough. Record it in the repo (a testdata file, not a comment) so
   the derivation is reproducible.
4. Sweep candidate floors across `[0.10, 0.55]` against that set. Report precision/recall
   of the neighbour set at each. Pick the value, and write the argument down — the number
   without the argument is tuning, which B-008 forbids by name.
5. Change both defaults together: `pkg/recall/doc.go:30` and
   `config/forge.config.example.md:47`. Check `config/presets/` too — eight packaged
   presets, any that restate the value must move with it.
6. Decide `intent.go`'s `0.7` **in the same session**. Two acceptable outcomes, one
   mechanism each: promote it to a config value (add to `config.Recall`, wire it in
   `cmd/forge/intent.go`, default it in the packaged config), or keep it hardcoded and
   replace its doc comment with the re-derived justification. Do not leave it as
   "re-derive both" without naming which.
7. Re-record the golden a second time and **paste the diff into the commit message**.

**Verification.**

- `go test ./cmd/forge -run TestCalibration` → passes against the re-recorded golden.
- `go test ./...` → green, both lanes.
- Manual: `forge recall "Storybook interaction testing with play functions"` against
  `examples/vault` emits a non-empty `neighbours` array. That row is the one that
  motivated the item — before: CREATE 0.217, both Storybook notes under 0.30, zero
  neighbours.

**Done when** the floor is re-derived from an independent query set, both default sites
agree, `intent.go`'s gate has a stated resolution, and the golden diff is in the commit.

**Do not.** Do not move `Answer` (0.85) or `Update` (0.55) — B-008 forbids it and this item
does not reopen it. `neighbour_min_score` is a different knob, named nowhere in that
prohibition, and it is the only one in scope. Do not derive the floor from the nine
calibration queries.

---

# B-032 — activation vs. absent-term penalty

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
commit, spec §2.5 matches the code, and B-008's false positive has not returned.

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
- B-033 and B-032 should land first — both move the scale this item measures against.

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

# B-027 — decide the design-doc half

**Why it's open.** `.forge/code-index-<repo>.json` is per-repo, correctly — repeatable
`--repo` means one shared name would let repo two overwrite repo one. The code side and one
wrong agent instruction were fixed 2026-08-21. The design docs still say the singular
`.forge/code-index.json`, **deliberately**, under the standing "record, don't fix" rule.

**Anchors — the sites still saying the singular name.**

- `docs/KNOWLEDGE-FORGE-ADDENDUM.md:247`, `:318` (§B.6)
- `docs/KNOWLEDGE-FORGE-DESIGN.md:714`, `:954` (§15)
- `docs/CLAUDE-CODE-PROMPT.md:208`, `:365`, `:458`
- `docs/ROADMAP.md:53`
- Already correct: `pkg/drift/gitindex.go`'s `persist` doc comment (carries the
  explanation), `pkg/codeindex/index.go:52`, `store.go:9`,
  `agents/forge-codebase-scout.md:33`.
- Out of scope by construction: `examples/vault/` — scrubbed vault content, a historical
  artifact, not documentation.

**Prerequisites.** CLAUDE.md's precedence rule says design docs are **not** edited
mid-flight; §8.4-style decisions are what a later reader follows. This item is therefore a
*decision* about whether the rule still applies now that the roadmap is complete.

**Steps.**

1. Decide, and write the decision into B-027: either the docs are now editable because the
   roadmap is done and there is no in-flight phase to destabilise, or the rule stands and
   this item closes as "recorded, permanent".
2. If editable: update the eight sites above to show the `-<repo>` suffix. Nothing else —
   this is a find-and-replace with a sanity read of each surrounding sentence, not a
   rewrite of §B.6.
3. Check `docs/tr/` for a mirror of these lines (it needed nothing for B-023's enum, but
   these are different sections).
4. Close B-027 either way. It is currently half-open with no remaining code consequence,
   which is the worst state for a backlog entry to sit in.

**Verification.** `grep -rn "code-index\.json" docs/` returns nothing, or returns only
lines a written decision says stay.

**Done when** the entry is closed with a stated decision, not left half-open.

**Do not.** Do not rename the file on disk to match the docs — the per-repo suffix is
required, and `cachePath` in `pkg/drift` stays the single place a name is constructed.

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
   B-031, B-032, B-033.
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
