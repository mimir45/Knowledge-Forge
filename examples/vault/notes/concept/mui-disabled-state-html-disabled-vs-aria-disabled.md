---
title: "MUI Disabled State — HTML disabled vs aria-disabled"
slug: mui-disabled-state-html-disabled-vs-aria-disabled
type: concept
stack: [mui, react, typescript]
tags: [testing, accessibility]
depth: 3
confidence: low
created: 2026-04-21
updated: 2026-04-21
verified: 2026-04-21
freshness_days: 365
sources:
  - url: sources/daily/2026-04-21-react-resetpassword-refactor.md
    accessed: 2026-04-21
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# MUI Disabled State — HTML disabled vs aria-disabled

MUI `Button` with `disabled={true}` sets the actual HTML `<button disabled>` attribute.
This is a hard browser-level block, unlike CSS-only or `aria-disabled` states.

## Testing Implications

`@testing-library/user-event` throws a `pointer-events: none` error when attempting to click
a button with the HTML `disabled` attribute. Tests must:

- Assert `expect(button).toBeDisabled()` — never try to click it
- Remove tests that attempt to click disabled buttons with invalid form data

## aria-disabled Duplication

MUI `Button` already sets `aria-disabled` from its `isDisabled` prop. Don't add `aria-disabled`
manually to a component that wraps `Button` — it creates a double-assertion.

## Sources

- [[sources/daily/2026-04-21-react-resetpassword-refactor]]

## Related

- [[notes/concept/leprecoin-react-next-js-frontend-project]]
- [[notes/decision/decision-rewrite-tests-to-assert-tobedisabled-not-click]]
