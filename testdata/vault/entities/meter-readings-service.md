---
title: "MeterReadingsService"
tags: [spring-boot, java, service, entity]
source: sources/daily/2026-04-13-local-ai-spring.md
date: 2026-04-13
status: active
---

# MeterReadingsService

Spring Boot 4.x service exposing meter readings over a contract-first OpenAPI surface.

## Notes

- Soft delete throughout — see [[concepts/soft-delete]].
- Cursor pagination on the list endpoint — see [[TIL/databases/keyset-cursor-pagination]].
- Deployment topology is documented in [[entities/does-not-exist]].

## Decisions applying to this service

- [[decisions/liquibase-over-column-alias]]
