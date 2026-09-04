#!/bin/bash
# Phase B — the SERIAL latency pass. Nothing else may be running.
#
# Every number here is asserted against the budget the project documents for it.
# A budget you assert against is a bug detector; a bare number is a log entry.
#
# cold = the run right after the derived SQLite cache is deleted (and, for the
#        first command, the binary itself is not yet in the page cache)
# warm = the median of 10 subsequent runs, cache in place

S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign
F=$S/forge
V=$S/vtime
REPO=/Users/mimir45/knowledge-forge/.claude/worktrees/kf-test-campaign
export FORGE_CONFIG=$S/configs/base.md

rm -rf "$V"; cp -R "$S/vault-under-test" "$V"
mkdir -p "$S/snips"
printf 'echo hello\nif [ 1 -eq 1 ]; then echo yes; fi\n' > "$S/snips/t.sh"
printf 'class T { public static void main(String[] a){ System.out.println(1); } }\n' > "$S/snips/T.java"
echo '{"user_prompt":"how does dependency injection work in spring boot"}' > "$S/snips/intent.json"
echo '{}' > "$S/snips/empty.json"

ms() { echo $(( ($2 - $1) / 1000000 )); }

# run_n <label> <budget_ms> <stdin_file|-> <cmd...>
run_n() {
  local label="$1" budget="$2" stdin="$3"; shift 3
  local times=() t0 t1
  for i in $(seq 1 10); do
    t0=$(date +%s%N)
    if [ "$stdin" = "-" ]; then "$@" >/dev/null 2>&1; else "$@" < "$stdin" >/dev/null 2>&1; fi
    t1=$(date +%s%N)
    times+=( $(ms "$t0" "$t1") )
  done
  local sorted median p95
  sorted=$(printf '%s\n' "${times[@]}" | sort -n)
  median=$(echo "$sorted" | sed -n '5p')
  p95=$(echo "$sorted" | sed -n '10p')
  local flag="ok"
  if [ "$budget" != "-" ] && [ "$median" -gt "$budget" ]; then flag="OVER BUDGET"; fi
  printf "| %-26s | %8s | %8s | %8s | %-11s |\n" "$label" "${median}ms" "${p95}ms" "${budget}ms" "$flag"
}

cold_one() {
  local label="$1" stdin="$2"; shift 2
  local t0 t1
  rm -f "$V/.forge/cache/index.db"
  t0=$(date +%s%N)
  if [ "$stdin" = "-" ]; then "$@" >/dev/null 2>&1; else "$@" < "$stdin" >/dev/null 2>&1; fi
  t1=$(date +%s%N)
  printf "| %-26s | %8s |\n" "$label" "$(ms "$t0" "$t1")ms"
}

echo "### COLD (derived SQLite cache deleted immediately before each run)"
echo "| command                    |     cold |"
echo "|----------------------------|----------|"
cold_one "recall"          - $F recall --question "how does dependency injection work in spring boot" --vault "$V"
cold_one "index"           - $F index --vault "$V"
cold_one "check --offline" - $F check --vault "$V" --offline
cold_one "drift"           - $F drift --repo kf="$REPO" --vault "$V"
cold_one "session-context" "$S/snips/empty.json"  $F session-context --vault "$V"
cold_one "intent"          "$S/snips/intent.json" $F intent --vault "$V"
cold_one "verify-code bash" - $F verify-code --lang bash --file "$S/snips/t.sh"
cold_one "verify-code java" - $F verify-code --lang java --file "$S/snips/T.java"

echo
echo "### WARM (median and p95 of 10 serial runs, cache in place)"
echo "| command                    |   median |      p95 |   budget | verdict     |"
echo "|----------------------------|----------|----------|----------|-------------|"
$F index --vault "$V" >/dev/null 2>&1   # ensure cache is warm
run_n "recall"           -   -                     $F recall --question "how does dependency injection work in spring boot" --vault "$V"
run_n "index"            200 -                     $F index --vault "$V"
run_n "check --offline"  10000 -                   $F check --vault "$V" --offline
run_n "drift"            100 -                     $F drift --repo kf="$REPO" --vault "$V"
run_n "session-context"  200 "$S/snips/empty.json" $F session-context --vault "$V"
run_n "intent"           50  "$S/snips/intent.json" $F intent --vault "$V"
run_n "verify-code bash" -   -                     $F verify-code --lang bash --file "$S/snips/t.sh"
run_n "verify-code java" -   -                     $F verify-code --lang java --file "$S/snips/T.java"

echo
echo "### Hook timeouts declared in hooks/hooks.json (a hard ceiling, not a budget)"
echo "  SessionStart/session-context: 5s   UserPromptSubmit/intent: 2s   SessionEnd: 10s   PostToolUse: 10s"

echo
echo "### pkg/qualitygate benchmark (CLAUDE.md claims ~0.13ms for Run; the package is not in 'make bench')"
cd "$REPO" && CGO_ENABLED=0 go test ./pkg/qualitygate -run '^$' -bench . -benchmem 2>&1 | tail -12
