#!/usr/bin/env bash
#
# validate-schema.sh — schema-conformance layer of the validity strategy
# (D-031, layer 2). Validates the XML parts of an emitted deck against the
# vendored ISO/IEC 29500:2016 transitional XSDs using xmllint.
#
# SKIPs gracefully (exit 0) when xmllint or the schemas are absent. Test/CI-only
# tooling — never enters the shipped artifact (P4).
#
# It emits the full-surface showcase deck (shapes, rich text, images, sections,
# speaker notes, tables, and the scene renderer) so validation exercises every
# part type — presentation, slides, masters, layouts, notes, and theme.

set -uo pipefail
cd "$(dirname "$0")/.."

SCHEMA_DIR="docs/specifications/ooxml-transitional"
DECK="${1:-test-output/schema-check.pptx}"

if ! command -v xmllint >/dev/null 2>&1; then
	echo "SKIP: xmllint not installed (libxml2-utils)"
	exit 0
fi
if [ ! -d "$SCHEMA_DIR" ] || [ -z "$(find "$SCHEMA_DIR" -name '*.xsd' 2>/dev/null | head -1)" ]; then
	echo "SKIP: no vendored XSDs in $SCHEMA_DIR (see docs/specifications/README.md)"
	exit 0
fi

if [ ! -f "$DECK" ]; then
	go run ./_gen/genshowcase "$DECK" >/dev/null 2>&1 || { echo "FAIL: could not emit the showcase deck"; exit 1; }
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
unzip -qq "$DECK" -d "$work" || { echo "FAIL: deck is not a valid zip"; exit 1; }

status=0
# check <schema-file> <part>...  — pml.xsd covers PresentationML parts;
# dml-main.xsd covers the theme.
check() {
	local schema="$SCHEMA_DIR/$1"
	shift
	[ -f "$schema" ] || return 0
	for f in "$@"; do
		[ -f "$f" ] || continue
		if ! xmllint --noout --schema "$schema" "$f" 2>>"$work/err"; then
			echo "FAIL: ${f#"$work"/} does not validate against $(basename "$schema")"
			status=1
		fi
	done
}

check pml.xsd \
	"$work"/ppt/presentation.xml \
	"$work"/ppt/slides/slide*.xml \
	"$work"/ppt/slideMasters/*.xml \
	"$work"/ppt/slideLayouts/*.xml \
	"$work"/ppt/notesMasters/*.xml \
	"$work"/ppt/notesSlides/*.xml \
	"$work"/ppt/presProps.xml \
	"$work"/ppt/viewProps.xml
check dml-main.xsd "$work"/ppt/theme/theme*.xml "$work"/ppt/tableStyles.xml
check shared-documentPropertiesExtended.xsd "$work"/docProps/app.xml

if [ "$status" -eq 0 ]; then
	echo "OK: emitted parts validate against the vendored transitional schemas"
else
	grep -iE "validity error|fails to validate|is not expected" "$work/err" 2>/dev/null | sed 's/^/  /' | head -20
fi
exit "$status"
