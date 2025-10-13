SHELL := /bin/bash
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

.DEFAULT_GOAL: all

# Build metadata
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION ?= $(shell go version | cut -d' ' -f3)

# Build configuration
BUILD_DIR ?= ./bin
TOOLS_DIR ?= $(BUILD_DIR)/tools
CGO_ENABLED ?= 0

# Enhanced LDFLAGS with version information
LDFLAGS = -ldflags "-s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.Commit=$(COMMIT)' \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.GoVersion=$(GO_VERSION)'"

# Tool versions
GOLANGCI_LINT_VERSION ?= v1.61.0
AIR_VERSION ?= v1.61.5

# Server configuration
SERVER_PID_FILE ?= $(BUILD_DIR)/server.pid
SERVER_PORT ?= 8090
SERVER_HOST ?= 127.0.0.1

.PHONY: all build check clean format help serve serve-bg serve-stop serve-status serve-restart test test-unit test-integration tidy deps-update deps-check deps-audit deps-outdated

all: check test-unit build ## Default target: check, test-unit, build

build: ## Build all executables, located under ./bin/
	@echo "Building..."
	@CGO_ENABLED=0 go build -o ./bin/example -trimpath $(LDFLAGS) ./example/...
	@CGO_ENABLED=0 go build -o ./bin/pocketbase -trimpath $(LDFLAGS) ./cmd/pocketbase/...

serve: build ## Run the pocketbase server
	@echo "Starting PocketBase server on $(SERVER_HOST):$(SERVER_PORT)..."
	@./bin/pocketbase serve --dev --http=$(SERVER_HOST):$(SERVER_PORT)

serve-bg: build ## Run the pocketbase server in background
	@echo "Starting PocketBase server in background..."
	@if [ -f $(SERVER_PID_FILE) ]; then \
		echo "Server already running (PID: $$(cat $(SERVER_PID_FILE)))"; \
		exit 1; \
	fi
	@./bin/pocketbase serve --dev --http=$(SERVER_HOST):$(SERVER_PORT) & \
	echo $$! > $(SERVER_PID_FILE) && \
	echo "Server started with PID: $$! (saved to $(SERVER_PID_FILE))"

serve-stop: ## Stop the background pocketbase server
	@if [ ! -f $(SERVER_PID_FILE) ]; then \
		echo "No server PID file found at $(SERVER_PID_FILE)"; \
		exit 1; \
	fi
	@PID=$$(cat $(SERVER_PID_FILE)) && \
	if kill -0 $$PID 2>/dev/null; then \
		echo "Stopping server (PID: $$PID)..."; \
		kill $$PID && \
		rm -f $(SERVER_PID_FILE) && \
		echo "Server stopped successfully"; \
	else \
		echo "Server process (PID: $$PID) not running, cleaning up PID file"; \
		rm -f $(SERVER_PID_FILE); \
	fi

serve-status: ## Check if the pocketbase server is running
	@if [ -f $(SERVER_PID_FILE) ]; then \
		PID=$$(cat $(SERVER_PID_FILE)) && \
		if kill -0 $$PID 2>/dev/null; then \
			echo "Server is running (PID: $$PID)"; \
		else \
			echo "Server PID file exists but process not running"; \
			rm -f $(SERVER_PID_FILE); \
		fi; \
	else \
		echo "Server is not running"; \
	fi

serve-restart: serve-stop serve-bg ## Restart the pocketbase server

clean: ## Remove all artifacts from ./bin/ and ./resources
	@echo "Cleaning build artifacts..."
	@if [ -f $(SERVER_PID_FILE) ]; then \
		echo "Stopping running server before cleanup..."; \
		$(MAKE) serve-stop; \
	fi
	@rm -rf ./bin/* ./resources/*
	@echo "Clean complete"

format: ## Format go code with goimports
	@go install golang.org/x/tools/cmd/goimports@latest
	@goimports -l -w .

test: ## Run tests (requires PocketBase server running on :8090)
	@go test -shuffle=on -race ./...

test-unit: ## Run unit tests only (short mode)
	@go test -shuffle=on -race -short ./...

test-integration: build ## Run integration tests with automatic server management
	@echo "Starting integration tests with automatic server management..."
	@$(MAKE) serve-bg
	@echo "Waiting for server to be ready..."
	@sleep 3
	@echo "Running integration tests..."
	@go test -shuffle=on -race ./... || TEST_RESULT=$$?; \
	echo "Stopping test server..."; \
	$(MAKE) serve-stop; \
	exit $${TEST_RESULT:-0}

tidy: ## Run go mod tidy
	@go mod tidy

check: ## Linting and static analysis
	@if test ! -e ./bin/golangci-lint; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh; \
	fi

	@./bin/golangci-lint run -c .golangci.yml

	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

deps-update: ## Update all dependencies to latest versions
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

deps-check: ## Verify dependencies and check for issues
	@echo "Verifying dependencies..."
	@go mod verify
	@go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod or go.sum has changes after tidy" && exit 1)

deps-audit: ## Audit dependencies for security vulnerabilities
	@echo "Auditing dependencies for vulnerabilities..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

deps-outdated: ## Check for outdated dependencies
	@echo "Checking for outdated dependencies..."
	@go list -u -m all | grep -v "indirect" | grep "\["

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
