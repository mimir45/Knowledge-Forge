---
title: "Storybook + Next.js App Router Requires appDirectory: true Parameter"
slug: storybook-plus-next-js-app-router-requires-appdirectory-true-parameter
type: howto
stack: [storybook, nextjs, react]
tags: [frontend]
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

# Storybook + Next.js App Router Requires appDirectory: true Parameter

## What is it?

`@storybook/nextjs` defaults to mocking the **Pages Router** (`next/router`). Components that use App Router hooks (`next/navigation` — `useRouter`, `usePathname`, `useSearchParams`) will throw `SB_FRAMEWORK_NEXTJS_0002` unless you explicitly tell Storybook to initialize App Router mocks instead.

## How it works

Without the parameter, `next/navigation` hooks are not mocked and throw at story render time:

```
Error: SB_FRAMEWORK_NEXTJS_0002 - This hook can only be used in an App Router component
```

## Key implementation steps

Add `nextjs.appDirectory: true` to the story's parameters:

```typescript
// Component-level (applies to all stories in the file)
const meta: Meta<typeof WelcomePage> = {
  title: 'Widgets/WelcomePage',
  component: WelcomePage,
  parameters: {
    layout: 'fullscreen',
    nextjs: {
      appDirectory: true,   // initializes next/navigation mocks
    },
  },
};
```

Or in `.storybook/preview.ts` to apply globally:
```typescript
export const parameters = {
  nextjs: {
    appDirectory: true,
  },
};
```

**For RTK Query mutations without MSW**, use `@storybook/test` fn mocks in `beforeEach`:

```typescript
import { fn } from '@storybook/test';
import * as mutations from '@/features/auth/api';

const mockMutate = fn();

export const Loading: Story = {
  beforeEach() {
    mockMutate.mockReturnValue([fn(), { isLoading: true }]);
    vi.spyOn(mutations, 'useResetPasswordMutation')
      .mockReturnValue(mockMutate());
  },
};
```

## Common pitfalls

- Only `next/navigation` hooks are affected — Pages Router components don't need this flag
- The flag must be set wherever the component is rendered, not just at the root level if using decorators
- MSW is not set up in Storybook by default — use fn-based mocks for API calls

## When to use / not use

Add `appDirectory: true` to any story that renders a component using `useRouter`, `usePathname`, `useSearchParams`, or `useParams` from `next/navigation`. Pages Router components using `next/router` work without this flag.
