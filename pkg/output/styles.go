package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	// Color styles
	SuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("10")) // Green

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9")) // Red

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("11")) // Yellow

	InfoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")) // Blue

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("13")) // Magenta

	PathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")) // Cyan

	SizeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // Bright Black/Gray

	FoundStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")) // Green

	MissingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9")) // Red

	// Specialized styles
	DirectoryHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("11")) // Yellow

	SeparatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // Gray

	SummaryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")) // Cyan

	FileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")) // White

	DirStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("13")) // Magenta

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
	fileInfo, _ := os.Stdout.Stat()
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