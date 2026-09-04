#!/bin/bash
S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign
F=$S/forge
REPO=/Users/mimir45/knowledge-forge/.claude/worktrees/kf-test-campaign

echo "=== 1) which commands answer --help, and which print nothing ==="
for c in recall index reindex validate gate drift check config init slug stats capture \
         verify-code logback scrub export-dataset dataset-stats \
         session-context intent session-capture cache-source; do
  n=$($F "$c" --help 2>&1 | wc -c | tr -d ' ')
  if [ "$n" -lt 60 ]; then echo "  NO HELP: $c  (--help produced $n bytes)"; fi
done

echo
echo "=== 2) forge engine --help ==="
$F engine --help > "$S/eh.txt" 2>&1; echo "  exit=$?"
head -2 "$S/eh.txt"

echo
echo "=== 3) README state ==="
echo "  in this worktree: $([ -f "$REPO/README.md" ] && echo present || echo missing)"
echo "  in the main checkout working tree:"
git -C /Users/mimir45/knowledge-forge status --short -- README.md 2>&1 | head -2

echo
echo "=== 4) USAGE.md claims that the campaign measured ==="
grep -nE 'exit|0\.7|budget|ms\b' "$REPO/docs/USAGE.md" | head -20

echo
echo "=== 5) does any doc still name the deleted design docs? ==="
grep -rn "DESIGN §\|ADDENDUM §\|ADDENDUM section\|ROADMAP.md\|KNOWLEDGE-FORGE-" \
  "$REPO/docs" "$REPO/references" "$REPO/skills" "$REPO/agents" "$REPO/config" 2>/dev/null \
  | sed "s|$REPO/||" | wc -l
echo "  (count of surviving references to documents that no longer exist)"
grep -rln "DESIGN §\|ADDENDUM §\|ADDENDUM section\|KNOWLEDGE-FORGE-" \
  "$REPO/docs" "$REPO/references" "$REPO/skills" "$REPO/agents" "$REPO/config" "$REPO/cmd" 2>/dev/null \
  | sed "s|$REPO/||"

echo
echo "=== 6) Makefile bench comment vs its package list ==="
grep -n -A4 "^## bench" "$REPO/Makefile"
