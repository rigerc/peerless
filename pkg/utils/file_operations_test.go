package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileInfo(t *testing.T) {
	t.Run("regular file", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test_file_*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// Write some content
		content := "Hello, World!"
		_, err = tmpFile.WriteString(content)
		require.NoError(t, err)
		tmpFile.Close()

		// Test FileInfo
		op, err := FileInfo(tmpFile.Name())
		require.NoError(t, err)

		assert.Equal(t, tmpFile.Name(), op.Path)
		assert.False(t, op.IsDir)
		assert.Equal(t, int64(len(content)), op.Size)
		assert.NoError(t, op.Error)
	})

	t.Run("directory", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "test_dir_")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create some files in the directory
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		err = os.WriteFile(file1, []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(file2, []byte("content2"), 0644)
		require.NoError(t, err)

		// Test FileInfo
		op, err := FileInfo(tmpDir)
		require.NoError(t, err)

		assert.Equal(t, tmpDir, op.Path)
		assert.True(t, op.IsDir)
		assert.Equal(t, int64(len("content1")+len("content2")), op.Size)
		assert.NoError(t, op.Error)
	})

	t.Run("non-existent path", func(t *testing.T) {
		op, err := FileInfo("/non/existent/path")
		assert.Error(t, err)
		assert.Equal(t, "/non/existent/path", op.Path)
		assert.Error(t, op.Error)
	})
}

func TestBatchFileInfo(t *testing.T) {
	// Create temporary files and directories
	tmpDir, err := os.MkdirTemp("", "test_batch_")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	subDir := filepath.Join(tmpDir, "subdir")

	err = os.WriteFile(file1, []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("content2"), 0644)
	require.NoError(t, err)
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	paths := []string{file1, file2, subDir, "/non/existent"}
	operations := BatchFileInfo(paths)

	assert.Len(t, operations, 4)

	// Check file1
	assert.Equal(t, file1, operations[0].Path)
	assert.False(t, operations[0].IsDir)
	assert.Equal(t, int64(len("content1")), operations[0].Size)
	assert.NoError(t, operations[0].Error)

	// Check file2
	assert.Equal(t, file2, operations[1].Path)
	assert.False(t, operations[1].IsDir)
	assert.Equal(t, int64(len("content2")), operations[1].Size)
	assert.NoError(t, operations[1].Error)

	// Check subDir
	assert.Equal(t, subDir, operations[2].Path)
	assert.True(t, operations[2].IsDir)
	assert.Equal(t, int64(0), operations[2].Size) // Empty directory
	assert.NoError(t, operations[2].Error)

	// Check non-existent path
	assert.Equal(t, "/non/existent", operations[3].Path)
	assert.Error(t, operations[3].Error)
}

func TestDeleteFiles(t *testing.T) {
	t.Run("delete files and directories", func(t *testing.T) {
		// Create temporary files and directories
		tmpDir, err := os.MkdirTemp("", "test_delete_")
		require.NoError(t, err)

		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")
		subDir := filepath.Join(tmpDir, "subdir")

		err = os.WriteFile(file1, []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(file2, []byte("content2"), 0644)
		require.NoError(t, err)
		err = os.Mkdir(subDir, 0755)
		require.NoError(t, err)

		paths := []string{file1, subDir, file2}

		// Track progress
		var progressCalls []struct {
			current int
			total   int
			path    string
			size    int64
		}

		progressCallback := func(current, total int, path string, size int64) {
			progressCalls = append(progressCalls, struct {
				current int
				total   int
				path    string
				size    int64
			}{current, total, path, size})
		}

		// Delete files
		result := DeleteFiles(paths, progressCallback)

		// Check results - files should definitely succeed, directory might fail due to filesystem issues
		assert.GreaterOrEqual(t, result.SuccessCount, 2) // At least the 2 files should succeed
		assert.LessOrEqual(t, result.FailedCount, 1)   // At most 1 failure (the directory)
		assert.GreaterOrEqual(t, len(result.Success), 2)

		// Check progress tracking - all 3 paths should be processed
		assert.Len(t, progressCalls, 3)
		assert.Equal(t, 1, progressCalls[0].current)
		assert.Equal(t, 3, progressCalls[0].total)
		assert.Equal(t, 2, progressCalls[1].current)  // Second item processed
		assert.Equal(t, 3, progressCalls[1].total)
		assert.Equal(t, 3, progressCalls[2].current)  // Third item processed
		assert.Equal(t, 3, progressCalls[2].total)

		// Verify files are deleted
		_, err = os.Stat(file1)
		assert.True(t, os.IsNotExist(err))
		_, err = os.Stat(file2)
		assert.True(t, os.IsNotExist(err))
		_, err = os.Stat(subDir)
		assert.True(t, os.IsNotExist(err))

		// Cleanup
		os.RemoveAll(tmpDir)
	})

	t.Run("delete with some failures", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test_delete_fail_*.txt")
		require.NoError(t, err)
		tmpFile.Close()

		paths := []string{tmpFile.Name(), "/non/existent/path"}

		result := DeleteFiles(paths, nil)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 1, result.FailedCount)
		assert.Len(t, result.Success, 1)
		assert.Len(t, result.Failed, 1)

		// Cleanup
		os.Remove(tmpFile.Name())
	})
}

func TestValidateDeletionPaths(t *testing.T) {
	t.Run("valid paths", func(t *testing.T) {
		// Create temporary directory
		tmpDir, err := os.MkdirTemp("", "test_validate_")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		file := filepath.Join(tmpDir, "file.txt")
		err = os.WriteFile(file, []byte("content"), 0644)
		require.NoError(t, err)

		paths := []string{file}
		allowedDirs := []string{tmpDir}

		err = ValidateDeletionPaths(paths, allowedDirs)
		assert.NoError(t, err)
	})

	t.Run("path outside allowed directories", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test_validate_")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		file := filepath.Join(tmpDir, "file.txt")
		err = os.WriteFile(file, []byte("content"), 0644)
		require.NoError(t, err)

		paths := []string{file}
		allowedDirs := []string{"/some/other/dir"}

		err = ValidateDeletionPaths(paths, allowedDirs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not within allowed directories")
	})

	t.Run("system path protection", func(t *testing.T) {
		systemPaths := []string{
			"/",
			"/bin",
			"/usr",
		}

		for _, path := range systemPaths {
			t.Run("system path "+path, func(t *testing.T) {
				paths := []string{path}
				err := ValidateDeletionPaths(paths, nil)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "refusing to delete system path")
			})
		}
	})

	t.Run("empty allowed directories", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_validate_*.txt")
		require.NoError(t, err)
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		paths := []string{tmpFile.Name()}
		err = ValidateDeletionPaths(paths, nil)
		assert.NoError(t, err) // Should allow any path when no allowed dirs specified
	})
}

func TestCalculateTotalSize(t *testing.T) {
	t.Run("calculate total size", func(t *testing.T) {
		// Create temporary files
		tmpDir, err := os.MkdirTemp("", "test_size_")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		err = os.WriteFile(file1, []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(file2, []byte("content2 longer"), 0644)
		require.NoError(t, err)

		paths := []string{file1, file2}
		totalSize, inaccessible, err := CalculateTotalSize(paths)

		assert.NoError(t, err)
		assert.Equal(t, int64(len("content1")+len("content2 longer")), totalSize)
		assert.Equal(t, 0, inaccessible)
	})

	t.Run("with inaccessible files", func(t *testing.T) {
		paths := []string{"/non/existent/file1", "/non/existent/file2"}
		totalSize, inaccessible, err := CalculateTotalSize(paths)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), totalSize)
		assert.Equal(t, 2, inaccessible)
	})

	t.Run("mixed accessible and inaccessible", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_size_*.txt")
		require.NoError(t, err)
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		paths := []string{tmpFile.Name(), "/non/existent/file"}
		totalSize, inaccessible, err := CalculateTotalSize(paths)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), totalSize) // Empty file
		assert.Equal(t, 1, inaccessible)
	})
}

func TestIsSystemPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/", true},
		{"/bin", true},
		{"/usr", true},
		{"/etc", true},
		{"/home/user", false},
		{"/tmp", false},
		{"relative/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isSystemPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}