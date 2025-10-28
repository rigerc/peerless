package client

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"peerless/pkg/types"
)

func TestNewTransmissionClient(t *testing.T) {
	config := types.Config{
		Host:     "localhost",
		Port:     9091,
		User:     "admin",
		Password: "secret",
	}

	client := NewTransmissionClient(config)

	assert.Equal(t, config, client.config)
}

func TestGetSessionID(t *testing.T) {
	t.Run("successful session ID retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/transmission/rpc", r.URL.Path)
			// Note: The initial session ID request doesn't set Content-Type

			// Read and discard the request body
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)

			w.Header().Set("X-Transmission-Session-Id", "test-session-id-123")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Extract host and port from test server
		host, port := extractHostPort(server.URL)

		config := types.Config{
			Host: host,
			Port: port,
		}
		client := NewTransmissionClient(config)

		sessionID, err := client.GetSessionID()
		require.NoError(t, err)
		assert.Equal(t, "test-session-id-123", sessionID)
	})

	t.Run("with authentication", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "admin", username)
			assert.Equal(t, "secret", password)

			w.Header().Set("X-Transmission-Session-Id", "auth-session-id")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		host, port := extractHostPort(server.URL)

		config := types.Config{
			Host:     host,
			Port:     port,
			User:     "admin",
			Password: "secret",
		}
		client := NewTransmissionClient(config)

		sessionID, err := client.GetSessionID()
		require.NoError(t, err)
		assert.Equal(t, "auth-session-id", sessionID)
	})

	t.Run("missing session ID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		host, port := extractHostPort(server.URL)

		config := types.Config{
			Host: host,
			Port: port,
		}
		client := NewTransmissionClient(config)

		sessionID, err := client.GetSessionID()
		assert.Error(t, err)
		assert.Equal(t, "", sessionID)
		assert.Contains(t, err.Error(), "no session ID received")
	})

	t.Run("connection error", func(t *testing.T) {
		config := types.Config{
			Host: "non-existent-host",
			Port: 9999,
		}
		client := NewTransmissionClient(config)

		sessionID, err := client.GetSessionID()
		assert.Error(t, err)
		assert.Equal(t, "", sessionID)
	})
}

func TestGetTorrents(t *testing.T) {
	t.Run("successful torrent retrieval", func(t *testing.T) {
		sessionID := "test-session-id"

		mockResponse := `{
			"arguments": {
				"torrents": [
					{
						"id": 1,
						"name": "Test Torrent 1",
						"downloadDir": "/downloads",
						"hashString": "abc123"
					},
					{
						"id": 2,
						"name": "Test Torrent 2",
						"downloadDir": "/downloads/movies",
						"hashString": "def456"
					}
				]
			},
			"result": "success"
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "test-session-id", r.Header.Get("X-Transmission-Session-Id"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Verify request body contains correct fields
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), "torrent-get")
			assert.Contains(t, string(body), "id")
			assert.Contains(t, string(body), "name")
			assert.Contains(t, string(body), "downloadDir")
			assert.Contains(t, string(body), "hashString")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, mockResponse)
		}))
		defer server.Close()

		host, port := extractHostPort(server.URL)

		config := types.Config{
			Host: host,
			Port: port,
		}
		client := NewTransmissionClient(config)

		torrents, err := client.GetTorrents(sessionID)
		require.NoError(t, err)

		assert.Len(t, torrents, 2)
		assert.Equal(t, 1, torrents[0].ID)
		assert.Equal(t, "Test Torrent 1", torrents[0].Name)
		assert.Equal(t, "/downloads", torrents[0].DownloadDir)
		assert.Equal(t, "abc123", torrents[0].HashString)

		assert.Equal(t, 2, torrents[1].ID)
		assert.Equal(t, "Test Torrent 2", torrents[1].Name)
		assert.Equal(t, "/downloads/movies", torrents[1].DownloadDir)
		assert.Equal(t, "def456", torrents[1].HashString)
	})

	t.Run("transmission error response", func(t *testing.T) {
		sessionID := "test-session-id"

		mockResponse := `{
			"result": "error",
			"arguments": {}
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, mockResponse)
		}))
		defer server.Close()

		host, port := extractHostPort(server.URL)

		config := types.Config{
			Host: host,
			Port: port,
		}
		client := NewTransmissionClient(config)

		torrents, err := client.GetTorrents(sessionID)
		assert.Error(t, err)
		assert.Nil(t, torrents)
		assert.Contains(t, err.Error(), "transmission returned: error")
	})
}

func TestGetAllTorrentPaths(t *testing.T) {
	t.Run("successful path retrieval with sorting", func(t *testing.T) {
		sessionID := "test-session-id"

		// Return torrents in unsorted order to test sorting
		mockResponse := `{
			"arguments": {
				"torrents": [
					{
						"id": 2,
						"name": "Z Torrent",
						"downloadDir": "/downloads",
						"hashString": "def456"
					},
					{
						"id": 1,
						"name": "A Torrent",
						"downloadDir": "/downloads",
						"hashString": "abc123"
					},
					{
						"id": 3,
						"name": "M Torrent",
						"downloadDir": "/movies",
						"hashString": "ghi789"
					}
				]
			},
			"result": "success"
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, mockResponse)
		}))
		defer server.Close()

		host, port := extractHostPort(server.URL)

		config := types.Config{
			Host: host,
			Port: port,
		}
		client := NewTransmissionClient(config)

		paths, err := client.GetAllTorrentPaths(sessionID)
		require.NoError(t, err)

		// Verify paths are sorted alphabetically
		expected := []string{
			"/downloads/A Torrent",
			"/downloads/Z Torrent",
			"/movies/M Torrent",
		}

		assert.Equal(t, expected, paths)
	})
}

func TestListDownloadDirectories(t *testing.T) {
	t.Run("successful directory listing", func(t *testing.T) {
		sessionID := "test-session-id"

		mockResponse := `{
			"arguments": {
				"torrents": [
					{
						"id": 1,
						"name": "Torrent 1",
						"downloadDir": "/downloads/movies",
						"hashString": "abc123"
					},
					{
						"id": 2,
						"name": "Torrent 2",
						"downloadDir": "/downloads/movies",
						"hashString": "def456"
					},
					{
						"id": 3,
						"name": "Torrent 3",
						"downloadDir": "/downloads/tv",
						"hashString": "ghi789"
					}
				]
			},
			"result": "success"
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, mockResponse)
		}))
		defer server.Close()

		host, port := extractHostPort(server.URL)

		config := types.Config{
			Host: host,
			Port: port,
		}
		client := NewTransmissionClient(config)

		// Capture stdout to verify output
		// This test just verifies no error is returned
		err := client.ListDownloadDirectories(sessionID)
		require.NoError(t, err)
	})
}

// Helper function to extract host and port from test server URL
func extractHostPort(url string) (string, int) {
	// Remove protocol prefix
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	// Split host and port
	parts := strings.Split(url, ":")
	host := parts[0]

	// Default port if not specified
	port := 80
	if len(parts) > 1 {
		_, err := fmt.Sscanf(parts[1], "%d", &port)
		if err != nil {
			port = 80
		}
	}

	return host, port
}