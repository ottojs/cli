#!/usr/bin/env bash
set -e

# Remove and recreate bin directory
rm -rf bin
mkdir -p bin

# Build for x86_64
echo "Building x86_64..."
GOARCH=amd64 go build -o bin/otto-x86_64 ./cmd/otto

echo "Building for arm64..."
GOARCH=arm64 go build -o bin/otto-arm64 ./cmd/otto

echo ""
echo "Build complete! Binaries are in the 'bin' directory:"
ls -la bin/
