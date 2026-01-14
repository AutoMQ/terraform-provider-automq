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
		{
			name: "FSWAL configuration",
			input: KafkaInstanceResourceModel{
				Name:    types.StringValue("fswal-instance"),
				Version: types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(4),
					FileSystemParam: &FileSystemParamModel{
						ThroughputMibpsPerFileSystem: types.Int64Value(1000),
						FileSystemCount:              types.Int64Value(2),
						SecurityGroups:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-12345")}),
					},
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("FSWAL"),
				},
			},
			expected: client.InstanceCreateParam{
				Name:    "fswal-instance",
				Version: "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku: 4,
					NodeConfig:  &client.NodeConfigParam{},
					FileSystem: &client.FileSystemParam{
						ThroughputMiBpsPerFileSystem: 1000,
						FileSystemCount:              2,
						SecurityGroups:               []string{"sg-12345"},
					},
				},
				Features: &client.InstanceFeatureParam{
					WalMode: stringPtr("FSWAL"),
				},
			},
		},
		{
			name: "FSWAL configuration without security group",
			input: KafkaInstanceResourceModel{
				Name:    types.StringValue("fswal-no-sg"),
				Version: types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(4),
					FileSystemParam: &FileSystemParamModel{
						ThroughputMibpsPerFileSystem: types.Int64Value(500),
						FileSystemCount:              types.Int64Value(1),
						SecurityGroups:               types.ListNull(types.StringType),
					},
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("FSWAL"),
				},
			},
			expected: client.InstanceCreateParam{
				Name:    "fswal-no-sg",
				Version: "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku: 4,
					NodeConfig:  &client.NodeConfigParam{},
					FileSystem: &client.FileSystemParam{
						ThroughputMiBpsPerFileSystem: 500,
						FileSystemCount:              1,
						SecurityGroups:               nil, // Should not be included when null/empty
					},
				},
				Features: &client.InstanceFeatureParam{
					WalMode: stringPtr("FSWAL"),
				},
			},
		},
		{
			name: "Configuration with compute_specs security_groups",
			input: KafkaInstanceResourceModel{
				Name:    types.StringValue("instance-with-sg"),
				Version: types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku:    types.Int64Value(4),
					SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-abc123"), types.StringValue("sg-def456")}),
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("EBSWAL"),
				},
			},
			expected: client.InstanceCreateParam{
				Name:    "instance-with-sg",
				Version: "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku:    4,
					NodeConfig:     &client.NodeConfigParam{},
					SecurityGroups: []string{"sg-abc123", "sg-def456"},
				},
				Features: &client.InstanceFeatureParam{
					WalMode: stringPtr("EBSWAL"),
				},
			},
		},
		{
			name: "Configuration with null compute_specs security_groups",
			input: KafkaInstanceResourceModel{
				Name:    types.StringValue("instance-no-sg"),
				Version: types.StringValue("1.0.0"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku:    types.Int64Value(4),
					SecurityGroups: types.ListNull(types.StringType),
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("EBSWAL"),
				},
			},
			expected: client.InstanceCreateParam{
				Name:    "instance-no-sg",
				Version: "1.0.0",
				Spec: client.SpecificationParam{
					ReservedAku:    4,
					NodeConfig:     &client.NodeConfigParam{},
					SecurityGroups: nil, // Should not be included when null
				},
				Features: &client.InstanceFeatureParam{
					WalMode: stringPtr("EBSWAL"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &client.InstanceCreateParam{}
			err := ExpandKafkaInstanceResource(context.Background(), tt.input, request)
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

			diags := FlattenKafkaInstanceModel(context.Background(), instance, resource)
			assert.False(t, diags.HasError())
			if resource.Features == nil {
				t.Fatalf("features unexpectedly nil")
			}
			assert.Nil(t, resource.Features.MetricsExporter)
		})
	}
}

func TestFlattenKafkaInstanceModel_FSWAL(t *testing.T) {
	tests := []struct {
		name     string
		input    *client.InstanceVO
		expected *KafkaInstanceResourceModel
	}{
		{
			name: "FSWAL with all parameters",
			input: &client.InstanceVO{
				InstanceId:  strPtr("fswal-instance"),
				Name:        strPtr("test-fswal"),
				Description: strPtr("FSWAL test instance"),
				Version:     strPtr("1.0.0"),
				State:       strPtr("Running"),
				Spec: &client.SpecificationVO{
					ReservedAku: int32Ptr(4),
					FileSystem: &client.FileSystemVO{
						ThroughputMiBpsPerFileSystem: int32Ptr(1000),
						FileSystemCount:              int32Ptr(2),
						SecurityGroups:               []string{"sg-12345"},
					},
				},
				Features: &client.InstanceFeatureVO{
					WalMode: strPtr("FSWAL"),
				},
			},
			expected: &KafkaInstanceResourceModel{
				InstanceID:     types.StringValue("fswal-instance"),
				Name:           types.StringValue("test-fswal"),
				Description:    types.StringValue("FSWAL test instance"),
				Version:        types.StringValue("1.0.0"),
				InstanceStatus: types.StringValue("Running"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(4),
					FileSystemParam: &FileSystemParamModel{
						ThroughputMibpsPerFileSystem: types.Int64Value(1000),
						FileSystemCount:              types.Int64Value(2),
						SecurityGroups:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-12345")}),
					},
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("FSWAL"),
				},
			},
		},
		{
			name: "FSWAL without security group",
			input: &client.InstanceVO{
				InstanceId:  strPtr("fswal-no-sg"),
				Name:        strPtr("test-fswal-no-sg"),
				Description: strPtr("FSWAL without security group"),
				Version:     strPtr("1.0.0"),
				State:       strPtr("Running"),
				Spec: &client.SpecificationVO{
					ReservedAku: int32Ptr(2),
					FileSystem: &client.FileSystemVO{
						ThroughputMiBpsPerFileSystem: int32Ptr(500),
						FileSystemCount:              int32Ptr(1),
						SecurityGroups:               nil,
					},
				},
				Features: &client.InstanceFeatureVO{
					WalMode: strPtr("FSWAL"),
				},
			},
			expected: &KafkaInstanceResourceModel{
				InstanceID:     types.StringValue("fswal-no-sg"),
				Name:           types.StringValue("test-fswal-no-sg"),
				Description:    types.StringValue("FSWAL without security group"),
				Version:        types.StringValue("1.0.0"),
				InstanceStatus: types.StringValue("Running"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku: types.Int64Value(2),
					FileSystemParam: &FileSystemParamModel{
						ThroughputMibpsPerFileSystem: types.Int64Value(500),
						FileSystemCount:              types.Int64Value(1),
						SecurityGroups:               types.ListNull(types.StringType),
					},
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("FSWAL"),
				},
			},
		},
		{
			name: "Non-FSWAL instance without file system",
			input: &client.InstanceVO{
				InstanceId:  strPtr("ebswal-instance"),
				Name:        strPtr("test-ebswal"),
				Description: strPtr("EBSWAL test instance"),
				Version:     strPtr("1.0.0"),
				State:       strPtr("Running"),
				Spec: &client.SpecificationVO{
					ReservedAku: int32Ptr(4),
					FileSystem:  nil, // No file system for non-FSWAL
				},
				Features: &client.InstanceFeatureVO{
					WalMode: strPtr("EBSWAL"),
				},
			},
			expected: &KafkaInstanceResourceModel{
				InstanceID:     types.StringValue("ebswal-instance"),
				Name:           types.StringValue("test-ebswal"),
				Description:    types.StringValue("EBSWAL test instance"),
				Version:        types.StringValue("1.0.0"),
				InstanceStatus: types.StringValue("Running"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku:     types.Int64Value(4),
					FileSystemParam: nil, // Should be nil for non-FSWAL
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("EBSWAL"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &KafkaInstanceResourceModel{}
			diags := FlattenKafkaInstanceModel(context.Background(), tt.input, resource)

			assert.False(t, diags.HasError(), "FlattenKafkaInstanceModel should not return errors")

			// Check basic fields
			assert.Equal(t, tt.expected.InstanceID, resource.InstanceID)
			assert.Equal(t, tt.expected.Name, resource.Name)
			assert.Equal(t, tt.expected.Description, resource.Description)
			assert.Equal(t, tt.expected.Version, resource.Version)
			assert.Equal(t, tt.expected.InstanceStatus, resource.InstanceStatus)

			// Check compute specs
			if tt.expected.ComputeSpecs == nil {
				assert.Nil(t, resource.ComputeSpecs)
			} else {
				assert.NotNil(t, resource.ComputeSpecs)
				assert.Equal(t, tt.expected.ComputeSpecs.ReservedAku, resource.ComputeSpecs.ReservedAku)

				// Check file system parameters
				if tt.expected.ComputeSpecs.FileSystemParam == nil {
					assert.Nil(t, resource.ComputeSpecs.FileSystemParam)
				} else {
					assert.NotNil(t, resource.ComputeSpecs.FileSystemParam)
					assert.Equal(t, tt.expected.ComputeSpecs.FileSystemParam.ThroughputMibpsPerFileSystem,
						resource.ComputeSpecs.FileSystemParam.ThroughputMibpsPerFileSystem)
					assert.Equal(t, tt.expected.ComputeSpecs.FileSystemParam.FileSystemCount,
						resource.ComputeSpecs.FileSystemParam.FileSystemCount)
					assert.Equal(t, tt.expected.ComputeSpecs.FileSystemParam.SecurityGroups,
						resource.ComputeSpecs.FileSystemParam.SecurityGroups)
				}
			}

			// Check features
			if tt.expected.Features == nil {
				assert.Nil(t, resource.Features)
			} else {
				assert.NotNil(t, resource.Features)
				assert.Equal(t, tt.expected.Features.WalMode, resource.Features.WalMode)
			}
		})
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func timePtr(s string) *time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return &t
}

func TestFlattenKafkaInstanceModel_ComputeSpecsSecurityGroups(t *testing.T) {
	tests := []struct {
		name     string
		input    *client.InstanceVO
		expected *KafkaInstanceResourceModel
	}{
		{
			name: "Instance with compute_specs security_groups",
			input: &client.InstanceVO{
				InstanceId:  strPtr("instance-with-sg"),
				Name:        strPtr("test-with-sg"),
				Description: strPtr("Instance with security groups"),
				Version:     strPtr("1.0.0"),
				State:       strPtr("Running"),
				Spec: &client.SpecificationVO{
					ReservedAku:    int32Ptr(4),
					SecurityGroups: []string{"sg-abc123", "sg-def456"},
				},
				Features: &client.InstanceFeatureVO{
					WalMode: strPtr("EBSWAL"),
				},
			},
			expected: &KafkaInstanceResourceModel{
				InstanceID:     types.StringValue("instance-with-sg"),
				Name:           types.StringValue("test-with-sg"),
				Description:    types.StringValue("Instance with security groups"),
				Version:        types.StringValue("1.0.0"),
				InstanceStatus: types.StringValue("Running"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku:    types.Int64Value(4),
					SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-abc123"), types.StringValue("sg-def456")}),
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("EBSWAL"),
				},
			},
		},
		{
			name: "Instance without compute_specs security_groups",
			input: &client.InstanceVO{
				InstanceId:  strPtr("instance-no-sg"),
				Name:        strPtr("test-no-sg"),
				Description: strPtr("Instance without security groups"),
				Version:     strPtr("1.0.0"),
				State:       strPtr("Running"),
				Spec: &client.SpecificationVO{
					ReservedAku:    int32Ptr(4),
					SecurityGroups: nil,
				},
				Features: &client.InstanceFeatureVO{
					WalMode: strPtr("EBSWAL"),
				},
			},
			expected: &KafkaInstanceResourceModel{
				InstanceID:     types.StringValue("instance-no-sg"),
				Name:           types.StringValue("test-no-sg"),
				Description:    types.StringValue("Instance without security groups"),
				Version:        types.StringValue("1.0.0"),
				InstanceStatus: types.StringValue("Running"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku:    types.Int64Value(4),
					SecurityGroups: types.ListNull(types.StringType),
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("EBSWAL"),
				},
			},
		},
		{
			name: "Instance with single security group",
			input: &client.InstanceVO{
				InstanceId:  strPtr("instance-single-sg"),
				Name:        strPtr("test-single-sg"),
				Description: strPtr("Instance with single security group"),
				Version:     strPtr("1.0.0"),
				State:       strPtr("Running"),
				Spec: &client.SpecificationVO{
					ReservedAku:    int32Ptr(6),
					SecurityGroups: []string{"sg-only-one"},
				},
				Features: &client.InstanceFeatureVO{
					WalMode: strPtr("S3WAL"),
				},
			},
			expected: &KafkaInstanceResourceModel{
				InstanceID:     types.StringValue("instance-single-sg"),
				Name:           types.StringValue("test-single-sg"),
				Description:    types.StringValue("Instance with single security group"),
				Version:        types.StringValue("1.0.0"),
				InstanceStatus: types.StringValue("Running"),
				ComputeSpecs: &ComputeSpecsModel{
					ReservedAku:    types.Int64Value(6),
					SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-only-one")}),
				},
				Features: &FeaturesModel{
					WalMode: types.StringValue("S3WAL"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &KafkaInstanceResourceModel{}
			diags := FlattenKafkaInstanceModel(context.Background(), tt.input, resource)

			assert.False(t, diags.HasError(), "FlattenKafkaInstanceModel should not return errors")

			// Check basic fields
			assert.Equal(t, tt.expected.InstanceID, resource.InstanceID)
			assert.Equal(t, tt.expected.Name, resource.Name)
			assert.Equal(t, tt.expected.Description, resource.Description)
			assert.Equal(t, tt.expected.Version, resource.Version)
			assert.Equal(t, tt.expected.InstanceStatus, resource.InstanceStatus)

			// Check compute specs
			assert.NotNil(t, resource.ComputeSpecs)
			assert.Equal(t, tt.expected.ComputeSpecs.ReservedAku, resource.ComputeSpecs.ReservedAku)

			// Check security groups
			assert.Equal(t, tt.expected.ComputeSpecs.SecurityGroups, resource.ComputeSpecs.SecurityGroups)

			// Check features
			assert.NotNil(t, resource.Features)
			assert.Equal(t, tt.expected.Features.WalMode, resource.Features.WalMode)
		})
	}
}

// TestFlattenKafkaInstanceModel_SecurityGroupsUnknownToNull tests that
// when previous state has unknown security_groups and API returns null,
// the result should be ListNull (not unknown).
func TestFlattenKafkaInstanceModel_SecurityGroupsUnknownToNull(t *testing.T) {
	// Simulate a resource with unknown security_groups (like during plan/apply)
	resource := &KafkaInstanceResourceModel{
		ComputeSpecs: &ComputeSpecsModel{
			ReservedAku:    types.Int64Value(4),
			SecurityGroups: types.ListUnknown(types.StringType), // Unknown during planning
		},
	}

	// API returns null security_groups
	input := &client.InstanceVO{
		InstanceId:  strPtr("test-instance"),
		Name:        strPtr("test"),
		Version:     strPtr("1.0.0"),
		State:       strPtr("Running"),
		Spec: &client.SpecificationVO{
			ReservedAku:    int32Ptr(4),
			SecurityGroups: nil, // API returns null
		},
		Features: &client.InstanceFeatureVO{
			WalMode: strPtr("EBSWAL"),
		},
	}

	diags := FlattenKafkaInstanceModel(context.Background(), input, resource)

	assert.False(t, diags.HasError(), "FlattenKafkaInstanceModel should not return errors")
	assert.NotNil(t, resource.ComputeSpecs)

	// The key assertion: security_groups should be ListNull, not ListUnknown
	assert.True(t, resource.ComputeSpecs.SecurityGroups.IsNull(),
		"security_groups should be null when API returns null, got: %v", resource.ComputeSpecs.SecurityGroups)
	assert.False(t, resource.ComputeSpecs.SecurityGroups.IsUnknown(),
		"security_groups should not be unknown after flatten")
}
