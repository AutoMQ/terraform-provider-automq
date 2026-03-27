package models

import (
	"fmt"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectorResourceModel struct {
	EnvironmentID            types.String                   `tfsdk:"environment_id"`
	ID                       types.String                   `tfsdk:"id"`
	Name                     types.String                   `tfsdk:"name"`
	Description              types.String                   `tfsdk:"description"`
	PluginID                 types.String                   `tfsdk:"plugin_id"`
	PluginType               types.String                   `tfsdk:"plugin_type"`
	ConnectorClass           types.String                   `tfsdk:"connector_class"`
	IamRole                  types.String                   `tfsdk:"iam_role"`
	KubernetesClusterID      types.String                   `tfsdk:"kubernetes_cluster_id"`
	KubernetesNamespace      types.String                   `tfsdk:"kubernetes_namespace"`
	KubernetesServiceAccount types.String                   `tfsdk:"kubernetes_service_account"`
	Capacity                 *ConnectorCapacityModel        `tfsdk:"capacity"`
	TaskCount                types.Int64                    `tfsdk:"task_count"`
	WorkerConfig             types.Map                      `tfsdk:"worker_config"`
	ConnectorConfig          types.Map                      `tfsdk:"connector_config"`
	KafkaCluster             *ConnectorKafkaClusterModel    `tfsdk:"kafka_cluster"`
	MetricExporter           *ConnectorMetricsExporterModel `tfsdk:"metric_exporter"`
	Labels                   types.Map                      `tfsdk:"labels"`
	Version                  types.String                   `tfsdk:"version"`
	SchedulingSpec           types.String                   `tfsdk:"scheduling_spec"`
	State                    types.String                   `tfsdk:"state"`
	KafkaConnectVersion      types.String                   `tfsdk:"kafka_connect_version"`
	CreatedAt                timetypes.RFC3339              `tfsdk:"created_at"`
	UpdatedAt                timetypes.RFC3339              `tfsdk:"updated_at"`
	Timeouts                 timeouts.Value                 `tfsdk:"timeouts"`
}

type ConnectorCapacityModel struct {
	WorkerCount        types.Int64  `tfsdk:"worker_count"`
	WorkerResourceSpec types.String `tfsdk:"worker_resource_spec"`
}

type ConnectorKafkaClusterModel struct {
	KafkaInstanceID types.String                 `tfsdk:"kafka_instance_id"`
	Security        *SecurityProtocolConfigModel `tfsdk:"security_protocol"`
}

type SecurityProtocolConfigModel struct {
	SecurityProtocol types.String `tfsdk:"security_protocol"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
	SaslMechanism    types.String `tfsdk:"sasl_mechanism"`
	TruststoreCerts  types.String `tfsdk:"truststore_certs"`
	KeyPassword      types.String `tfsdk:"key_password"`
	ClientCert       types.String `tfsdk:"client_cert"`
	PrivateKey       types.String `tfsdk:"private_key"`
}

type ConnectorMetricsExporterModel struct {
	RemoteWrite *ConnectorRemoteWriteModel `tfsdk:"remote_write"`
}

type ConnectorRemoteWriteModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	Endpoint           types.String `tfsdk:"endpoint"`
	AuthType           types.String `tfsdk:"auth_type"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Token              types.String `tfsdk:"token"`
	Region             types.String `tfsdk:"region"`
	PrometheusArn      types.String `tfsdk:"prometheus_arn"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	Headers            types.Map    `tfsdk:"headers"`
	Labels             types.Map    `tfsdk:"labels"`
}

// ExpandConnectorCreate converts the Terraform plan into an API create request.
func ExpandConnectorCreate(plan ConnectorResourceModel) (*client.ConnectorCreateParam, diag.Diagnostics) {
	var diags diag.Diagnostics
	if plan.Capacity == nil {
		diags.AddError("Missing capacity", "capacity block must be specified")
		return nil, diags
	}
	if plan.KafkaCluster == nil || plan.KafkaCluster.Security == nil {
		diags.AddError("Missing kafka_cluster", "kafka_cluster with security_protocol must be specified")
		return nil, diags
	}

	request := &client.ConnectorCreateParam{
		Name:                     plan.Name.ValueString(),
		KubernetesClusterId:      plan.KubernetesClusterID.ValueString(),
		PluginId:                 plan.PluginID.ValueString(),
		Capacity:                 cExpandCapacity(plan.Capacity),
		TaskCount:                int32(plan.TaskCount.ValueInt64()),
		KubernetesServiceAccount: plan.KubernetesServiceAccount.ValueString(),
		KubernetesNamespace:      plan.KubernetesNamespace.ValueString(),
		KafkaCluster: client.ConnectorKafkaClusterParam{
			KafkaInstanceId:        plan.KafkaCluster.KafkaInstanceID.ValueString(),
			SecurityProtocolConfig: *cExpandSecurityProtocol(plan.KafkaCluster.Security),
		},
	}
	if s := cOptStr(plan.Description); s != nil {
		request.Description = s
	}
	if s := cOptStr(plan.PluginType); s != nil {
		request.Type = s
	}
	connClass := plan.ConnectorClass.ValueString()
	request.ConnectorClass = &connClass
	if s := cOptStr(plan.IamRole); s != nil {
		request.IamRole = s
	}
	if s := cOptStr(plan.Version); s != nil {
		request.Version = s
	}
	if s := cOptStr(plan.SchedulingSpec); s != nil {
		request.SchedulingSpec = s
	}
	if m := cExpandStringMap(plan.WorkerConfig); m != nil {
		request.WorkerConfig = &client.ConnectorWorkerConfigParam{Properties: m}
	}
	if m := cExpandStringMap(plan.ConnectorConfig); m != nil {
		request.ConnectorConfig = &client.ConnectorConnectorConfigParam{Properties: m}
	}
	if cfg := cExpandMetrics(plan.MetricExporter); cfg != nil {
		request.MetricExporter = cfg
	}
	return request, diags
}

// ExpandConnectorUpdate converts the Terraform plan into an API update request.
func ExpandConnectorUpdate(plan ConnectorResourceModel) (*client.ConnectorUpdateParam, diag.Diagnostics) {
	var diags diag.Diagnostics
	request := &client.ConnectorUpdateParam{}
	if s := cOptStr(plan.Name); s != nil {
		request.Name = s
	}
	if s := cOptStr(plan.Description); s != nil {
		request.Description = s
	}
	if s := cOptStr(plan.PluginID); s != nil {
		request.PluginId = s
	}
	if !plan.TaskCount.IsNull() && !plan.TaskCount.IsUnknown() {
		tc := int32(plan.TaskCount.ValueInt64())
		request.TaskCount = &tc
	}
	if plan.Capacity != nil {
		expanded := cExpandCapacity(plan.Capacity)
		request.Capacity = &expanded
	}
	if plan.KafkaCluster != nil && plan.KafkaCluster.Security != nil {
		request.SecurityProtocolConfig = cExpandSecurityProtocol(plan.KafkaCluster.Security)
	}
	if s := cOptStr(plan.Version); s != nil {
		request.Version = s
	}
	if s := cOptStr(plan.SchedulingSpec); s != nil {
		request.SchedulingSpec = s
	}
	if m := cExpandStringMap(plan.WorkerConfig); m != nil {
		request.WorkerConfig = &client.ConnectorWorkerConfigParam{Properties: m}
	}
	if m := cExpandStringMap(plan.ConnectorConfig); m != nil {
		request.ConnectorConfig = &client.ConnectorConnectorConfigParam{Properties: m}
	}
	if m := cExpandStringMap(plan.Labels); m != nil {
		request.Labels = m
	}
	if cfg := cExpandMetrics(plan.MetricExporter); cfg != nil {
		request.MetricExporter = cfg
	}
	return request, diags
}

// FlattenConnector maps an API response into the Terraform state model.
func FlattenConnector(vo *client.ConnectorVO, state *ConnectorResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if vo == nil {
		diags.AddError("Invalid connector", "nil connector received")
		return diags
	}
	state.ID = cToStr(vo.Id)
	state.Name = cToStr(vo.Name)
	state.Description = cToStr(vo.Description)
	if vo.Plugin != nil && vo.Plugin.Id != nil {
		state.PluginID = types.StringValue(*vo.Plugin.Id)
	}
	state.PluginType = cToStr(vo.ConnType)
	state.ConnectorClass = cToStr(vo.ConnClass)
	state.IamRole = cToStr(vo.IamRole)
	state.KubernetesClusterID = cToStr(vo.KubernetesClusterId)
	state.KubernetesNamespace = cToStr(vo.KubernetesNamespace)
	state.KubernetesServiceAccount = cToStr(vo.KubernetesServiceAccount)
	state.TaskCount = cToInt64(vo.TaskCount)
	state.State = cToStr(vo.State)
	state.Version = cToStr(vo.Version)
	state.KafkaConnectVersion = cToStr(vo.KafkaConnectVersion)
	state.SchedulingSpec = cToStr(vo.SchedulingSpec)
	if vo.CreateTime != nil {
		state.CreatedAt = timetypes.NewRFC3339TimePointerValue(vo.CreateTime)
	}
	if vo.UpdateTime != nil {
		state.UpdatedAt = timetypes.NewRFC3339TimePointerValue(vo.UpdateTime)
	}
	state.Capacity = cFlattenCapacity(vo, state.Capacity)
	state.KafkaCluster = cFlattenKafkaCluster(vo, state.KafkaCluster)
	state.WorkerConfig = cFlattenInterfaceMap(vo.WorkerConfig)
	state.ConnectorConfig = cFlattenInterfaceMap(vo.ConnectorConfig)
	state.Labels = cFlattenStringMap(vo.Labels)
	state.MetricExporter = cFlattenMetrics(vo.MetricExporter, state.MetricExporter)
	return diags
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func cExpandCapacity(m *ConnectorCapacityModel) client.ConnectorCapacityParam {
	return client.ConnectorCapacityParam{
		WorkerCount:        int32(m.WorkerCount.ValueInt64()),
		WorkerResourceSpec: m.WorkerResourceSpec.ValueString(),
	}
}

func cExpandSecurityProtocol(m *SecurityProtocolConfigModel) *client.SecurityProtocolConfig {
	if m == nil {
		return nil
	}
	return &client.SecurityProtocolConfig{
		SecurityProtocol: cOptStr(m.SecurityProtocol),
		Username:         cOptStr(m.Username),
		Password:         cOptStr(m.Password),
		SaslMechanism:    cOptStr(m.SaslMechanism),
		TruststoreCerts:  cOptStr(m.TruststoreCerts),
		KeyPassword:      cOptStr(m.KeyPassword),
		ClientCert:       cOptStr(m.ClientCert),
		PrivateKey:       cOptStr(m.PrivateKey),
	}
}

func cExpandMetrics(m *ConnectorMetricsExporterModel) *client.ConnectMetricsConfigParam {
	if m == nil || m.RemoteWrite == nil {
		return nil
	}
	rw := m.RemoteWrite
	return &client.ConnectMetricsConfigParam{
		RemoteWrite: &client.ConnectRemoteWriteConfigParam{
			Enabled:            cOptBool(rw.Enabled),
			EndPoint:           cOptStr(rw.Endpoint),
			AuthType:           cOptStr(rw.AuthType),
			Username:           cOptStr(rw.Username),
			Password:           cOptStr(rw.Password),
			Token:              cOptStr(rw.Token),
			Region:             cOptStr(rw.Region),
			PrometheusArn:      cOptStr(rw.PrometheusArn),
			InsecureSkipVerify: cOptBool(rw.InsecureSkipVerify),
			Headers:            cExpandStringMap(rw.Headers),
			Labels:             cExpandStringMap(rw.Labels),
		},
	}
}

func cExpandStringMap(v types.Map) map[string]string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	result := make(map[string]string, len(v.Elements()))
	for k, val := range v.Elements() {
		sv, ok := val.(types.String)
		if !ok || sv.IsNull() || sv.IsUnknown() {
			continue
		}
		result[k] = sv.ValueString()
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func cFlattenCapacity(vo *client.ConnectorVO, prev *ConnectorCapacityModel) *ConnectorCapacityModel {
	m := &ConnectorCapacityModel{WorkerCount: types.Int64Null(), WorkerResourceSpec: types.StringNull()}
	if prev != nil {
		m = prev
	}
	if vo.WorkerCount != nil {
		m.WorkerCount = types.Int64Value(int64(*vo.WorkerCount))
	}
	if vo.WorkerResourceSpec != nil {
		m.WorkerResourceSpec = types.StringValue(*vo.WorkerResourceSpec)
	}
	return m
}

func cFlattenKafkaCluster(vo *client.ConnectorVO, prev *ConnectorKafkaClusterModel) *ConnectorKafkaClusterModel {
	if prev == nil {
		prev = &ConnectorKafkaClusterModel{KafkaInstanceID: types.StringNull()}
	}
	if vo.KafkaInstanceId != nil {
		prev.KafkaInstanceID = types.StringValue(*vo.KafkaInstanceId)
	}
	prev.Security = cFlattenSecurityProtocol(vo.SecurityProtocolConfig, prev.Security)
	return prev
}

func cFlattenSecurityProtocol(cfg *client.SecurityProtocolConfig, prev *SecurityProtocolConfigModel) *SecurityProtocolConfigModel {
	m := &SecurityProtocolConfigModel{
		SecurityProtocol: types.StringNull(),
		Username:         types.StringNull(),
		Password:         types.StringNull(),
		SaslMechanism:    types.StringNull(),
		TruststoreCerts:  types.StringNull(),
		KeyPassword:      types.StringNull(),
		ClientCert:       types.StringNull(),
		PrivateKey:       types.StringNull(),
	}
	if prev != nil {
		m = prev
	}
	if cfg == nil {
		return m
	}
	m.SecurityProtocol = cRetainStr(cfg.SecurityProtocol, m.SecurityProtocol)
	m.Username = cRetainStr(cfg.Username, m.Username)
	m.Password = cRetainStr(cfg.Password, m.Password)
	m.SaslMechanism = cRetainStr(cfg.SaslMechanism, m.SaslMechanism)
	m.TruststoreCerts = cRetainStr(cfg.TruststoreCerts, m.TruststoreCerts)
	m.KeyPassword = cRetainStr(cfg.KeyPassword, m.KeyPassword)
	m.ClientCert = cRetainStr(cfg.ClientCert, m.ClientCert)
	m.PrivateKey = cRetainStr(cfg.PrivateKey, m.PrivateKey)
	return m
}

func cFlattenMetrics(cfg *client.ConnectMetricsConfigVO, prev *ConnectorMetricsExporterModel) *ConnectorMetricsExporterModel {
	if cfg == nil || cfg.RemoteWrite == nil {
		return nil
	}
	m := &ConnectorMetricsExporterModel{}
	if prev != nil {
		m = prev
	}
	m.RemoteWrite = cFlattenRemoteWrite(cfg.RemoteWrite, m.RemoteWrite)
	return m
}

func cFlattenRemoteWrite(cfg *client.ConnectRemoteWriteConfigVO, prev *ConnectorRemoteWriteModel) *ConnectorRemoteWriteModel {
	m := &ConnectorRemoteWriteModel{
		Enabled: types.BoolNull(), Endpoint: types.StringNull(), AuthType: types.StringNull(),
		Username: types.StringNull(), Password: types.StringNull(), Token: types.StringNull(),
		Region: types.StringNull(), PrometheusArn: types.StringNull(), InsecureSkipVerify: types.BoolNull(),
		Headers: types.MapNull(types.StringType), Labels: types.MapNull(types.StringType),
	}
	if prev != nil {
		m = prev
	}
	m.Enabled = types.BoolValue(cfg.Enabled)
	m.Endpoint = cRetainStr(cfg.EndPoint, m.Endpoint)
	m.AuthType = cRetainStr(cfg.AuthType, m.AuthType)
	m.Username = cRetainStr(cfg.Username, m.Username)
	m.Region = cRetainStr(cfg.Region, m.Region)
	m.PrometheusArn = cRetainStr(cfg.PrometheusArn, m.PrometheusArn)
	m.InsecureSkipVerify = types.BoolValue(cfg.InsecureSkipVerify)
	m.Headers = cFlattenStringMap(cfg.Headers)
	m.Labels = cFlattenStringMap(cfg.Labels)
	return m
}

func cFlattenInterfaceMap(src map[string]interface{}) types.Map {
	if len(src) == 0 {
		return types.MapNull(types.StringType)
	}
	vals := make(map[string]attr.Value, len(src))
	for k, v := range src {
		vals[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	return types.MapValueMust(types.StringType, vals)
}

func cFlattenStringMap(src map[string]string) types.Map {
	if len(src) == 0 {
		return types.MapNull(types.StringType)
	}
	vals := make(map[string]attr.Value, len(src))
	for k, v := range src {
		vals[k] = types.StringValue(v)
	}
	return types.MapValueMust(types.StringType, vals)
}

func cRetainStr(api *string, existing types.String) types.String {
	if api != nil {
		return types.StringValue(*api)
	}
	if existing.IsNull() || existing.IsUnknown() {
		return types.StringNull()
	}
	return existing
}

func cOptStr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func cOptBool(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

func cToStr(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func cToInt64(v *int32) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}
