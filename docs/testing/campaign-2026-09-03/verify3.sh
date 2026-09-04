#!/bin/bash
S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign
F=$S/forge
V=$S/vverify
REPO=/Users/mimir45/knowledge-forge/.claude/worktrees/kf-test-campaign
export FORGE_CONFIG=$S/configs/base.md
Q="how does dependency injection work in spring boot"

echo "########## X1 — forge intent against the REAL Claude Code payload ##########"
for f in user_input prompt user_prompt; do
  out=$(echo "{\"session_id\":\"s1\",\"hook_event_name\":\"UserPromptSubmit\",\"cwd\":\"/tmp\",\"$f\":\"$Q\"}" | $F intent --vault "$V" 2>&1)
  printf "  field=%-12s exit=%s  output=[%s]\n" "$f" "$?" "$(echo "$out" | head -c 70)"
done

echo
echo "########## X2 — what field names do the OTHER hook commands decode? ##########"
grep -n 'json:"' "$REPO"/cmd/forge/intent.go "$REPO"/cmd/forge/session_capture.go "$REPO"/cmd/forge/cache_source.go "$REPO"/cmd/forge/session_context.go 2>/dev/null

echo
echo "########## X3 — do those hooks produce any effect on a realistic payload? ##########"
echo '{"session_id":"s1","hook_event_name":"SessionEnd","transcript_path":"/nonexistent.jsonl","cwd":"/tmp"}' | $F session-capture --vault "$V" >/dev/null 2>&1
echo "  session-capture exit=$?  (_inbox count now: $(ls "$V/_inbox" 2>/dev/null | wc -l | tr -d ' '))"
echo '{"session_id":"s1","hook_event_name":"PostToolUse","tool_name":"WebFetch","tool_input":{"url":"https://example.com"},"tool_response":{"result":"hello world"}}' | $F cache-source --vault "$V" >/dev/null 2>&1
echo "  cache-source exit=$?  (.forge/cache/*.md count: $(ls "$V/.forge/cache"/*.md 2>/dev/null | wc -l | tr -d ' '))"

echo
echo "########## X4 — does pkg/qualitygate have ANY Benchmark at all? ##########"
grep -rn "func Benchmark" "$REPO"/pkg/qualitygate/ 2>/dev/null || echo "  none — CLAUDE.md's ~0.13ms figure has no benchmark behind it"
echo "  packages that DO have benchmarks:"
grep -rln "func Benchmark" "$REPO"/pkg/ "$REPO"/cmd/ 2>/dev/null | sed "s|$REPO/||"
echo "  packages listed in 'make bench':"
grep -n "bench:" -A3 "$REPO"/Makefile | head -6

echo
echo "########## X5 — drift over budget: is it the git work or the vault load? ##########"
for i in 1 2 3; do
  t0=$(date +%s%N); $F drift --repo kf="$REPO" --vault "$V" >/dev/null 2>&1; t1=$(date +%s%N)
  echo "  drift full           $(( (t1-t0)/1000000 ))ms"
done
HEAD=$(git -C "$REPO" rev-parse HEAD)
for i in 1 2 3; do
  t0=$(date +%s%N); $F drift --repo kf="$REPO" --vault "$V" --since-commit "$HEAD" >/dev/null 2>&1; t1=$(date +%s%N)
  echo "  drift --since-commit $(( (t1-t0)/1000000 ))ms   (the form the git hooks actually use)"
done

echo
echo "########## X6 — version reporting ##########"
echo "  plugin.json declares:  $(jq -r .version "$REPO"/.claude-plugin/plugin.json)"
echo "  build stamped:         $(grep -o 'main.version=[^ ]*' "$REPO"/Makefile || echo '(from git describe)')"
for a in --version -v version -V; do printf "  forge %-10s -> %s\n" "$a" "$($F $a 2>&1 | head -1)"; done
