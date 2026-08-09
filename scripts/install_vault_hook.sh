#!/bin/sh
# Install the D3 capture hook into a vault repository.
#
#   scripts/install_vault_hook.sh <vault-dir> [forge-binary]
#
# Idempotent, and refuses rather than clobbers: an existing post-commit hook that is not
# ours is left alone, because it is the user's and this script has no way to merge it.
# Nothing here touches note content — it writes .git/hooks/post-commit and .forge/forge-bin.

set -eu

vault=${1:?usage: install_vault_hook.sh <vault-dir> [forge-binary]}
forge=${2:-$(command -v forge)}
here=$(cd "$(dirname "$0")/.." && pwd)
marker="Knowledge Forge — D3 human-correction capture"

[ -d "$vault/.git" ] || { echo "not a git repository: $vault" >&2; exit 1; }
[ -x "$forge" ] || { echo "not an executable forge binary: $forge" >&2; exit 1; }

hook="$vault/.git/hooks/post-commit"
if [ -e "$hook" ] && ! grep -q "$marker" "$hook"; then
    echo "refusing to overwrite an existing post-commit hook: $hook" >&2
    exit 1
fi

mkdir -p "$vault/.forge"
printf '%s\n' "$forge" > "$vault/.forge/forge-bin"
cp "$here/hooks/vault-post-commit" "$hook"
chmod +x "$hook"

echo "installed $hook"
echo "forge binary: $forge"
echo "dataset:      $vault/.forge/datasets/d3.jsonl (gitignored, local only)"
