package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type TransmissionRequest struct {
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type TorrentInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DownloadDir string `json:"downloadDir"`
}

type TransmissionResponse struct {
	Arguments struct {
		Torrents []TorrentInfo `json:"torrents"`
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

type dirList []string

func (d *dirList) String() string {
	return strings.Join(*d, ",")
}

func (d *dirList) Set(value string) error {
	*d = append(*d, value)
	return nil
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
	cfg := Config{}
	var dirs dirList
	var getDirs bool
	var outputFile string
	var missingOnly bool
	var foundOnly bool
	var checkPaths bool
	var orphanedTorrents bool
	
	flag.StringVar(&cfg.Host, "host", "localhost", "Transmission host")
	flag.IntVar(&cfg.Port, "port", 9091, "Transmission port")
	flag.StringVar(&cfg.User, "user", "", "Transmission username")
	flag.StringVar(&cfg.Password, "pass", "", "Transmission password")
	flag.Var(&dirs, "dir", "Directory to check (can be specified multiple times)")
	flag.BoolVar(&getDirs, "get-directories", false, "List all download directories from Transmission")
	flag.StringVar(&outputFile, "o", "", "Output file for absolute paths of missing items")
	flag.BoolVar(&missingOnly, "missing-only", false, "Show only items NOT in Transmission")
	flag.BoolVar(&foundOnly, "found-only", false, "Show only items found in Transmission")
	flag.BoolVar(&checkPaths, "check-paths", false, "Verify Transmission download paths exist on disk")
	flag.BoolVar(&orphanedTorrents, "orphaned-torrents", false, "Find torrents whose files don't exist on disk")
	flag.Parse()

	// Get session ID
	sessionID, err := getSessionID(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting session ID: %v\n", err)
		os.Exit(1)
	}

	// If get-directories flag is set, just list directories and exit
	if getDirs {
		err := listDownloadDirectories(cfg, sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing directories: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// If no directories specified, use current directory
	if len(dirs) == 0 {
		dirs = append(dirs, ".")
	}
	cfg.Dirs = dirs

	// Get session ID
	sessionID, err = getSessionID(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting session ID: %v\n", err)
		os.Exit(1)
	}

	// Get all torrents from Transmission
	torrents, err := getTorrents(cfg, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting torrents: %v\n", err)
		os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "\nError writing to output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nWrote %d missing item paths to: %s\n", len(missingPaths), outputFile)
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