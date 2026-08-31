package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestEnvironmentCredentialsSchema(t *testing.T) {
	var response resource.SchemaResponse
	NewEnvironmentResource().Schema(context.Background(), resource.SchemaRequest{}, &response)

	for _, name := range []string{"client_id", "client_secret"} {
		attribute, ok := response.Schema.Attributes[name]
		if !ok {
			t.Fatalf("schema does not contain %q", name)
		}
		if !attribute.IsComputed() {
			t.Errorf("%s must be computed", name)
		}
		if !attribute.IsSensitive() {
			t.Errorf("%s must be sensitive", name)
		}
		if attribute.IsOptional() || attribute.IsRequired() {
			t.Errorf("%s must be read-only", name)
		}
	}
}
