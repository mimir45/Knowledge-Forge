---
title: "Docker Compose Init Container Pattern with Health-Gated Sequencing"
slug: docker-compose-init-container-pattern-with-health-gated-sequencing
type: howto
stack: [docker, keycloak]
tags: [docker-compose, devops]
depth: 3
confidence: low
created: 2026-04-17
updated: 2026-04-17
verified: 2026-04-17
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Docker Compose Init Container Pattern with Health-Gated Sequencing

## What is it?

Docker Compose (v2) supports an init container pattern: a short-lived service that runs one-time setup tasks only after another service is healthy, then exits. This is the correct way to automate tasks like patching Keycloak realm settings that can't be done via configuration files.

## How it works

```yaml
services:
  keycloak:
    image: quay.io/keycloak/keycloak:25
    healthcheck:
      test: ["CMD", "curl", "-f",
             "http://localhost:8080/realms/master/protocol/openid-connect/token"]
      interval: 10s
      timeout: 5s
      retries: 10

  keycloak-init:
    image: quay.io/keycloak/keycloak:25
    depends_on:
      keycloak:
        condition: service_healthy   # waits for healthcheck to pass
    restart: "no"                    # run once, don't restart
    stop_grace_period: 120s          # give kcadm.sh time to finish
    entrypoint: ["/bin/sh", "-c"]
    command:
      - |
        /opt/keycloak/bin/kcadm.sh config credentials \
          --server http://keycloak:8080 --realm master \
          --user admin --password admin1234
        /opt/keycloak/bin/kcadm.sh update realms/master -s sslRequired=NONE
```

## Common pitfalls

- **`restart: "no"` is required** — without it, Docker restarts the init container on every daemon restart
- **Containers on the same network reach each other by service name** — use `http://keycloak:8080` not `http://localhost:8080`
- **Keycloak 25 `start-dev` does NOT expose `/health/ready`** — requires `--health-enabled=true` flag. Use the token endpoint as a health probe instead
- **JVM properties via `JAVA_OPTS_APPEND` don't override realm settings already stored in the DB volume** — only `kcadm.sh` reliably patches a running realm
- **`--import-realm` silently skips any realm that already exists** (including `master`) — never use JSON import to patch `master` realm settings

## When to use / not use

Use this pattern for one-time setup that depends on a service being fully ready: database seeding, Keycloak realm patching, schema migrations, secret injection. Don't use it for tasks that should run on every container restart.
