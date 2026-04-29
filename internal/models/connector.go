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
	EnvironmentID            types.String                `tfsdk:"environment_id"`
	ID                       types.String                `tfsdk:"id"`
	ConnectClusterID         types.String                `tfsdk:"connect_cluster_id"`
	Name                     types.String                `tfsdk:"name"`
	Description              types.String                `tfsdk:"description"`
	ConnectorClass           types.String                `tfsdk:"connector_class"`
	TaskCount                types.Int64                 `tfsdk:"task_count"`
	KafkaCluster             *ConnectorKafkaClusterModel `tfsdk:"kafka_cluster"`
	ConnectorConfig          types.Map                   `tfsdk:"connector_config"`
	ConnectorConfigSensitive types.Map                   `tfsdk:"connector_config_sensitive"`
	InitialOffsets           []InitialOffsetModel        `tfsdk:"initial_offsets"`
	State                    types.String                `tfsdk:"state"`
	ConnectorType            types.String                `tfsdk:"connector_type"`
	PluginID                 types.String                `tfsdk:"plugin_id"`
	CreatedAt                timetypes.RFC3339           `tfsdk:"created_at"`
	UpdatedAt                timetypes.RFC3339           `tfsdk:"updated_at"`
	Timeouts                 timeouts.Value              `tfsdk:"timeouts"`
}

type ConnectorKafkaClusterModel struct {
	Security *SecurityProtocolConfigModel `tfsdk:"security_protocol"`
}

type SecurityProtocolConfigModel struct {
	Protocol        types.String `tfsdk:"protocol"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
	SaslMechanism   types.String `tfsdk:"sasl_mechanism"`
	TruststoreCerts types.String `tfsdk:"truststore_certs"`
	ClientCert      types.String `tfsdk:"client_cert"`
	PrivateKey      types.String `tfsdk:"private_key"`
}

type InitialOffsetModel struct {
	Partition types.Map `tfsdk:"partition"`
	Offset    types.Map `tfsdk:"offset"`
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

func ExpandConnectorCreate(plan ConnectorResourceModel) (*client.ConnectorCreateParam, diag.Diagnostics) {
	var diags diag.Diagnostics
	request := &client.ConnectorCreateParam{
		ConnectClusterId: plan.ConnectClusterID.ValueString(),
		Name:             plan.Name.ValueString(),
		ConnectorClass:   plan.ConnectorClass.ValueString(),
		TaskCount:        int32(plan.TaskCount.ValueInt64()),
	}
	if s := cOptStr(plan.Description); s != nil {
		request.Description = s
	}
	if plan.KafkaCluster != nil && plan.KafkaCluster.Security != nil {
		request.KafkaCluster = &client.ConnectorKafkaClusterParam{
			SecurityProtocolConfig: cExpandSecurityProtocol(plan.KafkaCluster.Security),
		}
	}
	if m := cExpandStringMap(plan.ConnectorConfig); m != nil {
		request.ConnectorConfig = &client.ConnectorConnectorConfigParam{Properties: m}
	}
	if m := cExpandStringMap(plan.ConnectorConfigSensitive); m != nil {
		request.ConnectorConfigSensitive = &client.ConnectorConnectorConfigParam{Properties: m}
	}
	if offsets := cExpandInitialOffsets(plan.InitialOffsets); len(offsets) > 0 {
		request.InitialOffsets = offsets
	}
	return request, diags
}

func ExpandConnectorUpdate(plan ConnectorResourceModel) (*client.ConnectorUpdateParam, diag.Diagnostics) {
	var diags diag.Diagnostics
	request := &client.ConnectorUpdateParam{}
	if s := cOptStr(plan.Name); s != nil {
		request.Name = s
	}
	if s := cOptStr(plan.Description); s != nil {
		request.Description = s
	}
	if !plan.TaskCount.IsNull() && !plan.TaskCount.IsUnknown() {
		tc := int32(plan.TaskCount.ValueInt64())
		request.TaskCount = &tc
	}
	if plan.KafkaCluster != nil && plan.KafkaCluster.Security != nil {
		request.SecurityProtocolConfig = cExpandSecurityProtocol(plan.KafkaCluster.Security)
	}
	if m := cExpandStringMap(plan.ConnectorConfig); m != nil {
		request.ConnectorConfig = &client.ConnectorConnectorConfigParam{Properties: m}
	}
	if m := cExpandStringMap(plan.ConnectorConfigSensitive); m != nil {
		request.ConnectorConfigSensitive = &client.ConnectorConnectorConfigParam{Properties: m}
	}
	return request, diags
}

func FlattenConnector(vo *client.ConnectorVO, state *ConnectorResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if vo == nil {
		diags.AddError("Invalid connector", "nil connector received")
		return diags
	}
	state.ID = cToStr(vo.Id)
	state.ConnectClusterID = cRetainStr(vo.ConnectClusterId, state.ConnectClusterID)
	state.Name = cToStr(vo.Name)
	state.Description = cToStr(vo.Description)
	state.TaskCount = cToInt64(vo.TaskCount)
	state.State = cToStr(vo.State)
	state.ConnectorType = cToStr(firstString(vo.ConnectorType, vo.ConnType))
	state.ConnectorClass = cToStr(firstString(vo.ConnectorClass, vo.ConnClass))
	state.PluginID = cToStr(firstString(vo.PluginId, pluginSummaryID(vo.Plugin)))
	if vo.CreateTime != nil {
		state.CreatedAt = timetypes.NewRFC3339TimePointerValue(vo.CreateTime)
	}
	if vo.UpdateTime != nil {
		state.UpdatedAt = timetypes.NewRFC3339TimePointerValue(vo.UpdateTime)
	}
	state.KafkaCluster = cFlattenConnectorKafkaCluster(vo.SecurityProtocolConfig, state.KafkaCluster)
	state.ConnectorConfig = cFlattenInterfaceMap(vo.ConnectorConfig)
	state.ConnectorConfigSensitive = cRetainMap(cFlattenInterfaceMap(vo.ConnectorConfigSensitive), state.ConnectorConfigSensitive)
	return diags
}

func cExpandInitialOffsets(offsets []InitialOffsetModel) []client.InitialOffsetParam {
	if len(offsets) == 0 {
		return nil
	}
	result := make([]client.InitialOffsetParam, 0, len(offsets))
	for _, offset := range offsets {
		result = append(result, client.InitialOffsetParam{
			Partition: cExpandStringMap(offset.Partition),
			Offset:    cExpandStringMap(offset.Offset),
		})
	}
	return result
}

func cExpandSecurityProtocol(m *SecurityProtocolConfigModel) *client.SecurityProtocolConfig {
	if m == nil {
		return nil
	}
	protocol := cOptStr(m.Protocol)
	return &client.SecurityProtocolConfig{
		Protocol:         protocol,
		SecurityProtocol: protocol,
		Username:         cOptStr(m.Username),
		Password:         cOptStr(m.Password),
		SaslMechanism:    cOptStr(m.SaslMechanism),
		TruststoreCerts:  cOptStr(m.TruststoreCerts),
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

func cFlattenConnectorKafkaCluster(cfg *client.SecurityProtocolConfig, prev *ConnectorKafkaClusterModel) *ConnectorKafkaClusterModel {
	if cfg == nil && prev == nil {
		return nil
	}
	if prev == nil {
		prev = &ConnectorKafkaClusterModel{}
	}
	prev.Security = cFlattenSecurityProtocol(cfg, prev.Security)
	return prev
}

func cFlattenSecurityProtocol(cfg *client.SecurityProtocolConfig, prev *SecurityProtocolConfigModel) *SecurityProtocolConfigModel {
	m := &SecurityProtocolConfigModel{
		Protocol:        types.StringNull(),
		Username:        types.StringNull(),
		Password:        types.StringNull(),
		SaslMechanism:   types.StringNull(),
		TruststoreCerts: types.StringNull(),
		ClientCert:      types.StringNull(),
		PrivateKey:      types.StringNull(),
	}
	if prev != nil {
		m = prev
	}
	if cfg == nil {
		return m
	}
	m.Protocol = cRetainStr(firstString(cfg.Protocol, cfg.SecurityProtocol), m.Protocol)
	m.Username = cRetainStr(cfg.Username, m.Username)
	m.Password = cRetainStr(cfg.Password, m.Password)
	m.SaslMechanism = cRetainStr(cfg.SaslMechanism, m.SaslMechanism)
	m.TruststoreCerts = cRetainStr(cfg.TruststoreCerts, m.TruststoreCerts)
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

func cRetainMap(api types.Map, existing types.Map) types.Map {
	if !api.IsNull() && !api.IsUnknown() {
		return api
	}
	if existing.IsNull() || existing.IsUnknown() {
		return types.MapNull(types.StringType)
	}
	return existing
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

func firstString(values ...*string) *string {
	for _, value := range values {
		if value != nil && *value != "" {
			return value
		}
	}
	return nil
}

func pluginSummaryID(plugin *client.ConnectPluginSummaryVO) *string {
	if plugin == nil {
		return nil
	}
	return plugin.Id
}
