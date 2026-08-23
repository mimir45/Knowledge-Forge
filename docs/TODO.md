# TODO — executable plans for the open backlog

`docs/BACKLOG.md` is the record of *why* each item exists. This file is the record of
*how to close it*. One section per workable item, each written so a future session can
open it cold, follow the steps, and stop when "Done when" is satisfied.

**Read `docs/BACKLOG.md`'s own entry before starting any item here.** These plans are
derived from it and deliberately do not restate its argument. Where a backlog entry has a
`###` closing section, that section is more current than its `Status:` line.

Written 2026-08-23 from a full scan of B-001…B-035; B-036 and B-037 added the same day as
B-033 and B-032 closed; B-038 added 2026-08-24 when B-031 closed and split it out. Anchors
were verified against the tree at that date; re-grep before trusting a line number.
**Closed and recorded items are dropped from this file
entirely, not just their plan sections** — `docs/BACKLOG.md` is the full census (every ID,
closed or open, with the reasoning) and the durable record of *why*; this file exists only
to track *how to close* whatever is still open. If an ID isn't below, it's done — look it
up in BACKLOG.

## Two status classes

Not every open backlog item can take a plan, and the absence of steps below is a decision,
not an oversight:

- **PLANNED** — workable, has a full six-field section in this file. Three now: B-015,
  B-034, B-035. B-029, B-027, B-033, B-032, B-023 and B-031 closed 2026-08-23/24; B-033's
  closure opened B-036 and B-032's opened B-037, both NO STEPS by their own argument —
  each names a measurement to run before any design is chosen, not a fix to apply.
- **NO STEPS** — open but not actionable by an implementation session: blocked on external
  observation, or a user decision, or "record, don't fix" by standing rule. Listed with
  its unblock condition instead of steps.

## Execution order

```
B-029  →  B-033  →  B-032  →  B-023  →  B-031  →  B-015  →  B-035  →  B-034  →  B-027
 done     done      done      done      done      head     run_id    D6       done
```

**B-029, B-027, B-033, B-032, B-023 and B-031 all closed** (the first four 2026-08-23,
B-023 2026-08-24, B-031 2026-08-24). B-031 closed with no code change — see its BACKLOG
entry's closing section: neither of its two shapes survives measurement against the corpus.
**B-015 is now the head.**

B-029 was first because it is independent of everything else and was the largest single item.
B-027 was last because it is documentation with no code consequence left.

---

## Index — 10 open items (28 closed/recorded IDs live in BACKLOG only)

| ID | Subject | Class | Section |
|---|---|---|---|
| B-003 | Repo directory still named `TIL` | NO STEPS — user decision | [below](#no-steps) |
| B-004 | Module path has no VCS host prefix | NO STEPS — deferred by decision | [below](#no-steps) |
| B-012 | `code_refs` has no live producer | NO STEPS — blocked on packaging | [below](#no-steps) |
| B-015 | `CodeGroup.DependsOn` never populated | **PLANNED** | [§B-015](#b-015--populate-codegroupdependson) |
| B-025 | `PostToolUse`/WebFetch payload shape | NO STEPS — **BLOCKED** | [below](#no-steps) |
| B-034 | D6 (code↔knowledge) not built | **PLANNED** | [§B-034](#b-034--build-d6-as-an-export-view) |
| B-035 | D1 has no outcome label | **PLANNED** | [§B-035](#b-035--mint-a-run_id-so-d1-can-carry-an-outcome) |
| B-036 | Broad query links ten neighbours | NO STEPS — measure first | [below](#no-steps) |
| B-037 | Intent gate FIRE/QUIET margin now negative | NO STEPS — measure first | [below](#no-steps) |
| B-038 | `bodyPass` window allocated by path, not relevance | NO STEPS — measure first | [below](#b-038--bodypasss-top-20-window-is-allocated-by-path-not-by-relevance) |

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
