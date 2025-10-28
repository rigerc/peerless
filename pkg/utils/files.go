package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	if !info.IsDir() {
		return info.Size(), nil
	}

	// For directories, calculate total size recursively
	var totalSize int64
	err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

func WriteMissingPaths(filename string, paths []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, path := range paths {
		_, err := file.WriteString(path + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// DirectoryInfo represents a directory with its torrent count
type DirectoryInfo struct {
	Path  string
	Count int
}

// WriteDirectoryList writes a list of directories to a file
func WriteDirectoryList(filename string, dirs []DirectoryInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, dir := range dirs {
		_, err := file.WriteString(fmt.Sprintf("%s (%d torrents)\n", dir.Path, dir.Count))
		if err != nil {
			return err
		}
	}

	return nil
}