# Knowledge Forge — external CLI test campaign, 2026-09-03

**180 recorded command invocations across ten parallel agents, plus a serial latency
pass and three orchestrator verification rounds. 273 raw records. No panics, no hangs,
no injection. Twelve confirmed defects, four agent claims refuted on verification.**

Environment, binary hash and method: [`env.md`](env.md). Raw data: `runs/*.jsonl`.
Every number below is re-derivable with `bash aggregate.sh`.

---

## Why this campaign existed

Two gaps in the project's own verification made it worth running:

- **No test drives the compiled binary.** Every "e2e" test under `cmd/forge/` is an
  in-process function call (`runValidate(...)`, `cmdCheck([...])`). `main.go`'s `run()`
  is never invoked and no exit code is ever asserted. The CLI's *argv → exit code*
  contract — spelled out in detail in each `--help` — had never been executed.
- **No latency harness exists.** The numbers in `CLAUDE.md` (`drift` 60–70ms, `index`
  20ms, `qualitygate.Run` ~0.13ms) cannot be produced from anything committed: all nine
  `Benchmark` functions are library micro-benchmarks, and `pkg/qualitygate` has no
  benchmark at all.

Determinism was already covered by `evals/run.sh` and was not re-tested.

## Method

Ten Haiku agents ran **in parallel** against ten independent copies of the real vault,
each recording through a shell recorder that captured the actual exit code, wall time
and both streams — so no number in `runs/*.jsonl` was typed by an agent. Timings from
that phase are stamped `contended: true` and are **never compared to a budget**: ten
processes on one machine and one SQLite cache measure the harness, not the product.

Latency was measured afterwards in a **serial** pass with nothing else running.

Everything an agent reported was re-derived by the orchestrator before entering this
report. Four claims did not survive that step and are listed under *Refuted*.

---

## Confirmed defects

### B1 — Telemetry writes the user's question in plaintext ▲ critical

The stated invariant is *"Telemetry logs the topic and a hash. Never raw question text,
code, or file contents."* The `topic` field is in fact the **whole question** put
through `vault.Slug`, so the question is recoverable — including anything secret pasted
into it.

```
$ forge recall --question "how do I rotate my AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI credential" --vault V
$ tail -1 V/.forge/log.jsonl | jq '{topic, q_hash}'
{
  "topic": "how-do-i-rotate-my-aws-secret-access-key-wjalrxutnfemi-credential",
  "q_hash": "1f133d315c8e"
}
```

The secret value survives verbatim (lower-cased). Reproduced independently by agent 10
with a different nonce (`sk-live-SECRET-9931` → `sk-live-secret-9931`) and by agent 08
in Turkish. `q_hash` is a sha256 truncated to **12 hex characters**, which the docs
describe simply as "a hash".

Source: `cmd/forge/recall.go:108` (`topic: vault.Slug(q.Question)`), `pkg/vault/slug.go`.
`telemetry.enabled: false` does correctly suppress the write.

### B2 — The `UserPromptSubmit` hook is inert for every real user ▲ critical

`forge intent` decodes exactly one stdin field, `user_prompt`
(`cmd/forge/intent.go:38-46`). Claude Code sends the prompt as **`user_input`**
(documented stdin payload for `UserPromptSubmit`: `session_id`, `prompt_id`,
`transcript_path`, `cwd`, `permission_mode`, `hook_event_name`, `user_input`, `model`).
The shim `hooks/user-prompt-intent` pipes stdin through unmodified, so the field never
matches and the command always sees an empty prompt.

```
$ Q="how does dependency injection work in spring boot"     # recall top_score = 0.75
$ echo "{\"user_input\":\"$Q\"}"  | forge intent --vault V   ->  (no output)
$ echo "{\"prompt\":\"$Q\"}"      | forge intent --vault V   ->  (no output)
$ echo "{\"user_prompt\":\"$Q\"}" | forge intent --vault V   ->  {"additionalContext":"The vault may already answer this — …
```

Because the hook is contracted to be silent below threshold and to always exit 0, this
failure is **completely invisible in normal use**. The gate calibration behind it
(`testdata/intent-gate-labels.txt`, `intent-gate.golden`, the derivation of the 0.50
constant) has never fired in production.

The other three hook commands were checked for the same class of bug and are **correct**:
`session-capture` reads `session_id` / `transcript_path` / `reason`, `cache-source` reads
`tool_name` / `tool_input` / `tool_response`. `intent` is the only one.

### B3 — A mistyped `--vault` silently creates a new vault ▲ high

`--vault` pointing at a path that does not exist is not rejected. Where the parent is
writable, `forge` **creates the directory** and initialises `.forge/` inside it:

```
$ forge recall --question test --vault /tmp/typo-in-my-vault-path
$ echo $?          # 0
$ ls -a /tmp/typo-in-my-vault-path
.  ..  .forge
```

The user gets an empty result set and a new empty vault, with no indication that they
misspelled the path. Where the parent is *not* writable the failure surfaces as an
implementation detail with the wrong exit code — the CLI's documented usage-error code
is 2:

```
$ forge recall --question test --vault /nonexistent/definitely/not/here
forge recall: mkdir /nonexistent: read-only file system
$ echo $?          # 1, not 2
$ forge recall --question test --vault /etc/hosts
forge recall: mkdir /etc/hosts: not a directory
$ echo $?          # 1, not 2
```

### B4 — The locked-stage guard has a hole, and the two guards disagree ▲ high

`recall`, `write` and `index` are locked to engine `none`, and the binary is documented
to refuse to start rather than silently override. `pkg/config/validate.go` checks only
`Stage.Engine` — it never inspects `fallback` or `then`. So a paid tier on a locked
stage loads cleanly:

```
# config: pipeline: { write: {engine: none, fallback: api} }
$ FORGE_CONFIG=…/h3-locked-fallback.md forge config
$ echo $?          # 0  — accepted

$ FORGE_CONFIG=…/h3-locked-fallback.md forge engine select --stage write --vault V
forge engine select: engine: pipeline.write: "api" is not allowed — [recall write index] are locked to "none" (T0 static core)
$ echo $?          # 2  — rejected
```

Two guards, two different verdicts on the same file, and two differently-worded errors.
Compare the direct violation, which `forge config` does catch, with a much better
message that names the offending layer:

```
$ FORGE_CONFIG=…/locked-direct.md forge config      # pipeline.recall.engine: host
forge config: pipeline.recall: engine "host" is not allowed — stages recall, write, index are locked
to engine "none" (the T0 static core makes zero model calls); set by …/locked-direct.md
$ echo $?          # 2
```

### B5 — `forge drift` misses its budget, and `--since-commit` is not the cheap path ▬ medium

`drift` runs on the git-hook path and is the project's binding latency constraint:
budget **<100ms**, `CLAUDE.md` records 60–70ms measured. Serial pass, 94-note vault,
nothing else running:

```
drift                 median 151ms   p95 208ms      budget 100ms   OVER
drift --since-commit  147ms, 147ms, 148ms (3 runs)
```

`--since-commit` is documented as "the cheap gate … the git hooks always pass it", but
it is **not measurably cheaper** than evaluating every citation. Whatever dominates the
cost is not the citation scan.

### B6 — `verify-code` accepts an unknown language and exits 0 ▬ medium

Documented: `--lang java|ts|bash|auto`, exit 2 on a usage error.

```
$ forge verify-code --lang cobol --file ok.sh
$ echo $?          # 0 — silently "skipped"
```

A typo'd `--lang` reports success. Combined with the exit-0-on-skip contract, a quality
gate wired to this would pass every snippet.

### B7 — `verify-code` reports a missing file as a compile failure ▬ medium

```
$ forge verify-code --lang bash --file /nonexistent/x.sh
forge verify-code: open /nonexistent/x.sh: no such file or directory
$ echo $?          # 1 (= "the snippet failed"), documented as 2 (= usage error)
```

A caller distinguishing "your snippet is broken" from "you called me wrong" cannot.

### B8 — `forge check` reports "0 skipped" after skipping two reports ▬ medium

`--help`: *"Without `--repo`, drift.md and moc/codebase.md … are skipped rather than
written empty … the run says which files it wrote and which it skipped."*

The skipping is correct — verified, neither file is created. The **accounting is wrong**:

```
$ forge check --vault V --offline        # no --repo
… 10 files listed …
10 written, 0 skipped
```

Two files were skipped and the run says zero, so the promise that the run tells you what
it skipped is not kept.

### B9 — `engine select` accepts any stage name ▽ low

`Pipeline` is a `map[string]Stage` with no enum, so a typo resolves to nothing and
succeeds:

```
$ forge engine select --stage nosuchstage --json --vault V
$ echo $?          # 0
```

Silently mis-routing a mistyped stage is worse than rejecting it.

### B10 — Three telemetry fields are always empty, and `forge stats` derives from them ▽ low

`duration_ms` and `sources` are always `0` and `project` always `""` on every `ask`
event (`cmd/forge/recall.go:95-97`), confirmed across every recall the campaign made.
`forge stats` reports a hit rate and an "approximate research time saved" built on top
of those fields — with them hard-wired to zero, the estimate cannot be meaningful.

### B11 — `telemetry.scope` is validated but never read ▽ low

`local|team` is enforced (`pkg/config/validate.go:93`) and that is the only reference to
it in non-test Go source. `scope: team` changes nothing observable — no `scope` key
appears in `log.jsonl`, and output is byte-identical to `scope: local`.

### B12 — The documented latency table does not match this machine ▽ perf

Serial pass, 10 runs each, median / p95, same class of machine (Apple M4) the documented
figures were taken on:

| command | cold | warm median | warm p95 | budget | `CLAUDE.md` claims | verdict |
|---|---|---|---|---|---|---|
| `recall` | 188ms | 57ms | 114ms | — | — | — |
| `index` | 116ms | 127ms | 140ms | 200ms | 20ms | under budget, **6× the claim** |
| `check --offline` | 206ms | 147ms | 160ms | 10s | 390ms warm / 930ms cold | comfortably **better** than claimed |
| `drift` | 164ms | 151ms | 208ms | **100ms** | 60–70ms | **OVER BUDGET** (B5) |
| `session-context` | 40ms | 36ms | 38ms | 200ms | — | ok |
| `intent` | 128ms | 49ms | 58ms | **50ms** | — | median just inside; **p95 and cold are over** |
| `verify-code bash` | 48ms | 46ms | 90ms | — | ~10ms warm / ~470ms cold | **4.6× the claim** |
| `verify-code java` | 1238ms | 1030ms | 1373ms | — | ~170ms warm / ~370ms cold | **6× the claim** |

`intent`'s 50ms budget holds only warm; a first invocation in a session costs 128ms.
Both stay far inside the 2s hook timeout in `hooks/hooks.json`, so nothing user-visible
breaks — but the budget as written is not met on the path that matters (the first prompt
of a session).

`pkg/qualitygate`'s documented ~0.13ms could not be checked **because the package
contains no benchmark at all** (`go test ./pkg/qualitygate -bench .` runs nothing), and
it is absent from `make bench`'s package list.

---

## Refuted — claimed by an agent, did not survive verification

Recorded because a test report that only lists confirmations is not trustworthy.

| Claim | Verdict |
|---|---|
| *"`ANSWER_FROM_VAULT` (≥0.85) is unreachable in practice"* — agents 01 and 03 | **False.** Across the campaign's 109 recall responses the verdict distribution was `CREATE` 57, `UPDATE(extend)` 45, **`ANSWER_FROM_VAULT` 7**, with scores up to 1.0. Those two agents' topic packs happened to top out at 0.75 and 0.80; agent 06 saw 0.882 and a direct title query scores 0.95. The threshold is reachable — it is topic-dependent. |
| *"`forge check` writes `drift.md` and `moc/codebase.md` even without `--repo`"* — agent 05 | **Artifact.** The vault copy already shipped those two files from the user's real vault. With them removed first, `check --offline` creates neither. (The related *accounting* defect is real and is **B8**.) |
| *"ALL-CAPS questions score higher than mixed case"* — agent 08 | **False.** Casing is irrelevant: `Bölüm I Teori Temelleri`, `bölüm i teori temelleri` and `BÖLÜM I TEORİ TEMELLERİ` all score **0.95**. The 0.80 the agent compared against came from a *different string* — one with the extra word `nedir` — so what it measured was query-length dilution, not case sensitivity. |
| *"An unresolvable import masks real compile errors"* — agent 06, filed as critical | **False.** A genuine syntax error still fails even alongside an unresolvable import (`exit 1`, `verdict: fail`). The `skipped` verdict on *import-only* failures is consistent with the declared contract that `verify-code` is "never a dependency resolver". Not a defect. |

---

## Turkish-language recall — real, but not a code defect

Turkish questions score far below their English equivalents for the same concept:

| concept | Turkish | English |
|---|---|---|
| Spring Boot dependency injection | 0.45 | 0.75 |
| Hibernate soft delete | 0.303 | 0.623 |
| Testcontainers integration test | 0.17 | 0.427 |
| *the one Turkish-titled note in the vault* | **0.80** | 0.0 |

The mechanism is not a tokenizer bug. Diacritics are handled correctly
(`ö→o ü→u ş→s ğ→g ı→i ç→c`), casing is irrelevant, and nothing crashes on Turkish-only
input. The notes themselves are written in English, so a Turkish question matches only on
the shared technical tokens and every Turkish function word dilutes the score — visible
directly in `0.95` for a bare title versus `0.80` for the same title plus `nedir`. The
last row is the proof: against the one Turkish-titled note, Turkish wins outright and
English misses entirely.

So this is corpus-language behaviour working as designed — but for a Turkish-speaking
owner it means native-language questions retrieve materially worse, which is worth a
product decision rather than a bug fix.

---

## Robustness — nothing broke

38 degenerate-input cases and 20 hook cases produced **no panic, no stack trace, no
hang**, across: empty / whitespace / single-character / 10 KB / 100 KB questions
(84ms and 203ms respectively, valid JSON both times), emoji, full-width and astral
Unicode, embedded newlines, and path traversal.

- Shell metacharacters (`$(whoami) && rm -rf / ; …`) were treated as text; no file was
  created, nothing executed.
- `'; DROP TABLE notes; --` was treated as text; the SQLite index still answered
  correctly afterwards.
- All four hook commands returned exit 0 on empty stdin, malformed JSON, plain text and
  a nonexistent vault — the fail-silent contract holds. (That it holds is also what hides
  **B2**.)
- `forge index` is genuinely idempotent within a day (identical sha256 across two runs).
- Deleting `.forge/cache/index.db` entirely and running `forge reindex` rebuilt the vault
  from markdown alone and recall worked afterwards — markdown really is the only source
  of truth. The `budget` table survives, though by omission from `Reset()`'s table list
  rather than by an explicit guard.
- `forge recall` is byte-identical across runs once `run_id` is stripped, and `--explain`
  leaves stdout untouched, writing only to stderr.

---

## Documentation drift

Reported separately at the user's request; no fixes were applied.

| ID | Finding |
|---|---|
| **D1** | There is no way to ask the binary its version: `--version`, `-v`, `-V` and `version` are all `unknown command`. `make build` stamps `-X main.version -X main.commit` and nothing ever prints them. Three different `forge` binaries were reachable on this machine and none could be told apart. `plugin.json` declares `0.1.2` while the build stamps `v0.1.1-13-g33b94b1`. |
| **D2** | All eight paths in `CLAUDE.md`'s "Read the docs in this order" are gone — removed in `df5ccea refactor: cleaned old docs`, `CLAUDE.md` never updated. `docs/` holds only `ARCHITECTURE.md`, `USAGE.md`, `datasets.md`. |
| **D3** | Live code and specs still cite those deleted documents: `references/recall-spec.md:12` ("Source of truth: DESIGN §5.3"), `forge gate --help` ("DESIGN §12"), `forge check --help` ("ADDENDUM section B.4"), `config/forge.config.example.md` ("DESIGN §5.3's decision tree"). None resolves. |
| **D4** | `.claude/agents/` does not exist, but `CLAUDE.md` documents six workflow agents living there (`finder`, `executor`, `explainer`, `vault-analyst`, `doc-auditor`, `cross-checker`) and instructs the reader to prefer delegating to them. |
| **D5** | Drift in the opposite direction: `skills/forge/SKILL.md:228-236` states that nothing loads agents from a root-level `agents/` directory and prescribes a manual Agent-tool dispatch. In fact the four product agents *are* discovered as `forge:forge-researcher`, `forge:forge-codebase-scout`, `forge:forge-verifier`, `forge:forge-librarian`. The skill therefore steers callers away from a working mechanism — a documentation defect with runtime consequences. |
| **D6** | `CLAUDE.md` describes the cgo lane as `-tags codeindex`; the real constraint is `//go:build cgo` and `make full` passes no tag. |
| **D7** | `CLAUDE.md` quotes `pkg/qualitygate.Run` at ~0.13ms. The package has no benchmark and is not in `make bench`'s list. `make bench`'s own comment says "the three the phase brief names" above a list of six packages. |
| **D8** | `hooks/user-prompt-intent` and `docs/USAGE.md:228` both still describe the intent gate as `> 0.7`; the constant is `0.50` (`cmd/forge/intent.go:84`), deliberately re-derived from measurement. Moot in practice while **B2** stands. |

---

## Isolation

No `forge` invocation in this campaign was ever pointed at the user's real vault.
`FORGE_CONFIG` was pinned to a scratch config on every call and `--vault` was passed
explicitly. Verified afterwards: nothing under `~/.forge/` changed, and
`~/.forge/forge.config.md` was untouched.

**One side effect did reach the real vault, and it was not the campaign's `forge` calls.**
Three files appeared under `~/Documents/Base/.forge/cache/` during the run
(e.g. `url: https://code.claude.com/docs/en/claude_code_docs_map.md`, `fetched: 2026-09-03`).
These were written by the *installed* plugin's own `PostToolUse` hook, firing on a
subagent's `WebFetch` in this session and resolving the vault from the user's own config.
Worth noting on its own terms: `forge cache-source` writes into the configured vault from
any session, including a subagent's fetches, with no scoping to the project being worked
on.

## Regression baseline

Run after the campaign, all clean:

```
CGO_ENABLED=0 go build ./...                              ok
go test ./cmd/forge -run TestFixtureIsNeverMutated        ok
testdata/vault/                                           clean (no .git, _index.md, .forge)
evals/run.sh                                              PASS (recall + check determinism)
```

## Coverage

| Exercised | Not exercised |
|---|---|
| `recall` (109 responses), `index`, `reindex`, `check`, `drift`, `validate`, `gate`, `verify-code`, `config`, `engine select`, `slug`, `stats`, `logback --dry-run`, all four hook commands | `capture`, `export-dataset`, `dataset-stats`, `scrub`, `engine run` against a live tier, `logback` writing, `drift --apply`, `drift --deep` |
| exit-code contracts, JSON schema, threshold bands, determinism, idempotency, cache rebuild, degenerate input, injection, Unicode, config-chain validation, latency vs budget | the cgo/`codeindex` lane (out of scope), the `/forge` skill pipeline end-to-end, multi-repo `drift`, `on_exhausted` against a real paid tier |

## Reproducing any finding

```bash
make build                                    # from commit 33b94b1
export FORGE_CONFIG=<a config whose vault_path is a COPY of your vault>
./forge <the command quoted in the finding> --vault <that copy>
```

Then re-derive every total in this report from the raw data:

```bash
bash docs/testing/campaign-2026-09-03/aggregate.sh
```
