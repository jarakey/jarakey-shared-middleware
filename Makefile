.PHONY: all test test-coverage clean install help

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty)"

all: clean test

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...
	@echo "✅ Tests completed"

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run tests with coverage and show summary
test-coverage-summary:
	@echo "🧪 Running tests with coverage summary..."
	@go test -v -cover ./...
	@echo "✅ Coverage summary completed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f coverage.out coverage.html
	@echo "✅ Cleaned"

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# Build the library (just compile check)
build:
	@echo "🔨 Building library..."
	@go build ./...
	@echo "✅ Library builds successfully"

# Lint the code
lint:
	@echo "🔍 Running linter..."
	@go vet ./...
	@echo "✅ Linting completed"

# Format the code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Development setup
dev-setup: install
	@echo "🚀 Development environment ready!"
	@echo "Available commands:"
	@echo "  make test                    - Run tests"
	@echo "  make test-coverage          - Run tests with coverage report"
	@echo "  make test-coverage-summary  - Run tests with coverage summary"
	@echo "  make build                  - Build library (compile check)"
	@echo "  make lint                   - Run linter"
	@echo "  make fmt                    - Format code"
	@echo "  make clean                  - Clean artifacts"

# Help
help:
	@echo "Available targets:"
	@echo "  all                    - Clean and test"
	@echo "  test                   - Run tests"
	@echo "  test-coverage          - Run tests with coverage report"
	@echo "  test-coverage-summary  - Run tests with coverage summary"
	@echo "  build                  - Build library (compile check)"
	@echo "  lint                   - Run linter"
	@echo "  fmt                    - Format code"
	@echo "  clean                  - Clean build artifacts"
	@echo "  install                - Install dependencies"
	@echo "  dev-setup              - Setup development environment"
	@echo "  help                   - Show this help message"