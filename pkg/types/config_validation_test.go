package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"peerless/pkg/constants"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := Config{
			Host:     "localhost",
			Port:     9091,
			User:     "admin",
			Password: "securepassword123",
			Dirs:     []string{"/downloads", "/movies"},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		config := Config{
			Host:     "",
			Port:     99999,
			User:     "admin",
			Password: "",
			Dirs:     []string{"/downloads", "/downloads"}, // duplicate
		}

		err := config.Validate()
		assert.Error(t, err)

		validationErrors, ok := err.(ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrors, 4) // host, port, password, dirs
	})
}

func TestConfig_ValidateHost(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid IPv4",
			host:        "192.168.1.100",
			expectError: false,
		},
		{
			name:        "valid IPv6",
			host:        "::1",
			expectError: false,
		},
		{
			name:        "valid hostname",
			host:        "transmission.local",
			expectError: false,
		},
		{
			name:        "empty host",
			host:        "",
			expectError: true,
			errorMsg:    "host is required",
		},
		{
			name:        "whitespace only",
			host:        "   ",
			expectError: true,
			errorMsg:    "host cannot be empty or whitespace",
		},
		{
			name:        "host with surrounding whitespace",
			host:        "  localhost  ",
			expectError: false,
		},
		{
			name:        "invalid hostname with special chars",
			host:        "invalid@hostname",
			expectError: true,
			errorMsg:    "host must be a valid IP address or hostname",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Host: tt.host}
			err := config.ValidateHost()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				// Check that whitespace was trimmed
				assert.Equal(t, strings.TrimSpace(tt.host), config.Host)
			}
		})
	}
}

func TestConfig_ValidatePort(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		expectError bool
	}{
		{
			name:        "valid port",
			port:        9091,
			expectError: false,
		},
		{
			name:        "minimum valid port",
			port:        constants.MinPort,
			expectError: false,
		},
		{
			name:        "maximum valid port",
			port:        constants.MaxPort,
			expectError: false,
		},
		{
			name:        "port too low",
			port:        0,
			expectError: true,
		},
		{
			name:        "port too high",
			port:        99999,
			expectError: true,
		},
		{
			name:        "negative port",
			port:        -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Port: tt.port}
			err := config.ValidatePort()

			if tt.expectError {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ValidateAuth(t *testing.T) {
	tests := []struct {
		name        string
		user        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid auth",
			user:        "admin",
			password:    "securepassword123",
			expectError: false,
		},
		{
			name:        "no auth provided",
			user:        "",
			password:    "",
			expectError: false,
		},
		{
			name:        "username without password",
			user:        "admin",
			password:    "",
			expectError: true,
			errorMsg:    "password is required when username is provided",
		},
		{
			name:        "weak password - password",
			user:        "admin",
			password:    "password",
			expectError: true,
			errorMsg:    "weak password detected",
		},
		{
			name:        "weak password - 123456",
			user:        "admin",
			password:    "123456",
			expectError: true,
			errorMsg:    "weak password detected",
		},
		{
			name:        "weak password - admin",
			user:        "admin",
			password:    "admin",
			expectError: true,
			errorMsg:    "weak password detected",
		},
		{
			name:        "weak password - guest",
			user:        "admin",
			password:    "guest",
			expectError: true,
			errorMsg:    "weak password detected",
		},
		{
			name:        "case-insensitive weak password check",
			user:        "admin",
			password:    "PASSWORD",
			expectError: true,
			errorMsg:    "weak password detected",
		},
		{
			name:        "password without username",
			user:        "",
			password:    "somepassword",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				User:     tt.user,
				Password: tt.password,
			}
			err := config.ValidateAuth()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ValidateDirs(t *testing.T) {
	tests := []struct {
		name        string
		dirs        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty dirs",
			dirs:        []string{},
			expectError: false,
		},
		{
			name:        "single valid dir",
			dirs:        []string{"/downloads"},
			expectError: false,
		},
		{
			name:        "multiple valid dirs",
			dirs:        []string{"/downloads", "/movies", "/tv"},
			expectError: false,
		},
		{
			name:        "duplicate directories",
			dirs:        []string{"/downloads", "/movies", "/downloads"},
			expectError: true,
			errorMsg:    "duplicate directory: /downloads",
		},
		{
			name:        "multiple duplicates",
			dirs:        []string{"/downloads", "/downloads", "/downloads"},
			expectError: true,
			errorMsg:    "duplicate directory: /downloads",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Dirs: tt.dirs}
			err := config.ValidateDirs()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	t.Run("set default port", func(t *testing.T) {
		config := Config{
			Host: "localhost",
			Port: 0, // Uninitialized
		}

		config.SetDefaults()
		assert.Equal(t, constants.DefaultPort, config.Port)
	})

	t.Run("keep existing port", func(t *testing.T) {
		config := Config{
			Host: "localhost",
			Port: 8080,
		}

		config.SetDefaults()
		assert.Equal(t, 8080, config.Port)
	})
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "host",
		Message: "host is required",
	}

	expected := "host: host is required"
	assert.Equal(t, expected, err.Error())
}

func TestValidationErrors(t *testing.T) {
	t.Run("single error", func(t *testing.T) {
		errors := ValidationErrors{
			{Field: "host", Message: "host is required"},
		}

		expected := "host: host is required"
		assert.Equal(t, expected, errors.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		errors := ValidationErrors{
			{Field: "host", Message: "host is required"},
			{Field: "port", Message: "port must be between 1 and 65535"},
		}

		expected := "host: host is required; port: port must be between 1 and 65535"
		assert.Equal(t, expected, errors.Error())
	})

	t.Run("no errors", func(t *testing.T) {
		errors := ValidationErrors{}
		expected := "no validation errors"
		assert.Equal(t, expected, errors.Error())
	})
}

func TestIsValidHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		valid    bool
	}{
		{
			name:     "valid simple hostname",
			hostname: "transmission",
			valid:    true,
		},
		{
			name:     "valid hostname with subdomain",
			hostname: "transmission.local",
			valid:    true,
		},
		{
			name:     "valid hostname with hyphen",
			hostname: "my-server",
			valid:    true,
		},
		{
			name:     "empty hostname",
			hostname: "",
			valid:    false,
		},
		{
			name:     "hostname starting with hyphen",
			hostname: "-invalid",
			valid:    false,
		},
		{
			name:     "hostname ending with hyphen",
			hostname: "invalid-",
			valid:    false,
		},
		{
			name:     "hostname with invalid characters",
			hostname: "invalid@host",
			valid:    false,
		},
		{
			name:     "hostname with consecutive dots",
			hostname: "invalid..hostname",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidHostname(tt.hostname)
			assert.Equal(t, tt.valid, result)
		})
	}
}
