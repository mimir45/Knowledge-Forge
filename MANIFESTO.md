# The Knowledge Forge Manifesto

You asked a model to explain something hard. It gave you a genuinely good answer.
Then you closed the tab.

Next month the same question comes back, and it costs exactly what it cost the first
time. Not because the answer was wrong — because nothing kept it.

That is the problem Knowledge Forge exists for: **an explanation you paid for once
should never have to be paid for twice.** Every "explain X" moment becomes a permanent,
linked, verified markdown note in your own vault. The second time the question comes up,
it's a file read instead of a research run.

This document is not a feature list. It's the set of positions the software takes — the
things it refuses to do, and why.

---

## 1. No model call where a deterministic answer exists.

Finding which of your notes already answers a question is string matching, ranking, and
graph traversal. It is not reasoning. So it doesn't get a model.

The entire static core — recall, deduplication, schema validation, drift detection,
quality gates, report generation — makes **zero model calls**. No API key. No network.
It runs on a plane.

An LLM layer sits on top, and it's optional, and it's configurable per pipeline stage.
It is never load-bearing for correctness. Only for enrichment.

> If a step can be right without a model, giving it a model makes it slower, more
> expensive, and less trustworthy — in exchange for nothing.

## 2. No embeddings.

Recall is hand-rolled MinHash + LSH over lexical signatures. Deliberately.

Cosine similarity over embeddings is confidently vague. It cannot tell the difference
between *the same idea stated twice* and *two adjacent topics* — and for a duplicate
report, that distinction is the entire job. A near-duplicate you merge by mistake is
knowledge you destroyed.

A vector database is also a service, a dimension count, a model version, and a
reindexing story. All of that, to answer a question a bag of shingles answers better.

## 3. Markdown is the only source of truth.

Your notes are plain files, in your vault, in your git repo. You can read them in a text
editor in twenty years.

There is a SQLite cache. It is derived, disposable, and rebuildable from the markdown by
a single command — never the other way around. Nothing lives only in the database.
Delete it and lose nothing but a few milliseconds.

> A knowledge base you can only read through its own software is not your knowledge.
> It's someone's schema.

## 4. The strongest tier critiques. It never returns a rewrite.

The top tier is an advisor, and an advisor's output is a list of disputed claims and a
patch. Never a replacement draft.

The moment the reviewing model is allowed to rewrite the note wholesale, the note stops
being your understanding and becomes the model's opinion of it. Drafting is a step you
can accept or reject. Rewriting is a step that quietly replaces you.

## 5. Nothing gets published silently.

Seven quality gates — schema, citations, freshness, link integrity, duplication,
compiled code, anti-slop — run before a note enters the vault.

A draft that fails does not get fixed quietly and shipped. It goes to `_inbox/` with
`confidence: low`, and it waits for you.

> Silent success on unverified content is how a knowledge base becomes a liability.
> A quarantine folder is a feature.

## 6. Never auto-mutate the vault on a schedule.

No background daemon rewriting your notes at 3am. No cron job "improving" prose you
already approved.

Every mutation is either something you ran, or something a git commit triggered — and
you can read exactly what it did afterwards.

## 7. Drift is anchored to git, never to your editor.

Notes cite code. Code moves. A note whose citations no longer match reality is worse
than no note — it's confidently stale.

So drift checking runs on `post-commit`, `post-merge`, `post-checkout`. Never on file
save. Never against your uncommitted working tree, which is a construction site, not a
statement of fact.

Verdicts are a pure function of (note references, tree state). Revert the commit and the
demoted notes come back — symmetrically, mechanically, no history rewriting. It runs on
the git hook path, so it has a hard latency budget: under 100ms, or it's a bug.

## 8. Never log what you asked.

The ask log stores a topic label and a sha256 hash. Never the question text. Never your
code. Never file contents.

Training pairs accumulate as a byproduct of normal use, under the same rule. Exporting
them is a manual command, anonymized by default, and it **fails closed** — it emits
nothing rather than emit something it couldn't redact.

## 9. Nothing leaves your machine.

There is no upload path in this codebase. No telemetry endpoint. No sync. No phone-home.

The only outbound network access anywhere is an HTTP HEAD request against URLs *your own
notes cite*, to check they still resolve (`pkg/linkcheck`) — plus whichever LLM tier you
configured yourself, if you configured one (`pkg/engine`, `cmd/forge/engine_run.go`).
That is the complete list. Nothing else in the tree opens a socket.

This is not a privacy policy. It's an architectural fact, and you can grep for it.

## 10. No daemon on speculation. Measure first.

CLI only for v1. Not because a daemon would be wrong, but because nobody has measured
that it's needed.

Every latency budget in the project is a number, checked against a real vault, not a
vibe. Performance claims that can't be reproduced aren't claims.

---

## What this adds up to

A Go static-analysis engine, ~17,500 lines across 18 packages, with ~10,800 lines of
tests. Twenty-two CLI subcommands. Zero model calls in the core. It ships as a Claude
Code plugin, and it works standalone.

Not because the model layer is bad. Because **the durable part of a knowledge system has
to be the part that doesn't depend on a model being available, affordable, or the same
version it was last year.**

Your understanding compounds. Your notes should too.

---

## It's still early

Knowledge Forge is in beta. Things will change, edges are rough, and I'm working through
them in the open. The positions above are settled; the software implementing them is
not.

If you try it: bug reports, feature ideas, and blunt feedback are all genuinely useful
right now.

**[github.com/mimir45/Knowledge-Forge](https://github.com/mimir45/Knowledge-Forge)** ·
MIT
