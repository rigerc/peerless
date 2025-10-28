package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransmissionRequest_MarshalJSON(t *testing.T) {
	t.Run("basic request", func(t *testing.T) {
		req := TransmissionRequest{
			Method: "torrent-get",
			Arguments: map[string]interface{}{
				"fields": []string{"id", "name", "downloadDir"},
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "torrent-get", result["method"])
		assert.Contains(t, result, "arguments")
	})

	t.Run("request without arguments", func(t *testing.T) {
		req := TransmissionRequest{
			Method: "session-get",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "session-get", result["method"])
		assert.NotContains(t, result, "arguments")
	})
}

func TestTransmissionResponse_UnmarshalJSON(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		jsonData := `{
			"arguments": {
				"torrents": [
					{
						"id": 1,
						"name": "Test Torrent",
						"downloadDir": "/downloads",
						"hashString": "abc123def456"
					}
				]
			},
			"result": "success"
		}`

		var resp TransmissionResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)

		assert.Equal(t, "success", resp.Result)
		assert.Len(t, resp.Arguments.Torrents, 1)

		torrent := resp.Arguments.Torrents[0]
		assert.Equal(t, 1, torrent.ID)
		assert.Equal(t, "Test Torrent", torrent.Name)
		assert.Equal(t, "/downloads", torrent.DownloadDir)
		assert.Equal(t, "abc123def456", torrent.HashString)
	})

	t.Run("empty torrents list", func(t *testing.T) {
		jsonData := `{
			"arguments": {
				"torrents": []
			},
			"result": "success"
		}`

		var resp TransmissionResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)

		assert.Equal(t, "success", resp.Result)
		assert.Len(t, resp.Arguments.Torrents, 0)
	})

	t.Run("error response", func(t *testing.T) {
		jsonData := `{
			"arguments": {},
			"result": "error: invalid request"
		}`

		var resp TransmissionResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)

		assert.Equal(t, "error: invalid request", resp.Result)
	})
}

func TestTorrentInfo_Fields(t *testing.T) {
	torrent := TorrentInfo{
		ID:          42,
		Name:        "Example Torrent",
		DownloadDir: "/path/to/downloads",
		HashString:  "1234567890abcdef",
	}

	assert.Equal(t, 42, torrent.ID)
	assert.Equal(t, "Example Torrent", torrent.Name)
	assert.Equal(t, "/path/to/downloads", torrent.DownloadDir)
	assert.Equal(t, "1234567890abcdef", torrent.HashString)
}

func TestConfig_Fields(t *testing.T) {
	config := Config{
		Host:     "192.168.1.100",
		Port:     9091,
		User:     "admin",
		Password: "secret123",
		Dirs:     []string{"/downloads", "/media/torrents"},
	}

	assert.Equal(t, "192.168.1.100", config.Host)
	assert.Equal(t, 9091, config.Port)
	assert.Equal(t, "admin", config.User)
	assert.Equal(t, "secret123", config.Password)
	assert.Equal(t, []string{"/downloads", "/media/torrents"}, config.Dirs)
}

func TestConfig_DefaultValues(t *testing.T) {
	var config Config

	// Test zero values
	assert.Empty(t, config.Host)
	assert.Zero(t, config.Port)
	assert.Empty(t, config.User)
	assert.Empty(t, config.Password)
	assert.Nil(t, config.Dirs)
}

func TestConfig_WithEmptyDirs(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     9091,
		User:     "test",
		Password: "test",
		Dirs:     []string{}, // Empty slice
	}

	assert.NotNil(t, config.Dirs)
	assert.Len(t, config.Dirs, 0)
}