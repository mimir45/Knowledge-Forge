---
title: "docker-compose-local.yaml — Local Development Docker Compose"
slug: docker-compose-local-yaml-local-development-docker-compose
type: concept
stack: [docker, keycloak, postgresql, testcontainers]
tags: [local-dev]
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

# docker-compose-local.yaml — Local Development Docker Compose

Located at project root. Runs the full local development stack.

## Services
- **PostgreSQL** — primary database
- **Keycloak** — identity provider; mounts `./keycloak` for realm import; health probe at `/realms/master` (not `/health/ready` — requires `--health-enabled=true` flag which is off by default in `start-dev`)
- **keycloak-init** — init container; runs `kcadm.sh update realms/master -s sslRequired=NONE` after Keycloak healthcheck passes; `restart: "no"`, `stop_grace_period: 120s`

## Key Notes
- Containers on the same Docker network reach each other by service name (e.g., `http://keycloak:8080`)
- `depends_on: condition: service_healthy` used for init container sequencing
- Pre-existing issue: Keycloak service missing `utility-network` definition (non-blocking)

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]
- [[sources/daily/2026-04-15-spring-refactor-testcontainers]]

## Related
- [[notes/concept/keycloak-realm-export-json-keycloak-realm-definition]]
- [[notes/decision/decision-keycloak-init-docker-init-container-for-ssl-fix]]
- [[notes/pitfall/issue-keycloak-master-realm-sslrequired-external]]
