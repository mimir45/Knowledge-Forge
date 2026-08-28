#!/usr/bin/env bash
# PostToolUse hook (Edit|Write): after a *.go file is written, runs
# `gofmt -l` and `go vet` on it and feeds any issues back to Claude as
# additionalContext. Non-blocking by design -- this is the cheap half of
# `make lint`, meant to catch what CI would catch anyway before the
# push round-trip, not to gate the edit.
set -uo pipefail

input=$(cat)
file_path=$(jq -r '.tool_input.file_path // .tool_response.filePath // empty' <<<"$input")

case "$file_path" in
  *.go) ;;
  *) exit 0 ;;
esac
[[ -f "$file_path" ]] || exit 0

gofmt_out=$(gofmt -l "$file_path" 2>&1)
vet_out=$(cd "$(dirname "$file_path")" && go vet . 2>&1)

msg=""
if [[ -n "$gofmt_out" ]]; then
  msg+="gofmt: ${file_path} is not formatted -- run: gofmt -w ${file_path}"$'\n'
fi
if [[ -n "$vet_out" ]]; then
  msg+="go vet (package $(dirname "$file_path")):"$'\n'"${vet_out}"$'\n'
fi

if [[ -n "$msg" ]]; then
  jq -n --arg msg "$msg" '{hookSpecificOutput: {hookEventName: "PostToolUse", additionalContext: $msg}}'
fi
exit 0
