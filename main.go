package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"go-tneat/pkg/client"
	"go-tneat/pkg/types"
	"go-tneat/pkg/utils"
)

func main() {
	app := &cli.Command{
		Name:  "go-tneat",
		Usage: "Transmission neat - check local directories against Transmission torrents",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Value:   "localhost",
				Usage:   "Transmission host",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   9091,
				Usage:   "Transmission port",
			},
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "Transmission username",
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"P"},
				Usage:   "Transmission password",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "check",
				Usage: "Check local directories against Transmission torrents",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "dir",
						Aliases: []string{"d"},
						Usage:   "Directory to check (can be specified multiple times)",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output file for absolute paths of missing items",
					},
				},
				Action: runCheck,
			},
			{
				Name:  "list-directories",
				Usage: "List all download directories from Transmission",
				Aliases: []string{"ls-dirs", "ld"},
				Action: runListDirectories,
			},
			{
				Name:  "list-torrents",
				Usage: "List all torrent paths from Transmission",
				Aliases: []string{"ls-torrents", "lt"},
				Action: runListTorrents,
			},
		},
		Action: runCheck, // Default action when no subcommand is provided
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createClient(cmd *cli.Command) (*client.TransmissionClient, string, error) {
	cfg := types.Config{
		Host:     cmd.String("host"),
		Port:     cmd.Int("port"),
		User:     cmd.String("user"),
		Password: cmd.String("password"),
	}

	client := client.NewTransmissionClient(cfg)
	sessionID, err := client.GetSessionID()
	if err != nil {
		return nil, "", fmt.Errorf("error getting session ID: %w", err)
	}

	return client, sessionID, nil
}

func runCheck(ctx context.Context, cmd *cli.Command) error {
	dirs := cmd.StringSlice("dir")
	outputFile := cmd.String("output")

	// If no directories specified, use current directory
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	client, sessionID, err := createClient(cmd)
	if err != nil {
		return err
	}

	// Get all torrents from Transmission
	torrents, err := client.GetTorrents(sessionID)
	if err != nil {
		return fmt.Errorf("error getting torrents: %w", err)
	}

	// Create a map of torrent names for quick lookup
	torrentMap := make(map[string]bool)
	for _, t := range torrents {
		torrentMap[strings.ToLower(t.Name)] = true
	}

	fmt.Printf("Found %d torrents in Transmission\n\n", len(torrents))

	totalItems := 0
	totalFound := 0
	totalMissingSize := int64(0)
	var missingPaths []string

	// Check each directory
	for dirIdx, dir := range dirs {
		if dirIdx > 0 {
			fmt.Println()
		}

		// List directory contents
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", dir, err)
			continue
		}

		fmt.Printf("Directory: %s\n", dir)
		fmt.Println(strings.Repeat("-", 80))

		found := 0
		missingSize := int64(0)

		for _, entry := range entries {
			name := entry.Name()
			inTransmission := torrentMap[strings.ToLower(name)]

			var entryType string
			if entry.IsDir() {
				entryType = "[DIR] "
			} else {
				entryType = "[FILE]"
			}

			status := "✗"
			if inTransmission {
				status = "✓"
				found++
			} else {
				// Get size for missing items
				fullPath := filepath.Join(dir, name)

				// Get absolute path
				absPath, err := filepath.Abs(fullPath)
				if err != nil {
					absPath = fullPath
				}
				missingPaths = append(missingPaths, absPath)

				size, err := utils.GetSize(fullPath)
				if err == nil {
					missingSize += size
				}
			}

			fmt.Printf("%s %s %s\n", status, entryType, name)
		}

		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("Directory Summary: %d/%d items found in Transmission\n", found, len(entries))
		if missingSize > 0 {
			fmt.Printf("Missing items total size: %s\n", utils.FormatSize(missingSize))
		}

		totalItems += len(entries)
		totalFound += found
		totalMissingSize += missingSize
	}

	// Overall summary if multiple directories
	if len(dirs) > 1 {
		fmt.Println()
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("Overall Summary: %d/%d items found in Transmission across %d directories\n",
			totalFound, totalItems, len(dirs))
		if totalMissingSize > 0 {
			fmt.Printf("Total missing items size: %s\n", utils.FormatSize(totalMissingSize))
		}
	}

	// Write missing paths to output file if specified
	if outputFile != "" {
		err := utils.WriteMissingPaths(outputFile, missingPaths)
		if err != nil {
			return fmt.Errorf("error writing to output file: %w", err)
		}
		fmt.Printf("\nWrote %d missing item paths to: %s\n", len(missingPaths), outputFile)
	}

	return nil
}

func runListDirectories(ctx context.Context, cmd *cli.Command) error {
	client, sessionID, err := createClient(cmd)
	if err != nil {
		return err
	}

	return client.ListDownloadDirectories(sessionID)
}

func runListTorrents(ctx context.Context, cmd *cli.Command) error {
	client, sessionID, err := createClient(cmd)
	if err != nil {
		return err
	}

	paths, err := client.GetAllTorrentPaths(sessionID)
	if err != nil {
		return fmt.Errorf("error getting all torrent paths: %w", err)
	}

	// Output each path on its own line
	for _, path := range paths {
		fmt.Println(path)
	}

	return nil
}