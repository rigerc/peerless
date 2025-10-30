package service

import (
	"bytes"
	"io"
	"net/http"
)

// MockHTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// NewMockResponse creates a mock HTTP response
func NewMockResponse(statusCode int, body string, headers map[string]string) *http.Response {
	resp := &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}

	for key, value := range headers {
		resp.Header.Set(key, value)
	}

	return resp
}
