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
	DeployProfile  types.String       `tfsdk:"deploy_profile"`
	Version        types.String       `tfsdk:"version"`
	ComputeSpecs   *ComputeSpecsModel `tfsdk:"compute_specs"`
	Features       *FeaturesModel     `tfsdk:"features"`
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
	DeployProfile  types.String          `tfsdk:"deploy_profile"`
	Version        types.String          `tfsdk:"version"`
	ComputeSpecs   *ComputeSpecsModel    `tfsdk:"compute_specs"`
	Features       *FeaturesSummaryModel `tfsdk:"features"`
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
	ReservedAku           types.Int64            `tfsdk:"reserved_aku"`
	Networks              []NetworkModel         `tfsdk:"networks"`
	KubernetesNodeGroups  []NodeGroupModel       `tfsdk:"kubernetes_node_groups"`
	BucketProfiles        []BucketProfileIDModel `tfsdk:"bucket_profiles"`
	DeployType            types.String           `tfsdk:"deploy_type"`
	DnsZone               types.String           `tfsdk:"dns_zone"`
	KubernetesClusterID   types.String           `tfsdk:"kubernetes_cluster_id"`
	KubernetesNamespace   types.String           `tfsdk:"kubernetes_namespace"`
	KubernetesServiceAcct types.String           `tfsdk:"kubernetes_service_account"`
	InstanceRole          types.String           `tfsdk:"instance_role"`
	DataBuckets           types.List             `tfsdk:"data_buckets"`
}

type NodeGroupModel struct {
	ID types.String `tfsdk:"id"`
}

type DataBucketModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
}

var DataBucketObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"bucket_name": types.StringType,
	},
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

func coalesceStringAttr(apiValue *string, previous *types.String) types.String {
	if apiValue != nil {
		return types.StringValue(*apiValue)
	}
	if previous != nil && !previous.IsNull() && !previous.IsUnknown() {
		return *previous
	}
	return types.StringNull()
}

type FeaturesModel struct {
	WalMode         types.String          `tfsdk:"wal_mode"`
	InstanceConfigs types.Map             `tfsdk:"instance_configs"`
	Integrations    types.Set             `tfsdk:"integrations"`
	Security        *SecurityModel        `tfsdk:"security"`
	MetricsExporter *MetricsExporterModel `tfsdk:"metrics_exporter"`
	TableTopic      *TableTopicModel      `tfsdk:"table_topic"`
}

type FeaturesSummaryModel struct {
	WalMode         types.String          `tfsdk:"wal_mode"`
	InstanceConfigs types.Map             `tfsdk:"instance_configs"`
	Integrations    types.Set             `tfsdk:"integrations"`
	Security        *SecuritySummaryModel `tfsdk:"security"`
	MetricsExporter *MetricsExporterModel `tfsdk:"metrics_exporter"`
	TableTopic      *TableTopicModel      `tfsdk:"table_topic"`
}

type BucketProfileIDModel struct {
	ID types.String `tfsdk:"id"`
}

type SecurityModel struct {
	AuthenticationMethods  types.Set    `tfsdk:"authentication_methods"`
	TransitEncryptionModes types.Set    `tfsdk:"transit_encryption_modes"`
	CertificateAuthority   types.String `tfsdk:"certificate_authority"`
	CertificateChain       types.String `tfsdk:"certificate_chain"`
	PrivateKey             types.String `tfsdk:"private_key"`
	DataEncryptionMode     types.String `tfsdk:"data_encryption_mode"`
}

type SecuritySummaryModel struct {
	AuthenticationMethods  types.Set    `tfsdk:"authentication_methods"`
	TransitEncryptionModes types.Set    `tfsdk:"transit_encryption_modes"`
	DataEncryptionMode     types.String `tfsdk:"data_encryption_mode"`
}

type MetricsExporterModel struct {
	Prometheus *PrometheusExporterModel `tfsdk:"prometheus"`
}

type PrometheusExporterModel struct {
	AuthType      types.String `tfsdk:"auth_type"`
	EndPoint      types.String `tfsdk:"end_point"`
	PrometheusArn types.String `tfsdk:"prometheus_arn"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	Token         types.String `tfsdk:"token"`
	Labels        types.Map    `tfsdk:"labels"`
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

// ExpandKafkaInstanceResource converts a KafkaInstanceResourceModel to a client.InstanceCreateParam.
// It handles the conversion of all nested structures and validates required fields.
func ExpandKafkaInstanceResource(instance KafkaInstanceResourceModel, request *client.InstanceCreateParam) error {
	if request == nil {
		return fmt.Errorf("request parameter cannot be nil")
	}

	// Basic fields
	request.Name = instance.Name.ValueString()
	request.Description = instance.Description.ValueString()
	if !instance.DeployProfile.IsNull() && !instance.DeployProfile.IsUnknown() {
		request.DeployProfile = instance.DeployProfile.ValueString()
	}
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
		request.Spec = client.SpecificationParam{
			ReservedAku: int32(instance.ComputeSpecs.ReservedAku.ValueInt64()),
			NodeConfig:  &client.NodeConfigParam{},
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
		if !instance.ComputeSpecs.InstanceRole.IsNull() && !instance.ComputeSpecs.InstanceRole.IsUnknown() {
			role := instance.ComputeSpecs.InstanceRole.ValueString()
			request.Spec.InstanceRole = &role
		}
		// ignore Node Configs

		// Networks
		// Kubernetes Node Groups
		if len(instance.ComputeSpecs.KubernetesNodeGroups) > 0 {
			nodeGroups := make([]client.KubernetesNodeGroupParam, 0, len(instance.ComputeSpecs.KubernetesNodeGroups))
			for _, ng := range instance.ComputeSpecs.KubernetesNodeGroups {
				id := ng.ID.ValueString()
				nodeGroups = append(nodeGroups, client.KubernetesNodeGroupParam{
					Id: &id,
				})
			}
			request.Spec.KubernetesNodeGroups = nodeGroups
		}
		if len(instance.ComputeSpecs.Networks) > 0 {
			networks := make([]client.InstanceNetworkParam, 0, len(instance.ComputeSpecs.Networks))
			for _, network := range instance.ComputeSpecs.Networks {
				subnet := ""
				if !network.Subnets.IsNull() {
					subnets := ExpandStringValueList(network.Subnets)
					if len(subnets) > 0 {
						subnet = subnets[0]
					}
				}
				networks = append(networks, client.InstanceNetworkParam{
					Zone:   network.Zone.ValueString(),
					Subnet: &subnet,
				})
			}
			request.Spec.Networks = networks
		}
		// Bucket Profiles
		if len(instance.ComputeSpecs.BucketProfiles) > 0 {
			bucketProfiles := make([]client.BucketProfileBindParam, 0, len(instance.ComputeSpecs.BucketProfiles))
			for _, ng := range instance.ComputeSpecs.BucketProfiles {
				id := ng.ID.ValueString()
				bucketProfiles = append(bucketProfiles, client.BucketProfileBindParam{
					Id: &id,
				})
			}
			request.Spec.BucketProfiles = bucketProfiles
		}

		if !instance.ComputeSpecs.DataBuckets.IsNull() && !instance.ComputeSpecs.DataBuckets.IsUnknown() {
			dataBucketModels, dataBucketDiags := DataBucketListToModels(context.TODO(), instance.ComputeSpecs.DataBuckets)
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

	}

	// Features
	if instance.Features != nil {
		// WAL
		walMode := instance.Features.WalMode.ValueString()
		request.Features = &client.InstanceFeatureParam{
			WalMode: &walMode,
		}

		// Security
		if instance.Features.Security != nil {
			if !instance.Features.Security.DataEncryptionMode.IsNull() {
				dataEncryptionMode := instance.Features.Security.DataEncryptionMode.ValueString()
				request.Features.Security = &client.InstanceSecurityParam{
					DataEncryptionMode: &dataEncryptionMode,
				}
			}

			if !instance.Features.Security.AuthenticationMethods.IsNull() {
				var authMethods []string
				diags := instance.Features.Security.AuthenticationMethods.ElementsAs(context.TODO(), &authMethods, false)
				if diags.HasError() {
					return fmt.Errorf("failed to parse authentication methods: %v", diags)
				}
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.AuthenticationMethods = authMethods
			}

			if !instance.Features.Security.TransitEncryptionModes.IsNull() {
				var encryptionModes []string
				diags := instance.Features.Security.TransitEncryptionModes.ElementsAs(context.TODO(), &encryptionModes, false)
				if diags.HasError() {
					return fmt.Errorf("failed to parse transit encryption modes: %v", diags)
				}
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.TransitEncryptionModes = encryptionModes
			}

			if !instance.Features.Security.CertificateAuthority.IsNull() {
				certAuth := instance.Features.Security.CertificateAuthority.ValueString()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.CertificateAuthority = &certAuth
			}

			if !instance.Features.Security.CertificateChain.IsNull() {
				certChain := instance.Features.Security.CertificateChain.ValueString()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.CertificateChain = &certChain
			}

			if !instance.Features.Security.PrivateKey.IsNull() {
				privateKey := instance.Features.Security.PrivateKey.ValueString()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.PrivateKey = &privateKey
			}
		}

		// Metrics exporter
		if instance.Features.MetricsExporter != nil {
			exporter := client.InstanceMetricsExporterParam{}
			hasConfig := false
			if promConfig := instance.Features.MetricsExporter.Prometheus; promConfig != nil {
				prom := &client.InstancePrometheusExporterParam{}
				enabled := true
				prom.Enabled = &enabled
				if !promConfig.AuthType.IsNull() && !promConfig.AuthType.IsUnknown() {
					auth := promConfig.AuthType.ValueString()
					prom.AuthType = &auth
				}
				if !promConfig.EndPoint.IsNull() && !promConfig.EndPoint.IsUnknown() {
					endpoint := promConfig.EndPoint.ValueString()
					prom.EndPoint = &endpoint
				}
				if !promConfig.PrometheusArn.IsNull() && !promConfig.PrometheusArn.IsUnknown() {
					arn := promConfig.PrometheusArn.ValueString()
					prom.PrometheusArn = &arn
				}
				if !promConfig.Username.IsNull() && !promConfig.Username.IsUnknown() {
					username := promConfig.Username.ValueString()
					prom.Username = &username
				}
				if !promConfig.Password.IsNull() && !promConfig.Password.IsUnknown() {
					password := promConfig.Password.ValueString()
					prom.Password = &password
				}
				if !promConfig.Token.IsNull() && !promConfig.Token.IsUnknown() {
					token := promConfig.Token.ValueString()
					prom.Token = &token
				}
				if !promConfig.Labels.IsNull() && !promConfig.Labels.IsUnknown() {
					configs := ExpandStringValueMap(promConfig.Labels)
					if len(configs) > 0 {
						labelSlice := make([]client.MetricsLabelParam, len(configs))
						for i, cfg := range configs {
							labelSlice[i] = client.MetricsLabelParam{Name: cfg.Key, Value: cfg.Value}
						}
						prom.Labels = labelSlice
					}
				}
				exporter.Prometheus = prom
				hasConfig = true
			}
			if hasConfig {
				request.Features.MetricsExporter = &exporter
			}
		}

		// Table topic
		if instance.Features.TableTopic != nil {
			topicModel := instance.Features.TableTopic
			if topicModel.Warehouse.IsNull() || topicModel.Warehouse.IsUnknown() {
				return fmt.Errorf("features.table_topic.warehouse is required when table_topic is set")
			}
			if topicModel.CatalogType.IsNull() || topicModel.CatalogType.IsUnknown() {
				return fmt.Errorf("features.table_topic.catalog_type is required when table_topic is set")
			}
			topic := &client.TableTopicParam{
				Warehouse:   topicModel.Warehouse.ValueString(),
				CatalogType: topicModel.CatalogType.ValueString(),
			}
			if !topicModel.MetastoreURI.IsNull() && !topicModel.MetastoreURI.IsUnknown() {
				uri := topicModel.MetastoreURI.ValueString()
				topic.MetastoreUri = &uri
			}
			if !topicModel.HiveAuthMode.IsNull() && !topicModel.HiveAuthMode.IsUnknown() {
				mode := topicModel.HiveAuthMode.ValueString()
				topic.HiveAuthMode = &mode
			}
			if !topicModel.KerberosPrincipal.IsNull() && !topicModel.KerberosPrincipal.IsUnknown() {
				principal := topicModel.KerberosPrincipal.ValueString()
				topic.KerberosPrincipal = &principal
			}
			if !topicModel.UserPrincipal.IsNull() && !topicModel.UserPrincipal.IsUnknown() {
				principal := topicModel.UserPrincipal.ValueString()
				topic.UserPrincipal = &principal
			}
			if !topicModel.KeytabFile.IsNull() && !topicModel.KeytabFile.IsUnknown() {
				keytab := topicModel.KeytabFile.ValueString()
				topic.KeytabFile = &keytab
			}
			if !topicModel.Krb5ConfFile.IsNull() && !topicModel.Krb5ConfFile.IsUnknown() {
				conf := topicModel.Krb5ConfFile.ValueString()
				topic.Krb5ConfFile = &conf
			}
			request.Features.TableTopic = topic
		}

		// Integrations
		if !instance.Features.Integrations.IsNull() && len(instance.Features.Integrations.Elements()) > 0 {
			ids := ExpandSetValueList(instance.Features.Integrations)
			integrations := make([]client.IntegrationBindParam, 0, len(ids))
			for _, id := range ids {
				integrationID := id
				integrations = append(integrations, client.IntegrationBindParam{
					Id: &integrationID,
				})
			}
			request.Features.Integrations = integrations
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
	model.DeployProfile = resource.DeployProfile
	model.Version = resource.Version
	model.ComputeSpecs = resource.ComputeSpecs
	if resource.Features != nil {
		features := &FeaturesSummaryModel{
			WalMode:         resource.Features.WalMode,
			InstanceConfigs: resource.Features.InstanceConfigs,
			Integrations:    resource.Features.Integrations,
			MetricsExporter: resource.Features.MetricsExporter,
			TableTopic:      resource.Features.TableTopic,
		}
		if resource.Features.Security != nil {
			features.Security = &SecuritySummaryModel{
				AuthenticationMethods:  resource.Features.Security.AuthenticationMethods,
				TransitEncryptionModes: resource.Features.Security.TransitEncryptionModes,
				DataEncryptionMode:     resource.Features.Security.DataEncryptionMode,
			}
		}
		model.Features = features
	} else {
		model.Features = nil
	}
	model.Endpoints = resource.Endpoints
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
	if instance.DeployProfile != nil {
		resource.DeployProfile = types.StringValue(*instance.DeployProfile)
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
func FlattenKafkaInstanceModel(instance *client.InstanceVO, resource *KafkaInstanceResourceModel) diag.Diagnostics {
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
	if instance.DeployProfile != nil {
		resource.DeployProfile = types.StringValue(*instance.DeployProfile)
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
				Networks:              []NetworkModel{},
				KubernetesNodeGroups:  []NodeGroupModel{},
				BucketProfiles:        []BucketProfileIDModel{},
				DataBuckets:           types.ListNull(DataBucketObjectType),
				DeployType:            types.StringNull(),
				DnsZone:               types.StringNull(),
				KubernetesClusterID:   types.StringNull(),
				KubernetesNamespace:   types.StringNull(),
				KubernetesServiceAcct: types.StringNull(),
				InstanceRole:          types.StringNull(),
			}
		}
		// Reserved AKU
		if instance.Spec.ReservedAku != nil {
			resource.ComputeSpecs.ReservedAku = types.Int64Value(int64(*instance.Spec.ReservedAku))
		}
		var prevDeploy, prevDnsZone *types.String
		var prevClusterID, prevNamespace, prevServiceAccount, prevInstanceRole *types.String
		if previousSpecs != nil {
			prevDeploy = &previousSpecs.DeployType
			prevDnsZone = &previousSpecs.DnsZone
			prevClusterID = &previousSpecs.KubernetesClusterID
			prevNamespace = &previousSpecs.KubernetesNamespace
			prevServiceAccount = &previousSpecs.KubernetesServiceAcct
			prevInstanceRole = &previousSpecs.InstanceRole
		}
		resource.ComputeSpecs.DeployType = coalesceStringAttr(instance.Spec.DeployType, prevDeploy)
		resource.ComputeSpecs.DnsZone = coalesceStringAttr(instance.Spec.DnsZone, prevDnsZone)
		resource.ComputeSpecs.KubernetesClusterID = coalesceStringAttr(instance.Spec.KubernetesClusterId, prevClusterID)
		resource.ComputeSpecs.KubernetesNamespace = coalesceStringAttr(instance.Spec.KubernetesNamespace, prevNamespace)
		resource.ComputeSpecs.KubernetesServiceAcct = coalesceStringAttr(instance.Spec.KubernetesServiceAccount, prevServiceAccount)
		resource.ComputeSpecs.InstanceRole = coalesceStringAttr(instance.Spec.InstanceRole, prevInstanceRole)

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
			resource.ComputeSpecs.KubernetesNodeGroups = nodeGroups
			// Networks
		}
		if instance.Spec.Networks != nil {
			networks, networkDiags := flattenNetworks(instance.Spec.Networks)
			if networkDiags.HasError() {
				diags.Append(networkDiags...)
			} else {
				resource.ComputeSpecs.Networks = networks
			}
		}

		// Bucket Profiles
		if instance.Spec.BucketProfiles != nil {
			bucketProfiles := make([]BucketProfileIDModel, 0, len(instance.Spec.BucketProfiles))
			for _, bp := range instance.Spec.BucketProfiles {
				if bp.Id != nil {
					bucketProfiles = append(bucketProfiles, BucketProfileIDModel{
						ID: types.StringValue(*bp.Id),
					})
				}
			}
			resource.ComputeSpecs.BucketProfiles = bucketProfiles
		}

		var previousDataBuckets []DataBucketModel
		if previousSpecs != nil {
			prevBuckets, prevDiags := DataBucketListToModels(context.TODO(), previousSpecs.DataBuckets)
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
			listValue, listDiags := DataBucketModelsToList(context.TODO(), dataBuckets)
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
				previousSecurity = previousFeatures.Security
			}
			if previousSecurity != nil {
				clone := *previousSecurity
				resource.Features.Security = &clone
			} else {
				resource.Features.Security = &SecurityModel{}
			}

			if instance.Features.Security.DataEncryptionMode != nil {
				resource.Features.Security.DataEncryptionMode = types.StringValue(*instance.Features.Security.DataEncryptionMode)
			}

			// Authentication Methods
			if instance.Features.Security.AuthenticationMethods != nil {
				values := make([]attr.Value, len(instance.Features.Security.AuthenticationMethods))
				for i, v := range instance.Features.Security.AuthenticationMethods {
					values[i] = types.StringValue(v)
				}
				set, listDiags := types.SetValue(types.StringType, values)
				if !listDiags.HasError() {
					resource.Features.Security.AuthenticationMethods = set
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
					resource.Features.Security.TransitEncryptionModes = set
				} else {
					diags.Append(listDiags...)
				}
			}
		}

		// Metrics exporter
		var previousMetrics *MetricsExporterModel
		if previousFeatures != nil {
			previousMetrics = previousFeatures.MetricsExporter
		}
		if instance.Features.MetricsExporter != nil && !isMetricsExporterVOEmpty(instance.Features.MetricsExporter) {
			metrics, metricsDiags := flattenMetricsExporterVO(instance.Features.MetricsExporter, previousMetrics)
			diags.Append(metricsDiags...)
			resource.Features.MetricsExporter = metrics
		} else if previousMetrics != nil {
			resource.Features.MetricsExporter = previousMetrics
		} else {
			resource.Features.MetricsExporter = nil
		}

		// Table topic
		var previousTableTopic *TableTopicModel
		if previousFeatures != nil {
			previousTableTopic = previousFeatures.TableTopic
		}
		resource.Features.TableTopic = flattenTableTopicVO(instance.Features.TableTopic, previousTableTopic)

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
		prom = *previous
	}

	prom.AuthType = retainString(cleanAPIString(vo.AuthType), prom.AuthType)
	prom.EndPoint = retainString(cleanAPIString(vo.EndPoint), prom.EndPoint)
	prom.PrometheusArn = retainString(cleanAPIString(vo.PrometheusArn), prom.PrometheusArn)
	prom.Username = retainString(cleanAPIString(vo.Username), prom.Username)
	prom.Password = prom.Password
	prom.Token = prom.Token

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
	if vo == nil || !vo.Enabled {
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

// FlattenKafkaInstanceModelWithIntegrations adds integration information to the KafkaInstanceResourceModel.
func FlattenKafkaInstanceModelWithIntegrations(integrations []client.IntegrationVO, resource *KafkaInstanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if resource.Features == nil {
		resource.Features = &FeaturesModel{}
	}

	if len(integrations) == 0 {
		resource.Features.Integrations = types.SetNull(types.StringType)
		return diags
	}

	integrationIds := make([]attr.Value, 0, len(integrations))
	for _, integration := range integrations {
		if integration.Code == "" {
			continue
		}
		integrationIds = append(integrationIds, types.StringValue(integration.Code))
	}

	if len(integrationIds) == 0 {
		resource.Features.Integrations = types.SetNull(types.StringType)
		return diags
	}

	resource.Features.Integrations = types.SetValueMust(types.StringType, integrationIds)
	return diags
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
