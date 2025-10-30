# Peerless

A Go CLI tool for comparing local directories with Transmission torrents to identify missing files and keep your torrent library organized.

## Quick Start

### Installation

```bash
# Build from source
go build -o peerless main.go
```

or

# Download pre-built binaries from [releases] (Linux x86, i386 and arm-7, Darwin x86, Windows x86, i386)

### Basic Usage

```bash
# Check current directory against torrents
./peerless --host localhost --user admin --password secret

# Show Transmission status
./peerless --host localhost --user admin --password secret status

# List all download directories
./peerless --host localhost --user admin --password secret list-directories
```

### Common Options

```bash
# Check specific directories
./peerless --host localhost --user admin --password secret check --dir /downloads/movies --dir /downloads/tv

# Export results to file
./peerless --host localhost --user admin --password secret check --output missing.txt

# Show compact status
./peerless --host localhost --user admin --password secret status --compact

# Connect to remote Transmission
./peerless --host 192.168.1.100 --port 9091 --user admin --password secret

# Enable verbose output
./peerless --host localhost --user admin --password secret --verbose check

# Enable debug logging
./peerless --host localhost --user admin --password secret --debug status

# Preview file deletion (safe dry run)
./peerless --host localhost --user admin --password secret check --dry-run

# List all torrent paths
./peerless --host localhost --user admin --password secret list-torrents
```

## Features

- **Directory Comparison**: Find local files/directories not tracked in Transmission torrents
- **Status Monitoring**: View Transmission statistics and session information
- **File Management**: Safely delete missing files with confirmation and dry-run support
- **Multiple Formats**: Styled console output or plain text file exports
- **Secure Authentication**: Mandatory authentication for all connections

## Commands

- `check` - Compare directories with torrents (default)
- `status` - Show Transmission statistics
- `list-directories` - List all download directories
- `list-torrents` - List all torrent paths

## Example Usage

```bash
# Check specific directories
./peerless --host localhost --user admin --password secret \
  check --dir /downloads/movies --dir /downloads/tv

# Export missing items to file
./peerless --host localhost --user admin --password secret \
  check --output missing.txt

# Show compact status
./peerless --host localhost --user admin --password secret \
  status --compact

# Preview file deletion (dry run)
./peerless --host localhost --user admin --password secret \
  check --dry-run

# Delete missing files (after review)
./peerless --host localhost --user admin --password secret \
  check --rm
```

## Authentication Required

All operations require Transmission credentials:
```bash
./peerless --host <host> --user <username> --password <password> <command>
```

## Development

```bash
# Run tests
go test -v ./...

# Build with goreleaser
goreleaser build --clean

# Install dependencies
go mod tidy
```

## Architecture

- **pkg/client/** - Transmission RPC client with session management
- **pkg/service/** - Business logic for torrent operations
- **pkg/types/** - Data structures and configuration validation
- **pkg/utils/** - File system utilities and batch operations
- **pkg/output/** - Styled terminal output
- **pkg/errors/** - Specialized error handling