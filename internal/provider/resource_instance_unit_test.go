package provider

import (
	"context"
	"errors"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/models"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStringPtr(s string) *string {
	return &s
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
				KubernetesNodeGroups: nil,
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
		plan.Features.MetricsExporter = &models.MetricsExporterModel{
			Prometheus: &models.PrometheusExporterModel{
				AuthType: types.StringNull(),
			},
		}
		diags := validateInstanceContract(context.Background(), &plan)
		require.True(t, diags.HasError())
		assert.Len(t, diags.Errors(), 1)
	})
}

func TestInstanceUpdateContractValidation(t *testing.T) {
	t.Run("removing instance config key is rejected", func(t *testing.T) {
		plan := newConfigOnlyPlan(map[string]string{"b": "2"})
		state := newConfigOnlyPlan(map[string]string{"a": "1", "b": "2"})
		diags := validateInstanceUpdateContract("inst-1", plan, state)
		require.True(t, diags.HasError())
		assert.Len(t, diags.Errors(), 1)
	})

	t.Run("updating instance config value is allowed", func(t *testing.T) {
		plan := newConfigOnlyPlan(map[string]string{"a": "2"})
		state := newConfigOnlyPlan(map[string]string{"a": "1"})
		diags := validateInstanceUpdateContract("inst-1", plan, state)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})

	t.Run("null state configs are tolerated", func(t *testing.T) {
		plan := newConfigOnlyPlan(map[string]string{"a": "1"})
		state := models.KafkaInstanceResourceModel{
			Features: &models.FeaturesModel{
				InstanceConfigs: types.MapNull(types.StringType),
			},
		}
		diags := validateInstanceUpdateContract("inst-1", plan, state)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})
}

func TestInstanceUpdateParamBuilder(t *testing.T) {
	t.Run("no supported changes returns empty update plan", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()

		param, updatePlan := buildInstanceUpdateParam(plan, state)
		assert.False(t, updatePlan.hasUpdate)
		assert.Equal(t, client.InstanceUpdateParam{}, param)
	})

	t.Run("certificate change builds security patch and wait plan", func(t *testing.T) {
		plan := newValidUsageBasedIAASPlan()
		state := newValidUsageBasedIAASPlan()
		plan.Features.Security = &models.SecurityModel{
			CertificateAuthority: types.StringValue("new-ca"),
			CertificateChain:     types.StringValue("new-chain"),
			PrivateKey:           types.StringValue("new-key"),
		}
		state.Features.Security = &models.SecurityModel{
			CertificateAuthority: types.StringValue("old-ca"),
			CertificateChain:     types.StringValue("old-chain"),
			PrivateKey:           types.StringValue("old-key"),
		}

		param, updatePlan := buildInstanceUpdateParam(plan, state)
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

		param, updatePlan := buildInstanceUpdateParam(plan, state)
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
		plan.Features.MetricsExporter = nil
		state.Features.MetricsExporter = &models.MetricsExporterModel{
			Prometheus: &models.PrometheusExporterModel{
				AuthType: types.StringValue("noauth"),
			},
		}

		param, updatePlan := buildInstanceUpdateParam(plan, state)
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
		plan.ComputeSpecs.FileSystemParam = &models.FileSystemParamModel{
			ThroughputMibpsPerFileSystem: types.Int64Value(512),
			FileSystemCount:              types.Int64Value(2),
		}
		state.ComputeSpecs.FileSystemParam = &models.FileSystemParamModel{
			ThroughputMibpsPerFileSystem: types.Int64Value(256),
			FileSystemCount:              types.Int64Value(2),
		}

		param, updatePlan := buildInstanceUpdateParam(plan, state)
		require.True(t, updatePlan.hasUpdate)
		assert.True(t, updatePlan.shouldWait)
		require.NotNil(t, param.Spec)
		require.NotNil(t, param.Spec.FileSystem)
		assert.Equal(t, int32(512), param.Spec.FileSystem.ThroughputMiBpsPerFileSystem)
		assert.Equal(t, int32(2), param.Spec.FileSystem.FileSystemCount)
	})
}

func TestInstanceUpdateStatePreservation(t *testing.T) {
	t.Run("preserves instance configs and certificates from plan", func(t *testing.T) {
		state := models.KafkaInstanceResourceModel{}
		plan := models.KafkaInstanceResourceModel{
			Features: &models.FeaturesModel{
				InstanceConfigs: mustStringMap(map[string]string{"a": "1"}),
				Security: &models.SecurityModel{
					CertificateAuthority: types.StringValue("ca"),
					CertificateChain:     types.StringValue("chain"),
					PrivateKey:           types.StringValue("key"),
				},
			},
		}

		applyUpdateStatePreservation(&state, plan, instanceUpdatePlan{
			instanceConfigsChanged: true,
			certificateChanged:     true,
		})

		require.NotNil(t, state.Features)
		assert.True(t, state.Features.InstanceConfigs.Equal(plan.Features.InstanceConfigs))
		require.NotNil(t, state.Features.Security)
		assert.Equal(t, plan.Features.Security.CertificateAuthority, state.Features.Security.CertificateAuthority)
		assert.Equal(t, plan.Features.Security.CertificateChain, state.Features.Security.CertificateChain)
		assert.Equal(t, plan.Features.Security.PrivateKey, state.Features.Security.PrivateKey)
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

		_, found, diags := refreshKafkaInstanceState(context.Background(), resource, "inst-1", &state)
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

		_, found, diags := refreshKafkaInstanceState(context.Background(), resource, "inst-2", &state)
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

		_, found, diags := refreshKafkaInstanceState(context.Background(), resource, "inst-missing", &state)
		assert.False(t, found)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})
}

func newValidUsageBasedIAASPlan() models.KafkaInstanceResourceModel {
	return models.KafkaInstanceResourceModel{
		Name:    types.StringValue("test-instance"),
		Version: types.StringValue("1.0.0"),
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
			Security:        &models.SecurityModel{},
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
