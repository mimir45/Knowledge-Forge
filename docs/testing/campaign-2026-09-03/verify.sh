#!/bin/bash
# Orchestrator's own verification of the agents' most important claims.
# Nothing an agent reported goes into the report until it is reproduced here.
S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign
F=$S/forge
V=$S/vverify
export FORGE_CONFIG=$S/configs/base.md

rm -rf "$V"; cp -R "$S/vault-under-test" "$V"
mkdir -p "$S/snips"

echo "########## V1 — H1: does telemetry leak the raw question? ##########"
before=$(wc -l < "$V/.forge/log.jsonl")
$F recall --question "how do I rotate my AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI credential" --vault "$V" >/dev/null 2>&1
after=$(wc -l < "$V/.forge/log.jsonl")
echo "log lines: $before -> $after"
tail -1 "$V/.forge/log.jsonl" | jq -c '{topic, q_hash, q_hash_len:(.q_hash|length)}'

echo
echo "########## V2 — verify-code declared exit codes ##########"
printf 'echo hi\n' > "$S/snips/ok.sh"
$F verify-code --lang cobol --file "$S/snips/ok.sh" >/dev/null 2>&1; echo "  --lang cobol (invalid)      exit=$?  (help says 2 = usage error)"
$F verify-code --lang bash --file /nonexistent/x.sh >/dev/null 2>&1; echo "  --file nonexistent          exit=$?  (help says 2 = usage error)"
printf 'import org.springframework.stereotype.Service;\n@Service\nclass Foo {}\n' > "$S/snips/Foo.java"
$F verify-code --lang java --file "$S/snips/Foo.java" > "$S/snips/import.json" 2>&1; echo "  unresolvable spring import  exit=$?"
echo "  its JSON:"; cat "$S/snips/import.json" | head -c 600; echo

echo
echo "########## V3 — vault path handling (does it mkdir?) ##########"
$F recall --question test --vault /nonexistent/definitely/not/here >/dev/null 2>&1; echo "  nonexistent vault exit=$?"
$F recall --question test --vault /nonexistent/definitely/not/here 2>&1 | head -2
NEW=$S/snips/should-not-exist-vault
rm -rf "$NEW"
$F recall --question test --vault "$NEW" >/dev/null 2>&1; echo "  writable nonexistent vault exit=$?"
if [ -d "$NEW" ]; then echo "  !! forge CREATED the vault directory at $NEW"; ls -a "$NEW"; else echo "  directory not created"; fi

echo
echo "########## V4 — forge check without --repo ##########"
rm -rf "$S/vcheck"; cp -R "$S/vault-under-test" "$S/vcheck"
rm -f "$S/vcheck/reports/drift.md" "$S/vcheck/moc/codebase.md"
$F check --vault "$S/vcheck" --offline > "$S/snips/check.out" 2>&1
echo "  exit=$?"
echo "  --- what it said about drift/codebase ---"
grep -iE 'drift|codebase|skip' "$S/snips/check.out" | head -10
echo "  --- do the files exist now? ---"
ls -l "$S/vcheck/reports/drift.md" 2>&1 | tail -1
ls -l "$S/vcheck/moc/codebase.md" 2>&1 | tail -1
echo "  --- first 12 lines of drift.md ---"
head -12 "$S/vcheck/reports/drift.md" 2>/dev/null

echo
echo "########## V5 — H3: locked-stage guard disagreement ##########"
FORGE_CONFIG=$S/configs/h3-locked-fallback.md $F config >/dev/null 2>&1; echo "  forge config  with write.fallback=api   exit=$?"
FORGE_CONFIG=$S/configs/h3-locked-fallback.md $F engine select --stage write --vault "$V" >/dev/null 2>&1; echo "  engine select with write.fallback=api   exit=$?"
FORGE_CONFIG=$S/configs/h3-locked-fallback.md $F engine select --stage write --vault "$V" 2>&1 | head -2
echo "  --- direct violation for comparison ---"
FORGE_CONFIG=$S/configs/locked-direct.md $F config >/dev/null 2>&1; echo "  forge config with recall.engine=host    exit=$?"
FORGE_CONFIG=$S/configs/locked-direct.md $F config 2>&1 | head -2

echo
echo "########## V6 — H2: the intent gate threshold ##########"
for q in "what does saveAndFlush do differently from save" \
         "how does the continue.dev rag context provider protocol work" \
         "how do I configure CORS allowed origins in spring boot" \
         "how do kafka consumer groups rebalance"; do
  sc=$($F recall --question "$q" --vault "$V" 2>/dev/null | jq -r '.top_score')
  out=$(echo "{\"prompt\":\"$q\"}" | $F intent --vault "$V" 2>/dev/null)
  if [ -n "$out" ]; then emitted="EMITTED"; else emitted="silent"; fi
  printf "  score=%-6s intent=%-8s  %s\n" "$sc" "$emitted" "$q"
done

echo
echo "########## V7 — the real recall ceiling across every question the campaign asked ##########"
jq -s -r '[.[] | select(.kind!="note" and (.cmd|test("recall")) and .stdout_head != "")
           | (.stdout_head | capture("\"top_score\": (?<s>[0-9.]+)") | .s | tonumber)]
          | {n: length, max: max, min: min}' "$S"/runs/agent-*.jsonl
echo "  (verdicts actually observed:)"
jq -s -r '[.[] | select(.stdout_head != null) | (.stdout_head | capture("\"verdict\": \"(?<v>[^\"]+)\"") | .v)] | group_by(.) | map({v: .[0], n: length})' "$S"/runs/agent-*.jsonl

echo
echo "########## V8 — case sensitivity anomaly (agent 08: ALL CAPS scored higher) ##########"
for q in "bölüm i teori temelleri nedir" "BÖLÜM I TEORİ TEMELLERİ" "bolum i teori temelleri nedir" "Bölüm I Teori Temelleri"; do
  sc=$($F recall --question "$q" --vault "$V" 2>/dev/null | jq -r '.top_score')
  slug=$($F recall --question "$q" --vault "$V" 2>/dev/null | jq -r '.candidates[0].slug // "none"')
  printf "  %-8s  %-28s  %s\n" "$sc" "$slug" "$q"
done
