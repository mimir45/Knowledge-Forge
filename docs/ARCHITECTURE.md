# Knowledge Forge — Architecture

> This document answers the **how** question. For usage, see
> [`USAGE.md`](USAGE.md).

---

## 1. Overview

The system has four layers. Read them bottom-up — none of the upper layers mean
anything without the ones below, but not the other way around.

```
┌──────────────────────────────────────────────────────────────────────┐
│  LAYER 3 — Claude Code integration                                   │
│  .claude-plugin/  hooks/  skills/  agents/                           │
│  (lifecycle hooks, slash commands, product agent specs)              │
└──────────────────────────────────────────────────────────────────────┘
                              │  exec
┌──────────────────────────────────────────────────────────────────────┐
│  LAYER 2 — CLI:  cmd/forge  (20 subcommands)                         │
│  flag parsing · orchestration · output formatting · exit codes       │
└──────────────────────────────────────────────────────────────────────┘
                              │  import
┌──────────────────────────────────────────────────────────────────────┐
│  LAYER 1 — Libraries:  pkg/*  (18 packages)                          │
│  all business logic; each one independently testable                │
└──────────────────────────────────────────────────────────────────────┘
                              │  read/write
┌──────────────────────────────────────────────────────────────────────┐
│  LAYER 0 — Data                                                      │
│  vault/*.md (source of truth) · code repos (git) ·                   │
│  .forge/*.db + .forge/*.json (derived cache)                         │
└──────────────────────────────────────────────────────────────────────┘
```

**Instructive note:** the real test of this separation is that no file under
`cmd/forge` contains business logic. `cmd/forge/drift.go` is 250 lines long but never
makes a single verdict decision itself; it reads flags, calls `pkg/drift`, and prints
the result as JSON or text. That's why `pkg/drift`'s tests run without a CLI, and the
CLI's tests never re-test business logic.

---

## 2. Package map and the import DAG

There are 18 packages. The dependencies between them are **acyclic** and surprisingly
sparse:

```
cmd/forge ──────────────► (all 18 pkgs) + profiles

pkg/config    ──► config (embed)
pkg/vault     ──► pkg/config, references
pkg/dataset   ──► pkg/vault
pkg/scrub     ──► pkg/vault
pkg/engine    ──► pkg/config
pkg/drift     ──► pkg/codeindex, pkg/coderef, pkg/vault
pkg/qualitygate ──► pkg/config, pkg/recall, pkg/similarity, pkg/vault, references
pkg/report    ──► pkg/drift, pkg/gitsig, pkg/graph, pkg/linkcheck, pkg/similarity

leaves (import no internal package):
  codeindex · coderef · gitsig · graph · linkcheck · recall
  sentinel  · similarity · store · telemetry
```

Three things to read off this graph:

**(a) Ten packages are leaves.** `pkg/recall` imports nothing — not vault, not store.
Its input is `[]recall.Doc`, its output `recall.Result`. That turns the scoring
algorithm into a pure function, which is why `pkg/recall/score_test.go` can run
without ever standing up a vault.

**(b) `pkg/report` does NOT import `pkg/codeindex`** — this is a written rule, not an
accident. The reason is cgo: `pkg/codeindex` is the one cgo package (go-tree-sitter).
If the report layer imported it, the pure-Go build lane would break. Instead,
`cmd/forge/check_codebase.go` defines a `symbolFinder` interface and wires the
concrete type together in the CLI.

**(c) `pkg/drift` looks like it imports the cgo-carrying `pkg/codeindex`**, but stays
pure Go thanks to its own `Source` interface:

```go
type Source interface {
    At(...)        // file content at a commit
    RevBefore(...) // the previous revision
    Head(...)      // HEAD sha
    Find(...)      // symbol lookup
    ResolveAt(...) // resolution at a specific era
}
```

That's exactly what the interface is for: drift's logic is testable in pure Go, and
the concrete implementation carrying cgo stays outside.

### 2.1 Each package's job, in one sentence

| Package | Job |
|---|---|
| `pkg/vault` | Frontmatter + markdown AST (goldmark), mtime-cached. Note read/write/validate/quarantine. |
| `pkg/recall` | Deterministic question → note scoring. Zero model calls. |
| `pkg/similarity` | Hand-rolled MinHash + LSH banding. Near-duplicate detection. |
| `pkg/graph` | Note link graph: components, hubs, orphans, centrality. |
| `pkg/codeindex` | go-tree-sitter (Java + TypeScript). **The one cgo package**, build-tag gated. |
| `pkg/coderef` | Extracts and resolves code citations from note bodies and frontmatter. |
| `pkg/gitsig` | Churn, ownership, co-change coupling — via the git CLI (not go-git). |
| `pkg/drift` | **The key package.** Decay detection via AST comparison, not line diffs. |
| `pkg/linkcheck` | HTTP HEAD on sources, cached, rate-limited. |
| `pkg/report` | Renders analyses to markdown. Must not import `pkg/codeindex`. |
| `pkg/store` | SQLite (`modernc.org/sqlite`). Derived cache only, except the budget table. |
| `pkg/config` | The four-layer config chain. |
| `pkg/engine` | none/host/api/advisor backends, per-stage selection + fallback, engine_trail. |
| `pkg/qualitygate` | The seven gates (§8) + `Run`/`Report` orchestration + `_inbox/` quarantine. |
| `pkg/sentinel` | Id-based begin/end managed comment blocks; `Upsert`/`UpsertBefore`/`Remove`. |
| `pkg/scrub` | Redacts secret/PII-shaped content from a vault copy; fails closed. |
| `pkg/dataset` | D2–D4 training-pair capture (JSONL). |
| `pkg/telemetry` | The `ask` event; sha256 topic hash, never raw question text. |

---

## 3. Two build lanes

This is the most commonly misunderstood part of the architecture. There are **two
distinct build modes**, and `pkg/codeindex` is what separates them.

```
LANE A — pure Go (default, shipped)
  CGO_ENABLED=0 go build ./...
  → NO tree-sitter, code indexing disabled
  → cross-compiles to six targets
  → make build

LANE B — cgo
  CGO_ENABLED=1 go build -tags codeindex ./...
  → go-tree-sitter compiled in, Java + TypeScript parsed
  → needs a host toolchain, does not cross-compile
  → make full
```

The mechanism is the classic two-file Go pattern:

- `pkg/codeindex/parse_cgo.go` — the real tree-sitter parser, build-tag gated.
- `pkg/codeindex/parse_nocgo.go` — same signatures, empty/degraded behavior.

That way, code importing `pkg/codeindex` **compiles** in both lanes; only the
behavior differs. The relevant invariant:

> `CGO_ENABLED=0` for every package except `pkg/codeindex` (go-tree-sitter needs cgo).

`make test` deliberately enforces this from both directions: first
`CGO_ENABLED=1 go test ./...` (with the real parser), then
`CGO_ENABLED=0 go build ./...` (to prove the pure lane still compiles). Running only
one silently breaks the other.

Measured state: all 18 packages report `ok`, green under both `CGO_ENABLED` values.
(`config`, `profiles`, `references` are data-only packages — no test files.)

---

## 4. Config: the four-layer chain

```
1. $FORGE_CONFIG                          (highest precedence)
2. <project>/.forge.config.md
3. ~/.forge/forge.config.md
4. config/forge.config.example.md         (embedded in the binary, lowest)
```

The rules, in `pkg/config/load.go` and `merge.go`:

- **A missing optional layer is skipped.** But **a missing `$FORGE_CONFIG` is an
  error** — because the user named it explicitly. Silently ignoring it would let
  someone run against the wrong config without noticing.
- **Maps merge key by key.** Scalars and lists are replaced wholesale. So overriding
  `engines.budget.api_usd_per_day` doesn't require rewriting the whole `engines`
  block; but adding one element to the `check.reports` list requires writing out the
  entire list.
- Config is a **markdown file** with YAML frontmatter. `frontmatter()` strips the BOM
  and normalizes CRLF — so a config coming from Windows doesn't silently fail to
  parse.
- `decode()` round-trips through YAML, *"so the struct tags stay the single
  definition of the schema"* — meaning the schema isn't defined in two places.
- `expandHome()` resolves `~/`.

**Who writes it?** Only `forge init`. And only to two files:
`~/.forge/forge.config.md` and `<vault>/profiles/me.md`. **Never**
`config/forge.config.md` — that stays a packaged template. Violating this separation
would mean the next `go build` overwrites the user's own settings.

The schema is the **union** of the engine/config blocks and the pipeline keys those don't
restate.

`forge config --layers` prints which layers are present; `forge config --json`
prints the resolved result. This is the first command to run for a "why this
value?" question.

### 4.1 Main blocks of the config tree

`config/forge.config.example.md` (215 lines) carries the complete default tree:

| Block | Content |
|---|---|
| `vault_path`, `repo_path: auto`, `paths` | Where to operate. |
| `trigger.mode: ask` | Whether to ask at the "explain X" moment, or go automatic. |
| `recall` | Lexical; thresholds **0.85 / 0.55**, duplicate 0.30. |
| `freshness_days` | By type: concept 365, howto 180, api 90, pattern 365, pitfall 365, incident 0, decision 0. |
| `engines` | default `host`; api `anthropic`/`claude-sonnet-5`/`ANTHROPIC_API_KEY`; advisor `claude-opus-5` mode `critique`; local disabled; budget 2.00/1.00 USD; `on_exhausted: queue`; `routing.advisor_when`. |
| `pipeline` | Nine stages (below). |
| `research`, `verify`, `write` | Research depth; `run_code: auto`, `duplicate_threshold: 0.40`; language `en`, 1200 words, mermaid. |
| `static` | code_index (java/kotlin/python/typescript), git_signals, `cache_ttl_days: 30`, drift, linkcheck, logback (`inline_markers: false`). |
| `check` | `schedule: "0 9 * * MON"`, nine reports, churn 90 days, duplicate 0.40. |
| `garden`, `dataset`, `telemetry` | Gardening; D1–D5 capture + `anonymize_on_export`; telemetry `local` scope. |

The prose section also records three groups **deliberately left in code**:
`pkg/vault`'s `excludedPrefixes`/`hubNames`, `pkg/report/duplicates.go`'s
`specThreshold = 0.85`, and `$HOME/.forge/bin` in the Makefile. These weren't
promoted to config because they're not a user decision — they're a spec constant.

---

## 5. The engine layer: tiers, chains, and budget

### 5.1 Four tiers

| Tier | Implementation | Note |
|---|---|---|
| `none` | `pkg/engine/none.go` | No model. |
| `host` | `host.go` | The Claude Code session itself. |
| `api` | `api.go` + `api_provider.go` | The Anthropic API. |
| `advisor` | `advisor.go` | Critique-only. |
| `local` | *alias* | **Not a fifth engine** — "`api.go` under a different `base_url`". Maps to `TierAPI`. |

### 5.2 The selection algorithm

`Resolve()` in `pkg/engine/select.go` walks the stage's chain:

```go
func chain(cfg, stage, st) []string {
    if st.Engine != "" { out = append(out, st.Engine) }
    else if cfg.Engines.Default != "" { out = append(out, cfg.Engines.Default) }
    if st.Fallback != "" { out = append(out, st.Fallback) }
    if st.Then != "" { out = append(out, st.Then) }
    if len(out) == 0 { out = []string{"none"} }
    return out
}
```

An instructive detail, in the source's own comment: *"an unconfigured stage is not a
claim of being locked to none — it's silence, and it's `cfg.Engines.Default` that
fills it."* The difference between treating an empty field as "none" versus as "the
default" is a judgment call about the config author's intent, and it's made
explicitly here.

`Resolve` returns not just the winning name but a **human-readable reason** too. That
lets `forge engine select --json` answer "why did offline fall back to none" —
not just say that it did. If no candidate qualifies:
`"no candidate in the chain was available; degrading to none"`.

### 5.3 Locked stages — defense in two layers

Three stages accept nothing but `none`: **`recall`, `write`, `index`**.

The check happens in two places:

1. **At load time** — `pkg/config/validate.go`, `LockedStageError`.
2. **At selection time** — `pkg/engine/select.go`, `checkLocked()`.

The second one might look redundant, but its source comment states the reason:
`checkLocked` looks at `Fallback` and `Then` as well as `Engine` — *"a tamper hiding
behind `pipeline.write.fallback` instead of `pipeline.write.engine` must be caught
here too, or this layer is decorative."*

The error message is instructive too, not a silent override:

```
engine: pipeline.write: "api" is not allowed — [recall write index] are locked to
"none" (T0 static core)
```

### 5.4 Budget accounting

- Counters live in **SQLite** (`pkg/store/budget.go`), under `.forge/`.
- It **survives** `forge reindex` — the only exception to the cache rule. Otherwise
  reindex would be a budget-reset trick.
- `Exhausted()` distinguishes "no budget today" from "no tier measured here at all";
  `on_exhausted: queue` wouldn't behave correctly without that distinction.
- `on_exhausted` defaults to **`queue`**. The three accepted values are
  `queue | degrade | stop`: `queue` stamps `pending_advisor: true` for processing on
  the next budget cycle and falls through to `none`; `degrade` is the same as
  today's silent fallback to `none`, because that's the honest reading of the word
  anyway; `stop` exits non-zero for real — `pkg/engine` itself never reads
  `OnExhausted`, that distinction is made one layer up, in `cmd/forge`.

### 5.5 engine_trail

`pkg/engine/trail.go` records which stage went to which tier in the note's
frontmatter. `forge engine record` writes it and **refuses to record against a
locked stage** (`isLockedStage`). That way, "how was this note produced?" is
answerable from the note itself.

---

## 6. Anatomy of the recall engine

`pkg/recall/score.go` (246 lines) is the most heavily thought-through file in the
system.

### 6.1 Four channels

```go
const (wTitle = 0.4; wTags = 0.3; wStack = 0.2; wBody = 0.1)
```

### 6.2 Blending — a mean over active channels

```go
func blend(chs []Channel) (score float64, matched []string) {
    num, den := 0.0, 0.0
    for _, c := range chs {
        if !c.Active { continue }
        num += c.Weight * c.Value
        den += c.Weight
        if c.Value > 0 { matched = append(matched, c.Name) }
    }
    if den == 0 { return 0, matched }
    return num / den, matched
}
```

A channel with `Active == false` drops out of both the numerator and the
denominator. `weighted()` returns `ok=false` on an empty denominator, and the
channel is disabled.

### 6.3 IDF

```go
const idfCap = 3.5

func idf(df, n int) float64 {
    if df <= 0 || n <= 0 { return 0 }
    return math.Min(math.Log(1+float64(n)/float64(df)), idfCap)
}
```

`log(1+n/df)`, not `log(n/df)` — the latter gives exactly zero whenever a term
appears in every note, wiping the term out of the equation entirely.

### 6.4 Title: F₂

```go
func f2(hits, queryTerms, titleTokens int) float64 {
    if hits == 0 || queryTerms == 0 || titleTokens == 0 { return 0 }
    p := float64(hits) / float64(titleTokens)
    r := float64(hits) / float64(queryTerms)
    return 5 * p * r / (4*p + r)
}
```

### 6.5 Body

`bodyChannel` saturates every term after **3 repeats**. A note that mentions a word
50 times isn't more relevant than one that mentions it 3 times.

### 6.6 The calibration gap IDF weighting found and fixed

When IDF weighting first shipped, it didn't fix the case it was aimed at. Why: terms
carrying a question's actual meaning were being filtered out of the denominator
whenever no note carried them — so "nobody knows this" was behaving like
"unimportant." The fix needed two changes: the `inVocab` filter switched sides to
apply to the `--stack` hint instead of the question, and a missing term's weight
became the average of the present terms' weights instead of zero.
**Don't nudge the thresholds in response to this** — the fix came from remeasuring
and re-deriving the §3.1 calibration table, not from tweaking a constant.

---

## 7. Drift: git-anchored decay detection

### 7.1 Contract

```go
type Verdict string
const (OK; Repaired; Suspect; Broken; Skipped)

func (f Finding) Demoting() bool { return f.Verdict == Broken }

type Changed struct {
    Touched map[string]bool
    Deleted map[string]string  // repo-relative path -> repo name
}

type Opts struct{ Deep bool }
```

### 7.2 Why the git object store, not the working tree

Drift reads the `HEAD` tree and `--since-commit <sha>`. It **never** looks at the
working tree. Why:

- A half-written file is not a claim. A note shouldn't get demoted over an edit
  that hasn't even been saved yet.
- Determinism: (note refs, tree state) → verdict is a pure function. The working
  tree would make that function time-dependent.
- Symmetry: `git revert` → same tree → same verdict → the note is restored.

The only state stored under `.forge/` is the confidence value from before a
demotion — a restore target, never a verdict input.

### 7.3 Two paths: the hook path and the full sweep

| | Hook path (`forge drift`, default) | Full sweep (`forge check`) |
|---|---|---|
| Trigger | post-commit / post-merge / post-checkout | Weekly |
| Scope | files changed since `--since-commit` | Whole vault |
| Registry | from the `HEAD` tree | `HEAD` + (with `--deep`) historical `ResolveAt` |
| Calls `drift.Apply`? | yes, with `--apply` | **No** |
| Budget | **< 100 ms** (the binding constraint) | < 10 s |

This distinction matters for understanding how a reference to a deleted file ends up
with a `Broken` verdict:

Because `registryOf` always builds the registry from the current `HEAD` tree, a
reference to a fully deleted file could never come back `Broken` — it stayed
`Skipped` forever. The fix has two parts: in the full sweep (`opts.Deep`, no
`--since-commit`), fall back to a verified-era `ResolveAt` scan — but since the full
sweep in `forge check` never calls `drift.Apply`, that alone only makes `drift.md`
correct without ever auto-demoting anything. The automatic demotion happens on the
hook path: the hook already computed a cheap gate
(`coderef.ChangedFilesStatus`); switching from `--name-only` to `--name-status`
made that gate carry **deletion evidence** (`drift.Changed{Touched, Deleted}`). Now
an `Unresolved` reference that matches a deletion in the same commit gets a
`Broken` verdict **immediately** under `--apply` — no `--deep` or historical
registry scan needed. This is the only automatic demotion path a deleted-file
reference has.

A critical architectural detail: a miss on the hook path that doesn't match
**produces no finding at all**, never a `Skipped` one. Otherwise an unrelated later
commit could still flip a still-broken note back to `high`.
`TestRollbackSymmetryOnDeletion` pins this down.

This approach has one known limitation: basename collisions.

### 7.4 AST comparison, not line diffs

A file's 200 lines might have changed while the cited symbol was never touched; or a
single line changed and that line is the method signature. A line diff can't tell
the two apart — AST comparison can. This is where `pkg/codeindex`'s tree-sitter
comes in — and exactly why it needs cgo.

---

## 8. Quality gates and quarantine

The order of `Run` in `pkg/qualitygate/gate.go` is fixed:

```
schema → citation → code → freshness → antislop → link → duplicate
```

`Remedy` is an iota but serializes to JSON as a **name, not an ordinal**
(`MarshalJSON`) — so inserting a remedy in a future release doesn't break saved
reports.

```go
func blocksWrite(r Remedy) bool  // None, DelegateToLibrarian, SwitchToUpdate → false
```

A blocking failure → `_inbox/` quarantine, `confidence: low`.
`cmd/forge/gate.go` calls `reindexAfterQuarantine` after quarantining: the index
needs to reflect that the note is now in `_inbox/`.

The `code` gate has an interesting subsystem: `pkg/qualitygate/compile*.go` —
`compile_bash.go`, `compile_java.go`, `compile_ts.go`. These call the system
toolchain, in a **throwaway directory**. The cost is dominated by toolchain startup,
not gate logic — see §13 for the measured figures. `tsc` isn't installed in this
environment so the TypeScript lane went untested;
`TestCompileTSSkippedWhenToolchainAbsent` covers the no-toolchain path instead.

The six in-process gates besides `code` are expected to be far cheaper than invoking a
compiler, but **that has never been measured**: `pkg/qualitygate` contains no benchmark
and is absent from `make bench`'s package list. An earlier revision of this document
quoted ~0.13 ms; that number could not be reproduced from anything in the repo.

---

## 9. Sentinel: idempotent managed blocks

A 30-line package, but all of `logback` rests on it.

```go
type Style struct{ Open, Close string }

var (
    Markdown = Style{"<!--", "-->"}
    Slash    = Style{"//"}
    Hash     = Style{"#"}
)
```

Markers render as `forge:<id>:begin` / `forge:<id>:end`. The contract in one
sentence: *"everything outside a block's own begin/end pair is left byte-for-byte
untouched."*

`Upsert` / `UpsertBefore` / `Remove` — idempotent and **position-independent.** So
if the user moves the block elsewhere in the file, the next `logback` run keeps
writing it there instead of creating a second copy.

This is the correct solution to "writing generated content into the user's file":
instead of owning the whole file, you own one named section of it.

---

## 10. Three main data flows

### Flow A — Question → Note

```
user: "explain X"
   │
   ├─ [hook] UserPromptSubmit → forge intent
   │     read prompt from stdin → recall → if score >= 0.50
   │     print the best hit as additionalContext
   │     budget < 50 ms · fail-silent · exit 0
   │
   ▼
forge recall --explain
   │  load notes via pkg/vault (warm from SQLite cache)
   │  score with pkg/recall (4 channels, IDF, F₂)
   │  verdict: reuse | update | create
   │  pkg/telemetry: ask event (topic hash only)
   ▼
verdict = create
   │
   ├─ pipeline: intake → plan → research → synthesize → verify → write
   │  (each stage picks its own engine tier from pkg/engine)
   ▼
draft note
   │
   ▼
forge gate --file draft.md
   │  seven gates, in order
   │  blocking failure? → _inbox/, confidence: low → reindex
   │  clean? → notes/<type>/<slug>.md
   ▼
forge index → _index.md + SQLite
```

### Flow B — Commit → Drift

```
git commit in a code repo
   │
   ▼
.git/hooks/post-commit  (installed via scripts/install_drift_hook.sh)
   │
   ▼
forge drift --repo myapp:<path> --since-commit <sha> --apply
   │
   ├─ coderef.ChangedFilesStatus (--name-status)
   │     → drift.Changed{Touched, Deleted}          ← cheap gate
   │
   ├─ no reference is affected: exit (the common case)
   │
   ├─ for affected references:
   │     registryOf(HEAD tree) → coderef resolution
   │     AST comparison (pkg/codeindex, tree-sitter)
   │     → OK | Repaired | Suspect | Broken
   │
   ▼
Broken → demote
   │  previous confidence saved to .forge/ (restore target)
   │  the note's confidence is lowered
   │
   ▼
git revert → same tree → same verdict → the note is restored symmetrically
```

Budget **< 100 ms**. This is the binding latency constraint for the whole project —
it's the only thing running on the git hook path. **The current build does not meet
it**: measured at 151 ms median / 208 ms p95 (§13).

### Flow C — Weekly check

```
forge check   (schedule: "0 9 * * MON", but NO scheduled automatic mutation)
   │
   ├─ collectVault: notes, graph, similarity, git history, budget snapshot
   │
   ├─ render nine reports → <vault>/reports/
   │     coverage · staleness · duplicates · orphans · gaps
   │     graph-health · churn · deadlinks · drift
   │     (+ cost.md, codebase.md, moc/weekly/YYYY-WW.md)
   │
   ├─ weekly rollup: week-over-week delta via .forge/weekly-stats.json
   │
   ├─ aiPass (optional): draft refresh · duplicate merge · ADR stub suggestions
   │     — prints INSTRUCTIONS only, never writes itself
   │
   └─ drainAdvisorQueue: process queued notes if budget has rolled over
```

Measured: **390 ms warm / 930 ms cold**, budget 10 s. The nine reports render
deterministically — six consecutive runs, md5-identical.

Measured against the real vault: **9 notes** cite changed code (2 broken,
7 suspect), out of 140 references; **21 of 94 notes** are orphans; **23 graph
components**; **3 duplicate pairs** ≥ 0.40; **39 of 41 stacks** covered.

`writeReport` only writes when content actually changed (`writeIfChanged`) — so an
unchanged report's mtime doesn't move, and the vault's git diff stays clean.

---

## 11. The Claude Code integration layer

### 11.1 Plugin manifest

`.claude-plugin/plugin.json` — name `forge`, displayName "Knowledge Forge", v0.1.0,
MIT, repo `github.com/mimir45/Knowledge-Forge`.
`.claude-plugin/marketplace.json` — the marketplace listing.

These two close the **packaging gap** that didn't exist before Phase 6: the
top-level `agents/` directory and `hooks/hooks.json` are auto-discovered once the
plugin is installed.

### 11.2 Four lifecycle hooks

| Event | Matcher | Command | Timeout |
|---|---|---|---|
| `SessionStart` | — | `hooks/session-context` | 5 s |
| `UserPromptSubmit` | — | `hooks/user-prompt-intent` | 2 s |
| `SessionEnd` | — | `hooks/session-end-capture` | 10 s |
| `PostToolUse` | `WebFetch` | `hooks/post-tool-cache-source` | 10 s |

Paths are `"${CLAUDE_PLUGIN_ROOT}"/hooks/...` — moved off Phase 6's hardcoded
absolute paths to this.

**Shared contract:** every one of them is fail-silent, **always exits 0.** A hook
must never be able to break a session or a commit.

**The resume trap** (documented at the source): `SessionStart` re-runs on
`--continue`/`--resume` (with `source: resume`) — expected, and its output is
idempotent and cheap. But **every other hook's output is replayed from the saved
transcript**, not re-run. Consequence: seeing a stale recall hit on resume is not a
bug in `forge intent`, it's expected behavior. That's why no hook carries
time-sensitive work.

### 11.3 Git hooks

There are three separate hook families, and they should not be confused:

| Hook | Where | What it does | Install |
|---|---|---|---|
| `vault-post-commit` | in the **vault** repo | `forge capture` — D3 training-pair harvest | `scripts/install_vault_hook.sh` |
| `code-post-commit` / `-merge` / `-checkout` | in **code** repos | `forge drift` | `scripts/install_drift_hook.sh` |
| the four in `hooks/hooks.json` | in a Claude Code session | context/intent/capture/cache | plugin install |

The vault hook calls **`~/.forge/bin/forge`** — not the repo's own build output.
The absolute path is pinned at `<vault>/.forge/forge-bin`; `$FORGE_BIN` overrides
it. This is a **copy**: it needs rebuilding whenever `pkg/dataset` or
`cmd/forge/capture.go` changes:

```bash
CGO_ENABLED=0 go build -o ~/.forge/bin/forge ./cmd/forge
```

Because the hook prints nothing by design, a stale binary is **silent**.
Diagnostic: `<vault>/.forge/capture.log`.

### 11.4 Four skills

`skills/forge/`, `skills/forge-init/`, `skills/forge-check/`, `skills/forge-stats/`.
Skills ask questions and shell out to the binary — they carry no business logic.

### 11.5 Four product agents

`agents/forge-researcher.md`, `forge-codebase-scout.md`, `forge-verifier.md`,
`forge-librarian.md`.

`forge-librarian`'s prompt stamps **`Forge-Write: true`** on every commit it
authors. This matters: without it, `pkg/dataset` would record the agent's own
output as a *human correction*, and the training data would be contaminated.
`pkg/dataset/d3_forge_write_test.go` pins the guard both ways.

**Recorded gap:** these must not be confused with the **workflow** agents under
`.claude/agents/`. The ones under `.claude/agents/` (`finder`, `executor`,
`explainer`, `vault-analyst`, `doc-auditor`, `cross-checker`) exist to *build* this
project. The ones under `agents/` are the *product's own* agents.

---

## 12. Invariant table

Each row is stated in a different document, and each one is easy to violate by
accident.

| # | Invariant | Enforced where |
|---|---|---|
| 1 | The T0 static core makes zero model calls. | `cmd/forge/main.go` doc comment; code review |
| 2 | `recall`/`write`/`index` accept only `none`; otherwise **refuse to start with a clear error**. | `pkg/config/validate.go` + `pkg/engine/select.go:checkLocked` |
| 3 | Drift is git-anchored; never on file save, never against the working tree. Verdicts are a pure function; a revert restores symmetrically. Demotion history lives in `.forge/`, never in the note body. | `pkg/drift`, `rollback_test.go` |
| 4 | `CGO_ENABLED=0` for every package except `pkg/codeindex`. | `parse_cgo.go`/`parse_nocgo.go`, `make test` |
| 5 | Markdown is the single source of truth; SQLite is derived; `forge reindex` rebuilds it entirely. | `cmd/forge/index.go:cmdReindex` |
| 6 | `pkg/similarity` is hand-rolled MinHash + LSH. **No embeddings.** | `pkg/similarity/*` |
| 7 | Never auto-mutate the vault on a schedule; gate failures go to `_inbox/` with `confidence: low`. | `pkg/qualitygate/quarantine.go` |
| 8 | Code verification compiles in a throwaway directory, never in the user's project. | `pkg/qualitygate/compile.go` |
| 9 | The advisor tier is critique-only: disputed claims + a patch, never a rewrite. | `pkg/engine/advisor.go` |
| 10 | Telemetry logs topic + hash. Never raw questions, code, or file contents. | `pkg/telemetry/qhash.go` |
| 11 | CLI only for v1. Don't build a daemon on speculation — measure first. | (measured; not needed) |
| 12 | `pkg/report` must not import `pkg/codeindex`. | import DAG |
| 13 | Scrub fails closed; a note that can't be re-validated cancels the whole run. | `pkg/scrub/scrub.go` |

---

## 13. Latency budgets and measured values

Two different things live in this section and they must not be confused.

**Budgets** are targets the design commits to. **Measurements** are what a specific
build did on a specific machine, and they are only worth printing next to the harness
that reproduces them.

The figures below come from the external CLI campaign in
`docs/testing/campaign-2026-09-03/` — a serial pass on an idle Apple M4 against a
94-note vault, 10 runs per command, with the raw records under `runs/*.jsonl`. Every
number is re-derivable with `bash docs/testing/campaign-2026-09-03/aggregate.sh`.

| Operation | Budget | Cold | Warm median | Warm p95 | Verdict |
|---|---|---|---|---|---|
| `forge drift` | **< 100 ms** | 164 ms | **151 ms** | **208 ms** | **OVER BUDGET** — and it is the binding constraint |
| `forge index` | < 200 ms | 116 ms | 127 ms | 140 ms | under budget |
| `forge check --offline` | < 10 s | 206 ms | 147 ms | 160 ms | comfortably under |
| `forge recall` | — | 188 ms | 57 ms | 114 ms | — |
| `forge session-context` | < 200 ms | 40 ms | 36 ms | 38 ms | under budget |
| `forge intent` | **< 50 ms** | 128 ms | 49 ms | 58 ms | median just inside; **p95 and cold are over** |
| `forge verify-code` bash | — | 48 ms | 46 ms | 90 ms | dominated by toolchain startup |
| `forge verify-code` java | — | 1238 ms | 1030 ms | 1373 ms | same |
| `forge verify-code` ts | — | — | — | — | not measured; `tsc` absent here |
| `qualitygate.Run` (6 gates, excl. `code`) | — | — | — | — | **no benchmark exists** |

Two things this table says that the previous revision did not:

- **`forge drift` misses its own budget** — tracked as **B5** in
  [`docs/testing/campaign-2026-09-03/report.md`](testing/campaign-2026-09-03/report.md).
  `MANIFESTO.md` §7 says missing this budget is a bug, and that reading stands: this is
  an open defect, not an accepted cost. It runs on the git-hook path, so this is the one
  number that matters most. `--since-commit` is documented as the cheap gate, but it
  measured 147–148 ms against the full run's 151 ms — so whatever dominates the cost is
  not the citation scan.
- **`forge intent` meets 50 ms only warm.** A first invocation in a session costs 128 ms.
  Both hook commands stay far inside the 2 s timeout in `hooks/hooks.json`, so nothing
  user-visible breaks — but the budget as written is not met on the path that matters,
  the first prompt of a session.

`forge intent` is able to approach its budget at all only because it reuses `forge
recall`'s already-warm SQLite cache: an architectural decision (the derived cache) is
what makes the latency budget reachable.

**Caveats.** These were taken by direct invocation, not from a live Claude Code session.
There is no committed latency harness — the nine `Benchmark` functions are library
micro-benchmarks, and `make bench` runs none of the commands above. Per `MANIFESTO.md`,
performance claims that can't be reproduced aren't claims; if you change a number here,
land the harness that produces it in the same commit.

---

## 14. Physical layout of the data layer

```
<vault>/                          (a git repo, e.g. ~/Documents/Vault)
├── notes/<type>/<slug>.md        7 types: concept howto api pattern pitfall incident decision
├── moc/                          Map of Content; moc/weekly/YYYY-WW.md rollup
├── _inbox/                       quarantine, confidence: low
├── _archive/
├── profiles/me.md                developer profile (written by forge init)
├── reports/                      forge check's nine reports
├── raw/  sources/                outside the note contract, live
├── _index.md                     forge index's output
└── .forge/
    ├── forge-bin                 absolute path the vault hook will call
    ├── capture.log                the D3 hook's only diagnostic channel
    ├── <cache>.db                SQLite — derived + the budget table
    ├── code-index-<repo>.json    per-repo code index cache
    ├── weekly-stats.json         week-over-week delta persistence
    └── cache/<url-hash>.md       WebFetch source cache, TTL'd
```

`pkg/drift/gitindex.go` deliberately writes the cache per-repo as
`.forge/code-index-<repo>.json`, not a single shared `.forge/code-index.json` —
since `--repo` is repeatable, one shared name would collide across repos.
`persist`'s doc comment carries this reasoning.

---

## 15. Test strategy

| Layer | How it's tested |
|---|---|
| Pure functions (`recall`, `similarity`, `graph`, `sentinel`) | Direct unit tests, no I/O |
| Vault operations | `testdata/vault/` fixture copied to a temp dir, `git init`'d |
| Drift | A temp git repo is set up, committed to, reverted (`rollback_test.go`) |
| Engine | Fake API via `httptest` (`engine_run_httptest_test.go`) |
| CLI | `cmd/forge/e2e_test.go` |
| Cross-lane | `make test`: `CGO_ENABLED=1 go test` **then** `CGO_ENABLED=0 go build` |

### The `testdata/vault/` fixture

13 notes reproducing the real vault's **pre-migration** topology, plus **twelve
deliberate defects (F1–F12)**: mixed frontmatter shapes, a dangling wikilink, a
dangling `source:` path, an orphan, a near-duplicate pair, notes with no
frontmatter at all, status carried as body prose.

Two hard rules:

> **The defects are the test surface. Do not fix them.**

> **It has no `.git`, deliberately.** A nested repo would become a stray gitlink
> once this repo is `git init`-ed. The harness copies the fixture to a temp dir and
> `git init`s **the copy**. **Never `git init` it in place.**

Catalogue: `testdata/README.md`.

This should not be confused with `examples/vault/` — a separate Phase 6 deliverable:
93 files generated from the real vault via `forge scrub`, scoped to `notes/` +
`moc/` only.
