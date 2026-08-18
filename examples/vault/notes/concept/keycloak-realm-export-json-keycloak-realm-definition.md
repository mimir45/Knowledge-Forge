---
title: "keycloak/realm-export.json — Keycloak Realm Definition"
slug: keycloak-realm-export-json-keycloak-realm-definition
type: concept
stack: [keycloak, oauth2]
tags: [google-idp, jwt, security]
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

# keycloak/realm-export.json — Keycloak Realm Definition

Located at `./keycloak/realm-export.json` relative to MeterReadingsService root.
Mounted into the Keycloak container via docker-compose at `/opt/keycloak/data/import`.

## Contents
- Realm: `meter-readings-realm`
- Clients: `postman` (public), `meter-readings-service` (confidential)
- Google IDP stub pre-wired (Client ID/Secret placeholders — fill before first run)
- `sslRequired: none` — only applies to this realm, NOT to `master` realm

## Import Caveat
`--import-realm` uses `IGNORE_EXISTING` strategy — silently skips any realm that already exists.
To force re-import: `docker-compose down -v` first.

`master` realm SSL must be patched via `kcadm.sh`, not JSON import — see [[notes/pitfall/issue-keycloak-master-realm-sslrequired-external]].

## Sources
- [[sources/daily/2026-04-14-spring-keycloak-postman]]

## Related
- [[notes/concept/docker-compose-local-yaml-local-development-docker-compose]]
- [[notes/decision/decision-keycloak-init-docker-init-container-for-ssl-fix]]
- [[notes/pitfall/issue-keycloak-master-realm-sslrequired-external]]
