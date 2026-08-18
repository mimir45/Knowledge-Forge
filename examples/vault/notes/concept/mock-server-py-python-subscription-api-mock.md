---
title: "mock_server.py — Python Subscription API Mock"
slug: mock-server-py-python-subscription-api-mock
type: concept
stack: [python, stripe]
tags: [mock, testing, api]
depth: 3
confidence: low
created: 2026-04-20
updated: 2026-04-20
verified: 2026-04-20
freshness_days: 365
sources:
  - url: sources/daily/2026-04-20-python-mock-server.md
    accessed: 2026-04-20
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# mock_server.py — Python Subscription API Mock

Python mock server for `PUT /api/v1/subscription` endpoint. Runs alongside an XML-driven test
runner for 263-step subscription flow validation.

## Auth Logic

| Condition | Response |
|-----------|----------|
| No auth header | 401 |
| Non-admin token | 403 |
| Invalid body | 400 |
| Blocked card ID | 400 |
| Unknown real-looking ID | 404 |
| Valid `pm_card_*` | 204 |

## Configuration

Uses `python-dotenv` for `.env` file support. Priority: `.env` < environment exports < CLI flags.
Non-admin token must be identical in both mock server and test runner.

## Known Issues

- XML test exports can contain malformed JSON (missing commas); test runner normalizes before parsing
- `--non-admin-token` CLI flag was mistyped as `--on-admin-token` — causes unexpected 401s
- Error response bodies return bare `{"status": 400}` — realistic error messages pending

## Sources

- [[sources/daily/2026-04-20-python-mock-server]]

## Related

- [[notes/concept/python-dotenv-env-file-loading-for-python-scripts]]
