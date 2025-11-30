.PHONY: help build test test-verbose test-coverage lint fmt clean install-tools example

BINARY_NAME=mt-toon
GO=go
GOFLAGS=-v -race -timeout 30s

help:
	@echo "mt-toon - Production-Grade Toon API Response Handler"
	@echo ""
	@echo "Available targets:"
	@echo "  make build          - Build the project"
	@echo "  make test           - Run tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make install-tools  - Install development tools"
	@echo "  make example        - Run example"

build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) ./...

test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...

test-verbose:
	@echo "Running tests with verbose output..."
	$(GO) test $(GOFLAGS) -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || true

clean:
	@echo "Cleaning..."
	$(GO) clean
	rm -f coverage.out coverage.html

install-tools:
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest

example:
	@echo "Running example..."
	$(GO) run cmd/example/main.go

.DEFAULT_GOAL := help
