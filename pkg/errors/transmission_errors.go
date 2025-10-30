package errors

import (
	"fmt"
	"net/http"
)

// TransmissionError represents an error from the Transmission RPC API
type TransmissionError struct {
	StatusCode int
	Host       string
	Port       int
	Message    string
	Err        error
}

func (e *TransmissionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s at %s:%d: %v", e.Message, e.Host, e.Port, e.Err)
	}
	return fmt.Sprintf("%s at %s:%d", e.Message, e.Host, e.Port)
}

func (e *TransmissionError) Unwrap() error {
	return e.Err
}

// NewTransmissionError creates a new TransmissionError from HTTP response
func NewTransmissionError(statusCode int, host string, port int, err error) *TransmissionError {
	var message string

	switch statusCode {
	case http.StatusUnauthorized:
		message = "authentication failed: invalid username or password"
	case http.StatusForbidden:
		message = "access forbidden: insufficient permissions"
	case http.StatusNotFound:
		message = "RPC endpoint not found. Ensure Transmission is running"
	case http.StatusConflict:
		message = "session conflict: invalid session ID"
	case http.StatusInternalServerError:
		message = "Transmission server error (500)"
	default:
		if statusCode >= 400 {
			message = fmt.Sprintf("HTTP %d error", statusCode)
		} else {
			message = "unknown error"
		}
	}

	return &TransmissionError{
		StatusCode: statusCode,
		Host:       host,
		Port:       port,
		Message:    message,
		Err:        err,
	}
}

// IsAuthenticationError checks if the error is authentication failure
func IsAuthenticationError(err error) bool {
	if te, ok := err.(*TransmissionError); ok {
		return te.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsConnectionError checks if the error is connection failure
func IsConnectionError(err error) bool {
	if te, ok := err.(*TransmissionError); ok {
		return te.StatusCode == 0
	}
	return false
}