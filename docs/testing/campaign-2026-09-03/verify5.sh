#!/bin/bash
# H6: the gate's own comments say MarkUnverified = "publish flagged" and
# DropConfidence = "publish, but confidence drops". blocksWrite blocks both.
# Build a draft that fails ONLY the citation gate (remedy: mark_unverified) and see
# whether it is published or quarantined.
S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign
F=$S/forge
V=$S/vh6
export FORGE_CONFIG=$S/configs/base.md
rm -rf "$V"; cp -R "$S/vault-under-test" "$V"
mkdir -p "$S/snips"

D="$S/snips/h6-citation.md"
cat > "$D" <<'EOF'
---
title: "Spring Boot 3.4 Actuator Endpoint Defaults"
slug: spring-boot-3-4-actuator-endpoint-defaults
type: concept
stack: [spring-boot, java]
tags: [actuator, configuration]
depth: 2
confidence: medium
created: 2026-09-03
updated: 2026-09-03
verified: 2026-09-03
freshness_days: 365
sources: []
related: ["[[dependency-injection-in-spring-boot]]", "[[meterreadingsservice-spring-boot-4-x-project]]"]
supersedes: []
forge_version: 0.1.2
origin: ask
---

# Spring Boot 3.4 Actuator Endpoint Defaults

In Spring Boot 3.4 the actuator exposes only `health` over HTTP by default, and
`management.endpoints.web.exposure.include` must name any other endpoint explicitly.
This changed in 3.4 specifically: earlier versions also exposed `info`.

Enabling all endpoints costs roughly 40ms of extra startup time on a cold JVM.

## Mechanism

The exposure filter runs during `WebEndpointAutoConfiguration`, before the servlet
container binds, so an endpoint excluded there is never registered at all rather than
being registered and refused at request time.
EOF

echo "=== inbox before: $(ls "$V/_inbox" 2>/dev/null | wc -l | tr -d ' ') files"
$F gate --file "$D" --rel notes/concept/spring-boot-3-4-actuator-endpoint-defaults.md --vault "$V" > "$S/snips/h6.json" 2>"$S/snips/h6.err"
echo "=== forge gate exit=$?"
echo "--- gate outcomes ---"
jq -r '.outcomes[] | "  \(.gate)\t\(.verdict)\t\(.remedy)\t\(.detail // "")"' "$S/snips/h6.json" 2>/dev/null || head -c 800 "$S/snips/h6.json"
echo "--- quarantine flag ---"
jq -r '.quarantine' "$S/snips/h6.json" 2>/dev/null
echo "=== inbox after: $(ls "$V/_inbox" 2>/dev/null | wc -l | tr -d ' ') files"
ls "$V/_inbox" 2>/dev/null | tail -3
echo "=== was it published to notes/concept/ instead? ==="
ls "$V/notes/concept/spring-boot-3-4-actuator-endpoint-defaults.md" 2>&1 | tail -1
echo "=== if quarantined, what did the frontmatter become? ==="
Q="$V/_inbox/spring-boot-3-4-actuator-endpoint-defaults.md"
if [ -f "$Q" ]; then grep -E '^confidence:|^title:' "$Q"; echo "  --- Open questions section ---"; sed -n '/## Open questions/,/^$/p' "$Q" | head -12; fi
echo
echo "=== the comments this hypothesis is about ==="
sed -n '14,24p' /Users/mimir45/knowledge-forge/.claude/worktrees/kf-test-campaign/pkg/qualitygate/gate.go
echo "  --- and what blocksWrite actually does ---"
sed -n '110,120p' /Users/mimir45/knowledge-forge/.claude/worktrees/kf-test-campaign/pkg/qualitygate/gate.go
