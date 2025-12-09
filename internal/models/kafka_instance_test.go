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
				Name:        types.StringValue("test-instance"),
				Description: types.StringValue("test-description"),
				Version:     types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(4),
					DeployType:  types.StringValue("IAAS"),
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
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("wal-mode-1"),
					Security: &SecurityModel{
						DataEncryptionMode:           types.StringValue("encryption-mode-1"),
						CertificateAuthority:         types.StringValue("ca-1"),
						CertificateChain:             types.StringValue("chain-1"),
						PrivateKey:                   types.StringValue("key-1"),
						TlsHostnameValidationEnabled: types.BoolValue(true),
					},
					InstanceConfigs: types.MapValueMust(types.StringType, map[string]attr.Value{
						"config-key": types.StringValue("config-value"),
					}),
					MetricsExporter: &MetricsExporterModel{
						Prometheus: &PrometheusExporterModel{
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
				},
			},
			expected: client.InstanceCreateParam{
				Name:        "test-instance",
				Description: "test-description",
				Version:     "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku: 4,
					NodeConfig:  &client.NodeConfigParam{},
					DeployType:  stringPtr("IAAS"),
					Networks: []client.InstanceNetworkParam{
						{
							Zone:   "zone-1",
							Subnet: stringPtr("subnet-1"),
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
						DataEncryptionMode:           stringPtr("encryption-mode-1"),
						CertificateAuthority:         stringPtr("ca-1"),
						CertificateChain:             stringPtr("chain-1"),
						PrivateKey:                   stringPtr("key-1"),
						TlsHostnameValidationEnabled: boolPtr(true),
					},
					InstanceConfigs: []client.ConfigItemParam{
						{
							Key:   testStringPtr("config-key"),
							Value: testStringPtr("config-value"),
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
				},
			},
		},
		{
			name: "Minimal configuration",
			input: KafkaInstanceResourceModel{
				Name:    types.StringValue("minimal-instance"),
				Version: types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(1),
				},
			},
			expected: client.InstanceCreateParam{
				Name:    "minimal-instance",
				Version: "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku: 1,
					NodeConfig:  &client.NodeConfigParam{},
				},
			},
		},
		{
			name: "Empty arrays configuration",
			input: KafkaInstanceResourceModel{
				Name:    types.StringValue("empty-arrays"),
				Version: types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku:          types.Int64Value(2),
					Networks:             []NetworkModel{},
					KubernetesNodeGroups: []NodeGroupModel{},
				},
				Features: &FeaturesModel{
					WalMode:         types.StringValue(""),
					InstanceConfigs: types.MapValueMust(types.StringType, map[string]attr.Value{}),
				},
			},
			expected: client.InstanceCreateParam{
				Name:    "empty-arrays",
				Version: "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku:          2,
					NodeConfig:           &client.NodeConfigParam{},
					Networks:             nil,
					KubernetesNodeGroups: nil,
				},
				Features: &client.InstanceFeatureParam{
					WalMode:         stringPtr(""),
					InstanceConfigs: []client.ConfigItemParam{},
				},
			},
		},
		{
			name: "Nil features configuration",
			input: KafkaInstanceResourceModel{
				Name:    types.StringValue("nil-features"),
				Version: types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(2),
				},
				Features: nil,
			},
			expected: client.InstanceCreateParam{
				Name:    "nil-features",
				Version: "1.0.0",
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
				InstanceId:  strPtr("test-id"),
				Name:        strPtr("test-name"),
				Description: strPtr("test-description"),
				Version:     strPtr("1.0.0"),
				State:       strPtr("Running"),
				GmtCreate:   timePtr("2024-01-01T00:00:00Z"),
				GmtModified: timePtr("2024-01-02T00:00:00Z"),
			},
			expected: &KafkaInstanceResourceModel{
				InstanceID:     types.StringValue("test-id"),
				Name:           types.StringValue("test-name"),
				Description:    types.StringValue("test-description"),
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
			assert.Equal(t, tt.expected.Version, actual.Version)
			assert.Equal(t, tt.expected.InstanceStatus, actual.InstanceStatus)
			assert.Equal(t, tt.expected.CreatedAt, actual.CreatedAt)
			assert.Equal(t, tt.expected.LastUpdated, actual.LastUpdated)
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

func TestFlattenKafkaInstanceModel_RemovesMetricsExporterWhenAPIEmitsNone(t *testing.T) {
	tests := []struct {
		name     string
		exporter *client.InstanceMetricsExporterVO
	}{
		{name: "nil metrics exporter", exporter: nil},
		{name: "empty metrics exporter", exporter: &client.InstanceMetricsExporterVO{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &KafkaInstanceResourceModel{
				Features: &FeaturesModel{
					MetricsExporter: &MetricsExporterModel{
						Prometheus: &PrometheusExporterModel{
							AuthType: types.StringValue("noauth"),
							EndPoint: types.StringValue("https://previous.example.com"),
						},
					},
				},
			}
			instance := &client.InstanceVO{
				InstanceId: strPtr("test-instance"),
				Features: &client.InstanceFeatureVO{
					WalMode:         strPtr("EBSWAL"),
					MetricsExporter: tt.exporter,
				},
			}

			diags := FlattenKafkaInstanceModel(instance, resource)
			assert.False(t, diags.HasError())
			if resource.Features == nil {
				t.Fatalf("features unexpectedly nil")
			}
			assert.Nil(t, resource.Features.MetricsExporter)
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
