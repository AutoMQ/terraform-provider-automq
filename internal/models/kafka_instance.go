package models

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Instance states that represent the lifecycle of a Kafka instance
const (
	StateCreating = "Creating"
	StateRunning  = "Running"
	StateChanging = "Changing"
	StateDeleting = "Deleting"
	StateNotFound = "NotFound"
	StateError    = "Error"
	StateUnknown  = "Unknown"
)

type KafkaInstanceResourceModel struct {
	EnvironmentID  types.String       `tfsdk:"environment_id"`
	InstanceID     types.String       `tfsdk:"id"`
	Name           types.String       `tfsdk:"name"`
	Description    types.String       `tfsdk:"description"`
	Version        types.String       `tfsdk:"version"`
	ComputeSpecs   *ComputeSpecsModel `tfsdk:"compute_specs"`
	Features       *FeaturesModel     `tfsdk:"features"`
	Tags           types.Map          `tfsdk:"tags"`
	Endpoints      types.List         `tfsdk:"endpoints"`
	CreatedAt      timetypes.RFC3339  `tfsdk:"created_at"`
	LastUpdated    timetypes.RFC3339  `tfsdk:"last_updated"`
	InstanceStatus types.String       `tfsdk:"status"`
	Timeouts       timeouts.Value     `tfsdk:"timeouts"`
}

type KafkaInstanceModel struct {
	EnvironmentID  types.String          `tfsdk:"environment_id"`
	InstanceID     types.String          `tfsdk:"id"`
	Name           types.String          `tfsdk:"name"`
	Description    types.String          `tfsdk:"description"`
	Version        types.String          `tfsdk:"version"`
	ComputeSpecs   *ComputeSpecsModel    `tfsdk:"compute_specs"`
	Features       *FeaturesSummaryModel `tfsdk:"features"`
	Tags           types.Map             `tfsdk:"tags"`
	Endpoints      types.List            `tfsdk:"endpoints"`
	CreatedAt      timetypes.RFC3339     `tfsdk:"created_at"`
	LastUpdated    timetypes.RFC3339     `tfsdk:"last_updated"`
	InstanceStatus types.String          `tfsdk:"status"`
}

type InstanceAccessInfo struct {
	DisplayName      types.String `tfsdk:"display_name"`
	NetworkType      types.String `tfsdk:"network_type"`
	Protocol         types.String `tfsdk:"protocol"`
	Mechanisms       types.String `tfsdk:"mechanisms"`
	BootstrapServers types.String `tfsdk:"bootstrap_servers"`
}

type NetworkModel struct {
	Zone    types.String `tfsdk:"zone"`
	Subnets types.List   `tfsdk:"subnets"`
}

type ComputeSpecsModel struct {
	ReservedAku           types.Int64  `tfsdk:"reserved_aku"`
	PricingMode           types.String `tfsdk:"pricing_mode"`
	ReservedNodeCount     types.Int64  `tfsdk:"reserved_node_count"`
	InstanceTypes         types.List   `tfsdk:"instance_types"`
	Networks              types.List   `tfsdk:"networks"`
	KubernetesNodeGroups  types.List   `tfsdk:"kubernetes_node_groups"`
	DeployType            types.String `tfsdk:"deploy_type"`
	DnsZone               types.String `tfsdk:"dns_zone"`
	KubernetesClusterID   types.String `tfsdk:"kubernetes_cluster_id"`
	KubernetesNamespace   types.String `tfsdk:"kubernetes_namespace"`
	KubernetesServiceAcct types.String `tfsdk:"kubernetes_service_account"`
	ScheduleSpec          types.String `tfsdk:"schedule_spec"`
	InstanceRole          types.String `tfsdk:"instance_role"`
	DataBuckets           types.List   `tfsdk:"data_buckets"`
	SecurityGroups        types.List   `tfsdk:"security_groups"`
	FileSystemParam       types.Object `tfsdk:"file_system_param"`
}

type NodeGroupModel struct {
	ID types.String `tfsdk:"id"`
}

type DataBucketModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
}

type FileSystemParamModel struct {
	FileSystemType               types.String `tfsdk:"file_system_type"`
	ThroughputMibpsPerFileSystem types.Int64  `tfsdk:"throughput_mibps_per_file_system"`
	FileSystemCount              types.Int64  `tfsdk:"file_system_count"`
	SecurityGroups               types.List   `tfsdk:"security_groups"`
}

var NetworkObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"zone":    types.StringType,
		"subnets": types.ListType{ElemType: types.StringType},
	},
}

var NodeGroupObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id": types.StringType,
	},
}

var DataBucketObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"bucket_name": types.StringType,
	},
}

var FileSystemParamObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"file_system_type":                 types.StringType,
		"throughput_mibps_per_file_system": types.Int64Type,
		"file_system_count":                types.Int64Type,
		"security_groups":                  types.ListType{ElemType: types.StringType},
	},
}

var SecurityObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"authentication_methods":          types.SetType{ElemType: types.StringType},
		"transit_encryption_modes":        types.SetType{ElemType: types.StringType},
		"certificate_authority":           types.StringType,
		"certificate_chain":               types.StringType,
		"private_key":                     types.StringType,
		"data_encryption_mode":            types.StringType,
		"tls_hostname_validation_enabled": types.BoolType,
	},
}

var SecuritySummaryObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"authentication_methods":          types.SetType{ElemType: types.StringType},
		"transit_encryption_modes":        types.SetType{ElemType: types.StringType},
		"data_encryption_mode":            types.StringType,
		"tls_hostname_validation_enabled": types.BoolType,
	},
}

var PrometheusExporterObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"auth_type":      types.StringType,
		"endpoint":       types.StringType,
		"prometheus_arn": types.StringType,
		"username":       types.StringType,
		"password":       types.StringType,
		"token":          types.StringType,
		"labels":         types.MapType{ElemType: types.StringType},
	},
}

var MetricsExporterObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"prometheus": PrometheusExporterObjectType,
	},
}

var TableTopicObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"warehouse":          types.StringType,
		"catalog_type":       types.StringType,
		"metastore_uri":      types.StringType,
		"hive_auth_mode":     types.StringType,
		"kerberos_principal": types.StringType,
		"user_principal":     types.StringType,
		"keytab_file":        types.StringType,
		"krb5conf_file":      types.StringType,
	},
}

func NetworkListToModels(ctx context.Context, list types.List) ([]NetworkModel, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var networks []NetworkModel
	diags := list.ElementsAs(ctx, &networks, false)
	return networks, diags
}

func NetworkModelsToList(ctx context.Context, networks []NetworkModel) (types.List, diag.Diagnostics) {
	if networks == nil {
		return types.ListNull(NetworkObjectType), nil
	}
	return types.ListValueFrom(ctx, NetworkObjectType, networks)
}

func NodeGroupListToModels(ctx context.Context, list types.List) ([]NodeGroupModel, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var groups []NodeGroupModel
	diags := list.ElementsAs(ctx, &groups, false)
	return groups, diags
}

func NodeGroupModelsToList(ctx context.Context, groups []NodeGroupModel) (types.List, diag.Diagnostics) {
	if groups == nil {
		return types.ListNull(NodeGroupObjectType), nil
	}
	return types.ListValueFrom(ctx, NodeGroupObjectType, groups)
}

func DataBucketListToModels(ctx context.Context, list types.List) ([]DataBucketModel, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var buckets []DataBucketModel
	diags := list.ElementsAs(ctx, &buckets, false)
	return buckets, diags
}

func DataBucketModelsToList(ctx context.Context, buckets []DataBucketModel) (types.List, diag.Diagnostics) {
	if len(buckets) == 0 {
		return types.ListNull(DataBucketObjectType), nil
	}
	listValue, diags := types.ListValueFrom(ctx, DataBucketObjectType, buckets)
	return listValue, diags
}

func FileSystemParamObjectToModel(ctx context.Context, object types.Object) (*FileSystemParamModel, diag.Diagnostics) {
	if object.IsNull() || object.IsUnknown() {
		return nil, nil
	}
	var model FileSystemParamModel
	diags := object.As(ctx, &model, basetypes.ObjectAsOptions{})
	return &model, diags
}

func FileSystemParamModelToObject(ctx context.Context, model *FileSystemParamModel) (types.Object, diag.Diagnostics) {
	if model == nil {
		return types.ObjectNull(FileSystemParamObjectType.AttrTypes), nil
	}
	normalized := *model
	if normalized.FileSystemType.IsNull() {
		normalized.FileSystemType = types.StringNull()
	}
	if normalized.ThroughputMibpsPerFileSystem.IsNull() {
		normalized.ThroughputMibpsPerFileSystem = types.Int64Null()
	}
	if normalized.FileSystemCount.IsNull() {
		normalized.FileSystemCount = types.Int64Null()
	}
	if normalized.SecurityGroups.IsNull() {
		normalized.SecurityGroups = types.ListNull(types.StringType)
	}
	return types.ObjectValueFrom(ctx, FileSystemParamObjectType.AttrTypes, normalized)
}

func SecurityObjectToModel(ctx context.Context, object types.Object) (*SecurityModel, diag.Diagnostics) {
	if object.IsNull() || object.IsUnknown() {
		return nil, nil
	}
	var model SecurityModel
	diags := object.As(ctx, &model, basetypes.ObjectAsOptions{})
	return &model, diags
}

func SecurityModelToObject(ctx context.Context, model *SecurityModel) (types.Object, diag.Diagnostics) {
	if model == nil {
		return types.ObjectNull(SecurityObjectType.AttrTypes), nil
	}
	normalized := *model
	if normalized.AuthenticationMethods.IsNull() {
		normalized.AuthenticationMethods = types.SetNull(types.StringType)
	}
	if normalized.TransitEncryptionModes.IsNull() {
		normalized.TransitEncryptionModes = types.SetNull(types.StringType)
	}
	if normalized.CertificateAuthority.IsNull() {
		normalized.CertificateAuthority = types.StringNull()
	}
	if normalized.CertificateChain.IsNull() {
		normalized.CertificateChain = types.StringNull()
	}
	if normalized.PrivateKey.IsNull() {
		normalized.PrivateKey = types.StringNull()
	}
	if normalized.DataEncryptionMode.IsNull() {
		normalized.DataEncryptionMode = types.StringNull()
	}
	if normalized.TlsHostnameValidationEnabled.IsNull() {
		normalized.TlsHostnameValidationEnabled = types.BoolNull()
	}
	return types.ObjectValueFrom(ctx, SecurityObjectType.AttrTypes, normalized)
}

func SecuritySummaryModelToObject(ctx context.Context, model *SecuritySummaryModel) (types.Object, diag.Diagnostics) {
	if model == nil {
		return types.ObjectNull(SecuritySummaryObjectType.AttrTypes), nil
	}
	normalized := *model
	if normalized.AuthenticationMethods.IsNull() {
		normalized.AuthenticationMethods = types.SetNull(types.StringType)
	}
	if normalized.TransitEncryptionModes.IsNull() {
		normalized.TransitEncryptionModes = types.SetNull(types.StringType)
	}
	if normalized.DataEncryptionMode.IsNull() {
		normalized.DataEncryptionMode = types.StringNull()
	}
	if normalized.TlsHostnameValidationEnabled.IsNull() {
		normalized.TlsHostnameValidationEnabled = types.BoolNull()
	}
	return types.ObjectValueFrom(ctx, SecuritySummaryObjectType.AttrTypes, normalized)
}

func MetricsExporterObjectToModel(ctx context.Context, object types.Object) (*MetricsExporterModel, diag.Diagnostics) {
	if object.IsNull() || object.IsUnknown() {
		return nil, nil
	}
	var model MetricsExporterModel
	diags := object.As(ctx, &model, basetypes.ObjectAsOptions{})
	return &model, diags
}

func MetricsExporterModelToObject(ctx context.Context, model *MetricsExporterModel) (types.Object, diag.Diagnostics) {
	if model == nil {
		return types.ObjectNull(MetricsExporterObjectType.AttrTypes), nil
	}
	normalized := *model
	if normalized.Prometheus != nil {
		prom := *normalized.Prometheus
		if prom.AuthType.IsNull() {
			prom.AuthType = types.StringNull()
		}
		if prom.EndPoint.IsNull() {
			prom.EndPoint = types.StringNull()
		}
		if prom.PrometheusArn.IsNull() {
			prom.PrometheusArn = types.StringNull()
		}
		if prom.Username.IsNull() {
			prom.Username = types.StringNull()
		}
		if prom.Password.IsNull() {
			prom.Password = types.StringNull()
		}
		if prom.Token.IsNull() {
			prom.Token = types.StringNull()
		}
		if prom.Labels.IsNull() {
			prom.Labels = types.MapNull(types.StringType)
		}
		normalized.Prometheus = &prom
	}
	return types.ObjectValueFrom(ctx, MetricsExporterObjectType.AttrTypes, normalized)
}

func TableTopicObjectToModel(ctx context.Context, object types.Object) (*TableTopicModel, diag.Diagnostics) {
	if object.IsNull() || object.IsUnknown() {
		return nil, nil
	}
	var model TableTopicModel
	diags := object.As(ctx, &model, basetypes.ObjectAsOptions{})
	return &model, diags
}

func TableTopicModelToObject(ctx context.Context, model *TableTopicModel) (types.Object, diag.Diagnostics) {
	if model == nil {
		return types.ObjectNull(TableTopicObjectType.AttrTypes), nil
	}
	normalized := *model
	if normalized.Warehouse.IsNull() {
		normalized.Warehouse = types.StringNull()
	}
	if normalized.CatalogType.IsNull() {
		normalized.CatalogType = types.StringNull()
	}
	if normalized.MetastoreURI.IsNull() {
		normalized.MetastoreURI = types.StringNull()
	}
	if normalized.HiveAuthMode.IsNull() {
		normalized.HiveAuthMode = types.StringNull()
	}
	if normalized.KerberosPrincipal.IsNull() {
		normalized.KerberosPrincipal = types.StringNull()
	}
	if normalized.UserPrincipal.IsNull() {
		normalized.UserPrincipal = types.StringNull()
	}
	if normalized.KeytabFile.IsNull() {
		normalized.KeytabFile = types.StringNull()
	}
	if normalized.Krb5ConfFile.IsNull() {
		normalized.Krb5ConfFile = types.StringNull()
	}
	return types.ObjectValueFrom(ctx, TableTopicObjectType.AttrTypes, normalized)
}

func coalesceStringAttr(apiValue *string, previous *types.String) types.String {
	if apiValue != nil {
		return types.StringValue(*apiValue)
	}
	if previous != nil && !previous.IsNull() && !previous.IsUnknown() {
		return *previous
	}
	return types.StringNull()
}

func coalesceBoolAttr(apiValue *bool, previous *types.Bool) types.Bool {
	if apiValue != nil {
		return types.BoolValue(*apiValue)
	}
	if previous != nil && !previous.IsNull() && !previous.IsUnknown() {
		return *previous
	}
	return types.BoolNull()
}

type FeaturesModel struct {
	WalMode               types.String `tfsdk:"wal_mode"`
	InstanceConfigs       types.Map    `tfsdk:"instance_configs"`
	Security              types.Object `tfsdk:"security"`
	MetricsExporter       types.Object `tfsdk:"metrics_exporter"`
	TableTopic            types.Object `tfsdk:"table_topic"`
	SchemaRegistryEnabled types.Bool   `tfsdk:"schema_registry_enabled"`
}

type FeaturesSummaryModel struct {
	WalMode               types.String `tfsdk:"wal_mode"`
	InstanceConfigs       types.Map    `tfsdk:"instance_configs"`
	Security              types.Object `tfsdk:"security"`
	MetricsExporter       types.Object `tfsdk:"metrics_exporter"`
	TableTopic            types.Object `tfsdk:"table_topic"`
	SchemaRegistryEnabled types.Bool   `tfsdk:"schema_registry_enabled"`
}

type SecurityModel struct {
	AuthenticationMethods        types.Set    `tfsdk:"authentication_methods"`
	TransitEncryptionModes       types.Set    `tfsdk:"transit_encryption_modes"`
	CertificateAuthority         types.String `tfsdk:"certificate_authority"`
	CertificateChain             types.String `tfsdk:"certificate_chain"`
	PrivateKey                   types.String `tfsdk:"private_key"`
	DataEncryptionMode           types.String `tfsdk:"data_encryption_mode"`
	TlsHostnameValidationEnabled types.Bool   `tfsdk:"tls_hostname_validation_enabled"`
}

type SecuritySummaryModel struct {
	AuthenticationMethods        types.Set    `tfsdk:"authentication_methods"`
	TransitEncryptionModes       types.Set    `tfsdk:"transit_encryption_modes"`
	DataEncryptionMode           types.String `tfsdk:"data_encryption_mode"`
	TlsHostnameValidationEnabled types.Bool   `tfsdk:"tls_hostname_validation_enabled"`
}

type MetricsExporterModel struct {
	Prometheus *PrometheusExporterModel `tfsdk:"prometheus"`
}

type PrometheusExporterModel struct {
	AuthType      types.String `tfsdk:"auth_type"`
	EndPoint      types.String `tfsdk:"endpoint"`
	PrometheusArn types.String `tfsdk:"prometheus_arn"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	Token         types.String `tfsdk:"token"`
	Labels        types.Map    `tfsdk:"labels"`
}

// MetricsExporterHasConfig reports whether any nested exporter attributes are
// explicitly configured (non-null and known).
func MetricsExporterHasConfig(model *MetricsExporterModel) bool {
	if model == nil {
		return false
	}
	return PrometheusExporterHasConfig(model.Prometheus)
}

// PrometheusExporterHasConfig inspects the nested Prometheus block for any user
// supplied values. It mirrors the provider-side logic so both expansion and
// flattening share the same definition of an \"empty\" block.
func PrometheusExporterHasConfig(model *PrometheusExporterModel) bool {
	if model == nil {
		return false
	}
	return isKnownStringSet(model.AuthType) ||
		isKnownStringSet(model.EndPoint) ||
		isKnownStringSet(model.PrometheusArn) ||
		isKnownStringSet(model.Username) ||
		isKnownStringSet(model.Password) ||
		isKnownStringSet(model.Token) ||
		(!model.Labels.IsNull() && !model.Labels.IsUnknown() && len(model.Labels.Elements()) > 0)
}

func isKnownStringSet(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func BuildMetricsExporterParam(model *MetricsExporterModel) (*client.InstanceMetricsExporterParam, bool) {
	if model == nil {
		return nil, false
	}
	exporter := client.InstanceMetricsExporterParam{}
	hasConfig := false
	if model.Prometheus != nil {
		prom, ok := BuildPrometheusExporterParam(model.Prometheus)
		if ok {
			exporter.Prometheus = prom
			hasConfig = true
		}
	}
	if !hasConfig {
		return nil, false
	}
	return &exporter, true
}

func BuildPrometheusExporterParam(model *PrometheusExporterModel) (*client.InstancePrometheusExporterParam, bool) {
	if model == nil {
		return nil, false
	}
	if !PrometheusExporterHasConfig(model) {
		return nil, false
	}
	prom := &client.InstancePrometheusExporterParam{}
	enabled := true
	prom.Enabled = &enabled
	if !model.AuthType.IsNull() && !model.AuthType.IsUnknown() {
		auth := model.AuthType.ValueString()
		prom.AuthType = &auth
	}
	if !model.EndPoint.IsNull() && !model.EndPoint.IsUnknown() {
		endpoint := model.EndPoint.ValueString()
		prom.EndPoint = &endpoint
	}
	if !model.PrometheusArn.IsNull() && !model.PrometheusArn.IsUnknown() {
		arn := model.PrometheusArn.ValueString()
		prom.PrometheusArn = &arn
	}
	if !model.Username.IsNull() && !model.Username.IsUnknown() {
		username := model.Username.ValueString()
		prom.Username = &username
	}
	if !model.Password.IsNull() && !model.Password.IsUnknown() {
		password := model.Password.ValueString()
		prom.Password = &password
	}
	if !model.Token.IsNull() && !model.Token.IsUnknown() {
		token := model.Token.ValueString()
		prom.Token = &token
	}
	if !model.Labels.IsNull() && !model.Labels.IsUnknown() && len(model.Labels.Elements()) > 0 {
		labels := ExpandStringValueMap(model.Labels)
		if len(labels) > 0 {
			promLabels := make([]client.MetricsLabelParam, len(labels))
			for i, label := range labels {
				name := ""
				if label.Key != nil {
					name = *label.Key
				}
				value := ""
				if label.Value != nil {
					value = *label.Value
				}
				promLabels[i] = client.MetricsLabelParam{Name: name, Value: value}
			}
			prom.Labels = promLabels
		}
	}
	return prom, true
}

type TableTopicModel struct {
	Warehouse         types.String `tfsdk:"warehouse"`
	CatalogType       types.String `tfsdk:"catalog_type"`
	MetastoreURI      types.String `tfsdk:"metastore_uri"`
	HiveAuthMode      types.String `tfsdk:"hive_auth_mode"`
	KerberosPrincipal types.String `tfsdk:"kerberos_principal"`
	UserPrincipal     types.String `tfsdk:"user_principal"`
	KeytabFile        types.String `tfsdk:"keytab_file"`
	Krb5ConfFile      types.String `tfsdk:"krb5conf_file"`
}

func BuildTableTopicParam(model *TableTopicModel) *client.TableTopicParam {
	if model == nil {
		return nil
	}
	enabled := true
	topic := &client.TableTopicParam{
		Enabled:     &enabled,
		Warehouse:   model.Warehouse.ValueString(),
		CatalogType: model.CatalogType.ValueString(),
	}
	if !model.MetastoreURI.IsNull() && !model.MetastoreURI.IsUnknown() {
		value := model.MetastoreURI.ValueString()
		topic.MetastoreUri = &value
	}
	if !model.HiveAuthMode.IsNull() && !model.HiveAuthMode.IsUnknown() {
		value := model.HiveAuthMode.ValueString()
		topic.HiveAuthMode = &value
	}
	if !model.KerberosPrincipal.IsNull() && !model.KerberosPrincipal.IsUnknown() {
		value := model.KerberosPrincipal.ValueString()
		topic.KerberosPrincipal = &value
	}
	if !model.UserPrincipal.IsNull() && !model.UserPrincipal.IsUnknown() {
		value := model.UserPrincipal.ValueString()
		topic.UserPrincipal = &value
	}
	if !model.KeytabFile.IsNull() && !model.KeytabFile.IsUnknown() {
		value := model.KeytabFile.ValueString()
		topic.KeytabFile = &value
	}
	if !model.Krb5ConfFile.IsNull() && !model.Krb5ConfFile.IsUnknown() {
		value := model.Krb5ConfFile.ValueString()
		topic.Krb5ConfFile = &value
	}
	return topic
}

// ExpandKafkaInstanceResource converts a KafkaInstanceResourceModel to a client.InstanceCreateParam.
// It handles the conversion of all nested structures and validates required fields.
func ExpandKafkaInstanceResource(ctx context.Context, instance KafkaInstanceResourceModel, request *client.InstanceCreateParam) error {
	if request == nil {
		return fmt.Errorf("request parameter cannot be nil")
	}

	// Basic fields
	request.Name = instance.Name.ValueString()
	request.Description = instance.Description.ValueString()
	request.Version = instance.Version.ValueString()

	// Validate required fields
	if request.Name == "" {
		return fmt.Errorf("instance name is required")
	}

	if request.Version == "" {
		return fmt.Errorf("instance version is required")
	}

	// Compute Specs
	if instance.ComputeSpecs != nil {
		// Reserved AKU
		request.Spec = client.SpecificationParam{}
		if !instance.ComputeSpecs.ReservedAku.IsNull() && !instance.ComputeSpecs.ReservedAku.IsUnknown() {
			request.Spec.ReservedAku = int32(instance.ComputeSpecs.ReservedAku.ValueInt64())
		}

		// Pricing Mode
		if !instance.ComputeSpecs.PricingMode.IsNull() && !instance.ComputeSpecs.PricingMode.IsUnknown() {
			pricingMode := instance.ComputeSpecs.PricingMode.ValueString()
			request.Spec.PricingMode = &pricingMode
		}

		// Reserved Node Count
		if !instance.ComputeSpecs.ReservedNodeCount.IsNull() && !instance.ComputeSpecs.ReservedNodeCount.IsUnknown() {
			reservedNodeCount := int32(instance.ComputeSpecs.ReservedNodeCount.ValueInt64())
			request.Spec.ReservedNodeCount = &reservedNodeCount
		}

		// Instance Types (for UsageBased pricing mode)
		if !instance.ComputeSpecs.InstanceTypes.IsNull() && !instance.ComputeSpecs.InstanceTypes.IsUnknown() {
			var instanceTypes []string
			diags := instance.ComputeSpecs.InstanceTypes.ElementsAs(ctx, &instanceTypes, false)
			if diags.HasError() {
				return fmt.Errorf("failed to parse instance_types: %v", diags)
			}
			if len(instanceTypes) > 0 {
				if request.Spec.NodeConfig == nil {
					request.Spec.NodeConfig = &client.NodeConfigParam{}
				}
				request.Spec.NodeConfig.InstanceTypes = instanceTypes
			}
		}

		if !instance.ComputeSpecs.DeployType.IsNull() && !instance.ComputeSpecs.DeployType.IsUnknown() {
			deployType := instance.ComputeSpecs.DeployType.ValueString()
			request.Spec.DeployType = &deployType
		}
		if !instance.ComputeSpecs.DnsZone.IsNull() && !instance.ComputeSpecs.DnsZone.IsUnknown() {
			dns := instance.ComputeSpecs.DnsZone.ValueString()
			request.Spec.DnsZone = &dns
		}
		if !instance.ComputeSpecs.KubernetesClusterID.IsNull() && !instance.ComputeSpecs.KubernetesClusterID.IsUnknown() {
			clusterID := instance.ComputeSpecs.KubernetesClusterID.ValueString()
			request.Spec.KubernetesClusterId = &clusterID
		}
		if !instance.ComputeSpecs.KubernetesNamespace.IsNull() && !instance.ComputeSpecs.KubernetesNamespace.IsUnknown() {
			namespace := instance.ComputeSpecs.KubernetesNamespace.ValueString()
			request.Spec.KubernetesNamespace = &namespace
		}
		if !instance.ComputeSpecs.KubernetesServiceAcct.IsNull() && !instance.ComputeSpecs.KubernetesServiceAcct.IsUnknown() {
			serviceAccount := instance.ComputeSpecs.KubernetesServiceAcct.ValueString()
			request.Spec.KubernetesServiceAccount = &serviceAccount
		}
		if !instance.ComputeSpecs.ScheduleSpec.IsNull() && !instance.ComputeSpecs.ScheduleSpec.IsUnknown() {
			scheduleSpec := instance.ComputeSpecs.ScheduleSpec.ValueString()
			request.Spec.ScheduleSpec = &scheduleSpec
		}
		if !instance.ComputeSpecs.InstanceRole.IsNull() && !instance.ComputeSpecs.InstanceRole.IsUnknown() {
			role := instance.ComputeSpecs.InstanceRole.ValueString()
			request.Spec.InstanceRole = &role
		}
		// Kubernetes Node Groups
		nodeGroupModels, nodeGroupDiags := NodeGroupListToModels(ctx, instance.ComputeSpecs.KubernetesNodeGroups)
		if nodeGroupDiags.HasError() {
			return fmt.Errorf("failed to parse compute_specs.kubernetes_node_groups: %v", nodeGroupDiags.Errors())
		}
		if len(nodeGroupModels) > 0 {
			nodeGroups := make([]client.KubernetesNodeGroupParam, 0, len(nodeGroupModels))
			for _, ng := range nodeGroupModels {
				id := ng.ID.ValueString()
				nodeGroups = append(nodeGroups, client.KubernetesNodeGroupParam{
					Id: &id,
				})
			}
			request.Spec.KubernetesNodeGroups = nodeGroups
		}
		networkModels, networkDiags := NetworkListToModels(ctx, instance.ComputeSpecs.Networks)
		if networkDiags.HasError() {
			return fmt.Errorf("failed to parse compute_specs.networks: %v", networkDiags.Errors())
		}
		if len(networkModels) > 0 {
			networks := make([]client.InstanceNetworkParam, 0, len(networkModels))
			for _, network := range networkModels {
				var subnet *string
				if !network.Subnets.IsNull() {
					subnets := ExpandStringValueList(network.Subnets)
					if len(subnets) > 0 {
						subnet = &subnets[0]
					}
				}
				networks = append(networks, client.InstanceNetworkParam{
					Zone:   network.Zone.ValueString(),
					Subnet: subnet,
				})
			}
			request.Spec.Networks = networks
		}
		if !instance.ComputeSpecs.DataBuckets.IsNull() && !instance.ComputeSpecs.DataBuckets.IsUnknown() {
			dataBucketModels, dataBucketDiags := DataBucketListToModels(ctx, instance.ComputeSpecs.DataBuckets)
			if dataBucketDiags.HasError() {
				return fmt.Errorf("failed to parse compute_specs.data_buckets: %v", dataBucketDiags.Errors())
			}
			dataBuckets := make([]client.BucketProfileParam, 0, len(dataBucketModels))
			for _, bucket := range dataBucketModels {
				if bucket.BucketName.IsNull() || bucket.BucketName.IsUnknown() {
					return fmt.Errorf("compute_specs.data_buckets.bucket_name is required")
				}
				profile := client.BucketProfileParam{
					BucketName: bucket.BucketName.ValueString(),
				}
				dataBuckets = append(dataBuckets, profile)
			}
			request.Spec.DataBuckets = dataBuckets
		}

		// Security Groups for compute specs
		if !instance.ComputeSpecs.SecurityGroups.IsNull() &&
			!instance.ComputeSpecs.SecurityGroups.IsUnknown() {
			var securityGroups []string
			diags := instance.ComputeSpecs.SecurityGroups.ElementsAs(ctx, &securityGroups, false)
			if !diags.HasError() && len(securityGroups) > 0 {
				request.Spec.SecurityGroups = securityGroups
			}
		}

		// File System Parameters for FSWAL
		fileSystemModel, fileSystemDiags := FileSystemParamObjectToModel(ctx, instance.ComputeSpecs.FileSystemParam)
		if fileSystemDiags.HasError() {
			return fmt.Errorf("failed to parse compute_specs.file_system_param: %v", fileSystemDiags.Errors())
		}
		if fileSystemModel != nil {
			fileSystemParam := &client.FileSystemParam{
				ThroughputMiBpsPerFileSystem: int32(fileSystemModel.ThroughputMibpsPerFileSystem.ValueInt64()),
				FileSystemCount:              int32(fileSystemModel.FileSystemCount.ValueInt64()),
			}

			// File system type
			if !fileSystemModel.FileSystemType.IsNull() &&
				!fileSystemModel.FileSystemType.IsUnknown() {
				fsType := fileSystemModel.FileSystemType.ValueString()
				fileSystemParam.FileSystemType = &fsType
			}

			// Security groups protection logic: only include if not empty
			if !fileSystemModel.SecurityGroups.IsNull() &&
				!fileSystemModel.SecurityGroups.IsUnknown() {
				var securityGroups []string
				diags := fileSystemModel.SecurityGroups.ElementsAs(ctx, &securityGroups, false)
				if !diags.HasError() && len(securityGroups) > 0 {
					fileSystemParam.SecurityGroups = securityGroups
				}
			}

			request.Spec.FileSystem = fileSystemParam
		}

	}

	// Tags - expand map to TagParam slice
	if !instance.Tags.IsNull() && !instance.Tags.IsUnknown() {
		tagsMap := make(map[string]string)
		diags := instance.Tags.ElementsAs(ctx, &tagsMap, false)
		if diags.HasError() {
			return fmt.Errorf("failed to parse tags: %v", diags)
		}
		tags := make([]client.TagParam, 0, len(tagsMap))
		for name, value := range tagsMap {
			tags = append(tags, client.TagParam{
				Name:  name,
				Value: value,
			})
		}
		request.Tags = tags
	}

	// Features
	if instance.Features != nil {
		// WAL
		walMode := instance.Features.WalMode.ValueString()
		request.Features = &client.InstanceFeatureParam{
			WalMode: &walMode,
		}

		// Security
		securityModel, securityDiags := SecurityObjectToModel(ctx, instance.Features.Security)
		if securityDiags.HasError() {
			return fmt.Errorf("failed to parse features.security: %v", securityDiags.Errors())
		}
		if securityModel != nil {
			if !securityModel.DataEncryptionMode.IsNull() {
				dataEncryptionMode := securityModel.DataEncryptionMode.ValueString()
				request.Features.Security = &client.InstanceSecurityParam{
					DataEncryptionMode: &dataEncryptionMode,
				}
			}

			if !securityModel.AuthenticationMethods.IsNull() {
				var authMethods []string
				diags := securityModel.AuthenticationMethods.ElementsAs(ctx, &authMethods, false)
				if diags.HasError() {
					return fmt.Errorf("failed to parse authentication methods: %v", diags)
				}
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.AuthenticationMethods = authMethods
			}

			if !securityModel.TransitEncryptionModes.IsNull() {
				var encryptionModes []string
				diags := securityModel.TransitEncryptionModes.ElementsAs(ctx, &encryptionModes, false)
				if diags.HasError() {
					return fmt.Errorf("failed to parse transit encryption modes: %v", diags)
				}
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.TransitEncryptionModes = encryptionModes
			}

			if !securityModel.CertificateAuthority.IsNull() {
				certAuth := securityModel.CertificateAuthority.ValueString()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.CertificateAuthority = &certAuth
			}

			if !securityModel.CertificateChain.IsNull() {
				certChain := securityModel.CertificateChain.ValueString()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.CertificateChain = &certChain
			}

			if !securityModel.PrivateKey.IsNull() {
				privateKey := securityModel.PrivateKey.ValueString()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.PrivateKey = &privateKey
			}

			if !securityModel.TlsHostnameValidationEnabled.IsNull() && !securityModel.TlsHostnameValidationEnabled.IsUnknown() {
				enabled := securityModel.TlsHostnameValidationEnabled.ValueBool()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.TlsHostnameValidationEnabled = &enabled
			}
		}

		// Metrics exporter
		metricsModel, metricsDiags := MetricsExporterObjectToModel(ctx, instance.Features.MetricsExporter)
		if metricsDiags.HasError() {
			return fmt.Errorf("failed to parse features.metrics_exporter: %v", metricsDiags.Errors())
		}
		if metricsModel != nil {
			if exporter, hasConfig := BuildMetricsExporterParam(metricsModel); hasConfig {
				request.Features.MetricsExporter = exporter
			}
		}

		// Table topic
		topicModel, topicDiags := TableTopicObjectToModel(ctx, instance.Features.TableTopic)
		if topicDiags.HasError() {
			return fmt.Errorf("failed to parse features.table_topic: %v", topicDiags.Errors())
		}
		if topicModel != nil {
			if topicModel.Warehouse.IsNull() || topicModel.Warehouse.IsUnknown() {
				return fmt.Errorf("features.table_topic.warehouse is required when table_topic is set")
			}
			if topicModel.CatalogType.IsNull() || topicModel.CatalogType.IsUnknown() {
				return fmt.Errorf("features.table_topic.catalog_type is required when table_topic is set")
			}
			request.Features.TableTopic = BuildTableTopicParam(topicModel)
		}

		// Schema Registry
		if !instance.Features.SchemaRegistryEnabled.IsNull() && !instance.Features.SchemaRegistryEnabled.IsUnknown() {
			enabled := instance.Features.SchemaRegistryEnabled.ValueBool()
			request.Features.SchemaRegistryEnabled = &enabled
		}

		// Instance Configs
		if !instance.Features.InstanceConfigs.IsNull() {
			instanceConfigs := ExpandStringValueMap(instance.Features.InstanceConfigs)
			request.Features.InstanceConfigs = instanceConfigs
		}
	}
	return nil
}

// ConvertKafkaInstanceModel copies data from a KafkaInstanceResourceModel to a KafkaInstanceModel.
// This function is used when converting between different model representations.
func ConvertKafkaInstanceModel(resource *KafkaInstanceResourceModel, model *KafkaInstanceModel) error {
	if resource == nil || model == nil {
		return fmt.Errorf("both resource and model parameters must not be nil")
	}

	model.EnvironmentID = resource.EnvironmentID
	model.InstanceID = resource.InstanceID
	model.Name = resource.Name
	model.Description = resource.Description
	model.Version = resource.Version
	model.ComputeSpecs = resource.ComputeSpecs
	if resource.Features != nil {
		features := &FeaturesSummaryModel{
			WalMode:               resource.Features.WalMode,
			InstanceConfigs:       resource.Features.InstanceConfigs,
			MetricsExporter:       resource.Features.MetricsExporter,
			TableTopic:            resource.Features.TableTopic,
			SchemaRegistryEnabled: resource.Features.SchemaRegistryEnabled,
			Security:              types.ObjectNull(SecuritySummaryObjectType.AttrTypes),
		}
		security, securityDiags := SecurityObjectToModel(context.Background(), resource.Features.Security)
		if securityDiags.HasError() {
			return fmt.Errorf("failed to parse features.security: %v", securityDiags.Errors())
		}
		if security != nil {
			summary := &SecuritySummaryModel{
				AuthenticationMethods:        security.AuthenticationMethods,
				TransitEncryptionModes:       security.TransitEncryptionModes,
				DataEncryptionMode:           security.DataEncryptionMode,
				TlsHostnameValidationEnabled: security.TlsHostnameValidationEnabled,
			}
			securityObject, objectDiags := SecuritySummaryModelToObject(context.Background(), summary)
			if objectDiags.HasError() {
				return fmt.Errorf("failed to build features.security: %v", objectDiags.Errors())
			}
			features.Security = securityObject
		}
		model.Features = features
	} else {
		model.Features = nil
	}
	model.Endpoints = resource.Endpoints
	// Ensure Tags has proper type information
	if resource.Tags.IsNull() {
		model.Tags = types.MapNull(types.StringType)
	} else if resource.Tags.IsUnknown() {
		model.Tags = types.MapUnknown(types.StringType)
	} else {
		model.Tags = resource.Tags
	}
	model.CreatedAt = resource.CreatedAt
	model.LastUpdated = resource.LastUpdated
	model.InstanceStatus = resource.InstanceStatus
	return nil
}

// FlattenKafkaInstanceBasicModel converts a client.InstanceSummaryVO to a KafkaInstanceResourceModel.
// It handles the basic fields of the instance model.
func FlattenKafkaInstanceBasicModel(instance *client.InstanceSummaryVO, resource *KafkaInstanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if instance == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Instance",
			"Cannot flatten nil instance",
		)}
	}

	if resource == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Resource",
			"Cannot flatten instance into nil resource",
		)}
	}

	// Basic fields
	if instance.InstanceId != nil {
		resource.InstanceID = types.StringValue(*instance.InstanceId)
	}
	if instance.Name != nil {
		resource.Name = types.StringValue(*instance.Name)
	}
	if instance.Description != nil {
		resource.Description = types.StringValue(*instance.Description)
	}
	if instance.Version != nil {
		resource.Version = types.StringValue(*instance.Version)
	}
	if instance.State != nil {
		resource.InstanceStatus = types.StringValue(*instance.State)
	}

	// Timestamps
	if instance.GmtCreate != nil {
		resource.CreatedAt = timetypes.NewRFC3339TimePointerValue(instance.GmtCreate)
	}
	if instance.GmtModified != nil {
		resource.LastUpdated = timetypes.NewRFC3339TimePointerValue(instance.GmtModified)
	}

	return diags
}

// FlattenKafkaInstanceModel converts a client.InstanceVO to a KafkaInstanceResourceModel.
// It handles all fields including nested structures.
func FlattenKafkaInstanceModel(ctx context.Context, instance *client.InstanceVO, resource *KafkaInstanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if instance == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Instance",
			"Cannot flatten nil instance",
		)}
	}

	if resource == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Resource",
			"Cannot flatten instance into nil resource",
		)}
	}

	// Basic fields
	if instance.InstanceId != nil {
		resource.InstanceID = types.StringValue(*instance.InstanceId)
	}
	if instance.Name != nil {
		resource.Name = types.StringValue(*instance.Name)
	}
	if instance.Description != nil {
		resource.Description = types.StringValue(*instance.Description)
	}
	if instance.Version != nil {
		resource.Version = types.StringValue(*instance.Version)
	}
	if instance.State != nil {
		resource.InstanceStatus = types.StringValue(*instance.State)
	}

	// Compute Specs
	if instance.Spec != nil {
		var previousSpecs *ComputeSpecsModel
		if resource.ComputeSpecs != nil {
			prevCopy := *resource.ComputeSpecs
			previousSpecs = &prevCopy
		}
		if resource.ComputeSpecs == nil {
			resource.ComputeSpecs = &ComputeSpecsModel{
				ReservedAku:           types.Int64Null(),
				PricingMode:           types.StringNull(),
				ReservedNodeCount:     types.Int64Null(),
				InstanceTypes:         types.ListNull(types.StringType),
				Networks:              types.ListNull(NetworkObjectType),
				KubernetesNodeGroups:  types.ListNull(NodeGroupObjectType),
				DataBuckets:           types.ListNull(DataBucketObjectType),
				SecurityGroups:        types.ListNull(types.StringType),
				DeployType:            types.StringNull(),
				DnsZone:               types.StringNull(),
				KubernetesClusterID:   types.StringNull(),
				KubernetesNamespace:   types.StringNull(),
				KubernetesServiceAcct: types.StringNull(),
				ScheduleSpec:          types.StringNull(),
				InstanceRole:          types.StringNull(),
				FileSystemParam:       types.ObjectNull(FileSystemParamObjectType.AttrTypes),
			}
		}
		// Reserved AKU
		if instance.Spec.ReservedAku != nil {
			resource.ComputeSpecs.ReservedAku = types.Int64Value(int64(*instance.Spec.ReservedAku))
		}
		// Pricing Mode
		var prevPricingMode *types.String
		if previousSpecs != nil {
			prevPricingMode = &previousSpecs.PricingMode
		}
		resource.ComputeSpecs.PricingMode = coalesceStringAttr(instance.Spec.PricingMode, prevPricingMode)
		// Reserved Node Count
		if instance.Spec.ReservedNodeCount != nil {
			resource.ComputeSpecs.ReservedNodeCount = types.Int64Value(int64(*instance.Spec.ReservedNodeCount))
		} else if previousSpecs != nil && !previousSpecs.ReservedNodeCount.IsNull() && !previousSpecs.ReservedNodeCount.IsUnknown() {
			resource.ComputeSpecs.ReservedNodeCount = previousSpecs.ReservedNodeCount
		} else {
			resource.ComputeSpecs.ReservedNodeCount = types.Int64Null()
		}
		var prevDeploy, prevDnsZone *types.String
		var prevClusterID, prevNamespace, prevServiceAccount, prevScheduleSpec, prevInstanceRole *types.String
		if previousSpecs != nil {
			prevDeploy = &previousSpecs.DeployType
			prevDnsZone = &previousSpecs.DnsZone
			prevClusterID = &previousSpecs.KubernetesClusterID
			prevNamespace = &previousSpecs.KubernetesNamespace
			prevServiceAccount = &previousSpecs.KubernetesServiceAcct
			prevScheduleSpec = &previousSpecs.ScheduleSpec
			prevInstanceRole = &previousSpecs.InstanceRole
		}
		resource.ComputeSpecs.DeployType = coalesceStringAttr(instance.Spec.DeployType, prevDeploy)
		resource.ComputeSpecs.DnsZone = coalesceStringAttr(instance.Spec.DnsZone, prevDnsZone)
		resource.ComputeSpecs.KubernetesClusterID = coalesceStringAttr(instance.Spec.KubernetesClusterId, prevClusterID)
		resource.ComputeSpecs.KubernetesNamespace = coalesceStringAttr(instance.Spec.KubernetesNamespace, prevNamespace)
		resource.ComputeSpecs.KubernetesServiceAcct = coalesceStringAttr(instance.Spec.KubernetesServiceAccount, prevServiceAccount)
		resource.ComputeSpecs.ScheduleSpec = coalesceStringAttr(instance.Spec.ScheduleSpec, prevScheduleSpec)
		resource.ComputeSpecs.InstanceRole = coalesceStringAttr(instance.Spec.InstanceRole, prevInstanceRole)
		// Instance Types (from NodeConfig)
		shouldRetainInstanceTypes := strings.EqualFold(resource.ComputeSpecs.DeployType.ValueString(), "K8S") ||
			(strings.EqualFold(resource.ComputeSpecs.PricingMode.ValueString(), "UsageBased") && strings.EqualFold(resource.ComputeSpecs.DeployType.ValueString(), "IAAS"))
		if shouldRetainInstanceTypes && instance.Spec.NodeConfig != nil && len(instance.Spec.NodeConfig.InstanceTypes) > 0 {
			instanceTypesList, itDiags := types.ListValueFrom(ctx, types.StringType, instance.Spec.NodeConfig.InstanceTypes)
			if !itDiags.HasError() {
				resource.ComputeSpecs.InstanceTypes = instanceTypesList
			}
		} else if shouldRetainInstanceTypes && previousSpecs != nil && !previousSpecs.InstanceTypes.IsNull() && !previousSpecs.InstanceTypes.IsUnknown() {
			resource.ComputeSpecs.InstanceTypes = previousSpecs.InstanceTypes
		} else {
			resource.ComputeSpecs.InstanceTypes = types.ListNull(types.StringType)
		}

		// Kubernetes Node Groups
		if instance.Spec.KubernetesNodeGroups != nil {
			nodeGroups := make([]NodeGroupModel, 0, len(instance.Spec.KubernetesNodeGroups))
			for _, ng := range instance.Spec.KubernetesNodeGroups {
				if ng.Id != nil {
					nodeGroups = append(nodeGroups, NodeGroupModel{
						ID: types.StringValue(*ng.Id),
					})
				}
			}
			nodeGroupsValue, nodeGroupDiags := NodeGroupModelsToList(ctx, nodeGroups)
			if nodeGroupDiags.HasError() {
				diags.Append(nodeGroupDiags...)
			} else {
				resource.ComputeSpecs.KubernetesNodeGroups = nodeGroupsValue
			}
		}
		if instance.Spec.Networks != nil {
			networks, networkDiags := flattenNetworks(instance.Spec.Networks)
			if networkDiags.HasError() {
				diags.Append(networkDiags...)
			} else {
				networksValue, valueDiags := NetworkModelsToList(ctx, networks)
				if valueDiags.HasError() {
					diags.Append(valueDiags...)
				} else {
					resource.ComputeSpecs.Networks = networksValue
				}
			}
		}

		var previousDataBuckets []DataBucketModel
		if previousSpecs != nil {
			prevBuckets, prevDiags := DataBucketListToModels(ctx, previousSpecs.DataBuckets)
			if prevDiags.HasError() {
				diags.Append(prevDiags...)
			}
			previousDataBuckets = prevBuckets
		}
		if instance.Spec.DataBuckets != nil {
			dataBuckets := make([]DataBucketModel, 0, len(instance.Spec.DataBuckets))
			for idx, bucket := range instance.Spec.DataBuckets {
				var base DataBucketModel
				if idx < len(previousDataBuckets) {
					base = previousDataBuckets[idx]
				} else {
					base = DataBucketModel{
						BucketName: types.StringNull(),
					}
				}
				if bucket.BucketName != nil {
					base.BucketName = types.StringValue(*bucket.BucketName)
				}
				dataBuckets = append(dataBuckets, base)
			}
			listValue, listDiags := DataBucketModelsToList(ctx, dataBuckets)
			if listDiags.HasError() {
				diags.Append(listDiags...)
			} else {
				resource.ComputeSpecs.DataBuckets = listValue
			}
		} else if previousSpecs != nil {
			resource.ComputeSpecs.DataBuckets = previousSpecs.DataBuckets
		} else {
			resource.ComputeSpecs.DataBuckets = types.ListNull(DataBucketObjectType)
		}

		// Security Groups for compute specs
		if len(instance.Spec.SecurityGroups) > 0 {
			securityGroupsList, sgDiags := types.ListValueFrom(ctx, types.StringType, instance.Spec.SecurityGroups)
			if !sgDiags.HasError() {
				resource.ComputeSpecs.SecurityGroups = securityGroupsList
			}
		} else if previousSpecs != nil && !previousSpecs.SecurityGroups.IsNull() && !previousSpecs.SecurityGroups.IsUnknown() {
			resource.ComputeSpecs.SecurityGroups = previousSpecs.SecurityGroups
		} else {
			resource.ComputeSpecs.SecurityGroups = types.ListNull(types.StringType)
		}

		// File System Parameters for FSWAL
		var previousFileSystemParam *FileSystemParamModel
		if previousSpecs != nil {
			var fsDiags diag.Diagnostics
			previousFileSystemParam, fsDiags = FileSystemParamObjectToModel(ctx, previousSpecs.FileSystemParam)
			if fsDiags.HasError() {
				diags.Append(fsDiags...)
			}
		}
		if instance.Spec.FileSystem != nil {
			fileSystemParam := &FileSystemParamModel{
				FileSystemType:               types.StringNull(),
				ThroughputMibpsPerFileSystem: types.Int64Null(),
				FileSystemCount:              types.Int64Null(),
				SecurityGroups:               types.ListNull(types.StringType),
			}

			// Copy previous values if they exist
			if previousFileSystemParam != nil {
				fileSystemParam.FileSystemType = previousFileSystemParam.FileSystemType
				fileSystemParam.ThroughputMibpsPerFileSystem = previousFileSystemParam.ThroughputMibpsPerFileSystem
				fileSystemParam.FileSystemCount = previousFileSystemParam.FileSystemCount
				fileSystemParam.SecurityGroups = previousFileSystemParam.SecurityGroups
			}

			// Update with API response values
			if instance.Spec.FileSystem.FileSystemType != nil {
				fileSystemParam.FileSystemType = types.StringValue(*instance.Spec.FileSystem.FileSystemType)
			}
			if instance.Spec.FileSystem.ThroughputMiBpsPerFileSystem != nil {
				fileSystemParam.ThroughputMibpsPerFileSystem = types.Int64Value(int64(*instance.Spec.FileSystem.ThroughputMiBpsPerFileSystem))
			}
			if instance.Spec.FileSystem.FileSystemCount != nil {
				fileSystemParam.FileSystemCount = types.Int64Value(int64(*instance.Spec.FileSystem.FileSystemCount))
			}
			if len(instance.Spec.FileSystem.SecurityGroups) > 0 {
				securityGroupsList, diags := types.ListValueFrom(ctx, types.StringType, instance.Spec.FileSystem.SecurityGroups)
				if !diags.HasError() {
					fileSystemParam.SecurityGroups = securityGroupsList
				}
			}

			fileSystemObject, objectDiags := FileSystemParamModelToObject(ctx, fileSystemParam)
			if objectDiags.HasError() {
				diags.Append(objectDiags...)
			} else {
				resource.ComputeSpecs.FileSystemParam = fileSystemObject
			}
		} else if previousFileSystemParam != nil {
			// Preserve previous file system parameters if API doesn't return them
			resource.ComputeSpecs.FileSystemParam = previousSpecs.FileSystemParam
		} else {
			resource.ComputeSpecs.FileSystemParam = types.ObjectNull(FileSystemParamObjectType.AttrTypes)
		}

	}

	// Features
	var previousFeatures *FeaturesModel
	if resource.Features != nil {
		prevCopy := *resource.Features
		previousFeatures = &prevCopy
	}
	if instance.Features != nil {
		if resource.Features == nil {
			resource.Features = &FeaturesModel{}
		}

		// WAL Mode
		if instance.Features.WalMode != nil {
			resource.Features.WalMode = types.StringValue(*instance.Features.WalMode)
		}

		// Security
		if instance.Features.Security != nil {
			var previousSecurity *SecurityModel
			if previousFeatures != nil {
				var securityDiags diag.Diagnostics
				previousSecurity, securityDiags = SecurityObjectToModel(ctx, previousFeatures.Security)
				if securityDiags.HasError() {
					diags.Append(securityDiags...)
				}
			}
			security := &SecurityModel{}
			if previousSecurity != nil {
				securityClone := *previousSecurity
				security = &securityClone
			}

			if instance.Features.Security.DataEncryptionMode != nil {
				security.DataEncryptionMode = types.StringValue(*instance.Features.Security.DataEncryptionMode)
			}
			var previousTls *types.Bool
			if previousSecurity != nil {
				previousTls = &previousSecurity.TlsHostnameValidationEnabled
			}
			security.TlsHostnameValidationEnabled = coalesceBoolAttr(
				instance.Features.Security.TlsHostnameValidationEnabled,
				previousTls,
			)

			// Authentication Methods
			if instance.Features.Security.AuthenticationMethods != nil {
				values := make([]attr.Value, len(instance.Features.Security.AuthenticationMethods))
				for i, v := range instance.Features.Security.AuthenticationMethods {
					values[i] = types.StringValue(v)
				}
				set, listDiags := types.SetValue(types.StringType, values)
				if !listDiags.HasError() {
					security.AuthenticationMethods = set
				} else {
					diags.Append(listDiags...)
				}
			}
			// Transit Encryption Modes
			if instance.Features.Security.TransitEncryptionModes != nil {
				values := make([]attr.Value, len(instance.Features.Security.TransitEncryptionModes))
				for i, v := range instance.Features.Security.TransitEncryptionModes {
					values[i] = types.StringValue(v)
				}
				set, listDiags := types.SetValue(types.StringType, values)
				if !listDiags.HasError() {
					security.TransitEncryptionModes = set
				} else {
					diags.Append(listDiags...)
				}
			}
			securityObject, objectDiags := SecurityModelToObject(ctx, security)
			if objectDiags.HasError() {
				diags.Append(objectDiags...)
			} else {
				resource.Features.Security = securityObject
			}
		}

		// Metrics exporter
		var previousMetrics *MetricsExporterModel
		if previousFeatures != nil {
			var metricsDiags diag.Diagnostics
			previousMetrics, metricsDiags = MetricsExporterObjectToModel(ctx, previousFeatures.MetricsExporter)
			if metricsDiags.HasError() {
				diags.Append(metricsDiags...)
			}
		}
		if instance.Features.MetricsExporter != nil && !isMetricsExporterVOEmpty(instance.Features.MetricsExporter) {
			metrics, metricsDiags := flattenMetricsExporterVO(instance.Features.MetricsExporter, previousMetrics)
			diags.Append(metricsDiags...)
			metricsObject, objectDiags := MetricsExporterModelToObject(ctx, metrics)
			if objectDiags.HasError() {
				diags.Append(objectDiags...)
			} else {
				resource.Features.MetricsExporter = metricsObject
			}
		} else {
			resource.Features.MetricsExporter = types.ObjectNull(MetricsExporterObjectType.AttrTypes)
		}

		// Table topic
		var previousTableTopic *TableTopicModel
		if previousFeatures != nil {
			var topicDiags diag.Diagnostics
			previousTableTopic, topicDiags = TableTopicObjectToModel(ctx, previousFeatures.TableTopic)
			if topicDiags.HasError() {
				diags.Append(topicDiags...)
			}
		}
		topic := flattenTableTopicVO(instance.Features.TableTopic, previousTableTopic)
		topicObject, topicDiags := TableTopicModelToObject(ctx, topic)
		if topicDiags.HasError() {
			diags.Append(topicDiags...)
		} else {
			resource.Features.TableTopic = topicObject
		}

		var previousSchemaRegistryEnabled *types.Bool
		if previousFeatures != nil {
			previousSchemaRegistryEnabled = &previousFeatures.SchemaRegistryEnabled
		}
		resource.Features.SchemaRegistryEnabled = coalesceBoolAttr(
			instance.Features.SchemaRegistryEnabled,
			previousSchemaRegistryEnabled,
		)

	}

	// Timestamps
	if instance.GmtCreate != nil {
		resource.CreatedAt = timetypes.NewRFC3339TimePointerValue(instance.GmtCreate)
	}
	if instance.GmtModified != nil {
		resource.LastUpdated = timetypes.NewRFC3339TimePointerValue(instance.GmtModified)
	}

	// Tags - flatten TagVO slice to map
	if len(instance.Tags) > 0 {
		tagsMap := make(map[string]string, len(instance.Tags))
		for _, tag := range instance.Tags {
			if tag.Name != nil && tag.Value != nil {
				tagsMap[*tag.Name] = *tag.Value
			}
		}
		if len(tagsMap) > 0 {
			tagsValue, tagsDiags := types.MapValueFrom(ctx, types.StringType, tagsMap)
			if tagsDiags.HasError() {
				diags.Append(tagsDiags...)
			} else {
				resource.Tags = tagsValue
			}
		}
	} else if resource.Tags.IsUnknown() {
		resource.Tags = types.MapNull(types.StringType)
	}

	return diags
}

func isMetricsExporterVOEmpty(vo *client.InstanceMetricsExporterVO) bool {
	if vo == nil {
		return true
	}
	if vo.Prometheus != nil && !isPrometheusVOEmpty(vo.Prometheus) {
		return false
	}
	return true
}

func flattenMetricsExporterVO(vo *client.InstanceMetricsExporterVO, previous *MetricsExporterModel) (*MetricsExporterModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if vo == nil || isMetricsExporterVOEmpty(vo) {
		return nil, diags
	}

	metrics := MetricsExporterModel{}
	if previous != nil {
		metrics = *previous
	}

	var previousProm *PrometheusExporterModel
	if previous != nil {
		previousProm = previous.Prometheus
	}

	prom, promDiags := flattenPrometheusExporterVO(vo.Prometheus, previousProm)
	diags.Append(promDiags...)
	metrics.Prometheus = prom

	if metrics.Prometheus == nil {
		return nil, diags
	}
	return &metrics, diags
}

func flattenPrometheusExporterVO(vo *client.InstancePrometheusExporterVO, previous *PrometheusExporterModel) (*PrometheusExporterModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if vo == nil || isPrometheusVOEmpty(vo) {
		return nil, diags
	}

	prom := PrometheusExporterModel{
		AuthType:      types.StringNull(),
		EndPoint:      types.StringNull(),
		PrometheusArn: types.StringNull(),
		Username:      types.StringNull(),
		Password:      types.StringNull(),
		Token:         types.StringNull(),
		Labels:        types.MapNull(types.StringType),
	}
	if previous != nil {
		// Preserve any sensitive values (e.g. password/token) that are not returned by the API.
		prom = *previous
	}

	prom.AuthType = retainString(cleanAPIString(vo.AuthType), prom.AuthType)
	prom.EndPoint = retainString(cleanAPIString(vo.EndPoint), prom.EndPoint)
	prom.PrometheusArn = retainString(cleanAPIString(vo.PrometheusArn), prom.PrometheusArn)
	prom.Username = retainString(cleanAPIString(vo.Username), prom.Username)

	if len(vo.Labels) > 0 {
		labelMap := make(map[string]string, len(vo.Labels))
		for _, label := range vo.Labels {
			if label.Name != nil && label.Value != nil {
				labelMap[*label.Name] = *label.Value
			}
		}
		if len(labelMap) > 0 {
			labelsValue, mapDiags := types.MapValueFrom(context.Background(), types.StringType, labelMap)
			diags.Append(mapDiags...)
			if !mapDiags.HasError() {
				prom.Labels = labelsValue
			}
		}
	} else if previous == nil {
		prom.Labels = types.MapNull(types.StringType)
	}

	return &prom, diags
}

func flattenTableTopicVO(vo *client.TableTopicVO, previous *TableTopicModel) *TableTopicModel {
	if vo == nil || vo.CatalogType == nil {
		return nil
	}

	topic := TableTopicModel{
		Warehouse:         types.StringNull(),
		CatalogType:       types.StringNull(),
		MetastoreURI:      types.StringNull(),
		HiveAuthMode:      types.StringNull(),
		KerberosPrincipal: types.StringNull(),
		UserPrincipal:     types.StringNull(),
		KeytabFile:        types.StringNull(),
		Krb5ConfFile:      types.StringNull(),
	}
	if previous != nil {
		topic = *previous
	}

	topic.Warehouse = retainString(vo.Warehouse, topic.Warehouse)
	topic.CatalogType = retainString(vo.CatalogType, topic.CatalogType)
	topic.MetastoreURI = retainString(vo.MetastoreUri, topic.MetastoreURI)
	topic.HiveAuthMode = retainString(vo.HiveAuthMode, topic.HiveAuthMode)
	topic.KerberosPrincipal = retainString(vo.KerberosPrincipal, topic.KerberosPrincipal)
	topic.UserPrincipal = retainString(vo.UserPrincipal, topic.UserPrincipal)
	topic.KeytabFile = retainString(vo.KeytabFile, topic.KeytabFile)
	topic.Krb5ConfFile = retainString(vo.Krb5ConfFile, topic.Krb5ConfFile)

	return &topic
}

func retainString(api *string, previous types.String) types.String {
	if api != nil {
		return types.StringValue(*api)
	}
	return previous
}

func cleanAPIString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isPrometheusVOEmpty(vo *client.InstancePrometheusExporterVO) bool {
	if vo == nil {
		return true
	}
	if vo.AuthType != nil || vo.EndPoint != nil || vo.Username != nil || vo.PrometheusArn != nil {
		return false
	}
	return len(vo.Labels) == 0
}

// FlattenKafkaInstanceModelWithEndpoints adds endpoint information to the KafkaInstanceResourceModel.
func FlattenKafkaInstanceModelWithEndpoints(endpoints []client.InstanceAccessInfoVO, resource *KafkaInstanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if endpoints == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Endpoints",
			"Cannot flatten nil endpoints",
		)}
	}

	// Handle endpoints
	if len(endpoints) > 0 {
		if err := populateInstanceAccessInfoList(context.Background(), resource, endpoints); err.HasError() {
			diags.Append(err...)
		}
	}

	return diags
}

// populateInstanceAccessInfoList converts endpoint information into a list of InstanceAccessInfo.
func populateInstanceAccessInfoList(ctx context.Context, data *KafkaInstanceResourceModel, in []client.InstanceAccessInfoVO) diag.Diagnostics {
	var diags diag.Diagnostics

	if data == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Resource",
			"Cannot populate access info for nil resource",
		)}
	}

	instanceAccessInfoList := make([]InstanceAccessInfo, 0, len(in))
	for _, item := range in {
		if item.DisplayName == nil || item.NetworkType == nil || item.Protocol == nil ||
			item.Mechanisms == nil || item.BootstrapServers == nil {
			continue
		}

		instanceAccessInfoList = append(instanceAccessInfoList, InstanceAccessInfo{
			DisplayName:      types.StringValue(*item.DisplayName),
			NetworkType:      types.StringValue(*item.NetworkType),
			Protocol:         types.StringValue(*item.Protocol),
			Mechanisms:       types.StringValue(*item.Mechanisms),
			BootstrapServers: types.StringValue(*item.BootstrapServers),
		})
	}

	endpointsList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{
		"display_name":      types.StringType,
		"network_type":      types.StringType,
		"protocol":          types.StringType,
		"mechanisms":        types.StringType,
		"bootstrap_servers": types.StringType,
	}}, instanceAccessInfoList)

	if !diags.HasError() {
		data.Endpoints = endpointsList
	}

	return diags
}

// flattenNetworks converts network information into a slice of NetworkModel.
func flattenNetworks(networks []client.InstanceZoneNetworkVO) ([]NetworkModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if networks == nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Networks",
			"Cannot flatten nil networks",
		)}
	}

	networksModel := make([]NetworkModel, 0, len(networks))
	for _, network := range networks {
		if network.Zone == nil {
			continue
		}

		zone := types.StringValue(*network.Zone)
		subnets := make([]attr.Value, 0)
		if network.Subnet != nil {
			subnets = append(subnets, types.StringValue(*network.Subnet))
		}

		subnetList, listDiags := types.ListValue(types.StringType, subnets)
		if listDiags.HasError() {
			diags.Append(listDiags...)
			continue
		}

		networksModel = append(networksModel, NetworkModel{
			Zone:    zone,
			Subnets: subnetList,
		})
	}

	return networksModel, diags
}
