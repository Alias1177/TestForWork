# Application name
APP_NAME := usdt-rates-service
BINARY_NAME := app

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Docker parameters
DOCKER_BUILD := docker build
DOCKER_TAG := $(APP_NAME):latest

# Build directory
BUILD_DIR := ./bin

# Source directory
SRC_DIR := ./cmd/server

# Test directories
TEST_DIRS := ./tests/... ./internal/...

.PHONY: all build clean test docker-build run lint help tidy deps generate

# Default target
all: clean build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out $(TEST_DIRS)
	@echo "Tests completed"

# Run tests with coverage
test-coverage: test
	@echo "Generating coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@$(DOCKER_BUILD) -t $(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_TAG)"

# Run the application locally
run: build
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run with docker-compose
docker-run:
	@echo "Starting services with docker-compose..."
	@docker-compose up -d

# Stop docker-compose services
docker-stop:
	@echo "Stopping services..."
	@docker-compose down

# View docker-compose logs
docker-logs:
	@docker-compose logs -f

# Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...
	@echo "Linting completed"

# Tidy go modules
tidy:
	@echo "Tidying go modules..."
	@$(GOMOD) tidy
	@echo "Go modules tidied"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@echo "Dependencies downloaded"

# Generate protobuf files
generate:
	@echo "Generating protobuf files..."
	@which protoc > /dev/null || (echo "protoc not found, please install Protocol Buffers compiler" && exit 1)
	@which protoc-gen-go > /dev/null || (echo "protoc-gen-go not found, installing..." && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)
	@which protoc-gen-go-grpc > /dev/null || (echo "protoc-gen-go-grpc not found, installing..." && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest)
	@export PATH=$$PATH:$(shell go env GOPATH)/bin && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/rates/rates.proto
	@echo "Protobuf files generated"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...
	@echo "Code formatted"

# Database migrations (for development)
migrate-up:
	@echo "Running database migrations..."
	@docker-compose exec app ./app --database.host=postgres --help || echo "Application not running, start with 'make docker-run' first"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@docker-compose up -d postgres
	@sleep 5
	@USDT_DATABASE_HOST=localhost $(GOTEST) -v -tags=integration ./tests/...
	@docker-compose down

# Show help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run unit tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Start services with docker-compose"
	@echo "  docker-stop   - Stop docker-compose services"
	@echo "  docker-logs   - View docker-compose logs"
	@echo "  run           - Build and run the application locally"
	@echo "  lint          - Run linter"
	@echo "  tidy          - Tidy go modules"
	@echo "  deps          - Download dependencies"
	@echo "  generate      - Generate protobuf files"
	@echo "  install-tools - Install development tools"
	@echo "  fmt           - Format code"
	@echo "  help          - Show this help message" 