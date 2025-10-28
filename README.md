# Peerless

A Go CLI tool that checks local directories against Transmission torrents to identify missing files and directories.

## Overview

**Peerless** connects to a Transmission BitTorrent client via its RPC API and compares local file/directory names with torrent names. This helps you identify which local items are not tracked in your Transmission instance, making it easier to maintain an organized torrent library.

## Features

- **Directory Comparison**: Check local directories against Transmission torrents
- **Multiple Directory Support**: Analyze multiple directories in a single run
- **Authentication Support**: Connect to Transmission with username/password
- **Remote Connection**: Connect to Transmission running on any host
- **Missing Items Export**: Export missing file paths to a text file
- **Color-coded Output**: Visual feedback with terminal colors
- **Verbose Logging**: Configurable output levels (error, info, debug)
- **Cross-platform**: Built for Linux, Windows, and macOS

## Installation

### Pre-built binaries

Binaries are available in release for linux, windows and os x, including a arm-7 one (Synology)

### From Source

```bash
git clone <repository-url>
cd peerless
go build -o peerless main.go
```

### Using GoReleaser

```bash
goreleaser build --clean
```

Binaries will be available in the `dist/` directory.

## Usage

### Basic Usage

Show the help screen with all available commands:

```bash
./peerless
```

To check the current directory against Transmission torrents, use the `check` command:

```bash
./peerless --host localhost --user admin --password secret check
```

### Check Specific Directories

```bash
./peerless --host localhost --user admin --password secret check --dir /path/to/movies --dir /path/to/tv
```

### Connect to Transmission

**Authentication is required** for all commands:

```bash
./peerless --host 192.168.1.100 --port 9091 --user admin --password secret
```

**Required Parameters:**
- `--host, -H`: Transmission host address
- `--user, -u`: Transmission username
- `--password, -p`: Transmission password

### Export Missing Items

```bash
./peerless --host localhost --user admin --password secret --output missing-items.txt
```

### List Management Commands

List all download directories configured in Transmission:

```bash
./peerless --host localhost --user admin --password secret list-directories

# Save directories to file
./peerless --host localhost --user admin --password secret list-directories --output directories.txt
```

List all torrent paths from Transmission:

```bash
./peerless --host localhost --user admin --password secret list-torrents

# Save torrent paths to file
./peerless --host localhost --user admin --password secret list-torrents --output torrents.txt
```

### Verbosity Control

```bash
# Show info-level output
./peerless --host localhost --user admin --password secret --verbose

# Show debug-level output
./peerless --host localhost --user admin --password secret --debug
```

## Commands

### `check`

Compare local directories with Transmission torrents.

**Flags:**
- `--dir, -d`: Directory to check (can be specified multiple times)
- `--output, -o`: Output file for absolute paths of missing items

**Example:**
```bash
./peerless check --dir /downloads/movies --dir /downloads/tv --output missing.txt
```

### `list-directories` (aliases: `ls-dirs`, `ld`)

List all download directories from Transmission with torrent counts.

**Flags:**
- `--output, -o`: Output file for directory list

**Example:**
```bash
./peerless list-directories

# Save to file (unstyled output)
./peerless list-directories --output directories.txt
```

### `list-torrents` (aliases: `ls-torrents`, `lt`)

List all torrent paths from Transmission.

**Flags:**
- `--output, -o`: Output file for torrent paths

**Example:**
```bash
./peerless list-torrents

# Save to file (unstyled output)
./peerless list-torrents --output torrents.txt
```

## Global Flags

- `--host, -H`: Transmission host (required)
- `--port, -po`: Transmission port (default: 9091)
- `--user, -u`: Transmission username (required)
- `--password, -p`: Transmission password (required)
- `--verbose, -v`: Enable verbose logging output
- `--debug, -d`: Enable debug logging output

## Output Examples

### Check Command Output

```
Directory: /downloads/movies
--------------------------------------------------------------------------------
✗ [FILE] Movie.2023.1080p.BluRay.x264
✓ [DIR] Movie.Collection.2023
✗ [DIR] Another.Movie.2024
--------------------------------------------------------------------------------
Directory Summary: 1/3 items found in Transmission
Missing items total size: 25.67 GB
```

### List Directories Output

```
Download Directories in Transmission (2 unique):
--------------------------------------------------------------------------------
/downloads/movies (2 torrents)
/downloads/tv (1 torrents)
```

## Color Legend

- ✓ **Green**: Item found in Transmission
- ✗ **Red**: Item missing from Transmission
- **Blue**: Directory headers and information
- **Cyan**: File paths and summaries
- **Gray**: File sizes and separators

## Configuration

peerless connects to Transmission via its RPC API. Authentication is **required** - you must provide host, username, and password. The default port is 9091.

### Transmission RPC Settings

Ensure Transmission's RPC interface is enabled:

1. Open Transmission preferences
2. Go to "Remote" or "Web" tab
3. Enable "Allow remote access"
4. Set username/password if desired
5. Note the RPC port (default: 9091)

## Development

### Prerequisites

- Go 1.25.3 or later

### Building

```bash
go build -o peerless main.go
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

### Project Structure

```
peerless/
├── main.go              # CLI entry point
├── pkg/
│   ├── client/          # Transmission RPC client
│   ├── output/          # Terminal output styling
│   ├── types/           # Data structures
│   └── utils/           # File system utilities
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── .goreleaser.yaml     # Release configuration
└── README.md            # This file
```

## Dependencies

- `github.com/urfave/cli/v3` - CLI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/log` - Structured logging
- `github.com/stretchr/testify` - Testing utilities

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License.

## Troubleshooting

### Connection Issues

- Ensure Transmission is running and RPC is enabled
- Check that the host and port are correct
- Verify firewall settings allow traffic to the RPC port
- Confirm authentication credentials if required

### Performance

- Large directories may take time to process, especially when calculating sizes
- Use `--verbose` to see progress information
- Consider checking smaller directories individually for faster results

### Color Output

- Colors are automatically disabled in non-terminal environments
- Force monochrome output by piping to another command or redirecting to a file

## Similar Tools

- [transmission-remote](https://transmissionbt.com/) - Official Transmission CLI
- [transmission-cli](https://github.com/transmission/transmission) - Command-line interface

## Support

For issues, feature requests, or questions, please open an issue on the project repository.