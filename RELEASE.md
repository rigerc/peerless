# Release Process

This document describes how to create releases for Peerless using the automated release script and Makefile.

## Prerequisites

1. **Clean Git Working Directory**: All changes must be committed
2. **Main Branch**: Releases should be made from the main/master branch
3. **Goreleaser**: Install goreleaser (https://goreleaser.com/install/)
4. **GitHub Token**: For automated releases, you need a GitHub token with `repo` scope

```bash
# Install goreleaser (macOS example)
brew install goreleaser

# Set GitHub token (environment variable) - optional but recommended
export GITHUB_TOKEN="ghp_your_github_token_here"

# Or pass token inline:
make release GITHUB_TOKEN="ghp_your_github_token_here"
```

### GitHub Token Benefits

- ✅ **Automated Authentication**: No interactive git credential prompts
- ✅ **CI/CD Friendly**: Works in automated environments
- ✅ **Secure**: Token-based authentication instead of SSH keys
- ✅ **Consistent**: Same authentication for both git operations and goreleaser

### Token Scopes Required

Your GitHub token needs the following scopes:
- `repo` - Full control of private repositories
- `workflow` (if using GitHub Actions) - Update GitHub Action workflows

## Release Methods

### Method 1: Using the Bash Script (Recommended)

The `release.sh` script provides an interactive release process with safety checks.

```bash
# Patch release (0.1.6 -> 0.1.7)
./release.sh patch

# Minor release (0.1.6 -> 0.2.0)
./release.sh minor

# Major release (0.1.6 -> 1.0.0)
./release.sh major
```

### Method 2: Using the Makefile

The Makefile provides convenient shortcuts for the same operations.

```bash
# Patch release
make release-patch

# Minor release
make release-minor

# Major release
make release-major

# Default patch release
make release
```

## Release Process Flow

When you run a release command, the following happens:

1. **Pre-flight Checks**:
   - Ensures git working directory is clean
   - Checks if you're on main/master branch (with warning if not)
   - Verifies goreleaser is installed

2. **Version Calculation**:
   - Gets the current version from git tags
   - Calculates the new version based on increment type
   - Shows the version change for confirmation

3. **Confirmation**:
   - Prompts for confirmation before proceeding
   - You can cancel at this point if needed

4. **Release Execution**:
   - Fetches latest changes from remote
   - Creates and pushes a git tag (e.g., `v0.1.7`)
   - Runs `goreleaser release --clean`
   - Uploads release artifacts to GitHub

## Version Examples

| Current | Command    | New Version |
|---------|------------|-------------|
| 0.1.6   | `patch`    | 0.1.7       |
| 0.1.6   | `minor`    | 0.2.0       |
| 0.1.6   | `major`    | 1.0.0       |

## Makefile Commands Reference

```bash
# Development
make build          # Build for current platform
make test           # Run tests
make lint           # Format and vet code
make dev            # Run linting and tests

# Release
make release-patch  # Create patch release
make release-minor  # Create minor release
make release-major  # Create major release

# Utilities
make version        # Show current version
make clean          # Clean build artifacts
make check-dirty    # Check if working directory is dirty
make help           # Show all available targets
```

## Troubleshooting

### Working Directory Not Clean
```bash
# Commit your changes
git add .
git commit -m "Your commit message"
```

### Goreleaser Not Found
```bash
# Install goreleaser
curl -sSfL https://goreleaser.com/install.sh | sh

# Or using homebrew (macOS)
brew install goreleaser
```

### GitHub Token Issues
Make sure your GitHub token has the `repo` scope and is set in your environment:
```bash
export GITHUB_TOKEN="ghp_your_token_here"
```

### Release Failed but Tag Was Pushed
If goreleaser fails after the tag is pushed, you may need to:
1. Delete the failed GitHub release manually
2. Fix the issue (usually goreleaser configuration)
3. Delete the local tag and try again:
   ```bash
   git tag -d v0.1.7
   git push origin :refs/tags/v0.1.7
   ./release.sh patch
   ```

## Manual Release Process

If you prefer to do releases manually without the script:

1. **Update version** (if needed)
2. **Commit and push changes**
3. **Create tag**:
   ```bash
   git tag -a v0.1.7 -m "Release v0.1.7"
   git push origin v0.1.7
   ```
4. **Run goreleaser**:
   ```bash
   goreleaser release --clean
   ```

## Best Practices

1. **Test before releasing**: Run `make test` and `make lint` first
2. **Check changelog**: Update CHANGELOG.md with the new version changes
3. **Release from main**: Always release from the main/master branch
4. **Verify release**: Check GitHub releases after publishing to ensure artifacts are correct

## Automation

For CI/CD automation, you can use the Makefile targets:

```yaml
# Example GitHub Actions step
- name: Create Release
  if: startsWith(github.ref, 'refs/tags/')
  run: make release
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```