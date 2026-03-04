#!/usr/bin/env pwsh
# Build script for Currency Converter API
# Usage: .\scripts\build.ps1

param(
    [switch]$Help
)

$ErrorActionPreference = "Stop"

function Show-Help {
    Write-Host @"
Currency Converter API - Build Script

Usage: .\scripts\build.ps1 [command]

Commands:
    build       Build the application (default)
    run         Run the application
    test        Run tests
    test-verbose    Run tests with verbose output
    test-cover      Run tests with coverage report
    clean       Clean build artifacts
    deps        Download dependencies
    docs        Generate Swagger documentation
    docs-clean  Remove Swagger documentation
    fmt         Format code
    help        Show this help message

Examples:
    .\scripts\build.ps1 build
    .\scripts\build.ps1 test
    .\scripts\build.ps1 run
"@
}

function Invoke-Build {
    Write-Host "Building currency_converter..." -ForegroundColor Cyan
    go build -o currency_converter.exe ./cmd/api
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build successful!" -ForegroundColor Green
    }
}

function Invoke-Run {
    Write-Host "Running currency_converter..." -ForegroundColor Cyan
    go run ./cmd/api/main.go
}

function Invoke-Test {
    Write-Host "Running tests..." -ForegroundColor Cyan
    go test ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "All tests passed!" -ForegroundColor Green
    }
}

function Invoke-TestVerbose {
    Write-Host "Running tests (verbose)..." -ForegroundColor Cyan
    go test -v ./...
}

function Invoke-TestCover {
    Write-Host "Running tests with coverage..." -ForegroundColor Cyan
    go test ./... -coverprofile=coverage.out
    if ($LASTEXITCODE -eq 0) {
        go tool cover -html=coverage.out -o coverage.html
        Remove-Item coverage.out
        Write-Host "Coverage report generated: coverage.html" -ForegroundColor Green
    }
}

function Invoke-Clean {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Cyan
    if (Test-Path "currency_converter.exe") {
        Remove-Item "currency_converter.exe"
    }
    go clean
    Write-Host "Clean complete!" -ForegroundColor Green
}

function Invoke-Deps {
    Write-Host "Downloading dependencies..." -ForegroundColor Cyan
    go mod download
    go mod tidy
    Write-Host "Dependencies updated!" -ForegroundColor Green
}

function Invoke-Docs {
    Write-Host "Generating Swagger documentation..." -ForegroundColor Cyan
    swag init -g cmd/api/main.go -o internal/handler/docs
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Documentation generated!" -ForegroundColor Green
    }
}

function Invoke-DocsClean {
    Write-Host "Removing Swagger documentation..." -ForegroundColor Cyan
    if (Test-Path "internal/handler/docs") {
        Remove-Item -Recurse -Force "internal/handler/docs"
    }
    Write-Host "Documentation removed!" -ForegroundColor Green
}

function Invoke-Fmt {
    Write-Host "Formatting code..." -ForegroundColor Cyan
    go fmt ./...
    Write-Host "Code formatted!" -ForegroundColor Green
}

# Main script logic
$command = $args[0]

if ($Help -or $command -eq "help") {
    Show-Help
    exit 0
}

switch ($command) {
    "build" { Invoke-Build }
    "run" { Invoke-Run }
    "test" { Invoke-Test }
    "test-verbose" { Invoke-TestVerbose }
    "test-cover" { Invoke-TestCover }
    "clean" { Invoke-Clean }
    "deps" { Invoke-Deps }
    "docs" { Invoke-Docs }
    "docs-clean" { Invoke-DocsClean }
    "fmt" { Invoke-Fmt }
    "" { Invoke-Build }
    default {
        Write-Host "Unknown command: $command" -ForegroundColor Red
        Write-Host "Use '.\scripts\build.ps1 help' for usage information." -ForegroundColor Yellow
        exit 1
    }
}
