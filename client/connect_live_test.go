package client

import (
	"context"
	"os"
	"testing"
)

func TestLiveConnectAPIs(t *testing.T) {
	if os.Getenv("AUTOMQ_LIVE_TEST") != "1" {
		t.Skip("set AUTOMQ_LIVE_TEST=1 to run against a live CMP endpoint")
	}
	endpoint := os.Getenv("AUTOMQ_BYOC_ENDPOINT")
	accessKey := os.Getenv("AUTOMQ_BYOC_ACCESS_KEY")
	secretKey := os.Getenv("AUTOMQ_BYOC_SECRET_KEY")
	envID := os.Getenv("AUTOMQ_ENVIRONMENT_ID")
	if endpoint == "" || accessKey == "" || secretKey == "" || envID == "" {
		t.Fatal("AUTOMQ_BYOC_ENDPOINT, AUTOMQ_BYOC_ACCESS_KEY, AUTOMQ_BYOC_SECRET_KEY, and AUTOMQ_ENVIRONMENT_ID are required")
	}
	ctx := context.WithValue(context.Background(), EnvIdKey, envID)
	c, err := NewClient(ctx, endpoint, AuthCredentials{AccessKeyID: accessKey, SecretAccessKey: secretKey})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	body, err := c.Get(ctx, connectClusterCollectionPath, map[string]string{"page": "1", "size": "1"})
	if err != nil {
		t.Fatalf("list connect clusters: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("list connect clusters returned empty body")
	}

	body, err = c.Get(ctx, connectorCollectionPath, map[string]string{"page": "1", "size": "1"})
	if err != nil {
		t.Fatalf("list connectors: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("list connectors returned empty body")
	}
}
