#!/usr/bin/env bash
# evals/run.sh — determinism checks against forge recall and forge check.
#
# This is not a quality eval (there's no model output to score — the T0 core makes
# zero model calls). What it checks is the property this project's docs repeatedly
# claim and measure by hand each phase: given the same vault, forge recall and forge
# check produce byte-identical output on repeated runs. See CLAUDE.md's Status
# section for the same check done manually each phase (Phase 2b's six-run md5 check,
# Phase 5b's byte-identical smoke test) — this script is that check, made repeatable.
#
# Uses a small fixture vault under evals/fixtures/vault/, never the real vault and
# never testdata/vault/ (that fixture is deliberately broken — see testdata/README.md).
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture="$root/evals/fixtures/vault"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

bin="$work/forge"
CGO_ENABLED=0 go build -o "$bin" "$root/cmd/forge"

vault="$work/vault"
cp -R "$fixture" "$vault"
(cd "$vault" && git init -q && git add -A && git -c user.email=eval@local -c user.name=eval commit -q -m "fixture")

fail=0

echo "== forge recall: determinism =="
# run_id (B-035) is a fresh random correlation key minted per call by design — it must
# differ every run, so it's excluded here the same way a timestamp field would be from
# any other snapshot comparison. Everything else in the envelope must still match.
q="how does kafka consumer group rebalancing work"
r1="$("$bin" recall --question "$q" --vault "$vault" | grep -v '"run_id"')"
r2="$("$bin" recall --question "$q" --vault "$vault" | grep -v '"run_id"')"
if [ "$r1" != "$r2" ]; then
  echo "FAIL: forge recall output differs across runs"
  diff <(echo "$r1") <(echo "$r2") || true
  fail=1
else
  echo "PASS: identical across two runs"
fi

echo "== forge check: determinism =="
v1="$work/vault1"; v2="$work/vault2"
cp -R "$vault" "$v1"; cp -R "$vault" "$v2"
"$bin" check --vault "$v1" --offline >/dev/null
"$bin" check --vault "$v2" --offline >/dev/null
if diff -r "$v1/reports" "$v2/reports" >/dev/null 2>&1; then
  echo "PASS: reports/ byte-identical across two runs"
else
  echo "FAIL: forge check reports differ across runs"
  diff -r "$v1/reports" "$v2/reports" || true
  fail=1
fi

exit $fail
