// Package conformance validates that an OPC package pptx-go emits is
// structurally sound — independent of the writer/reader round-trip. It is the
// pure-Go layer of the validity strategy (D-031, CLAUDE.md §11): it catches
// dangling relationships, parts missing a content type, unresolved targets,
// malformed pack URIs, and missing required parts. Schema conformance and
// "a real office app opens it" are checked by separate layers (xmllint XSD,
// LibreOffice headless).
//
// This package is test/CI tooling; it is never imported by the shipped
// builder or renderer.
package conformance

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hurtener/pptx-go/internal/opc"
)

// Severity classifies an Issue.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Issue is one conformance finding.
type Issue struct {
	Severity Severity
	Part     string // pack URI the issue is about ("" = package-level)
	Message  string
}

func (i Issue) String() string {
	where := i.Part
	if where == "" {
		where = "<package>"
	}
	return fmt.Sprintf("%s: [%s] %s", i.Severity, where, i.Message)
}

// Options tunes the checks.
type Options struct {
	// RequiredParts are pack URIs that must be present (full-deck
	// completeness). Empty skips the completeness check — useful while the
	// builder is incomplete.
	RequiredParts []string
}

// Report is the validation result.
type Report struct {
	Issues []Issue
}

// Errors returns only the error-severity issues.
func (r Report) Errors() []Issue {
	var out []Issue
	for _, i := range r.Issues {
		if i.Severity == SeverityError {
			out = append(out, i)
		}
	}
	return out
}

// OK reports whether there are no error-severity issues.
func (r Report) OK() bool { return len(r.Errors()) == 0 }

func (r Report) String() string {
	if len(r.Issues) == 0 {
		return "conformance: OK (no issues)"
	}
	var b strings.Builder
	for _, i := range r.Issues {
		fmt.Fprintln(&b, i.String())
	}
	return b.String()
}

// rIDPattern matches a relationship-ID attribute value (e.g. "rId5").
var rIDPattern = regexp.MustCompile(`"(rId\d+)"`)

// Validate runs the structural checks against an open package.
func Validate(pkg *opc.Package, opts Options) Report {
	var rep Report
	add := func(sev Severity, part, msg string) {
		rep.Issues = append(rep.Issues, Issue{Severity: sev, Part: part, Message: msg})
	}

	parts := pkg.AllParts()
	// Deterministic order for stable output.
	sort.Slice(parts, func(a, b int) bool {
		return parts[a].PartURI().URI() < parts[b].PartURI().URI()
	})

	for _, part := range parts {
		uri := part.PartURI().URI()

		// 1. Content-type coverage.
		if strings.TrimSpace(part.ContentType()) == "" {
			add(SeverityError, uri, "part has no content type (missing from [Content_Types].xml)")
		}

		// 2. Pack-URI validity.
		if !strings.HasPrefix(uri, "/") {
			add(SeverityError, uri, "pack URI is not absolute (must start with '/')")
		}
		if strings.Contains(uri, "\\") || strings.Contains(uri, "..") {
			add(SeverityError, uri, "pack URI contains a backslash or '..' segment")
		}

		// 3 & 4. Relationship integrity. relIDs is the set the part declares;
		// it stays empty for parts with no rels, so any rId the XML references
		// is then flagged as dangling below.
		relIDs := map[string]bool{}
		for _, rel := range part.Relationships().All() {
			relIDs[rel.RID()] = true
			if rel.IsExternal() {
				continue
			}
			target := part.GetRelatedPart(rel.RID())
			if target == nil || !pkg.ContainsPart(target) {
				tref := ""
				if t := rel.TargetURI(); t != nil {
					tref = t.URI()
				}
				add(SeverityError, uri, fmt.Sprintf("relationship %s targets a part not in the package (%s)", rel.RID(), tref))
			}
		}

		// 4. Dangling rId references in the part XML.
		seen := map[string]bool{}
		for _, m := range rIDPattern.FindAllStringSubmatch(string(part.Blob()), -1) {
			rid := m[1]
			if seen[rid] {
				continue
			}
			seen[rid] = true
			if !relIDs[rid] {
				add(SeverityError, uri, fmt.Sprintf("XML references relationship %s but the part has no such relationship", rid))
			}
		}
	}

	// 5. Required parts present.
	for _, req := range opts.RequiredParts {
		if !pkg.ContainsPart(opc.NewPackURI(req)) {
			add(SeverityError, req, "required part is missing from the package")
		}
	}

	return rep
}

// ValidateBytes opens a .pptx byte slice and validates it.
func ValidateBytes(data []byte, opts Options) (Report, error) {
	pkg, err := opc.Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return Report{}, fmt.Errorf("open package: %w", err)
	}
	defer func() { _ = pkg.Close() }()
	return Validate(pkg, opts), nil
}
