# evals/

Not a quality eval in the LLM-scoring sense — the T0 static core makes zero model
calls, so there's no model output to grade here. What this checks is the property
this project's docs have measured by hand each phase and now claims as an invariant:
given the same vault, `forge recall` and `forge check` produce byte-identical output
on repeated runs (see CLAUDE.md's Status section — Phase 2b's six-consecutive-run md5
check, Phase 5b's byte-identical `logback` rerun check — this is that check, scripted
and repeatable in CI).

## Running

```sh
./evals/run.sh
```

Builds `forge` (portable, `CGO_ENABLED=0` lane), copies `evals/fixtures/vault/` into a
temp dir, `git init`s the copy (matching this repo's established
copy-then-`git init` convention for exercising git-anchored behavior — see
`testdata/vault/`'s own harness), then:

1. Runs `forge recall --question "..." --vault <dir>` twice and diffs the JSON output.
2. Runs `forge check --vault <dir> --offline` twice against two separate copies and
   diffs the rendered `reports/` trees.

Exits non-zero if either check finds a difference.

## Fixtures

`evals/fixtures/vault/` is a small, deliberately clean fixture — two schema-valid,
cross-linked notes (`notes/concept/kafka-consumer-group-rebalancing.md`,
`notes/howto/configure-kafka-consumer-timeouts.md`) plus the empty DESIGN §7
topology directories (`moc/`, `_inbox/`, `_archive/`, `profiles/`) as `.gitkeep`
shells.

This is **not** `testdata/vault/` (that fixture's F1–F12 defects are deliberate and
exercised by `pkg/vault`, `pkg/graph`, etc.'s own tests — see `testdata/README.md`;
don't reuse it here, and don't fix its defects) and **not** `examples/vault/` (Phase 6's
separate, real-vault-derived, `forge scrub`-processed, human-reviewed example content).
Three different fixtures, three different purposes — see CONTRIBUTING.md's "Fixture
vaults" section.
