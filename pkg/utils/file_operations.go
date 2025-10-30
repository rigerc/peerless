package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileOperation represents an operation on a file or directory
type FileOperation struct {
	Path  string
	Size  int64
	IsDir bool
	Error error
}

// FileOperationResult tracks the result of file operations
type FileOperationResult struct {
	Success      []FileOperation
	Failed       []FileOperation
	TotalSize    int64
	SuccessCount int
	FailedCount  int
}

// DeleteProgressCallback is called for each file during deletion
type DeleteProgressCallback func(current, total int, path string, size int64)

// FileInfo retrieves detailed information about a file or directory
func FileInfo(path string) (*FileOperation, error) {
	info, err := os.Stat(path)
	if err != nil {
		return &FileOperation{Path: path, Error: err}, err
	}

	op := &FileOperation{
		Path:  path,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		op.Size = info.Size()
	} else {
		size, err := GetSize(path)
		if err != nil {
			op.Error = err
		} else {
			op.Size = size
		}
	}

	return op, nil
}

// BatchFileInfo retrieves information for multiple paths
func BatchFileInfo(paths []string) []*FileOperation {
	operations := make([]*FileOperation, 0, len(paths))

	for _, path := range paths {
		op, _ := FileInfo(path)
		operations = append(operations, op)
	}

	return operations
}

// DeleteFiles deletes multiple files/directories with progress tracking
func DeleteFiles(paths []string, progressCallback DeleteProgressCallback) *FileOperationResult {
	result := &FileOperationResult{
		Success: make([]FileOperation, 0),
		Failed:  make([]FileOperation, 0),
	}

	total := len(paths)

	for i, path := range paths {
		op, err := FileInfo(path)

		if progressCallback != nil {
			progressCallback(i+1, total, path, op.Size)
		}

		if err != nil {
			op.Error = err
			result.Failed = append(result.Failed, *op)
			result.FailedCount++
			continue
		}

		var deleteErr error
		if op.IsDir {
			deleteErr = os.RemoveAll(path)
		} else {
			deleteErr = os.Remove(path)
		}

		if deleteErr != nil {
			op.Error = deleteErr
			result.Failed = append(result.Failed, *op)
			result.FailedCount++
		} else {
			result.Success = append(result.Success, *op)
			result.SuccessCount++
			result.TotalSize += op.Size
		}
	}

	return result
}

// ValidateDeletionPaths validates paths before deletion
func ValidateDeletionPaths(paths []string, allowedDirs []string) error {
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", path, err)
		}

		// Check if path is within allowed directories
		if len(allowedDirs) > 0 {
			allowed := false
			for _, allowedDir := range allowedDirs {
				absAllowedDir, err := filepath.Abs(allowedDir)
				if err != nil {
					continue
				}

				rel, err := filepath.Rel(absAllowedDir, absPath)
				if err == nil && !filepath.IsAbs(rel) && len(rel) > 0 && rel[0] != '.' {
					allowed = true
					break
				}
			}

			if !allowed {
				return fmt.Errorf("path %s is not within allowed directories", path)
			}
		}

		// Prevent deletion of system directories
		if isSystemPath(absPath) {
			return fmt.Errorf("refusing to delete system path: %s", absPath)
		}
	}

	return nil
}

// isSystemPath checks if a path is a critical system directory
func isSystemPath(path string) bool {
	systemPaths := []string{
		"/", "/bin", "/boot", "/dev", "/etc", "/lib", "/lib64",
		"/proc", "/root", "/sbin", "/sys", "/usr", "/var",
		"C:\\Windows", "C:\\Program Files", "C:\\Program Files (x86)",
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, sysPath := range systemPaths {
		if absPath == sysPath {
			return true
		}
	}

	return false
}

// CalculateTotalSize calculates total size for a list of paths
func CalculateTotalSize(paths []string) (int64, int, error) {
	var totalSize int64
	var inaccessible int

	for _, path := range paths {
		size, err := GetSize(path)
		if err != nil {
			inaccessible++
			continue
		}
		totalSize += size
	}

	return totalSize, inaccessible, nil
}