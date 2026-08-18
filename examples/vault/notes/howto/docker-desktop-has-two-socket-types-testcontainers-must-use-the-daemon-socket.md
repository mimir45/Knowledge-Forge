---
title: "Docker Desktop Has Two Socket Types — Testcontainers Must Use the Daemon Socket"
slug: docker-desktop-has-two-socket-types-testcontainers-must-use-the-daemon-socket
type: howto
stack: [docker, testcontainers, java]
tags: [testing, macos]
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

# Docker Desktop Has Two Socket Types — Testcontainers Must Use the Daemon Socket

## What is it?

Docker Desktop on macOS exposes two different Unix sockets. The CLI socket (`~/.docker/desktop/docker.sock`) rejects old API versions (minimum v1.40); Testcontainers' bundled docker-java defaults to API version `v1.24`, causing HTTP 400 errors. The fix is to point Testcontainers at the daemon socket, not the CLI socket.

## How it works

| Socket | Path | Purpose | Accepts |
|---|---|---|---|
| Daemon socket | `/var/run/docker.sock` | Docker daemon | Any API version |
| Daemon socket (alt) | `~/.docker/run/docker.sock` | Docker daemon | Any API version |
| CLI socket | `~/.docker/desktop/docker.sock` | Docker Desktop CLI | v1.40+ only |

Testcontainers' shaded docker-java bundle hardcodes `v1.24` as the minimum negotiated version. System properties (`DOCKER_API_VERSION`, `docker.io.client.api.version`) do **not** override this in the shaded bundle — socket selection is the only effective lever.

## Key implementation steps

Edit `~/.testcontainers.properties`:

```properties
# WRONG — CLI socket, rejects v1.24
tc.host=unix://[REDACTED-PATH]

# CORRECT — daemon socket
tc.host=unix:///var/run/docker.sock
```

Verify the daemon socket works:
```bash
curl --unix-socket /var/run/docker.sock http://localhost/version
```
Should return a JSON response with `ServerVersion`.

## Common pitfalls

- `docker.io.client.api.version` and `DOCKER_API_VERSION` system/env properties are ineffective in the Testcontainers shaded bundle
- The error (`HTTP 400 Bad Request`) looks like a Docker server error, not a client configuration issue
- Docker Desktop on fresh installs may set `tc.host` to the CLI socket automatically

## When to use / not use

Any time you use Testcontainers on macOS with Docker Desktop. The daemon socket is always the correct target. The CLI socket is only for Docker Desktop's own UI communication.
