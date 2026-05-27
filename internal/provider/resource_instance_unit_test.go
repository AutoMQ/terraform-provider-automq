package provider

import (
	"context"
	"errors"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/models"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStringPtr(s string) *string {
	return &s
}

func testSecurityObject(t *testing.T, model *models.SecurityModel) types.Object {
	t.Helper()
	value, diags := models.SecurityModelToObject(context.Background(), model)
	require.False(t, diags.HasError(), "failed to build security object: %v", diags)
	return value
}

func testMetricsExporterObject(t *testing.T, model *models.MetricsExporterModel) types.Object {
	t.Helper()
	value, diags := models.MetricsExporterModelToObject(context.Background(), model)
	require.False(t, diags.HasError(), "failed to build metrics exporter object: %v", diags)
	return value
}

func testTableTopicObject(t *testing.T, model *models.TableTopicModel) types.Object {
	t.Helper()
	value, diags := models.TableTopicModelToObject(context.Background(), model)
	require.False(t, diags.HasError(), "failed to build table topic object: %v", diags)
	return value
}

func testFileSystemObject(t *testing.T, model *models.FileSystemParamModel) types.Object {
	t.Helper()
	value, diags := models.FileSystemParamModelToObject(context.Background(), model)
	require.False(t, diags.HasError(), "failed to build file system object: %v", diags)
	return value
}

func testNetworkList(t *testing.T, networks []models.NetworkModel) types.List {
	t.Helper()
	value, diags := models.NetworkModelsToList(context.Background(), networks)
	require.False(t, diags.HasError(), "failed to build network list: %v", diags)
	return value
}

func testNodeGroupList(t *testing.T, groups []models.NodeGroupModel) types.List {
	t.Helper()
	value, diags := models.NodeGroupModelsToList(context.Background(), groups)
	require.False(t, diags.HasError(), "failed to build node group list: %v", diags)
	return value
}

func testValidateInstanceUpdateContract(plan, state models.KafkaInstanceResourceModel) diag.Diagnostics {
	return validateInstanceUpdateContract(context.Background(), "inst-1", plan, state)
}

func testBuildInstanceUpdateParam(t *testing.T, plan, state models.KafkaInstanceResourceModel) (client.InstanceUpdateParam, instanceUpdatePlan) {
	t.Helper()
	param, updatePlan, diags := buildInstanceUpdateParam(context.Background(), plan, state)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	return param, updatePlan
}

func testApplyUpdateStatePreservation(t *testing.T, state *models.KafkaInstanceResourceModel, plan models.KafkaInstanceResourceModel, updatePlan instanceUpdatePlan) {
	t.Helper()
	diags := applyUpdateStatePreservation(context.Background(), state, plan, updatePlan)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
}

type stubKafkaInstanceAPI struct {
	instance         *client.InstanceVO
	endpoints        []client.InstanceAccessInfoVO
	getInstanceErr   error
	getEndpointsErr  error
	getEndpointsCall int
}

func (s *stubKafkaInstanceAPI) CreateKafkaInstance(context.Context, client.InstanceCreateParam) (*client.InstanceSummaryVO, error) {
	return nil, errors.New("unexpected CreateKafkaInstance call")
}

func (s *stubKafkaInstanceAPI) GetKafkaInstance(context.Context, string) (*client.InstanceVO, error) {
	return s.instance, s.getInstanceErr
}

func (s *stubKafkaInstanceAPI) GetKafkaInstanceByName(context.Context, string) (*client.InstanceVO, error) {
	return nil, errors.New("unexpected GetKafkaInstanceByName call")
}

func (s *stubKafkaInstanceAPI) DeleteKafkaInstance(context.Context, string) error {
	return errors.New("unexpected DeleteKafkaInstance call")
}

func (s *stubKafkaInstanceAPI) UpdateKafkaInstance(context.Context, string, client.InstanceUpdateParam) error {
	return errors.New("unexpected UpdateKafkaInstance call")
}

func (s *stubKafkaInstanceAPI) GetInstanceEndpoints(context.Context, string) ([]client.InstanceAccessInfoVO, error) {
	s.getEndpointsCall++
	return s.endpoints, s.getEndpointsErr
}

// Contract validation tests keep provider-side plan rules explicit and separate
// from the lower-level model expansion/flattening tests.
func TestInstanceContractValidation(t *testing.T) {
	t.Run("valid usage based iaas plan passes", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		diags := validateInstanceContract(context.Background(), &plan)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})

	t.Run("k8s deploy type requires cluster and node groups", func(t *testing.T) {
		plan := models.KafkaInstanceResourceModel{
			ComputeSpecs: &models.ComputeSpecsModel{
				DeployType:           types.StringValue("K8S"),
				KubernetesClusterID:  types.StringNull(),
				KubernetesNodeGroups: types.ListNull(models.NodeGroupObjectType),
			},
			Features: &models.FeaturesModel{
				WalMode: types.StringValue("EBSWAL"),
			},
		}
		diags := validateInstanceContract(context.Background(), &plan)
		require.True(t, diags.HasError())
		assert.Len(t, diags.Errors(), 2)
	})

	t.Run("usage based iaas requires reserved node count and instance types", func(t *testing.T) {
		plan := models.KafkaInstanceResourceModel{
			ComputeSpecs: &models.ComputeSpecsModel{
				PricingMode:       types.StringValue("UsageBased"),
				DeployType:        types.StringValue("IAAS"),
				ReservedNodeCount: types.Int64Null(),
				InstanceTypes:     types.ListNull(types.StringType),
			},
			Features: &models.FeaturesModel{
				WalMode: types.StringValue("EBSWAL"),
			},
		}
		diags := validateInstanceContract(context.Background(), &plan)
		require.True(t, diags.HasError())
		assert.Len(t, diags.Errors(), 2)
	})

	t.Run("fswal requires file system contract and rejects k8s", func(t *testing.T) {
		plan := models.KafkaInstanceResourceModel{
			ComputeSpecs: &models.ComputeSpecsModel{
				DeployType: types.StringValue("K8S"),
			},
			Features: &models.FeaturesModel{
				WalMode: types.StringValue("FSWAL"),
			},
		}
		diags := validateInstanceContract(context.Background(), &plan)
		require.True(t, diags.HasError())
		assert.GreaterOrEqual(t, len(diags.Errors()), 2)
	})

	t.Run("data buckets require bucket name", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		plan.ComputeSpecs.DataBuckets = types.ListValueMust(
			models.DataBucketObjectType,
			[]attr.Value{
				types.ObjectValueMust(
					models.DataBucketObjectType.AttrTypes,
					map[string]attr.Value{
						"bucket_name": types.StringNull(),
					},
				),
			},
		)
		diags := validateInstanceContract(context.Background(), &plan)
		require.True(t, diags.HasError())
		assert.Len(t, diags.Errors(), 1)
	})

	t.Run("metrics exporter block requires configured prometheus block", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		plan.Features.MetricsExporter = testMetricsExporterObject(t, &models.MetricsExporterModel{
			Prometheus: &models.PrometheusExporterModel{
				AuthType: types.StringNull(),
			},
		})
		diags := validateInstanceContract(context.Background(), &plan)
		require.True(t, diags.HasError())
		assert.Len(t, diags.Errors(), 1)
	})

	t.Run("table topic requires explicit schema registry enablement", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		plan.Features.SchemaRegistryEnabled = types.BoolNull()
		plan.Features.TableTopic = testTableTopicObject(t, &models.TableTopicModel{
			Warehouse:   types.StringValue("warehouse"),
			CatalogType: types.StringValue("glue"),
		})
		diags := validateInstanceContract(context.Background(), &plan)
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "schema_registry_enabled")
	})

	t.Run("table topic with schema registry enablement passes", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		plan.Features.SchemaRegistryEnabled = types.BoolValue(true)
		plan.Features.TableTopic = testTableTopicObject(t, &models.TableTopicModel{
			Warehouse:   types.StringValue("warehouse"),
			CatalogType: types.StringValue("glue"),
		})
		diags := validateInstanceContract(context.Background(), &plan)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})
}

func TestInstanceUpdateContractValidation(t *testing.T) {
	t.Run("removing instance config key is rejected", func(t *testing.T) {
		plan := newConfigOnlyPlan(map[string]string{"b": "2"})
		state := newConfigOnlyPlan(map[string]string{"a": "1", "b": "2"})
		diags := testValidateInstanceUpdateContract(plan, state)
		require.True(t, diags.HasError())
		assert.Len(t, diags.Errors(), 1)
	})

	t.Run("updating instance config value is allowed", func(t *testing.T) {
		plan := newConfigOnlyPlan(map[string]string{"a": "2"})
		state := newConfigOnlyPlan(map[string]string{"a": "1"})
		diags := testValidateInstanceUpdateContract(plan, state)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})

	t.Run("null state configs are tolerated", func(t *testing.T) {
		plan := newConfigOnlyPlan(map[string]string{"a": "1"})
		state := models.KafkaInstanceResourceModel{
			Features: &models.FeaturesModel{
				InstanceConfigs: types.MapNull(types.StringType),
			},
		}
		diags := testValidateInstanceUpdateContract(plan, state)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})

	t.Run("null table topic plan is rejected after enablement", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		state.Features.TableTopic = testTableTopicObject(t, &models.TableTopicModel{
			Warehouse:   types.StringValue("warehouse"),
			CatalogType: types.StringValue("glue"),
		})
		diags := testValidateInstanceUpdateContract(plan, state)
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "cannot be disabled")
	})

	t.Run("changing enabled table topic is rejected", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.Features.TableTopic = testTableTopicObject(t, &models.TableTopicModel{
			Warehouse:   types.StringValue("new-warehouse"),
			CatalogType: types.StringValue("glue"),
		})
		state.Features.TableTopic = testTableTopicObject(t, &models.TableTopicModel{
			Warehouse:   types.StringValue("old-warehouse"),
			CatalogType: types.StringValue("glue"),
		})
		diags := testValidateInstanceUpdateContract(plan, state)
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "cannot be changed")
	})
}

func TestInstanceUpdateParamBuilder(t *testing.T) {
	t.Run("no supported changes returns empty update plan", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		assert.False(t, updatePlan.hasUpdate)
		assert.Equal(t, client.InstanceUpdateParam{}, param)
	})

	t.Run("certificate change builds security patch and wait plan", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.Features.Security = testSecurityObject(t, &models.SecurityModel{
			CertificateAuthority: types.StringValue("new-ca"),
			CertificateChain:     types.StringValue("new-chain"),
			PrivateKey:           types.StringValue("new-key"),
		})
		state.Features.Security = testSecurityObject(t, &models.SecurityModel{
			CertificateAuthority: types.StringValue("old-ca"),
			CertificateChain:     types.StringValue("old-chain"),
			PrivateKey:           types.StringValue("old-key"),
		})

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		assert.True(t, updatePlan.certificateChanged)
		require.NotNil(t, param.Features)
		require.NotNil(t, param.Features.Security)
		require.NotNil(t, param.Features.Security.CertificateAuthority)
		assert.Equal(t, "new-ca", *param.Features.Security.CertificateAuthority)
		assert.Equal(t, "new-chain", *param.Features.Security.CertificateChain)
		assert.Equal(t, "new-key", *param.Features.Security.PrivateKey)
	})

	t.Run("instance config change builds feature patch", func(t *testing.T) {
		plan := newConfigOnlyPlan(map[string]string{"auto.create.topics.enable": "false"})
		state := newConfigOnlyPlan(map[string]string{"auto.create.topics.enable": "true"})

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		assert.True(t, updatePlan.instanceConfigsChanged)
		require.NotNil(t, param.Features)
		assert.Len(t, param.Features.InstanceConfigs, 1)
		assert.Equal(t, "auto.create.topics.enable", *param.Features.InstanceConfigs[0].Key)
		assert.Equal(t, "false", *param.Features.InstanceConfigs[0].Value)
	})

	t.Run("metrics exporter disable sends enabled false patch", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.Features.MetricsExporter = types.ObjectNull(models.MetricsExporterObjectType.AttrTypes)
		state.Features.MetricsExporter = testMetricsExporterObject(t, &models.MetricsExporterModel{
			Prometheus: &models.PrometheusExporterModel{
				AuthType: types.StringValue("noauth"),
			},
		})

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		require.NotNil(t, param.Features)
		require.NotNil(t, param.Features.MetricsExporter)
		require.NotNil(t, param.Features.MetricsExporter.Prometheus)
		require.NotNil(t, param.Features.MetricsExporter.Prometheus.Enabled)
		assert.False(t, *param.Features.MetricsExporter.Prometheus.Enabled)
	})

	t.Run("file system throughput change builds spec patch", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.ComputeSpecs.FileSystemParam = testFileSystemObject(t, &models.FileSystemParamModel{
			ThroughputMibpsPerFileSystem: types.Int64Value(512),
			FileSystemCount:              types.Int64Value(2),
		})
		state.ComputeSpecs.FileSystemParam = testFileSystemObject(t, &models.FileSystemParamModel{
			ThroughputMibpsPerFileSystem: types.Int64Value(256),
			FileSystemCount:              types.Int64Value(2),
		})

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		require.NotNil(t, param.Spec)
		require.NotNil(t, param.Spec.FileSystem)
		assert.Equal(t, int32(512), param.Spec.FileSystem.ThroughputMiBpsPerFileSystem)
		assert.Equal(t, int32(2), param.Spec.FileSystem.FileSystemCount)
	})

	t.Run("name and description changes build top level patch", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.Name = types.StringValue("updated-name")
		plan.Description = types.StringValue("updated-description")
		state.Description = types.StringValue("old-description")

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		require.NotNil(t, param.Name)
		require.NotNil(t, param.Description)
		assert.Equal(t, "updated-name", *param.Name)
		assert.Equal(t, "updated-description", *param.Description)
		assert.False(t, updatePlan.shouldWait)
	})

	t.Run("version change builds wait patch", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.Version = types.StringValue("2.0.0")

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		require.NotNil(t, param.Version)
		assert.Equal(t, "2.0.0", *param.Version)
	})

	t.Run("reserved aku change builds spec patch", func(t *testing.T) {
		plan := newValidSubscriptionPlan()
		state := newValidSubscriptionPlan()
		plan.ComputeSpecs.ReservedAku = types.Int64Value(8)
		state.ComputeSpecs.ReservedAku = types.Int64Value(6)

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		require.NotNil(t, param.Spec)
		require.NotNil(t, param.Spec.ReservedAku)
		assert.Equal(t, int32(8), *param.Spec.ReservedAku)
	})

	t.Run("reserved node count change builds spec patch", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.ComputeSpecs.ReservedNodeCount = types.Int64Value(5)
		state.ComputeSpecs.ReservedNodeCount = types.Int64Value(3)

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		require.NotNil(t, param.Spec)
		require.NotNil(t, param.Spec.ReservedNodeCount)
		assert.Equal(t, int32(5), *param.Spec.ReservedNodeCount)
	})

	t.Run("schema registry enablement builds feature patch", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.Features.SchemaRegistryEnabled = types.BoolValue(true)
		state.Features.SchemaRegistryEnabled = types.BoolValue(false)

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		require.NotNil(t, param.Features)
		require.NotNil(t, param.Features.SchemaRegistryEnabled)
		assert.True(t, *param.Features.SchemaRegistryEnabled)
	})

	t.Run("adding table topic builds enable patch", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.Features.SchemaRegistryEnabled = types.BoolValue(true)
		state.Features.SchemaRegistryEnabled = types.BoolValue(true)
		plan.Features.TableTopic = testTableTopicObject(t, &models.TableTopicModel{
			Warehouse:   types.StringValue("warehouse"),
			CatalogType: types.StringValue("glue"),
		})

		param, updatePlan := testBuildInstanceUpdateParam(t, plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		assert.True(t, updatePlan.tableTopicChanged)
		require.NotNil(t, param.Features)
		require.NotNil(t, param.Features.TableTopic)
		require.NotNil(t, param.Features.TableTopic.Enabled)
		assert.True(t, *param.Features.TableTopic.Enabled)
		assert.Equal(t, "warehouse", param.Features.TableTopic.Warehouse)
		assert.Equal(t, "glue", param.Features.TableTopic.CatalogType)
	})
}

func TestInstanceUpdateStatePreservation(t *testing.T) {
	t.Run("preserves instance configs and certificates from plan", func(t *testing.T) {
		state := models.KafkaInstanceResourceModel{}
		plan := models.KafkaInstanceResourceModel{
			Features: &models.FeaturesModel{
				InstanceConfigs: mustStringMap(map[string]string{"a": "1"}),
				Security: testSecurityObject(t, &models.SecurityModel{
					CertificateAuthority: types.StringValue("ca"),
					CertificateChain:     types.StringValue("chain"),
					PrivateKey:           types.StringValue("key"),
				}),
			},
		}

		testApplyUpdateStatePreservation(t, &state, plan, instanceUpdatePlan{
			instanceConfigsChanged: true,
			certificateChanged:     true,
		})

		require.NotNil(t, state.Features)
		assert.True(t, state.Features.InstanceConfigs.Equal(plan.Features.InstanceConfigs))
		stateSecurity, stateSecurityDiags := models.SecurityObjectToModel(context.Background(), state.Features.Security)
		require.False(t, stateSecurityDiags.HasError())
		planSecurity, planSecurityDiags := models.SecurityObjectToModel(context.Background(), plan.Features.Security)
		require.False(t, planSecurityDiags.HasError())
		require.NotNil(t, stateSecurity)
		assert.Equal(t, planSecurity.CertificateAuthority, stateSecurity.CertificateAuthority)
		assert.Equal(t, planSecurity.CertificateChain, stateSecurity.CertificateChain)
		assert.Equal(t, planSecurity.PrivateKey, stateSecurity.PrivateKey)
	})

	t.Run("preserves only instance configs when certificate flag is false", func(t *testing.T) {
		state := models.KafkaInstanceResourceModel{}
		plan := models.KafkaInstanceResourceModel{
			Features: &models.FeaturesModel{
				InstanceConfigs: mustStringMap(map[string]string{"a": "1"}),
				Security: testSecurityObject(t, &models.SecurityModel{
					CertificateAuthority: types.StringValue("ca"),
				}),
			},
		}

		testApplyUpdateStatePreservation(t, &state, plan, instanceUpdatePlan{
			instanceConfigsChanged: true,
		})

		require.NotNil(t, state.Features)
		assert.True(t, state.Features.InstanceConfigs.Equal(plan.Features.InstanceConfigs))
		assert.True(t, state.Features.Security.IsNull())
	})

	t.Run("preserves only certificates when config flag is false", func(t *testing.T) {
		state := models.KafkaInstanceResourceModel{}
		plan := models.KafkaInstanceResourceModel{
			Features: &models.FeaturesModel{
				Security: testSecurityObject(t, &models.SecurityModel{
					CertificateAuthority: types.StringValue("ca"),
					CertificateChain:     types.StringValue("chain"),
					PrivateKey:           types.StringValue("key"),
				}),
			},
		}

		testApplyUpdateStatePreservation(t, &state, plan, instanceUpdatePlan{
			certificateChanged: true,
		})

		require.NotNil(t, state.Features)
		stateSecurity, stateSecurityDiags := models.SecurityObjectToModel(context.Background(), state.Features.Security)
		require.False(t, stateSecurityDiags.HasError())
		planSecurity, planSecurityDiags := models.SecurityObjectToModel(context.Background(), plan.Features.Security)
		require.False(t, planSecurityDiags.HasError())
		require.NotNil(t, stateSecurity)
		assert.Equal(t, planSecurity.CertificateAuthority, stateSecurity.CertificateAuthority)
		assert.True(t, state.Features.InstanceConfigs.IsNull())
	})

	t.Run("preserves table topic while readback may omit sensitive fields", func(t *testing.T) {
		state := models.KafkaInstanceResourceModel{}
		plan := models.KafkaInstanceResourceModel{
			Features: &models.FeaturesModel{
				TableTopic: testTableTopicObject(t, &models.TableTopicModel{
					Warehouse:    types.StringValue("warehouse"),
					CatalogType:  types.StringValue("hive"),
					KeytabFile:   types.StringValue("keytab"),
					Krb5ConfFile: types.StringValue("krb5"),
				}),
			},
		}

		testApplyUpdateStatePreservation(t, &state, plan, instanceUpdatePlan{
			tableTopicChanged: true,
		})

		require.NotNil(t, state.Features)
		stateTopic, stateTopicDiags := models.TableTopicObjectToModel(context.Background(), state.Features.TableTopic)
		require.False(t, stateTopicDiags.HasError())
		planTopic, planTopicDiags := models.TableTopicObjectToModel(context.Background(), plan.Features.TableTopic)
		require.False(t, planTopicDiags.HasError())
		require.NotNil(t, stateTopic)
		assert.Equal(t, planTopic.KeytabFile, stateTopic.KeytabFile)
		assert.Equal(t, planTopic.Krb5ConfFile, stateTopic.Krb5ConfFile)
	})
}

// Refresh tests cover the runtime readback path shared by Create/Read/Update.
func TestInstanceRefreshState(t *testing.T) {
	t.Run("running instance refreshes endpoints", func(t *testing.T) {
		api := &stubKafkaInstanceAPI{
			instance: &client.InstanceVO{
				InstanceId: testStringPtr("inst-1"),
				Name:       testStringPtr("test"),
				State:      testStringPtr(models.StateRunning),
			},
			endpoints: []client.InstanceAccessInfoVO{
				{
					DisplayName:      testStringPtr("private"),
					NetworkType:      testStringPtr("private"),
					Protocol:         testStringPtr("SASL_PLAINTEXT"),
					BootstrapServers: testStringPtr("broker:9092"),
				},
			},
		}
		resource := &KafkaInstanceResource{api: api}
		state := models.KafkaInstanceResourceModel{}

		found, diags := refreshKafkaInstanceState(context.Background(), resource, "inst-1", &state)
		if !found {
			t.Fatalf("expected instance to be found")
		}
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if api.getEndpointsCall != 1 {
			t.Fatalf("expected exactly one endpoints call, got %d", api.getEndpointsCall)
		}
	})

	t.Run("non-running instance skips endpoints", func(t *testing.T) {
		api := &stubKafkaInstanceAPI{
			instance: &client.InstanceVO{
				InstanceId: testStringPtr("inst-2"),
				Name:       testStringPtr("test"),
				State:      testStringPtr(models.StateCreating),
			},
		}
		resource := &KafkaInstanceResource{api: api}
		state := models.KafkaInstanceResourceModel{}

		found, diags := refreshKafkaInstanceState(context.Background(), resource, "inst-2", &state)
		if !found {
			t.Fatalf("expected instance to be found")
		}
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if api.getEndpointsCall != 0 {
			t.Fatalf("expected no endpoints call, got %d", api.getEndpointsCall)
		}
		if !state.Endpoints.IsNull() && !state.Endpoints.IsUnknown() && len(state.Endpoints.Elements()) != 0 {
			t.Fatalf("expected endpoints to remain unset for non-running instance")
		}
	})
	t.Run("not found is surfaced without diagnostics", func(t *testing.T) {
		resource := &KafkaInstanceResource{
			api: &stubKafkaInstanceAPI{
				getInstanceErr: &client.ErrorResponse{Code: 404},
			},
		}
		state := models.KafkaInstanceResourceModel{
			InstanceID: types.StringValue("inst-missing"),
		}

		found, diags := refreshKafkaInstanceState(context.Background(), resource, "inst-missing", &state)
		assert.False(t, found)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})
}

func newValidUsageBasedIAASPlan() models.KafkaInstanceResourceModel {
	return models.KafkaInstanceResourceModel{
		Name:        types.StringValue("test-instance"),
		Description: types.StringValue("description"),
		Version:     types.StringValue("1.0.0"),
		ComputeSpecs: &models.ComputeSpecsModel{
			DeployType:        types.StringValue("IAAS"),
			PricingMode:       types.StringValue("UsageBased"),
			ReservedNodeCount: types.Int64Value(3),
			ReservedAku:       types.Int64Value(0),
			InstanceTypes:     mustStringList("m5.large"),
		},
		Features: &models.FeaturesModel{
			WalMode:         types.StringValue("EBSWAL"),
			InstanceConfigs: types.MapNull(types.StringType),
			Security:        types.ObjectNull(models.SecurityObjectType.AttrTypes),
		},
	}
}

func newValidSubscriptionPlan() models.KafkaInstanceResourceModel {
	return models.KafkaInstanceResourceModel{
		Name:        types.StringValue("test-instance"),
		Description: types.StringValue("description"),
		Version:     types.StringValue("1.0.0"),
		ComputeSpecs: &models.ComputeSpecsModel{
			DeployType:  types.StringValue("IAAS"),
			PricingMode: types.StringValue("SubscriptionBased"),
			ReservedAku: types.Int64Value(6),
		},
		Features: &models.FeaturesModel{
			WalMode:         types.StringValue("EBSWAL"),
			InstanceConfigs: types.MapNull(types.StringType),
			Security:        types.ObjectNull(models.SecurityObjectType.AttrTypes),
		},
	}
}

func newConfigOnlyPlan(values map[string]string) models.KafkaInstanceResourceModel {
	return models.KafkaInstanceResourceModel{
		Features: &models.FeaturesModel{
			InstanceConfigs: mustStringMap(values),
		},
	}
}

func mustStringList(values ...string) types.List {
	attrs := make([]attr.Value, 0, len(values))
	for _, value := range values {
		attrs = append(attrs, types.StringValue(value))
	}
	return types.ListValueMust(types.StringType, attrs)
}

func mustStringMap(values map[string]string) types.Map {
	attrs := make(map[string]attr.Value, len(values))
	for key, value := range values {
		attrs[key] = types.StringValue(value)
	}
	return types.MapValueMust(types.StringType, attrs)
}
