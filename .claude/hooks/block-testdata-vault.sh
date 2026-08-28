#!/usr/bin/env bash
# PreToolUse hook (Edit|Write): refuses any write into testdata/vault/.
#
# testdata/vault/ is Knowledge Forge's fixture vault -- 13 notes carrying 12
# deliberate defects (F1-F12: dangling wikilinks, mixed frontmatter shapes,
# a near-duplicate pair, etc). CLAUDE.md states three separate times:
# "The defects are the test surface. Do not fix them." This is exactly the
# kind of file an agent "helpfully" repairs without re-reading that rule.
set -euo pipefail

input=$(cat)
file_path=$(jq -r '.tool_input.file_path // empty' <<<"$input")

if [[ "$file_path" == *"testdata/vault/"* ]]; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "testdata/vault/ is the fixture vault -- its 12 defects (F1-F12) ARE the test surface. CLAUDE.md: \"The defects are the test surface. Do not fix them.\" If this edit is genuinely intended (e.g. adding a new labeled fixture), ask the user to make it directly, or confirm explicitly before retrying."
    }
  }'
fi
