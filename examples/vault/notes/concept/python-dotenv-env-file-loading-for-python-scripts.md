---
title: "python-dotenv — .env File Loading for Python Scripts"
slug: python-dotenv-env-file-loading-for-python-scripts
type: concept
stack: [python]
tags: [config, dotenv, cli]
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

# python-dotenv — .env File Loading for Python Scripts

`python-dotenv` loads key-value pairs from a `.env` file into environment variables at script
startup. Useful for keeping credentials and config out of source-tracked CLI flags.

## Priority Order

`.env` file < shell environment exports < explicit CLI flags

CLI flags always win, so scripts can be overridden at runtime without editing `.env`.

## Pattern

```python
from dotenv import load_dotenv
load_dotenv()  # reads .env from CWD
token = os.getenv("ADMIN_TOKEN")
```

Provide a `.env.example` template in the repo; add `.env` to `.gitignore`.

## Sources

- [[sources/daily/2026-04-20-python-mock-server]]

## Related

- [[notes/concept/mock-server-py-python-subscription-api-mock]]
