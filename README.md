# go-tneat

**Transmission neat** - A CLI tool to check local directories against Transmission BitTorrent client torrents.

go-tneat helps you identify missing files by comparing what's in your local directories with what exists in your Transmission BitTorrent client. It's perfect for ensuring your downloaded content is complete and finding any missing items.

## Features

- 🔍 **Directory Check**: Compare local directories with Transmission torrents
- 📁 **List Download Directories**: View all download directories from Transmission
- 📄 **List Torrents**: Display all torrent paths from Transmission
- 📊 **Summary Reports**: Get detailed statistics on found/missing items
- 💾 **Missing Items Export**: Save missing item paths to a file
- 🎨 **Colorful Output**: Clear visual indicators for found/missing items
- 🔧 **Configurable**: Custom host, port, username, and password support

## Installation

### From Binary

Download the latest release from [GitHub Releases](https://github.com/your-username/go-tneat/releases):

```bash
# Linux
curl -L -o go-tneat https://github.com/your-username/go-tneat/releases/latest/download/go-tneat-linux-amd64
chmod +x go-tneat
sudo mv go-tneat /usr/local/bin/

# macOS
curl -L -o go-tneat https://github.com/your-username/go-tneat/releases/latest/download/go-tneat-darwin-amd64
chmod +x go-tneat
sudo mv go-tneat /usr/local/bin/

# Windows
curl -L -o go-tneat.exe https://github.com/your-username/go-tneat/releases/latest/download/go-tneat-windows-amd64.exe
```

### From Source

```bash
git clone https://github.com/your-username/go-tneat.git
cd go-tneat
go build -o go-tneat
```

## Usage

### Basic Commands

```bash
# Check current directory against Transmission
./go-tneat

# Check specific directories
./go-tneat check --dir /path/to/downloads --dir /path/to/media

# List all download directories from Transmission
./go-tneat list-directories

# List all torrent paths
./go-tneat list-torrents
```

### Connection Options

```bash
# Connect to remote Transmission instance
./go-tneat --host 192.168.1.100 --port 9091 --username myuser --password mypass

# Enable verbose logging
./go-tneat --verbose check --dir /downloads

# Enable debug logging
./go-tneat --debug check --dir /downloads
```

### Export Missing Items

```bash
# Save missing item paths to a file
./go-tneat check --dir /downloads --output missing-items.txt
```

## Options

### Global Flags

- `--host, -H`: Transmission host (default: localhost)
- `--port, -po`: Transmission port (default: 9091)
- `--user, -u`: Transmission username
- `--password, -p`: Transmission password
- `--verbose, -v`: Enable verbose logging output
- `--debug, -d`: Enable debug logging output

### Check Command Options

- `--dir, -d`: Directory to check (can be specified multiple times)
- `--output, -o`: Output file for absolute paths of missing items

### Command Aliases

- `list-directories` → `ls-dirs`, `ld`
- `list-torrents` → `ls-torrents`, `lt`

## Examples

### Example 1: Basic Directory Check

```bash
$ ./go-tneat check --dir ~/Downloads

Found 15 torrents in Transmission
================================================================================
Directory: /home/user/Downloads
================================================================================
✓ movie1.mkv (Found in Transmission)
✓ tv-show-s01e01.mp4 (Found in Transmission)
✗ old-file.txt (Missing from Transmission)
✓ document.pdf (Found in Transmission)

Directory Summary: 3/4 items found in Transmission
Missing items total size: 2.3 MB
```

### Example 2: Multiple Directories with Export

```bash
$ ./go-tneat check \
  --dir ~/Downloads \
  --dir ~/Media \
  --output missing.txt \
  --verbose

[INFO] Connecting to Transmission host=localhost port=9091 user=true
[INFO] Retrieved torrents from Transmission count=15
[INFO] Starting directory check directories=[/home/user/Downloads /home/user/Media]

Found 15 torrents in Transmission

Directory: /home/user/Downloads
================================================================================
✓ movie1.mkv
✓ tv-show-s01e01.mp4
✗ incomplete-download/
Directory Summary: 2/3 items found in Transmission
Missing items total size: 1.2 GB

Directory: /home/user/Media
================================================================================
✓ music-album/
✓ video-tutorials/
Directory Summary: 2/2 items found in Transmission

Overall Summary: 4/5 items found in Transmission across 2 directories
Total missing items size: 1.2 GB

Wrote 2 missing item paths to: missing.txt
```

### Example 3: Remote Transmission Server

```bash
$ ./go-tneat \
  --host nas.local \
  --port 9091 \
  --username admin \
  --password secret123 \
  check --dir /volume1/downloads
```

## Configuration

### Environment Variables

You can set the following environment variables to avoid passing credentials on the command line:

- `TRANSMISSION_HOST`: Transmission host
- `TRANSMISSION_PORT`: Transmission port
- `TRANSMISSION_USER`: Transmission username
- `TRANSMISSION_PASSWORD`: Transmission password

```bash
export TRANSMISSION_HOST=nas.local
export TRANSMISSION_USER=admin
export TRANSMISSION_PASSWORD=secret123
./go-tneat check --dir ~/Downloads
```

### Transmission RPC Settings

Make sure your Transmission daemon allows RPC connections:

1. Enable RPC in Transmission preferences
2. Set the RPC port (default: 9091)
3. Configure authentication if needed
4. For remote access, bind to `0.0.0.0` instead of `127.0.0.1`

## Output Explanation

### Status Indicators

- **✓**: Item found in Transmission
- **✗**: Item missing from Transmission
- **📁**: Directory (for folder status)

### Size Formatting

Sizes are automatically formatted with appropriate units:
- Bytes: `256 B`
- Kilobytes: `1.2 KB`
- Megabytes: `15.3 MB`
- Gigabytes: `2.1 GB`
- Terabytes: `1.5 TB`

## Troubleshooting

### Connection Issues

```bash
# Test connection with verbose logging
./go-tneat --verbose list-torrents

# Common issues:
# 1. Transmission RPC not enabled
# 2. Wrong host/port combination
# 3. Authentication failure
# 4. Firewall blocking connection
```

### Permission Issues

```bash
# Ensure you have read access to directories
ls -la /path/to/directory

# Run with appropriate permissions
./go-tneat check --dir /protected/directory
```

### Debug Mode

For detailed troubleshooting:

```bash
./go-tneat --debug check --dir /path/to/directory
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Release Information

This project uses GoReleaser for automated releases. See [.goreleaser.yaml](.goreleaser.yaml) for build configuration.