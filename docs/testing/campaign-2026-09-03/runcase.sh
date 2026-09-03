#!/bin/bash
# runcase.sh — the ONLY way a campaign agent records a result.
#
# usage:
#   bash /Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/runcase.sh \
#        <AGENT> <CASE_ID> <EXPECT_EXIT|-> <ASSERT_TEXT> -- <command> [args...]
#
# It runs the command, measures wall time, captures the real exit code and both
# streams, and appends one JSON object to runs/agent-<AGENT>.jsonl.
# Nothing about the result is typed by hand, so nothing can be fabricated.
#
# EXPECT_EXIT: an integer the command is contracted to return, or "-" when the
# case is exploratory and there is no contracted code. verdict is computed:
#   pass  -> exit == EXPECT_EXIT
#   fail  -> exit != EXPECT_EXIT   (this is a finding: declared contract broken)
#   n/a   -> EXPECT_EXIT was "-"

S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign

AGENT="$1"; shift
CASE="$1"; shift
EXPECT="$1"; shift
ASSERT="$1"; shift
if [ "$1" = "--" ]; then shift; fi

if [ -z "$AGENT" ] || [ -z "$CASE" ] || [ $# -eq 0 ]; then
  echo "runcase: usage: runcase.sh AGENT CASE EXPECT_EXIT ASSERT -- cmd..." >&2
  exit 64
fi

OUT=$(mktemp); ERR=$(mktemp)

T0=$(date +%s%N)
"$@" >"$OUT" 2>"$ERR"
CODE=$?
T1=$(date +%s%N)
MS=$(( (T1 - T0) / 1000000 ))

if [ "$EXPECT" = "-" ]; then
  VERDICT="n/a"
elif [ "$CODE" -eq "$EXPECT" ] 2>/dev/null; then
  VERDICT="pass"
else
  VERDICT="fail"
fi

CMD="$*"
mkdir -p "$S/runs"

jq -c -n \
  --arg agent   "$AGENT" \
  --arg case    "$CASE" \
  --arg cmd     "$CMD" \
  --argjson exit "$CODE" \
  --arg expect  "$EXPECT" \
  --argjson ms  "$MS" \
  --arg assert  "$ASSERT" \
  --arg verdict "$VERDICT" \
  --arg stdout  "$(head -c 800 "$OUT")" \
  --arg stderr  "$(head -c 800 "$ERR")" \
  --argjson stdout_bytes "$(wc -c <"$OUT" | tr -d ' ')" \
  --argjson stderr_bytes "$(wc -c <"$ERR" | tr -d ' ')" \
  '{agent:$agent, case:$case, cmd:$cmd, exit:$exit, expect:$expect, ms:$ms,
    contended:true, assert:$assert, verdict:$verdict,
    stdout_head:$stdout, stderr_head:$stderr,
    stdout_bytes:$stdout_bytes, stderr_bytes:$stderr_bytes}' \
  >> "$S/runs/agent-$AGENT.jsonl"

echo "[$CASE] exit=$CODE (expected $EXPECT) ${MS}ms verdict=$VERDICT"
echo "--- stdout (first 400) ---"; head -c 400 "$OUT"; echo
echo "--- stderr (first 400) ---"; head -c 400 "$ERR"; echo

rm -f "$OUT" "$ERR"
exit 0
