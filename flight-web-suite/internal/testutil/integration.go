package testutil

import (
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v5"
)

// IntegrationClient wraps http.Client with state for workflow assertions.
type IntegrationClient struct {
	t            *testing.T
	client       *http.Client
	lastResponse *http.Response
	lastBody     string
	lastRedirect *url.URL
}

// SetupIntegrationServer starts a test server and returns a client that
// preserves cookies and records the latest response and redirect target.
func SetupIntegrationServer(t *testing.T, e *echo.Echo) (*httptest.Server, *IntegrationClient) {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("failed to create cookie jar: %v", err)
	}

	ic := &IntegrationClient{t: t}
	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	client := srv.Client()
	client.Jar = jar
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		ic.lastRedirect = req.URL
		return http.ErrUseLastResponse
	}

	ic.client = client
	return srv, ic
}

func (c *IntegrationClient) Get(target string) (*http.Response, string) {
	c.t.Helper()

	resp, err := c.client.Get(target)
	if err != nil {
		c.t.Fatalf("GET %q failed: %v", target, err)
	}
	return c.record(resp)
}

func (c *IntegrationClient) PostForm(target string, data url.Values) (*http.Response, string) {
	c.t.Helper()

	resp, err := c.client.PostForm(target, data)
	if err != nil && !errors.Is(err, http.ErrUseLastResponse) {
		c.t.Fatalf("POST %q failed: %v", target, err)
	}
	return c.record(resp)
}

func (c *IntegrationClient) FollowRedirect() (*http.Response, string) {
	c.t.Helper()

	if c.lastRedirect == nil {
		c.t.Fatal("expected redirect target, but no redirect was recorded")
	}
	return c.Get(c.lastRedirect.String())
}

func (c *IntegrationClient) AssertRedirect(t *testing.T, wantCode int, wantLocation string) {
	t.Helper()

	if c.lastResponse == nil {
		t.Fatal("expected a recorded response, but none was available")
	}

	if got := c.lastResponse.StatusCode; got != wantCode {
		t.Fatalf("expected status %d, got %d", wantCode, got)
	}

	if got := c.lastResponse.Header.Get("Location"); got != wantLocation {
		t.Fatalf("expected redirect location %q, got %q", wantLocation, got)
	}
}

func (c *IntegrationClient) AssertResponseCode(t *testing.T, wantCode int) {
	t.Helper()

	if c.lastResponse == nil {
		t.Fatal("expected a recorded response, but none was available")
	}

	if got := c.lastResponse.StatusCode; got != wantCode {
		t.Fatalf("expected status %d, got %d", wantCode, got)
	}
}

func (c *IntegrationClient) record(resp *http.Response) (*http.Response, string) {
	c.t.Helper()

	if resp == nil {
		c.t.Fatal("expected response, got nil")
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close() // nolint:errcheck
	if err != nil {
		c.t.Fatalf("failed to read response body: %v", err)
	}

	c.lastResponse = resp
	c.lastBody = string(body)
	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		c.lastRedirect = nil
	}

	return resp, c.lastBody
}
