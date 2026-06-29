package client

import "testing"

func TestConnectorPluginPathsMatchCmpRoutes(t *testing.T) {
	if pluginCollectionPath != "/api/v1/connect/plugins" {
		t.Fatalf("pluginCollectionPath = %q", pluginCollectionPath)
	}
	if pluginItemPath != "/api/v1/connect/plugins/%s" {
		t.Fatalf("pluginItemPath = %q", pluginItemPath)
	}
}
