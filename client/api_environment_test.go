package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestEnvironmentAPIUsesGlobalPaths(t *testing.T) {
	type capturedRequest struct {
		method string
		path   string
		body   string
	}
	var requests []capturedRequest
	client, err := NewClient(context.Background(), "https://example.com", AuthCredentials{
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-key",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.HTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var body []byte
		if req.Body != nil {
			body, err = io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("ReadAll(request body) error = %v", err)
			}
		}
		requests = append(requests, capturedRequest{method: req.Method, path: req.URL.Path, body: string(body)})
		responseBody := `{}`
		if req.Method == http.MethodPost {
			responseBody = `{"envId":"env-123","name":"production"}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}, nil
	})}

	ctx := context.Background()
	created, err := client.CreateEnvironment(ctx, EnvironmentCreateParam{
		Name:          "production",
		CloudProvider: "aws",
		Region:        "us-east-1",
		Scope:         "123456789012",
		Product:       "kafka",
	})
	if err != nil {
		t.Fatalf("CreateEnvironment() error = %v", err)
	}
	if created.EnvID != "env-123" {
		t.Fatalf("created environment ID = %q", created.EnvID)
	}
	if _, err := client.GetEnvironment(ctx, "env-123"); err != nil {
		t.Fatalf("GetEnvironment() error = %v", err)
	}
	if err := client.UpdateEnvironment(ctx, "env-123", EnvironmentUpdateParam{Name: "renamed"}); err != nil {
		t.Fatalf("UpdateEnvironment() error = %v", err)
	}
	if err := client.DeleteEnvironment(ctx, "env-123"); err != nil {
		t.Fatalf("DeleteEnvironment() error = %v", err)
	}

	want := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/environments"},
		{http.MethodGet, "/api/v1/environments/env-123"},
		{http.MethodPatch, "/api/v1/environments/env-123"},
		{http.MethodDelete, "/api/v1/environments/env-123"},
	}
	if len(requests) != len(want) {
		t.Fatalf("captured %d requests, want %d", len(requests), len(want))
	}
	for i := range want {
		if requests[i].method != want[i].method || requests[i].path != want[i].path {
			t.Errorf("request[%d] = %s %s, want %s %s", i, requests[i].method, requests[i].path, want[i].method, want[i].path)
		}
	}

	var createBody EnvironmentCreateParam
	if err := json.Unmarshal([]byte(requests[0].body), &createBody); err != nil {
		t.Fatalf("Unmarshal(create body) error = %v", err)
	}
	if createBody.Scope != "123456789012" || createBody.Product != "kafka" {
		t.Fatalf("create body = %#v", createBody)
	}
	if !strings.Contains(requests[2].body, `"description":null`) {
		t.Fatalf("update body = %s, want explicit null description", requests[2].body)
	}
}
