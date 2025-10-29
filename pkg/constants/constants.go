package constants

import "time"

// Network and HTTP constants
const (
	// Default transmission port
	DefaultPort = 9091

	// HTTP timeout duration
	HTTPTimeout = 30 * time.Second

	// Port range limits
	MinPort = 1
	MaxPort = 65535
)

// File system constants
const (
	// File size units in bytes
	BytesPerKB = 1024
	BytesPerMB = 1024 * 1024
	BytesPerGB = 1024 * 1024 * 1024
	BytesPerTB = 1024 * 1024 * 1024 * 1024
	BytesPerPB = 1024 * 1024 * 1024 * 1024 * 1024
)

// Display constants
const (
	// Separator width for terminal output
	SeparatorWidth = 80

	// File size unit names
	SizeUnits = "KBMBGBTBPB"
)

// Unicode control characters to filter out
const (
	LTRMark = '\u200E'
	RTLMark = '\u200F'
	LRE     = '\u202A'
	RLE     = '\u202B'
	PDF     = '\u202C'
	LRO     = '\u202D'
	RLO     = '\u202E'
)