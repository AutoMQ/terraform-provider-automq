package models

import (
	"fmt"
	"math"
	"strings"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"
)

type ConnectClusterResourceModel struct {
	EnvironmentID       types.String                   `tfsdk:"environment_id"`
	ID                  types.String                   `tfsdk:"id"`
	Name                types.String                   `tfsdk:"name"`
	Description         types.String                   `tfsdk:"description"`
	Plugins             []ConnectClusterPluginModel    `tfsdk:"plugins"`
	KafkaCluster        *ConnectClusterKafkaModel      `tfsdk:"kafka_cluster"`
	Capacity            *ConnectClusterCapacityModel   `tfsdk:"capacity"`
	Compute             *ConnectClusterComputeModel    `tfsdk:"compute"`
	WorkerConfig        types.Map                      `tfsdk:"worker_config"`
	MetricExporter      *ConnectorMetricsExporterModel `tfsdk:"metric_exporter"`
	Tags                types.Map                      `tfsdk:"tags"`
	Version             types.String                   `tfsdk:"version"`
	State               types.String                   `tfsdk:"state"`
	KafkaConnectVersion types.String                   `tfsdk:"kafka_connect_version"`
	CreatedAt           timetypes.RFC3339              `tfsdk:"created_at"`
	UpdatedAt           timetypes.RFC3339              `tfsdk:"updated_at"`
	Timeouts            timeouts.Value                 `tfsdk:"timeouts"`
}

type ConnectClusterPluginModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

type ConnectClusterKafkaModel struct {
	KafkaInstanceID types.String `tfsdk:"kafka_instance_id"`
}

type ConnectClusterCapacityModel struct {
	Type        types.String                            `tfsdk:"type"`
	Provisioned *ConnectClusterProvisionedCapacityModel `tfsdk:"provisioned"`
	Autoscaling *ConnectClusterAutoscalingCapacityModel `tfsdk:"autoscaling"`
}

type ConnectClusterProvisionedCapacityModel struct {
	WorkerResourceSpec types.String `tfsdk:"worker_resource_spec"`
	WorkerCount        types.Int64  `tfsdk:"worker_count"`
}

type ConnectClusterAutoscalingCapacityModel struct {
	WorkerResourceSpec types.String              `tfsdk:"worker_resource_spec"`
	MinWorkerCount     types.Int64               `tfsdk:"min_worker_count"`
	MaxWorkerCount     types.Int64               `tfsdk:"max_worker_count"`
	ScaleInPolicy      *ConnectClusterScaleModel `tfsdk:"scale_in_policy"`
	ScaleOutPolicy     *ConnectClusterScaleModel `tfsdk:"scale_out_policy"`
}

type ConnectClusterScaleModel struct {
	CPUUtilizationPercentage types.Int64 `tfsdk:"cpu_utilization_percentage"`
}

type ConnectClusterComputeModel struct {
	Type       types.String                   `tfsdk:"type"`
	Kubernetes *ConnectClusterKubernetesModel `tfsdk:"kubernetes"`
	IamRole    types.String                   `tfsdk:"iam_role"`
}

type ConnectClusterKubernetesModel struct {
	ClusterID      types.String `tfsdk:"cluster_id"`
	Namespace      types.String `tfsdk:"namespace"`
	ServiceAccount types.String `tfsdk:"service_account"`
	SchedulingSpec types.String `tfsdk:"scheduling_spec"`
}

func ExpandConnectClusterCreate(plan ConnectClusterResourceModel) (*client.ConnectClusterCreateParam, diag.Diagnostics) {
	diags := ValidateConnectClusterCapacity(plan.Capacity)
	if diags.HasError() {
		return nil, diags
	}
	request := &client.ConnectClusterCreateParam{
		Name:         plan.Name.ValueString(),
		Plugins:      expandClusterPlugins(plan.Plugins),
		KafkaCluster: expandClusterKafka(plan.KafkaCluster),
		Capacity:     expandClusterCapacity(plan.Capacity),
		Compute:      expandClusterCompute(plan.Compute),
	}
	if s := cOptStr(plan.Description); s != nil {
		request.Description = s
	}
	if m := cExpandStringMap(plan.WorkerConfig); m != nil {
		request.WorkerConfig = &client.ConnectorWorkerConfigParam{Properties: m}
	}
	if cfg := cExpandMetrics(plan.MetricExporter); cfg != nil {
		request.MetricExporter = cfg
	}
	if m := cExpandStringMap(plan.Tags); m != nil {
		request.Tags = m
	}
	if s := cOptStr(plan.Version); s != nil {
		request.Version = s
	}
	return request, diags
}

func ExpandConnectClusterUpdate(plan ConnectClusterResourceModel) (*client.ConnectClusterUpdateParam, diag.Diagnostics) {
	var diags diag.Diagnostics
	if plan.Capacity != nil {
		diags.Append(ValidateConnectClusterCapacity(plan.Capacity)...)
		if diags.HasError() {
			return nil, diags
		}
	}
	request := &client.ConnectClusterUpdateParam{
		Plugins: expandClusterPlugins(plan.Plugins),
	}
	if s := cOptStr(plan.Name); s != nil {
		request.Name = s
	}
	if s := cOptStr(plan.Description); s != nil {
		request.Description = s
	}
	if plan.Capacity != nil {
		capacity := expandClusterCapacity(plan.Capacity)
		request.Capacity = &capacity
	}
	if m := cExpandStringMap(plan.WorkerConfig); m != nil {
		request.WorkerConfig = &client.ConnectorWorkerConfigParam{Properties: m}
	}
	if cfg := cExpandMetrics(plan.MetricExporter); cfg != nil {
		request.MetricExporter = cfg
	}
	if m := cExpandStringMap(plan.Tags); m != nil {
		request.Tags = m
	}
	if s := cOptStr(plan.Version); s != nil {
		request.Version = s
	}
	return request, diags
}

func FlattenConnectCluster(vo *client.ConnectClusterVO, state *ConnectClusterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if vo == nil {
		diags.AddError("Invalid connect cluster", "nil connect cluster received")
		return diags
	}
	capacity, capacityDiags := flattenClusterCapacity(vo)
	diags.Append(capacityDiags...)
	if diags.HasError() {
		return diags
	}
	state.ID = cToStr(vo.Id)
	state.Name = cToStr(vo.Name)
	state.Description = cToStr(vo.Description)
	state.Plugins = flattenClusterPlugins(vo.Plugins)
	state.KafkaCluster = flattenClusterKafka(vo, state.KafkaCluster)
	state.Capacity = capacity
	state.Compute = flattenClusterCompute(vo, state.Compute)
	state.WorkerConfig = cFlattenInterfaceMapRetainEmpty(vo.WorkerConfig, state.WorkerConfig)
	state.MetricExporter = cFlattenMetrics(vo.MetricExporter, state.MetricExporter)
	state.Tags = cFlattenStringMap(vo.Tags)
	state.Version = cToStr(vo.Version)
	state.State = cToStr(vo.State)
	state.KafkaConnectVersion = cToStr(vo.KafkaConnectVersion)
	if vo.CreateTime != nil {
		state.CreatedAt = timetypes.NewRFC3339TimePointerValue(vo.CreateTime)
	}
	if vo.UpdateTime != nil {
		state.UpdatedAt = timetypes.NewRFC3339TimePointerValue(vo.UpdateTime)
	}
	return diags
}

func expandClusterPlugins(plugins []ConnectClusterPluginModel) []client.ClusterPluginParam {
	if len(plugins) == 0 {
		return nil
	}
	result := make([]client.ClusterPluginParam, 0, len(plugins))
	for _, plugin := range plugins {
		result = append(result, client.ClusterPluginParam{
			Name:    plugin.Name.ValueString(),
			Version: plugin.Version.ValueString(),
		})
	}
	return result
}

func expandClusterKafka(kafka *ConnectClusterKafkaModel) client.ConnectClusterKafkaParam {
	if kafka == nil {
		return client.ConnectClusterKafkaParam{}
	}
	return client.ConnectClusterKafkaParam{KafkaInstanceId: kafka.KafkaInstanceID.ValueString()}
}

func expandClusterCapacity(capacity *ConnectClusterCapacityModel) client.ConnectClusterCapacityParam {
	if capacity == nil {
		return client.ConnectClusterCapacityParam{}
	}
	result := client.ConnectClusterCapacityParam{Type: capacity.Type.ValueString()}
	if capacity.Type.ValueString() == "provisioned" && capacity.Provisioned != nil {
		result.Provisioned = &client.ConnectClusterProvisionedParam{
			WorkerResourceSpec: capacity.Provisioned.WorkerResourceSpec.ValueString(),
			WorkerCount:        int32(capacity.Provisioned.WorkerCount.ValueInt64()),
		}
	}
	if capacity.Type.ValueString() == "autoscaling" && capacity.Autoscaling != nil {
		result.Autoscaling = &client.ConnectClusterAutoscalingParam{
			WorkerResourceSpec: capacity.Autoscaling.WorkerResourceSpec.ValueString(),
			MinWorkerCount:     int32(capacity.Autoscaling.MinWorkerCount.ValueInt64()),
			MaxWorkerCount:     int32(capacity.Autoscaling.MaxWorkerCount.ValueInt64()),
			ScaleInPolicy:      expandScalePolicy(capacity.Autoscaling.ScaleInPolicy),
			ScaleOutPolicy:     expandScalePolicy(capacity.Autoscaling.ScaleOutPolicy),
		}
	}
	return result
}

func expandScalePolicy(policy *ConnectClusterScaleModel) *client.ScalePolicyParam {
	if policy == nil {
		return nil
	}
	return &client.ScalePolicyParam{CpuUtilizationPercentage: int32(policy.CPUUtilizationPercentage.ValueInt64())}
}

func expandClusterCompute(compute *ConnectClusterComputeModel) client.ConnectClusterComputeParam {
	if compute == nil {
		return client.ConnectClusterComputeParam{}
	}
	result := client.ConnectClusterComputeParam{
		Type:    compute.Type.ValueString(),
		IamRole: cOptStr(compute.IamRole),
	}
	if compute.Kubernetes != nil {
		result.Kubernetes = &client.ConnectClusterK8sParam{
			ClusterId:      compute.Kubernetes.ClusterID.ValueString(),
			Namespace:      compute.Kubernetes.Namespace.ValueString(),
			ServiceAccount: compute.Kubernetes.ServiceAccount.ValueString(),
			SchedulingSpec: cOptNormalizedYAML(compute.Kubernetes.SchedulingSpec),
		}
	}
	return result
}

func flattenClusterPlugins(plugins []client.ClusterPluginVO) []ConnectClusterPluginModel {
	if len(plugins) == 0 {
		return nil
	}
	result := make([]ConnectClusterPluginModel, 0, len(plugins))
	for _, plugin := range plugins {
		result = append(result, ConnectClusterPluginModel{
			Name:    cToStr(plugin.Name),
			Version: cToStr(plugin.Version),
		})
	}
	return result
}

func flattenClusterKafka(vo *client.ConnectClusterVO, prev *ConnectClusterKafkaModel) *ConnectClusterKafkaModel {
	if prev == nil {
		prev = &ConnectClusterKafkaModel{}
	}
	prev.KafkaInstanceID = cRetainStr(vo.KafkaInstanceId, prev.KafkaInstanceID)
	return prev
}

// ValidateConnectClusterCapacity checks the complete capacity before converting Terraform
// int64 values into API int32 values. Schema validators also run during planning.
func ValidateConnectClusterCapacity(capacity *ConnectClusterCapacityModel) diag.Diagnostics {
	var diags diag.Diagnostics
	invalid := func(detail string) { diags.AddError("Invalid Connect Cluster capacity", detail) }
	if capacity == nil {
		invalid("capacity is required.")
		return diags
	}
	validateSpec := func(spec types.String) {
		switch spec.ValueString() {
		case "TIER1", "TIER2", "TIER3", "TIER4":
		default:
			invalid("worker_resource_spec must be TIER1, TIER2, TIER3, or TIER4.")
		}
	}
	validateCount := func(name string, value types.Int64) {
		if value.IsNull() || value.IsUnknown() || value.ValueInt64() < 1 || value.ValueInt64() > math.MaxInt32 {
			invalid(fmt.Sprintf("%s must be an integer from 1 to %d, the API integer limit.", name, math.MaxInt32))
		}
	}
	switch capacity.Type.ValueString() {
	case "provisioned":
		if capacity.Provisioned == nil || capacity.Autoscaling != nil {
			invalid("provisioned capacity requires only the provisioned object.")
			return diags
		}
		validateSpec(capacity.Provisioned.WorkerResourceSpec)
		validateCount("worker_count", capacity.Provisioned.WorkerCount)
	case "autoscaling":
		if capacity.Autoscaling == nil || capacity.Provisioned != nil {
			invalid("autoscaling capacity requires only the autoscaling object.")
			return diags
		}
		auto := capacity.Autoscaling
		validateSpec(auto.WorkerResourceSpec)
		validateCount("min_worker_count", auto.MinWorkerCount)
		validateCount("max_worker_count", auto.MaxWorkerCount)
		if auto.MaxWorkerCount.ValueInt64() < auto.MinWorkerCount.ValueInt64() {
			invalid("max_worker_count must be greater than or equal to min_worker_count.")
		}
		if auto.ScaleInPolicy == nil || auto.ScaleOutPolicy == nil {
			invalid("scale_in_policy and scale_out_policy are both required for autoscaling.")
			return diags
		}
		for _, policy := range []*ConnectClusterScaleModel{auto.ScaleInPolicy, auto.ScaleOutPolicy} {
			cpu := policy.CPUUtilizationPercentage
			if cpu.IsNull() || cpu.IsUnknown() || cpu.ValueInt64() < 1 || cpu.ValueInt64() > 100 {
				invalid("cpu_utilization_percentage must be an integer from 1 to 100.")
			}
		}
		if !auto.ScaleInPolicy.CPUUtilizationPercentage.Equal(auto.ScaleOutPolicy.CPUUtilizationPercentage) {
			invalid("scale_in_policy and scale_out_policy must use the same target CPU utilization percentage.")
		}
	default:
		invalid("capacity.type must be provisioned or autoscaling.")
	}
	return diags
}

func flattenClusterCapacity(vo *client.ConnectClusterVO) (*ConnectClusterCapacityModel, diag.Diagnostics) {
	capacity := &ConnectClusterCapacityModel{}
	var diags diag.Diagnostics
	if vo.CapacityType == nil {
		diags.AddError("Invalid Connect Cluster capacity response", "The API response is missing capacityType.")
		return nil, diags
	}
	// AUTOSCALED is the API enum; lowercasing it does not produce the Terraform mode.
	switch *vo.CapacityType {
	case "PROVISIONED":
		capacity.Type = types.StringValue("provisioned")
		capacity.Provisioned = &ConnectClusterProvisionedCapacityModel{
			WorkerResourceSpec: cToStr(vo.WorkerResourceSpec),
			WorkerCount:        cToInt64(vo.WorkerCount),
		}
	case "AUTOSCALED":
		capacity.Type = types.StringValue("autoscaling")
		capacity.Autoscaling = &ConnectClusterAutoscalingCapacityModel{
			WorkerResourceSpec: cToStr(vo.WorkerResourceSpec),
			MinWorkerCount:     cToInt64(vo.MinWorkerCount),
			MaxWorkerCount:     cToInt64(vo.MaxWorkerCount),
			ScaleInPolicy:      &ConnectClusterScaleModel{CPUUtilizationPercentage: cToInt64(vo.ScaleInCpuPercent)},
			ScaleOutPolicy:     &ConnectClusterScaleModel{CPUUtilizationPercentage: cToInt64(vo.ScaleOutCpuPercent)},
		}
	default:
		diags.AddError("Invalid Connect Cluster capacity response", fmt.Sprintf("Unsupported capacityType %q.", *vo.CapacityType))
		return nil, diags
	}
	diags.Append(ValidateConnectClusterCapacity(capacity)...)
	return capacity, diags
}

func flattenClusterCompute(vo *client.ConnectClusterVO, prev *ConnectClusterComputeModel) *ConnectClusterComputeModel {
	if prev == nil {
		prev = &ConnectClusterComputeModel{Type: types.StringValue("k8s")}
	}
	prev.IamRole = cRetainStr(vo.IamRole, prev.IamRole)
	if vo.KubernetesClusterId != nil || vo.KubernetesNamespace != nil || vo.KubernetesServiceAccount != nil || vo.SchedulingSpec != nil || prev.Kubernetes != nil {
		if prev.Kubernetes == nil {
			prev.Kubernetes = &ConnectClusterKubernetesModel{}
		}
		prev.Kubernetes.ClusterID = cRetainStr(vo.KubernetesClusterId, prev.Kubernetes.ClusterID)
		prev.Kubernetes.Namespace = cRetainStr(vo.KubernetesNamespace, prev.Kubernetes.Namespace)
		prev.Kubernetes.ServiceAccount = cRetainStr(vo.KubernetesServiceAccount, prev.Kubernetes.ServiceAccount)
		prev.Kubernetes.SchedulingSpec = cRetainNormalizedYAML(vo.SchedulingSpec, prev.Kubernetes.SchedulingSpec)
	}
	return prev
}

func cOptNormalizedYAML(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := normalizeYAML(v.ValueString())
	if s == "" {
		return nil
	}
	return &s
}

func cRetainNormalizedYAML(api *string, existing types.String) types.String {
	if api != nil {
		if normalized := normalizeYAML(*api); normalized != "" {
			return types.StringValue(normalized)
		}
		return types.StringNull()
	}
	if existing.IsNull() || existing.IsUnknown() {
		return types.StringNull()
	}
	return existing
}

func normalizeYAML(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	var value interface{}
	if err := yaml.Unmarshal([]byte(trimmed), &value); err != nil {
		return trimmed
	}
	if isEmptyYAMLValue(value) {
		return ""
	}
	out, err := yaml.Marshal(value)
	if err != nil {
		return trimmed
	}
	return strings.TrimSpace(string(out))
}

func isEmptyYAMLValue(value interface{}) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case map[string]interface{}:
		return len(typed) == 0
	case []interface{}:
		return len(typed) == 0
	default:
		return false
	}
}
