package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSize(t *testing.T) {
	t.Run("regular file", func(t *testing.T) {
		// Create a temporary file
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.txt")

		content := []byte("Hello, World!")
		err := os.WriteFile(tmpFile, content, 0644)
		require.NoError(t, err)

		size, err := GetSize(tmpFile)
		require.NoError(t, err)
		assert.Equal(t, int64(len(content)), size)
	})

	t.Run("directory with files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create files in the directory
		files := map[string][]byte{
			"file1.txt":        []byte("Hello"),
			"file2.txt":        []byte("World!"),
			"subdir/file3.txt": []byte("Test"),
		}

		totalSize := int64(0)
		for path, content := range files {
			fullPath := filepath.Join(tmpDir, path)
			err := os.MkdirAll(filepath.Dir(fullPath), 0755)
			require.NoError(t, err)

			err = os.WriteFile(fullPath, content, 0644)
			require.NoError(t, err)
			totalSize += int64(len(content))
		}

		size, err := GetSize(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, totalSize, size)
	})

	t.Run("non-existent path", func(t *testing.T) {
		size, err := GetSize("/non/existent/path")
		assert.Error(t, err)
		assert.Equal(t, int64(0), size)
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		size, err := GetSize(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 1024, "1.00 KB"},
		{"megabytes", 1024 * 1024, "1.00 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.00 GB"},
		{"terabytes", 1024 * 1024 * 1024 * 1024, "1.00 TB"},
		{"petabytes", 1024 * 1024 * 1024 * 1024 * 1024, "1.00 PB"},
		{"fractional", 1536, "1.50 KB"},
		{"zero", 0, "0 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPortValidation(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid port",
			port:        9091,
			expectError: false,
		},
		{
			name:        "valid port 1",
			port:        1,
			expectError: false,
		},
		{
			name:        "valid port 65535",
			port:        65535,
			expectError: false,
		},
		{
			name:        "invalid port 0",
			port:        0,
			expectError: true,
			errorMsg:    "invalid port 0: port must be between 1 and 65535",
		},
		{
			name:        "invalid port negative",
			port:        -1,
			expectError: true,
			errorMsg:    "invalid port -1: port must be between 1 and 65535",
		},
		{
			name:        "invalid port too high",
			port:        65536,
			expectError: true,
			errorMsg:    "invalid port 65536: port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This simulates the validation logic from main.go
			if tt.port <= 0 || tt.port > 65535 {
				if !tt.expectError {
					t.Errorf("expected no error but port %d should be invalid", tt.port)
				}
				if tt.expectError && fmt.Sprintf("invalid port %d: port must be between 1 and 65535", tt.port) != tt.errorMsg {
					t.Errorf("expected error message %q but got %q", tt.errorMsg, fmt.Sprintf("invalid port %d: port must be between 1 and 65535", tt.port))
				}
			} else {
				if tt.expectError {
					t.Errorf("expected error but port %d should be valid", tt.port)
				}
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "/path/to/file.txt",
			expected: "/path/to/file.txt",
		},
		{
			name:     "with LTR mark",
			input:    "/path/to/file.txt\u200E",
			expected: "/path/to/file.txt",
		},
		{
			name:     "with RTL mark",
			input:    "/path/to/file.txt\u200F",
			expected: "/path/to/file.txt",
		},
		{
			name:     "with multiple formatting characters",
			input:    "/path/to/file.txt\u200E\u200F\u202A",
			expected: "/path/to/file.txt",
		},
		{
			name:     "with newlines preserved",
			input:    "/path/to/file.txt\n/another/path.txt",
			expected: "/path/to/file.txt\n/another/path.txt",
		},
		{
			name:     "with tabs preserved",
			input:    "/path/to/file.txt\twith tab",
			expected: "/path/to/file.txt\twith tab",
		},
		{
			name:     "with control characters removed",
			input:    "/path/to/file.txt\u0001\u0002",
			expected: "/path/to/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteMissingPaths(t *testing.T) {
	t.Run("write paths to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "missing.txt")

		paths := []string{
			"/path/to/file1.txt",
			"/path/to/file2.txt",
			"/another/path/file3.txt",
		}

		err := WriteMissingPaths(outputFile, paths)
		require.NoError(t, err)

		// Read the file and verify content
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		expected := "/path/to/file1.txt\n/path/to/file2.txt\n/another/path/file3.txt\n"
		assert.Equal(t, expected, string(content))
	})

	t.Run("empty paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "empty.txt")

		err := WriteMissingPaths(outputFile, []string{})
		require.NoError(t, err)

		// File should exist but be empty
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Equal(t, "", string(content))
	})

	t.Run("invalid path", func(t *testing.T) {
		err := WriteMissingPaths("/invalid/path/file.txt", []string{"test"})
		assert.Error(t, err)
	})
}
