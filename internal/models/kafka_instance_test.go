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
			name: "FSWAL with file system param",
			input: KafkaInstanceResourceModel{
				Name:          types.StringValue("fswal-instance"),
				DeployProfile: types.StringValue("fswal-profile"),
				Version:       types.StringValue("1.2.3"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(8),
					FileSystemParam: &FileSystemParamModel{
						ThroughputMiBpsPerFileSystem: types.Int64Value(256),
						FileSystemCount:              types.Int64Value(4),
					},
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("FSWAL"),
				},
			},
			expected: client.InstanceCreateParam{
				Name:          "fswal-instance",
				DeployProfile: "fswal-profile",
				Version:       "1.2.3",
				Spec: client.SpecificationParam{
					ReservedAku: 8,
					NodeConfig:  &client.NodeConfigParam{},
					FileSystem: &client.FileSystemParam{
						ThroughputMiBpsPerFileSystem: 256,
						FileSystemCount:              4,
					},
				},
				Features: &client.InstanceFeatureParam{
					WalMode: stringPtr("FSWAL"),
				},
			},
		},
		{
			name: "Full configuration",
			input: KafkaInstanceResourceModel{
				Name:          types.StringValue("test-instance"),
				Description:   types.StringValue("test-description"),
				DeployProfile: types.StringValue("test-profile"),
				Version:       types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(4),
					DeployType:  types.StringValue("IAAS"),
					Provider:    types.StringValue("aws"),
					Region:      types.StringValue("us-east-1"),
					Vpc:         types.StringValue("vpc-1"),
					DataBuckets: types.ListValueMust(
						DataBucketObjectType,
						[]attr.Value{
							types.ObjectValueMust(
								DataBucketObjectType.AttrTypes,
								map[string]attr.Value{
									"bucket_name": types.StringValue("data-bucket-1"),
								},
							),
						},
					),
					Networks: []NetworkModel{
						{
							Zone:    types.StringValue("zone-1"),
							Subnets: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-1")}),
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
					Integrations: types.SetValueMust(types.StringType, []attr.Value{
						types.StringValue("integration-1"),
					}),
					InstanceConfigs: types.MapValueMust(types.StringType, map[string]attr.Value{
						"config-key": types.StringValue("config-value"),
					}),
					MetricsExporter: &MetricsExporterModel{
						Prometheus: &PrometheusExporterModel{
							Enabled:  types.BoolValue(true),
							EndPoint: types.StringValue("http://prometheus"),
							Labels: types.MapValueMust(types.StringType, map[string]attr.Value{
								"env": types.StringValue("test"),
							}),
						},
					},
					TableTopic: &TableTopicModel{
						Warehouse:   types.StringValue("warehouse-1"),
						CatalogType: types.StringValue("HIVE"),
					},
					ExtendListeners: []InstanceListenerModel{
						{
							ListenerName:     types.StringValue("PUBLIC"),
							SecurityProtocol: types.StringValue("PLAINTEXT"),
							Port:             types.Int64Value(19092),
						},
					},
					InboundRules: []InboundRuleModel{
						{
							ListenerName: types.StringValue("PUBLIC"),
							Cidrs:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("0.0.0.0/0")}),
						},
					},
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
					DeployType:  stringPtr("IAAS"),
					Provider:    stringPtr("aws"),
					Region:      stringPtr("us-east-1"),
					Vpc:         stringPtr("vpc-1"),
					Networks: []client.InstanceNetworkParam{
						{
							Zone:   "zone-1",
							Subnet: stringPtr("subnet-1"),
						},
					},
					BucketProfiles: []client.BucketProfileBindParam{
						{
							Id: stringPtr("bucket-1"),
						},
					},
					DataBuckets: []client.BucketProfileParam{
						{
							BucketName: "data-bucket-1",
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
					MetricsExporter: &client.InstanceMetricsExporterParam{
						Prometheus: &client.InstancePrometheusExporterParam{
							Enabled:  boolPtr(true),
							EndPoint: stringPtr("http://prometheus"),
							Labels: []client.MetricsLabelParam{
								{Name: "env", Value: "test"},
							},
						},
					},
					TableTopic: &client.TableTopicParam{
						Warehouse:   "warehouse-1",
						CatalogType: "HIVE",
					},
					ExtendListeners: []client.InstanceListenerParam{
						{
							ListenerName:     "PUBLIC",
							SecurityProtocol: stringPtr("PLAINTEXT"),
							Port:             int32Ptr(19092),
						},
					},
					InboundRules: []client.InboundRuleParam{
						{
							ListenerName: "PUBLIC",
							Cidrs:        []string{"0.0.0.0/0"},
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
					Integrations:    types.SetValueMust(types.StringType, []attr.Value{}),
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

func TestExpandKafkaInstanceResource_FswalMissingFields(t *testing.T) {
	request := &client.InstanceCreateParam{}
	model := KafkaInstanceResourceModel{
		Name:          types.StringValue("bad-fswal"),
		DeployProfile: types.StringValue("profile"),
		Version:       types.StringValue("1.0.0"),
		ComputeSpecs: &ComputeSpecsModel{
			ReservedAku: types.Int64Value(6),
			FileSystemParam: &FileSystemParamModel{
				ThroughputMiBpsPerFileSystem: types.Int64Null(),
				FileSystemCount:              types.Int64Value(2),
			},
		},
		Features: &FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	err := ExpandKafkaInstanceResource(model, request)
	assert.Error(t, err)
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
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

func TestFlattenKafkaInstanceModel_FileSystemParam(t *testing.T) {
	instance := &client.InstanceVO{
		Spec: &client.SpecificationVO{
			ReservedAku: int32Ptr(8),
			FileSystem: &client.FileSystemVO{
				ThroughputMiBpsPerFileSystem: int32Ptr(512),
				FileSystemCount:              int32Ptr(6),
			},
		},
		Features: &client.InstanceFeatureVO{
			WalMode: stringPtr("FSWAL"),
		},
	}
	resource := &KafkaInstanceResourceModel{}
	diags := FlattenKafkaInstanceModel(instance, resource)
	assert.False(t, diags.HasError())
	if assert.NotNil(t, resource.ComputeSpecs) && assert.NotNil(t, resource.ComputeSpecs.FileSystemParam) {
		assert.Equal(t, int64(512), resource.ComputeSpecs.FileSystemParam.ThroughputMiBpsPerFileSystem.ValueInt64())
		assert.Equal(t, int64(6), resource.ComputeSpecs.FileSystemParam.FileSystemCount.ValueInt64())
	}
	assert.Equal(t, types.StringValue("FSWAL"), resource.Features.WalMode)
}

func TestFlattenKafkaInstanceModelWithIntegrations(t *testing.T) {
	tests := []struct {
		name         string
		integrations []client.IntegrationVO
		expected     types.Set
	}{
		{
			name: "normal case",
			integrations: []client.IntegrationVO{
				{Code: "integration1"},
				{Code: "integration2"},
			},
			expected: types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("integration1"),
					types.StringValue("integration2"),
				},
			),
		},
		{
			name:         "empty integrations",
			integrations: []client.IntegrationVO{},
			expected:     types.SetValueMust(types.StringType, []attr.Value{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &KafkaInstanceResourceModel{}
			diags := FlattenKafkaInstanceModelWithIntegrations(tt.integrations, resource)

			assert.False(t, diags.HasError())

			assert.Equal(t, len(tt.expected.Elements()), len(resource.Features.Integrations.Elements()))
			expectedElements := tt.expected.Elements()
			integrationElements := resource.Features.Integrations.Elements()
			for i, expected := range expectedElements {
				assert.Equal(t, expected, integrationElements[i])
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

func int32Ptr(v int32) *int32 {
	return &v
}
