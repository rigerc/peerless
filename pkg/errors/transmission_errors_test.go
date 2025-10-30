package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTransmissionError(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected string
	}{
		{
			name:     "401 unauthorized",
			status:   http.StatusUnauthorized,
			expected: "authentication failed: invalid username or password",
		},
		{
			name:     "403 forbidden",
			status:   http.StatusForbidden,
			expected: "access forbidden: insufficient permissions",
		},
		{
			name:     "404 not found",
			status:   http.StatusNotFound,
			expected: "RPC endpoint not found. Ensure Transmission is running",
		},
		{
			name:     "409 conflict",
			status:   http.StatusConflict,
			expected: "session conflict: invalid session ID",
		},
		{
			name:     "500 server error",
			status:   http.StatusInternalServerError,
			expected: "Transmission server error (500)",
		},
		{
			name:     "418 custom error",
			status:   418,
			expected: "HTTP 418 error",
		},
		{
			name:     "200 success",
			status:   200,
			expected: "unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewTransmissionError(tt.status, "localhost", 9091, nil)
			assert.Equal(t, tt.status, err.StatusCode)
			assert.Equal(t, "localhost", err.Host)
			assert.Equal(t, 9091, err.Port)
			assert.Equal(t, tt.expected, err.Message)
			assert.Nil(t, err.Err)
		})
	}
}

func TestTransmissionError_Error(t *testing.T) {
	t.Run("error without underlying error", func(t *testing.T) {
		err := &TransmissionError{
			StatusCode: 401,
			Host:       "localhost",
			Port:       9091,
			Message:    "authentication failed",
			Err:        nil,
		}

		expected := "authentication failed at localhost:9091"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("error with underlying error", func(t *testing.T) {
		underlying := &TransmissionError{
			StatusCode: 0,
			Host:       "localhost",
			Port:       9091,
			Message:    "connection refused",
			Err:        nil,
		}

		err := &TransmissionError{
			StatusCode: 401,
			Host:       "localhost",
			Port:       9091,
			Message:    "authentication failed",
			Err:        underlying,
		}

		expected := "authentication failed at localhost:9091: connection refused at localhost:9091"
		assert.Equal(t, expected, err.Error())
	})
}

func TestTransmissionError_Unwrap(t *testing.T) {
	t.Run("error without underlying error", func(t *testing.T) {
		err := &TransmissionError{
			StatusCode: 401,
			Host:       "localhost",
			Port:       9091,
			Message:    "authentication failed",
			Err:        nil,
		}

		assert.Nil(t, err.Unwrap())
	})

	t.Run("error with underlying error", func(t *testing.T) {
		underlying := &TransmissionError{
			StatusCode: 0,
			Host:       "localhost",
			Port:       9091,
			Message:    "connection refused",
			Err:        nil,
		}

		err := &TransmissionError{
			StatusCode: 401,
			Host:       "localhost",
			Port:       9091,
			Message:    "authentication failed",
			Err:        underlying,
		}

		assert.Equal(t, underlying, err.Unwrap())
	})
}

func TestIsAuthenticationError(t *testing.T) {
	t.Run("authentication error", func(t *testing.T) {
		err := NewTransmissionError(http.StatusUnauthorized, "localhost", 9091, nil)
		assert.True(t, IsAuthenticationError(err))
	})

	t.Run("non-authentication error", func(t *testing.T) {
		err := NewTransmissionError(http.StatusNotFound, "localhost", 9091, nil)
		assert.False(t, IsAuthenticationError(err))
	})

	t.Run("different error type", func(t *testing.T) {
		err := assert.AnError
		assert.False(t, IsAuthenticationError(err))
	})
}

func TestIsConnectionError(t *testing.T) {
	t.Run("connection error", func(t *testing.T) {
		err := NewTransmissionError(0, "localhost", 9091, nil)
		assert.True(t, IsConnectionError(err))
	})

	t.Run("non-connection error", func(t *testing.T) {
		err := NewTransmissionError(http.StatusUnauthorized, "localhost", 9091, nil)
		assert.False(t, IsConnectionError(err))
	})

	t.Run("different error type", func(t *testing.T) {
		err := assert.AnError
		assert.False(t, IsConnectionError(err))
	})
}