.PHONY: build run test clean docs help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Swagger parameters
SWAGCMD=swag

# Binary name
BINARY_NAME=currency_converter

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/api

run:
	$(GORUN) ./cmd/api/main.go

test:
	$(GOTEST) ./...

test-verbose:
	$(GOTEST) -v ./...

test-cover:
	$(GOTEST) ./... -coverprofile=coverage.out
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	rm coverage.out

clean:
	rm -f $(BINARY_NAME)
	$(GOCMD) clean

deps:
	$(GOMOD) download
	$(GOMOD) tidy

docs:
	$(SWAGCMD) init -g cmd/api/main.go -o internal/handler/docs

docs-clean:
	rm -rf internal/handler/docs

fmt:
	$(GOCMD) fmt ./...

lint:
	golangci-lint run

help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  test-cover   - Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  docs         - Generate Swagger documentation"
	@echo "  docs-clean   - Remove Swagger documentation"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  help         - Show this help message"
