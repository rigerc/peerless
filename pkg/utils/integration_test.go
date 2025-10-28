package utils_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"peerless/pkg/utils"
)

// TestFileOutputIntegration tests file output functionality
func TestFileOutputIntegration(t *testing.T) {
	t.Run("WriteDirectoryListToFile", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "directories.txt")

		// Test data
		dirs := []utils.DirectoryInfo{
			{Path: "/downloads/movies", Count: 5},
			{Path: "/downloads/tv", Count: 3},
			{Path: "/downloads/documentaries", Count: 2},
		}

		// Write to file
		err := utils.WriteDirectoryList(outputFile, dirs)
		if err != nil {
			t.Fatalf("Failed to write directory list: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Fatalf("Output file was not created: %s", outputFile)
		}

		// Read and verify content
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		contentStr := string(content)
		expectedLines := []string{
			"/downloads/documentaries (2 torrents)",
			"/downloads/movies (5 torrents)",
			"/downloads/tv (3 torrents)",
		}

		for _, expectedLine := range expectedLines {
			if !strings.Contains(contentStr, expectedLine) {
				t.Errorf("Expected line '%s' not found in content: %s", expectedLine, contentStr)
			}
		}

		// Verify content is plain text (no control characters)
		if strings.Contains(contentStr, "\u200E") || strings.Contains(contentStr, "\u200F") {
			t.Errorf("Output contains unwanted Unicode control characters: %s", contentStr)
		}
	})

	t.Run("WriteMissingPathsToFile", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "missing.txt")

		// Test data with potential Unicode characters
		paths := []string{
			"/downloads/movies/Movie.2023.1080p.BluRay.x264",
			"/downloads/tv/TV.Series.S01\u200E", // With LTR mark
			"/downloads/documentaries/Nature.Doc.2024",
			"/downloads/music/Album.2023\u200F",  // With RTL mark
		}

		// Write to file
		err := utils.WriteMissingPaths(outputFile, paths)
		if err != nil {
			t.Fatalf("Failed to write missing paths: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Fatalf("Output file was not created: %s", outputFile)
		}

		// Read and verify content
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		contentStr := string(content)

		// Verify all paths are present (without Unicode marks)
		expectedPaths := []string{
			"/downloads/movies/Movie.2023.1080p.BluRay.x264",
			"/downloads/tv/TV.Series.S01",
			"/downloads/documentaries/Nature.Doc.2024",
			"/downloads/music/Album.2023",
		}

		for _, expectedPath := range expectedPaths {
			if !strings.Contains(contentStr, expectedPath) {
				t.Errorf("Expected path '%s' not found in content: %s", expectedPath, contentStr)
			}
		}

		// Verify Unicode control characters are removed
		if strings.Contains(contentStr, "\u200E") || strings.Contains(contentStr, "\u200F") {
			t.Errorf("Output contains Unicode control characters that should have been sanitized: %s", contentStr)
		}
	})

	t.Run("EmptyFileOutput", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "empty.txt")

		// Test empty data
		dirs := []utils.DirectoryInfo{}
		err := utils.WriteDirectoryList(outputFile, dirs)
		if err != nil {
			t.Fatalf("Failed to write empty directory list: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Fatalf("Output file was not created: %s", outputFile)
		}

		// Read and verify content is empty
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		if len(content) != 0 {
			t.Errorf("Expected empty file, but got content: %s", string(content))
		}
	})
}

// TestFilePermissions tests file creation permissions
func TestFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "permissions.txt")

	// Test data
	dirs := []utils.DirectoryInfo{
		{Path: "/downloads/test", Count: 1},
	}

	// Write to file
	err := utils.WriteDirectoryList(outputFile, dirs)
	if err != nil {
		t.Fatalf("Failed to write directory list: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// Verify file is readable by owner
	if info.Mode().Perm()&0400 == 0 {
		t.Error("File is not readable by owner")
	}
}

// TestSpecialCharactersInPaths tests handling of special characters
func TestSpecialCharactersInPaths(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "special.txt")

	// Test data with various special characters
	paths := []string{
		"/downloads/Normal File",
		"/downloads/File-With-Dashes",
		"/downloads/File_With_Underscores",
		"/downloads/File.With.Dots",
		"/downloads/File With Spaces",
		"/downloads/File'With'Quotes",
		"/downloads/File(With)Parens",
		"/downloads/File[With]Brackets",
		"/downloads/File{With}Braces",
		"/downloads/File#With#Hash",
		"/downloads/File@With@At",
		"/downloads/File$With$Dollar",
		"/downloads/File%With%Percent",
		"/downloads/File^With^Caret",
		"/downloads/File&With&Ampersand",
		"/downloads/File*With*Asterisk",
		"/downloads/File+With+Plus",
		"/downloads/File=With=Equals",
		"/downloads/File|With|Pipe",
		"/downloads/File\\With\\Backslash",
		"/downloads/File/With/Slashes",
		"/downloads/File:With:Colon",
		"/downloads/File;With;Semicolon",
		"/downloads/File\"With\"Quotes",
		"/downloads/File<With>Brackets",
		"/downloads/File?With?Question",
		// Unicode characters (should be preserved)
		"/downloads/电影.MOV", // Chinese characters
		"/downloads/Фильм.MOV", // Cyrillic characters
		"/downloads/فيلم.MOV", // Arabic characters
	}

	// Write to file
	err := utils.WriteMissingPaths(outputFile, paths)
	if err != nil {
		t.Fatalf("Failed to write paths with special characters: %v", err)
	}

	// Read and verify content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify all paths are present
	for _, expectedPath := range paths {
		if !strings.Contains(contentStr, expectedPath) {
			t.Errorf("Expected path '%s' not found in content", expectedPath)
		}
	}

	// Verify control characters are removed
	controlChars := []string{"\u200E", "\u200F", "\u202A", "\u202B", "\u202C", "\u202D", "\u202E"}
	for _, char := range controlChars {
		if strings.Contains(contentStr, char) {
			t.Errorf("Output contains control character '%s' that should have been removed", char)
		}
	}
}