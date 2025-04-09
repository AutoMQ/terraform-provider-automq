package models

import (
	"terraform-provider-automq/client"
	"testing"
	"time"

	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestExpandKafkaInstanceResource(t *testing.T) {
	tests := []struct {
		name     string
		input    KafkaInstanceResourceModel
		expected client.InstanceCreateParam
	}{
		{
			name: "Full configuration",
			input: KafkaInstanceResourceModel{
				Name:          types.StringValue("test-instance"),
				Description:   types.StringValue("test-description"),
				DeployProfile: types.StringValue("test-profile"),
				Version:       types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(4),
					Networks: []NetworkModel{
						{
							Zone:    types.StringValue("zone-1"),
							Subnets: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-1")}),
						},
					},
					KubernetesNodeGroups: []NodeGroupModel{
						{
							ID: types.StringValue("node-group-1"),
						},
					},
					BucketProfiles: []BucketProfileIDModel{
						{
							ID: types.StringValue("bucket-1"),
						},
					},
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("wal-mode-1"),
					Security: &SecurityModel{
						DataEncryptionMode:   types.StringValue("encryption-mode-1"),
						CertificateAuthority: types.StringValue("ca-1"),
						CertificateChain:     types.StringValue("chain-1"),
						PrivateKey:           types.StringValue("key-1"),
					},
					Integrations: []IntegrationModel{
						{
							ID: types.StringValue("integration-1"),
						},
					},
					InstanceConfigs: types.MapValueMust(types.StringType, map[string]attr.Value{
						"config-key": types.StringValue("config-value"),
					}),
				},
			},
			expected: client.InstanceCreateParam{
				Name:          "test-instance",
				Description:   "test-description",
				DeployProfile: "test-profile",
				Version:       "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku: 4,
					NodeConfig:  &client.NodeConfigParam{},
					Networks: []client.InstanceNetworkParam{
						{
							Zone:   "zone-1",
							Subnet: stringPtr("subnet-1"),
						},
					},
					KubernetesNodeGroups: []client.KubernetesNodeGroupParam{
						{
							Id: stringPtr("node-group-1"),
						},
					},
					BucketProfiles: []client.BucketProfileBindParam{
						{
							Id: stringPtr("bucket-1"),
						},
					},
				},
				Features: &client.InstanceFeatureParam{
					WalMode: stringPtr("wal-mode-1"),
					Security: &client.InstanceSecurityParam{
						DataEncryptionMode:   stringPtr("encryption-mode-1"),
						CertificateAuthority: stringPtr("ca-1"),
						CertificateChain:     stringPtr("chain-1"),
						PrivateKey:           stringPtr("key-1"),
					},
					Integrations: []client.IntegrationBindParam{
						{
							Id: stringPtr("integration-1"),
						},
					},
					InstanceConfigs: []client.ConfigItemParam{
						{
							Key:   "config-key",
							Value: "config-value",
						},
					},
				},
			},
		},
		{
			name: "Minimal configuration",
			input: KafkaInstanceResourceModel{
				Name:          types.StringValue("minimal-instance"),
				DeployProfile: types.StringValue("minimal-profile"),
				Version:       types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(1),
				},
			},
			expected: client.InstanceCreateParam{
				Name:          "minimal-instance",
				DeployProfile: "minimal-profile",
				Version:       "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku: 1,
					NodeConfig:  &client.NodeConfigParam{},
				},
			},
		},
		{
			name: "Empty arrays configuration",
			input: KafkaInstanceResourceModel{
				Name:          types.StringValue("empty-arrays"),
				DeployProfile: types.StringValue("test-profile"),
				Version:       types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku:          types.Int64Value(2),
					Networks:             []NetworkModel{},
					KubernetesNodeGroups: []NodeGroupModel{},
					BucketProfiles:       []BucketProfileIDModel{},
				},
				Features: &FeaturesModel{
					WalMode:         types.StringValue(""),
					Integrations:    []IntegrationModel{},
					InstanceConfigs: types.MapValueMust(types.StringType, map[string]attr.Value{}),
				},
			},
			expected: client.InstanceCreateParam{
				Name:          "empty-arrays",
				DeployProfile: "test-profile",
				Version:       "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku:          2,
					NodeConfig:           &client.NodeConfigParam{},
					Networks:             nil,
					KubernetesNodeGroups: nil,
					BucketProfiles:       nil,
				},
				Features: &client.InstanceFeatureParam{
					WalMode:         stringPtr(""),
					Integrations:    nil,
					InstanceConfigs: []client.ConfigItemParam{},
				},
			},
		},
		{
			name: "Nil features configuration",
			input: KafkaInstanceResourceModel{
				Name:          types.StringValue("nil-features"),
				DeployProfile: types.StringValue("test-profile"),
				Version:       types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(2),
				},
				Features: nil,
			},
			expected: client.InstanceCreateParam{
				Name:          "nil-features",
				DeployProfile: "test-profile",
				Version:       "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku: 2,
					NodeConfig:  &client.NodeConfigParam{},
				},
				Features: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &client.InstanceCreateParam{}
			err := ExpandKafkaInstanceResource(tt.input, request)
			assert.Equal(t, tt.expected, *request)
			assert.NoError(t, err)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestFlattenKafkaInstanceBasicModel(t *testing.T) {
	tests := []struct {
		name     string
		input    *client.InstanceSummaryVO
		expected *KafkaInstanceResourceModel
	}{
		{
			name: "normal case",
			input: &client.InstanceSummaryVO{
				InstanceId:    strPtr("test-id"),
				Name:          strPtr("test-name"),
				Description:   strPtr("test-description"),
				DeployProfile: strPtr("test-profile"),
				Version:       strPtr("1.0.0"),
				State:         strPtr("Running"),
				GmtCreate:     timePtr("2024-01-01T00:00:00Z"),
				GmtModified:   timePtr("2024-01-02T00:00:00Z"),
			},
			expected: &KafkaInstanceResourceModel{
				InstanceID:     types.StringValue("test-id"),
				Name:           types.StringValue("test-name"),
				Description:    types.StringValue("test-description"),
				DeployProfile:  types.StringValue("test-profile"),
				Version:        types.StringValue("1.0.0"),
				InstanceStatus: types.StringValue("Running"),
				CreatedAt:      timetypes.NewRFC3339TimePointerValue(timePtr("2024-01-01T00:00:00Z")),
				LastUpdated:    timetypes.NewRFC3339TimePointerValue(timePtr("2024-01-02T00:00:00Z")),
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: &KafkaInstanceResourceModel{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := &KafkaInstanceResourceModel{}
			diags := FlattenKafkaInstanceBasicModel(tt.input, actual)

			if tt.input == nil {
				assert.True(t, diags.HasError())
				return
			}

			assert.False(t, diags.HasError())
			assert.Equal(t, tt.expected.InstanceID, actual.InstanceID)
			assert.Equal(t, tt.expected.Name, actual.Name)
			assert.Equal(t, tt.expected.Description, actual.Description)
			assert.Equal(t, tt.expected.DeployProfile, actual.DeployProfile)
			assert.Equal(t, tt.expected.Version, actual.Version)
			assert.Equal(t, tt.expected.InstanceStatus, actual.InstanceStatus)
			assert.Equal(t, tt.expected.CreatedAt, actual.CreatedAt)
			assert.Equal(t, tt.expected.LastUpdated, actual.LastUpdated)
		})
	}
}

func TestFlattenKafkaInstanceModelWithIntegrations(t *testing.T) {
	tests := []struct {
		name         string
		integrations []client.IntegrationVO
		expected     []IntegrationModel
	}{
		{
			name: "normal case",
			integrations: []client.IntegrationVO{
				{Code: "integration1"},
				{Code: "integration2"},
			},
			expected: []IntegrationModel{
				{ID: types.StringValue("integration1")},
				{ID: types.StringValue("integration2")},
			},
		},
		{
			name:         "empty integrations",
			integrations: []client.IntegrationVO{},
			expected:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &KafkaInstanceResourceModel{}
			diags := FlattenKafkaInstanceModelWithIntegrations(tt.integrations, resource)

			assert.False(t, diags.HasError())
			if tt.expected == nil {
				assert.Nil(t, resource.Features)
				return
			}

			assert.Equal(t, len(tt.expected), len(resource.Features.Integrations))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.ID, resource.Features.Integrations[i].ID)
			}
		})
	}
}

func TestFlattenKafkaInstanceModelWithEndpoints(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []client.InstanceAccessInfoVO
		expected  []attr.Value
	}{
		{
			name: "normal case",
			endpoints: []client.InstanceAccessInfoVO{
				{
					DisplayName:      strPtr("endpoint1"),
					NetworkType:      strPtr("private"),
					Protocol:         strPtr("SASL_PLAINTEXT"),
					Mechanisms:       strPtr("PLAIN"),
					BootstrapServers: strPtr("localhost:9092"),
				},
			},
			expected: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"display_name":      types.StringType,
						"network_type":      types.StringType,
						"protocol":          types.StringType,
						"mechanisms":        types.StringType,
						"bootstrap_servers": types.StringType,
					},
					map[string]attr.Value{
						"display_name":      types.StringValue("endpoint1"),
						"network_type":      types.StringValue("private"),
						"protocol":          types.StringValue("SASL_PLAINTEXT"),
						"mechanisms":        types.StringValue("PLAIN"),
						"bootstrap_servers": types.StringValue("localhost:9092"),
					},
				),
			},
		},
		{
			name:      "empty endpoints",
			endpoints: []client.InstanceAccessInfoVO{},
			expected:  []attr.Value{},
		},
		{
			name:      "nil endpoints",
			endpoints: nil,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &KafkaInstanceResourceModel{}
			diags := FlattenKafkaInstanceModelWithEndpoints(tt.endpoints, resource)

			if tt.endpoints == nil {
				assert.True(t, diags.HasError())
				return
			}

			assert.False(t, diags.HasError())

			assert.Equal(t, len(tt.expected), len(resource.Endpoints.Elements()))

			if len(tt.expected) == 0 {
				return
			}

			var endpoints []InstanceAccessInfo
			diags = resource.Endpoints.ElementsAs(context.Background(), &endpoints, false)
			assert.False(t, diags.HasError())

			for i, endpoint := range endpoints {
				expectedObj, ok := tt.expected[i].(types.Object)
				if !ok {
					t.Fatalf("expected[%d] is not of type types.Object", i)
				}
				assert.Equal(t, expectedObj.Attributes()["display_name"], endpoint.DisplayName)
				assert.Equal(t, expectedObj.Attributes()["network_type"], endpoint.NetworkType)
				assert.Equal(t, expectedObj.Attributes()["protocol"], endpoint.Protocol)
				assert.Equal(t, expectedObj.Attributes()["mechanisms"], endpoint.Mechanisms)
				assert.Equal(t, expectedObj.Attributes()["bootstrap_servers"], endpoint.BootstrapServers)
			}
		})
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func timePtr(s string) *time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return &t
}
