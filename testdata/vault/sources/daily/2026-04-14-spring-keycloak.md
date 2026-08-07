---
title: "2026-04-14 — Keycloak IdP, soft delete revisited"
tags: [daily, source, keycloak, spring-boot]
source: raw/daily/2026-04-14.md
date: 2026-04-14
status: processed
---

# 2026-04-14 — Keycloak IdP, soft delete revisited

Keycloak wired in as the IdP for the meter-readings service; the realm export is
checked in so the init container can import it on a clean start. Revisited soft delete
while auditing the repository layer — the flag-based approach stays.

## Extracted

- Concepts: [[concepts/soft-deletion]]
