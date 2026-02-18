# ==============================================================================
# emv-merchant-emvqr Makefile
# ==============================================================================

MODULE      := github.com/hussainpithawala/emv-merchant-qr-lib
PACKAGE     := ./...
BINARY_NAME := emv-merchant-qr-lib

# Version is derived from the latest git tag; falls back to "dev" if no tag exists.
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION  := $(shell go version | awk '{print $$3}')

# Tool versions
GOLANGCI_LINT_VERSION := v2.4.0
STATICCHECK_VERSION   := latest

# Directories
BUILD_DIR   := ./bin
COVERAGE    := coverage.out
COVERAGE_HTML := coverage.html

# Go flags
GOFLAGS     := -trimpath
LDFLAGS     := -s -w \
               -X '$(MODULE).Version=$(VERSION)' \
               -X '$(MODULE).Commit=$(COMMIT)' \
               -X '$(MODULE).BuildDate=$(BUILD_DATE)'
TEST_FLAGS  := -race -count=1
BENCH_FLAGS := -bench=. -benchmem

# Colours (safe-guards for terminals that don't support colours)
RESET  := $(shell tput sgr0    2>/dev/null || echo "")
BOLD   := $(shell tput bold    2>/dev/null || echo "")
GREEN  := $(shell tput setaf 2 2>/dev/null || echo "")
YELLOW := $(shell tput setaf 3 2>/dev/null || echo "")
CYAN   := $(shell tput setaf 6 2>/dev/null || echo "")
RED    := $(shell tput setaf 1 2>/dev/null || echo "")

# Helper macro — prints a section header
define section
	@echo "$(BOLD)$(CYAN) ▶ $(1)$(RESET)"
endef

# Helper macro — prints success
define ok
	@echo "$(GREEN)✔ $(1)$(RESET)"
endef

# Helper macro — prints a warning
define warn
	@echo "$(YELLOW)⚠ $(1)$(RESET)"
endef

.DEFAULT_GOAL := help
.PHONY: install-lint lint fmt vet static-check check-mod build examples \
        deps update-deps validate ci pre-release release release-ci \
        clean godoc version info test help test-unit

# ==============================================================================
# DEPENDENCY MANAGEMENT
# ==============================================================================

## deps: Download and tidy module dependencies
deps:
	$(call section,Downloading dependencies)
	@go mod download
	@go mod tidy
	$(call ok,Dependencies ready)

## update-deps: Upgrade all dependencies to their latest minor/patch versions
update-deps:
	$(call section,Updating dependencies)
	@go get -u $(PACKAGE)
	@go mod tidy
	$(call ok,Dependencies updated — review go.sum before committing)

## check-mod: Verify go.mod and go.sum are consistent and tidy
check-mod:
	$(call section,Checking module consistency)
	@go mod verify
	@go mod tidy
	$(call ok,Module is tidy and verified)

# ==============================================================================
# TOOLING INSTALLATION
# ==============================================================================

## install-lint: Install golangci-lint ($(GOLANGCI_LINT_VERSION)) and staticcheck
install-lint:
	$(call section,Installing linting tools)
	@echo "  Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@echo "  Installing staticcheck..."
	@go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	$(call ok,Linting tools installed)

# ==============================================================================
# CODE QUALITY
# ==============================================================================

## fmt: Format all Go source files with gofmt
fmt:
	$(call section,Formatting source files)
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		gofmt -w .; \
		echo "$(YELLOW)Formatted:$(RESET)"; \
		echo "$$UNFORMATTED" | sed 's/^/  /'; \
	else \
		echo "  All files already formatted."; \
	fi
	$(call ok,Formatting complete)

## vet: Run go vet to catch common errors
vet:
	$(call section,Running go vet)
	@go vet $(PACKAGE)
	$(call ok,go vet passed)

## lint: Run golangci-lint with project configuration
lint:
	$(call section,Running golangci-lint)
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(YELLOW)golangci-lint not found — run 'make install-lint'$(RESET)"; \
		exit 1; \
	fi
	@golangci-lint run $(PACKAGE)
	$(call ok,golangci-lint passed)

## static-check: Run staticcheck for advanced static analysis
static-check:
	$(call section,Running staticcheck)
	@if ! command -v staticcheck >/dev/null 2>&1; then \
		echo "$(YELLOW)staticcheck not found — run 'make install-lint'$(RESET)"; \
		exit 1; \
	fi
	@staticcheck $(PACKAGE)
	$(call ok,staticcheck passed)

# ==============================================================================
# TESTING
# ==============================================================================

## test: Run the full test suite (unit + example tests) with race detector
test:
	$(call section,Running full test suite)
	@go test $(TEST_FLAGS) -coverprofile=$(COVERAGE) -covermode=atomic $(PACKAGE)
	@go tool cover -func=$(COVERAGE) | tail -1
	$(call ok,All tests passed)

## test-unit: Run only non-example unit tests (faster inner-loop target)
test-unit:
	$(call section,Running unit tests)
	@go test $(TEST_FLAGS) -run '^Test' $(PACKAGE)
	$(call ok,Unit tests passed)

# ==============================================================================
# BUILD
# ==============================================================================

## build: Compile the library and verify it builds cleanly (no binary produced)
build:
	$(call section,Building $(MODULE))
	@mkdir -p $(BUILD_DIR)
	@go build $(GOFLAGS) $(PACKAGE)
	$(call ok,Build succeeded — module: $(MODULE) version: $(VERSION))

## examples: Build and run all Example* functions as a smoke test
examples:
	$(call section,Running example functions)
	@go test $(TEST_FLAGS) -v -run '^Example' $(PACKAGE)
	$(call ok,All examples passed)

# ==============================================================================
# DOCUMENTATION
# ==============================================================================

## godoc: Serve package documentation locally on http://localhost:6060
godoc:
	$(call section,Starting godoc server)
	@if ! command -v godoc >/dev/null 2>&1; then \
		echo "  Installing godoc..."; \
		go install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@echo "  Documentation: $(BOLD)http://localhost:6060/pkg/$(MODULE)/$(RESET)"
	@godoc -http=:6060

# ==============================================================================
# VERSION & INFO
# ==============================================================================

## version: Print the current version derived from git tags
version:
	@echo "$(BOLD)Version:$(RESET)    $(VERSION)"
	@echo "$(BOLD)Commit:$(RESET)     $(COMMIT)"
	@echo "$(BOLD)Build date:$(RESET) $(BUILD_DATE)"

## info: Print full build and environment information
info:
	$(call section,Environment)
	@echo "  Module:       $(MODULE)"
	@echo "  Version:      $(VERSION)"
	@echo "  Commit:       $(COMMIT)"
	@echo "  Build date:   $(BUILD_DATE)"
	@echo "  Go version:   $(GO_VERSION)"
	@echo "  GOPATH:       $$(go env GOPATH)"
	@echo "  GOROOT:       $$(go env GOROOT)"
	@echo "  GOOS/GOARCH:  $$(go env GOOS)/$$(go env GOARCH)"
	@echo "  CGO enabled:  $$(go env CGO_ENABLED)"

# ==============================================================================
# COMPOSITE WORKFLOWS
# ==============================================================================

## validate: Run fmt-check, vet, lint, static-check, check-mod, and tests
validate: _fmt-check vet lint static-check check-mod test
	$(call ok,All validation checks passed)

# Internal target: check formatting without writing (used in validate / CI)
_fmt-check:
	$(call section,Checking formatting)
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "$(RED)The following files need formatting (run 'make fmt'):$(RESET)"; \
		echo "$$UNFORMATTED" | sed 's/^/  /'; \
		exit 1; \
	fi
	$(call ok,All files are correctly formatted)

## ci: Full CI pipeline — deps, validate, build, examples
ci: deps validate build examples
	$(call ok,CI pipeline completed successfully)

## pre-release: Validate, then confirm the version tag does not already exist
pre-release: validate
	$(call section,Pre-release checks)
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "$(RED)No git tag found. Tag a version first:$(RESET)"; \
		echo "  git tag v1.x.x && git push origin v1.x.x"; \
		exit 1; \
	fi
	@if echo "$(VERSION)" | grep -q "dirty"; then \
		echo "$(RED)Working tree is dirty — commit or stash changes before releasing$(RESET)"; \
		exit 1; \
	fi
	@echo "  Version $(BOLD)$(VERSION)$(RESET) is clean and ready."
	$(call ok,Pre-release checks passed)

## release: Tag and push HEAD — set VERSION variable, e.g. make release VERSION=v1.2.3
release:
	$(call section,Releasing $(VERSION))
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "$(RED)Provide a semantic version: make release VERSION=v1.2.3$(RESET)"; \
		exit 1; \
	fi
	@if ! echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+'; then \
		echo "$(RED)Version must follow semver: vMAJOR.MINOR.PATCH$(RESET)"; \
		exit 1; \
	fi
	@if git rev-parse "$(VERSION)" >/dev/null 2>&1; then \
		echo "$(RED)Tag $(VERSION) already exists$(RESET)"; \
		exit 1; \
	fi
	$(MAKE) pre-release
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	$(call ok,Tagged and pushed $(VERSION) — GitHub Actions will create the release)

## release-ci: CI-side release step — verifies tag, runs full pipeline (no push)
release-ci: deps validate build examples
	$(call section,Release CI verification)
	@echo "  Tag:     $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@if echo "$(VERSION)" | grep -q "dirty"; then \
		echo "$(RED)Dirty build detected in CI — aborting$(RESET)"; \
		exit 1; \
	fi
	$(call ok,Release CI pipeline passed for $(VERSION))

# ==============================================================================
# CLEANUP
# ==============================================================================

## clean: Remove build artefacts, coverage reports, and tool caches
clean:
	$(call section,Cleaning build artefacts)
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE) $(COVERAGE_HTML)
	@go clean -cache -testcache
	$(call ok,Clean complete)

# ==============================================================================
# HELP
# ==============================================================================

## help: Show this help message (default target)
help:
	@echo ""
	@echo "$(BOLD)emv-merchant-qr-lib$(RESET) — EMV® QR Code encoder/decoder library"
	@echo "$(BOLD)Module:$(RESET) $(MODULE)"
	@echo "$(BOLD)Version:$(RESET) $(VERSION)"
	@echo ""
	@echo "$(BOLD)$(CYAN)Usage:$(RESET)"
	@echo "  make $(BOLD)<target>$(RESET) [VARIABLE=value ...]"
	@echo ""
	@echo "$(BOLD)$(CYAN)Targets:$(RESET)"
	@grep -E '^## ' $(MAKEFILE_LIST) \
		| sed 's/^## //' \
		| awk -F': ' '{ printf "  $(BOLD)%-18s$(RESET) %s\n", $$1, $$2 }'
	@echo ""
	@echo "$(BOLD)$(CYAN)Variables:$(RESET)"
	@echo "  $(BOLD)VERSION$(RESET)   Semantic version tag for release (e.g. v1.2.3)"
	@echo ""
	@echo "$(BOLD)$(CYAN)Examples:$(RESET)"
	@echo "  make ci                    # Full CI pipeline"
	@echo "  make test                  # Run all tests"
	@echo "  make test-unit             # Run unit tests only (faster)"
	@echo "  make release VERSION=v1.1.0"
	@echo "  make godoc                 # Browse docs at http://localhost:6060"
	@echo ""