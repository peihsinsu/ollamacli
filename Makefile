# Project information
PROJECT_NAME := ollamacli
BINARY_NAME := ollamacli
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | awk '{print $$3}')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build variables
BUILD_DIR := build
DIST_DIR := dist
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
DESTDIR ?=
INSTALL ?= /usr/bin/install
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"

# Go variables
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

# Tools
GOLANGCI_LINT_VERSION := v1.54.2
GORELEASER_VERSION := v1.20.0

# Colors for terminal output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

.PHONY: all build build-dev run run-direct deps tidy test test-coverage benchmark fmt vet lint clean install uninstall package build-all check-gpg info help

# Default target
all: clean deps build

# Build targets
build: deps ## Build the application (production)
	@echo "$(BLUE)Building $(PROJECT_NAME) for $(GOOS)/$(GOARCH)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(PROJECT_NAME)
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

build-dev: ## Quick build for development
	@echo "$(BLUE)Building $(PROJECT_NAME) (development mode)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(PROJECT_NAME)
	@echo "$(GREEN)Development build completed: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

run: build ## Build and run the application
	@echo "$(BLUE)Running $(PROJECT_NAME)...$(RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

run-direct: ## Run without building (go run)
	@echo "$(BLUE)Running $(PROJECT_NAME) directly...$(RESET)"
	go run ./cmd/$(PROJECT_NAME) $(ARGS)

# Dependencies
deps: ## Install dependencies
	@echo "$(BLUE)Installing dependencies...$(RESET)"
	go mod download
	go mod verify
	@echo "$(GREEN)Dependencies installed$(RESET)"

tidy: ## Tidy Go modules
	@echo "$(BLUE)Tidying modules...$(RESET)"
	go mod tidy
	@echo "$(GREEN)Modules tidied$(RESET)"

# Testing
test: ## Run tests
	@echo "$(BLUE)Running tests...$(RESET)"
	go test -v -race ./...
	@echo "$(GREEN)Tests completed$(RESET)"

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	go test -v -race -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "$(GREEN)Coverage report generated: $(BUILD_DIR)/coverage.html$(RESET)"

benchmark: ## Run benchmark tests
	@echo "$(BLUE)Running benchmarks...$(RESET)"
	go test -bench=. -benchmem ./...
	@echo "$(GREEN)Benchmarks completed$(RESET)"

# Code quality
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(RESET)"
	go fmt ./...
	goimports -w .
	@echo "$(GREEN)Code formatted$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	go vet ./...
	@echo "$(GREEN)go vet completed$(RESET)"

lint: ## Run golangci-lint (requires installation)
	@echo "$(BLUE)Running golangci-lint...$(RESET)"
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not found. Installing...$(RESET)" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))
	golangci-lint run
	@echo "$(GREEN)Linting completed$(RESET)"

# Cleanup
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	go clean -cache -testcache -modcache
	@echo "$(GREEN)Cleanup completed$(RESET)"

# Installation
install: build ## Install to system PATH (override PREFIX or DESTDIR as needed)
	@echo "$(BLUE)Installing $(BINARY_NAME) to $(DESTDIR)$(BINDIR)...$(RESET)"
	@mkdir -p $(DESTDIR)$(BINDIR)
	@$(INSTALL) -m 0755 $(BUILD_DIR)/$(BINARY_NAME) $(DESTDIR)$(BINDIR)/$(BINARY_NAME)
	@echo "$(GREEN)$(BINARY_NAME) installed to $(DESTDIR)$(BINDIR)$(RESET)"

uninstall: ## Remove from system PATH (honors PREFIX/DESTDIR)
	@echo "$(BLUE)Removing $(BINARY_NAME) from $(DESTDIR)$(BINDIR)...$(RESET)"
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY_NAME)
	@echo "$(GREEN)$(BINARY_NAME) uninstalled from $(DESTDIR)$(BINDIR)$(RESET)"

# Packaging
package: build ## Create distribution package
	@echo "$(BLUE)Creating distribution package...$(RESET)"
	@mkdir -p $(DIST_DIR)
	tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz -C $(BUILD_DIR) $(BINARY_NAME)
	@echo "$(GREEN)Package created: $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz$(RESET)"

# Cross-compilation
build-all: clean deps ## Cross-compile for multiple platforms
	@echo "$(BLUE)Cross-compiling for multiple platforms...$(RESET)"
	@mkdir -p $(DIST_DIR)

	# Linux AMD64
	@echo "$(YELLOW)Building for linux/amd64...$(RESET)"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/$(PROJECT_NAME)

	# Linux ARM64
	@echo "$(YELLOW)Building for linux/arm64...$(RESET)"
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/$(PROJECT_NAME)

	# macOS AMD64
	@echo "$(YELLOW)Building for darwin/amd64...$(RESET)"
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/$(PROJECT_NAME)

	# macOS ARM64
	@echo "$(YELLOW)Building for darwin/arm64...$(RESET)"
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/$(PROJECT_NAME)

	# Windows AMD64
	@echo "$(YELLOW)Building for windows/amd64...$(RESET)"
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(PROJECT_NAME)

	# Windows ARM64
	@echo "$(YELLOW)Building for windows/arm64...$(RESET)"
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd/$(PROJECT_NAME)

	@echo "$(GREEN)Cross-compilation completed. Binaries in $(DIST_DIR)/$(RESET)"

# Utilities
check-gpg: ## Check GPG availability
	@echo "$(BLUE)Checking GPG configuration...$(RESET)"
	@gpg --version > /dev/null 2>&1 && echo "$(GREEN)GPG available$(RESET)" || echo "$(YELLOW)GPG not available$(RESET)"

info: ## Show project information
	@echo "$(BLUE)Project Information:$(RESET)"
	@echo "  Name:          $(PROJECT_NAME)"
	@echo "  Binary:        $(BINARY_NAME)"
	@echo "  Version:       $(VERSION)"
	@echo "  Build Time:    $(BUILD_TIME)"
	@echo "  Commit Hash:   $(COMMIT_HASH)"
	@echo "  Go Version:    $(GO_VERSION)"
	@echo "  GOOS/GOARCH:   $(GOOS)/$(GOARCH)"
	@echo "  Build Dir:     $(BUILD_DIR)"
	@echo "  Dist Dir:      $(DIST_DIR)"

help: ## Show this help message
	@echo "$(BLUE)Available targets:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(RESET) - %s\n", $$1, $$2}' $(MAKEFILE_LIST)
