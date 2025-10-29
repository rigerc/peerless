package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"peerless/pkg/client"
	"peerless/pkg/constants"
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
				Value:   constants.DefaultPort,
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
					&cli.BoolFlag{
						Name:    "rm",
						Aliases: []string{"delete", "remove"},
						Usage:   "Delete missing files after confirmation (DESTRUCTIVE)",
					},
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"dry", "simulate"},
						Usage:   "Show what would be deleted without actually deleting files",
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

func createClient(ctx context.Context, cmd *cli.Command) (*client.TransmissionClient, string, error) {
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
	if port < constants.MinPort || port > constants.MaxPort {
		return nil, "", fmt.Errorf("invalid port %d: port must be between %d and %d", port, constants.MinPort, constants.MaxPort)
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

	sessionID, err := client.GetSessionID(ctx)
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
	deleteMissing := cmd.Bool("rm")
	dryRun := cmd.Bool("dry-run")

	// If no directories specified, use current directory
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	// Validate conflicting options
	if deleteMissing && dryRun {
		output.PrintError("❌ Cannot use --rm and --dry-run together")
		output.PrintInfo("💡 Use --dry-run to preview what would be deleted, then use --rm to actually delete")
		return fmt.Errorf("conflicting options: --rm and --dry-run cannot be used together")
	}

	output.Logger.Info("Starting directory check", "directories", dirs)

	client, sessionID, err := createClient(ctx, cmd)
	if err != nil {
		return err
	}

	// Get all torrents from Transmission
	torrents, err := client.GetTorrents(ctx, sessionID)
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
		output.PrintSeparator(constants.SeparatorWidth)

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

		output.PrintSeparator(constants.SeparatorWidth)
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
		output.PrintSeparator(constants.SeparatorWidth)
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

	// Handle deletion of missing files if requested
	if (deleteMissing || dryRun) && len(missingPaths) > 0 {
		if dryRun {
			fmt.Println()
			output.PrintInfo("🔍 DRY RUN MODE - No files will actually be deleted")
			fmt.Println()
		} else {
			fmt.Println()
			output.PrintWarning("⚠️  DELETE MODE ENABLED - This will permanently delete files!")
			fmt.Println()
		}

		// Show what will be deleted
		headerText := "Files and directories to be deleted:"
		if dryRun {
			headerText = "Files and directories that WOULD be deleted:"
		}
		output.PrintError(headerText)

		for i, path := range missingPaths {
			// Get file info for display
			if info, err := os.Stat(path); err == nil {
				sizeStr := ""
				if !info.IsDir() {
					sizeStr = fmt.Sprintf(" (%s)", utils.FormatSize(info.Size()))
				}
				fmt.Printf("  %d. %s%s\n", i+1, path, sizeStr)
			} else {
				fmt.Printf("  %d. %s (error getting info)\n", i+1, path)
			}
		}
		fmt.Println()

		// Calculate total size
		var totalSize int64
		for _, path := range missingPaths {
			if size, err := utils.GetSize(path); err == nil {
				totalSize += size
			}
		}

		actionText := "Total to delete:"
		if dryRun {
			actionText = "Total that would be deleted:"
		}
		fmt.Printf("%s %d items (%s)\n", actionText, len(missingPaths), utils.FormatSize(totalSize))
		fmt.Println()

		if dryRun {
			// In dry run mode, just show what would happen
			output.PrintInfo("🔍 DRY RUN COMPLETED - No files were actually deleted")
			fmt.Println()
			output.PrintSuccess("💡 To actually delete these files, run the same command with --rm instead of --dry-run")
		} else {
			// Ask for confirmation for actual deletion
			fmt.Print("❓ Are you sure you want to delete these files? This action cannot be undone! (yes/No): ")
			var response string
			fmt.Scanln(&response)

			response = strings.ToLower(strings.TrimSpace(response))
			if response == "yes" || response == "y" {
				fmt.Println()
				output.PrintWarning("Deleting files...")

				var deletedCount int
				var deletedSize int64
				var failedDeletions []string

				for _, path := range missingPaths {
					output.Logger.Debug("Attempting to delete", "path", path)

					var err error
					var size int64

					// Get file info once and use it for both size and deletion
					var info os.FileInfo
					info, err = os.Stat(path)
					if err == nil {
						if !info.IsDir() {
							size = info.Size()
						}

						// Attempt deletion
						if info.IsDir() {
							err = os.RemoveAll(path) // Remove directory and contents
						} else {
							err = os.Remove(path) // Remove file
						}
					} else {
						err = fmt.Errorf("file not found: %v", err)
					}

					if err != nil {
						output.Logger.Error("Failed to delete", "path", path, "error", err)
						output.PrintError(fmt.Sprintf("❌ Failed to delete %s: %v", path, err))
						failedDeletions = append(failedDeletions, path)
					} else {
						output.Logger.Debug("Successfully deleted", "path", path)
						deletedCount++
						deletedSize += size
					}
				}

				fmt.Println()
				if deletedCount > 0 {
					output.PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d items (%s)", deletedCount, utils.FormatSize(deletedSize)))
				}

				if len(failedDeletions) > 0 {
					fmt.Println()
					output.PrintError(fmt.Sprintf("❌ Failed to delete %d items:", len(failedDeletions)))
					for _, path := range failedDeletions {
						fmt.Printf("  • %s\n", path)
					}
				}

				if len(failedDeletions) == 0 && deletedCount > 0 {
					fmt.Println()
					output.PrintSuccess("🎉 All missing files deleted successfully!")
				}
			} else {
				fmt.Println()
				output.PrintInfo("❌ Deletion cancelled by user")
			}
		}
	} else if (deleteMissing || dryRun) && len(missingPaths) == 0 {
		fmt.Println()
		output.PrintSuccess("✅ No missing files found - nothing to delete!")
	}

	output.Logger.Info("Directory check completed successfully")

	return nil
}

func runListDirectories(ctx context.Context, cmd *cli.Command) error {
	outputFile := cmd.String("output")
	output.Logger.Info("Starting directory listing command")

	client, sessionID, err := createClient(ctx, cmd)
	if err != nil {
		return err
	}

	output.Logger.Info("Retrieving download directories from Transmission")
	dirs, err := client.GetDownloadDirectories(ctx, sessionID)
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
		output.PrintSeparator(constants.SeparatorWidth)

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

	client, sessionID, err := createClient(ctx, cmd)
	if err != nil {
		return err
	}

	output.Logger.Info("Retrieving all torrent paths from Transmission")
	paths, err := client.GetAllTorrentPaths(ctx, sessionID)
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