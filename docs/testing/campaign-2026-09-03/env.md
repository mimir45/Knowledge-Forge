# Campaign environment

| | |
|---|---|
| Date | 2026-09-03 |
| Repo commit | `33b94b19d6782aea3294cc8103f26878978bb40b` |
| Branch | `worktree-kf-test-campaign` |
| Binary under test | built from that commit with `make build` (`CGO_ENABLED=0`) |
| Binary sha256 | `f5ef75a72d811bb2977538fcd1eb88c6046556fe28075be4995001749125bc66` |
| Build lane | pure Go only. The cgo/`codeindex` lane was **out of scope** for this campaign. |
| Go | `go1.26.2 darwin/arm64` |
| Host | macOS 26.5.2, `arm64` (Apple M4) |
| javac | `24.0.2` |
| node | `v24.13.1` |
| tsc | **absent** — every TypeScript `verify-code` case therefore exercised the *skipped* path, not the compile path |

## Why the binary was pinned

Three different `forge` binaries were reachable on this machine at campaign time: the
repo's own build output, the marketplace plugin cache at `0.1.2`, and the pinned copy
recorded in `<vault>/.forge/forge-bin`. Since the CLI has no way to report its own
version (finding **D1**), the only way to know which build produced a result is to build
one from a known commit and address it by absolute path. That is what every case did.

`bin/forge` was deliberately **not** used: it is a SHA256-verifying shim, not a binary,
and rejects a pin mismatch with exit 127 — which from outside looks like a crash.

## Vault under test

A copy of the user's real Obsidian vault (`~/Documents/Base`, 94 notes / 488 files),
copied to scratch and given a baseline commit. **Ten independent copies** were made, one
per agent, because almost every `forge` subcommand mutates the vault it is pointed at
(`_index.md`, the SQLite cache, `.forge/log.jsonl`, `.forge/datasets/`, `_inbox/`) and
ten agents sharing one vault would have corrupted each other's assertions.

`forge scrub` was deliberately not used to build the copy: it redacts content, which
shifts recall scores and would have made every measurement unrepresentative of the real
vault.

## Files here

| File | What it is |
|---|---|
| `report.md` | the findings |
| `runs/agent-NN.jsonl` | 273 raw records — one JSON object per command invocation or observation |
| `aggregate.txt` | the aggregation output every number in `report.md` is derived from |
| `runcase.sh`, `addnote.sh` | the recorders the agents were required to use |
| `aggregate.sh` | rebuilds every total from the raw JSONL |
| `verify.sh` … `verify6.sh` | the orchestrator's own re-derivation of the agents' claims and of hypotheses H1–H6. `verify4.sh` covers the documentation sweep (D3, D9–D12); `verify5.sh` and `verify6.sh` isolate H6/B13 |
| `timing.sh` | the serial latency pass |
| `BRIEFING.md` | the contract every agent was given |
