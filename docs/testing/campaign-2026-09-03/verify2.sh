#!/bin/bash
S=/Users/mimir45/.claude/jobs/c679826c/tmp/kf-campaign
F=$S/forge
V=$S/vverify
export FORGE_CONFIG=$S/configs/base.md

echo "########## W1 — forge intent: does it EVER emit? ##########"
Q="how does dependency injection work in spring boot"
echo "recall score for the control question:"
$F recall --question "$Q" --vault "$V" 2>/dev/null | jq -r '.top_score'
echo
for payload in \
  "{\"prompt\":\"$Q\"}" \
  "{\"user_prompt\":\"$Q\"}" \
  "{\"message\":\"$Q\"}" \
  "{\"prompt\":\"$Q\",\"hook_event_name\":\"UserPromptSubmit\",\"session_id\":\"abc\",\"cwd\":\"/tmp\"}" ; do
  out=$(echo "$payload" | $F intent --vault "$V" 2>&1)
  printf "  %-92s -> [%s]\n" "$(echo "$payload" | head -c 90)" "$(echo "$out" | head -c 90)"
done
echo "  raw text on stdin:"
printf "  %-92s -> [%s]\n" "$Q" "$(echo "$Q" | $F intent --vault "$V" 2>&1 | head -c 90)"

echo
echo "########## W2 — verify-code: does an unresolvable import MASK a real syntax error? ##########"
mkdir -p "$S/snips"
# a) genuine syntax error, no imports  -> should FAIL (exit 1)
printf 'class A { void f() { int x = 1 } }\n' > "$S/snips/SyntaxOnly.java"
$F verify-code --lang java --file "$S/snips/SyntaxOnly.java" > "$S/snips/a.json" 2>&1; echo "  a) syntax error only          exit=$?"
jq -c '{verdict, detail}' "$S/snips/a.json" 2>/dev/null || head -c 200 "$S/snips/a.json"
# b) genuine syntax error AND an unresolvable import -> ?
printf 'import org.springframework.stereotype.Service;\nclass B { void f() { int x = 1 } }\n' > "$S/snips/Both.java"
$F verify-code --lang java --file "$S/snips/Both.java" > "$S/snips/b.json" 2>&1; echo "  b) syntax error + bad import  exit=$?"
jq -c '{verdict, detail}' "$S/snips/b.json" 2>/dev/null || head -c 200 "$S/snips/b.json"
echo "  -> if (a) fails but (b) is 'skipped', an unresolvable import MASKS real syntax errors."

echo
echo "########## W3 — forge check: the skipped-file accounting ##########"
rm -rf "$S/vcheck2"; cp -R "$S/vault-under-test" "$S/vcheck2"
rm -f "$S/vcheck2/reports/drift.md" "$S/vcheck2/moc/codebase.md"
echo "  before: drift.md exists? $([ -f "$S/vcheck2/reports/drift.md" ] && echo yes || echo no)"
$F check --vault "$S/vcheck2" --offline 2>&1 | tail -20
echo "  after:  drift.md exists? $([ -f "$S/vcheck2/reports/drift.md" ] && echo yes || echo no)"
echo "  after:  moc/codebase.md exists? $([ -f "$S/vcheck2/moc/codebase.md" ] && echo yes || echo no)"

echo
echo "########## W4 — was agent 05's finding an artifact of pre-existing files? ##########"
echo "  does the pristine baseline copy already ship reports/drift.md?"
ls -l "$S/vault-under-test/reports/drift.md" 2>&1 | tail -1
ls -l "$S/vault-under-test/moc/codebase.md" 2>&1 | tail -1

echo
echo "########## W5 — the casing claim: is it case, or the extra word? ##########"
for q in "Bölüm I Teori Temelleri" "bölüm i teori temelleri" "BÖLÜM I TEORİ TEMELLERİ" \
         "bölüm i teori temelleri nedir" "Bölüm I Teori Temelleri nedir"; do
  sc=$($F recall --question "$q" --vault "$V" 2>/dev/null | jq -r '.top_score')
  printf "  %-8s  %s\n" "$sc" "$q"
done
