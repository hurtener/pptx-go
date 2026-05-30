#!/usr/bin/env bash
#
# preflight.sh — the pre-commit / CI gate (CLAUDE.md §4.1).
#
# Builds the library CGo-free, runs every per-phase smoke script (each SKIPs
# gracefully where its surface is not built yet), then runs drift-audit.
# Exits non-zero on the first hard failure.

set -uo pipefail
cd "$(dirname "$0")/.."

status=0

echo "==> build (CGO_ENABLED=0)"
if CGO_ENABLED=0 go build ./...; then
	echo "OK: build"
else
	echo "FAIL: build"
	status=1
fi

echo
echo "==> smoke scripts"
shopt -s nullglob
smokes=(scripts/smoke/phase-*.sh)
if [ ${#smokes[@]} -eq 0 ]; then
	echo "SKIP: no per-phase smoke scripts yet"
else
	for s in "${smokes[@]}"; do
		echo "--- $s"
		if ! bash "$s"; then
			echo "FAIL: $s reported failures"
			status=1
		fi
	done
fi

echo
echo "==> drift-audit"
if ! ./scripts/drift-audit.sh; then
	status=1
fi

echo
echo "==> schema validation (D-031 layer 2; SKIPs without xmllint/schemas)"
if ! ./scripts/validate-schema.sh; then
	status=1
fi

echo
if [ "$status" -eq 0 ]; then
	echo "preflight: PASS"
else
	echo "preflight: FAIL"
fi
exit "$status"
