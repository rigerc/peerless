package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"peerless/pkg/constants"
)

func GetSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if !info.IsDir() {
		return info.Size(), nil
	}

	var totalSize int64
	var walkErr error

	err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			// Log but don't fail entirely - collect the error but continue walking
			walkErr = fmt.Errorf("error accessing %s: %w", p, err)
			return nil
		}
		if !d.IsDir() {
			fileInfo, err := d.Info()
			if err == nil {
				totalSize += fileInfo.Size()
			}
		}
		return nil
	})

	if err != nil {
		return totalSize, err
	}

	// Return any walk errors that occurred but don't fail if we have some size data
	if walkErr != nil {
		return totalSize, walkErr
	}

	return totalSize, nil
}

func FormatSize(bytes int64) string {
	if bytes < constants.BytesPerKB {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(constants.BytesPerKB), 0
	for n := bytes / constants.BytesPerKB; n >= constants.BytesPerKB; n /= constants.BytesPerKB {
		div *= constants.BytesPerKB
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

func WriteMissingPaths(filename string, paths []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	for _, path := range paths {
		cleanPath := SanitizeString(path)
		if _, err := file.WriteString(cleanPath + "\n"); err != nil {
			return fmt.Errorf("failed to write path %s to file %s: %w", path, filename, err)
		}
	}

	// Ensure all data is flushed to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file %s: %w", filename, err)
	}

	return nil
}

// NormalizeName normalizes a name for comparison based on OS case sensitivity
func NormalizeName(name string) string {
	if isCaseSensitive() {
		return name
	}
	return strings.ToLower(name)
}

// isCaseSensitive determines if the current file system is case-sensitive
func isCaseSensitive() bool {
	// Windows is case-insensitive by default
	// macOS can be case-insensitive (APFS default) or case-sensitive (APFS case-sensitive)
	// Linux is typically case-sensitive
	return runtime.GOOS != "windows"
}

// SanitizeString removes control characters and LTR/RTL marks from strings
func SanitizeString(s string) string {
	var result strings.Builder
	for _, r := range s {
		// Skip control characters except newline, tab, and carriage return
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			continue
		}
		// Skip specific Unicode formatting characters
		if r == constants.LTRMark || r == constants.RTLMark || r == constants.LRE || r == constants.RLE || r == constants.PDF || r == constants.LRO || r == constants.RLO {
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
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
		cleanPath := SanitizeString(dir.Path)
		_, err := file.WriteString(fmt.Sprintf("%s (%d torrents)\n", cleanPath, dir.Count))
		if err != nil {
			return err
		}
	}

	return nil
}