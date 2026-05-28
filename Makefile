# pptx-go — canonical build / test / lint commands.
#
# Every target is run by CI and documented in CLAUDE.md §4. Targets no-op
# gracefully before the code they act on exists.

# The shipped artifact is CGo-free (P4). Tests use CGo only for the race
# detector (-race requires it); test binaries are not shipped (CLAUDE.md §5).
GO              ?= go
GOFLAGS         ?=
PKGS            ?= ./...
COVERAGE_OUT    ?= coverage.out
COVERAGE_CONFIG ?= internal/coveragecheck/coverage.json

.DEFAULT_GOAL := build

.PHONY: build test coverage bench vet lint drift-audit check-mirror \
        preflight install-hooks tidy clean help

## build: compile the library CGo-free.
build:
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(PKGS)

## test: run the full suite with the race detector (CGo-enabled test binaries).
test:
	CGO_ENABLED=1 $(GO) test -race $(GOFLAGS) $(PKGS)

## coverage: produce a per-package coverage profile and run the band gate.
coverage:
	CGO_ENABLED=1 $(GO) test -race -covermode=atomic -coverprofile=$(COVERAGE_OUT) $(PKGS)
	$(GO) run ./internal/coveragecheck/cmd/coveragecheck -profile=$(COVERAGE_OUT) -config=$(COVERAGE_CONFIG)

## bench: run the Go benchmarks (on demand — not a CI gate).
bench:
	$(GO) test -run='^$$' -bench=. -benchmem $(PKGS)

## vet: go vet.
vet:
	$(GO) vet $(PKGS)

## lint: golangci-lint (skips with a notice when not installed).
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "SKIP: golangci-lint not installed (https://golangci-lint.run/usage/install/)"; \
	fi

## drift-audit: design-coherence checks (RFC/plans/mirror/forbidden names).
drift-audit:
	./scripts/drift-audit.sh

## check-mirror: verify AGENTS.md == CLAUDE.md.
check-mirror:
	@diff -q AGENTS.md CLAUDE.md && echo "OK: AGENTS.md == CLAUDE.md"

## preflight: build + per-phase smoke checks + drift-audit (the pre-commit/CI gate).
preflight:
	./scripts/preflight.sh

## install-hooks: install the pre-commit hook (one-time, per clone).
install-hooks:
	./scripts/install-hooks.sh

## tidy: go mod tidy.
tidy:
	$(GO) mod tidy

## clean: remove build/coverage artifacts.
clean:
	rm -f $(COVERAGE_OUT)
	$(GO) clean

## help: list documented targets.
help:
	@grep -hE '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'
