---
title: "A Test That Queries the DOM Without Rendering Always Passes Vacuously"
slug: a-test-that-queries-the-dom-without-rendering-always-passes-vacuously
type: howto
stack: [react, jsdom, jest]
tags: [testing, frontend]
depth: 3
confidence: low
created: 2026-04-17
updated: 2026-04-17
verified: 2026-04-17
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# A Test That Queries the DOM Without Rendering Always Passes Vacuously

## What is it?

In JSDOM (used by Jest/Vitest), querying the DOM without first rendering a component returns an empty collection — which passes all "not found" assertions. A test structured this way appears green but tests nothing at all.
 
## How it works

```typescript
// BROKEN — component never rendered; queries run against bare JSDOM
it('shows skeletons', () => {
  const skeletons = document.querySelectorAll('*');
  expect(skeletons.length).toBeGreaterThan(0); // PASSES vacuously if DOM has anything
  // or worse:
  expect(skeletons.length).toBe(0); // also PASSES on empty DOM
});
```

```typescript
// CORRECT — component rendered first, then queried
import { render, screen } from '@testing-library/react';

it('shows skeletons while loading', () => {
  render(<AccountsOverviewLoader />);
  const skeletons = screen.getAllByTestId('skeleton');
  expect(skeletons).toHaveLength(3);
});
```

## Key implementation steps

Always follow the render → query → assert pattern:

```typescript
import { render, screen } from '@testing-library/react';
import { AccountsOverviewLoader } from './index';

describe('AccountsOverviewLoader', () => {
  it('renders the correct number of skeletons', () => {
    render(<AccountsOverviewLoader />);
    // Query AFTER render
    expect(screen.getAllByTestId('skeleton')).toHaveLength(3);
  });
});
```

Use `data-qa` or `data-testid` attributes on Skeleton instances to make assertions precise:
```tsx
<Skeleton data-qa="skeleton" width="100%" height={48} />
```

## Common pitfalls

- `document.querySelectorAll('*')` against bare JSDOM always returns at least `<html>`, `<head>`, `<body>` — length > 0 passes even without a component
- Tests written this way give false confidence — CI stays green while the component is broken or deleted
- Check that you see the component actually imported and called in the test file before trusting it

## When to use / not use

Always render before querying. Use `@testing-library/react`'s `render()` and `screen` API. The only exception is testing pure functions or hooks (with `renderHook()`), which have their own render lifecycle.
