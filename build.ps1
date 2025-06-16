
# Set ErrorActionPreference to stop on errors (equivalent to set -e)
$ErrorActionPreference = "Stop"

# Set version (can be overridden by environment variable)
$Version = if ($env:VERSION) { $env:VERSION } else { "mainline" }

# Get Go version
$GoVersionOutput = go version
$GoVersion = ($GoVersionOutput -split ' ')[2]

# Build flags
$LDFlags = "-X main.Version=$Version -X main.GoVersion=$GoVersion"

# Remove and recreate bin directory
if (Test-Path "bin") {
    Remove-Item -Path "bin" -Recurse -Force
}
New-Item -ItemType Directory -Path "bin" | Out-Null

# Build for x86_64
Write-Host "Building x86_64 (version $Version)..."
$env:GOARCH = "amd64"
go build -ldflags $LDFlags -o bin/otto-x86_64.exe ./cmd/otto

# Build for arm64
Write-Host "Building for arm64 (version $Version)..."
$env:GOARCH = "arm64"
go build -ldflags $LDFlags -o bin/otto-arm64.exe ./cmd/otto

# Reset GOARCH
Remove-Item Env:GOARCH

# Display results
Write-Host ""
Write-Host "Build complete! Binaries are in the 'bin' directory:"
Get-ChildItem -Path "bin"

Write-Host ""
Write-Host "Version: $Version"
Write-Host "Go Version: $GoVersion"
