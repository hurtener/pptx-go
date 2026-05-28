// Command coveragecheck runs the pptx-go per-package coverage band gate.
//
// Usage:
//
//	go run ./internal/coveragecheck/cmd/coveragecheck -profile=coverage.out -config=internal/coveragecheck/coverage.json
//
// It exits non-zero when any package is below its configured band (or, when
// require_configured is set, when a covered package has no configured band).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hurtener/pptx-go/internal/coveragecheck"
)

func main() {
	profile := flag.String("profile", "coverage.out", "path to the Go coverage profile")
	config := flag.String("config", "internal/coveragecheck/coverage.json", "path to coverage.json")
	flag.Parse()

	if err := run(*profile, *config); err != nil {
		fmt.Fprintln(os.Stderr, "coveragecheck:", err)
		os.Exit(1)
	}
}

func run(profilePath, configPath string) error {
	cfg, err := coveragecheck.LoadConfig(configPath)
	if err != nil {
		return err
	}

	f, err := os.Open(profilePath)
	if err != nil {
		return fmt.Errorf("open profile: %w", err)
	}
	defer func() { _ = f.Close() }()

	cov, err := coveragecheck.ParseProfile(f)
	if err != nil {
		return err
	}

	rep := coveragecheck.Check(cov, cfg)
	fmt.Print(coveragecheck.Format(rep))
	if rep.Failed() {
		return fmt.Errorf("coverage gate failed")
	}
	return nil
}
