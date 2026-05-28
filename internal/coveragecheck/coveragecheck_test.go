package coveragecheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseProfile(t *testing.T) {
	profile := strings.Join([]string{
		"mode: atomic",
		"github.com/hurtener/pptx-go/pkg/a/file.go:1.1,3.2 4 1",
		"github.com/hurtener/pptx-go/pkg/a/file.go:5.1,6.2 2 0",
		"github.com/hurtener/pptx-go/pkg/b/x.go:1.1,2.2 1 1",
		"",
	}, "\n")

	cov, err := ParseProfile(strings.NewReader(profile))
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}
	a := cov["github.com/hurtener/pptx-go/pkg/a"]
	if a == nil {
		t.Fatal("package a missing")
	}
	if a.Stmts != 6 || a.Covered != 4 {
		t.Fatalf("a: got stmts=%d covered=%d, want 6/4", a.Stmts, a.Covered)
	}
	if got, want := a.Percent, 100*4.0/6.0; got != want {
		t.Fatalf("a percent: got %v want %v", got, want)
	}
	b := cov["github.com/hurtener/pptx-go/pkg/b"]
	if b == nil || b.Percent != 100 {
		t.Fatalf("b: got %+v, want 100%%", b)
	}
}

func TestParseProfileNoModeLine(t *testing.T) {
	cov, err := ParseProfile(strings.NewReader("github.com/x/y/f.go:1.1,2.2 3 1\n"))
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}
	if cov["github.com/x/y"].Covered != 3 {
		t.Fatalf("got %+v", cov["github.com/x/y"])
	}
}

func TestParseProfileErrors(t *testing.T) {
	cases := map[string]string{
		"too few fields": "mode: atomic\nbadline\n",
		"no position":    "mode: atomic\nnopkgfile 2 1\n",
		"bad stmt count": "mode: atomic\ngithub.com/x/f.go:1.1,2.2 NaN 1\n",
		"bad exec count": "mode: atomic\ngithub.com/x/f.go:1.1,2.2 2 NaN\n",
	}
	for name, in := range cases {
		if _, err := ParseProfile(strings.NewReader(in)); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

func TestCheckBands(t *testing.T) {
	cov := map[string]*PackageCoverage{
		"github.com/x/ok":     {Pkg: "github.com/x/ok", Percent: 90},
		"github.com/x/low":    {Pkg: "github.com/x/low", Percent: 50},
		"github.com/x/extra":  {Pkg: "github.com/x/extra", Percent: 33},
		"github.com/x/ignore": {Pkg: "github.com/x/ignore", Percent: 0},
	}
	cfg := Config{
		Ignore: []string{"github.com/x/ignore"},
		Packages: map[string]Band{
			"github.com/x/ok":  {Min: 85},
			"github.com/x/low": {Min: 85},
		},
	}

	rep := Check(cov, cfg)
	if !rep.Failed() {
		t.Fatal("expected failure: github.com/x/low is below its band")
	}
	status := map[string]string{}
	for _, r := range rep.Results {
		status[r.Pkg] = r.Status
	}
	if status["github.com/x/ok"] != "ok" {
		t.Errorf("ok: got %q", status["github.com/x/ok"])
	}
	if status["github.com/x/low"] != "below" {
		t.Errorf("low: got %q", status["github.com/x/low"])
	}
	if status["github.com/x/extra"] != "ok" {
		t.Errorf("extra (unconfigured, lenient): got %q", status["github.com/x/extra"])
	}
	if _, ok := status["github.com/x/ignore"]; ok {
		t.Error("ignored package should not appear in the report")
	}
}

func TestCheckLenientUnconfiguredPasses(t *testing.T) {
	cov := map[string]*PackageCoverage{
		"github.com/x/extra": {Pkg: "github.com/x/extra", Percent: 1},
		"github.com/x/ok":    {Pkg: "github.com/x/ok", Percent: 90},
	}
	cfg := Config{Packages: map[string]Band{"github.com/x/ok": {Min: 85}}}
	rep := Check(cov, cfg)
	if rep.Failed() {
		t.Fatal("with RequireConfigured=false, an unconfigured package must not fail the gate")
	}
}

func TestCheckRequireConfigured(t *testing.T) {
	cov := map[string]*PackageCoverage{
		"github.com/x/extra": {Pkg: "github.com/x/extra", Percent: 90},
	}
	cfg := Config{
		RequireConfigured: true,
		Packages: map[string]Band{
			"github.com/x/missing": {Min: 80}, // configured but absent from profile
		},
	}
	rep := Check(cov, cfg)
	if !rep.Failed() {
		t.Fatal("expected failure: unconfigured covered package + missing configured package")
	}
	var sawUnconfigured, sawMissing bool
	for _, r := range rep.Results {
		if r.Pkg == "github.com/x/extra" && r.Status == "unconfigured" {
			sawUnconfigured = true
		}
		if r.Pkg == "github.com/x/missing" && r.Status == "below" {
			sawMissing = true
		}
	}
	if !sawUnconfigured {
		t.Error("expected 'unconfigured' status for the covered-but-unbanded package")
	}
	if !sawMissing {
		t.Error("expected 'below' status for the configured-but-uncovered package")
	}
}

func TestCheckDeterministicOrder(t *testing.T) {
	cov := map[string]*PackageCoverage{
		"github.com/z": {Pkg: "github.com/z", Percent: 90},
		"github.com/a": {Pkg: "github.com/a", Percent: 90},
		"github.com/m": {Pkg: "github.com/m", Percent: 90},
	}
	rep := Check(cov, Config{})
	got := []string{rep.Results[0].Pkg, rep.Results[1].Pkg, rep.Results[2].Pkg}
	want := []string{"github.com/a", "github.com/m", "github.com/z"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order: got %v want %v", got, want)
		}
	}
}

func TestFormat(t *testing.T) {
	rep := Report{Results: []Result{
		{Pkg: "github.com/x/ok", Percent: 90, Min: 85, Status: "ok"},
		{Pkg: "github.com/x/low", Percent: 50, Min: 85, Status: "below"},
		{Pkg: "github.com/x/un", Percent: 10, Status: "unconfigured"},
	}}
	out := Format(rep)
	for _, want := range []string{"OK:", "FAIL:", "3 package(s), 2 failing"} {
		if !strings.Contains(out, want) {
			t.Errorf("Format output missing %q:\n%s", want, out)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.json")
	if err := os.WriteFile(good, []byte(`{"require_configured":true,"packages":{"a":{"min":80}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(good)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !cfg.RequireConfigured || cfg.Packages["a"].Min != 80 {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	if _, err := LoadConfig(filepath.Join(dir, "missing.json")); err == nil {
		t.Error("expected error for missing file")
	}

	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(bad); err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestLoadConfigNilPackagesInitialized(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "c.json")
	if err := os.WriteFile(p, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Packages == nil {
		t.Fatal("Packages map should be initialized")
	}
}
