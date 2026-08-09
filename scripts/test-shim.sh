#!/bin/sh
# Fixture test for bin/forge. The shim is the only thing standing between the vault's
# silent post-commit hook and an unverified binary, so "it works on my machine" is not
# enough — a shim that fails open would be indistinguishable from one that works.
set -eu

cd "$(dirname "$0")/.."
shim="$PWD/bin/forge"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

fake="$tmp/forge"
printf '#!/bin/sh\necho RAN "$@"\n' > "$fake"
chmod +x "$fake"
good=$(shasum -a 256 "$fake" 2>/dev/null | awk '{print $1}' \
	|| sha256sum "$fake" | awk '{print $1}')

fail=0
check() { # name, expected-exit, actual-exit
	if [ "$2" = "$3" ]; then echo "ok   $1"
	else echo "FAIL $1: exit $3, want $2"; fail=1; fi
}

set +e
FORGE_BIN="$fake" FORGE_BIN_SHA256="$good" "$shim" x >"$tmp/out" 2>&1
check "runs a binary whose hash matches" 0 $?
grep -q 'RAN x' "$tmp/out" || { echo "FAIL arguments not forwarded"; fail=1; }

FORGE_BIN="$fake" FORGE_BIN_SHA256=deadbeef "$shim" >/dev/null 2>&1
check "refuses a hash mismatch" 127 $?

printf '%s\n' "$good" > "$fake.sha256"
FORGE_BIN="$fake" "$shim" >/dev/null 2>&1
check "accepts a sidecar pin" 0 $?

printf '# tampered\n' >> "$fake"
FORGE_BIN="$fake" "$shim" >/dev/null 2>&1
check "refuses a binary edited after pinning" 127 $?

rm -f "$fake.sha256"
FORGE_BIN="$fake" env -u FORGE_BIN_SHA256 "$shim" >/dev/null 2>&1
check "refuses when no pin exists at all" 127 $?

FORGE_BIN="$tmp/absent" "$shim" >/dev/null 2>&1
check "refuses a missing binary" 127 $?
set -e

[ "$fail" = 0 ] || { echo "shim tests failed"; exit 1; }
echo "all shim tests passed"
