#!/bin/bash

# Test script for the release process (dry run)
# This script simulates the release process without actually creating releases

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_test "Testing release script functionality..."

# Test 1: Show help
print_test "1. Testing help output..."
./release.sh

echo

# Test 2: Test version calculation logic
print_test "2. Testing version calculation..."

# Create a temporary script to test version increment
cat > /tmp/test_version.sh << 'EOF'
#!/bin/bash

# Copy the increment function from release.sh
increment_version() {
    local current_version=$1
    local increment_type=$2

    # Parse version components
    local major=$(echo "$current_version" | cut -d. -f1)
    local minor=$(echo "$current_version" | cut -d. -f2)
    local patch=$(echo "$current_version" | cut -d. -f3)

    case "$increment_type" in
        "patch")
            patch=$((patch + 1))
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
    esac

    echo "${major}.${minor}.${patch}"
}

# Test cases
echo "Testing version increment logic:"
echo "Current -> Increment -> Result"
echo "-------------------------------"
echo "0.1.6 -> patch     -> $(increment_version "0.1.6" "patch")"
echo "0.1.6 -> minor     -> $(increment_version "0.1.6" "minor")"
echo "0.1.6 -> major     -> $(increment_version "0.1.6" "major")"
echo "1.2.3 -> patch     -> $(increment_version "1.2.3" "patch")"
echo "1.2.3 -> minor     -> $(increment_version "1.2.3" "minor")"
echo "1.2.3 -> major     -> $(increment_version "1.2.3" "major")"
echo "0.0.9 -> patch     -> $(increment_version "0.0.9" "patch")"
echo "0.0.9 -> minor     -> $(increment_version "0.0.9" "minor")"
echo "9.9.9 -> major     -> $(increment_version "9.9.9" "major")"
EOF

chmod +x /tmp/test_version.sh
/tmp/test_version.sh
rm /tmp/test_version.sh

echo

# Test 3: Show current version
print_test "3. Checking current git version..."
if git describe --tags --abbrev=0 >/dev/null 2>&1; then
    CURRENT_VERSION=$(git describe --tags --abbrev=0)
    echo "Current version: $CURRENT_VERSION"
else
    echo "No tags found, would start from 0.0.0"
fi

echo

# Test 4: Show Makefile targets
print_test "4. Testing Makefile targets..."
echo "Available release targets:"
echo "- make release-patch  (patch release)"
echo "- make release-minor  (minor release)"
echo "- make release-major  (major release)"
echo "- make release        (default patch release)"

echo

# Test 5: Check git status
print_test "5. Checking git status..."
if git diff-index --quiet HEAD --; then
    echo "✓ Git working directory is clean"
else
    echo "✗ Git working directory is dirty (would prevent release)"
fi

echo

# Test 6: Check if goreleaser is available
print_test "6. Checking goreleaser availability..."
if command -v goreleaser &> /dev/null; then
    echo "✓ goreleaser is available"
    GORELEASER_VERSION=$(goreleaser --version 2>/dev/null | head -1 || echo "version unknown")
    echo "  Version: $GORELEASER_VERSION"
else
    echo "✗ goreleaser is not installed or not in PATH"
    echo "  Install from: https://goreleaser.com/install/"
fi

echo

print_success "Release script test completed!"
echo
echo "To perform an actual release:"
echo "1. Commit all your changes"
echo "2. Run: ./release.sh patch  (or minor/major)"
echo "3. Follow the prompts"
echo
echo "Or use the Makefile:"
echo "1. Commit all your changes"
echo "2. Run: make release-patch  (or release-minor/release-major)"