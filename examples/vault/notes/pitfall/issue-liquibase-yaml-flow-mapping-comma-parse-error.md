---
title: "Issue: Liquibase YAML Flow Mapping Comma Parse Error"
slug: issue-liquibase-yaml-flow-mapping-comma-parse-error
type: pitfall
stack: [liquibase]
tags: [yaml, issue]
depth: 3
confidence: low
created: 2026-04-14
updated: 2026-04-14
verified: 2026-04-14
freshness_days: 365
sources:
  - url: sources/daily/2026-04-14-spring-keycloak-postman.md
    accessed: 2026-04-14
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Issue: Liquibase YAML Flow Mapping Comma Parse Error

## Symptom
`ChangeLogParseException: Unexpected node: Apt 1` on application startup.

## Root Cause
Address value `123 Maple Ave, Apt 1` appeared inside a YAML flow mapping
(`{ key: val, ... }`). YAML flow mappings use `,` as a field separator — the parser split
the value mid-address and treated `Apt 1` as a spurious second key.

## Fix
Double-quote the address value so the comma is treated as a literal character:
`"123 Maple Ave, Apt 1"`

## Affected Files
- `src/main/resources/db/changelog/2026-04-12-007-insert-mock-data.yaml`

## Notes
- Error message names the fragment (`Apt 1`), not the full bad value — look one token earlier
- Single-quoted strings also work; common in Liquibase YAML `value:` fields

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
