BINARY  := gogen
PKG     := ./cmd/gogen
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

# Example generation defaults (override on the command line, e.g. MODULE=... ARCH=...)
MODULE   ?= github.com/you/app
ARCH     ?= clean
ROUTER   ?= gin
FRONTEND ?= none
AUTH     ?=
ENTITIES ?= --entity User --entity Product
AUTH_FLAG := $(if $(filter true,$(AUTH)),--auth,)

.DEFAULT_GOAL := help
.PHONY: help build install run gen serve test test-coverage bench lint lint-fix \
	examples clean-examples release-snapshot clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "} {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

build: ## Build the gogen binary into ./gogen
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

install: ## Install gogen into GOBIN (or GOPATH/bin)
	go install -ldflags "$(LDFLAGS)" $(PKG)

run: ## Generate a sample project (clean architecture, gin)
	go run $(PKG) --module github.com/indalyadav/example-project --arch clean --router gin $(ENTITIES)

gen: ## Generate a project from variables: make gen ARCH=monolith FRONTEND=react AUTH=true
	go run $(PKG) --module $(MODULE) --arch $(ARCH) --router $(ROUTER) --frontend $(FRONTEND) $(AUTH_FLAG) $(ENTITIES)

serve: ## Start the web UI (browser configurator)
	go run $(PKG) serve

test: ## Run all tests
	go test -v ./...

test-coverage: ## Run tests with coverage and open the HTML report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

bench: ## Run benchmarks
	go test -bench=. ./...

lint: ## Run golangci-lint
	golangci-lint run

lint-fix: ## Run golangci-lint with --fix
	golangci-lint run --fix

examples: build ## Generate one example project per architecture into ./examples
	mkdir -p examples
	cd examples && rm -rf simple-example && ../$(BINARY) --module github.com/indalyadav/simple-example --arch simple --router gin $(ENTITIES)
	cd examples && rm -rf layered-example && ../$(BINARY) --module github.com/indalyadav/layered-example --arch layered --router gin $(ENTITIES)
	cd examples && rm -rf clean-example && ../$(BINARY) --module github.com/indalyadav/clean-example --arch clean --router gin $(ENTITIES)
	cd examples && rm -rf microservice-example && ../$(BINARY) --module github.com/indalyadav/microservice-example --arch microservice --router gin $(ENTITIES)
	cd examples && rm -rf monolith-example && ../$(BINARY) --module github.com/indalyadav/monolith-example --arch monolith --router chi $(ENTITIES)
	cd examples && rm -rf auth-example && ../$(BINARY) --module github.com/indalyadav/auth-example --arch clean --router gin --auth $(ENTITIES)
	cd examples && rm -rf fullstack-example && ../$(BINARY) --module github.com/indalyadav/fullstack-example --arch monolith --router gin --frontend react --auth $(ENTITIES)
	@echo "✅ examples generated in ./examples"

clean-examples: ## Remove generated example projects
	rm -rf examples

release-snapshot: ## Build release archives + Homebrew formula locally (no publish, output in ./dist)
	goreleaser release --snapshot --clean

clean: ## Remove built binary, coverage output and dist/
	rm -rf $(BINARY) coverage.out dist
