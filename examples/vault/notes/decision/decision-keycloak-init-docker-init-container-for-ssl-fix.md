---
title: "Decision: keycloak-init Docker Init Container for SSL Fix"
slug: decision-keycloak-init-docker-init-container-for-ssl-fix
type: decision
stack: [keycloak, docker]
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

# Decision: keycloak-init Docker Init Container for SSL Fix

**Decision:** Added a `keycloak-init` init container to `docker-compose-local.yaml` that runs
`kcadm.sh update realms/master -s sslRequired=NONE` automatically after Keycloak is healthy.

**Rationale:** Keycloak 25 `--import-realm` skips `master` realm if it already exists in the
volume. The SSL setting on `master` must be patched via `kcadm.sh`. Manual UI step was error-prone
and required after every `docker-compose down -v`.

**Config:**
- `restart: "no"` — must not re-run on Docker daemon restart
- `stop_grace_period: 120s` — `kcadm.sh` handshake takes time
- Health probe: token endpoint (not `/health/ready` which requires `--health-enabled=true`)

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/concept/docker-compose-local-yaml-local-development-docker-compose]]
- [[notes/concept/keycloak-realm-export-json-keycloak-realm-definition]]
- [[notes/pitfall/issue-keycloak-master-realm-sslrequired-external]]
