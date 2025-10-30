package client

import (
	"context"
	"net/http"
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

	assert.NotNil(t, client.httpClient)
}

func TestNewTransmissionClientWithHTTPClient(t *testing.T) {
	config := types.Config{
		Host:     "localhost",
		Port:     9091,
		User:     "admin",
		Password: "secret",
	}

	mockHTTP := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return NewMockResponse(200, "{}", nil), nil
		},
	}

	client := NewTransmissionClientWithHTTPClient(config, mockHTTP)

	assert.Equal(t, mockHTTP, client.httpClient)
}

func TestGetSessionID(t *testing.T) {
	t.Run("successful session ID retrieval", func(t *testing.T) {
		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return NewMockResponse(409, "{}", map[string]string{
					"X-Transmission-Session-Id": "test-session-id-123",
				}), nil
			},
		}

		config := types.Config{
			Host: "localhost",
			Port: 9091,
		}
		client := NewTransmissionClientWithHTTPClient(config, mockHTTP)

		sessionID, err := client.getSessionID(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "test-session-id-123", sessionID)
	})

	t.Run("with authentication", func(t *testing.T) {
		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				username, password, ok := req.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, "admin", username)
				assert.Equal(t, "secret", password)

				return NewMockResponse(409, "{}", map[string]string{
					"X-Transmission-Session-Id": "auth-session-id",
				}), nil
			},
		}

		config := types.Config{
			Host:     "localhost",
			Port:     9091,
			User:     "admin",
			Password: "secret",
		}
		client := NewTransmissionClientWithHTTPClient(config, mockHTTP)

		sessionID, err := client.getSessionID(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "auth-session-id", sessionID)
	})

	t.Run("missing session ID", func(t *testing.T) {
		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return NewMockResponse(200, "{}", nil), nil
			},
		}

		config := types.Config{
			Host: "localhost",
			Port: 9091,
		}
		client := NewTransmissionClientWithHTTPClient(config, mockHTTP)

		sessionID, err := client.getSessionID(context.Background())
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

		sessionID, err := client.getSessionID(context.Background())
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

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// First call returns session ID
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": sessionID,
					}), nil
				}

				// Second call returns torrents
				assert.Equal(t, "test-session-id", req.Header.Get("X-Transmission-Session-Id"))
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{
			Host: "localhost",
			Port: 9091,
		}
		client := NewTransmissionClientWithHTTPClient(config, mockHTTP)

		torrents, err := client.GetTorrents(context.Background())
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

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// First call returns session ID
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": sessionID,
					}), nil
				}

				// Second call returns error
				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{
			Host: "localhost",
			Port: 9091,
		}
		client := NewTransmissionClientWithHTTPClient(config, mockHTTP)

		torrents, err := client.GetTorrents(context.Background())
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

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// First call returns session ID
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": sessionID,
					}), nil
				}

				// Second call returns torrents
				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{
			Host: "localhost",
			Port: 9091,
		}
		client := NewTransmissionClientWithHTTPClient(config, mockHTTP)

		paths, err := client.GetAllTorrentPaths(context.Background())
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

func TestGetDownloadDirectories(t *testing.T) {
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

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// First call returns session ID
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": sessionID,
					}), nil
				}

				// Second call returns torrents
				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{
			Host: "localhost",
			Port: 9091,
		}
		client := NewTransmissionClientWithHTTPClient(config, mockHTTP)

		dirs, err := client.GetDownloadDirectories(context.Background())
		require.NoError(t, err)

		assert.Len(t, dirs, 2)
		assert.Equal(t, "/downloads/movies", dirs[0].Path)
		assert.Equal(t, 2, dirs[0].Count)
		assert.Equal(t, "/downloads/tv", dirs[1].Path)
		assert.Equal(t, 1, dirs[1].Count)
	})
}

func TestBaseURL(t *testing.T) {
	config := types.Config{
		Host: "localhost",
		Port: 9091,
	}
	client := NewTransmissionClient(config)

	expected := "http://localhost:9091/transmission/rpc"
	assert.Equal(t, expected, client.baseURL())
}