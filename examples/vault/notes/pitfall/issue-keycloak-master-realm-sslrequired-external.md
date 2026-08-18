---
title: "Issue: Keycloak master Realm sslRequired=external"
slug: issue-keycloak-master-realm-sslrequired-external
type: pitfall
stack: [keycloak, docker]
tags: [security, issue]
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

# Issue: Keycloak master Realm sslRequired=external

## Symptom
Keycloak admin console returns "HTTPS required" even though `realm-export.json` has `sslRequired: none`.

## Root Cause
The `sslRequired: none` setting in `realm-export.json` only applies to `meter-readings-realm`,
not to `master`. Keycloak bootstraps `master` before the import runs, and `--import-realm`
uses `IGNORE_EXISTING` strategy — `master` already exists, so it is skipped entirely.

JVM system properties and environment variables (`KC_SPI_REALM_DEFAULT_SSL_REQUIRED`) are
unreliable because `master` is already persisted in the DB volume before they are read.

## Fix
Run `kcadm.sh update realms/master -s sslRequired=NONE` inside the Keycloak container after startup.
Automated via `keycloak-init` init container.

## Affected Files
- `docker-compose-local.yaml` — keycloak-init service added

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/decision/decision-keycloak-init-docker-init-container-for-ssl-fix]]
- [[notes/concept/docker-compose-local-yaml-local-development-docker-compose]]
- [[notes/concept/keycloak-realm-export-json-keycloak-realm-definition]]
