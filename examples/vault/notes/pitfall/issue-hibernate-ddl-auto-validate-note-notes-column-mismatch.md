---
title: "Issue: Hibernate ddl-auto=validate note/notes Column Mismatch"
slug: issue-hibernate-ddl-auto-validate-note-notes-column-mismatch
type: pitfall
stack: [hibernate, liquibase, spring-boot]
tags: [issue]
depth: 3
confidence: low
created: 2026-04-13
updated: 2026-04-13
verified: 2026-04-13
freshness_days: 365
sources:
  - url: sources/daily/2026-04-13-local-ai-continue-rag-spring.md
    accessed: 2026-04-13
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Issue: Hibernate ddl-auto=validate note/notes Column Mismatch

## Symptom
Application fails to start with Hibernate `ddl-auto=validate` complaining about a missing column.

## Root Cause
Entity field named `notes` (plural) but the database column was `note` (singular) — a
singular/plural mismatch between the Liquibase migration and the Java entity field name.

## Fix
Added Liquibase changeset 006 to `RENAME COLUMN note TO notes`. Entity field name and DB column
name are now in sync with no `@Column` alias needed.

## Affected Files
- Liquibase changeset 006 (new)
- `Meter.java` entity (no change needed — field was already `notes`)

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/decision/decision-liquibase-migration-over-column-alias-for-column-rename]]
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
