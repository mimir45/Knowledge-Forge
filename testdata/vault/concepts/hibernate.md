---
title: "Hibernate — ORM Patterns and Gotchas"
tags: [hibernate, spring-boot, jpa, orm, java]
source: sources/daily/2026-04-13-local-ai-spring.md
date: 2026-04-13
status: active
---

# Hibernate — ORM Patterns and Gotchas

Key patterns and known pitfalls when using Hibernate 6 with Spring Boot.

## Soft Delete

Use `@SQLRestriction("deleted_at IS NULL")` — the Hibernate 6 replacement for the
deprecated `@Where` annotation. See [[concepts/soft-delete]].

## Flush Timing

`@CreationTimestamp` is populated at flush, not at `save()`. A DTO built from the
returned entity therefore carries a null timestamp unless the method calls
`saveAndFlush()`. See [[decisions/liquibase-over-column-alias]] for the related column
naming decision and [[issues/hibernate-column-mismatch]] for the failure it caused.
