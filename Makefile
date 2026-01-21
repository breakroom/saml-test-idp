.PHONY: build test run clean fmt lint help generate-certs

# Binary name
BINARY_NAME=saml-test-idp
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/saml-test-idp

## test: Run unit tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## run: Build and run with example config
run: build
	@echo "Starting SAML IDP server..."
	./$(BUILD_DIR)/$(BINARY_NAME) --config config.example.yaml

## run-dev: Run with go run (no build step)
run-dev:
	$(GOCMD) run ./cmd/saml-test-idp --config config.example.yaml

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## tidy: Tidy go modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

## generate-certs: Generate self-signed certificates for testing
generate-certs:
	@echo "Generating self-signed certificates..."
	@mkdir -p certs
	openssl req -x509 -newkey rsa:2048 -keyout certs/idp.key -out certs/idp.crt -days 365 -nodes -subj "/CN=SAML Test IDP"
	@echo "Certificates generated in certs/ directory"

## help: Show this help message
help:
	@echo "SAML Test IDP - Makefile commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
