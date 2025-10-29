#!/bin/bash

# Release script for peerless
# Increments semver tag and runs goreleaser

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [patch|minor|major]"
    echo ""
    echo "Arguments:"
    echo "  patch  Increment patch version (e.g., 0.1.6 -> 0.1.7)"
    echo "  minor  Increment minor version (e.g., 0.1.6 -> 0.2.0)"
    echo "  major  Increment major version (e.g., 0.1.6 -> 1.0.0)"
    echo ""
    echo "Examples:"
    echo "  $0 patch   # Bump to 0.1.7"
    echo "  $0 minor   # Bump to 0.2.0"
    echo "  $0 major   # Bump to 1.0.0"
    exit 1
}

# Function to validate semver
validate_semver() {
    local version=$1
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        return 1
    fi
    return 0
}

# Function to increment semver
increment_version() {
    local current_version=$1
    local increment_type=$2

    if ! validate_semver "$current_version"; then
        print_error "Invalid current version: $current_version"
        exit 1
    fi

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
        *)
            print_error "Invalid increment type: $increment_type"
            show_usage
            ;;
    esac

    echo "${major}.${minor}.${patch}"
}

# Function to check if git working directory is clean
check_git_status() {
    if ! git diff-index --quiet HEAD --; then
        print_error "Git working directory is not clean"
        print_error "Please commit or stash your changes before releasing"
        exit 1
    fi
}

# Function to check if on main branch
check_git_branch() {
    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [[ "$current_branch" != "main" && "$current_branch" != "master" ]]; then
        print_warning "You are not on main/master branch (current: $current_branch)"
        read -p "Do you want to continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Release cancelled"
            exit 0
        fi
    fi
}

# Function to check if goreleaser is available
check_goreleaser() {
    if ! command -v goreleaser &> /dev/null; then
        print_error "goreleaser is not installed or not in PATH"
        print_info "Install goreleaser: https://goreleaser.com/install/"
        exit 1
    fi
}

# Function to check GitHub token
check_github_token() {
    if [[ -z "$GITHUB_TOKEN" ]]; then
        print_warning "GITHUB_TOKEN environment variable is not set"
        print_warning "Git operations will use your default credentials"
        print_info "Set GITHUB_TOKEN for automated releases:"
        print_info "  export GITHUB_TOKEN=\"ghp_your_token_here\""
        echo
        read -p "Do you want to continue without GITHUB_TOKEN? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Release cancelled"
            exit 0
        fi
    else
        print_info "GITHUB_TOKEN is set (length: ${#GITHUB_TOKEN} characters)"
    fi
}

# Main script
main() {
    local increment_type=${1:-}

    # Check arguments
    if [[ -z "$increment_type" ]]; then
        show_usage
    fi

    if [[ ! "$increment_type" =~ ^(patch|minor|major)$ ]]; then
        print_error "Invalid increment type: $increment_type"
        show_usage
    fi

    print_info "Starting release process..."
    print_info "Increment type: $increment_type"

    # Pre-flight checks
    print_info "Running pre-flight checks..."
    check_git_status
    check_git_branch
    check_github_token
    check_goreleaser

    # Get current version
    local current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")
    print_info "Current version: $current_version"

    # Calculate new version
    local new_version=$(increment_version "$current_version" "$increment_type")
    print_info "New version: $new_version"

    # Confirm release
    echo
    print_warning "About to create release $new_version"
    print_warning "This will:"
    print_warning "  1. Create and push a git tag v$new_version"
    print_warning "  2. Run 'goreleaser release --clean'"
    print_warning "  3. Upload release to GitHub"
    echo
    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Release cancelled"
        exit 0
    fi

    # Configure git to use GitHub token for authentication if available
    local git_remote_url="origin"
    local remote_url=""
    if [[ -n "$GITHUB_TOKEN" ]]; then
        print_info "Using GitHub token for authentication"
        # Extract the remote URL and modify it to include the token
        remote_url=$(git remote get-url "$git_remote_url" 2>/dev/null)
        if [[ "$remote_url" =~ ^https://github.com/ ]]; then
            # Convert HTTPS URL to include token
            local token_url="${remote_url/https:\/\//https:\/\/x-access-token:$GITHUB_TOKEN@}"
            git remote set-url "$git_remote_url" "$token_url"
            print_info "Updated remote URL to use token authentication"
        fi
    fi

    # Ensure we're up to date with remote
    print_info "Fetching latest changes from remote..."
    git fetch "$git_remote_url"

    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    if ! git merge "$git_remote_url"/"$current_branch" --ff-only; then
        print_error "Failed to merge changes from $git_remote_url/$current_branch"
        print_error "Please resolve conflicts and try again"
        exit 1
    fi

    # Create tag
    print_info "Creating tag v$new_version..."
    git tag -a "v$new_version" -m "Release v$new_version"

    # Push tag
    print_info "Pushing tag to remote..."
    if git push "$git_remote_url" "v$new_version"; then
        print_success "Tag v$new_version pushed successfully"
    else
        print_error "Failed to push tag v$new_version"
        # Restore original remote URL if we modified it
        if [[ -n "$GITHUB_TOKEN" && -n "$remote_url" ]]; then
            git remote set-url "$git_remote_url" "$remote_url" 2>/dev/null || true
        fi
        exit 1
    fi

    # Restore original remote URL if we modified it
    if [[ -n "$GITHUB_TOKEN" && -n "$remote_url" ]]; then
        git remote set-url "$git_remote_url" "$remote_url"
        print_info "Restored original remote URL"
    fi

    # Run goreleaser
    print_info "Running goreleaser..."
    if goreleaser release --clean; then
        print_success "Release v$new_version completed successfully!"
        print_success "Check GitHub releases for the uploaded artifacts"
    else
        print_error "goreleaser failed!"
        print_error "The tag has been pushed, but the release failed"
        print_error "You may need to manually clean up the release"
        exit 1
    fi
}

# Run main function with all arguments
main "$@"