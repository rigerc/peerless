package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"peerless/pkg/types"
	"peerless/pkg/utils"
)

type TransmissionClient struct {
	config types.Config
}

func NewTransmissionClient(config types.Config) *TransmissionClient {
	return &TransmissionClient{
		config: config,
	}
}

func (c *TransmissionClient) GetSessionID() (string, error) {
	url := fmt.Sprintf("http://%s:%d/transmission/rpc", c.config.Host, c.config.Port)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return "", err
	}

	if c.config.User != "" {
		req.SetBasicAuth(c.config.User, c.config.Password)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for HTTP authentication errors
	switch resp.StatusCode {
	case 401:
		return "", fmt.Errorf("authentication failed: invalid username or password for Transmission at %s:%d", c.config.Host, c.config.Port)
	case 403:
		return "", fmt.Errorf("access forbidden: insufficient permissions to access Transmission at %s:%d", c.config.Host, c.config.Port)
	case 404:
		return "", fmt.Errorf("Transmission RPC endpoint not found at %s:%d. Ensure Transmission is running and RPC is enabled", c.config.Host, c.config.Port)
	case 409:
		// This is the normal session establishment flow - extract session ID from response
		sessionID := resp.Header.Get("X-Transmission-Session-Id")
		if sessionID == "" {
			return "", fmt.Errorf("session conflict response missing X-Transmission-Session-Id header from Transmission at %s:%d", c.config.Host, c.config.Port)
		}
		return sessionID, nil
	case 500:
		return "", fmt.Errorf("Transmission server error (500) at %s:%d. Check Transmission logs", c.config.Host, c.config.Port)
	}

	// Check for other HTTP errors
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d error from Transmission at %s:%d", resp.StatusCode, c.config.Host, c.config.Port)
	}

	// For successful responses (200 OK), extract session ID from header
	sessionID := resp.Header.Get("X-Transmission-Session-Id")
	if sessionID == "" {
		return "", fmt.Errorf("no session ID received from Transmission at %s:%d. Ensure RPC interface is enabled", c.config.Host, c.config.Port)
	}

	return sessionID, nil
}

func (c *TransmissionClient) GetTorrents(sessionID string) ([]types.TorrentInfo, error) {
	url := fmt.Sprintf("http://%s:%d/transmission/rpc", c.config.Host, c.config.Port)

	reqBody := types.TransmissionRequest{
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

	if c.config.User != "" {
		req.SetBasicAuth(c.config.User, c.config.Password)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for HTTP authentication errors
	switch resp.StatusCode {
	case 401:
		return nil, fmt.Errorf("authentication failed: invalid username or password for Transmission at %s:%d", c.config.Host, c.config.Port)
	case 403:
		return nil, fmt.Errorf("access forbidden: insufficient permissions to access Transmission at %s:%d", c.config.Host, c.config.Port)
	case 404:
		return nil, fmt.Errorf("Transmission RPC endpoint not found at %s:%d. Ensure Transmission is running and RPC is enabled", c.config.Host, c.config.Port)
	case 409:
		return nil, fmt.Errorf("session conflict: invalid session ID. Re-authentication required")
	case 500:
		return nil, fmt.Errorf("Transmission server error (500) at %s:%d. Check Transmission logs", c.config.Host, c.config.Port)
	}

	// Check for other HTTP errors
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d error from Transmission at %s:%d", resp.StatusCode, c.config.Host, c.config.Port)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result types.TransmissionResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Result != "success" {
		return nil, fmt.Errorf("transmission returned: %s", result.Result)
	}

	return result.Arguments.Torrents, nil
}

func (c *TransmissionClient) GetAllTorrentPaths(sessionID string) ([]string, error) {
	torrents, err := c.GetTorrents(sessionID)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, torrent := range torrents {
		absPath := filepath.Join(torrent.DownloadDir, torrent.Name)
		// Sanitize the path to remove any control characters
		cleanPath := utils.SanitizeString(absPath)
		paths = append(paths, cleanPath)
	}

	// Sort paths alphabetically
	sort.Strings(paths)

	return paths, nil
}

// GetDownloadDirectories returns download directories with their torrent counts
func (c *TransmissionClient) GetDownloadDirectories(sessionID string) ([]utils.DirectoryInfo, error) {
	torrents, err := c.GetTorrents(sessionID)
	if err != nil {
		return nil, err
	}

	// Collect unique download directories
	dirMap := make(map[string]int)
	for _, t := range torrents {
		dirMap[t.DownloadDir]++
	}

	// Convert to sorted slice
	var dirs []utils.DirectoryInfo
	for path, count := range dirMap {
		cleanPath := utils.SanitizeString(path)
		dirs = append(dirs, utils.DirectoryInfo{Path: cleanPath, Count: count})
	}

	// Sort by path using Go's built-in sort
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Path < dirs[j].Path
	})

	return dirs, nil
}

// ListDownloadDirectories prints download directories (for backward compatibility)
func (c *TransmissionClient) ListDownloadDirectories(sessionID string) error {
	dirs, err := c.GetDownloadDirectories(sessionID)
	if err != nil {
		return err
	}

	fmt.Printf("Download Directories in Transmission (%d unique):\n", len(dirs))
	fmt.Println(strings.Repeat("-", 80))

	for _, d := range dirs {
		fmt.Printf("%s (%d torrents)\n", d.Path, d.Count)
	}

	return nil
}