package client

import "testing"

func TestConnectorPluginPathsMatchCmpRoutes(t *testing.T) {
	if pluginCollectionPath != "/connect/plugins" {
		t.Fatalf("pluginCollectionPath = %q", pluginCollectionPath)
	}
	if pluginItemPath != "/connect/plugins/%s" {
		t.Fatalf("pluginItemPath = %q", pluginItemPath)
	}
}
