# Makefile for Peerless

# Variables
BINARY_NAME=peerless
GORELEASER=goreleaser
GOFMT=gofmt
GO_FILES=$(shell find . -name "*.go" -type f | grep -v vendor)
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Peerless - Check local directories against Transmission torrents"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
.PHONY: build
build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) main.go

.PHONY: build-all
build-all: ## Build the binary for all platforms (requires goreleaser)
	@echo "Building $(BINARY_NAME) for all platforms..."
	$(GORELEASER) build --clean

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	go clean

# Development targets
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	$(GOFMT) -s -w $(GO_FILES)

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: fmt vet ## Run all linting (format and vet)

# Release targets
.PHONY: release-patch
release-patch: ## Create patch release (e.g., 0.1.6 -> 0.1.7)
	@echo "Creating patch release..."
	./release.sh patch

.PHONY: release-minor
release-minor: ## Create minor release (e.g., 0.1.6 -> 0.2.0)
	@echo "Creating minor release..."
	./release.sh minor

.PHONY: release-major
release-major: ## Create major release (e.g., 0.1.6 -> 1.0.0)
	@echo "Creating major release..."
	./release.sh major

.PHONY: release
release: release-patch ## Default to patch release

# Installation targets
.PHONY: install
install: build ## Install binary locally
	@echo "Installing $(BINARY_NAME)..."
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/ || cp bin/$(BINARY_NAME) $$HOME/go/bin/ || echo "Install manually: cp bin/$(BINARY_NAME) to a directory in PATH"

.PHONY: install-local
install-local: build ## Install binary to /usr/local/bin (requires sudo)
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp bin/$(BINARY_NAME) /usr/local/bin/

# Utility targets
.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

.PHONY: update-deps
update-deps: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

.PHONY: version
version: ## Show current version
	@echo "Current version: $(VERSION)"

.PHONY: check-dirty
check-dirty: ## Check if working directory is dirty
	@! git diff-index --quiet HEAD -- && (echo "Working directory is dirty" && exit 1) || echo "Working directory is clean"

# Quick development workflow
.PHONY: dev
dev: lint test ## Run linting and tests (common development workflow)
	@echo "Development checks completed!"

.PHONY: ci
ci: deps lint test ## Run CI pipeline locally
	@echo "CI pipeline completed!"

# Pre-release checks
.PHONY: pre-release-checks
pre-release-checks: check-dirty deps lint test ## Run all pre-release checks
	@echo "Pre-release checks completed!"

# Show available goreleaser targets
.PHONY: goreleaser-help
goreleaser-help: ## Show goreleaser help
	@echo "Goreleaser targets:"
	@echo "  build          Build binaries"
	@echo "  build --clean  Clean and build binaries"
	@echo "  release        Create release (requires tag)"
	@echo "  release --clean Clean and create release"

# Docker targets (if applicable)
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -it $(BINARY_NAME):$(VERSION)

# Quick shortcuts
.PHONY: b
b: build ## Shortcut for build
.PHONY: t
t: test ## Shortcut for test
.PHONY: c
c: clean ## Shortcut for clean