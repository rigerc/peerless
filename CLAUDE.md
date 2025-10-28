# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**go-tneat** is a Go CLI application that checks local directories against Transmission torrents. It connects to a Transmission BitTorrent client via its RPC API and compares local file/directory names with torrent names to identify missing items.

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

## Development Commands

### Building
```bash
go build -o go-tneat main.go
```

### Testing
```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./pkg/client
go test -v ./pkg/utils
go test -v ./pkg/types
```

### Release Building (with GoReleaser)
```bash
goreleaser build --clean
```

## Dependencies

Main dependencies from `go.mod`:
- `github.com/urfave/cli/v3`: CLI framework
- `github.com/stretchr/testify`: Testing utilities
- `github.com/charmbracelet/lipgloss`: Terminal styling
- `github.com/charmbracelet/log`: Structured logging

## Usage Examples

```bash
# Check current directory against torrents
./go-tneat

# Check specific directories
./go-tneat check --dir /path/to/movies --dir /path/to/tv

# Connect to remote Transmission with authentication
./go-tneat --host 192.168.1.100 --port 9091 --user admin --password secret

# Export missing items to file
./go-tneat --output missing.txt

# List all Transmission download directories
./go-tneat list-directories

# List all torrent paths
./go-tneat list-torrents

# Enable verbose/debug logging
./go-tneat --verbose
./go-tneat --debug
```

## Configuration

The application connects to Transmission via HTTP RPC API:
- Default host: `localhost`
- Default port: `9091`
- Optional basic authentication support
- Session ID management for request authentication

## Testing Notes

Tests use `httptest` for mocking Transmission RPC responses. No running Transmission instance is required for unit tests. Test files follow Go's standard `_test.go` naming convention.

## Code Style

- Standard Go formatting (`gofmt`)
- Package-based architecture with clear separation of concerns
- Error handling with explicit error returns
- Structured logging with different verbosity levels
- Color-aware terminal output with fallback for non-TTY environments