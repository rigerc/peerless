package types

import (
	"fmt"
	"net"
	"strings"

	"peerless/pkg/constants"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors ValidationErrors

	if err := c.ValidateHost(); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			errors = append(errors, *ve)
		}
	}

	if err := c.ValidatePort(); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			errors = append(errors, *ve)
		}
	}

	if err := c.ValidateAuth(); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			errors = append(errors, *ve)
		}
	}

	if err := c.ValidateDirs(); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			errors = append(errors, *ve)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateHost validates the host configuration
func (c *Config) ValidateHost() error {
	if c.Host == "" {
		return &ValidationError{Field: "host", Message: "host is required"}
	}

	trimmed := strings.TrimSpace(c.Host)
	if trimmed == "" {
		return &ValidationError{Field: "host", Message: "host cannot be empty or whitespace"}
	}

	if net.ParseIP(trimmed) == nil {
		if !isValidHostname(trimmed) {
			return &ValidationError{Field: "host", Message: "host must be a valid IP address or hostname"}
		}
	}

	c.Host = trimmed
	return nil
}

// ValidatePort validates the port configuration
func (c *Config) ValidatePort() error {
	if c.Port < constants.MinPort || c.Port > constants.MaxPort {
		return &ValidationError{
			Field:   "port",
			Message: fmt.Sprintf("port must be between %d and %d, got %d", constants.MinPort, constants.MaxPort, c.Port),
		}
	}
	return nil
}

// ValidateAuth validates the authentication configuration
func (c *Config) ValidateAuth() error {
	if c.User != "" && c.Password == "" {
		return &ValidationError{Field: "password", Message: "password is required when username is provided"}
	}

	// Check for weak passwords
	if c.Password != "" {
		weakPasswords := []string{"password", "123456", "admin", "guest"}
		for _, weak := range weakPasswords {
			if strings.ToLower(c.Password) == weak {
				return &ValidationError{
					Field:   "password",
					Message: fmt.Sprintf("weak password detected: '%s' should not be used", weak),
				}
			}
		}
	}

	return nil
}

// ValidateDirs validates the directories configuration
func (c *Config) ValidateDirs() error {
	if len(c.Dirs) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	for _, dir := range c.Dirs {
		if seen[dir] {
			return &ValidationError{Field: "dirs", Message: fmt.Sprintf("duplicate directory: %s", dir)}
		}
		seen[dir] = true
	}

	return nil
}

// isValidHostname checks if a string is a valid hostname
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Check for consecutive dots
	for i := 0; i < len(hostname)-1; i++ {
		if hostname[i] == '.' && hostname[i+1] == '.' {
			return false
		}
	}

	for i, char := range hostname {
		if char == '.' {
			continue
		}
		if char == '-' && i > 0 && i < len(hostname)-1 {
			continue
		}
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') {
			return false
		}
	}

	return true
}

// SetDefaults sets default values for optional configuration fields
func (c *Config) SetDefaults() {
	if c.Port == 0 {
		c.Port = constants.DefaultPort
	}
}
