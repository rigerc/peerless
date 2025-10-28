package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"
)

type TransmissionRequest struct {
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type TorrentInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DownloadDir string `json:"downloadDir"`
	HashString  string `json:"hashString"`
}

type TorrentFile struct {
	Name   string `json:"name"`
	Length int64  `json:"length"`
}

type TorrentDetailedInfo struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	DownloadDir string        `json:"downloadDir"`
	HashString  string        `json:"hashString"`
	Files       []TorrentFile `json:"files"`
}

type TransmissionResponse struct {
	Arguments struct {
		Torrents []TorrentInfo `json:"torrents"`
	} `json:"arguments"`
	Result string `json:"result"`
}

type TransmissionDetailedResponse struct {
	Arguments struct {
		Torrents []TorrentDetailedInfo `json:"torrents"`
	} `json:"arguments"`
	Result string `json:"result"`
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Dirs     []string
}

func getSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	if !info.IsDir() {
		return info.Size(), nil
	}

	// For directories, calculate total size recursively
	var totalSize int64
	err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

func writeMissingPaths(filename string, paths []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, path := range paths {
		_, err := file.WriteString(path + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	app := &cli.Command{
		Name:  "go-tneat",
		Usage: "Transmission neat - check local directories against Transmission torrents",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"h"},
				Value:   "localhost",
				Usage:   "Transmission host",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"po"},
				Value:   9091,
				Usage:   "Transmission port",
			},
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "Transmission username",
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "Transmission password",
			},
			&cli.StringSliceFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Usage:   "Directory to check (can be specified multiple times)",
			},
			&cli.BoolFlag{
				Name:    "get-directories",
				Aliases: []string{"gd"},
				Usage:   "List all download directories from Transmission",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file for absolute paths of missing items",
			},
			&cli.BoolFlag{
				Name:    "get-all-torrents",
				Aliases: []string{"ga"},
				Usage:   "Get absolute paths of all torrents in Transmission",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := Config{
				Host:     cmd.String("host"),
				Port:     cmd.Int("port"),
				User:     cmd.String("user"),
				Password: cmd.String("password"),
				Dirs:     cmd.StringSlice("dir"),
			}

			getDirs := cmd.Bool("get-directories")
			getAllTorrents := cmd.Bool("get-all-torrents")
			outputFile := cmd.String("output")

			// Get session ID
			sessionID, err := getSessionID(cfg)
			if err != nil {
				return fmt.Errorf("error getting session ID: %w", err)
			}

			// If get-directories flag is set, just list directories and exit
			if getDirs {
				err := listDownloadDirectories(cfg, sessionID)
				if err != nil {
					return fmt.Errorf("error listing directories: %w", err)
				}
				return nil
			}

			// If get-all-torrents flag is set, output all torrent paths and exit
			if getAllTorrents {
				paths, err := getAllTorrentPaths(cfg, sessionID)
				if err != nil {
					return fmt.Errorf("error getting all torrent paths: %w", err)
				}

				// Output each path on its own line
				for _, path := range paths {
					fmt.Println(path)
				}
				return nil
			}

			// If no directories specified, use current directory
			if len(cfg.Dirs) == 0 {
				cfg.Dirs = []string{"."}
			}

			// Get all torrents from Transmission
			torrents, err := getTorrents(cfg, sessionID)
			if err != nil {
				return fmt.Errorf("error getting torrents: %w", err)
			}

			// Create a map of torrent names for quick lookup
			torrentMap := make(map[string]bool)
			for _, t := range torrents {
				torrentMap[strings.ToLower(t.Name)] = true
			}

			fmt.Printf("Found %d torrents in Transmission\n\n", len(torrents))

			totalItems := 0
			totalFound := 0
			totalMissingSize := int64(0)
			var missingPaths []string

			// Check each directory
			for dirIdx, dir := range cfg.Dirs {
				if dirIdx > 0 {
					fmt.Println()
				}

				// List directory contents
				entries, err := os.ReadDir(dir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", dir, err)
					continue
				}

				fmt.Printf("Directory: %s\n", dir)
				fmt.Println(strings.Repeat("-", 80))

				found := 0
				missingSize := int64(0)

				for _, entry := range entries {
					name := entry.Name()
					inTransmission := torrentMap[strings.ToLower(name)]

					var entryType string
					if entry.IsDir() {
						entryType = "[DIR] "
					} else {
						entryType = "[FILE]"
					}

					status := "✗"
					if inTransmission {
						status = "✓"
						found++
					} else {
						// Get size for missing items
						fullPath := filepath.Join(dir, name)

						// Get absolute path
						absPath, err := filepath.Abs(fullPath)
						if err != nil {
							absPath = fullPath
						}
						missingPaths = append(missingPaths, absPath)

						size, err := getSize(fullPath)
						if err == nil {
							missingSize += size
						}
					}

					fmt.Printf("%s %s %s\n", status, entryType, name)
				}

				fmt.Println(strings.Repeat("-", 80))
				fmt.Printf("Directory Summary: %d/%d items found in Transmission\n", found, len(entries))
				if missingSize > 0 {
					fmt.Printf("Missing items total size: %s\n", formatSize(missingSize))
				}

				totalItems += len(entries)
				totalFound += found
				totalMissingSize += missingSize
			}

			// Overall summary if multiple directories
			if len(cfg.Dirs) > 1 {
				fmt.Println()
				fmt.Println(strings.Repeat("=", 80))
				fmt.Printf("Overall Summary: %d/%d items found in Transmission across %d directories\n",
					totalFound, totalItems, len(cfg.Dirs))
				if totalMissingSize > 0 {
					fmt.Printf("Total missing items size: %s\n", formatSize(totalMissingSize))
				}
			}

			// Write missing paths to output file if specified
			if outputFile != "" {
				err := writeMissingPaths(outputFile, missingPaths)
				if err != nil {
					return fmt.Errorf("error writing to output file: %w", err)
				}
				fmt.Printf("\nWrote %d missing item paths to: %s\n", len(missingPaths), outputFile)
			}

			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getSessionID(cfg Config) (string, error) {
	url := fmt.Sprintf("http://%s:%d/transmission/rpc", cfg.Host, cfg.Port)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return "", err
	}

	if cfg.User != "" {
		req.SetBasicAuth(cfg.User, cfg.Password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	sessionID := resp.Header.Get("X-Transmission-Session-Id")
	if sessionID == "" {
		return "", fmt.Errorf("no session ID received")
	}

	return sessionID, nil
}

func getTorrents(cfg Config, sessionID string) ([]TorrentInfo, error) {
	url := fmt.Sprintf("http://%s:%d/transmission/rpc", cfg.Host, cfg.Port)

	reqBody := TransmissionRequest{
		Method: "torrent-get",
		Arguments: map[string]interface{}{
			"fields": []string{"id", "name", "downloadDir"},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transmission-Session-Id", sessionID)
	
	if cfg.User != "" {
		req.SetBasicAuth(cfg.User, cfg.Password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TransmissionResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Result != "success" {
		return nil, fmt.Errorf("transmission returned: %s", result.Result)
	}

	return result.Arguments.Torrents, nil
}

func getAllTorrentPaths(cfg Config, sessionID string) ([]string, error) {
	url := fmt.Sprintf("http://%s:%d/transmission/rpc", cfg.Host, cfg.Port)

	reqBody := TransmissionRequest{
		Method: "torrent-get",
		Arguments: map[string]interface{}{
			"fields": []string{"id", "name", "downloadDir", "hashString"},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transmission-Session-Id", sessionID)

	if cfg.User != "" {
		req.SetBasicAuth(cfg.User, cfg.Password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TransmissionResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Result != "success" {
		return nil, fmt.Errorf("transmission returned: %s", result.Result)
	}

	var paths []string
	for _, torrent := range result.Arguments.Torrents {
		// Construct absolute path using the torrent name and download directory
		absPath := filepath.Join(torrent.DownloadDir, torrent.Name)
		paths = append(paths, absPath)
	}

	// Sort paths alphabetically
	sort.Strings(paths)

	return paths, nil
}

func listDownloadDirectories(cfg Config, sessionID string) error {
	torrents, err := getTorrents(cfg, sessionID)
	if err != nil {
		return err
	}

	// Collect unique download directories
	dirMap := make(map[string]int)
	for _, t := range torrents {
		dirMap[t.DownloadDir]++
	}

	// Convert to sorted slice
	type dirCount struct {
		path  string
		count int
	}
	
	var dirs []dirCount
	for path, count := range dirMap {
		dirs = append(dirs, dirCount{path: path, count: count})
	}

	// Sort by path
	for i := 0; i < len(dirs); i++ {
		for j := i + 1; j < len(dirs); j++ {
			if dirs[i].path > dirs[j].path {
				dirs[i], dirs[j] = dirs[j], dirs[i]
			}
		}
	}

	fmt.Printf("Download Directories in Transmission (%d unique):\n", len(dirs))
	fmt.Println(strings.Repeat("-", 80))
	
	for _, d := range dirs {
		fmt.Printf("%s (%d torrents)\n", d.path, d.count)
	}

	return nil
}