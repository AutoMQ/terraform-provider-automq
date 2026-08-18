package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/models"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaAclImportStateInitializesLogicalIdentity(t *testing.T) {
	aclResource := &KafkaAclResource{}
	var schemaResp resource.SchemaResponse
	aclResource.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "unexpected schema diagnostics: %v", schemaResp.Diagnostics)

	importResp := resource.ImportStateResponse{State: tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), nil),
	}}
	aclResource.ImportState(context.Background(), resource.ImportStateRequest{
		ID: "env-test@kf-test@orders-writer|TOPIC|ALLOW|orders|LITERAL|PRODUCE",
	}, &importResp)
	require.False(t, importResp.Diagnostics.HasError(), "unexpected import diagnostics: %v", importResp.Diagnostics)

	var state models.KafkaAclResourceModel
	stateDiags := importResp.State.Get(context.Background(), &state)
	require.False(t, stateDiags.HasError(), "failed to decode prepared import state: %v", stateDiags)
	assert.Equal(t, types.StringValue("env-test"), state.EnvironmentID)
	assert.Equal(t, types.StringValue("kf-test"), state.KafkaInstance)
	assert.Equal(t, types.StringValue("orders-writer|TOPIC|ALLOW|orders"), state.ID)
	assert.Equal(t, types.StringValue("User:orders-writer"), state.Principal)
	assert.Equal(t, types.StringValue("TOPIC"), state.ResourceType)
	assert.Equal(t, types.StringValue("orders"), state.ResourceName)
	assert.Equal(t, types.StringValue("LITERAL"), state.PatternType)
	assert.Equal(t, types.StringValue("PRODUCE"), state.OperationGroup)
	assert.Equal(t, types.StringValue("ALLOW"), state.Permission)
}

func TestKafkaAclImportStateRejectsInvalidIDs(t *testing.T) {
	aclResource := &KafkaAclResource{}
	var schemaResp resource.SchemaResponse
	aclResource.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "unexpected schema diagnostics: %v", schemaResp.Diagnostics)

	tests := map[string]string{
		"missing resource context": "orders-writer|TOPIC|ALLOW|orders|LITERAL|PRODUCE",
		"missing identity field":   "env-test@kf-test@orders-writer|TOPIC|ALLOW|orders|LITERAL",
		"invalid identity escape":  "env-test@kf-test@orders%ZZwriter|TOPIC|ALLOW|orders|LITERAL|PRODUCE",
		"unsupported operation":    "env-test@kf-test@orders-writer|GROUP|ALLOW|orders|LITERAL|CONSUME",
	}

	for name, importID := range tests {
		t.Run(name, func(t *testing.T) {
			importResp := resource.ImportStateResponse{State: tfsdk.State{
				Schema: schemaResp.Schema,
				Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), nil),
			}}
			aclResource.ImportState(context.Background(), resource.ImportStateRequest{ID: importID}, &importResp)
			require.True(t, importResp.Diagnostics.HasError(), "expected import ID %q to be rejected", importID)
		})
	}
}

func TestKafkaAclReadUsesCompleteLogicalIdentity(t *testing.T) {
	wildcardHost := "*"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := client.PageNumResultKafkaAclBindingVO{
			List: []client.KafkaAclBindingVO{
				newProviderTestKafkaAcl("orders-writer", wildcardHost, "TOPIC", "orders", "LITERAL", "CONSUME", "ALLOW"),
				newProviderTestKafkaAcl("orders-writer", wildcardHost, "TOPIC", "orders", "LITERAL", "PRODUCE", "ALLOW"),
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer server.Close()

	api, err := client.NewClient(context.Background(), server.URL, client.AuthCredentials{AccessKeyID: "key", SecretAccessKey: "secret"})
	require.NoError(t, err)
	api.MaxRetries = 0
	aclResource := &KafkaAclResource{client: api}
	var schemaResp resource.SchemaResponse
	aclResource.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	state := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), nil),
	}
	stateDiags := state.Set(context.Background(), &models.KafkaAclResourceModel{
		EnvironmentID:  types.StringValue("env-test"),
		KafkaInstance:  types.StringValue("kf-test"),
		ID:             types.StringValue("orders-writer|TOPIC|ALLOW|orders"),
		ResourceType:   types.StringValue("TOPIC"),
		ResourceName:   types.StringValue("orders"),
		PatternType:    types.StringValue("LITERAL"),
		Principal:      types.StringValue("User:orders-writer"),
		OperationGroup: types.StringValue("PRODUCE"),
		Permission:     types.StringValue("ALLOW"),
	})
	require.False(t, stateDiags.HasError())

	readResp := resource.ReadResponse{State: tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), nil),
	}}
	aclResource.Read(context.Background(), resource.ReadRequest{State: state}, &readResp)
	require.False(t, readResp.Diagnostics.HasError(), "unexpected read diagnostics: %v", readResp.Diagnostics)

	var actual models.KafkaAclResourceModel
	readDiags := readResp.State.Get(context.Background(), &actual)
	require.False(t, readDiags.HasError())
	assert.Equal(t, types.StringValue("PRODUCE"), actual.OperationGroup)
}

func newProviderTestKafkaAcl(user, host, resourceType, resourceName, patternType, operationGroup, permission string) client.KafkaAclBindingVO {
	return client.KafkaAclBindingVO{
		AccessControl: &client.KafkaAccessControlVO{
			User:           user,
			Host:           &host,
			OperationGroup: client.OperationGroup{Name: operationGroup},
			PermissionType: permission,
		},
		ResourcePattern: &client.KafkaResourcePatternVO{
			ResourceType: resourceType,
			Name:         resourceName,
			PatternType:  patternType,
		},
	}
}
