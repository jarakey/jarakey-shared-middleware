.PHONY: all build test clean install migrate-tool

# Build configuration
BINARY_NAME=migrate-tool
BUILD_DIR=bin
MAIN_PATH=migrations/cmd/main.go

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty)"

all: clean build test

# Build the migration tool
build:
	@echo "🔨 Building migration tool..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Built $(BINARY_NAME) in $(BUILD_DIR)/"

# Build for multiple platforms
build-all: clean
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "✅ Built for Linux AMD64"
	
	# macOS AMD64
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "✅ Built for macOS AMD64"
	
	# macOS ARM64
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "✅ Built for macOS ARM64"
	
	# Windows AMD64
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ Built for Windows AMD64"

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

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Cleaned"

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# Build and install locally
install-local: build
	@echo "📦 Installing migration tool locally..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Migration tool installed to /usr/local/bin/$(BINARY_NAME)"

# Run the migration tool
migrate-tool: build
	@echo "🚀 Running migration tool..."
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Example usage
example-up:
	@echo "📖 Example: Run migrations up"
	@echo "make migrate-tool ARGS='-command=up -database=\"postgres://user:pass@localhost:5432/db?sslmode=disable\"'"

example-down:
	@echo "📖 Example: Rollback migrations"
	@echo "make migrate-tool ARGS='-command=down -database=\"postgres://user:pass@localhost:5432/db?sslmode=disable\"'"

example-status:
	@echo "📖 Example: Check migration status"
	@echo "make migrate-tool ARGS='-command=status -database=\"postgres://user:pass@localhost:5432/db?sslmode=disable\"'"

example-validate:
	@echo "📖 Example: Validate migration files"
	@echo "make migrate-tool ARGS='-command=validate -path=./infrastructure/scripts'"

# Development helpers
dev-setup: install
	@echo "🚀 Development environment ready!"
	@echo "Available commands:"
	@echo "  make build          - Build migration tool"
	@echo "  make test           - Run tests"
	@echo "  make migrate-tool   - Run with custom args"
	@echo "  make example-up     - Show up migration example"
	@echo "  make example-down   - Show down migration example"
	@echo "  make example-status - Show status example"
	@echo "  make example-validate - Show validation example"

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Clean, build, and test"
	@echo "  build         - Build migration tool"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install dependencies"
	@echo "  install-local - Install migration tool locally"
	@echo "  migrate-tool  - Run migration tool with custom args"
	@echo "  dev-setup     - Setup development environment"
	@echo "  help          - Show this help message"
