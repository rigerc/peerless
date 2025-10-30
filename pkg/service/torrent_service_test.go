package service

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"peerless/pkg/client"
	"peerless/pkg/types"
)

func TestNewTorrentService(t *testing.T) {
	config := types.Config{
		Host: "localhost",
		Port: 9091,
	}

	mockHTTP := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return NewMockResponse(409, "{}", map[string]string{
				"X-Transmission-Session-Id": "test-session",
			}), nil
		},
	}

	transmissionClient := client.NewTransmissionClientWithHTTPClient(config, mockHTTP)
	service := NewTorrentService(transmissionClient)

	assert.NotNil(t, service)
	assert.Equal(t, transmissionClient, service.client)
}

func TestTorrentService_CheckDirectories(t *testing.T) {
	t.Run("successful directory check", func(t *testing.T) {
		// Create temporary directories and files
		tmpDir, err := os.MkdirTemp("", "test_service_")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create test files
		file1 := filepath.Join(tmpDir, "Movie1.2024.1080p.BluRay.x264")
		file2 := filepath.Join(tmpDir, "Movie2.2024.720p.WEBRip.x264")
		file3 := filepath.Join(tmpDir, "LocalFile.txt")

		err = os.WriteFile(file1, []byte("movie1 content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(file2, []byte("movie2 content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(file3, []byte("local content"), 0644)
		require.NoError(t, err)

		// Mock torrent data
		mockResponse := `{
			"arguments": {
				"torrents": [
					{
						"id": 1,
						"name": "Movie1.2024.1080p.BluRay.x264",
						"downloadDir": "/downloads",
						"hashString": "abc123"
					},
					{
						"id": 2,
						"name": "Movie2.2024.720p.WEBRip.x264",
						"downloadDir": "/downloads",
						"hashString": "def456"
					}
				]
			},
			"result": "success"
		}`

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": "test-session",
					}), nil
				}
				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{Host: "localhost", Port: 9091}
		transmissionClient := client.NewTransmissionClientWithHTTPClient(config, mockHTTP)
		service := NewTorrentService(transmissionClient)

		// Test directory check
		result, err := service.CheckDirectories(context.Background(), []string{tmpDir})
		require.NoError(t, err)

		// Verify results
		assert.Len(t, result.Directories, 1)
		dirResult := result.Directories[0]
		assert.Equal(t, tmpDir, dirResult.Path)
		assert.Equal(t, 3, dirResult.TotalItems)        // 3 files in directory
		assert.Equal(t, 2, dirResult.FoundItems)        // 2 files found in torrents
		assert.Equal(t, 1, len(dirResult.MissingPaths)) // 1 file missing
		assert.Contains(t, dirResult.MissingPaths, file3)

		// Verify overall results
		assert.Equal(t, 3, result.TotalItems)
		assert.Equal(t, 2, result.TotalFound)
		assert.Len(t, result.MissingPaths, 1)
	})

	t.Run("multiple directories", func(t *testing.T) {
		// Create temporary directories
		tmpDir1, err := os.MkdirTemp("", "test_service1_")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir1)

		tmpDir2, err := os.MkdirTemp("", "test_service2_")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir2)

		// Create test files
		file1 := filepath.Join(tmpDir1, "Movie1.2024.1080p.BluRay.x264")
		file2 := filepath.Join(tmpDir2, "Series1.S01E01.1080p.BluRay.x264")

		err = os.WriteFile(file1, []byte("movie1 content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(file2, []byte("series1 content"), 0644)
		require.NoError(t, err)

		// Mock torrent data
		mockResponse := `{
			"arguments": {
				"torrents": [
					{
						"id": 1,
						"name": "Movie1.2024.1080p.BluRay.x264",
						"downloadDir": "/downloads",
						"hashString": "abc123"
					}
				]
			},
			"result": "success"
		}`

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": "test-session",
					}), nil
				}
				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{Host: "localhost", Port: 9091}
		transmissionClient := client.NewTransmissionClientWithHTTPClient(config, mockHTTP)
		service := NewTorrentService(transmissionClient)

		// Test directory check
		result, err := service.CheckDirectories(context.Background(), []string{tmpDir1, tmpDir2})
		require.NoError(t, err)

		// Verify results
		assert.Len(t, result.Directories, 2)
		assert.Equal(t, 2, result.TotalItems)
		assert.Equal(t, 1, result.TotalFound)
		assert.Len(t, result.MissingPaths, 1)
	})
}

func TestTorrentService_GetTorrentStatistics(t *testing.T) {
	t.Run("successful statistics retrieval", func(t *testing.T) {
		mockResponse := `{
			"arguments": {
				"torrents": [
					{
						"id": 1,
						"name": "Movie1.2024",
						"downloadDir": "/downloads/movies",
						"hashString": "abc123"
					},
					{
						"id": 2,
						"name": "Movie2.2024",
						"downloadDir": "/downloads/movies",
						"hashString": "def456"
					},
					{
						"id": 3,
						"name": "Series1.S01E01",
						"downloadDir": "/downloads/tv",
						"hashString": "ghi789"
					}
				]
			},
			"result": "success"
		}`

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": "test-session",
					}), nil
				}
				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{Host: "localhost", Port: 9091}
		transmissionClient := client.NewTransmissionClientWithHTTPClient(config, mockHTTP)
		service := NewTorrentService(transmissionClient)

		stats, err := service.GetTorrentStatistics(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 3, stats.TotalTorrents)
		assert.Len(t, stats.Directories, 2)
		assert.Equal(t, 2, stats.Directories["/downloads/movies"])
		assert.Equal(t, 1, stats.Directories["/downloads/tv"])
	})
}

func TestTorrentService_CompareLocalWithTransmission(t *testing.T) {
	t.Run("successful comparison", func(t *testing.T) {
		// Create temporary directory with test files
		tmpDir, err := os.MkdirTemp("", "test_compare_")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		localFile := filepath.Join(tmpDir, "LocalFile.txt")
		torrentFile := filepath.Join(tmpDir, "TorrentFile.txt")

		err = os.WriteFile(localFile, []byte("local content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(torrentFile, []byte("torrent content"), 0644)
		require.NoError(t, err)

		// Mock torrent data with path matching torrentFile
		mockResponse := `{
			"arguments": {
				"torrents": [
					{
						"id": 1,
						"name": "TorrentFile.txt",
						"downloadDir": "` + tmpDir + `",
						"hashString": "abc123"
					},
					{
						"id": 2,
						"name": "RemoteOnly.txt",
						"downloadDir": "/downloads/remote",
						"hashString": "def456"
					}
				]
			},
			"result": "success"
		}`

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": "test-session",
					}), nil
				}
				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{Host: "localhost", Port: 9091}
		transmissionClient := client.NewTransmissionClientWithHTTPClient(config, mockHTTP)
		service := NewTorrentService(transmissionClient)

		result, err := service.CompareLocalWithTransmission(context.Background(), tmpDir)
		require.NoError(t, err)

		// Verify comparison results
		assert.Equal(t, 2, result.TotalLocal)        // 2 local files
		assert.Equal(t, 2, result.TotalTransmission) // 2 torrent files
		assert.Len(t, result.InBoth, 1)              // 1 file in both
		assert.Len(t, result.LocalOnly, 1)           // 1 local-only file
		assert.Len(t, result.InTransmissionOnly, 1)  // 1 transmission-only file

		// Check that the common file is identified correctly
		absTorrentFile, _ := filepath.Abs(torrentFile)
		assert.Contains(t, result.InBoth, absTorrentFile)

		absLocalFile, _ := filepath.Abs(localFile)
		assert.Contains(t, result.LocalOnly, absLocalFile)
	})
}

func TestTorrentService_GetDownloadDirectories(t *testing.T) {
	t.Run("successful directory listing", func(t *testing.T) {
		mockResponse := `{
			"arguments": {
				"torrents": [
					{
						"id": 1,
						"name": "Movie1.2024",
						"downloadDir": "/downloads/movies",
						"hashString": "abc123"
					},
					{
						"id": 2,
						"name": "Movie2.2024",
						"downloadDir": "/downloads/movies",
						"hashString": "def456"
					},
					{
						"id": 3,
						"name": "Series1.S01E01",
						"downloadDir": "/downloads/tv",
						"hashString": "ghi789"
					}
				]
			},
			"result": "success"
		}`

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": "test-session",
					}), nil
				}
				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{Host: "localhost", Port: 9091}
		transmissionClient := client.NewTransmissionClientWithHTTPClient(config, mockHTTP)
		service := NewTorrentService(transmissionClient)

		dirs, err := service.GetDownloadDirectories(context.Background())
		require.NoError(t, err)

		assert.Len(t, dirs, 2)
		assert.Equal(t, "/downloads/movies", dirs[0].Path)
		assert.Equal(t, 2, dirs[0].Count)
		assert.Equal(t, "/downloads/tv", dirs[1].Path)
		assert.Equal(t, 1, dirs[1].Count)
	})
}

func TestTorrentService_GetAllTorrentPaths(t *testing.T) {
	t.Run("successful path retrieval", func(t *testing.T) {
		mockResponse := `{
			"arguments": {
				"torrents": [
					{
						"id": 1,
						"name": "Movie1.2024",
						"downloadDir": "/downloads/movies",
						"hashString": "abc123"
					},
					{
						"id": 2,
						"name": "Movie2.2024",
						"downloadDir": "/downloads/tv",
						"hashString": "def456"
					}
				]
			},
			"result": "success"
		}`

		mockHTTP := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("X-Transmission-Session-Id") == "" {
					return NewMockResponse(409, "{}", map[string]string{
						"X-Transmission-Session-Id": "test-session",
					}), nil
				}
				return NewMockResponse(200, mockResponse, map[string]string{
					"Content-Type": "application/json",
				}), nil
			},
		}

		config := types.Config{Host: "localhost", Port: 9091}
		transmissionClient := client.NewTransmissionClientWithHTTPClient(config, mockHTTP)
		service := NewTorrentService(transmissionClient)

		paths, err := service.GetAllTorrentPaths(context.Background())
		require.NoError(t, err)

		assert.Len(t, paths, 2)
		assert.Contains(t, paths, "/downloads/movies/Movie1.2024")
		assert.Contains(t, paths, "/downloads/tv/Movie2.2024")
	})
}
