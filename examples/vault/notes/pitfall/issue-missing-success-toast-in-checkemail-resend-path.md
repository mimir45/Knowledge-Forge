---
title: "Issue: Missing Success Toast in CheckEmail Resend Path"
slug: issue-missing-success-toast-in-checkemail-resend-path
type: pitfall
stack: [react]
tags: [leprecoin, frontend, ux]
depth: 3
confidence: low
created: 2026-04-17
updated: 2026-04-17
verified: 2026-04-17
freshness_days: 365
sources:
  - url: sources/daily/2026-04-17-storybook-llm-wiki.md
    accessed: 2026-04-17
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Issue: Missing Success Toast in CheckEmail Resend Path

## Status

**Not implemented** — identified as a one-liner fix.

## Symptom

The `CheckEmail` resend button submits successfully but shows no feedback to the user.
A success toast should appear confirming the email was resent.

## Fix

Add a toast call in the resend success handler and an i18n key for the message:

```ts
onResendSuccess: () => toast.success(t('checkEmail.resendSuccess'))
```

Add the key `checkEmail.resendSuccess` to translation files.

## Sources

- [[sources/daily/2026-04-17-storybook-llm-wiki]]

## Related

- [[notes/concept/leprecoin-react-next-js-frontend-project]]
- [[notes/concept/open-questions-unresolved-topics-and-open-threads]]
