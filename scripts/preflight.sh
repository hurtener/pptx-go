#!/usr/bin/env bash
#
# preflight.sh — the pre-commit / CI gate (CLAUDE.md §4.1).
#
# Builds the library CGo-free, runs every per-phase smoke script (each SKIPs
# gracefully where its surface is not built yet), then runs drift-audit and the
# schema-validation layer. Exits non-zero on the first hard failure.
#
# Performance (the smoke loop dominates wall-clock as phases accumulate):
#   1. Test binaries are PRE-COMPILED once up front (`go test -run='^$' ./...`).
#      Each smoke runs several `go test -run <pat>` calls that would otherwise
#      each trigger — and, run in parallel, concurrently contend on — the same
#      compilation. Warming the build cache once makes those calls reuse the
#      compiled binaries instead of racing the toolchain to rebuild them.
#   2. Smoke scripts run in a bounded PARALLEL pool. They are independent (each
#      only reads the repo and runs `go` against the shared, concurrency-safe
#      build cache), so they fan out across `xargs -P`. Output is captured
#      per-script and replayed in phase order, so the log reads exactly as the
#      old sequential run did.
#
# Knobs:
#   PREFLIGHT_JOBS=N   pool width (default: min(CPU count, 8); N=1 forces the
#                      old sequential behaviour).
#   PREFLIGHT_NO_PRECOMPILE=1   skip the test-binary pre-warm.

set -uo pipefail
cd "$(dirname "$0")/.."

status=0

# Pool width. Smokes are pre-warmed (light on CPU — mostly test execution), so a
# pool near the core count is fine; cap at 8 to bound toolchain oversubscription
# on a cold cache. Floor 2. Override via PREFLIGHT_JOBS.
jobs="${PREFLIGHT_JOBS:-0}"
case "$jobs" in ''|*[!0-9]*) jobs=0 ;; esac
if [ "$jobs" -le 0 ]; then
	cores="$(sysctl -n hw.ncpu 2>/dev/null || nproc 2>/dev/null || echo 4)"
	case "$cores" in ''|*[!0-9]*) cores=4 ;; esac
	jobs="$cores"
	[ "$jobs" -gt 8 ] && jobs=8
	[ "$jobs" -lt 2 ] && jobs=2
fi

echo "==> build (CGO_ENABLED=0)"
if CGO_ENABLED=0 go build ./...; then
	echo "OK: build"
else
	echo "FAIL: build"
	status=1
fi

if [ "${PREFLIGHT_NO_PRECOMPILE:-0}" != "1" ]; then
	echo
	echo "==> precompile test binaries (warms the cache for the smoke pool)"
	# -run='^$' matches no test, so this compiles + caches every test binary
	# without running anything. A genuine break surfaces (with detail) in the
	# smokes below, so this step is advisory: never fail the gate on it.
	if go test -run='^$' ./... >/dev/null 2>&1; then
		echo "OK: test binaries compiled"
	else
		echo "WARN: precompile reported an issue (continuing; smokes will detail it)"
	fi
fi

echo
echo "==> smoke scripts (parallel: ${jobs} jobs)"
shopt -s nullglob
smokes=(scripts/smoke/phase-*.sh)
if [ ${#smokes[@]} -eq 0 ]; then
	echo "SKIP: no per-phase smoke scripts yet"
elif [ "$jobs" -eq 1 ]; then
	# Sequential path (PREFLIGHT_JOBS=1): identical to the historical behaviour.
	for s in "${smokes[@]}"; do
		echo "--- $s"
		if ! bash "$s"; then
			echo "FAIL: $s reported failures"
			status=1
		fi
	done
else
	outdir="$(mktemp -d)"
	trap 'rm -rf "$outdir"' EXIT
	# Fan out. The worker body is inline (no `export -f` — exported-function
	# inheritance across `bash -c` is fragile on bash 3.2, which macOS ships);
	# it takes the smoke path and the output dir as positional args. Each worker
	# writes its combined stdout/stderr to a per-script log and touches a
	# .ok/.fail sentinel for the collector.
	printf '%s\n' "${smokes[@]}" | xargs -P "$jobs" -I {} bash -c '
		s="$1"; outdir="$2"; base="$(basename "$s" .sh)"
		if bash "$s" >"$outdir/$base.log" 2>&1; then
			: >"$outdir/$base.ok"
		else
			: >"$outdir/$base.fail"
		fi
	' _ {} "$outdir"
	# Replay in phase order so the log is deterministic regardless of schedule.
	for s in "${smokes[@]}"; do
		base="$(basename "$s" .sh)"
		echo "--- $s"
		cat "$outdir/$base.log" 2>/dev/null
		if [ -e "$outdir/$base.fail" ]; then
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
