---
title: "Storybook Decorator Pattern — Provider Wrappers for Redux and Context"
slug: storybook-decorator-pattern-provider-wrappers-for-redux-and-context
type: concept
stack: [storybook, react, redux, nextjs]
tags: [testing, leprecoin]
depth: 3
confidence: low
created: 2026-04-22
updated: 2026-04-22
verified: 2026-04-22
freshness_days: 365
sources:
  - url: sources/daily/2026-04-22-storybook-coverage-audit.md
    accessed: 2026-04-22
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Storybook Decorator Pattern — Provider Wrappers for Redux and Context

Stories for components that depend on context or Redux must wrap the story in the appropriate
provider decorators. Failing to do so causes runtime errors in Storybook.

## Decorator Map

| Component type | Required decorators |
|---------------|---------------------|
| Redux-dependent forms | `Provider + ToastProvider` |
| OTP / SignUp flows | `SignUpContextProvider` |
| Plain components | none |

## Convention

```ts
export default {
  decorators: [
    (Story) => (
      <Provider store={store}>
        <ToastProvider>
          <Story />
        </ToastProvider>
      </Provider>
    ),
  ],
} satisfies Meta<typeof MyComponent>;
```

## Page-Level Wrappers

Full-page wrappers (`LoginPage`, `SignUpPage`, etc.) that only compose already-storied children
with no isolated visual logic are skipped — they add no value and double maintenance cost.

See [[notes/decision/decision-skip-storybook-stories-for-page-level-wrapper-components]].

## Sources

- [[sources/daily/2026-04-22-storybook-coverage-audit]]

## Related

- [[notes/concept/leprecoin-react-next-js-frontend-project]]
- [[notes/decision/decision-skip-storybook-stories-for-page-level-wrapper-components]]
