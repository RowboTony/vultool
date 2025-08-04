#!/bin/bash
# Build script for vultool on Linux/macOS
# This script reads the version from the VERSION file and builds the binary

set -e

# Default output binary name
OUTPUT="${1:-vultool}"

# Check if VERSION file exists
if [ ! -f "VERSION" ]; then
    echo "ERROR: VERSION file not found in current directory" >&2
    exit 1
fi

# Read version from file
VERSION=$(cat VERSION)
VERSION="${VERSION//[$'\r\n']/}" # Remove any newlines/carriage returns

echo "Building vultool version: $VERSION"

# Build the binary
echo "Running: go build -ldflags \"-X main.version=$VERSION\" -o \"$OUTPUT\" ./cmd/vultool"
go build -ldflags "-X main.version=$VERSION" -o "$OUTPUT" ./cmd/vultool

if [ $? -eq 0 ]; then
    echo "Build successful! Binary created: $OUTPUT"
    
    # Test the binary
    echo ""
    echo "Testing binary..."
    ./"$OUTPUT" --version
else
    echo "Build failed" >&2
    exit 1
fi
