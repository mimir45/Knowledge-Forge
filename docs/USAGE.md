# Knowledge Forge — Usage Guide

> This document answers the **how do I use it** question. For the design
> rationale, see [`ARCHITECTURE.md`](ARCHITECTURE.md).

---

## 0. Quick start

```bash
# 1. Build (pure Go lane — this is the lane that ships)
cd knowledge-forge
CGO_ENABLED=0 go build ./...
make build                       # → ./dist/forge

# 2. If you want code indexing (tree-sitter), the cgo lane
make full                        # CGO_ENABLED=1, -tags codeindex

# 3. Setup wizard
forge init --vault ~/Documents/Base --language java \
           --frameworks spring-boot,hibernate --seniority senior

# 4. See what config actually resolved to
forge config --layers
forge config --json

# 5. First real use
forge recall --question "how does keyset pagination work" --explain
```

**Stale binary trap:** the checked-in root `./forge` binary predates Phase 5b/6 —
it doesn't know the `logback` or `scrub` subcommands. Rebuild with `make build`
before relying on it.

---

## 1. Installation

### 1.1 Build from source

| Command | Effect |
|---|---|
| `make build` | Pure-Go build → `./dist/forge`. This is the lane that ships. |
| `make full` | `CGO_ENABLED=1 -tags codeindex` — with tree-sitter code indexing. |
| `make test` | `CGO_ENABLED=1 go test ./...` then `CGO_ENABLED=0 go build ./...` — both lanes. |
| `make bench` | Runs the benchmark suite. |
| `make vet` / `make fmt` / `make lint` | Static checks / formatting / linting. |
| `make dist` | Cross-compiles all six release targets. |
| `make checksums` | Generates checksums for the release artifacts. |
| `make install-hook` | Installs the git hooks (see §4). |
| `make clean` / `make help` | Cleanup / command list. |

### 1.2 As a Claude Code plugin

```
claude plugin marketplace add mimir45/Knowledge-Forge
```

**Caveat:** the release/checksum flow hasn't been exercised end-to-end on a clean
machine yet — treat it as unverified until someone confirms a fresh install works.
Once installed, `.claude-plugin/plugin.json`, `hooks/hooks.json`, and `agents/` are
auto-discovered — nothing else needs wiring by hand.

---

## 2. `forge init` — the setup wizard

`forge init` writes exactly two files and nothing else:

| File | Content |
|---|---|
| `~/.forge/forge.config.md` | Your settings — only the keys that differ from the packaged defaults, so a binary upgrade still brings you new defaults for everything you didn't decide. |
| `<vault>/profiles/me.md` | Your developer profile, rendered from `profiles/me.template.md`. |

Both are yours to edit afterwards. This command refuses to overwrite either
without `--force`. It does not prompt on its own — `skills/forge-init/` asks the
questions and calls this command.

```
usage: forge init --vault DIR [--language L] [--frameworks a,b] [--infra a,b]
                  [--seniority junior|mid|senior] [--depth 1-5] [--note-language en]
                  [--explain-style mechanism-first] [--trigger ask|auto|manual]
                  [--engine-preset claude-only] [--stack-preset java-backend]
                  [--force] [--dry-run]
```

| Flag | Default | Note |
|---|---|---|
| `--vault` | — | **Required.** |
| `--language` | — | Primary language, e.g. `java`. |
| `--frameworks` | — | Comma-separated, e.g. `spring-boot,hibernate`. |
| `--infra` | — | Comma-separated, e.g. `docker,postgres,kafka`. |
| `--seniority` | `mid` | `junior \| mid \| senior`. |
| `--depth` | `0` | 1–5; 0 derives from `--seniority`. |
| `--note-language` | `en` | Language of note bodies. |
| `--explain-style` | `mechanism-first` | `mechanism-first \| example-first \| analogy-first`. |
| `--trigger` | `ask` | `ask \| auto \| manual`. |
| `--engine-preset` | `claude-only` | `offline \| claude-only \| byo-api \| max`. |
| `--stack-preset` | — | `java-backend \| frontend \| devops \| minimal`. |
| `--force` | false | Overwrite existing files. |
| `--dry-run` | false | Print what would be written, write nothing. |

If run through the skill instead of raw flags, five questions are asked:
language, frameworks, infra, seniority, trigger. Exit codes 3 and 4 signal
vault-precondition failures (e.g. vault path doesn't exist, or isn't a git repo).

---

## 3. Command reference

### 3.1 `forge config`, `forge slug`

`forge config --layers` — lists which of the four config layers are present and
where. `forge config --json` — prints the fully resolved, merged config.
`forge slug --title "..."` — deterministic slug generation for note filenames.

### 3.2 `forge validate`, `forge index` / `forge reindex`

`forge validate <file>` — checks a note against the frontmatter/body contract.
`forge index` — incrementally updates `_index.md` and the SQLite cache.
`forge reindex` — rebuilds the SQLite cache **entirely** from markdown (the only
authoritative rebuild path; markdown is always the source of truth).

### 3.3 `forge recall`

```
forge recall --question "..." [--explain] [--stack "java,spring"]
```

Verdict thresholds:

| Score | Verdict |
|---|---|
| ≥ 0.85 | reuse |
| 0.55 – 0.85 | update |
| < 0.55 | create |

`--explain` prints the per-channel breakdown (title/tags/stack/body) behind the
final score.

### 3.4 `forge drift`

```
forge drift --repo <name>:<path> [--since-commit <sha>] [--apply] [--deep]
```

Verdicts: `OK | Repaired | Suspect | Broken | Skipped`. Measured latency on the
hook path: 60–70 ms (budget < 100 ms). See
[`ARCHITECTURE.md` §7](ARCHITECTURE.md#7-drift-git-anchored-decay-detection) for
the full contract and the reasoning behind git-anchoring.

### 3.5 `forge check`

Runs the full weekly sweep: nine reports rendered to `<vault>/reports/`
(coverage, staleness, duplicates, orphans, gaps, graph-health, churn, deadlinks,
drift), plus a weekly rollup and an optional AI pass that only **prints**
suggestions — it never writes on its own.

### 3.6 `forge engine select/run/record`

`forge engine select --stage write` — resolves which tier a stage would use and
why. `forge engine run --stage <s> --prompt <p>` — actually invokes the resolved
tier. `forge engine record` — writes the `engine_trail` into a note's
frontmatter; refuses to record against a locked stage.

### 3.7 `forge gate`, `forge verify-code`

```
forge gate --file draft.md
```

Runs the seven quality gates in fixed order:
`schema → citation → code → freshness → antislop → link → duplicate`.
Exit codes: `0` clean, `1` quarantined to `_inbox/` (**not an error** — a
deliberate quarantine signal), `2`/`3` other failure classes.

```
forge verify-code --lang bash|java|ts --file <path>
```

Compiles/checks the code citation in a throwaway directory — never inside the
user's own project.

### 3.8 `forge logback`

```
forge logback [--dry-run] [--remove-markers]
```

Writes three independently-gated outputs back into a code repo:
`docs/knowledge-map.md`, per-module `CLAUDE.md` fragments, and (opt-in) inline
markers. All are sentinel-based managed blocks — see
[`ARCHITECTURE.md` §9](ARCHITECTURE.md#9-sentinel-idempotent-managed-blocks).

### 3.9 `forge scrub`

```
forge scrub <src-vault> <dst-vault>
```

Redacts secret/PII-shaped content from a vault copy. **Fails closed** — if a note
can't be verified as clean after scrubbing, the whole run is cancelled rather
than shipping a partially-scrubbed vault.

### 3.10 `forge capture`

Called from the vault's `post-commit` hook. Captures D3 training pairs (question
→ note) from the commit. Notes authored by a product agent are stamped
`Forge-Write: true` so they're never mistaken for a human-authored correction and
don't contaminate the training data.

### 3.11 `forge stats`

Prints five sections: hit rate, top topics, gaps, time saved, trend.

### 3.12 Hook commands

| Command | Used from | Behavior |
|---|---|---|
| `forge session-context` | `SessionStart` | Prints vault index context. |
| `forge intent` | `UserPromptSubmit` | Recall-scores the prompt, injects the best hit as context if score > 0.7. |
| `forge session-capture` | `SessionEnd` | Captures session-level training signal. |
| `forge cache-source` | `PostToolUse` (WebFetch) | Caches fetched sources under `.forge/cache/`. |

All four are **fail-silent and always exit 0** — a hook must never be able to
break a session.

---

## 4. Hook installation

### 4.1 Claude Code lifecycle hooks

Installed automatically with the plugin, via `hooks/hooks.json`:
`SessionStart`, `UserPromptSubmit`, `SessionEnd`, `PostToolUse` (matcher
`WebFetch`).

**Resume trap:** on `--continue`/`--resume`, `SessionStart` re-runs (expected,
cheap, idempotent) but every *other* hook's output is replayed from the saved
transcript rather than re-executed. A stale recall hit after a resume is expected
behavior, not a bug.

### 4.2 Vault D3 hook

```bash
scripts/install_vault_hook.sh <vault-path>
```

Installs `.git/hooks/post-commit` in the vault, calling a **pinned binary copy**
at `~/.forge/bin/forge` (or `$FORGE_BIN`) — not the repo's own build output.
Rebuild that copy after any change to `pkg/dataset` or `cmd/forge/capture.go`:

```bash
CGO_ENABLED=0 go build -o ~/.forge/bin/forge ./cmd/forge
```

By design the hook never prints and can never fail a commit — so a stale or
broken binary is silent. Diagnose via `<vault>/.forge/capture.log`.

### 4.3 Code repo drift hooks

```bash
scripts/install_drift_hook.sh <repo-path>
```

Installs three hooks: `post-commit` (diffs from `HEAD^`), `post-merge` (diffs
from `ORIG_HEAD`), `post-checkout` (branch-level switches only).

---

## 5. Skills (slash commands)

| Skill | Purpose |
|---|---|
| `skills/forge/` | Main entry point — question → note pipeline. |
| `skills/forge-init/` | Onboarding wizard (asks the questions, calls `forge init`). |
| `skills/forge-check/` | Runs the weekly sweep interactively. |
| `skills/forge-stats/` | Prints usage stats. |

---

## 6. Config reference

### 6.1 The chain

Same four layers as [`ARCHITECTURE.md` §4](ARCHITECTURE.md#4-config-the-four-layer-chain):
`$FORGE_CONFIG` > `.forge.config.md` (project) > `~/.forge/forge.config.md`
(user) > the packaged template. Maps merge key by key; scalars and lists replace
wholesale.

### 6.2 Key examples

```yaml
vault_path: ~/Documents/Base
trigger:
  mode: ask
recall:
  reuse_threshold: 0.85
  update_threshold: 0.55
freshness_days:
  concept: 365
  howto: 180
  api: 90
engines:
  default: host
  api:
    provider: anthropic
    model: claude-sonnet-5
    api_key_env: ANTHROPIC_API_KEY
  budget:
    api_usd_per_day: 2.00
  on_exhausted: queue
pipeline:
  recall: { engine: none }     # locked
  write: { engine: none }      # locked
  index: { engine: none }      # locked
  research: { engine: host, fallback: none }
```

### 6.3 Deliberately in code, not config

`pkg/vault`'s `excludedPrefixes`/`hubNames`, `pkg/report/duplicates.go`'s
`specThreshold = 0.85`, and `$HOME/.forge/bin` in the Makefile — these are spec
constants, not user decisions, so they stay out of the config schema.

### 6.4 Packaged presets

| Engine preset | Effect |
|---|---|
| `offline` | Everything `none`. |
| `claude-only` | `host` for research/write-adjacent stages, locked stages stay `none`. |
| `byo-api` | Adds the `api` tier with your own key. |
| `max` | Adds the `advisor` critique tier on top. |

| Stack preset | Effect |
|---|---|
| `java-backend` | Java/Kotlin code indexing, Spring-flavored stack hints. |
| `frontend` | TypeScript code indexing. |
| `devops` | Infra-focused stack hints, no code indexing. |
| `minimal` | The smallest useful default set. |

---

## 7. Vault layout

```
<vault>/
├── notes/<type>/<slug>.md    concept · howto · api · pattern · pitfall · incident · decision
├── moc/                       Map of Content
├── _inbox/                    quarantine (confidence: low)
├── _archive/
├── profiles/me.md
├── reports/
├── raw/  sources/
└── .forge/
```

Note: seven note types are in active use, one more than DESIGN §7's original
five-directory tree — three of the extra directories are currently empty
`.gitkeep` shells reserved for future use.

---

## 8. Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `logback`/`scrub` "unknown command" | Stale root `./forge` binary predates Phase 5b/6 | `make build`, use `./dist/forge` |
| No D3 pairs appearing | Hook fails silently by design | Check `<vault>/.forge/capture.log` |
| `forge drift` returns `Skipped` for everything | Wrong `--since-commit`, or repo path mismatch | Verify with `git log --oneline -1 <sha>` |
| Config value not what's expected | Layer precedence confusion | `forge config --layers` then `forge config --json` |
| "not allowed — locked to none" | Tried to set `recall`/`write`/`index` to a non-`none` engine | This is enforced deliberately — see `ARCHITECTURE.md` §5.3 |
| `forge gate` exits 1 | Note quarantined to `_inbox/` | Not an error — fix the note, or leave it for review |
| `forge check` slow | Network-bound reports (linkcheck) | Expected; budget is 10s, typically well under warm |
| Code indexing missing symbols | Running the pure-Go lane | Rebuild with `make full` (cgo, `-tags codeindex`) |
| Recall scores feel off | Calibration is measured, not guessed | Read `references/recall-spec.md` before changing weights |
| Dataset export looks incomplete | D6 is derived, not captured | Re-run `forge export-dataset` after new D1-D5 captures |
| Code-index cache reused across repos unexpectedly | Cache is named per-repo | Confirm `--repo <name>:<path>` names differ |
