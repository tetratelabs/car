# Copyright 2021 Tetrate
# Licensed under the Apache License, Version 2.0 (the "License")
#
# This script uses automatic variables (ex $<, $(@D)) and substitution references $(<:.signed=)
# Please see GNU make's documentation if unfamiliar: https://www.gnu.org/software/make/manual/html_node/
.PHONY: test build dist clean format lint check

# Include versions of tools we build on-demand
include Tools.mk

# This should be driven by automation and result in N.N.N, not vN.N.N
VERSION ?= dev

# This selects the goroot to use in the following priority order:
# 1. ${GOROOT}          - Ex actions/setup-go
# 2. ${GOROOT_1_17_X64} - Ex GitHub Actions runner
# 3. $(go env GOROOT)   - Implicit from the go binary in the path
#
# There may be multiple GOROOT variables, so pick the one matching go.mod.
go_release          := $(shell sed -ne 's/^go //gp' go.mod)
# https://github.com/actions/runner/blob/master/src/Runner.Common/Constants.cs
github_runner_arch  := $(if $(findstring $(shell uname -m),x86_64),X64,ARM64)
github_goroot_name  := GOROOT_$(subst .,_,$(go_release))_$(github_runner_arch)
github_goroot_val   := $(value $(github_goroot_name))
goroot_path         := $(shell go env GOROOT 2>/dev/null)
goroot              := $(firstword $(GOROOT) $(github_goroot_val) $(goroot_path))

ifndef goroot
$(error could not determine GOROOT)
endif

# We must ensure `go` executes with GOROOT and PATH variables exported:
# * GOROOT ensures versions don't conflict with /usr/local/go or c:\Go
# * PATH ensures tools like golint can fork and execute the correct go binary.
#
# We may be using a very old version of Make (ex. 3.81 on macOS). This means we
# can't re-set GOROOT or PATH via 'export' or use '.ONESHELL' to persist
# variables across lines. Hence, we set variables on one-line.
go := export PATH="$(goroot)/bin:$${PATH}" && export GOROOT="$(goroot)" && go

# Set variables corresponding to the selected goroot and the current host.
goarch := $(shell $(go) env GOARCH)
goexe  := $(shell $(go) env GOEXE)
goos   := $(shell $(go) env GOOS)

# Build the path to the car binary for the current runtime (goos,goarch)
current_binary_path := build/car_$(goos)_$(goarch)
current_binary      := $(current_binary_path)/car$(goexe)

# ANSI escape codes. f_ means foreground, b_ background.
# See https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
b_black            := $(shell printf "\33[40m")
f_white            := $(shell printf "\33[97m")
f_gray             := $(shell printf "\33[37m")
f_dark_gray        := $(shell printf "\33[90m")
f_bright_magenta   := $(shell printf "\33[95m")
b_bright_magenta   := $(shell printf "\33[105m")
ansi_reset         := $(shell printf "\33[0m")
ansi_car           := $(b_black)$(f_white)car$(ansi_reset)
ansi_format_dark   := $(f_gray)$(f_bright_magenta)%-10s$(ansi_reset) $(f_dark_gray)%s$(ansi_reset)\n
ansi_format_bright := $(f_white)$(f_bright_magenta)%-10s$(ansi_reset) $(f_white)$(b_bright_magenta)%s$(ansi_reset)\n

# This formats help statements in ANSI colors. To hide a target from help, don't comment it with a trailing '##'.
help: ## Describe how to use each target
	@printf "$(ansi_car)$(f_white)\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[0-9a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "$(ansi_format_dark)", $$1, $$2}' $(MAKEFILE_LIST)

build: $(current_binary) ## Build the car binary

test: ## Run all unit tests
	@printf "$(ansi_format_dark)" test "running unit tests"
	@$(go) test $(main_packages)
	@printf "$(ansi_format_bright)" test "ok"

coverpkg = $(main_packages: =,)
coverage: ## Generate test coverage
	@printf "$(ansi_format_dark)" coverage "running unit tests with coverage"
	@$(go) test -coverprofile=coverage.txt -covermode=atomic --coverpkg=$(coverpkg) $(main_packages)
	@$(go) tool cover -func coverage.txt
	@printf "$(ansi_format_bright)" coverage "ok"

platforms := darwin_amd64 darwin_arm64 linux_amd64 linux_arm64

# Make 3.81 doesn't support '**' globbing: Set explicitly instead of recursion.
all_sources   := $(wildcard *.go */*.go */*/*.go */*/*/*.go */*/*/*.go */*/*/*/*.go)
all_testdata  := $(wildcard testdata/* */testdata/* */*/testdata/* */*/*/testdata/*)
# main_sources compose the binary, so exclude test sources
main_sources  := $(wildcard $(filter-out %_test.go $(all_testdata), $(all_sources)))
# main_packages collect the unique main source directories (sort will dedupe).
# Paths need to all start with ./, so we do that manually vs foreach which strips it.
main_packages := $(sort $(foreach f,$(dir $(main_sources)),$(if $(findstring ./,$(f)),./,./$(f))))

build/car_%/car: $(main_sources)
	$(call go-build, $@)

dist/car_$(VERSION)_%.tar.gz: build/car_%/car
	@printf "$(ansi_format_dark)" tar.gz "tarring $@"
	@mkdir -p $(@D)
	@tar -C $(<D) -cpzf $@ $(<F)
	@printf "$(ansi_format_bright)" tar.gz "ok"

archives  := $(platforms:%=dist/car_$(VERSION)_%.tar.gz)
checksums := dist/car_$(VERSION)_checksums.txt

# Darwin doesn't have sha256sum. See https://github.com/actions/virtual-environments/issues/90
sha256sum := $(if $(findstring darwin,$(goos)),shasum -a 256,sha256sum)
$(checksums): $(archives) $(packages)
	@printf "$(ansi_format_dark)" sha256sum "generating $@"
	@$(sha256sum) $^ > $@
	@printf "$(ansi_format_bright)" sha256sum "ok"

# dist generates the assets that attach to a release
# Ex. https://github.com/tetratelabs/car/releases/tag/v$(VERSION)
dist: $(archives) $(packages) $(checksums) ## Generate release assets

clean: ## Ensure a clean build
	@printf "$(ansi_format_dark)" clean "deleting temporary files"
	@rm -rf dist build coverage.txt
	@$(go) clean -testcache
	@printf "$(ansi_format_bright)" clean "ok"

# format is a PHONY target, so always runs. This allows skipping when sources didn't change.
build/format: go.mod $(all_sources)
	@$(go) mod tidy
	@$(go) run $(licenser) apply -r "Tetrate"
	@$(go) run $(gofumpt) -l -w $(all_sources)
	@# -local ensures consistent ordering of our module in imports
	@$(go) run $(gosimports) -local github.com/tetratelabs/ -w $(shell find . -name '*.go' -type f)
	@mkdir -p $(@D) && touch $@

format:
	@printf "$(ansi_format_dark)" format "formatting project files"
	@$(MAKE) build/format
	@printf "$(ansi_format_bright)" format "ok"

golangci_lint_path := $(shell go env GOPATH)/bin/golangci-lint

$(golangci_lint_path):
	@go install $(golangci_lint)

golangci_lint_goarch ?= $(shell go env GOARCH)

# lint is a PHONY target, so always runs. This allows skipping when sources didn't change.
build/lint: $(all_sources) $(golangci_lint_path)
	@GOARCH=$(golangci_lint_goarch) CGO_ENABLED=0 $(golangci_lint_path) run --timeout 5m $(filter-out $(golangci_lint_path), $<)
	@mkdir -p $(@D) && touch $@

lint:
	@printf "$(ansi_format_dark)" lint "Running linters"
	@$(MAKE) build/lint
	@printf "$(ansi_format_bright)" lint "ok"

# CI blocks merge until this passes. If this fails, run "make check" locally and commit the difference.
# This formats code before running lint, as it is annoying to tell people to format first!
check: ## Verify contents of last commit
	@$(MAKE) lint
	@$(MAKE) format
	@# Make sure the check-in is clean
	@if [ ! -z "`git status -s`" ]; then \
		echo "The following differences will fail CI until committed:"; \
		git diff --exit-code; \
	fi

# define macros for multi-platform builds. these parse the filename being built
go-arch = $(if $(findstring amd64,$1),amd64,arm64)
go-os   = $(if $(findstring linux,$1),linux,darwin)
define go-build
	@printf "$(ansi_format_dark)" build "building $1"
	@# $(go:go=) removes the trailing 'go', so we can insert cross-build variables
	@$(go:go=) CGO_ENABLED=0 GOOS=$(call go-os,$1) GOARCH=$(call go-arch,$1) go build \
		-ldflags "-s -w -X main.version=$(VERSION)" \
		-o $1 ./cmd/car
	@printf "$(ansi_format_bright)" build "ok"
endef
