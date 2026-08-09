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

- Prefer official documentation and primary sources. Record every one in `sources:`
  with the date consulted.
- Version-pin claims that are version-dependent. "Spring Boot does X" ages badly;
  "Spring Boot 4.0 does X" can be checked.
- **Compile and run any code you assert works** — in a throwaway directory, never in
  the user's project. Untested code in a note is a future pitfall note.
- Mark anything you could not verify as such in the note body. A hedge in the text is
  honest; a confident sentence you could not check is a lie the vault will repeat.

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

Before publishing a new or extended note:

- `forge validate <path>` exits 0.
- Every claim has a source, or is explicitly marked unverified.
- Every code block was executed.
- Every `[[link]]` resolves.

**On failure, write to `_inbox/` with `confidence: low` and tell the user what failed.**
Never publish a note that did not pass, and never quietly fix the gate to make it pass.

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
- One question, at most one note. If the question spans two topics, ask which one —
  do not write both.
