#!/usr/bin/env bash
set -euo pipefail

# Set version (can be overridden by environment variable)
VERSION=${VERSION:-"mainline"}

# Get Go version
GO_VERSION=$(go version | awk '{print $3}')

# Build flags
LDFLAGS="-X main.Version=${VERSION} -X main.GoVersion=${GO_VERSION}"

# Remove and recreate bin directory
rm -rf bin
mkdir -p bin

# Build for x86_64
echo "Building x86_64 (version ${VERSION})..."
GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/otto-x86_64 ./cmd/otto

echo "Building for arm64 (version ${VERSION})..."
GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/otto-arm64 ./cmd/otto

echo ""
echo "Build complete! Binaries are in the 'bin' directory:"
ls -la bin/
echo ""
echo "Version: ${VERSION}"
echo "Go Version: ${GO_VERSION}"
