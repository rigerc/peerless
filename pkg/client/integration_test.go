package client_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"peerless/pkg/client"
	"peerless/pkg/types"
)

// TestTransmissionClientIntegration tests the Transmission client with mock server
func TestTransmissionClientIntegration(t *testing.T) {
	tests := []struct {
		name           string
		authRequired   bool
		username       string
		password       string
		expectSuccess  bool
		expectedError  string
		expectedCount  int
	}{
		{
			name:          "Valid authentication",
			authRequired:  true,
			username:      "admin",
			password:      "secret",
			expectSuccess: true,
			expectedCount: 3,
		},
		{
			name:          "Invalid authentication",
			authRequired:  true,
			username:      "admin",
			password:      "wrong",     // What client sends
			expectSuccess: false,
			expectedError: "authentication failed",
		},
		{
			name:          "No authentication required",
			authRequired:  false,
			username:      "",
			password:      "",
			expectSuccess: true,
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Log the request for debugging
				t.Logf("Mock server received request: %s %s", r.Method, r.URL)
				t.Logf("Headers: %+v", r.Header)

				// Check authentication if required
				if tt.authRequired {
					username, password, ok := r.BasicAuth()
					t.Logf("Auth check: user='%s', pass='%s', ok=%v", username, password, ok)

					// For the invalid auth test, we expect the password to be wrong
					expectedPassword := tt.password
					if tt.name == "Invalid authentication" {
						expectedPassword = "correctpassword" // This won't match what client sends
					}

					if !ok || username != tt.username || password != expectedPassword {
						t.Logf("Authentication failed, returning 401. Expected user='%s', pass='%s', got user='%s', pass='%s'", tt.username, expectedPassword, username, password)
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
				}

				// Check for session ID - for integration tests, we don't simulate the 409 conflict flow
				sessionID := r.Header.Get("X-Transmission-Session-Id")
				t.Logf("Session ID: %s", sessionID)

				// Always return 200 OK for integration tests to simplify the flow
				w.Header().Set("X-Transmission-Session-Id", "test-session-id")
				w.WriteHeader(http.StatusOK)

				// Return mock torrent data for any POST request
				response := `{
					"arguments": {
						"torrents": [
							{"id": 1, "name": "Movie.2023.1080p.BluRay.x264", "downloadDir": "/downloads/movies", "hashString": "abc123def456"},
							{"id": 2, "name": "TV.Series.S01", "downloadDir": "/downloads/tv", "hashString": "def456ghi789"},
							{"id": 3, "name": "Documentary.2024", "downloadDir": "/downloads/movies", "hashString": "ghi789jkl012"}
						]
					},
					"result": "success"
				}`
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(response))
			}))
			defer server.Close()

			// Extract host and port from server URL
			url := server.URL
			// Parse URL to get host and port correctly
			hostPort := strings.TrimPrefix(url, "http://")
			parts := strings.Split(hostPort, ":")
			host := parts[0]
			port := 80 // default, but httptest.Server uses dynamic ports
			if len(parts) > 1 {
				_, err := fmt.Sscanf(parts[1], "%d", &port)
				if err != nil {
					port = 80
				}
			}

			// Create client config
			config := types.Config{
				Host:     host,
				Port:     port,
				User:     tt.username,
				Password: tt.password,
			}

			// Create client
			transmissionClient := client.NewTransmissionClient(config)

			// Test GetSessionID
			sessionID, err := transmissionClient.GetSessionIDLegacy(context.Background())
			if tt.expectSuccess {
				if err != nil {
					t.Errorf("GetSessionID failed: %v", err)
					return
				}
				if sessionID != "test-session-id" {
					t.Errorf("Expected session ID 'test-session-id', got '%s'", sessionID)
				}

				// Test GetTorrents
				torrents, err := transmissionClient.GetTorrents(context.Background())
				if err != nil {
					t.Errorf("GetTorrents failed: %v", err)
					return
				}
				if len(torrents) != tt.expectedCount {
					t.Errorf("Expected %d torrents, got %d", tt.expectedCount, len(torrents))
				}

				// Test GetDownloadDirectories
				dirs, err := transmissionClient.GetDownloadDirectories(context.Background())
				if err != nil {
					t.Errorf("GetDownloadDirectories failed: %v", err)
					return
				}
				if len(dirs) == 0 {
					t.Error("Expected at least one directory, got none")
				}

				// Test GetAllTorrentPaths
				paths, err := transmissionClient.GetAllTorrentPaths(context.Background())
				if err != nil {
					t.Errorf("GetAllTorrentPaths failed: %v", err)
					return
				}
				if len(paths) != tt.expectedCount {
					t.Errorf("Expected %d paths, got %d", tt.expectedCount, len(paths))
				}

			} else {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

// TestTransmissionClientHTTPErrors tests HTTP error handling
func TestTransmissionClientHTTPErrors(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedError  string
	}{
		{
			name:          "401 Unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: "authentication failed",
		},
		{
			name:          "403 Forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: "access forbidden",
		},
		{
			name:          "404 Not Found",
			statusCode:    http.StatusNotFound,
			expectedError: "RPC endpoint not found",
		},
		{
			name:          "500 Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			expectedError: "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that returns error status
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			// Extract host and port from server URL
			url := server.URL
			hostPort := strings.TrimPrefix(url, "http://")
			parts := strings.Split(hostPort, ":")
			host := parts[0]
			port := 80 // default
			if len(parts) > 1 {
				_, err := fmt.Sscanf(parts[1], "%d", &port)
				if err != nil {
					port = 80
				}
			}

			// Create client config
			config := types.Config{
				Host:     host,
				Port:     port,
				User:     "admin",
				Password: "secret",
			}

			// Create client
			transmissionClient := client.NewTransmissionClient(config)

			// Test GetSessionID
			_, err := transmissionClient.GetSessionIDLegacy(context.Background())
			if err == nil {
				t.Errorf("Expected error for status %d, but got none", tt.statusCode)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
			}
		})
	}
}