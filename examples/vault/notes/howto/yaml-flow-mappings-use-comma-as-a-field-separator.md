---
title: "YAML Flow Mappings Use Comma as a Field Separator"
slug: yaml-flow-mappings-use-comma-as-a-field-separator
type: howto
stack: [liquibase]
tags: [yaml, debugging]
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

# YAML Flow Mappings Use Comma as a Field Separator

## What is it?

In YAML flow mappings (`{ key: value, key: value }`), a comma is a field separator — not a literal character. Any unquoted value containing a comma will be silently split, producing a cryptic "Unexpected node" parse error that names only the fragment after the comma, not the full offending value.

## How it works

```yaml
# BROKEN — comma in address splits the value
- { column: address, value: 123 Maple Ave, Apt 1 }
#                                         ↑
#                         YAML parser sees: value="123 Maple Ave" then "Apt 1" as a new key
```

Error: `ChangeLogParseException: Unexpected node: Apt 1`

The error names `Apt 1` (the fragment after the split), not the full address. Look one token earlier in the YAML source to find the actual bad value.

```yaml
# CORRECT — double-quoted value
- { column: address, value: "123 Maple Ave, Apt 1" }

# Also correct — single-quoted
- { column: address, value: '123 Maple Ave, Apt 1' }
```

## Key implementation steps

When writing YAML flow mappings, quote any value that might contain:
- Commas: addresses, lists, CSV fragments
- Colons: URLs, time values (`12:00`)
- Curly/square braces: template strings
- Special characters: `#`, `&`, `*`, `!`

For Liquibase YAML changesets, prefer block style for complex data:
```yaml
changes:
  - insert:
      tableName: users
      columns:
        - column:
            name: address
            value: "123 Maple Ave, Apt 1"
```

## Common pitfalls

- Error message names the fragment *after* the split, not the actual bad value — look earlier in the file
- Single-quoted strings prevent splitting and work the same as double-quoted for this purpose
- Block-style YAML (`key: value` on separate lines) does not have this problem — commas are literals

## When to use / not use

Prefer block-style YAML for data with complex values. Use flow mappings (`{...}`) only for short, simple records with no special characters in values.
