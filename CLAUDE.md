# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Peerless** is a Go CLI application that checks local directories against Transmission torrents. It connects to a Transmission BitTorrent client via its RPC API and compares local file/directory names with torrent names to identify missing items.

## Architecture

### Core Components

- **`main.go`**: CLI entry point using `urfave/cli/v3` with three main commands:
  - `check`: Compare local directories with Transmission torrents (default command)
  - `list-directories`: Show all download directories from Transmission
  - `list-torrents`: List all torrent paths from Transmission

- **`pkg/client/`**: Transmission RPC client implementation
  - `transmission.go`: Handles HTTP communication with Transmission's RPC API
  - Supports authentication and session management
  - Methods: `GetSessionID()`, `GetTorrents()`, `GetAllTorrentPaths()`, `ListDownloadDirectories()`

- **`pkg/types/`**: Data structures for Transmission API
  - `TransmissionRequest/Response`: RPC message formats
  - `TorrentInfo`: Torrent metadata (ID, name, downloadDir, hashString)
  - `Config`: Application configuration

- **`pkg/utils/`**: File system utilities
  - `GetSize()`: Calculate file/directory sizes recursively
  - `FormatSize()`: Human-readable size formatting
  - `WriteMissingPaths()`: Export missing file paths to file

- **`pkg/output/`**: Styled terminal output using Lipgloss
  - Color-coded status display (found/missing items)
  - Automatic color detection and terminal compatibility
  - Logger integration with configurable levels

- **`pkg/constants/`**: Application constants
  - Default ports, timeouts, file size units, display constants
  - Unicode control character definitions

## Development Commands

### Building
```bash
# Simple build
go build -o peerless main.go

# Clean build with goreleaser (includes cross-compilation)
goreleaser build --clean
```

### Testing
```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./pkg/client
go test -v ./pkg/utils
go test -v ./pkg/types

# Run with coverage
go test -v -cover ./...

# Run integration tests
go test -v -run Integration ./pkg/client
```

### Dependency Management
```bash
# Tidy dependencies
go mod tidy

# Generate code (if needed)
go generate ./...
```

## Dependencies

Main dependencies from `go.mod`:
- `github.com/urfave/cli/v3`: CLI framework
- `github.com/stretchr/testify`: Testing utilities
- `github.com/charmbracelet/lipgloss`: Terminal styling
- `github.com/charmbracelet/log`: Structured logging

## Usage Examples

**Important**: All operations require authentication with Transmission.

```bash
# Check current directory against torrents
./peerless --host localhost --user admin --password secret

# Check specific directories
./peerless --host localhost --user admin --password secret check --dir /path/to/movies --dir /path/to/tv

# Connect to remote Transmission with authentication
./peerless --host 192.168.1.100 --port 9091 --user admin --password secret

# Export missing items to file
./peerless --host localhost --user admin --password secret --output missing.txt

# List all Transmission download directories
./peerless --host localhost --user admin --password secret list-directories

# List all torrent paths
./peerless --host localhost --user admin --password secret list-torrents

# Enable verbose/debug logging
./peerless --host localhost --user admin --password secret --verbose
./peerless --host localhost --user admin --password secret --debug

# Delete missing files (DESTRUCTIVE - use with caution)
./peerless --host localhost --user admin --password secret check --rm

# Preview deletions (dry run)
./peerless --host localhost --user admin --password secret check --dry-run
```

## Configuration

The application connects to Transmission via HTTP RPC API:
- **Required parameters**: host, username, password (all mandatory)
- Default port: `9091` (configurable)
- Basic authentication support
- Session ID management for request authentication
- Port validation: must be between 1-65535

## Testing Notes

Tests use `httptest` for mocking Transmission RPC responses. No running Transmission instance is required for unit tests. Test files follow Go's standard `_test.go` naming convention. Integration tests are available for end-to-end testing.

## Code Style

- Standard Go formatting (`gofmt`)
- Package-based architecture with clear separation of concerns
- Error handling with explicit error returns and enhanced error messages
- Structured logging with different verbosity levels (error by default)
- Color-aware terminal output with fallback for non-TTY environments
- Comprehensive input validation for all required parameters

## Build Configuration

The project uses GoReleaser for cross-platform builds:
- Supports Linux, Windows, macOS (amd64, i386)
- ARM support for Linux (armv7)
- Automatic archive creation (tar.gz/zip)
- Configurable build matrix in `.goreleaser.yaml`