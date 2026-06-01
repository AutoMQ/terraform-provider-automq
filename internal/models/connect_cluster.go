package models

import (
	"strings"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	var diags diag.Diagnostics
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
	state.ID = cToStr(vo.Id)
	state.Name = cToStr(vo.Name)
	state.Description = cToStr(vo.Description)
	state.Plugins = flattenClusterPlugins(vo.Plugins)
	state.KafkaCluster = flattenClusterKafka(vo, state.KafkaCluster)
	state.Capacity = flattenClusterCapacity(vo, state.Capacity)
	state.Compute = flattenClusterCompute(vo, state.Compute)
	state.WorkerConfig = cFlattenInterfaceMap(vo.WorkerConfig)
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
	if capacity.Provisioned != nil {
		result.Provisioned = &client.ConnectClusterProvisionedParam{
			WorkerResourceSpec: capacity.Provisioned.WorkerResourceSpec.ValueString(),
			WorkerCount:        int32(capacity.Provisioned.WorkerCount.ValueInt64()),
		}
	}
	if capacity.Autoscaling != nil {
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
			SchedulingSpec: cOptStr(compute.Kubernetes.SchedulingSpec),
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

func flattenClusterCapacity(vo *client.ConnectClusterVO, prev *ConnectClusterCapacityModel) *ConnectClusterCapacityModel {
	if prev == nil {
		prev = &ConnectClusterCapacityModel{}
	}
	prev.Type = cRetainStrLower(vo.CapacityType, prev.Type)
	if vo.WorkerCount != nil || vo.WorkerResourceSpec != nil || prev.Provisioned != nil {
		if prev.Provisioned == nil {
			prev.Provisioned = &ConnectClusterProvisionedCapacityModel{}
		}
		prev.Provisioned.WorkerCount = cToInt64(vo.WorkerCount)
		prev.Provisioned.WorkerResourceSpec = cRetainStr(vo.WorkerResourceSpec, prev.Provisioned.WorkerResourceSpec)
	}
	if vo.MinWorkerCount != nil || vo.MaxWorkerCount != nil || prev.Autoscaling != nil {
		if prev.Autoscaling == nil {
			prev.Autoscaling = &ConnectClusterAutoscalingCapacityModel{}
		}
		prev.Autoscaling.WorkerResourceSpec = cRetainStr(vo.WorkerResourceSpec, prev.Autoscaling.WorkerResourceSpec)
		prev.Autoscaling.MinWorkerCount = cToInt64(vo.MinWorkerCount)
		prev.Autoscaling.MaxWorkerCount = cToInt64(vo.MaxWorkerCount)
		if vo.ScaleInCpuPercent != nil {
			prev.Autoscaling.ScaleInPolicy = &ConnectClusterScaleModel{CPUUtilizationPercentage: cToInt64(vo.ScaleInCpuPercent)}
		}
		if vo.ScaleOutCpuPercent != nil {
			prev.Autoscaling.ScaleOutPolicy = &ConnectClusterScaleModel{CPUUtilizationPercentage: cToInt64(vo.ScaleOutCpuPercent)}
		}
	}
	return prev
}

func cRetainStrLower(api *string, existing types.String) types.String {
	if api != nil {
		return types.StringValue(strings.ToLower(*api))
	}
	if existing.IsNull() || existing.IsUnknown() {
		return types.StringNull()
	}
	return existing
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
		prev.Kubernetes.SchedulingSpec = cRetainStr(vo.SchedulingSpec, prev.Kubernetes.SchedulingSpec)
	}
	return prev
}
