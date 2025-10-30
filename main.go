package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"peerless/pkg/client"
	"peerless/pkg/constants"
	"peerless/pkg/errors"
	"peerless/pkg/output"
	"peerless/pkg/service"
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
				Name:    "list-directories",
				Usage:   "List all download directories from Transmission",
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
				Name:    "list-torrents",
				Usage:   "List all torrent paths from Transmission",
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
			{
				Name:    "status",
				Usage:   "Show Transmission statistics and status information",
				Aliases: []string{"stat", "info"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "compact",
						Aliases: []string{"c"},
						Usage:   "Show compact status without detailed breakdown",
					},
				},
				Action: runStatus,
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

func createService(ctx context.Context, cmd *cli.Command) (*service.TorrentService, error) {
	setupLogging(cmd)

	// Create configuration
	cfg := types.Config{
		Host:     strings.TrimSpace(cmd.String("host")),
		Port:     cmd.Int("port"),
		User:     cmd.String("user"),
		Password: cmd.String("password"),
		Dirs:     cmd.StringSlice("dir"),
	}

	// Set defaults and validate configuration
	cfg.SetDefaults()
	if err := cfg.Validate(); err != nil {
		output.Logger.Error("Configuration validation failed", "error", err)
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	output.Logger.Info("Connecting to Transmission",
		"host", cfg.Host,
		"port", cfg.Port,
		"authenticated", cfg.User != "")

	// Create client and service
	client := client.NewTransmissionClient(cfg)
	svc := service.NewTorrentService(client)
	output.Logger.Debug("Created Transmission client and service")

	// Test connection by trying to get torrents
	_, err := client.GetTorrents(ctx)
	if err != nil {
		output.Logger.Error("Failed to connect to Transmission", "error", err)

		// Handle specific error types
		if errors.IsAuthenticationError(err) {
			return nil, fmt.Errorf("authentication failed: please check your username and password for Transmission at %s:%d. %w", cfg.Host, cfg.Port, err)
		} else if errors.IsConnectionError(err) {
			return nil, fmt.Errorf("cannot connect to Transmission at %s:%d. Please ensure:\n1. Transmission is running\n2. RPC interface is enabled\n3. Host and port are correct\nOriginal error: %w", cfg.Host, cfg.Port, err)
		} else {
			return nil, fmt.Errorf("failed to connect to Transmission at %s:%d: %w", cfg.Host, cfg.Port, err)
		}
	}

	output.Logger.Debug("Successfully connected to Transmission")
	return svc, nil
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

	svc, err := createService(ctx, cmd)
	if err != nil {
		return err
	}

	// Check directories using the service
	result, err := svc.CheckDirectories(ctx, dirs)
	if err != nil {
		output.Logger.Error("Failed to check directories", "error", err)
		return fmt.Errorf("error checking directories: %w", err)
	}

	output.Logger.Info("Directory check completed", "total_items", result.TotalItems, "total_found", result.TotalFound)
	output.PrintSummary(fmt.Sprintf("Found %d torrents in Transmission", result.TotalFound))
	fmt.Println()

	// Display results for each directory
	for i, dirResult := range result.Directories {
		if i > 0 {
			fmt.Println()
		}

		output.PrintDirectoryHeader(dirResult.Path)
		output.PrintSeparator(constants.SeparatorWidth)

		// List directory contents with status
		entries, err := os.ReadDir(dirResult.Path)
		if err != nil {
			output.Logger.Error("Error reading directory", "directory", dirResult.Path, "error", err)
			output.PrintError(fmt.Sprintf("Error reading directory %s: %v", dirResult.Path, err))
			continue
		}

		// Get torrent statistics for this directory (simplified approach)
		_, err = svc.GetTorrentStatistics(ctx)
		if err != nil {
			output.Logger.Error("Failed to get torrent statistics", "error", err)
			continue
		}

		for _, entry := range entries {
			name := entry.Name()
			// Check if this item is in the missing paths
			inTransmission := true
			for _, missingPath := range dirResult.MissingPaths {
				if filepath.Base(missingPath) == name {
					inTransmission = false
					break
				}
			}
			output.PrintTorrentStatus(inTransmission, name, entry.IsDir())
		}

		output.PrintSeparator(constants.SeparatorWidth)
		summary := fmt.Sprintf("Directory Summary: %d/%d items found in Transmission", dirResult.FoundItems, dirResult.TotalItems)
		output.PrintSummary(summary)

		if dirResult.MissingSize > 0 {
			fmt.Print("Missing items total size: ")
			output.PrintSize(utils.FormatSize(dirResult.MissingSize))
			fmt.Println()
		}
	}

	// Overall summary if multiple directories
	if len(dirs) > 1 {
		fmt.Println()
		output.PrintSeparator(constants.SeparatorWidth)
		summary := fmt.Sprintf("Overall Summary: %d/%d items found in Transmission across %d directories",
			result.TotalFound, result.TotalItems, len(dirs))
		output.PrintSummary(summary)

		if result.TotalMissingSize > 0 {
			fmt.Print("Total missing items size: ")
			output.PrintSize(utils.FormatSize(result.TotalMissingSize))
			fmt.Println()
		}

		// Show per-directory breakdown
		fmt.Println()
		output.PrintSummary("Per-Directory Breakdown:")
		for _, dirResult := range result.Directories {
			missingCount := dirResult.TotalItems - dirResult.FoundItems
			if missingCount > 0 {
				fmt.Printf("  %s: %d/%d missing (%.1f%%) - %s\n",
					dirResult.Path,
					missingCount,
					dirResult.TotalItems,
					float64(missingCount)/float64(dirResult.TotalItems)*100,
					utils.FormatSize(dirResult.MissingSize))
			} else {
				fmt.Printf("  %s: %d/%d found (100%%) - %s\n",
					dirResult.Path,
					dirResult.TotalItems,
					dirResult.TotalItems,
					utils.FormatSize(dirResult.MissingSize))
			}
		}
	}

	// Write missing paths to output file if specified
	if outputFile != "" {
		output.Logger.Info("Writing missing paths to file", "file", outputFile, "count", len(result.MissingPaths))
		err := utils.WriteMissingPaths(outputFile, result.MissingPaths)
		if err != nil {
			output.Logger.Error("Failed to write output file", "file", outputFile, "error", err)
			return fmt.Errorf("error writing to output file: %w", err)
		}
		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Wrote %d missing item paths to: %s", len(result.MissingPaths), outputFile))
	}

	// Handle deletion of missing files if requested
	if (deleteMissing || dryRun) && len(result.MissingPaths) > 0 {
		if dryRun {
			fmt.Println()
			output.PrintInfo("🔍 DRY RUN MODE - No files will actually be deleted")
			fmt.Println()
		} else {
			fmt.Println()
			output.PrintWarning("⚠️  DELETE MODE ENABLED - This will permanently delete files!")
			fmt.Println()
		}

		// Validate paths before deletion
		if err := utils.ValidateDeletionPaths(result.MissingPaths, dirs); err != nil {
			output.PrintError(fmt.Sprintf("❌ Path validation failed: %v", err))
			return fmt.Errorf("path validation failed: %w", err)
		}

		// Show what will be deleted
		headerText := "Files and directories to be deleted:"
		if dryRun {
			headerText = "Files and directories that WOULD be deleted:"
		}
		output.PrintError(headerText)

		// Get file operations info for display
		operations := utils.BatchFileInfo(result.MissingPaths)
		for i, op := range operations {
			if op.Error != nil {
				fmt.Printf("  %d. %s (error: %v)\n", i+1, op.Path, op.Error)
			} else {
				sizeStr := ""
				if op.IsDir {
					sizeStr = fmt.Sprintf(" (%s, directory)", utils.FormatSize(op.Size))
				} else {
					sizeStr = fmt.Sprintf(" (%s, file)", utils.FormatSize(op.Size))
				}
				fmt.Printf("  %d. %s%s\n", i+1, op.Path, sizeStr)
			}
		}
		fmt.Println()

		// Calculate total size using enhanced utility
		totalSize, inaccessibleItems, err := utils.CalculateTotalSize(result.MissingPaths)
		if err != nil {
			output.Logger.Warn("Failed to calculate total size", "error", err)
		}

		actionText := "Total to delete:"
		if dryRun {
			actionText = "Total that would be deleted:"
		}

		// Provide more informative total size display
		if inaccessibleItems > 0 {
			fmt.Printf("%s %d items (%s) - %d items inaccessible\n", actionText, len(result.MissingPaths), utils.FormatSize(totalSize), inaccessibleItems)
			fmt.Println("Note: Some items couldn't be sized due to permissions or other errors")
		} else {
			fmt.Printf("%s %d items (%s)\n", actionText, len(result.MissingPaths), utils.FormatSize(totalSize))
		}
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
			_, err := fmt.Scanln(&response)
			if err != nil {
				output.Logger.Warn("Failed to read input, cancelling deletion", "error", err)
				response = "no"
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response == "yes" || response == "y" {
				fmt.Println()
				output.PrintWarning("Deleting files...")

				// Use enhanced file operations with progress tracking
				deleteResult := utils.DeleteFiles(result.MissingPaths, func(current, total int, path string, size int64) {
					output.Logger.Debug("Deleting file", "current", current, "total", total, "path", path, "size", size)
				})

				fmt.Println()
				if deleteResult.SuccessCount > 0 {
					output.PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d items (%s)", deleteResult.SuccessCount, utils.FormatSize(deleteResult.TotalSize)))
				}

				if deleteResult.FailedCount > 0 {
					fmt.Println()
					output.PrintError(fmt.Sprintf("❌ Failed to delete %d items:", deleteResult.FailedCount))
					for _, failed := range deleteResult.Failed {
						fmt.Printf("  • %s: %v\n", failed.Path, failed.Error)
					}
				}

				if deleteResult.FailedCount == 0 && deleteResult.SuccessCount > 0 {
					fmt.Println()
					output.PrintSuccess("🎉 All missing files deleted successfully!")
				}
			} else {
				fmt.Println()
				output.PrintInfo("❌ Deletion cancelled by user")
			}
		}
	} else if (deleteMissing || dryRun) && len(result.MissingPaths) == 0 {
		fmt.Println()
		output.PrintSuccess("✅ No missing files found - nothing to delete!")
	}

	output.Logger.Info("Directory check completed successfully")

	return nil
}

func runListDirectories(ctx context.Context, cmd *cli.Command) error {
	outputFile := cmd.String("output")
	output.Logger.Info("Starting directory listing command")

	svc, err := createService(ctx, cmd)
	if err != nil {
		return err
	}

	output.Logger.Info("Retrieving download directories from Transmission")
	dirs, err := svc.GetDownloadDirectories(ctx)
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

	svc, err := createService(ctx, cmd)
	if err != nil {
		return err
	}

	output.Logger.Info("Retrieving all torrent paths from Transmission")
	paths, err := svc.GetAllTorrentPaths(ctx)
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

func runStatus(ctx context.Context, cmd *cli.Command) error {
	compact := cmd.Bool("compact")
	output.Logger.Info("Starting status command")

	svc, err := createService(ctx, cmd)
	if err != nil {
		return err
	}

	output.Logger.Info("Retrieving Transmission status information")
	status, err := svc.GetDetailedStatus(ctx)
	if err != nil {
		output.Logger.Error("Failed to get status", "error", err)
		return fmt.Errorf("error getting status: %w", err)
	}

	if compact {
		// Ultra-compact one-line output
		output.PrintCompactStatus(
			status.TotalTorrents,
			status.DownloadingTorrents,
			status.SeedingTorrents,
			status.PausedTorrents,
			status.TotalDownloadSpeed,
			status.TotalUploadSpeed,
			status.TotalSize,
			status.FreeSpace,
		)
	} else {
		// Concise multi-line output
		output.PrintStatusHeader("Transmission Status")
		output.PrintStatusSummary(
			status.TotalTorrents,
			status.DownloadingTorrents,
			status.SeedingTorrents,
			status.PausedTorrents,
			status.TotalDownloadSpeed,
			status.TotalUploadSpeed,
			status.TotalSize,
			status.DownloadedSize,
			status.RemainingSize,
			status.FreeSpace,
		)

		// Session info (single line)
		fmt.Printf("Directory: %s • Port: %s",
			output.PathStyle.Render(status.DownloadDir),
			fmt.Sprintf("%d", status.PeerPort))
		if status.AltSpeedEnabled {
			fmt.Printf(" • %s", output.WarningStyle.Render("Speed limits"))
		}
		fmt.Println()

		// Directory breakdown (simplified)
		if len(status.DirectoryBreakdown) > 1 {
			output.PrintSimpleDirectoryList(status.DirectoryBreakdown)
		}
	}

	output.Logger.Info("Status command completed successfully")
	return nil
}
