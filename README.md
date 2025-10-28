# Peerless

A powerful Go CLI tool for managing Transmission torrents by comparing local directories with your torrent library.

## Overview

Peerless connects to your Transmission BitTorrent client and helps you identify which local files and directories are not tracked in torrents. Perfect for maintaining an organized torrent library and ensuring all your media is properly managed.

## ✨ Features

- **🔐 Secure Authentication**: Mandatory authentication for all connections
- **📁 Directory Analysis**: Compare local directories against torrent names
- **📊 Multiple Output Formats**: Styled console output or plain text files
- **🌍 Unicode Support**: Handles international file names and paths
- **🚀 High Performance**: Fast comparison with efficient algorithms
- **🛡️ Error Resilience**: Comprehensive error handling with actionable messages
- **🎨 Beautiful Output**: Color-coded terminal output with automatic formatting

## 🚀 Quick Start

### Installation

### Pre-built binaries

Binaries are available in release for linux, windows and os x, including a arm-7 one (Synology)

#### From Source
```bash
git clone <repository-url>
cd peerless
go build -o peerless main.go
```

#### Using GoReleaser
```bash
goreleaser build --clean
```

### Basic Usage

```bash
# Show help screen
./peerless

# Check current directory against torrents
./peerless --host localhost --user admin --password secret check

# List all Transmission directories
./peerless --host localhost --user admin --password secret list-directories

# List all torrent paths
./peerless --host localhost --user admin --password secret list-torrents
```

## 📖 Commands

### `check` (Default Command)

Compare local directories with Transmission torrents to find missing items.

```bash
./peerless --host localhost --user admin --password secret check --dir /downloads/movies --dir /downloads/tv
```

**Flags:**
- `--dir, -d`: Directory to check (can be specified multiple times)
- `--output, -o`: Save missing items to file (unstyled output)

**Example Output:**
```
Directory: /downloads/movies
--------------------------------------------------------------------------------
✗ [DIR] Movie.2023.1080p.BluRay.x264
✓ [DIR] Movie.Collection.2023
✗ [FILE] Documentary.2024.720p.WEB-DL.x264
--------------------------------------------------------------------------------
Directory Summary: 1/3 items found in Transmission
Missing items total size: 25.67 GB
```

### `list-directories` (Aliases: `ls-dirs`, `ld`)

List all download directories configured in Transmission with torrent counts.

```bash
./peerless --host localhost --user admin --password secret list-directories

# Save to file (plain text)
./peerless --host localhost --user admin --password secret list-directories --output directories.txt
```

**Output Example:**
```
Download Directories in Transmission (3 unique):
--------------------------------------------------------------------------------
/downloads/movies (15 torrents)
/downloads/tv (8 torrents)
/downloads/documentaries (3 torrents)
```

### `list-torrents` (Aliases: `ls-torrents`, `lt`)

List all torrent paths from Transmission.

```bash
./peerless --host localhost --user admin --password secret list-torrents

# Save to file (plain text)
./peerless --host localhost --user admin --password secret list-torrents --output torrents.txt
```

## ⚙️ Configuration

### Authentication (Required)

Peerless requires authentication for all operations:

```bash
./peerless --host <host> --user <username> --password <password> <command>
```

**Required Parameters:**
- `--host, -H`: Transmission server host
- `--user, -u`: Transmission username
- `--password, -p`: Transmission password

**Example:**
```bash
./peerless --host 192.168.1.100 --user admin --password secret check
```

### Optional Parameters

- `--port, --po`: Transmission port (default: 9091)
- `--verbose, -v`: Enable verbose output
- `--debug, -d`: Enable debug output

## 🎨 Output Modes

### Console Output (Default)

Styled terminal output with colors and formatting:

```
Directory: /downloads/movies
🔹 Found: 15/20 items
❌ Missing: 5 items (12.3 GB)
```

### File Output (Plain Text)

Unstyled output suitable for scripts and automation:

```bash
./peerless --host localhost --user admin --password secret check --output missing.txt
```

**File Content:**
```
/downloads/movies/Missing.Movie.2023
/downloads/movies/Another.Movie.2024
```

## 🔧 Transmission Setup

### Enable RPC Access

1. Open Transmission
2. Go to **Edit → Preferences** (or **Transmission → Preferences** on macOS)
3. Navigate to **Remote** or **Web** tab
4. **Enable "Allow remote access"**
5. Set **Username** and **Password**
6. Note the **RPC port** (default: 9091)
7. Click **OK**

### Firewall Configuration

Ensure your firewall allows traffic to the Transmission RPC port:
- **Default Port**: 9091
- **Protocol**: HTTP
- **Source**: Your IP address or range

## 🐛 Troubleshooting

### Common Issues

#### Authentication Failed
```
authentication failed: please check your username and password for Transmission at host:port
```
**Solution**: Verify your Transmission RPC username and password.

#### Connection Refused
```
cannot connect to Transmission at host:port. Please ensure:
1. Transmission is running
2. RPC interface is enabled
3. Host and port are correct
```
**Solution**: Check if Transmission is running and RPC is enabled.

#### RPC Not Found
```
Transmission RPC endpoint not found at host:port. Ensure Transmission is running and RPC is enabled
```
**Solution**: Enable RPC interface in Transmission settings.

#### Port Issues
```
invalid port 99999: port must be between 1 and 65535
```
**Solution**: Use a valid port number (1-65535).

### Advanced Troubleshooting

#### Enable Debug Logging
```bash
./peerless --host localhost --user admin --password secret --debug check
```

#### Test Connection
```bash
# Test basic connectivity
./peerless --host localhost --user admin --password secret list-directories

# Test with verbose output
./peerless --host localhost --user admin --password secret --verbose check
```

#### Verify Transmission Configuration
1. Check Transmission is running: Open the application
2. Verify RPC enabled: Preferences → Remote → "Allow remote access"
3. Confirm credentials: Note exact username and password
4. Check port: Default is 9091, verify if changed

## 📋 Examples

### Basic Workflow
```bash
# 1. Check what directories Transmission knows about
./peerless --host localhost --user admin --password secret list-directories

# 2. Compare your media directory
./peerless --host localhost --user admin --password secret check --dir /media/Movies

# 3. Save missing items for later review
./peerless --host localhost --user admin --password secret check --dir /media/Movies --output missing.txt

# 4. Get complete torrent list
./peerless --host localhost --user admin --password secret list-torrents --output all-torrents.txt
```

### Batch Operations
```bash
# Check multiple directories
./peerless --host localhost --user admin --password secret \
  check \
  --dir /downloads/movies \
  --dir /downloads/tv \
  --dir /downloads/documentaries

# Export all directories
./peerless --host localhost --user admin --password secret \
  list-directories --output directories.txt

# Export all torrent paths
./peerless --host localhost --user admin --password secret \
  list-torrents --output torrents.txt
```

### Remote Management
```bash
# Connect to remote Transmission server
./peerless --host 192.168.1.100 --user admin --password secret \
  check --dir /media/server/Movies

# Different port
./peerless --host localhost --port 9092 --user admin --password secret \
  list-directories
```

## 🏗️ Development

### Building from Source

```bash
# Clone repository
git clone <repository-url>
cd peerless

# Build binary
go build -o peerless main.go

# Run tests
go test -v ./...

# Build release version
goreleaser build --clean
```

### Project Structure

```
peerless/
├── main.go              # CLI entry point and command definitions
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── .goreleaser.yaml     # Release configuration
├── README.md            # This file
└── pkg/
    ├── client/          # Transmission RPC client
    │   ├── transmission.go
    │   └── integration_test.go
    ├── types/           # Data structures
    │   └── types.go
    ├── utils/           # File system utilities
    │   ├── files.go
    │   └── integration_test.go
    └── output/          # Terminal styling
        └── styles.go
```

### Dependencies

- **[urfave/cli/v3](https://github.com/urfave/cli)**: Modern CLI framework
- **[charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)**: Terminal styling
- **[charmbracelet/log](https://github.com/charmbracelet/log)**: Structured logging
- **[stretchr/testify](https://github.com/stretchr/testify)**: Testing utilities

## 🧪 Testing

### Run Tests
```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./pkg/client
go test -v ./pkg/utils

# Run integration tests
go test -v -run Integration ./pkg/client
```

### Test Coverage
```bash
# Run with coverage
go test -v -cover ./...
```

## 📝 Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass: `go test -v ./...`
6. Commit your changes: `git commit -m "Add amazing feature"`
7. Push to branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Add comprehensive tests for new features
- Update documentation for breaking changes
- Ensure backward compatibility when possible
- Keep code clean and well-documented

## 📄 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## 🤝 Acknowledgments

- [Transmission](https://transmissionbt.com/) - The excellent BitTorrent client
- [urfave/cli](https://github.com/urfave/cli) - Powerful CLI framework
- [Charmbracelet](https://charmbracelet.com/) - Beautiful terminal UI tools

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/your-repo/peerless/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-repo/peerless/discussions)
- **Documentation**: [Wiki](https://github.com/your-repo/peerless/wiki)

---

**Peerless** - Keep your torrent library organized and complete. 🚀