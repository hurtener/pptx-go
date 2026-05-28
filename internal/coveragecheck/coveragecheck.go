// Package coveragecheck implements the mechanical per-package coverage band
// gate described in CLAUDE.md §11. It parses a Go coverage profile, computes
// per-package statement coverage, and compares each package against the
// thresholds declared in coverage.json.
//
// The gate is intentionally boring: a package below its configured band fails
// the build, and — when RequireConfigured is set — a package that appears in
// the profile with no configured threshold also fails (so a new package
// cannot silently ship without a coverage decision).
package coveragecheck

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Config is the parsed coverage.json. Thresholds are keyed by full package
// import path (e.g. "github.com/hurtener/pptx-go/internal/coveragecheck").
type Config struct {
	// RequireConfigured fails the gate when a package present in the profile
	// has no entry in Packages. CLAUDE.md §11 mandates this once the codebase
	// is reorganized; it defaults to false during early scaffolding.
	RequireConfigured bool `json:"require_configured"`

	// Packages maps an import path to its band.
	Packages map[string]Band `json:"packages"`

	// Ignore lists import-path prefixes excluded from the gate entirely
	// (e.g. example or command packages that are not coverage-gated).
	Ignore []string `json:"ignore"`
}

// Band is a single package's coverage requirement.
type Band struct {
	// Min is the minimum acceptable statement coverage, as a percentage.
	Min float64 `json:"min"`
	// Reason documents why the band is set where it is — required for any
	// override below the §11 class default (class + reason).
	Reason string `json:"reason,omitempty"`
}

// PackageCoverage is the computed coverage for one package.
type PackageCoverage struct {
	Pkg     string
	Stmts   int
	Covered int
	Percent float64
}

// Result is the outcome for a single configured/observed package.
type Result struct {
	Pkg     string
	Percent float64
	Min     float64
	Status  string // "ok", "below", "unconfigured"
}

// Report is the full gate outcome.
type Report struct {
	Results []Result
}

// Failed reports whether any result is a hard failure.
func (r Report) Failed() bool {
	for _, res := range r.Results {
		if res.Status == "below" || res.Status == "unconfigured" {
			return true
		}
	}
	return false
}

// LoadConfig reads and parses a coverage.json file.
func LoadConfig(pathname string) (Config, error) {
	var cfg Config
	b, err := os.ReadFile(pathname)
	if err != nil {
		return cfg, fmt.Errorf("read coverage config: %w", err)
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("parse coverage config %s: %w", pathname, err)
	}
	if cfg.Packages == nil {
		cfg.Packages = map[string]Band{}
	}
	return cfg, nil
}

// ParseProfile reads a Go coverage profile and returns per-package coverage,
// keyed by import path. A profile line has the form:
//
//	<importpath>/<file>.go:<sl>.<sc>,<el>.<ec> <numStmts> <count>
//
// The leading "mode:" line is skipped.
func ParseProfile(r io.Reader) (map[string]*PackageCoverage, error) {
	pkgs := map[string]*PackageCoverage{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	first := true
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if first {
			first = false
			if strings.HasPrefix(line, "mode:") {
				continue
			}
			// No mode line; fall through and parse as a block.
		}
		pkg, stmts, count, err := parseBlock(line)
		if err != nil {
			return nil, err
		}
		pc := pkgs[pkg]
		if pc == nil {
			pc = &PackageCoverage{Pkg: pkg}
			pkgs[pkg] = pc
		}
		pc.Stmts += stmts
		if count > 0 {
			pc.Covered += stmts
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan profile: %w", err)
	}
	for _, pc := range pkgs {
		if pc.Stmts > 0 {
			pc.Percent = 100 * float64(pc.Covered) / float64(pc.Stmts)
		}
	}
	return pkgs, nil
}

// parseBlock parses one coverage profile block line into its package import
// path, statement count, and execution count.
func parseBlock(line string) (pkg string, stmts, count int, err error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return "", 0, 0, fmt.Errorf("malformed profile line: %q", line)
	}
	colon := strings.LastIndex(fields[0], ":")
	if colon < 0 {
		return "", 0, 0, fmt.Errorf("malformed profile block (no position): %q", fields[0])
	}
	file := fields[0][:colon]
	stmts, err = strconv.Atoi(fields[1])
	if err != nil {
		return "", 0, 0, fmt.Errorf("bad statement count in %q: %w", line, err)
	}
	count, err = strconv.Atoi(fields[2])
	if err != nil {
		return "", 0, 0, fmt.Errorf("bad execution count in %q: %w", line, err)
	}
	return path.Dir(file), stmts, count, nil
}

// Check evaluates per-package coverage against the config and returns a
// deterministically ordered report.
func Check(coverage map[string]*PackageCoverage, cfg Config) Report {
	var rep Report
	seen := map[string]bool{}

	pkgNames := make([]string, 0, len(coverage))
	for name := range coverage {
		pkgNames = append(pkgNames, name)
	}
	sort.Strings(pkgNames)

	for _, name := range pkgNames {
		if ignored(name, cfg.Ignore) {
			continue
		}
		pc := coverage[name]
		seen[name] = true
		band, ok := cfg.Packages[name]
		switch {
		case ok:
			status := "ok"
			if pc.Percent+1e-9 < band.Min {
				status = "below"
			}
			rep.Results = append(rep.Results, Result{Pkg: name, Percent: pc.Percent, Min: band.Min, Status: status})
		case cfg.RequireConfigured:
			rep.Results = append(rep.Results, Result{Pkg: name, Percent: pc.Percent, Status: "unconfigured"})
		default:
			rep.Results = append(rep.Results, Result{Pkg: name, Percent: pc.Percent, Status: "ok"})
		}
	}

	// A configured package that produced no coverage data at all (no test
	// binary ran for it) is a failure when RequireConfigured is set.
	if cfg.RequireConfigured {
		cfgNames := make([]string, 0, len(cfg.Packages))
		for name := range cfg.Packages {
			cfgNames = append(cfgNames, name)
		}
		sort.Strings(cfgNames)
		for _, name := range cfgNames {
			if !seen[name] && !ignored(name, cfg.Ignore) {
				rep.Results = append(rep.Results, Result{Pkg: name, Percent: 0, Min: cfg.Packages[name].Min, Status: "below"})
			}
		}
	}
	return rep
}

func ignored(pkg string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(pkg, p) {
			return true
		}
	}
	return false
}

// Format renders a report as a human-readable, deterministic table.
func Format(rep Report) string {
	var b strings.Builder
	var failed int
	for _, r := range rep.Results {
		switch r.Status {
		case "ok":
			fmt.Fprintf(&b, "OK:    %-60s %6.2f%%\n", r.Pkg, r.Percent)
		case "below":
			failed++
			fmt.Fprintf(&b, "FAIL:  %-60s %6.2f%% < %.2f%%\n", r.Pkg, r.Percent, r.Min)
		case "unconfigured":
			failed++
			fmt.Fprintf(&b, "FAIL:  %-60s %6.2f%% (no band in coverage.json)\n", r.Pkg, r.Percent)
		}
	}
	fmt.Fprintf(&b, "\ncoverage gate: %d package(s), %d failing\n", len(rep.Results), failed)
	return b.String()
}
