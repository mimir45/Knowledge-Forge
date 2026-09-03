# Writing rules

The anti-slop gate reads this file. `## Banned phrases` is parsed by
`pkg/qualitygate/antislop.go` at gate-run time — add a line here to reject a phrase
without a recompile, the same way `references/schema.yaml` lets `forge validate` grow
without one. `## Structural requirements` names the one rule antislop.go also checks
mechanically; `## Rationale` is human prose only, never parsed.

## Banned phrases

- in today's fast-paced world
- it's important to note that
- it is important to note that
- let's dive in
- in conclusion
- game changer
- game-changer
- cutting-edge
- cutting edge
- leverage
- utilize
- delve into
- unlock the power of
- unlock the potential of
- seamlessly
- robust and scalable
- in the realm of
- at the end of the day
- when it comes to
- it goes without saying
- needless to say
- as an ai language model

## Structural requirements

- Notes of type `howto` or `api` must contain at least one fenced code block. A
  procedure or an API reference with zero code is a claim without a demonstration —
  exactly the gap this gate was built to catch. `concept`, `pattern`, `pitfall`,
  `decision`, and `incident` are not required to (a `decision` note may be pure prose).

## Rationale

The banned-phrase list targets stock LLM filler, not domain vocabulary — "leverage" and
"utilize" are on the list because they are near-universal padding for "use," not because
the words are wrong in every context. A note tripping this gate almost always reads
better with the phrase deleted outright rather than replaced.

The structural rule is deliberately narrow. It would be easy to also require a minimum
word count, a "Why it matters" heading, or a references section, but `cfg.Write.MaxNoteWords`
already owns the length axis and this gate has exactly one job — catching
prose that performs depth without demonstrating it. Anything else belongs in a template
(`templates/`), not a gate that blocks a write.

This list grows from `_inbox/` reject history, not from guessing in advance: when a
draft gets quarantined for a phrase not yet listed here, add it, don't special-case it
in Go.
