---
name: forge-init
description: Use when the user wants to set up Knowledge Forge for a vault they have not configured yet — "set up forge", "install knowledge forge here", "onboard this vault", "run forge init", or the first time `forge recall`/`forge check` fails because no config exists. Not for changing an existing config (edit `~/.forge/forge.config.md` or `<vault>/profiles/me.md` directly instead) and not for re-running the pipeline itself (that's `forge`).
---

# Forge Init

A conversational wrapper around one command. **This skill writes nothing.**
`forge init` (the Go binary) is the only writer of
`~/.forge/forge.config.md` and `<vault>/profiles/me.md` — this file asks the questions,
builds one shell command, and reports what the binary did. If you find yourself about to
`Write` or `Edit` either of those two paths directly, stop: that is a second writer,
and the whole point of routing through the binary is to keep there being one.

---

## Step 0 — Check whether this is already done

```bash
forge config --layers
```

If `~/.forge/forge.config.md` already appears, say so and stop — ask whether they want
to re-run with `--force` (destructive: it overwrites their hand edits) or just point them
at the file to edit directly. Do not silently re-init an existing setup.

## Step 1 — Ask, at most five questions

Every question below maps onto one `forge init` flag. Nothing here is asked twice, and
nothing not listed here is asked at all — nine flags is not five questions, so nothing
below prompts for `--note-language`, `--explain-style`, `--depth`, `--force`, or
`--dry-run`; they carry defaults `forge init --help` documents, and a user who wants
something other than the default edits the profile afterwards.

| # | Question | Flag(s) | Default if skipped |
|---|---|---|---|
| 1 | "What's your primary stack?" (language, plus frameworks and infra if they volunteer them) | `--language`, `--frameworks`, `--infra` | `en-agnostic`, empty, empty |
| 2 | "How should notes read — junior, mid, or senior?" | `--seniority` | `mid` |
| 3 | "Should notes get created automatically, or should I ask each time?" (`auto` / `ask` / `manual`) | `--trigger` | `ask` |
| 4 | "Which stack preset fits — `java-backend`, `frontend`, `devops`, `minimal`, or none?" | `--stack-preset` | none (empty overlay) |
| 5 | "Which model budget — `offline` (zero calls), `claude-only`, `byo-api`, or `max`?" | `--engine-preset` | `claude-only` |

Accept short free-text answers ("java, spring boot, postgres and docker" is fine) —
translate it into the comma-separated flag values yourself rather than asking the user
to format it. If they don't know, use the default and say so; don't block on an answer
they can change later by hand.

**Also needed, not asked:** the vault path. Use `$VAULT` if the session already has one
(the same variable `skills/forge/SKILL.md` uses); otherwise ask once, separately from
the five above, since it isn't a preference — it's where the two files go.

## Step 2 — Run it

```bash
forge init --vault "$VAULT" \
  --language "$LANG" --frameworks "$FRAMEWORKS" --infra "$INFRA" \
  --seniority "$SENIORITY" --trigger "$TRIGGER" \
  --stack-preset "$STACK_PRESET" --engine-preset "$ENGINE_PRESET"
```

Show the user the command before running it — it's the only place their five answers
become visible as a single, checkable thing. Omit a flag entirely rather than passing an
empty string for one they skipped; `forge init` already defaults every flag it defines.

If it exits non-zero, show the error verbatim and stop. Common cause: the vault path
isn't writable, or `~/.forge/forge.config.md` exists and `--force` wasn't offered — do
not retry with `--force` on the user's behalf without asking first, since that
overwrites hand edits per Step 0.

## Step 3 — Install the vault's capture hook

```bash
scripts/install_vault_hook.sh "$VAULT" "$HOME/.forge/bin/forge"
```

Pass the **pinned copy**, `$HOME/.forge/bin/forge` (overridden by `$FORGE_BIN` if the
user has that set) — never `$(command -v forge)`. CLAUDE.md is explicit that the hook
must point at a fixed copy, not a `PATH` lookup: `PATH` can resolve to the repo's build
shim or nothing at all, and because the hook "can never fail a commit and never prints,"
a wrong path breaks capture silently — nothing shows up until someone reads
`<vault>/.forge/capture.log`. If that copy doesn't exist yet, build it first:
`CGO_ENABLED=0 go build -o "$HOME/.forge/bin/forge" ./cmd/forge`.

This is the D3 human-correction hook (`.git/hooks/post-commit` in the vault, not the
code repo) — it is what lets `pkg/dataset` learn from edits the user makes to notes
after they're written. The script is idempotent and refuses rather than clobbers an
existing hook that isn't its own; if it refuses — the real vault already has one
installed — tell the user and do not force it. Skip this step entirely if `$VAULT` is
not a git repository; say so rather than trying to `git init` it (an uninitialized
vault is out of scope here, not an error to route around).

## Step 4 — Index

```bash
forge index --vault "$VAULT"
```

Makes the freshly-created profile and config immediately visible to `forge recall`
rather than leaving a first question to run against a stale or absent index.

## Step 5 — Summarize, don't repeat the file

`forge init`'s own stdout already prints the settings it wrote (see
`cmd/forge/init_write.go`'s `summarize`) — relay that, plus:

- the two file paths, so the user knows what to edit by hand later
- one line confirming the capture hook installed (or why it didn't)
- one line confirming the index ran and how many notes it found
- a pointer: "edit `<vault>/profiles/me.md` any time — it's not regenerated"

Do not restate the full rendered config or profile body in chat; the files are the
source of truth and a second copy in the transcript is the thing that goes stale.

---

## Invariants

- This skill never calls `Write` or `Edit` on `~/.forge/forge.config.md` or
  `<vault>/profiles/me.md`. `forge init` is the only writer (D-4).
- At most five preference questions, plus the vault path if it isn't already known.
- An existing config is never silently replaced — `--force` requires the user's
  explicit say-so, asked for by name, every time.
- If any step fails, stop and show the real error. Do not paper over a failed hook
  install or a failed index by declaring setup complete anyway.
