# Makefile for Peerless

# Variables
BINARY_NAME=peerless
GORELEASER=goreleaser
GOFMT=gofmt
GO_FILES=$(shell find . -name "*.go" -type f | grep -v vendor)
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")
GITHUB_TOKEN ?= $(GITHUB_TOKEN)
DRY_RUN ?= false

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
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "⚠️  GITHUB_TOKEN not set, using default git credentials"; \
	else \
		echo "✅ GITHUB_TOKEN is set, using token authentication"; \
	fi
	./release.sh patch

.PHONY: release-minor
release-minor: ## Create minor release (e.g., 0.1.6 -> 0.2.0)
	@echo "Creating minor release..."
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "⚠️  GITHUB_TOKEN not set, using default git credentials"; \
	else \
		echo "✅ GITHUB_TOKEN is set, using token authentication"; \
	fi
	./release.sh minor

.PHONY: release-major
release-major: ## Create major release (e.g., 0.1.6 -> 1.0.0)
	@echo "Creating major release..."
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "⚠️  GITHUB_TOKEN not set, using default git credentials"; \
	else \
		echo "✅ GITHUB_TOKEN is set, using token authentication"; \
	fi
	./release.sh major

.PHONY: release
release: release-patch ## Default to patch release

# Release targets with explicit token check
.PHONY: release-with-token
release-with-token: ## Ensure GITHUB_TOKEN is set before releasing
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "❌ GITHUB_TOKEN is required for this target"; \
		echo "Set it with: export GITHUB_TOKEN=\"your_token_here\""; \
		echo "Or run: make release GITHUB_TOKEN=your_token_here"; \
		exit 1; \
	fi
	@echo "✅ GITHUB_TOKEN verified, proceeding with release..."
	$(MAKE) release-patch

.PHONY: release-dry-run
release-dry-run: ## Simulate release without actually releasing
	@echo "🔍 Simulating release process..."
	@echo "Current version: $(VERSION)"
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "⚠️  Would use default git credentials"; \
	else \
		echo "✅ Would use GITHUB_TOKEN for authentication"; \
	fi
	@echo "To perform an actual release, run: make release"

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

.PHONY: check-token
check-token: ## Check GitHub token status
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "❌ GITHUB_TOKEN is not set"; \
		echo "Set it with: export GITHUB_TOKEN=\"ghp_your_token_here\""; \
		echo "Or run: make target GITHUB_TOKEN=your_token"; \
	else \
		echo "✅ GITHUB_TOKEN is set (length: $${#GITHUB_TOKEN} characters)"; \
		echo "Token starts with: $$(echo $(GITHUB_TOKEN) | cut -c1-10)..."; \
	fi

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