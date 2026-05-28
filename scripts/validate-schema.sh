#!/usr/bin/env bash
#
# validate-schema.sh — schema-conformance layer of the validity strategy
# (D-031, layer 2). Validates the XML parts of an emitted reference deck
# against the vendored ISO/IEC 29500 transitional XSDs using xmllint.
#
# SKIPs gracefully (exit 0) when the schemas are not yet vendored — see
# docs/specifications/README.md for how to vendor them. Test/CI-only tooling.

set -uo pipefail
cd "$(dirname "$0")/.."

SCHEMA_DIR="docs/specifications/ooxml-transitional"
DECK="${1:-test-output/reference.pptx}"

if ! command -v xmllint >/dev/null 2>&1; then
	echo "SKIP: xmllint not installed (libxml2-utils)"
	exit 0
fi
if [ ! -d "$SCHEMA_DIR" ] || [ -z "$(find "$SCHEMA_DIR" -name '*.xsd' 2>/dev/null | head -1)" ]; then
	echo "SKIP: no vendored XSDs in $SCHEMA_DIR (see docs/specifications/README.md)"
	exit 0
fi

# Emit the reference deck if absent.
if [ ! -f "$DECK" ]; then
	go run ./_gen/genrefdeck "$DECK" || { echo "FAIL: could not emit reference deck"; exit 1; }
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
unzip -qq "$DECK" -d "$work" || { echo "FAIL: deck is not a valid zip"; exit 1; }

# Convention: pml.xsd validates presentation/slide parts; dml.xsd the theme.
status=0
validate() { # <xml-glob> <schema>
	local schema="$SCHEMA_DIR/$2"
	[ -f "$schema" ] || return 0
	for f in $1; do
		[ -f "$f" ] || continue
		if ! xmllint --noout --schema "$schema" "$f" 2>>"$work/err"; then
			echo "FAIL: $f does not validate against $2"
			status=1
		fi
	done
}
validate "$work/ppt/presentation.xml" "pml.xsd"
validate "$work/ppt/slides/slide*.xml" "pml.xsd"
validate "$work/ppt/theme/theme*.xml" "dml.xsd"

if [ "$status" -eq 0 ]; then
	echo "OK: emitted parts validate against the vendored transitional schemas"
fi
exit "$status"
