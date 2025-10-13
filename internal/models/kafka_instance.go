package models

import (
	"context"
	"fmt"
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
	FileSystemParam       *FileSystemParamModel  `tfsdk:"file_system_param"`
	DeployType            types.String           `tfsdk:"deploy_type"`
	Provider              types.String           `tfsdk:"provider"`
	Region                types.String           `tfsdk:"region"`
	Scope                 types.String           `tfsdk:"scope"`
	Vpc                   types.String           `tfsdk:"vpc"`
	Domain                types.String           `tfsdk:"domain"`
	DnsZone               types.String           `tfsdk:"dns_zone"`
	KubernetesClusterID   types.String           `tfsdk:"kubernetes_cluster_id"`
	KubernetesNamespace   types.String           `tfsdk:"kubernetes_namespace"`
	KubernetesServiceAcct types.String           `tfsdk:"kubernetes_service_account"`
	Credential            types.String           `tfsdk:"credential"`
	InstanceRole          types.String           `tfsdk:"instance_role"`
	DataBuckets           []DataBucketModel      `tfsdk:"data_buckets"`
	TenantID              types.String           `tfsdk:"tenant_id"`
	VpcResourceGroup      types.String           `tfsdk:"vpc_resource_group"`
	K8sResourceGroup      types.String           `tfsdk:"k8s_resource_group"`
	DnsResourceGroup      types.String           `tfsdk:"dns_resource_group"`
}

type NodeGroupModel struct {
	ID types.String `tfsdk:"id"`
}

type FileSystemParamModel struct {
	ThroughputMiBpsPerFileSystem types.Int64 `tfsdk:"throughput_mibps_per_file_system"`
	FileSystemCount              types.Int64 `tfsdk:"file_system_count"`
}

type DataBucketModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Provider   types.String `tfsdk:"provider"`
	Region     types.String `tfsdk:"region"`
	Scope      types.String `tfsdk:"scope"`
	Credential types.String `tfsdk:"credential"`
	Endpoint   types.String `tfsdk:"endpoint"`
}

type FeaturesModel struct {
	WalMode         types.String            `tfsdk:"wal_mode"`
	InstanceConfigs types.Map               `tfsdk:"instance_configs"`
	Integrations    types.Set               `tfsdk:"integrations"`
	Security        *SecurityModel          `tfsdk:"security"`
	MetricsExporter *MetricsExporterModel   `tfsdk:"metrics_exporter"`
	TableTopic      *TableTopicModel        `tfsdk:"table_topic"`
	S3Failover      *FailoverModel          `tfsdk:"s3_failover"`
	InboundRules    []InboundRuleModel      `tfsdk:"inbound_rules"`
	ExtendListeners []InstanceListenerModel `tfsdk:"extend_listeners"`
}

type FeaturesSummaryModel struct {
	WalMode         types.String            `tfsdk:"wal_mode"`
	InstanceConfigs types.Map               `tfsdk:"instance_configs"`
	Integrations    types.Set               `tfsdk:"integrations"`
	Security        *SecuritySummaryModel   `tfsdk:"security"`
	MetricsExporter *MetricsExporterModel   `tfsdk:"metrics_exporter"`
	TableTopic      *TableTopicModel        `tfsdk:"table_topic"`
	S3Failover      *FailoverModel          `tfsdk:"s3_failover"`
	InboundRules    []InboundRuleModel      `tfsdk:"inbound_rules"`
	ExtendListeners []InstanceListenerModel `tfsdk:"extend_listeners"`
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
	Prometheus *PrometheusExporterModel   `tfsdk:"prometheus"`
	CloudWatch *CloudWatchExporterModel   `tfsdk:"cloudwatch"`
	Kafka      *KafkaMetricsExporterModel `tfsdk:"kafka"`
}

type PrometheusExporterModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	AuthType      types.String `tfsdk:"auth_type"`
	EndPoint      types.String `tfsdk:"end_point"`
	PrometheusArn types.String `tfsdk:"prometheus_arn"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	Token         types.String `tfsdk:"token"`
	Labels        types.Map    `tfsdk:"labels"`
}

type CloudWatchExporterModel struct {
	Enabled   types.Bool   `tfsdk:"enabled"`
	Namespace types.String `tfsdk:"namespace"`
}

type KafkaMetricsExporterModel struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	BootstrapServers types.String `tfsdk:"bootstrap_servers"`
	Topic            types.String `tfsdk:"topic"`
	CollectionPeriod types.Int64  `tfsdk:"collection_period"`
	SecurityProtocol types.String `tfsdk:"security_protocol"`
	SaslMechanism    types.String `tfsdk:"sasl_mechanism"`
	SaslUsername     types.String `tfsdk:"sasl_username"`
	SaslPassword     types.String `tfsdk:"sasl_password"`
}

type FailoverModel struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	StorageType       types.String `tfsdk:"storage_type"`
	EbsVolumeSizeInGB types.Int64  `tfsdk:"ebs_volume_size_gb"`
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

type InboundRuleModel struct {
	ListenerName types.String `tfsdk:"listener_name"`
	Cidrs        types.List   `tfsdk:"cidrs"`
}

type InstanceListenerModel struct {
	ListenerName     types.String `tfsdk:"listener_name"`
	SecurityProtocol types.String `tfsdk:"security_protocol"`
	Port             types.Int64  `tfsdk:"port"`
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
		if !instance.ComputeSpecs.Provider.IsNull() && !instance.ComputeSpecs.Provider.IsUnknown() {
			provider := instance.ComputeSpecs.Provider.ValueString()
			request.Spec.Provider = &provider
		}
		if !instance.ComputeSpecs.Region.IsNull() && !instance.ComputeSpecs.Region.IsUnknown() {
			region := instance.ComputeSpecs.Region.ValueString()
			request.Spec.Region = &region
		}
		if !instance.ComputeSpecs.Scope.IsNull() && !instance.ComputeSpecs.Scope.IsUnknown() {
			scope := instance.ComputeSpecs.Scope.ValueString()
			request.Spec.Scope = &scope
		}
		if !instance.ComputeSpecs.Vpc.IsNull() && !instance.ComputeSpecs.Vpc.IsUnknown() {
			vpc := instance.ComputeSpecs.Vpc.ValueString()
			request.Spec.Vpc = &vpc
		}
		if !instance.ComputeSpecs.Domain.IsNull() && !instance.ComputeSpecs.Domain.IsUnknown() {
			domain := instance.ComputeSpecs.Domain.ValueString()
			request.Spec.Domain = &domain
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
		if !instance.ComputeSpecs.Credential.IsNull() && !instance.ComputeSpecs.Credential.IsUnknown() {
			cred := instance.ComputeSpecs.Credential.ValueString()
			request.Spec.Credential = &cred
		}
		if !instance.ComputeSpecs.InstanceRole.IsNull() && !instance.ComputeSpecs.InstanceRole.IsUnknown() {
			role := instance.ComputeSpecs.InstanceRole.ValueString()
			request.Spec.InstanceRole = &role
		}
		if !instance.ComputeSpecs.TenantID.IsNull() && !instance.ComputeSpecs.TenantID.IsUnknown() {
			tenant := instance.ComputeSpecs.TenantID.ValueString()
			request.Spec.TenantId = &tenant
		}
		if !instance.ComputeSpecs.VpcResourceGroup.IsNull() && !instance.ComputeSpecs.VpcResourceGroup.IsUnknown() {
			rg := instance.ComputeSpecs.VpcResourceGroup.ValueString()
			request.Spec.VpcResourceGroup = &rg
		}
		if !instance.ComputeSpecs.K8sResourceGroup.IsNull() && !instance.ComputeSpecs.K8sResourceGroup.IsUnknown() {
			k8sRg := instance.ComputeSpecs.K8sResourceGroup.ValueString()
			request.Spec.K8sResourceGroup = &k8sRg
		}
		if !instance.ComputeSpecs.DnsResourceGroup.IsNull() && !instance.ComputeSpecs.DnsResourceGroup.IsUnknown() {
			dnsRg := instance.ComputeSpecs.DnsResourceGroup.ValueString()
			request.Spec.DnsResourceGroup = &dnsRg
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
		} else if len(instance.ComputeSpecs.Networks) > 0 {
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

		if len(instance.ComputeSpecs.DataBuckets) > 0 {
			dataBuckets := make([]client.BucketProfileParam, 0, len(instance.ComputeSpecs.DataBuckets))
			for _, bucket := range instance.ComputeSpecs.DataBuckets {
				if bucket.BucketName.IsNull() || bucket.BucketName.IsUnknown() {
					return fmt.Errorf("compute_specs.data_buckets.bucket_name is required")
				}
				profile := client.BucketProfileParam{
					BucketName: bucket.BucketName.ValueString(),
				}
				if !bucket.Provider.IsNull() && !bucket.Provider.IsUnknown() {
					val := bucket.Provider.ValueString()
					profile.Provider = &val
				}
				if !bucket.Region.IsNull() && !bucket.Region.IsUnknown() {
					val := bucket.Region.ValueString()
					profile.Region = &val
				}
				if !bucket.Scope.IsNull() && !bucket.Scope.IsUnknown() {
					val := bucket.Scope.ValueString()
					profile.Scope = &val
				}
				if !bucket.Credential.IsNull() && !bucket.Credential.IsUnknown() {
					val := bucket.Credential.ValueString()
					profile.Credential = &val
				}
				if !bucket.Endpoint.IsNull() && !bucket.Endpoint.IsUnknown() {
					val := bucket.Endpoint.ValueString()
					profile.Endpoint = &val
				}
				dataBuckets = append(dataBuckets, profile)
			}
			request.Spec.DataBuckets = dataBuckets
		}

		if instance.ComputeSpecs.FileSystemParam != nil {
			fsParam := instance.ComputeSpecs.FileSystemParam
			if fsParam.ThroughputMiBpsPerFileSystem.IsNull() || fsParam.ThroughputMiBpsPerFileSystem.IsUnknown() {
				return fmt.Errorf("compute_specs.file_system_param.throughput_mibps_per_file_system must be provided when file_system_param is set")
			}
			if fsParam.FileSystemCount.IsNull() || fsParam.FileSystemCount.IsUnknown() {
				return fmt.Errorf("compute_specs.file_system_param.file_system_count must be provided when file_system_param is set")
			}
			request.Spec.FileSystem = &client.FileSystemParam{
				ThroughputMiBpsPerFileSystem: int32(fsParam.ThroughputMiBpsPerFileSystem.ValueInt64()),
				FileSystemCount:              int32(fsParam.FileSystemCount.ValueInt64()),
			}
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

		// S3 Failover
		if instance.Features.S3Failover != nil {
			failoverModel := instance.Features.S3Failover
			failover := &client.InstanceFailoverParam{}
			if !failoverModel.Enabled.IsNull() && !failoverModel.Enabled.IsUnknown() {
				enabled := failoverModel.Enabled.ValueBool()
				failover.Enabled = &enabled
			}
			if !failoverModel.StorageType.IsNull() && !failoverModel.StorageType.IsUnknown() {
				storageType := failoverModel.StorageType.ValueString()
				failover.StorageType = &storageType
			}
			if !failoverModel.EbsVolumeSizeInGB.IsNull() && !failoverModel.EbsVolumeSizeInGB.IsUnknown() {
				size := int32(failoverModel.EbsVolumeSizeInGB.ValueInt64())
				failover.EbsVolumeSizeInGB = &size
			}
			request.Features.S3Failover = failover
		}

		// Metrics exporter
		if instance.Features.MetricsExporter != nil {
			exporter := client.InstanceMetricsExporterParam{}
			hasConfig := false
			if instance.Features.MetricsExporter.Prometheus != nil {
				promConfig := instance.Features.MetricsExporter.Prometheus
				prom := &client.InstancePrometheusExporterParam{}
				if !promConfig.Enabled.IsNull() && !promConfig.Enabled.IsUnknown() {
					enabled := promConfig.Enabled.ValueBool()
					prom.Enabled = &enabled
				}
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
			if instance.Features.MetricsExporter.CloudWatch != nil {
				cwConfig := instance.Features.MetricsExporter.CloudWatch
				cw := &client.InstanceCloudWatchExporterParam{}
				if !cwConfig.Enabled.IsNull() && !cwConfig.Enabled.IsUnknown() {
					enabled := cwConfig.Enabled.ValueBool()
					cw.Enabled = &enabled
				}
				if !cwConfig.Namespace.IsNull() && !cwConfig.Namespace.IsUnknown() {
					ns := cwConfig.Namespace.ValueString()
					cw.Namespace = &ns
				}
				exporter.CloudWatch = cw
				hasConfig = true
			}
			if instance.Features.MetricsExporter.Kafka != nil {
				kafkaConfig := instance.Features.MetricsExporter.Kafka
				kafka := &client.InstanceKafkaMetricsExporterParam{}
				if !kafkaConfig.Enabled.IsNull() && !kafkaConfig.Enabled.IsUnknown() {
					enabled := kafkaConfig.Enabled.ValueBool()
					kafka.Enabled = &enabled
				}
				if !kafkaConfig.BootstrapServers.IsNull() && !kafkaConfig.BootstrapServers.IsUnknown() {
					servers := kafkaConfig.BootstrapServers.ValueString()
					kafka.BootstrapServers = &servers
				}
				if !kafkaConfig.Topic.IsNull() && !kafkaConfig.Topic.IsUnknown() {
					topic := kafkaConfig.Topic.ValueString()
					kafka.Topic = &topic
				}
				if !kafkaConfig.CollectionPeriod.IsNull() && !kafkaConfig.CollectionPeriod.IsUnknown() {
					period := int32(kafkaConfig.CollectionPeriod.ValueInt64())
					kafka.CollectionPeriod = &period
				}
				if !kafkaConfig.SecurityProtocol.IsNull() && !kafkaConfig.SecurityProtocol.IsUnknown() {
					protocol := kafkaConfig.SecurityProtocol.ValueString()
					kafka.SecurityProtocol = &protocol
				}
				if !kafkaConfig.SaslMechanism.IsNull() && !kafkaConfig.SaslMechanism.IsUnknown() {
					mechanism := kafkaConfig.SaslMechanism.ValueString()
					kafka.SaslMechanism = &mechanism
				}
				if !kafkaConfig.SaslUsername.IsNull() && !kafkaConfig.SaslUsername.IsUnknown() {
					username := kafkaConfig.SaslUsername.ValueString()
					kafka.SaslUsername = &username
				}
				if !kafkaConfig.SaslPassword.IsNull() && !kafkaConfig.SaslPassword.IsUnknown() {
					password := kafkaConfig.SaslPassword.ValueString()
					kafka.SaslPassword = &password
				}
				exporter.Kafka = kafka
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

		// Inbound rules
		if len(instance.Features.InboundRules) > 0 {
			rules := make([]client.InboundRuleParam, 0, len(instance.Features.InboundRules))
			for _, rule := range instance.Features.InboundRules {
				if rule.ListenerName.IsNull() || rule.ListenerName.IsUnknown() {
					return fmt.Errorf("features.inbound_rules.listener_name is required")
				}
				cidrs := []string{}
				if !rule.Cidrs.IsNull() && !rule.Cidrs.IsUnknown() {
					cidrs = ExpandStringValueList(rule.Cidrs)
				}
				rules = append(rules, client.InboundRuleParam{
					ListenerName: rule.ListenerName.ValueString(),
					Cidrs:        cidrs,
				})
			}
			request.Features.InboundRules = rules
		}

		// Extended listeners
		if len(instance.Features.ExtendListeners) > 0 {
			listeners := make([]client.InstanceListenerParam, 0, len(instance.Features.ExtendListeners))
			for _, listener := range instance.Features.ExtendListeners {
				if listener.ListenerName.IsNull() || listener.ListenerName.IsUnknown() {
					return fmt.Errorf("features.extend_listeners.listener_name is required")
				}
				param := client.InstanceListenerParam{
					ListenerName: listener.ListenerName.ValueString(),
				}
				if !listener.SecurityProtocol.IsNull() && !listener.SecurityProtocol.IsUnknown() {
					protocol := listener.SecurityProtocol.ValueString()
					param.SecurityProtocol = &protocol
				}
				if !listener.Port.IsNull() && !listener.Port.IsUnknown() {
					port := int32(listener.Port.ValueInt64())
					param.Port = &port
				}
				listeners = append(listeners, param)
			}
			request.Features.ExtendListeners = listeners
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
			S3Failover:      resource.Features.S3Failover,
			InboundRules:    resource.Features.InboundRules,
			ExtendListeners: resource.Features.ExtendListeners,
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
		if resource.ComputeSpecs == nil {
			resource.ComputeSpecs = &ComputeSpecsModel{
				ReservedAku:          types.Int64Null(),
				Networks:             []NetworkModel{},
				KubernetesNodeGroups: []NodeGroupModel{},
				BucketProfiles:       []BucketProfileIDModel{},
			}
		}
		resource.ComputeSpecs.FileSystemParam = nil
		// Reserved AKU
		if instance.Spec.ReservedAku != nil {
			resource.ComputeSpecs.ReservedAku = types.Int64Value(int64(*instance.Spec.ReservedAku))
		}
		if instance.Spec.DeployType != nil {
			resource.ComputeSpecs.DeployType = types.StringValue(*instance.Spec.DeployType)
		} else {
			resource.ComputeSpecs.DeployType = types.StringNull()
		}
		if instance.Spec.Provider != nil {
			resource.ComputeSpecs.Provider = types.StringValue(*instance.Spec.Provider)
		} else {
			resource.ComputeSpecs.Provider = types.StringNull()
		}
		if instance.Spec.Region != nil {
			resource.ComputeSpecs.Region = types.StringValue(*instance.Spec.Region)
		} else {
			resource.ComputeSpecs.Region = types.StringNull()
		}
		if instance.Spec.Scope != nil {
			resource.ComputeSpecs.Scope = types.StringValue(*instance.Spec.Scope)
		} else {
			resource.ComputeSpecs.Scope = types.StringNull()
		}
		if instance.Spec.Vpc != nil {
			resource.ComputeSpecs.Vpc = types.StringValue(*instance.Spec.Vpc)
		} else {
			resource.ComputeSpecs.Vpc = types.StringNull()
		}
		if instance.Spec.Domain != nil {
			resource.ComputeSpecs.Domain = types.StringValue(*instance.Spec.Domain)
		} else {
			resource.ComputeSpecs.Domain = types.StringNull()
		}
		if instance.Spec.DnsZone != nil {
			resource.ComputeSpecs.DnsZone = types.StringValue(*instance.Spec.DnsZone)
		} else {
			resource.ComputeSpecs.DnsZone = types.StringNull()
		}
		if instance.Spec.KubernetesClusterId != nil {
			resource.ComputeSpecs.KubernetesClusterID = types.StringValue(*instance.Spec.KubernetesClusterId)
		} else {
			resource.ComputeSpecs.KubernetesClusterID = types.StringNull()
		}
		if instance.Spec.KubernetesNamespace != nil {
			resource.ComputeSpecs.KubernetesNamespace = types.StringValue(*instance.Spec.KubernetesNamespace)
		} else {
			resource.ComputeSpecs.KubernetesNamespace = types.StringNull()
		}
		if instance.Spec.KubernetesServiceAccount != nil {
			resource.ComputeSpecs.KubernetesServiceAcct = types.StringValue(*instance.Spec.KubernetesServiceAccount)
		} else {
			resource.ComputeSpecs.KubernetesServiceAcct = types.StringNull()
		}
		if instance.Spec.Credential != nil {
			resource.ComputeSpecs.Credential = types.StringValue(*instance.Spec.Credential)
		} else {
			resource.ComputeSpecs.Credential = types.StringNull()
		}
		if instance.Spec.InstanceRole != nil {
			resource.ComputeSpecs.InstanceRole = types.StringValue(*instance.Spec.InstanceRole)
		} else {
			resource.ComputeSpecs.InstanceRole = types.StringNull()
		}
		if instance.Spec.TenantId != nil {
			resource.ComputeSpecs.TenantID = types.StringValue(*instance.Spec.TenantId)
		} else {
			resource.ComputeSpecs.TenantID = types.StringNull()
		}
		if instance.Spec.VpcResourceGroup != nil {
			resource.ComputeSpecs.VpcResourceGroup = types.StringValue(*instance.Spec.VpcResourceGroup)
		} else {
			resource.ComputeSpecs.VpcResourceGroup = types.StringNull()
		}
		if instance.Spec.K8sResourceGroup != nil {
			resource.ComputeSpecs.K8sResourceGroup = types.StringValue(*instance.Spec.K8sResourceGroup)
		} else {
			resource.ComputeSpecs.K8sResourceGroup = types.StringNull()
		}
		if instance.Spec.DnsResourceGroup != nil {
			resource.ComputeSpecs.DnsResourceGroup = types.StringValue(*instance.Spec.DnsResourceGroup)
		} else {
			resource.ComputeSpecs.DnsResourceGroup = types.StringNull()
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
			resource.ComputeSpecs.KubernetesNodeGroups = nodeGroups
			// Networks
		} else if instance.Spec.Networks != nil {
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

		if instance.Spec.DataBuckets != nil {
			dataBuckets := make([]DataBucketModel, 0, len(instance.Spec.DataBuckets))
			for _, bucket := range instance.Spec.DataBuckets {
				model := DataBucketModel{
					BucketName: types.StringNull(),
					Provider:   types.StringNull(),
					Region:     types.StringNull(),
					Scope:      types.StringNull(),
					Credential: types.StringNull(),
					Endpoint:   types.StringNull(),
				}
				if bucket.BucketName != nil {
					model.BucketName = types.StringValue(*bucket.BucketName)
				}
				if bucket.Provider != nil {
					model.Provider = types.StringValue(*bucket.Provider)
				}
				if bucket.Region != nil {
					model.Region = types.StringValue(*bucket.Region)
				}
				if bucket.Scope != nil {
					model.Scope = types.StringValue(*bucket.Scope)
				}
				if bucket.Credential != nil {
					model.Credential = types.StringValue(*bucket.Credential)
				}
				if bucket.Endpoint != nil {
					model.Endpoint = types.StringValue(*bucket.Endpoint)
				}
				dataBuckets = append(dataBuckets, model)
			}
			resource.ComputeSpecs.DataBuckets = dataBuckets
		} else {
			resource.ComputeSpecs.DataBuckets = nil
		}

		if instance.Spec.FileSystem != nil {
			resource.ComputeSpecs.FileSystemParam = &FileSystemParamModel{}
			if instance.Spec.FileSystem.ThroughputMiBpsPerFileSystem != nil {
				resource.ComputeSpecs.FileSystemParam.ThroughputMiBpsPerFileSystem = types.Int64Value(int64(*instance.Spec.FileSystem.ThroughputMiBpsPerFileSystem))
			} else {
				resource.ComputeSpecs.FileSystemParam.ThroughputMiBpsPerFileSystem = types.Int64Null()
			}
			if instance.Spec.FileSystem.FileSystemCount != nil {
				resource.ComputeSpecs.FileSystemParam.FileSystemCount = types.Int64Value(int64(*instance.Spec.FileSystem.FileSystemCount))
			} else {
				resource.ComputeSpecs.FileSystemParam.FileSystemCount = types.Int64Null()
			}
		}
	}

	// Features
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
			if resource.Features.Security == nil {
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

		// S3 Failover
		if instance.Features.S3Failover != nil {
			resource.Features.S3Failover = &FailoverModel{
				Enabled:           types.BoolValue(instance.Features.S3Failover.Enabled),
				StorageType:       types.StringNull(),
				EbsVolumeSizeInGB: types.Int64Null(),
			}
			if instance.Features.S3Failover.StorageType != nil {
				resource.Features.S3Failover.StorageType = types.StringValue(*instance.Features.S3Failover.StorageType)
			}
			if instance.Features.S3Failover.EbsVolumeSizeInGB != nil {
				resource.Features.S3Failover.EbsVolumeSizeInGB = types.Int64Value(int64(*instance.Features.S3Failover.EbsVolumeSizeInGB))
			}
		} else {
			resource.Features.S3Failover = nil
		}

		// Metrics exporter
		if instance.Features.MetricsExporter != nil {
			metrics := &MetricsExporterModel{}
			if instance.Features.MetricsExporter.Prometheus != nil {
				prom := &PrometheusExporterModel{
					Enabled:       types.BoolValue(true),
					AuthType:      types.StringNull(),
					EndPoint:      types.StringNull(),
					PrometheusArn: types.StringNull(),
					Username:      types.StringNull(),
					Password:      types.StringNull(),
					Token:         types.StringNull(),
					Labels:        types.MapNull(types.StringType),
				}
				if instance.Features.MetricsExporter.Prometheus.AuthType != nil {
					prom.AuthType = types.StringValue(*instance.Features.MetricsExporter.Prometheus.AuthType)
				}
				if instance.Features.MetricsExporter.Prometheus.EndPoint != nil {
					prom.EndPoint = types.StringValue(*instance.Features.MetricsExporter.Prometheus.EndPoint)
				}
				if instance.Features.MetricsExporter.Prometheus.PrometheusArn != nil {
					prom.PrometheusArn = types.StringValue(*instance.Features.MetricsExporter.Prometheus.PrometheusArn)
				}
				if instance.Features.MetricsExporter.Prometheus.Username != nil {
					prom.Username = types.StringValue(*instance.Features.MetricsExporter.Prometheus.Username)
				}
				if len(instance.Features.MetricsExporter.Prometheus.Labels) > 0 {
					labelMap := make(map[string]string)
					for _, label := range instance.Features.MetricsExporter.Prometheus.Labels {
						if label.Name != nil && label.Value != nil {
							labelMap[*label.Name] = *label.Value
						}
					}
					if len(labelMap) > 0 {
						labelsValue, mapDiags := types.MapValueFrom(context.Background(), types.StringType, labelMap)
						if mapDiags.HasError() {
							diags.Append(mapDiags...)
						} else {
							prom.Labels = labelsValue
						}
					}
				}
				metrics.Prometheus = prom
			}
			if instance.Features.MetricsExporter.CloudWatch != nil {
				cw := &CloudWatchExporterModel{
					Enabled:   types.BoolValue(true),
					Namespace: types.StringNull(),
				}
				if instance.Features.MetricsExporter.CloudWatch.Namespace != nil {
					cw.Namespace = types.StringValue(*instance.Features.MetricsExporter.CloudWatch.Namespace)
				}
				metrics.CloudWatch = cw
			}
			if instance.Features.MetricsExporter.Kafka != nil {
				kafka := &KafkaMetricsExporterModel{
					Enabled:          types.BoolValue(instance.Features.MetricsExporter.Kafka.Enabled),
					BootstrapServers: types.StringNull(),
					Topic:            types.StringNull(),
					CollectionPeriod: types.Int64Null(),
					SecurityProtocol: types.StringNull(),
					SaslMechanism:    types.StringNull(),
					SaslUsername:     types.StringNull(),
					SaslPassword:     types.StringNull(),
				}
				if instance.Features.MetricsExporter.Kafka.BootstrapServers != nil {
					kafka.BootstrapServers = types.StringValue(*instance.Features.MetricsExporter.Kafka.BootstrapServers)
				}
				if instance.Features.MetricsExporter.Kafka.Topic != nil {
					kafka.Topic = types.StringValue(*instance.Features.MetricsExporter.Kafka.Topic)
				}
				if instance.Features.MetricsExporter.Kafka.CollectionPeriod != nil {
					kafka.CollectionPeriod = types.Int64Value(int64(*instance.Features.MetricsExporter.Kafka.CollectionPeriod))
				}
				if instance.Features.MetricsExporter.Kafka.SecurityProtocol != nil {
					kafka.SecurityProtocol = types.StringValue(*instance.Features.MetricsExporter.Kafka.SecurityProtocol)
				}
				if instance.Features.MetricsExporter.Kafka.SaslMechanism != nil {
					kafka.SaslMechanism = types.StringValue(*instance.Features.MetricsExporter.Kafka.SaslMechanism)
				}
				if instance.Features.MetricsExporter.Kafka.SaslUsername != nil {
					kafka.SaslUsername = types.StringValue(*instance.Features.MetricsExporter.Kafka.SaslUsername)
				}
				if instance.Features.MetricsExporter.Kafka.SaslPassword != nil {
					kafka.SaslPassword = types.StringValue(*instance.Features.MetricsExporter.Kafka.SaslPassword)
				}
				metrics.Kafka = kafka
			}
			resource.Features.MetricsExporter = metrics
		} else {
			resource.Features.MetricsExporter = nil
		}

		// Table topic
		if instance.Features.TableTopic != nil {
			topic := &TableTopicModel{
				Warehouse:         types.StringNull(),
				CatalogType:       types.StringNull(),
				MetastoreURI:      types.StringNull(),
				HiveAuthMode:      types.StringNull(),
				KerberosPrincipal: types.StringNull(),
				UserPrincipal:     types.StringNull(),
				KeytabFile:        types.StringNull(),
				Krb5ConfFile:      types.StringNull(),
			}
			if instance.Features.TableTopic.Warehouse != nil {
				topic.Warehouse = types.StringValue(*instance.Features.TableTopic.Warehouse)
			}
			if instance.Features.TableTopic.CatalogType != nil {
				topic.CatalogType = types.StringValue(*instance.Features.TableTopic.CatalogType)
			}
			if instance.Features.TableTopic.MetastoreUri != nil {
				topic.MetastoreURI = types.StringValue(*instance.Features.TableTopic.MetastoreUri)
			}
			if instance.Features.TableTopic.HiveAuthMode != nil {
				topic.HiveAuthMode = types.StringValue(*instance.Features.TableTopic.HiveAuthMode)
			}
			if instance.Features.TableTopic.KerberosPrincipal != nil {
				topic.KerberosPrincipal = types.StringValue(*instance.Features.TableTopic.KerberosPrincipal)
			}
			if instance.Features.TableTopic.UserPrincipal != nil {
				topic.UserPrincipal = types.StringValue(*instance.Features.TableTopic.UserPrincipal)
			}
			if instance.Features.TableTopic.KeytabFile != nil {
				topic.KeytabFile = types.StringValue(*instance.Features.TableTopic.KeytabFile)
			}
			if instance.Features.TableTopic.Krb5ConfFile != nil {
				topic.Krb5ConfFile = types.StringValue(*instance.Features.TableTopic.Krb5ConfFile)
			}
			resource.Features.TableTopic = topic
		} else {
			resource.Features.TableTopic = nil
		}

		// Extended listeners
		if instance.Features.ExtendListeners != nil {
			listeners := make([]InstanceListenerModel, 0, len(instance.Features.ExtendListeners))
			for _, listener := range instance.Features.ExtendListeners {
				model := InstanceListenerModel{
					ListenerName:     types.StringNull(),
					SecurityProtocol: types.StringNull(),
					Port:             types.Int64Null(),
				}
				if listener.ListenerName != nil {
					model.ListenerName = types.StringValue(*listener.ListenerName)
				}
				if listener.SecurityProtocol != nil {
					model.SecurityProtocol = types.StringValue(*listener.SecurityProtocol)
				}
				if listener.Port != nil {
					model.Port = types.Int64Value(int64(*listener.Port))
				}
				listeners = append(listeners, model)
			}
			resource.Features.ExtendListeners = listeners
		} else {
			resource.Features.ExtendListeners = nil
		}

		if instance.Features.InboundRules != nil {
			rules := make([]InboundRuleModel, 0, len(instance.Features.InboundRules))
			for _, rule := range instance.Features.InboundRules {
				model := InboundRuleModel{
					ListenerName: types.StringNull(),
					Cidrs:        types.ListNull(types.StringType),
				}
				if rule.ListenerName != nil {
					model.ListenerName = types.StringValue(*rule.ListenerName)
				}
				if len(rule.Cidrs) > 0 {
					cidrValues := make([]attr.Value, len(rule.Cidrs))
					for i, cidr := range rule.Cidrs {
						cidrValues[i] = types.StringValue(cidr)
					}
					cidrList, listDiags := types.ListValue(types.StringType, cidrValues)
					if listDiags.HasError() {
						diags.Append(listDiags...)
					} else {
						model.Cidrs = cidrList
					}
				}
				rules = append(rules, model)
			}
			resource.Features.InboundRules = rules
		} else {
			resource.Features.InboundRules = nil
		}
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
