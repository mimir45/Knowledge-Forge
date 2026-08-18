---
title: "useFocusOnMount — Shared React Focus Hook"
slug: usefocusonmount-shared-react-focus-hook
type: concept
stack: [react, typescript]
tags: [hooks, accessibility, leprecoin]
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

# useFocusOnMount — Shared React Focus Hook

Extracts the repeated `useRef + useEffect` focus-on-mount pattern shared across
`CheckEmail`, `ForgotPasswordForm`, and `ResetPasswordForm`.

## Implementation

```ts
// src/shared/lib/hooks/useFocusOnMount.ts
export function useFocusOnMount<T extends HTMLElement>(): React.RefObject<T | null> {
  const ref = useRef<T | null>(null);
  useEffect(() => { ref.current?.focus(); }, []);
  return ref;
}
```

Return type must be `React.RefObject<T | null>` — modern React's `useRef<T>(null)` returns
`RefObject<T | null>`, so omitting `| null` causes a type mismatch.

## Sources

- [[sources/daily/2026-04-21-react-resetpassword-refactor]]

## Related

- [[notes/decision/decision-extract-usefocusonmount-hook]]
- [[notes/concept/leprecoin-react-next-js-frontend-project]]
