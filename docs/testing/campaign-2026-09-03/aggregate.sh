#!/bin/bash
# Rebuilds every number in the report from the raw JSONL. Agent prose is never a source.
S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign
R=$S/runs

echo "=== per-agent case counts ==="
printf "%-8s %6s %6s %6s %6s %6s\n" agent cases pass fail na notes
for f in "$R"/agent-*.jsonl; do
  [ -e "$f" ] || continue
  a=$(basename "$f" .jsonl | sed 's/agent-//')
  cases=$(jq -s '[.[] | select(.kind != "note")] | length' "$f")
  pass=$(jq -s '[.[] | select(.verdict=="pass")] | length' "$f")
  fail=$(jq -s '[.[] | select(.verdict=="fail")] | length' "$f")
  na=$(jq -s '[.[] | select(.verdict=="n/a")] | length' "$f")
  notes=$(jq -s '[.[] | select(.kind=="note")] | length' "$f")
  printf "%-8s %6s %6s %6s %6s %6s\n" "$a" "$cases" "$pass" "$fail" "$na" "$notes"
done

echo
echo "=== TOTALS ==="
jq -s '{cases: ([.[] | select(.kind != "note")] | length),
        pass:  ([.[] | select(.verdict=="pass")] | length),
        fail:  ([.[] | select(.verdict=="fail")] | length),
        na:    ([.[] | select(.verdict=="n/a")] | length),
        notes: ([.[] | select(.kind=="note")] | length)}' "$R"/agent-*.jsonl

echo
echo "=== CONTRACT FAILURES (verdict=fail: declared exit code not met) ==="
jq -s -r '.[] | select(.verdict=="fail") |
  "[\(.agent)/\(.case)] exit=\(.exit) expected=\(.expect)  \(.assert)\n    cmd: \(.cmd)\n    err: \(.stderr_head | split("\n")[0])"' \
  "$R"/agent-*.jsonl

echo
echo "=== NOTES: severity=bug ==="
jq -s -r '.[] | select(.kind=="note" and .severity=="bug") |
  "[\(.agent)/\(.case)] \(.text)"' "$R"/agent-*.jsonl

echo
echo "=== NOTES: severity=suspect ==="
jq -s -r '.[] | select(.kind=="note" and .severity=="suspect") |
  "[\(.agent)/\(.case)] \(.text)"' "$R"/agent-*.jsonl

echo
echo "=== crashes / panics in captured stderr ==="
jq -s -r '.[] | select(.stderr_head != null and (.stderr_head | test("panic:|goroutine \\d+|runtime error"))) |
  "[\(.agent)/\(.case)] \(.stderr_head | split("\n")[0])"' "$R"/agent-*.jsonl
echo "(empty above = no panic observed)"
