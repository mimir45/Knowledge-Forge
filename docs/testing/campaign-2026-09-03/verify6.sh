#!/bin/bash
# Isolate the DropConfidence remedy: a draft that passes schema and citation but whose
# fenced code block does not compile. The code gate's remedy is DropConfidence, whose
# comment reads "publish, but confidence drops". Does it publish?
S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign
F=$S/forge
V=$S/vh6b
export FORGE_CONFIG=$S/configs/base.md
rm -rf "$V"; cp -R "$S/vault-under-test" "$V"

D="$S/snips/h6-code.md"
cat > "$D" <<'MDEOF'
---
title: "Restarting the Local Model Server Safely"
slug: restarting-the-local-model-server-safely
type: howto
stack: [llama-cpp]
tags: [shell, operations]
depth: 2
confidence: medium
created: 2026-09-03
updated: 2026-09-03
verified: 2026-09-03
freshness_days: 180
sources:
  - url: "https://github.com/ggerganov/llama.cpp"
    accessed: 2026-09-03
    kind: official
related: ["[[local-ai-stack-architecture-overview]]", "[[llama-server-sh-llama-cpp-startup-script]]"]
supersedes: []
forge_version: 0.1.2
origin: ask
---

# Restarting the Local Model Server Safely

See [[local-ai-stack-architecture-overview]] for where this sits. The restart has to
drain in-flight requests before the port is released, otherwise the next start races
the old socket.

```bash
if [ -f "$PIDFILE" ]; then
  kill "$(cat "$PIDFILE")"
  echo "stopped"
```

## Mechanism

The guard above is deliberately broken: the `if` block is never closed, so `bash -n`
reports a syntax error. That is what this draft is here to exercise.
MDEOF

echo "=== forge gate on a draft whose ONLY failing gate should be 'code' ==="
$F gate --file "$D" --rel notes/howto/restarting-the-local-model-server-safely.md --vault "$V" > "$S/snips/h6b.json" 2>&1
echo "exit=$?"
jq -r '.outcomes[] | "  \(.gate)\t\(.verdict)\t\(.remedy)\t\(.detail // "")"' "$S/snips/h6b.json" 2>/dev/null || head -c 700 "$S/snips/h6b.json"
echo "  quarantine: $(jq -r '.quarantine' "$S/snips/h6b.json" 2>/dev/null)"
echo
echo "published to notes/howto/ ?  $([ -f "$V/notes/howto/restarting-the-local-model-server-safely.md" ] && echo YES || echo NO)"
echo "quarantined to _inbox/  ?  $([ -f "$V/_inbox/restarting-the-local-model-server-safely.md" ] && echo YES || echo NO)"
if [ -f "$V/_inbox/restarting-the-local-model-server-safely.md" ]; then
  grep -E '^confidence:' "$V/_inbox/restarting-the-local-model-server-safely.md"
fi
