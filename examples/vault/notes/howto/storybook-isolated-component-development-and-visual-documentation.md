---
title: Storybook — Isolated Component Development & Visual Documentation
slug: storybook-isolated-component-development-and-visual-documentation
type: howto
stack: [storybook, react]
tags: [frontend, component-library, testing, documentation]
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

# Storybook — Isolated Component Development & Visual Documentation

## What is it?

Storybook is an open-source tool for building UI components in isolation from the rest of the application. Each component state is captured as a **story** — a named render of that component with a specific set of props/args. The result is an interactive catalogue you can browse, share with designers, and test without spinning up the full app.

## How it works

Storybook starts its own webpack (or Vite) dev server completely separate from the Next.js app server. It scans the file system for files matching the `stories` glob in `main.ts` (e.g. `../src/**/*.stories.@(js|jsx|ts|tsx)`), bundles them in its own sandbox, and mounts each story inside an `<iframe>` in the browser UI.

Key pipeline steps:

1. **`main.ts`** — declares the framework adapter (e.g. `@storybook/nextjs`), addon list, and any webpack/Vite customisations.
2. **`preview.ts/tsx`** — runs once before every story. This is where global decorators (theme providers, i18n) and global parameters (controls config, a11y rules) live.
3. **Addons** — plugins that inject panels into the Storybook UI. They communicate via an event bus (`@storybook/core-events`). Common addons: `addon-docs`, `addon-a11y`, `addon-controls`, `addon-actions`.
4. **Story bundling** — the framework adapter (e.g. `@storybook/nextjs`) shims Next.js-specific APIs (`next/image`, `next/router`, `next/navigation`) so they work inside the Storybook sandbox.

## Key concepts

### Meta object

```tsx
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from '@/shared/ui/Button';

const meta: Meta<typeof Button> = {
  title: 'shared/Button',   // Sidebar path — slash = folder nesting
  component: Button,
  tags: ['autodocs'],        // Auto-generate a Docs page from JSDoc + stories
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Primary action button used across the app.',
      },
    },
  },
  argTypes: {
    variation: {
      control: 'select',
      options: ['primary', 'secondary', 'tertiary'],
    },
  },
  args: { type: 'button', isDisabled: false },
};

export default meta;
```

### Story objects

```tsx
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: { variation: 'primary', children: 'Submit' },
};

export const Loading: Story = {
  args: { variation: 'primary', isLoading: true, children: 'Loading...' },
};
```

### Decorators

Decorators wrap every story with extra JSX — perfect for providers:

```tsx
// Applied globally in .storybook/preview.tsx
decorators: [
  (Story) => (
    <MuiThemeProvider theme={theme}>
      <ThemeProvider theme={theme}>
        <Story />
      </ThemeProvider>
    </MuiThemeProvider>
  ),
],
```

### Args & controls

`args` are the props passed to the component. The Controls panel auto-generates an interactive UI for them based on TypeScript types or explicit `argTypes`. Changing a control re-renders the story in real time.

### Parameters

Story-level configuration consumed by addons or Storybook itself. Examples: `layout`, `backgrounds`, `viewport`, `a11y.disable`, `docs.description`.

## Why teams use it

- **Isolation** — develop components without wiring up a real backend, auth flow, or page layout. Reproduce exact edge cases (empty state, loading, error) on demand.
- **Living documentation** — the `autodocs` tag generates a Docs page combining rendered previews, prop tables (from TS types), and description strings.
- **Design handoff** — designers review actual rendered components, not static mockups.
- **Visual regression testing** — tools like Chromatic snapshot every story and flag pixel-level diffs on each PR.
- **Accessibility auditing** — `@storybook/addon-a11y` runs axe-core against every story and surfaces violations in the A11y panel.
- **Reuse discovery** — browsing the catalogue prevents developers from re-implementing components that already exist.

## Integration with styled-components & MUI

This project uses both theme systems merged into a single `theme` object. The `preview.tsx` wraps every story with both providers and applies `CssBaseline` + `GlobalStyles`:

```tsx
// .storybook/preview.tsx
decorators: [
  (Story) => (
    <MuiThemeProvider theme={theme}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <GlobalStyles />
        <NextIntlClientProvider locale="en" messages={messages}>
          <Story />
        </NextIntlClientProvider>
      </ThemeProvider>
    </MuiThemeProvider>
  ),
],
```

The webpack alias in `main.ts` redirects `@mui/styled-engine` → `@mui/styled-engine-sc` so MUI uses styled-components as its CSS-in-JS engine — the same alias used in the Next.js app. Without this alias, MUI would fall back to Emotion and styles would conflict.

## Common pitfalls

- **Missing providers in preview.tsx** — components using `useTheme()` or `useTranslations()` will crash without the decorator wrapping them.
- **Next.js API mismatches** — `@storybook/nextjs` shims most Next.js APIs but they are mocks. Don't rely on real router state (e.g. `useSearchParams`) without a manual mock decorator.
- **JSX as args** — passing JSX as an `arg` (e.g. `children: <Icon />`) breaks the Controls panel; disable the control via `argTypes` for that prop.
- **`autodocs` with no named exports** — a story file with only a default export produces an empty Docs page. Always export at least one named story.
- **Stale story cache** — after renaming a story title, a full dev server restart may be needed.
- **Module resolution divergence** — path aliases (`@/`) must be mirrored in `webpackFinal`. Missing aliases cause "Cannot find module" errors only in Storybook.
- **`layout: 'centered'` on full-width components** — centering a component that expects `width: 100%` will make it collapse. Use `layout: 'fullscreen'` or a width wrapper decorator.

## When to use / not use

**Use when:**
- Building shared components in `shared/ui/*`
- The component has multiple visual states (loading, error, empty, success)
- You want designers to review work without a staging deploy
- Setting up visual regression CI (Chromatic)

**Skip or defer when:**
- The component is tightly coupled to a specific RTK Query slice and only makes sense end-to-end (write an integration test instead)
- The component is a one-off page layout with no reuse potential

## References

- [Storybook official docs](https://storybook.js.org/docs)
- [@storybook/nextjs framework adapter](https://storybook.js.org/docs/get-started/frameworks/nextjs)
- [Chromatic visual testing](https://www.chromatic.com/docs/)