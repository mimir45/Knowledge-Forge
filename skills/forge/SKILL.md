---
name: forge
description: Use when the user asks an explanatory question about a technology, concept, or integration — "how does X work", "what is X", "explain X", "how do I integrate X with Y", "what's the difference between X and Y", "best practices for X", "why does X do Y". Checks the vault before researching, then answers from it, extends an existing note, or writes a new one. Do NOT use for debugging a failure, refactoring or writing code, reviewing a diff, running commands, or recall of the user's own activity ("what did I do yesterday", "what's in my inbox") — those are not explanatory questions and this skill must not fire on them.
---

# Forge

Turns an explanatory question into a permanent, linked, verified note — **or into an
answer from a note that already exists.** The second outcome is the point: a vault that
grows a near-duplicate every time a topic comes up is worth less than no vault.

This file orchestrates. It holds no content. Read on demand:

| Need | Read |
|---|---|
| Scoring, thresholds, decision tree | `references/recall-spec.md` |
| Frontmatter fields, allowed values | `references/schema.yaml` |
| Tag and stack vocabulary | `references/taxonomy.md` |
| Note body structure | `templates/<type>.md` |
| Who you are writing for | `<vault>/profiles/me.md` |

---

## Stage 0 — Read the profile

**Read `<vault>/profiles/me.md` before writing anything.** It is written by `forge init`
and edited by hand afterwards. If it is missing, say so once and write at `mid` / depth 3
— do not invent a profile, and do not silently write generic prose as if none were asked
for.

A profile the model reads but does not act on is worse than no profile: it looks like
personalization while producing the same output for everyone. So each field below has a
concrete, checkable effect. If two notes written at different `seniority` values would
read the same, the profile is not being applied.

| Field | What it changes in the note you write |
|---|---|
| `primary_language` | every code example is in this language, unless the topic is itself about another one |
| `frameworks` | assumed available and assumed known; use them without introducing them |
| `infra` | worked examples target this infrastructure, not a generic stand-in |
| `seniority` | `junior` → define each term on first use, one worked example per claim, no unexplained jargon. `mid` → skip definitions, keep the tradeoffs. `senior` → skip tradeoff basics entirely, spend the space on edge cases, failure modes, and what the docs get wrong |
| `default_depth` | 1–5. Sets **section count** and how far down a mechanism you follow: 2 → what it does and one example; 3 → adds tradeoffs; 4 → adds failure modes and internals; 5 → adds source-level mechanism and version differences |
| `note_language` | the prose language of the body. Code, identifiers, and frontmatter values stay English regardless |
| `explain_style` | section order. `mechanism-first` → how it works, then when to use it. `example-first` → runnable example, then why it works. `analogy-first` → one analogy, then the mechanism, then drop the analogy |
| `assume_known` | referenced freely, **never re-explained**. Every entry here should delete a paragraph you would otherwise have written |
| `never_assume` | gets a one-line inline primer **every time** it appears, even mid-sentence |
| `code_style` | passed verbatim into how you write code blocks; treat it as a hard constraint, not a preference |
| `avoid` | negative constraints applied on a final pass over the finished draft |

`seniority` and `default_depth` move together but are not the same axis: seniority
decides *what you explain*, depth decides *how much*. A senior at depth 2 gets a short
note with no basics; a junior at depth 5 gets a long note that still defines its terms.

---

## Stage 1 — Recall. Always first.

```bash
forge recall --vault "$VAULT" --question "<the user's question, verbatim>" [--stack a,b]
```

**No research happens before recall has reported.** Not a web search, not a file read,
not a "let me quickly check". This is non-negotiable: recall is what makes the vault a
knowledge base rather than a pile, and research done first anchors every later judgement
to what was found instead of to what was already known.

Pass the question **verbatim**. Recall does its own normalization and stopword removal;
pre-summarizing it into keywords throws away the terms the scoring depends on. Pass
`--stack` only when the user named a technology the question does not already contain.

Recall is deterministic and makes **zero model calls**. It costs ~5 ms. There is no
budget argument for skipping it.

### Read the verdict, do not re-derive it

```json
{ "verdict": "UPDATE(extend)", "top_score": 0.617,
  "candidates": [ … ], "neighbours": [ … ] }
```

Branch on the `verdict` field. **Never compare `top_score` against a number yourself** —
the thresholds move into the config chain in Phase 3, and a copy of them in this prose
would diverge silently while still producing plausible-looking decisions.

`candidates` may be short or empty. Empty means no note matched on any channel: that is
`CREATE` with nothing to link to, and it is a correct answer, not a failure.

---

## Stage 2 — Branch on the verdict

### `ANSWER_FROM_VAULT` — the note exists and is fresh

1. Read `candidates[0].path` in full.
2. Answer the question **from that note**, not from memory and not from the web.
3. Show the link: `[[<slug>]]`, with the `verified` date so the user can judge its age.
4. Offer to deepen it — a section, a worked example, a source — and do nothing further
   unless they accept.

**Do not create a file.** Do not write, touch, or re-verify anything. If the note turns
out to be wrong or thin, say so and offer `UPDATE(refresh)`; do not silently repair it.

### `UPDATE(extend)` — a related note exists and this is new material for it

Target is `candidates[0].path`.

1. Research the question (stage 3), then read the target note in full.
2. **Insert a new section.** Never rewrite, reorder, condense, or "improve" existing
   body text — the user wrote it and did not ask you to edit it.
3. Add any new sources to `sources:`.
4. Bump `updated:` to today. Leave `verified:` alone — you did not re-check the
   existing claims.
5. Append one line to `## Changelog`: the date, what section was added, and why.
6. Show the user the added section and the frontmatter delta.

If the top candidate is clearly the wrong note — recall's tag/stack channels over-weight
ecosystem-wide terms like `spring` (BACKLOG B-008) — **say so and propose CREATE
instead.** Do not extend a note that is not about the topic. Check `candidates[1..]`
first; the right target is often second.

### `UPDATE(refresh)` — the note is on-topic but stale

Same target, different job: the claims are old, not missing.

1. Re-verify each existing claim against current sources.
2. Correct **only what actually changed.** A claim that still holds stays byte-identical.
3. Bump `verified:` (and `updated:` if anything changed).
4. **Show a diff and wait for approval before writing.** This is the one branch that
   touches text the user already trusted; it does not run unattended.

### `CREATE` — genuinely new

1. Research (stage 3).
2. `forge slug "<title>"` for the filename. Never hand-write a slug.
3. Pick a type — `concept howto pattern pitfall decision api incident` — and copy
   `templates/<type>.md`. Write into `notes/<type>/<slug>.md`.
4. Fill the frontmatter against `references/schema.yaml`; tags and stack come from
   `references/taxonomy.md`, not invented.
5. **Link every entry in `neighbours`** as `[[<slug>]]`, in a `## Related` section, each
   with a clause saying how it relates. An unexplained link is not a connection.
   If `neighbours` is empty, write no Related section — inventing links for a topic the
   vault does not cover is worse than leaving none.
6. Apply the profile (stage 0) to the draft: section count from `default_depth`, section
   order from `explain_style`, code from `primary_language` + `code_style`, and a final
   pass removing everything in `avoid`.
7. Run `forge validate <path>`. It must exit 0.

---

## Stage 3 — Research and verification

Applies to `CREATE` and both `UPDATE` modes.

### 3a — Research: dispatch in parallel

Dispatch `forge-researcher` and `forge-codebase-scout` **in the same turn** — two
independent calls, not one after the other. They have no dependency on each other: the
researcher answers "what does this technology do" from external sources, the scout
answers "how is it actually used in this repo" from local code, and running them
sequentially only adds latency, not accuracy.

Merge their reports into the draft:
- Researcher's findings + numbered sources → the note's `sources:` frontmatter and
  `## Sources` section.
- Scout's `file:line` examples → the `## In {{primary_stack}}` section (concept,
  pattern, api templates) — this section is the differentiator BACKLOG and DESIGN §11
  both call out; do not fill it with generic material the researcher could have found.
- If the scout found no local usage, say so in that section rather than omitting it —
  "not used in this repo yet" is a fact worth stating, not a gap to paper over.

### 3b — Verify: dispatch after research lands

Dispatch `forge-verifier` once the draft has code blocks and version-specific claims to
check. It delegates every compile/syntax check to `forge verify-code` itself — do not
compile snippets by hand in this stage, and do not duplicate the verifier's work.

- Prefer official documentation and primary sources. Record every one in `sources:`
  with the date consulted.
- Version-pin claims that are version-dependent. "Spring Boot does X" ages badly;
  "Spring Boot 4.0 does X" can be checked.
- Mark anything the verifier could not confirm as such in the note body. A hedge in the
  text is honest; a confident sentence that was never checked is a lie the vault will
  repeat.

**Advisor tier, when configured** (`verify.engine: advisor` or a fallback chain that
reaches it): treat this as a **two-pass verification** per DESIGN §15 — the deterministic
`forge-verifier` + `forge gate` pass runs first and is the pass that decides pass/fail;
the advisor tier only runs *after*, as critique-mode second pass over what already
passed. It returns disputed claims and a patch, never a rewrite (T3's invariant) — apply
its patch only to claims it disputed, not as a general polish pass.

### Packaging gap

Nothing in this repo currently loads agents from a root-level `agents/` directory —
Claude Code loads `.claude/agents/`, and there is no plugin manifest yet. Until that
packaging exists, "dispatch `forge-researcher`" means invoking the generic Agent tool
with the tool allowlist and prompt from `agents/forge-researcher.md` (and likewise for
the other three) — not live agent auto-discovery. Treat the four files as the spec to
follow, not as a `Task(subagent_type: "forge-researcher")` call that resolves today.

---

## Stage 4 — Gate: `forge gate` before every write

**Run this before writing to `notes/` or `_inbox/` in every branch** — `CREATE`,
`UPDATE(extend)`, and `UPDATE(refresh)` alike. It is a deterministic CLI, not an agent:
it makes no model calls and its verdict does not depend on which of the four agents
above ran or how.

```bash
forge gate --file <rendered draft> --rel <intended vault-relative path> \
           [--mode create|update] [--target-slug <slug>]
```

Read the JSON `Report` it prints. Branch on `Report.Quarantine` and the individual gate
outcomes — do not re-derive gate logic from this prose; the seven DESIGN §12 gates
(`pkg/qualitygate`) are the source of truth and this file does not restate their
thresholds:

- **`Quarantine: false`** — proceed to Stage 5 (link/publish). This applies whether
  every gate passed outright or a gate failed with a non-blocking remedy (`None`,
  `DelegateToLibrarian`, `SwitchToUpdate` never set `Quarantine`).
- **A `duplicate` gate outcome with `Remedy: SwitchToUpdate`** — this is a routing
  recommendation, not a hard block (`references/duplicate-spec.md` §6: 0.40 trips
  routinely on well-covered topics, not just on real duplicates). Reroute to
  `UPDATE(extend)` against the note the gate flagged, unless you have a stated reason to
  publish separately (DESIGN §12 permits it — e.g. a `pattern` and the `pitfall` that
  motivated it) — say that reason to the user rather than silently overriding.
- **`Quarantine: true`** — `forge gate` itself writes the draft to `_inbox/` with
  `confidence: low` and a `## Open questions` section naming every failed gate, then
  re-runs `forge index`. Do not write the file yourself in this case; the CLI already
  did it. Tell the user what failed and where the draft landed. This is what "gate
  failures go to `_inbox/`, never a silent publish" means concretely now — it is
  enforced by the CLI, not by this prompt's discipline.
- **Exit 3 (internal error)** — the draft was **not** handled: not published, not
  quarantined, left untouched at the path you gave `--file`. Do not treat this as either
  a pass or a quarantine; surface the stderr error to the user.

On a quarantine, `forge gate` prints a `--previous-draft` path to stderr. If you fix the
draft and re-run `forge gate` with that flag, a passing retry captures the (failing,
error, fixed) triple as training data (dataset D4) — pass it through when retrying
rather than starting a fresh invocation.

---

## Stage 5 — Link and publish: `forge-librarian`

Dispatch `forge-librarian` only after Stage 4 reports `Quarantine: false`. It is the
one agent with `Edit`/`Bash` in this pipeline and it never publishes a note that hasn't
already passed the gate — if you find yourself asking it to write something Stage 4
hasn't cleared, that is a bypass, not a shortcut.

- It populates `code_refs` (BACKLOG B-012) from the scout's `file:line` findings.
- It adds outbound/inbound links and MOC entries to satisfy the link gate's minimums.
- It runs `forge index` after writing.
- It commits with `git commit --trailer "Forge-Write: true"` on **every** commit it
  makes in the run — the note-write commit and any index-rebuild or link-fix commit
  alike (BACKLOG B-007; requires git ≥2.32). A commit missing the trailer reads to
  `pkg/dataset`'s D3 capture as a human correction, not a librarian action — that is the
  entire mechanism B-007 exists to close, so do not skip it "just for" a follow-up
  commit.

---

## Engine trail (Phase 3b)

`forge engine run` and this skill running a step in-session (`host`) both know they made
a model call, but neither one can stamp the note by itself: `run` has no note path, and
the binary cannot see a host step at all. After any research or verify step whose
`forge engine select --stage <name>` did not resolve to `none`, stamp it yourself
(`write` is locked to `none` always — `forge engine record` refuses any other tier there):

    forge engine record --stage <name> --tier <the tier that ran: host|api|advisor> --rel <path>

Skip this only for `none` — nothing ran, nothing to stamp. Do it once per stage per note —
a re-verified note gets one fresh `verify=host` line, not a growing history (`forge engine
record`'s own doc-comment).

---

## Quality gate

Superseded by Stage 4 above (`forge gate`, `pkg/qualitygate`'s seven DESIGN §12 gates) —
this section is deliberately short so there is nowhere left for the checklist to drift
out of sync with the CLI it now delegates to. `forge validate <path>` still runs as part
of the gate's schema check; you do not need to run it separately.

**On failure, `forge gate` writes to `_inbox/` with `confidence: low` itself.** Never
publish a note that did not pass, and never quietly fix the gate to make it pass instead
of fixing the note.

---

## Invariants

- Recall runs first. Always. No exceptions for "obvious" questions.
- The profile is read before any note is written, and its effects are visible in the
  output. Two notes on the same question at different `seniority` must not read alike.
- `ANSWER_FROM_VAULT` creates zero files.
- `extend` never rewrites existing body text.
- `refresh` shows a diff before writing.
- Slugs come from `forge slug`, never from you.
- Gate failures go to `_inbox/`, never to a silent publish.
- `forge gate` runs before every write, in every branch — `CREATE` and both `UPDATE`
  modes alike. No branch skips it.
- `forge-researcher` and `forge-codebase-scout` dispatch in parallel, never sequentially.
- `forge-librarian` never writes a note that Stage 4 hasn't already cleared, and stamps
  `Forge-Write: true` on every commit it makes, not only the first.
- One question, at most one note. If the question spans two topics, ask which one —
  do not write both.
