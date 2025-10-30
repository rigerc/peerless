package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"peerless/pkg/client"
	"peerless/pkg/types"
	"peerless/pkg/utils"
)

// TorrentService handles torrent-related business logic
type TorrentService struct {
	client *client.TransmissionClient
}

// NewTorrentService creates a new TorrentService
func NewTorrentService(client *client.TransmissionClient) *TorrentService {
	return &TorrentService{client: client}
}

// DirectoryCheckResult contains the results of checking directories
type DirectoryCheckResult struct {
	Directories      []DirectoryResult
	TotalItems       int
	TotalFound       int
	TotalMissingSize int64
	MissingPaths     []string
}

// DirectoryResult contains results for a single directory
type DirectoryResult struct {
	Path         string
	TotalItems   int
	FoundItems   int
	MissingSize  int64
	MissingPaths []string
}

// CheckDirectories checks local directories against Transmission torrents
func (s *TorrentService) CheckDirectories(ctx context.Context, dirs []string) (*DirectoryCheckResult, error) {
	torrents, err := s.client.GetTorrents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve torrents: %w", err)
	}

	torrentMap := make(map[string]bool)
	for _, t := range torrents {
		torrentMap[utils.NormalizeName(t.Name)] = true
	}

	result := &DirectoryCheckResult{
		Directories: make([]DirectoryResult, 0, len(dirs)),
	}

	for _, dir := range dirs {
		dirResult, err := s.checkSingleDirectory(dir, torrentMap)
		if err != nil {
			return nil, fmt.Errorf("failed to check directory %s: %w", dir, err)
		}

		result.Directories = append(result.Directories, *dirResult)
		result.TotalItems += dirResult.TotalItems
		result.TotalFound += dirResult.FoundItems
		result.TotalMissingSize += dirResult.MissingSize
		result.MissingPaths = append(result.MissingPaths, dirResult.MissingPaths...)
	}

	return result, nil
}

// checkSingleDirectory checks a single directory
func (s *TorrentService) checkSingleDirectory(dir string, torrentMap map[string]bool) (*DirectoryResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	result := &DirectoryResult{
		Path:         dir,
		TotalItems:   len(entries),
		MissingPaths: make([]string, 0),
	}

	for _, entry := range entries {
		name := entry.Name()
		inTransmission := torrentMap[utils.NormalizeName(name)]

		if inTransmission {
			result.FoundItems++
		} else {
			fullPath := filepath.Join(dir, name)
			absPath, err := filepath.Abs(fullPath)
			if err != nil {
				absPath = fullPath
			}

			result.MissingPaths = append(result.MissingPaths, absPath)

			size, err := utils.GetSize(fullPath)
			if err == nil {
				result.MissingSize += size
			}
		}
	}

	return result, nil
}

// TorrentStatistics contains statistics about torrents
type TorrentStatistics struct {
	TotalTorrents int
	Directories   map[string]int
}

// GetTorrentStatistics returns statistics about torrents
func (s *TorrentService) GetTorrentStatistics(ctx context.Context) (*TorrentStatistics, error) {
	torrents, err := s.client.GetTorrents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve torrents: %w", err)
	}

	stats := &TorrentStatistics{
		TotalTorrents: len(torrents),
		Directories:   make(map[string]int),
	}

	for _, t := range torrents {
		stats.Directories[t.DownloadDir]++
	}

	return stats, nil
}

// DetailedStatus contains comprehensive Transmission status information
type DetailedStatus struct {
	// Torrent counts
	TotalTorrents       int
	DownloadingTorrents int
	SeedingTorrents     int
	PausedTorrents      int
	CompletedTorrents   int

	// Size information
	TotalSize      int64
	DownloadedSize int64
	RemainingSize  int64

	// Speed information
	TotalDownloadSpeed int
	TotalUploadSpeed   int

	// Session information
	DownloadDir     string
	FreeSpace       int64
	PeerPort        int
	AltSpeedEnabled bool

	// Statistics
	CurrentSessionStats *types.SessionStats
	CumulativeStats     *types.SessionStats

	// Torrent breakdown by directory
	DirectoryBreakdown map[string]DirectoryStatus
}

// DirectoryStatus contains status for a specific download directory
type DirectoryStatus struct {
	TorrentCount   int
	TotalSize      int64
	DownloadedSize int64
	FreeSpace      int64
}

// GetDetailedStatus returns comprehensive Transmission status
func (s *TorrentService) GetDetailedStatus(ctx context.Context) (*DetailedStatus, error) {
	// Get all torrents with detailed information
	torrents, err := s.client.GetTorrents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve torrents: %w", err)
	}

	// Get session information
	sessionInfo, err := s.client.GetSessionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session info: %w", err)
	}

	// Get session statistics
	currentStats, cumulativeStats, err := s.client.GetSessionStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session stats: %w", err)
	}

	status := &DetailedStatus{
		TotalTorrents:       len(torrents),
		TotalSize:           0,
		DownloadedSize:      0,
		RemainingSize:       0,
		TotalDownloadSpeed:  0,
		TotalUploadSpeed:    0,
		DownloadDir:         sessionInfo.DownloadDir,
		FreeSpace:           sessionInfo.DownloadDirFree,
		PeerPort:            sessionInfo.PeerPort,
		AltSpeedEnabled:     sessionInfo.AltSpeedEnabled,
		CurrentSessionStats: currentStats,
		CumulativeStats:     cumulativeStats,
		DirectoryBreakdown:  make(map[string]DirectoryStatus),
	}

	// Process torrents
	for _, torrent := range torrents {
		status.TotalSize += torrent.TotalSize
		status.DownloadedSize += torrent.DownloadedEver
		status.RemainingSize += torrent.LeftUntilDone
		status.TotalDownloadSpeed += torrent.RateDownload
		status.TotalUploadSpeed += torrent.RateUpload

		// Count by status
		switch torrent.Status {
		case 0: // Stopped
			if torrent.PercentDone >= 1.0 {
				status.CompletedTorrents++
			} else {
				status.PausedTorrents++
			}
		case 1: // Queued to verify
		case 2: // Verifying
		case 3: // Queued to download
		case 4: // Downloading
			status.DownloadingTorrents++
		case 5: // Queued to seed
		case 6: // Seeding
			status.SeedingTorrents++
		}

		// Directory breakdown
		dirStatus, exists := status.DirectoryBreakdown[torrent.DownloadDir]
		if !exists {
			dirStatus = DirectoryStatus{
				FreeSpace: sessionInfo.DownloadDirFree, // Use global free space as fallback
			}
		}

		dirStatus.TorrentCount++
		dirStatus.TotalSize += torrent.TotalSize
		dirStatus.DownloadedSize += torrent.DownloadedEver

		status.DirectoryBreakdown[torrent.DownloadDir] = dirStatus
	}

	return status, nil
}

// CompareResult represents the result of comparing local vs Transmission
type CompareResult struct {
	InTransmissionOnly []string
	LocalOnly          []string
	InBoth             []string
	TotalLocal         int
	TotalTransmission  int
}

// CompareLocalWithTransmission compares local files with Transmission torrents
func (s *TorrentService) CompareLocalWithTransmission(ctx context.Context, dir string) (*CompareResult, error) {
	torrentPaths, err := s.client.GetAllTorrentPaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve torrent paths: %w", err)
	}

	torrentMap := make(map[string]bool)
	for _, path := range torrentPaths {
		torrentMap[path] = true
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	result := &CompareResult{
		InTransmissionOnly: make([]string, 0),
		LocalOnly:          make([]string, 0),
		InBoth:             make([]string, 0),
		TotalLocal:         len(entries),
		TotalTransmission:  len(torrentPaths),
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			absPath = fullPath
		}

		if torrentMap[absPath] {
			result.InBoth = append(result.InBoth, absPath)
			delete(torrentMap, absPath)
		} else {
			result.LocalOnly = append(result.LocalOnly, absPath)
		}
	}

	for path := range torrentMap {
		result.InTransmissionOnly = append(result.InTransmissionOnly, path)
	}

	return result, nil
}

// GetDownloadDirectories returns download directories with torrent counts
func (s *TorrentService) GetDownloadDirectories(ctx context.Context) ([]utils.DirectoryInfo, error) {
	return s.client.GetDownloadDirectories(ctx)
}

// GetAllTorrentPaths returns all torrent paths
func (s *TorrentService) GetAllTorrentPaths(ctx context.Context) ([]string, error) {
	return s.client.GetAllTorrentPaths(ctx)
}
