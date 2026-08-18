---
title: 'Spring CLI: Create project as folder instead of zip'
slug: spring-cli-create-project-as-folder-instead-of-zip
type: howto
stack: [spring-boot]
tags: [cli, tools]
depth: 3
confidence: low
created: 2026-04-30
updated: 2026-04-30
verified: 2026-04-30
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Spring CLI: Create project as folder instead of zip

By default, `spring init` downloads a `.zip` file to the current directory.

## Direct extraction with `--extract`

Use the `--extract` (or `-x`) flag with a directory path as the output target:

```bash
spring init -g=com.samir -a=test -j=21 --build=maven \
  -d=web,data-jpa,lombok,validation test/ --extract
```

The trailing slash on `test/` signals a directory target; `--extract` tells the CLI to unzip in place rather than save the archive.

## Alternative: unzip after the fact

```bash
spring init -g=com.samir -a=test ... test.zip
unzip test.zip -d myproject
```

## Summary

| Goal | Command |
|------|---------|
| Save as zip | `spring init ... myproject.zip` |
| Extract directly | `spring init ... myproject/ --extract` |
| Unzip manually | `unzip myproject.zip -d myproject` |
