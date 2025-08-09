#!/bin/bash

# Run tests for the notebook package

echo "Running notebook-use tests..."
echo "================================"

# Change to the notebook directory
cd "$(dirname "$0")"

# Run unit tests
echo -e "\n1. Running unit tests..."
go test -v -short ./...

# Run integration tests (requires Docker/Dagger)
echo -e "\n2. Running integration tests..."
if command -v docker &> /dev/null; then
    go test -v -run Integration ./...
else
    echo "Docker not available, skipping integration tests"
fi

# Run benchmarks
echo -e "\n3. Running benchmarks..."
go test -bench=. -benchmem ./...

# Generate coverage report
echo -e "\n4. Generating coverage report..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

echo -e "\nTest run complete!"
echo "Coverage report: coverage.html"