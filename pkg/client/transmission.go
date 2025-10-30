package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"sync"

	"peerless/pkg/constants"
	"peerless/pkg/errors"
	"peerless/pkg/types"
	"peerless/pkg/utils"
)

// HTTPClient interface for easier testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TransmissionClient manages interactions with Transmission RPC
type TransmissionClient struct {
	config      types.Config
	httpClient  HTTPClient
	sessionID   string
	sessionLock sync.RWMutex
}

func NewTransmissionClient(config types.Config) *TransmissionClient {
	return &TransmissionClient{
		config: config,
		httpClient: &http.Client{
			Timeout: constants.HTTPTimeout,
		},
	}
}

// NewTransmissionClientWithHTTPClient for testing with mock HTTP client
func NewTransmissionClientWithHTTPClient(config types.Config, httpClient HTTPClient) *TransmissionClient {
	return &TransmissionClient{
		config:     config,
		httpClient: httpClient,
	}
}

// baseURL returns the Transmission RPC endpoint URL
func (c *TransmissionClient) baseURL() string {
	return fmt.Sprintf("http://%s:%d/transmission/rpc", c.config.Host, c.config.Port)
}

// getSessionID retrieves the current session ID, or fetches a new one
func (c *TransmissionClient) getSessionID(ctx context.Context) (string, error) {
	c.sessionLock.RLock()
	if c.sessionID != "" {
		sessionID := c.sessionID
		c.sessionLock.RUnlock()
		return sessionID, nil
	}
	c.sessionLock.RUnlock()

	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()

	// Double-check after acquiring write lock
	if c.sessionID != "" {
		return c.sessionID, nil
	}

	sessionID, err := c.fetchSessionID(ctx)
	if err != nil {
		return "", err
	}

	c.sessionID = sessionID
	return sessionID, nil
}

// fetchSessionID fetches a new session ID from Transmission
func (c *TransmissionClient) fetchSessionID(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL(), bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if c.config.User != "" {
		req.SetBasicAuth(c.config.User, c.config.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.NewTransmissionError(0, c.config.Host, c.config.Port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != 409 {
		return "", errors.NewTransmissionError(resp.StatusCode, c.config.Host, c.config.Port, nil)
	}

	sessionID := resp.Header.Get("X-Transmission-Session-Id")
	if sessionID == "" {
		return "", fmt.Errorf("no session ID received from Transmission at %s:%d", c.config.Host, c.config.Port)
	}

	return sessionID, nil
}

// doRequest performs an authenticated request to Transmission
func (c *TransmissionClient) doRequest(ctx context.Context, reqBody types.TransmissionRequest) (*types.TransmissionResponse, error) {
	sessionID, err := c.getSessionID(ctx)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request to JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transmission-Session-Id", sessionID)

	if c.config.User != "" {
		req.SetBasicAuth(c.config.User, c.config.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewTransmissionError(0, c.config.Host, c.config.Port, err)
	}
	defer resp.Body.Close()

	// Handle session conflict - invalidate and retry once
	if resp.StatusCode == 409 {
		c.sessionLock.Lock()
		c.sessionID = ""
		c.sessionLock.Unlock()

		return c.doRequest(ctx, reqBody)
	}

	if resp.StatusCode >= 400 {
		return nil, errors.NewTransmissionError(resp.StatusCode, c.config.Host, c.config.Port, nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result types.TransmissionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if result.Result != "success" {
		return nil, fmt.Errorf("transmission returned: %s", result.Result)
	}

	return &result, nil
}

// GetTorrents retrieves all torrents from Transmission
func (c *TransmissionClient) GetTorrents(ctx context.Context) ([]types.TorrentInfo, error) {
	reqBody := types.TransmissionRequest{
		Method: "torrent-get",
		Arguments: map[string]interface{}{
			"fields": []string{"id", "name", "downloadDir", "hashString"},
		},
	}

	resp, err := c.doRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	return resp.Arguments.Torrents, nil
}

// GetAllTorrentPaths returns sorted list of all torrent paths
func (c *TransmissionClient) GetAllTorrentPaths(ctx context.Context) ([]string, error) {
	torrents, err := c.GetTorrents(ctx)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(torrents))
	for _, torrent := range torrents {
		absPath := filepath.Join(torrent.DownloadDir, torrent.Name)
		cleanPath := utils.SanitizeString(absPath)
		paths = append(paths, cleanPath)
	}

	sort.Strings(paths)
	return paths, nil
}

// GetDownloadDirectories returns download directories with torrent counts
func (c *TransmissionClient) GetDownloadDirectories(ctx context.Context) ([]utils.DirectoryInfo, error) {
	torrents, err := c.GetTorrents(ctx)
	if err != nil {
		return nil, err
	}

	dirMap := make(map[string]int)
	for _, t := range torrents {
		dirMap[t.DownloadDir]++
	}

	dirs := make([]utils.DirectoryInfo, 0, len(dirMap))
	for path, count := range dirMap {
		cleanPath := utils.SanitizeString(path)
		dirs = append(dirs, utils.DirectoryInfo{Path: cleanPath, Count: count})
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Path < dirs[j].Path
	})

	return dirs, nil
}

// Legacy methods for backward compatibility (deprecated)
func (c *TransmissionClient) GetSessionIDLegacy(ctx context.Context) (string, error) {
	return c.getSessionID(ctx)
}

func (c *TransmissionClient) GetTorrentsLegacy(ctx context.Context, sessionID string) ([]types.TorrentInfo, error) {
	return c.GetTorrents(ctx)
}

func (c *TransmissionClient) GetAllTorrentPathsLegacy(ctx context.Context, sessionID string) ([]string, error) {
	return c.GetAllTorrentPaths(ctx)
}

func (c *TransmissionClient) GetDownloadDirectoriesLegacy(ctx context.Context, sessionID string) ([]utils.DirectoryInfo, error) {
	return c.GetDownloadDirectories(ctx)
}
