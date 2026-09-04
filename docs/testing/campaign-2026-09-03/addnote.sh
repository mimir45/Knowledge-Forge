#!/bin/bash
# addnote.sh — record an OBSERVATION that a bare exit code cannot express.
#
# usage:
#   bash /Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign/addnote.sh \
#        <AGENT> <CASE_ID> <severity: bug|suspect|info> <TEXT>
#
# Use this for judgements about CONTENT — "the topic field contained the nonce",
# "the JSON had candidates: null", "reports/drift.md was written empty".
# Always cite the concrete evidence you saw in TEXT; a note with no observed
# string in it is worthless.

S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign
AGENT="$1"; CASE="$2"; SEV="$3"; shift 3; TEXT="$*"

if [ -z "$AGENT" ] || [ -z "$CASE" ] || [ -z "$SEV" ] || [ -z "$TEXT" ]; then
  echo "addnote: usage: addnote.sh AGENT CASE bug|suspect|info TEXT" >&2
  exit 64
fi

mkdir -p "$S/runs"
jq -c -n --arg agent "$AGENT" --arg case "$CASE" --arg sev "$SEV" --arg text "$TEXT" \
  '{agent:$agent, case:$case, kind:"note", severity:$sev, text:$text}' \
  >> "$S/runs/agent-$AGENT.jsonl"
echo "noted [$CASE/$SEV]: $TEXT"
