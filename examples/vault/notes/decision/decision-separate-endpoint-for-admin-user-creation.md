---
title: "Decision: Separate Endpoint for Admin User Creation"
slug: decision-separate-endpoint-for-admin-user-creation
type: decision
stack: [spring-boot, spring-security]
tags: [security, decision]
depth: 3
confidence: high
created: 2026-04-14
updated: 2026-04-14
verified: 2026-04-14
freshness_days: 730
sources:
  - url: sources/daily/2026-04-14-spring-keycloak-postman.md
    accessed: 2026-04-14
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Decision: Separate Endpoint for Admin User Creation

**Decision:** Admin users are created via a separate `POST /api/admin/users` endpoint secured with
`@PreAuthorize("hasRole('ADMIN')")`, rather than an `isAdmin` field in the request body.

**Rationale:** A role field in the request body is a security hole — any client can self-promote
to admin. A separate endpoint with `@PreAuthorize` enforces authorization at the method level.
Belt-and-suspenders: route is also `authenticated()` in `SecurityFilterChain`.

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/concept/meterreadingsservice-spring-boot-4-x-project]]
