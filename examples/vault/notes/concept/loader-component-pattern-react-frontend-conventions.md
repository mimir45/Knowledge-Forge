---
title: "Loader Component Pattern — React Frontend Conventions"
slug: loader-component-pattern-react-frontend-conventions
type: concept
stack: [react, storybook, styled-components]
tags: [accessibility, testing, frontend]
depth: 3
confidence: low
created: 2026-04-16
updated: 2026-04-16
verified: 2026-04-16
freshness_days: 365
sources:
  - url: sources/daily/2026-04-16-frontend-code-review.md
    accessed: 2026-04-16
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Loader Component Pattern — React Frontend Conventions

Canonical reference: `LegalDocumentLoader`. All loader components should follow this pattern.

## Required Accessibility Attributes
```tsx
<div
  role="status"
  aria-live="polite"
  aria-busy="true"
>
  <Skeleton data-qa="skeleton" />
</div>
```

## Styled Components Rules
- Use `styled.div` (not `styled.section`) for layout wrappers
- Use `width: 100%` (not hardcoded pixels)
- Use theme tokens for colors/spacing

## Testing Requirements
- Must render the component before querying the DOM (`render(<Component />)`)
- Never use `document.querySelectorAll('*')` against bare JSDOM — always passes vacuously
- Use `data-qa="skeleton"` selectors for skeleton count assertions
- Use `@testing-library/react`

## Storybook
- Use `autodocs` and `centered` layout
- Include stories for all meaningful states

## Sources
- [[sources/daily/2026-04-16-frontend-code-review]]
- [[sources/daily/2026-04-17-storybook-llm-wiki]]

## Related
