package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"peerless/pkg/client"
	"peerless/pkg/output"
	"peerless/pkg/types"
	"peerless/pkg/utils"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "peerless",
		Usage: "Peerless - check local directories against Transmission torrents",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Usage:   "Transmission host (required)",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"po"},
				Value:   9091,
				Usage:   "Transmission port",
			},
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "Transmission username (required)",
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "Transmission password (required)",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose logging output",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Enable debug logging output",
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output file for directory list",
					},
				},
				Action: runListDirectories,
			},
			{
				Name:  "list-torrents",
				Usage: "List all torrent paths from Transmission",
				Aliases: []string{"ls-torrents", "lt"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output file for torrent paths",
					},
				},
				Action: runListTorrents,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return cli.ShowAppHelp(cmd)
		}, // Show help when no subcommand is provided
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		output.Logger.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func setupLogging(cmd *cli.Command) {
	debug := cmd.Bool("debug")
	verbose := cmd.Bool("verbose")

	if debug {
		output.Logger.SetLevel(log.DebugLevel)
	} else if verbose {
		output.Logger.SetLevel(log.InfoLevel)
	} else {
		output.Logger.SetLevel(log.ErrorLevel) // Only show errors by default
	}
}

func createClient(cmd *cli.Command) (*client.TransmissionClient, string, error) {
	setupLogging(cmd)

	// Validate mandatory fields
	host := cmd.String("host")
	user := cmd.String("user")
	password := cmd.String("password")

	if host == "" {
		return nil, "", fmt.Errorf("host (-H/--host) is required")
	}
	if user == "" {
		return nil, "", fmt.Errorf("username (-u/--user) is required")
	}
	if password == "" {
		return nil, "", fmt.Errorf("password (-p/--password) is required")
	}

	port := cmd.Int("port")

	// Validate port range
	if port <= 0 || port > 65535 {
		return nil, "", fmt.Errorf("invalid port %d: port must be between 1 and 65535", port)
	}

	// Validate host format
	if strings.TrimSpace(host) == "" {
		return nil, "", fmt.Errorf("host cannot be empty")
	}

	cfg := types.Config{
		Host:     strings.TrimSpace(host),
		Port:     port,
		User:     user,
		Password: password,
	}

	output.Logger.Info("Connecting to Transmission",
		"host", cfg.Host,
		"port", cfg.Port,
		"authenticated", cfg.User != "")

	client := client.NewTransmissionClient(cfg)
	output.Logger.Debug("Created Transmission client")

	sessionID, err := client.GetSessionID()
	if err != nil {
		output.Logger.Error("Failed to get session ID", "error", err)

		// Provide enhanced error messages for common issues
		errMsg := err.Error()
		if strings.Contains(errMsg, "authentication failed") {
			return nil, "", fmt.Errorf("authentication failed: please check your username and password for Transmission at %s:%d. %w", cfg.Host, cfg.Port, err)
		} else if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "connect: connection refused") {
			return nil, "", fmt.Errorf("cannot connect to Transmission at %s:%d. Please ensure:\n1. Transmission is running\n2. RPC interface is enabled\n3. Host and port are correct\nOriginal error: %w", cfg.Host, cfg.Port, err)
		} else if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "name resolution") {
			return nil, "", fmt.Errorf("cannot resolve host '%s'. Please check the hostname and ensure DNS is working correctly. %w", cfg.Host, err)
		} else if strings.Contains(errMsg, "timeout") {
			return nil, "", fmt.Errorf("connection timeout to Transmission at %s:%d. Please check network connectivity and firewall settings. %w", cfg.Host, cfg.Port, err)
		} else if strings.Contains(errMsg, "RPC endpoint not found") {
			return nil, "", fmt.Errorf("Transmission RPC interface not available at %s:%d. Please enable RPC in Transmission settings. %w", cfg.Host, cfg.Port, err)
		} else {
			return nil, "", fmt.Errorf("failed to connect to Transmission at %s:%d: %w", cfg.Host, cfg.Port, err)
		}
	}

	output.Logger.Debug("Successfully obtained session ID")
	return client, sessionID, nil
}

func runCheck(ctx context.Context, cmd *cli.Command) error {
	dirs := cmd.StringSlice("dir")
	outputFile := cmd.String("output")

	// If no directories specified, use current directory
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	output.Logger.Info("Starting directory check", "directories", dirs)

	client, sessionID, err := createClient(cmd)
	if err != nil {
		return err
	}

	// Get all torrents from Transmission
	torrents, err := client.GetTorrents(sessionID)
	if err != nil {
		output.Logger.Error("Failed to get torrents", "error", err)
		return fmt.Errorf("error getting torrents: %w", err)
	}

	output.Logger.Info("Retrieved torrents from Transmission", "count", len(torrents))

	// Create a map of torrent names for quick lookup
	torrentMap := make(map[string]bool)
	for _, t := range torrents {
		torrentMap[strings.ToLower(t.Name)] = true
	}

	output.PrintSummary(fmt.Sprintf("Found %d torrents in Transmission", len(torrents)))
	fmt.Println()

	totalItems := 0
	totalFound := 0
	totalMissingSize := int64(0)
	var missingPaths []string

	// Check each directory
	for dirIdx, dir := range dirs {
		if dirIdx > 0 {
			fmt.Println()
		}

		output.Logger.Debug("Checking directory", "path", dir)

		// List directory contents
		entries, err := os.ReadDir(dir)
		if err != nil {
			output.Logger.Error("Error reading directory", "directory", dir, "error", err)
			output.PrintError(fmt.Sprintf("Error reading directory %s: %v", dir, err))
			continue
		}

		output.PrintDirectoryHeader(dir)
		output.PrintSeparator(80)

		found := 0
		missingSize := int64(0)

		for _, entry := range entries {
			name := entry.Name()
			inTransmission := torrentMap[strings.ToLower(name)]

			if inTransmission {
				found++
				output.Logger.Debug("Found item in Transmission", "name", name)
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

				output.Logger.Debug("Missing item", "name", name, "size", size)
			}

			// Print with colors
			output.PrintTorrentStatus(inTransmission, name, entry.IsDir())
		}

		output.PrintSeparator(80)
		summary := fmt.Sprintf("Directory Summary: %d/%d items found in Transmission", found, len(entries))
		output.PrintSummary(summary)

		if missingSize > 0 {
			fmt.Print("Missing items total size: ")
			output.PrintSize(utils.FormatSize(missingSize))
			fmt.Println()
		}

		totalItems += len(entries)
		totalFound += found
		totalMissingSize += missingSize

		output.Logger.Debug("Directory check completed",
			"directory", dir,
			"total", len(entries),
			"found", found,
			"missing_size", missingSize)
	}

	// Overall summary if multiple directories
	if len(dirs) > 1 {
		fmt.Println()
		output.PrintSeparator(80)
		summary := fmt.Sprintf("Overall Summary: %d/%d items found in Transmission across %d directories",
			totalFound, totalItems, len(dirs))
		output.PrintSummary(summary)

		if totalMissingSize > 0 {
			fmt.Print("Total missing items size: ")
			output.PrintSize(utils.FormatSize(totalMissingSize))
			fmt.Println()
		}

		output.Logger.Info("Overall check completed",
			"total_items", totalItems,
			"total_found", totalFound,
			"directories", len(dirs),
			"missing_size", totalMissingSize)
	}

	// Write missing paths to output file if specified
	if outputFile != "" {
		output.Logger.Info("Writing missing paths to file", "file", outputFile, "count", len(missingPaths))
		err := utils.WriteMissingPaths(outputFile, missingPaths)
		if err != nil {
			output.Logger.Error("Failed to write output file", "file", outputFile, "error", err)
			return fmt.Errorf("error writing to output file: %w", err)
		}
		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Wrote %d missing item paths to: %s", len(missingPaths), outputFile))
	}

	output.Logger.Info("Directory check completed successfully")

	return nil
}

func runListDirectories(ctx context.Context, cmd *cli.Command) error {
	outputFile := cmd.String("output")
	output.Logger.Info("Starting directory listing command")

	client, sessionID, err := createClient(cmd)
	if err != nil {
		return err
	}

	output.Logger.Info("Retrieving download directories from Transmission")
	dirs, err := client.GetDownloadDirectories(sessionID)
	if err != nil {
		output.Logger.Error("Failed to list directories", "error", err)
		return err
	}

	// Write to file if output flag is specified
	if outputFile != "" {
		output.Logger.Info("Writing directory list to file", "file", outputFile, "count", len(dirs))
		err := utils.WriteDirectoryList(outputFile, dirs)
		if err != nil {
			output.Logger.Error("Failed to write output file", "file", outputFile, "error", err)
			return fmt.Errorf("error writing to output file: %w", err)
		}
		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Wrote %d directories to: %s", len(dirs), outputFile))
	} else {
		// Display to console with styling
		output.PrintSummary(fmt.Sprintf("Download Directories in Transmission (%d unique)", len(dirs)))
		output.PrintSeparator(80)

		for _, d := range dirs {
			fmt.Printf("%s (%d torrents)\n", d.Path, d.Count)
		}
	}

	output.Logger.Info("Directory listing completed successfully")
	return nil
}

func runListTorrents(ctx context.Context, cmd *cli.Command) error {
	outputFile := cmd.String("output")
	output.Logger.Info("Starting torrent listing command")

	client, sessionID, err := createClient(cmd)
	if err != nil {
		return err
	}

	output.Logger.Info("Retrieving all torrent paths from Transmission")
	paths, err := client.GetAllTorrentPaths(sessionID)
	if err != nil {
		output.Logger.Error("Failed to get torrent paths", "error", err)
		return fmt.Errorf("error getting all torrent paths: %w", err)
	}

	output.Logger.Info("Found torrent paths", "count", len(paths))

	// Write to file if output flag is specified
	if outputFile != "" {
		output.Logger.Info("Writing torrent paths to file", "file", outputFile, "count", len(paths))
		err := utils.WriteMissingPaths(outputFile, paths)
		if err != nil {
			output.Logger.Error("Failed to write output file", "file", outputFile, "error", err)
			return fmt.Errorf("error writing to output file: %w", err)
		}
		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Wrote %d torrent paths to: %s", len(paths), outputFile))
	} else {
		// Display to console with styling
		for _, path := range paths {
			output.PrintPath(path)
		}
	}

	output.Logger.Info("Torrent listing completed successfully")
	return nil
}