package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"peerless/pkg/service"
	"peerless/pkg/utils"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// Color constants for better readability
const (
	ColorGreen   = lipgloss.Color("10")
	ColorRed     = lipgloss.Color("9")
	ColorYellow  = lipgloss.Color("11")
	ColorBlue    = lipgloss.Color("12")
	ColorMagenta = lipgloss.Color("13")
	ColorCyan    = lipgloss.Color("14")
	ColorGray    = lipgloss.Color("8")
	ColorWhite   = lipgloss.Color("15")
)

var (
	// Color styles
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGreen)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorRed)

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorYellow)

	InfoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBlue)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorMagenta)

	PathStyle = lipgloss.NewStyle().
			Foreground(ColorCyan)

	SizeStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	FoundStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGreen)

	MissingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorRed)

	// Specialized styles
	DirectoryHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorYellow)

	SeparatorStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	SummaryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	FileStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	DirStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorMagenta)

	// Status symbols
	SuccessSymbol = SuccessStyle.Render("✓")
	ErrorSymbol   = ErrorStyle.Render("✗")
	DirSymbol     = DirStyle.Render("[DIR]")
	FileSymbol    = FileStyle.Render("[FILE]")
)

// Initialize logger
var Logger *log.Logger

func init() {
	// Initialize logger with custom settings
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Prefix:          "peerless",
	})

	// Check if we're in a terminal that supports colors
	if !isTerminal() {
		disableColors()
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// disableColors disables all colored output for non-terminal environments
func disableColors() {
	SuccessStyle = SuccessStyle.UnsetBold().UnsetForeground()
	ErrorStyle = ErrorStyle.UnsetBold().UnsetForeground()
	WarningStyle = WarningStyle.UnsetBold().UnsetForeground()
	InfoStyle = InfoStyle.UnsetBold().UnsetForeground()
	HeaderStyle = HeaderStyle.UnsetBold().UnsetForeground()
	PathStyle = PathStyle.UnsetBold().UnsetForeground()
	SizeStyle = SizeStyle.UnsetBold().UnsetForeground()
	FoundStyle = FoundStyle.UnsetBold().UnsetForeground()
	MissingStyle = MissingStyle.UnsetBold().UnsetForeground()
	DirectoryHeaderStyle = DirectoryHeaderStyle.UnsetBold().UnsetForeground()
	SeparatorStyle = SeparatorStyle.UnsetBold().UnsetForeground()
	SummaryStyle = SummaryStyle.UnsetBold().UnsetForeground()
	FileStyle = FileStyle.UnsetBold().UnsetForeground()
	DirStyle = DirStyle.UnsetBold().UnsetForeground()

	// Status styles
	StatusTitleStyle = StatusTitleStyle.UnsetBold().UnsetForeground().UnsetUnderline()
	StatusLabelStyle = StatusLabelStyle.UnsetBold().UnsetForeground()
	StatusValueStyle = StatusValueStyle.UnsetBold().UnsetForeground()
	StatusHeaderStyle = StatusHeaderStyle.UnsetBold().UnsetForeground().UnsetBackground()
	StatusActiveStyle = StatusActiveStyle.UnsetBold().UnsetForeground()
	StatusInactiveStyle = StatusInactiveStyle.UnsetBold().UnsetForeground()
	StatusSpeedStyle = StatusSpeedStyle.UnsetBold().UnsetForeground()

	SuccessSymbol = "✓"
	ErrorSymbol = "✗"
	DirSymbol = "[DIR]"
	FileSymbol = "[FILE]"
}

// Helper functions for common output patterns

func PrintHeader(text string) {
	println(HeaderStyle.Render(text))
}

func PrintSeparator(width int) {
	separator := SeparatorStyle.Render(strings.Repeat("-", width))
	println(separator)
}

func PrintDirectoryHeader(dir string) {
	println(DirectoryHeaderStyle.Render("Directory: " + dir))
}

func PrintSummary(text string) {
	println(SummaryStyle.Render(text))
}

func PrintSuccess(text string) {
	println(SuccessStyle.Render(text))
}

func PrintError(text string) {
	println(ErrorStyle.Render(text))
}

func PrintWarning(text string) {
	println(WarningStyle.Render(text))
}

func PrintInfo(text string) {
	println(InfoStyle.Render(text))
}

func PrintPath(path string) {
	println(PathStyle.Render(path))
}

func PrintSize(size string) {
	print(SizeStyle.Render(size))
}

func PrintTorrentStatus(isFound bool, name string, isDir bool) {
	var statusSymbol string
	var entryType string

	if isFound {
		statusSymbol = SuccessSymbol
	} else {
		statusSymbol = ErrorSymbol
	}

	if isDir {
		entryType = DirSymbol + " "
	} else {
		entryType = FileSymbol
	}

	fmt.Printf("%s %s %s\n", statusSymbol, entryType, name)
}

// Status-specific styles
var (
	StatusTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorMagenta).
				Underline(true)

	StatusLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorCyan)

	StatusValueStyle = lipgloss.NewStyle().
				Foreground(ColorWhite)

	StatusHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorYellow).
				Background(lipgloss.Color("236"))

	StatusActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorGreen)

	StatusInactiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorGray)

	StatusSpeedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBlue)
)

// PrintStatusHeader prints a simple status header
func PrintStatusHeader(title string) {
	fmt.Println(StatusHeaderStyle.Render(title))
	fmt.Println()
}

// PrintCompactStatus prints a compact one-line status summary
func PrintCompactStatus(total, downloading, seeding, paused int, downloadSpeed, uploadSpeed int, totalSize, freeSpace int64) {
	// Torrent status
	status := fmt.Sprintf("%d torrents", total)
	if downloading > 0 {
		status += fmt.Sprintf(" (⬇️ %d)", downloading)
	}
	if seeding > 0 {
		status += fmt.Sprintf(" (⬆️ %d)", seeding)
	}
	if paused > 0 {
		status += fmt.Sprintf(" (⏸️ %d)", paused)
	}

	// Speeds
	speeds := ""
	if downloadSpeed > 0 || uploadSpeed > 0 {
		if downloadSpeed > 0 && uploadSpeed > 0 {
			speeds = fmt.Sprintf(" • %s ↓ / %s ↑", formatSpeed(downloadSpeed), formatSpeed(uploadSpeed))
		} else if downloadSpeed > 0 {
			speeds = fmt.Sprintf(" • %s ↓", formatSpeed(downloadSpeed))
		} else if uploadSpeed > 0 {
			speeds = fmt.Sprintf(" • %s ↑", formatSpeed(uploadSpeed))
		}
	}

	// Storage
	storage := ""
	if freeSpace > 0 {
		storage = fmt.Sprintf(" • %s free", formatSize(statusSize(freeSpace)))
	}

	fmt.Printf("%s%s%s\n\n", StatusValueStyle.Render(status), StatusSpeedStyle.Render(speeds), StatusValueStyle.Render(storage))
}

// PrintStatusSummary prints a concise status summary
func PrintStatusSummary(total, downloading, seeding, paused int, downloadSpeed, uploadSpeed int, totalSize, downloadedSize, remainingSize, freeSpace int64) {
	// Torrent counts in one line
	fmt.Printf("Torrents: %d", total)
	if downloading > 0 {
		fmt.Printf(" • %s downloading", StatusActiveStyle.Render(fmt.Sprintf("%d", downloading)))
	}
	if seeding > 0 {
		fmt.Printf(" • %s seeding", StatusActiveStyle.Render(fmt.Sprintf("%d", seeding)))
	}
	if paused > 0 {
		fmt.Printf(" • %s paused", WarningStyle.Render(fmt.Sprintf("%d", paused)))
	}
	fmt.Println()

	// Progress
	if totalSize > 0 {
		percent := float64(downloadedSize) / float64(totalSize) * 100
		fmt.Printf("Progress: %.1f%% • %s / %s", percent,
			StatusValueStyle.Render(formatSize(statusSize(downloadedSize))),
			StatusValueStyle.Render(formatSize(statusSize(totalSize))))
		if remainingSize > 0 {
			fmt.Printf(" • %s remaining", StatusValueStyle.Render(formatSize(statusSize(remainingSize))))
		}
		fmt.Println()
	}

	// Speeds
	if downloadSpeed > 0 || uploadSpeed > 0 {
		fmt.Print("Speed: ")
		if downloadSpeed > 0 {
			fmt.Printf("%s ↓", StatusSpeedStyle.Render(formatSpeed(downloadSpeed)))
		}
		if downloadSpeed > 0 && uploadSpeed > 0 {
			fmt.Print(" • ")
		}
		if uploadSpeed > 0 {
			fmt.Printf("%s ↑", StatusSpeedStyle.Render(formatSpeed(uploadSpeed)))
		}
		fmt.Println()
	}

	// Storage
	if freeSpace > 0 {
		fmt.Printf("Free Space: %s\n", StatusValueStyle.Render(formatSize(statusSize(freeSpace))))
	}
	fmt.Println()
}

// PrintSpeedInfo prints download/upload speeds
func PrintSpeedItem(label string, speed int) {
	if speed > 0 {
		formattedSpeed := formatSpeed(speed)
		fmt.Printf("  %s %s\n", StatusLabelStyle.Render(label+":"), StatusSpeedStyle.Render(formattedSpeed))
	} else {
		fmt.Printf("  %s %s\n", StatusLabelStyle.Render(label+":"), StatusInactiveStyle.Render("0 B/s"))
	}
}

// PrintSimpleDirectoryList prints a simple directory list
func PrintSimpleDirectoryList(breakdown map[string]service.DirectoryStatus) {
	if len(breakdown) == 0 {
		return
	}

	fmt.Print("Directories: ")
	i := 0
	for dir, status := range breakdown {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s (%d)", filepath.Base(dir), status.TorrentCount)
		i++
		if i >= 3 { // Limit to first 3 directories
			fmt.Printf(" + %d more", len(breakdown)-3)
			break
		}
	}
	fmt.Println()
}

// Helper types and functions for status display
type statusSize int64

func formatSize(s statusSize) string {
	return utils.FormatSize(int64(s))
}

func formatSpeed(bytesPerSecond int) string {
	if bytesPerSecond == 0 {
		return "0 B/s"
	}

	const unit = 1024
	if bytesPerSecond < unit {
		return fmt.Sprintf("%d B/s", bytesPerSecond)
	}

	div, exp := int64(unit), 0
	for n := bytesPerSecond / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB/s", "MB/s", "GB/s", "TB/s"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	value := float64(bytesPerSecond) / float64(div)
	return fmt.Sprintf("%.1f %s", value, units[exp])
}

func formatDuration(seconds int) string {
	if seconds == 0 {
		return "0s"
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	} else {
		return fmt.Sprintf("%ds", secs)
	}
}
