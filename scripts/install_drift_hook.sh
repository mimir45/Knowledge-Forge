#!/bin/sh
# Install the three git-anchored drift hooks into a CODE repository, paired with a vault.
#
#   scripts/install_drift_hook.sh <code-repo> <vault-dir> [forge-binary]
#
# Idempotent, and refuses rather than clobbers: an existing post-commit/post-merge/
# post-checkout hook that is not ours is left alone — this script has no way to merge it.
# Writes <code-repo>/.forge/vault-path (the only pairing a code repo has no other way to
# know) and each of the three git hooks; does not touch note content or the vault itself.

set -eu

repo=${1:?usage: install_drift_hook.sh <code-repo> <vault-dir> [forge-binary]}
vault=${2:?usage: install_drift_hook.sh <code-repo> <vault-dir> [forge-binary]}
forge=${3:-$(command -v forge)}
here=$(cd "$(dirname "$0")/.." && pwd)
marker="Knowledge Forge — git-anchored drift"

[ -d "$repo/.git" ] || { echo "not a git repository: $repo" >&2; exit 1; }
[ -d "$vault" ] || { echo "not a directory: $vault" >&2; exit 1; }
[ -x "$forge" ] || { echo "not an executable forge binary: $forge" >&2; exit 1; }

vault=$(cd "$vault" && pwd)

check_one() {
    hook="$repo/.git/hooks/$1"
    if [ -e "$hook" ] && ! grep -q "$marker" "$hook"; then
        echo "refusing to overwrite an existing $1 hook: $hook" >&2
        exit 1
    fi
}

install_one() {
    hook="$repo/.git/hooks/$1"
    cp "$here/hooks/$2" "$hook"
    chmod +x "$hook"
    echo "installed $hook"
}

# Check all three before touching any — a conflict on the second or third hook must not
# leave the first one installed with the other two untouched.
check_one post-commit
check_one post-merge
check_one post-checkout

install_one post-commit   code-post-commit
install_one post-merge    code-post-merge
install_one post-checkout code-post-checkout

mkdir -p "$repo/.forge"
printf '%s\n' "$vault" > "$repo/.forge/vault-path"

if [ -f "$repo/.gitignore" ] && ! grep -qx '\.forge/' "$repo/.gitignore"; then
    echo "note: add '.forge/' to $repo/.gitignore — $repo/.forge/vault-path is local pairing state, not something to commit" >&2
elif [ ! -f "$repo/.gitignore" ]; then
    echo "note: $repo has no .gitignore — add one excluding '.forge/' so vault-path stays local, not committed" >&2
fi

echo "vault:  $vault"
echo "forge:  resolved at hook-fire-time via \$FORGE_BIN -> ~/.forge/bin/forge -> \$PATH"
