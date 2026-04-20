package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

// RequestOption mutates an HTTP request before it is served in a test.
type RequestOption func(t *testing.T, req *http.Request)

// RenderPage serves an HTTP request through e, checks the expected status, and parses the HTML body.
func RenderPage(t *testing.T, e *echo.Echo, method, target string, body io.Reader, expectedStatus int, opts ...RequestOption) (*http.Response, string) {
	t.Helper()

	req := httptest.NewRequest(method, target, body)
	for _, opt := range opts {
		opt(t, req)
	}
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, rec.Code)
	}

	responseBody := rec.Body.String()
	return rec.Result(), responseBody
}
