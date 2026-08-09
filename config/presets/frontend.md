---
# frontend — for a React / Next.js developer.
#
# Code indexing covers TypeScript (the tree-sitter grammar handles .ts and .tsx).
# JavaScript is listed because notes cite it; the parser treats it as TypeScript.
#
# Two deliberate differences from the defaults. howto notes go stale in 90 rather than
# 180 days — the framework churn here is faster than the doc's default assumed, and a
# Next.js howto from two releases ago is usually wrong rather than merely dated. And
# linkcheck matters more, because frontend notes cite blog posts and changelogs that
# actually disappear, where backend notes cite javadoc that does not.

static:
  code_index: true
  languages: [typescript, javascript]
  git_signals: true
  linkcheck: {enabled: true, timeout_s: 8}

freshness_days:
  howto: 90
  api: 60

research:
  prefer: [official-docs, source-code]

write:
  diagrams: mermaid
---
