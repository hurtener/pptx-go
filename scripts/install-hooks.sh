#!/usr/bin/env bash
#
# install-hooks.sh — install the pptx-go git hooks (one-time, per clone).
#
# Symlinks scripts/hooks/* into .git/hooks/ so the tracked hook scripts stay
# the single source of truth.

set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

hooks_src="scripts/hooks"
hooks_dst="$(git rev-parse --git-path hooks)"

mkdir -p "$hooks_dst"

for hook in "$hooks_src"/*; do
	[ -f "$hook" ] || continue
	name="$(basename "$hook")"
	dst="$hooks_dst/$name"
	ln -sf "../../$hooks_src/$name" "$dst"
	chmod +x "$hook"
	echo "installed: $name -> $dst"
done

echo "Done. Hooks installed from $hooks_src/."
