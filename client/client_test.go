package client

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func requestCaptureClient(t *testing.T, capture func(*http.Request)) *Client {
	t.Helper()
	client, err := NewClient(context.Background(), "https://example.com", AuthCredentials{
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-key",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.HTTPClient = &http.Client{
		Timeout: time.Second,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			capture(req)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		}),
	}
	return client
}

func TestBuildQueryParamsEscapesSpecialCharacters(t *testing.T) {
	query := buildQueryParams(map[string]string{
		"exactUser":         "automq-connect",
		"fuzzyResourceName": "*",
		"path":              "foo/bar baz",
	})

	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("ParseQuery(%q) failed: %v", query, err)
	}

	if got := values.Get("fuzzyResourceName"); got != "*" {
		t.Fatalf("fuzzyResourceName = %q, want *", got)
	}
	if got := values.Get("path"); got != "foo/bar baz" {
		t.Fatalf("path = %q, want foo/bar baz", got)
	}
	if strings.Contains(query, "fuzzyResourceName=*") {
		t.Fatalf("wildcard query value was not URL encoded: %q", query)
	}
	if !strings.Contains(query, "fuzzyResourceName=%2A") {
		t.Fatalf("wildcard query value was not encoded as %%2A: %q", query)
	}
}

func TestEnvironmentScopedRequestPath(t *testing.T) {
	var requestPath string
	var legacyEnvironmentHeader string
	client := requestCaptureClient(t, func(r *http.Request) {
		requestPath = r.URL.RequestURI()
		legacyEnvironmentHeader = r.Header.Get("X-automq-environment-id")
	})
	ctx := context.WithValue(context.Background(), EnvIdKey, "env-123")
	if _, err := client.Get(ctx, "/instances", map[string]string{"name": "test instance"}); err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if requestPath != "/api/v1/environments/env-123/instances?name=test%20instance" {
		t.Fatalf("request path = %q", requestPath)
	}
	if legacyEnvironmentHeader != "" {
		t.Fatalf("legacy environment header = %q, want empty", legacyEnvironmentHeader)
	}
}

func TestGlobalRequestPathDoesNotRequireEnvironment(t *testing.T) {
	var requestPath string
	client := requestCaptureClient(t, func(r *http.Request) {
		requestPath = r.URL.RequestURI()
	})
	if _, err := client.Get(context.Background(), "/environments/env-123", nil); err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if requestPath != "/api/v1/environments/env-123" {
		t.Fatalf("request path = %q", requestPath)
	}
}

func TestEnvironmentOwnedRequestStillRequiresEnvironment(t *testing.T) {
	client, err := NewClient(context.Background(), "https://example.com", AuthCredentials{
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-key",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	_, err = client.Get(context.Background(), "/instances", nil)
	if err == nil || !strings.Contains(err.Error(), "Error getting environment ID from context") {
		t.Fatalf("Get() error = %v", err)
	}
}
