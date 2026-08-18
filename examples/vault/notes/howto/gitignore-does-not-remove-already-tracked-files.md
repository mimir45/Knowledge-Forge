---
title: ".gitignore Does Not Remove Already-Tracked Files"
slug: gitignore-does-not-remove-already-tracked-files
type: howto
stack: [git]
tags: [workflow]
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

# .gitignore Does Not Remove Already-Tracked Files

## What is it?

Adding a path to `.gitignore` has no effect if that path is already tracked by git. Git's index takes precedence — tracked files remain tracked regardless of `.gitignore`. You must explicitly remove the file from the index first.

## How it works

```
git track state ──→ .gitignore checked? ──→ file ignored?
     YES                  NO                  NO (still tracked)
     NO                   YES                 YES (ignored)
```

## Key implementation steps

```bash
# Remove from index without deleting from disk
git rm --cached <file>

# For a directory
git rm --cached -r <directory>/

# Then add to .gitignore
echo "directory/" >> .gitignore

# Commit both changes
git add .gitignore
git commit -m "stop tracking directory/"
```

## Common pitfalls

- Running `git rm --cached` without `-r` on a directory will fail — add the `-r` flag
- The file stays on your local disk — `--cached` only removes the git index entry
- Collaborators who already have the file tracked will need to run `git rm --cached` themselves after pulling the `.gitignore` change

## When to use / not use

Run this whenever you add a pattern to `.gitignore` for something already committed (build artifacts, IDE config, generated files, secrets accidentally committed). Always verify with `git status` after — the file should show as "untracked" not "modified."
