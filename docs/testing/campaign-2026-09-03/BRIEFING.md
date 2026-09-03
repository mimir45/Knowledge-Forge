# Knowledge Forge test campaign — agent briefing

You are one of ten agents driving the `forge` CLI from the outside to find bugs.
Read this whole file before running anything.

## Absolute paths (use these literally, never a relative path)

| Thing | Path |
|---|---|
| Pinned binary | `/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/forge` |
| Your vault (yours alone — NN is your agent number, zero-padded) | `/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/vaults/vNN` |
| Config variants | `/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/configs/` |
| Recorder | `/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/runcase.sh` |
| Observation recorder | `/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/addnote.sh` |
| Your scratch space for temp files | `/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/scratch/NN/` (mkdir it) |

Always export before you start:

```bash
export FORGE_CONFIG=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/configs/base.md
```

## Hard rules — violating any of these invalidates the campaign

1. **NEVER touch `/Users/mimir45/Documents/Base`** (the user's real vault),
   `~/.forge/`, `~/.claude/`, or the git repo at `/Users/mimir45/knowledge-forge`.
   Every `forge` call passes `--vault /Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/vaults/vNN`
   **explicitly**. If a command has no `--vault` flag, set `FORGE_CONFIG` to a config
   whose `vault_path` points at your copy.
2. **NEVER run `forge init`** — it is the only writer of `~/.forge/forge.config.md`
   and would destroy the user's configuration.
3. **Never type a number you did not measure.** Every command goes through
   `runcase.sh`, which records the real exit code and wall time for you. Do not
   report timings or exit codes from memory or estimation.
4. Do not `git commit`, `git push`, or modify anything in the knowledge-forge repo.
5. Stay in scope. If you find something interesting outside your assigned cases,
   record it with `addnote.sh` and move on — do not go exploring.

## How to run a case

```bash
bash /Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/runcase.sh \
  NN case-id EXPECT_EXIT "what the contract says should happen" \
  -- /Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/forge recall \
     --question "..." --vault /Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/vaults/vNN
```

`EXPECT_EXIT` is the exit code the documented contract promises, or `-` when the
case is exploratory. The script computes `verdict` by comparing. It appends one JSON
line to `runs/agent-NN.jsonl` and prints a short summary you can read.

For anything an exit code cannot express — JSON shape, file contents, a leaked
string, an empty report that should have been skipped — use:

```bash
bash /Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/addnote.sh \
  NN case-id bug "topic field contained the nonce: topic=\"how-does-nonce7q2x-work\""
```

Severity is `bug` (a contract is broken and you saw the evidence), `suspect`
(looks wrong, not certain), or `info`. **Always quote the concrete string you saw.**
A note with no observed evidence in it is worthless.

## Exit code semantics — read before judging anything

This CLI's exit codes are documented in each `--help`. Getting these wrong produces
false bug reports.

| Code | Meaning |
|---|---|
| 0 | success. **The four hook commands (`intent`, `session-context`, `session-capture`, `cache-source`) return 0 in ALL cases** — by design, fail-silent. You cannot judge them by exit code; check stderr and side effects instead. |
| 1 | **Overloaded, not automatically a bug.** Three distinct meanings: (a) `validate` found a defect; (b) `gate` quarantined the draft to `_inbox/` — this is correct handling, not an error; (c) an actual failure. |
| 2 | usage / flag parse error / config or vault resolution failure. The one reliable "you used it wrong" signal. |
| 3 | internal error or vault precondition. For `gate`, exit 3 means the draft was **not handled at all** and must still be sitting untouched at `--file`. |
| 4 | `forge init`'s vault precondition (you will not run init). |
| 127 | the `bin/forge` shim rejecting a hash pin — you are not using the shim, so if you ever see 127 something is wrong with your command. |

## Output streams

**Do not assume "empty stderr means no error."** Errors normally go to stderr in the
form `forge <subcommand>: <message>`, but eight call sites print error-shaped output
to **stdout** instead (`check_drain.go`, `engine_run.go`, `logback_*.go`,
`check_ai_pass.go`). `runcase.sh` captures both separately — read both.

## What to send back

A summary of **at most 10 lines**. State: how many cases you ran, how many
`verdict:fail`, and the two or three most important findings in one line each.
Do not paste JSON, do not paste command output, do not restate the case list.
Your JSONL file is the real deliverable; the orchestrator aggregates from it.
