package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaInstanceResource{}
var _ resource.ResourceWithConfigure = &KafkaInstanceResource{}
var _ resource.ResourceWithImportState = &KafkaInstanceResource{}
var _ resource.ResourceWithValidateConfig = &KafkaInstanceResource{}

func NewKafkaInstanceResource() resource.Resource {
	r := &KafkaInstanceResource{}
	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultUpdateTimeout(90 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)
	return r
}

// KafkaInstanceResource defines the resource implementation.
type KafkaInstanceResource struct {
	client *client.Client
	api    kafkaInstanceAPI
	framework.WithTimeouts
}

type kafkaInstanceAPI interface {
	CreateKafkaInstance(ctx context.Context, param client.InstanceCreateParam) (*client.InstanceSummaryVO, error)
	GetKafkaInstance(ctx context.Context, instanceId string) (*client.InstanceVO, error)
	GetKafkaInstanceByName(ctx context.Context, name string) (*client.InstanceVO, error)
	DeleteKafkaInstance(ctx context.Context, instanceId string) error
	UpdateKafkaInstance(ctx context.Context, instanceId string, param client.InstanceUpdateParam) error
	GetInstanceEndpoints(ctx context.Context, instanceId string) ([]client.InstanceAccessInfoVO, error)
}

type defaultKafkaInstanceAPI struct {
	client *client.Client
}

var allowedPrometheusAuthTypes = []string{"noauth", "basic", "bearer", "sigv4"}

func (a defaultKafkaInstanceAPI) CreateKafkaInstance(ctx context.Context, param client.InstanceCreateParam) (*client.InstanceSummaryVO, error) {
	return a.client.CreateKafkaInstance(ctx, param)
}

func (a defaultKafkaInstanceAPI) GetKafkaInstance(ctx context.Context, instanceId string) (*client.InstanceVO, error) {
	return a.client.GetKafkaInstance(ctx, instanceId)
}

func (a defaultKafkaInstanceAPI) GetKafkaInstanceByName(ctx context.Context, name string) (*client.InstanceVO, error) {
	return a.client.GetKafkaInstanceByName(ctx, name)
}

func (a defaultKafkaInstanceAPI) DeleteKafkaInstance(ctx context.Context, instanceId string) error {
	return a.client.DeleteKafkaInstance(ctx, instanceId)
}

func (a defaultKafkaInstanceAPI) UpdateKafkaInstance(ctx context.Context, instanceId string, param client.InstanceUpdateParam) error {
	return a.client.UpdateKafkaInstance(ctx, instanceId, param)
}

func (a defaultKafkaInstanceAPI) GetInstanceEndpoints(ctx context.Context, instanceId string) ([]client.InstanceAccessInfoVO, error) {
	return a.client.GetInstanceEndpoints(ctx, instanceId)
}

func (r *KafkaInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_instance"
}

func (r *KafkaInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)\n\n" +
			"Using the `automq_kafka_instance` resource type, you can create and manage Kafka instances, where each instance represents a physical cluster.\n\n" +
			"> **Note**: This provider version is only compatible with AutoMQ control plane versions 8.0 and later.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment identifier (e.g. `env-xxxxx`). Find this on the AutoMQ console System Settings page.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Kafka instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Kafka instance. It can contain letters (a-z or A-Z), numbers (0-9), underscores (_), and hyphens (-), with a length limit of 3 to 64 characters.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 64),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The instance description is used to differentiate the purpose of the instance. It supports letters (a-z or A-Z), numbers (0-9), underscores (_), spaces( ) and hyphens (-), with a length limit of 3 to 256 characters.",
				Optional:            true,
			},
			"version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The software version of AutoMQ instance. If you need to specify a version, refer to the [documentation](https://docs.automq.com/automq-cloud/release-notes) to choose the appropriate version number.",
			},
			"tags": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "A map of tags to assign to the Kafka instance. Tags are key-value pairs that help you identify and organize your resources. Once set, tags cannot be modified.",
				PlanModifiers:       []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"compute_specs": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "The compute specs of the instance",
				Attributes: map[string]schema.Attribute{
					"reserved_aku": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "AKU (AutoMQ Kafka Unit) defines the cluster scale. Each AKU provides up to 30 MiB/s write or 60 MiB/s read throughput. Minimum value is 3; maximum depends on your license quota. Required when `pricing_mode` is `SubscriptionBased`. For sizing guidance, refer to the [billing documentation](https://docs.automq.com/automq-cloud/subscriptions-and-billings/byoc-env-billings/billing-instructions-for-byoc#indicator-constraints).",
						Validators: []validator.Int64{
							int64validator.Between(3, 500),
						},
					},
					"pricing_mode": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Pricing mode for the instance. Supported values: `UsageBased` (pay-as-you-go based on actual usage, requires `reserved_node_count`), `SubscriptionBased` (subscription-based pricing, requires `reserved_aku`). Defaults to `SubscriptionBased`. Changes to pricing mode require instance replacement.",
						Default:             stringdefault.StaticString("SubscriptionBased"),
						Validators: []validator.String{
							stringvalidator.OneOf("UsageBased", "SubscriptionBased"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"reserved_node_count": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Number of reserved nodes for the instance. Valid range is 3 to 100. Required when `pricing_mode` is `UsageBased`.",
						Validators: []validator.Int64{
							int64validator.Between(3, 100),
						},
					},
					"instance_types": schema.ListAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: "Instance type list for the nodes. Maximum 1 entry. Required when `pricing_mode` is `UsageBased` and `deploy_type` is `IAAS`. Cannot be modified after creation.",
						Validators: []validator.List{
							listvalidator.SizeAtMost(1),
							listvalidator.SizeAtLeast(1),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
					"deploy_type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Deployment platform for the instance. `IAAS` deploys on EC2/VM instances; `K8S` deploys on a managed Kubernetes cluster (EKS/GKE/AKS). Supported values: `IAAS`, `K8S`. Changing deployment type requires instance replacement.",
						Default:             stringdefault.StaticString("IAAS"),
						Validators: []validator.String{
							stringvalidator.OneOf("IAAS", "K8S"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"dns_zone": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "DNS zone used when creating custom records. Changing a configured DNS zone requires instance replacement.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"networks": schema.ListNestedAttribute{
						Required:            true,
						MarkdownDescription: "To configure the network settings for an instance, you need to specify the availability zone(s) and subnet information. Currently, you can set either one availability zone or three availability zones.",
						Validators: []validator.List{
							listvalidator.UniqueValues(),
							listvalidator.SizeAtMost(3),
							listvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"zone": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Cloud provider availability zone ID (e.g. `us-east-1a` for AWS). Must match the zone of the specified subnet.",
									PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
								},
								"subnets": schema.ListAttribute{
									Optional:            true,
									MarkdownDescription: "Specify the subnet under the corresponding availability zone for deploying the instance. Currently, only one subnet can be set for each availability zone.",
									ElementType:         types.StringType,
									Validators: []validator.List{
										listvalidator.UniqueValues(),
										listvalidator.SizeAtMost(1),
									},
									PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
								},
							},
						},
					},
					"kubernetes_node_groups": schema.ListNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Node groups (or node pools) are units for unified configuration management of physical nodes in Kubernetes. Different Kubernetes providers may use different terms for node groups. Select target node groups that must be created in advance and configured for either single-AZ or three-AZ deployment. The instance node type must meet the requirements specified in the documentation. If you select a single-AZ node group, the AutoMQ instance will be deployed in a single availability zone; if you select a three-AZ node group, the instance will be deployed across three availability zones.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Node group identifier",
								},
							},
						},
						PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
					},
					"data_buckets": schema.ListNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Inline bucket configuration replacing legacy bucket profiles. Omit this field to let backend manage the data bucket. Changing configured data bucket settings requires instance replacement.",
						CustomType:          types.ListType{ElemType: models.DataBucketObjectType},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"bucket_name": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									MarkdownDescription: "Object storage bucket name used for data.",
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtMost(1),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplaceIfConfigured(),
							listplanmodifier.UseStateForUnknown(),
						},
					},
					"kubernetes_cluster_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Identifier for the target Kubernetes cluster when `deploy_type` is `K8S`. Changing the Kubernetes cluster requires instance replacement.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"kubernetes_namespace": schema.StringAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "Kubernetes namespace for the instance deployment. If not specified, the backend will auto-assign one. Changing a configured namespace requires instance replacement.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"kubernetes_service_account": schema.StringAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "Kubernetes service account for the instance pods. If not specified, the backend will auto-assign one. Changing a configured service account requires instance replacement.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"instance_role": schema.StringAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "IAM role ARN for the Kafka instance. If not specified, the backend will auto-generate an appropriate role. Format: `arn:aws:iam::<account-id>:role/<role-name>`. Changing a configured instance role requires instance replacement.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"security_groups": schema.ListAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Security groups for the instance. Omit this field entirely to let backend auto-generate. If specified, must contain at least one security group. Changing configured security groups requires instance replacement.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplaceIfConfigured(),
							listplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"file_system_param": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "File system configuration for FSWAL mode",
						Attributes: map[string]schema.Attribute{
							"file_system_type": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "File system type. Supported values:\n\n* `EFS_PROVISIONED` - Amazon Elastic File System (EFS), require control panel version ≥ 8.2.0\n* `ONTAP_V2` - Amazon FSx for NetApp ONTAP\n\nChanging this field requires resource replacement.",
								Validators: []validator.String{
									stringvalidator.OneOf("EFS_PROVISIONED", "ONTAP_V2"),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"throughput_mibps_per_file_system": schema.Int64Attribute{
								Required:            true,
								MarkdownDescription: "Throughput in MiBps per file system",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
							"file_system_count": schema.Int64Attribute{
								Required:            true,
								MarkdownDescription: "Number of file systems",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
							"security_groups": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Security groups for file systems. Omit this field entirely to let backend auto-generate. If specified, must contain at least one security group. Changing configured security groups requires instance replacement.",
								PlanModifiers: []planmodifier.List{
									listplanmodifier.RequiresReplaceIfConfigured(),
									listplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
						},
					},
				},
			},
			"features": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Feature configuration for the Kafka instance including WAL mode, security, metrics, and table topics.",
				Attributes: map[string]schema.Attribute{
					"wal_mode": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Write-Ahead Log storage mode: `EBSWAL`, `S3WAL`, or `FSWAL`. `FSWAL` requires `file_system_param` and is not supported with `K8S` deploy type. See [WAL mode documentation](https://docs.automq.com/automq-cloud/manage-instances/create-instance/choose-wal-mode) for details.",
						Validators: []validator.String{
							stringvalidator.OneOf("EBSWAL", "S3WAL", "FSWAL"),
						},
						// Default:       stringdefault.StaticString("EBSWAL"),
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"instance_configs": schema.MapAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "Additional configuration for the Kafka Instance. The currently supported parameters can be set by referring to the [documentation](https://docs.automq.com/automq-cloud/using-automq-for-kafka/restrictions#instance-level-configuration).",
						Optional:            true,
					},
					"security": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"authentication_methods": schema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
								MarkdownDescription: "Configure client authentication methods. Supported values:\n\n" +
									"* `anonymous` - No authentication required. Only available in VPC networks\n" +
									"* `sasl` - SASL protocol authentication. Supports PLAIN and SCRAM mechanisms\n" +
									"* `mtls` - Mutual TLS authentication. Each client uses unique TLS certificates mapped to ACL identities. Automatically supported when TLS encryption is enabled\n\n" +
									"Changes to authentication methods require instance replacement.",
								Validators: []validator.Set{
									setvalidator.ValueStringsAre(
										stringvalidator.OneOf("anonymous", "sasl", "mtls"),
									),
								},
								PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
							},
							"transit_encryption_modes": schema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
								MarkdownDescription: "Configure data transmission encryption. Supported values:\n\n	" +
									"* `plaintext` - No encryption. Only supported in VPC networks. Compatible with PLAINTEXT and SASL authentication protocols\n	" +
									"* `tls` - TLS encrypted transmission. Requires trusted CA certificates and server certificates\n\n" +
									"Changes to encryption modes require instance replacement.",
								Validators: []validator.Set{
									setvalidator.ValueStringsAre(
										stringvalidator.OneOf("plaintext", "tls"),
									),
								},
								PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
							},
							"data_encryption_mode": schema.StringAttribute{
								Optional: true,
								Computed: true,
								MarkdownDescription: "The encryption mode used to protect data stored in AutoMQ using cloud provider's storage encryption capabilities. Supported values:\n\n	" +
									"* `NONE` - No encryption (default)\n	" +
									"* `CPMK` - Cloud Provider Managed Key encryption using cloud provider's KMS service\n\n" +
									"Changes to encryption mode require instance replacement.",
								Default: stringdefault.StaticString("NONE"),
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "CPMK"),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
									stringplanmodifier.RequiresReplace(),
								},
							},
							"tls_hostname_validation_enabled": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Enable TLS hostname validation when AutoMQ brokers terminate TLS. Defaults to true. Changing this setting requires recreating the instance.",
								Default:             booldefault.StaticBool(true),
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
									boolplanmodifier.RequiresReplace(),
								},
							},
							"certificate_authority": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "The trusted CA certificate chain in PEM format used by AutoMQ to verify the validity of both server and client certificates. Required when `mtls` authentication method is enabled.",
							},
							"certificate_chain": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "The server certificate chain in PEM format issued by the CA. AutoMQ will deploy the instance with this certificate. Required when `mtls` authentication method is enabled.",
							},
							"private_key": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "The private key in PEM format corresponding to the server certificate. AutoMQ will deploy the instance with this key. Required when `mtls` authentication method is enabled.",
							},
						},
					},
					"metrics_exporter": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Configure Prometheus Remote Write metrics exporter.",
						Attributes: map[string]schema.Attribute{
							"prometheus": schema.SingleNestedAttribute{
								Optional:            true,
								MarkdownDescription: "Prometheus Remote Write configuration for exporting metrics.",
								Attributes: map[string]schema.Attribute{
									"auth_type": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Authentication type for Prometheus Remote Write. Supported values: `noauth` (no authentication), `basic` (HTTP Basic authentication), `bearer` (Bearer token authentication), `sigv4` (AWS Signature Version 4 for AWS Managed Service for Prometheus).",
										Validators: []validator.String{
											stringvalidator.OneOf(allowedPrometheusAuthTypes...),
										},
									},
									"endpoint": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Prometheus Remote Write endpoint URL. Must be a valid HTTP/HTTPS URL.",
									},
									"prometheus_arn": schema.StringAttribute{
										Optional:            true,
										MarkdownDescription: "AWS Managed Service for Prometheus workspace ARN. Required when `auth_type` is `sigv4`. Format: `arn:aws:aps:<region>:<account-id>:workspace/<workspace-id>`. Ensure the workspace region matches the instance deployment region.",
									},
									"username": schema.StringAttribute{
										Optional:            true,
										MarkdownDescription: "Username for HTTP Basic authentication. Required when `auth_type` is `basic`.",
									},
									"password": schema.StringAttribute{
										Optional:            true,
										Sensitive:           true,
										MarkdownDescription: "Password for HTTP Basic authentication. Required when `auth_type` is `basic`.",
									},
									"token": schema.StringAttribute{
										Optional:            true,
										Sensitive:           true,
										MarkdownDescription: "Bearer token for authentication. Required when `auth_type` is `bearer`.",
									},
									"labels": schema.MapAttribute{
										Optional:            true,
										MarkdownDescription: "Custom labels to attach to exported metrics as key-value pairs.",
										ElementType:         types.StringType,
									},
								},
							},
						},
					},
					"table_topic": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Inline table topic (Iceberg/Hive) configuration replacing legacy integration references.",
						Attributes: map[string]schema.Attribute{
							"warehouse": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Warehouse location for table data. For S3 Table catalog, provide the S3 Table Bucket ARN.",
							},
							"catalog_type": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Catalog type for managing Iceberg tables. Supported values: `s3Table` (S3 Table Catalog), `glue` (Glue Catalog), `hive` (Hive Catalog).",
								Validators: []validator.String{
									stringvalidator.OneOf("s3Table", "glue", "hive"),
								},
							},
							"metastore_uri": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Hive Metastore endpoint. Required when `catalog_type` is `hive`. Format: `thrift://<host>:<port>`.",
							},
							"hive_auth_mode": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Authentication mode for Hive Metastore. Supported values: `NONE` (no authentication), `KERBEROS` (Kerberos authentication).",
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "KERBEROS"),
								},
							},
							"kerberos_principal": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Kerberos principal configured on the Hive Metastore server. Required when `hive_auth_mode` is `KERBEROS`.",
							},
							"user_principal": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Kerberos user principal for authentication, e.g. `username@REALM`. Required when `hive_auth_mode` is `KERBEROS`.",
							},
							"keytab_file": schema.StringAttribute{
								Optional:            true,
								Sensitive:           true,
								MarkdownDescription: "Base64-encoded Kerberos keytab file content. Required when `hive_auth_mode` is `KERBEROS`.",
							},
							"krb5conf_file": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Base64-encoded Kerberos krb5.conf file content. Required when `hive_auth_mode` is `KERBEROS`.",
							},
						},
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				Computed:            true,
				MarkdownDescription: "Timestamp when the instance was created (RFC3339 format).",
			},
			"last_updated": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				Computed:            true,
				MarkdownDescription: "Timestamp when the instance was last updated (RFC3339 format).",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of instance. Currently supports statuses: `Creating`, `Running`, `Deleting`, `Changing` and `Abnormal`. For definitions and limitations of each status, please refer to the [documentation](https://docs.automq.com/automq-cloud/using-automq-for-kafka/manage-instances#lifecycle).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoints": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The bootstrap endpoints of instance. AutoMQ supports multiple access protocols; therefore, the Endpoint is a list.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of endpoint",
						},
						"network_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The network type of endpoint. Currently support `VPC` and `INTERNET`. `VPC` type is generally used for internal network access, while `INTERNET` type is used for accessing the AutoMQ cluster from the internet.",
						},
						"protocol": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The protocol of endpoint. Currently support `PLAINTEXT` and `SASL_PLAINTEXT`.",
						},
						"mechanisms": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The supported mechanisms of endpoint. Currently support `PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512`.",
						},
						"bootstrap_servers": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The bootstrap servers of endpoint.",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *KafkaInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
	r.api = defaultKafkaInstanceAPI{client: client}
}

func isStringValueSet(attr types.String) bool {
	return !attr.IsNull() && !attr.IsUnknown() && strings.TrimSpace(attr.ValueString()) != ""
}

func knownStringValue(attr types.String) (string, bool) {
	if attr.IsNull() || attr.IsUnknown() {
		return "", false
	}
	value := strings.TrimSpace(attr.ValueString())
	if value == "" {
		return "", false
	}
	return value, true
}

func knownInt64Value(attr types.Int64) (int64, bool) {
	if attr.IsNull() || attr.IsUnknown() {
		return 0, false
	}
	return attr.ValueInt64(), true
}

func validateInstanceContract(ctx context.Context, plan *models.KafkaInstanceResourceModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if plan == nil {
		return diagnostics
	}

	diagnostics.Append(validateDeployTypeContract(plan)...)
	diagnostics.Append(validatePricingModeContract(plan)...)
	diagnostics.Append(validateWalModeContract(plan)...)
	diagnostics.Append(validateManagedResourceContract(ctx, plan)...)
	diagnostics.Append(validateFeatureContract(plan)...)

	return diagnostics
}

func validateDeployTypeContract(plan *models.KafkaInstanceResourceModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if plan == nil || plan.ComputeSpecs == nil {
		return diagnostics
	}
	if deployType, ok := knownStringValue(plan.ComputeSpecs.DeployType); ok && strings.EqualFold(deployType, "K8S") {
		nodeGroups := plan.ComputeSpecs.KubernetesNodeGroups
		if len(nodeGroups) == 0 {
			diagnostics.AddError(
				"Invalid Configuration",
				"When compute_specs.deploy_type is K8S, at least one compute_specs.kubernetes_node_groups block must be provided.",
			)
		} else {
			for i, ng := range nodeGroups {
				if !isStringValueSet(ng.ID) {
					diagnostics.AddError(
						"Invalid Configuration",
						fmt.Sprintf("compute_specs.kubernetes_node_groups[%d].id must be provided when deploy_type is K8S.", i),
					)
				}
			}
		}

		if _, ok := knownStringValue(plan.ComputeSpecs.KubernetesClusterID); !ok {
			diagnostics.AddError(
				"Invalid Configuration",
				"When compute_specs.deploy_type is K8S, compute_specs.kubernetes_cluster_id must be provided.",
			)
		}
	}
	return diagnostics
}

func validatePricingModeContract(plan *models.KafkaInstanceResourceModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if plan == nil || plan.ComputeSpecs == nil {
		return diagnostics
	}
	if pricingMode, ok := knownStringValue(plan.ComputeSpecs.PricingMode); ok {
		if strings.EqualFold(pricingMode, "UsageBased") {
			if plan.ComputeSpecs.ReservedNodeCount.IsNull() || plan.ComputeSpecs.ReservedNodeCount.IsUnknown() {
				diagnostics.AddError(
					"Invalid Configuration",
					"compute_specs.reserved_node_count is required when compute_specs.pricing_mode is UsageBased.",
				)
			}
			if deployType, deployTypeSet := knownStringValue(plan.ComputeSpecs.DeployType); deployTypeSet && strings.EqualFold(deployType, "IAAS") {
				if plan.ComputeSpecs.InstanceTypes.IsNull() || plan.ComputeSpecs.InstanceTypes.IsUnknown() {
					diagnostics.AddError(
						"Invalid Configuration",
						"compute_specs.instance_types is required when compute_specs.pricing_mode is UsageBased and compute_specs.deploy_type is IAAS.",
					)
				}
			}
		} else if strings.EqualFold(pricingMode, "SubscriptionBased") {
			if plan.ComputeSpecs.ReservedAku.IsNull() || plan.ComputeSpecs.ReservedAku.IsUnknown() {
				diagnostics.AddError(
					"Invalid Configuration",
					"compute_specs.reserved_aku is required when compute_specs.pricing_mode is SubscriptionBased.",
				)
			}
		}
	}
	return diagnostics
}

func validateWalModeContract(plan *models.KafkaInstanceResourceModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if plan == nil || plan.ComputeSpecs == nil {
		return diagnostics
	}
	hasFileSystemParam := plan.ComputeSpecs.FileSystemParam != nil
	if plan.Features != nil {
		walMode, walModeSet := knownStringValue(plan.Features.WalMode)
		if walModeSet && strings.EqualFold(walMode, "FSWAL") {
			if !hasFileSystemParam {
				diagnostics.AddError(
					"Invalid Configuration",
					"file_system_param configuration is required when wal_mode is FSWAL",
				)
			} else {
				if plan.ComputeSpecs.FileSystemParam.FileSystemType.IsNull() ||
					plan.ComputeSpecs.FileSystemParam.FileSystemType.IsUnknown() {
					diagnostics.AddError(
						"Invalid Configuration",
						"file_system_type is required when wal_mode is FSWAL",
					)
				}
				if plan.ComputeSpecs.FileSystemParam.ThroughputMibpsPerFileSystem.IsNull() ||
					plan.ComputeSpecs.FileSystemParam.ThroughputMibpsPerFileSystem.IsUnknown() {
					diagnostics.AddError(
						"Invalid Configuration",
						"throughput_mibps_per_file_system is required when wal_mode is FSWAL",
					)
				}
				if plan.ComputeSpecs.FileSystemParam.FileSystemCount.IsNull() ||
					plan.ComputeSpecs.FileSystemParam.FileSystemCount.IsUnknown() {
					diagnostics.AddError(
						"Invalid Configuration",
						"file_system_count is required when wal_mode is FSWAL",
					)
				}
			}

			if deployType, ok := knownStringValue(plan.ComputeSpecs.DeployType); ok && strings.EqualFold(deployType, "K8S") {
				diagnostics.AddError(
					"Invalid Configuration",
					"FSWAL is not supported with K8S deployment type",
				)
			}
		} else if hasFileSystemParam {
			diagnostics.AddError(
				"Invalid Configuration",
				"file_system_param configuration is only valid when wal_mode is FSWAL",
			)
		}
	} else if hasFileSystemParam {
		diagnostics.AddError(
			"Invalid Configuration",
			"file_system_param configuration is only valid when wal_mode is FSWAL",
		)
	}
	return diagnostics
}

func validateManagedResourceContract(ctx context.Context, plan *models.KafkaInstanceResourceModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if plan == nil || plan.ComputeSpecs == nil {
		return diagnostics
	}
	if !plan.ComputeSpecs.DataBuckets.IsNull() && !plan.ComputeSpecs.DataBuckets.IsUnknown() {
		planBuckets, bucketDiags := models.DataBucketListToModels(ctx, plan.ComputeSpecs.DataBuckets)
		if bucketDiags.HasError() {
			diagnostics.Append(bucketDiags...)
		}
		for i, bucket := range planBuckets {
			if !isStringValueSet(bucket.BucketName) {
				diagnostics.AddError(
					"Invalid Configuration",
					fmt.Sprintf("compute_specs.data_buckets[%d].bucket_name must be provided when data_buckets is configured.", i),
				)
			}
		}
	}

	return diagnostics
}

func validateFeatureContract(plan *models.KafkaInstanceResourceModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if plan == nil || plan.Features == nil {
		return diagnostics
	}
	if plan.Features.MetricsExporter != nil {
		diagnostics.Append(validateMetricsExporterContract(plan.Features.MetricsExporter)...)
	}
	return diagnostics
}

func validateMetricsExporterContract(metrics *models.MetricsExporterModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if metrics == nil {
		return diagnostics
	}
	if metrics.Prometheus == nil || !models.PrometheusExporterHasConfig(metrics.Prometheus) {
		diagnostics.AddError(
			"Invalid Configuration",
			"features.metrics_exporter must include a prometheus block with required attributes. Remove the metrics_exporter block entirely to disable metrics export.",
		)
		return diagnostics
	}
	return diagnostics
}

func (r *KafkaInstanceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config models.KafkaInstanceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateInstanceContract(ctx, &config)...)
}

func (r *KafkaInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.KafkaInstanceResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateInstanceContract(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	// Generate API request body from plan
	in := client.InstanceCreateParam{}
	if err := models.ExpandKafkaInstanceResource(ctx, plan, &in); err != nil {
		resp.Diagnostics.AddError("Model Expansion Error", fmt.Sprintf("Failed to expand Kafka instance resource: %s", err))
		return
	}
	logFields := map[string]any{
		"environment_id": plan.EnvironmentID.ValueString(),
		"name":           plan.Name.ValueString(),
	}
	if plan.ComputeSpecs != nil {
		if !plan.ComputeSpecs.ReservedAku.IsNull() {
			logFields["reserved_aku"] = plan.ComputeSpecs.ReservedAku.ValueInt64()
		}
	}
	tflog.Debug(ctx, "Creating new Kafka Cluster", logFields)

	out, err := r.api.CreateKafkaInstance(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Kafka instance, got error: %s", err))
		return
	}
	// Start refresh from the original plan so backend-omitted fields still have
	// a state baseline during the post-create readback merge.
	state := plan
	// Flatten API response into Terraform state
	resp.Diagnostics.Append(models.FlattenKafkaInstanceBasicModel(out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Persist the initial state so Terraform is aware of the in-flight resource
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceId := state.InstanceID.ValueString()

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	if err := waitForKafkaClusterToProvisionFunc(ctx, r.client, instanceId, models.StateCreating, createTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}

	_, found, diags := refreshKafkaInstanceState(ctx, r, instanceId, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q not found after creation", instanceId))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.KafkaInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())
	instanceId := state.InstanceID.ValueString()
	_, found, diags := refreshKafkaInstanceState(ctx, r, instanceId, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		// Treat HTTP 404 Not Found status as a signal to recreate resource and return early
		resp.State.RemoveResource(ctx)
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.KafkaInstanceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate plan-only contracts before any backend call. These checks cover
	// required field combinations that Terraform schema cannot express.
	resp.Diagnostics.Append(validateInstanceContract(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	instanceId := plan.InstanceID.ValueString()
	updateTimeout := r.UpdateTimeout(ctx, state.Timeouts)

	// Validate update-only contracts that require comparing the new plan with
	// prior state, such as unsupported removal of instance config keys.
	resp.Diagnostics.Append(validateInstanceUpdateContract(instanceId, plan, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the backend PATCH payload from supported online-update fields only.
	// Unsupported in-place diffs are rejected before backend I/O.
	updateParam, updatePlan := buildInstanceUpdateParam(plan, state)
	if !updatePlan.hasUpdate {
		resp.Diagnostics.AddError(
			"Unsupported Kafka Instance Update",
			fmt.Sprintf("Terraform planned an in-place update for Kafka instance %q, but none of the changed attributes are supported by the backend PATCH API. "+
				"This usually means a create-only or backend-managed attribute is missing a replacement plan modifier.", instanceId),
		)
		return
	}

	// Preserve Terraform-only or sensitive plan values that the read API may omit,
	// so the refresh after PATCH does not erase valid configuration.
	applyUpdateStatePreservation(&state, plan, updatePlan)

	// Check backend runtime state only after a real PATCH is required. Local
	// contract failures and unsupported diffs return before this API call.
	instance, err := r.api.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	if instance == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q not found", instanceId))
		return
	}
	if instance.State == nil || *instance.State != models.StateRunning {
		current := models.StateUnknown
		if instance.State != nil {
			current = *instance.State
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q is Currently in %q state, only instances in 'Running' state can be updated", instanceId, current))
		return
	}

	// Execute the PATCH, wait for asynchronous changes when needed, then refresh
	// from backend so Terraform state reflects server-side readback.
	if err := r.api.UpdateKafkaInstance(ctx, instanceId, updateParam); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	if updatePlan.shouldWait {
		if err := waitForKafkaClusterToProvisionFunc(ctx, r.client, instanceId, models.StateChanging, updateTimeout); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
			return
		}
	}
	_, found, diags := refreshKafkaInstanceState(ctx, r, instanceId, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q not found after update", instanceId))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.KafkaInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	instanceId := state.InstanceID.ValueString()
	instance, err := r.api.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	if instance == nil {
		return
	}

	if *instance.State != models.StateDeleting {
		err = r.api.DeleteKafkaInstance(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Kafka instance %q, got error: %s", instanceId, err))
			return
		}
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	tflog.Info(ctx, "waiting for Kafka instance to be deleted", map[string]any{"instance_id": instanceId, "timeout": deleteTimeout.String()})
	// Wait until control plane reports NotFound so acceptance tests don't leave dangling clusters.
	if err := framework.WaitForKafkaClusterToDeleted(ctx, r.client, instanceId, deleteTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}
	tflog.Info(ctx, "Kafka instance deletion completed", map[string]any{"instance_id": instanceId})
	resp.State.RemoveResource(ctx)
}

func (r *KafkaInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "@")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: <environment_id>@<instance_id>, got: %q", req.ID),
		)
		return
	}
	environmentId := idParts[0]
	instanceId := idParts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), environmentId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), instanceId)...)
}

func shouldRefreshInstanceEndpoints(instance *client.InstanceVO) bool {
	return instance != nil && instance.State != nil && *instance.State == models.StateRunning
}

func refreshKafkaInstanceState(ctx context.Context, r *KafkaInstanceResource, instanceId string, state *models.KafkaInstanceResourceModel) (*client.InstanceVO, bool, diag.Diagnostics) {
	instance, err := r.api.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", instanceId, err))}
	}
	if instance == nil {
		return nil, false, diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Kafka instance %q not found", instanceId))}
	}

	diags := diag.Diagnostics{}
	diags.Append(models.FlattenKafkaInstanceModel(ctx, instance, state)...)
	if diags.HasError() {
		return instance, true, diags
	}
	if shouldRefreshInstanceEndpoints(instance) {
		endpoints, err := r.api.GetInstanceEndpoints(ctx, instanceId)
		if err != nil {
			return instance, true, diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get endpoints for Kafka instance %q, got error: %s", instanceId, err))}
		}
		diags.Append(models.FlattenKafkaInstanceModelWithEndpoints(endpoints, state)...)
	}
	return instance, true, diags
}

type instanceUpdatePlan struct {
	hasUpdate              bool
	shouldWait             bool
	instanceConfigsChanged bool
	certificateChanged     bool
}

func validateInstanceUpdateContract(instanceId string, plan, state models.KafkaInstanceResourceModel) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if plan.Features == nil || state.Features == nil || plan.Features.InstanceConfigs.IsUnknown() || state.Features.InstanceConfigs.IsUnknown() {
		return diags
	}

	planConfig := plan.Features.InstanceConfigs
	stateConfig := state.Features.InstanceConfigs
	if models.MapsEqual(planConfig, stateConfig) || stateConfig.IsNull() {
		return diags
	}

	for name := range stateConfig.Elements() {
		if _, ok := planConfig.Elements()[name]; !ok {
			diags.AddError("Config Update Error", fmt.Sprintf("Error occurred while updating Kafka Instance %q. "+
				" At present, we don't support the removal of instance settings from the 'configs' block, "+
				"meaning you can't reset to the instance's default settings. "+
				"As a workaround, you can find the default value and manually set the current value to match the default.", instanceId))
			return diags
		}
	}

	return diags
}

func applyUpdateStatePreservation(state *models.KafkaInstanceResourceModel, plan models.KafkaInstanceResourceModel, updatePlan instanceUpdatePlan) {
	if state == nil {
		return
	}
	if updatePlan.instanceConfigsChanged && plan.Features != nil {
		if state.Features == nil {
			state.Features = &models.FeaturesModel{}
		}
		state.Features.InstanceConfigs = plan.Features.InstanceConfigs
	}

	if updatePlan.certificateChanged && plan.Features != nil && plan.Features.Security != nil {
		if state.Features == nil {
			state.Features = &models.FeaturesModel{}
		}
		if state.Features.Security == nil {
			state.Features.Security = &models.SecurityModel{}
		}
		state.Features.Security.CertificateAuthority = plan.Features.Security.CertificateAuthority
		state.Features.Security.CertificateChain = plan.Features.Security.CertificateChain
		state.Features.Security.PrivateKey = plan.Features.Security.PrivateKey
	}
}

func buildInstanceUpdateParam(plan, state models.KafkaInstanceResourceModel) (client.InstanceUpdateParam, instanceUpdatePlan) {
	updateParam := client.InstanceUpdateParam{}
	updatePlan := instanceUpdatePlan{}

	ensureSpec := func() *client.SpecificationUpdateParam {
		if updateParam.Spec == nil {
			updateParam.Spec = &client.SpecificationUpdateParam{}
		}
		return updateParam.Spec
	}

	ensureFeatures := func() *client.InstanceFeatureParam {
		if updateParam.Features == nil {
			updateParam.Features = &client.InstanceFeatureParam{}
		}
		return updateParam.Features
	}

	if planName := plan.Name.ValueString(); planName != state.Name.ValueString() {
		name := planName
		updateParam.Name = &name
		updatePlan.hasUpdate = true
	}

	if planDesc := plan.Description.ValueString(); planDesc != state.Description.ValueString() {
		desc := planDesc
		updateParam.Description = &desc
		updatePlan.hasUpdate = true
	}

	if plan.Features != nil && state.Features != nil && !plan.Features.InstanceConfigs.IsUnknown() && !state.Features.InstanceConfigs.IsUnknown() {
		planConfig := plan.Features.InstanceConfigs
		stateConfig := state.Features.InstanceConfigs
		if !models.MapsEqual(planConfig, stateConfig) {
			features := ensureFeatures()
			features.InstanceConfigs = models.ExpandStringValueMap(planConfig)
			updatePlan.hasUpdate = true
			updatePlan.shouldWait = true
			updatePlan.instanceConfigsChanged = true
		}
	}

	if plan.Features != nil && plan.Features.Security != nil && state.Features != nil && state.Features.Security != nil {
		if !plan.Features.Security.CertificateAuthority.Equal(state.Features.Security.CertificateAuthority) ||
			!plan.Features.Security.CertificateChain.Equal(state.Features.Security.CertificateChain) ||
			!plan.Features.Security.PrivateKey.Equal(state.Features.Security.PrivateKey) {
			features := ensureFeatures()
			security := &client.InstanceSecurityParam{}
			ca := plan.Features.Security.CertificateAuthority.ValueString()
			security.CertificateAuthority = &ca
			chain := plan.Features.Security.CertificateChain.ValueString()
			security.CertificateChain = &chain
			privateKey := plan.Features.Security.PrivateKey.ValueString()
			security.PrivateKey = &privateKey
			features.Security = security
			updatePlan.hasUpdate = true
			updatePlan.shouldWait = true
			updatePlan.certificateChanged = true
		}
	}

	planVersion := plan.Version.ValueString()
	if planVersion != "" && planVersion != state.Version.ValueString() {
		version := planVersion
		updateParam.Version = &version
		updatePlan.hasUpdate = true
		updatePlan.shouldWait = true
	}

	if plan.Features != nil {
		var stateMetrics *models.MetricsExporterModel
		if state.Features != nil {
			stateMetrics = state.Features.MetricsExporter
		}
		if metricsExporterChanged(plan.Features.MetricsExporter, stateMetrics) {
			exporter, hasExporter := buildMetricsExporterParam(plan.Features.MetricsExporter)
			features := ensureFeatures()
			if hasExporter {
				features.MetricsExporter = exporter
				updatePlan.hasUpdate = true
				updatePlan.shouldWait = true
			} else if stateMetrics != nil {
				enabled := false
				prom := &client.InstancePrometheusExporterParam{Enabled: &enabled}
				features.MetricsExporter = &client.InstanceMetricsExporterParam{Prometheus: prom}
				updatePlan.hasUpdate = true
				updatePlan.shouldWait = true
			}
		}
	}

	if plan.ComputeSpecs != nil {
		if planAKUValue, ok := knownInt64Value(plan.ComputeSpecs.ReservedAku); ok {
			planAKU := int32(planAKUValue)
			stateAKU := int32(0)
			if state.ComputeSpecs != nil {
				if stateAKUValue, ok := knownInt64Value(state.ComputeSpecs.ReservedAku); ok {
					stateAKU = int32(stateAKUValue)
				}
			}
			if planAKU != stateAKU {
				spec := ensureSpec()
				spec.ReservedAku = &planAKU
				updatePlan.hasUpdate = true
				updatePlan.shouldWait = true
			}
		}
	}

	if plan.ComputeSpecs != nil {
		if planNodeCountValue, ok := knownInt64Value(plan.ComputeSpecs.ReservedNodeCount); ok {
			planNodeCount := int32(planNodeCountValue)
			stateNodeCount := int32(0)
			if state.ComputeSpecs != nil {
				if stateNodeCountValue, ok := knownInt64Value(state.ComputeSpecs.ReservedNodeCount); ok {
					stateNodeCount = int32(stateNodeCountValue)
				}
			}
			if planNodeCount != stateNodeCount {
				spec := ensureSpec()
				spec.ReservedNodeCount = &planNodeCount
				updatePlan.hasUpdate = true
				updatePlan.shouldWait = true
			}
		}
	}

	if plan.ComputeSpecs != nil {
		var stateFileSystemParam *models.FileSystemParamModel
		if state.ComputeSpecs != nil {
			stateFileSystemParam = state.ComputeSpecs.FileSystemParam
		}

		planFileSystemParam := plan.ComputeSpecs.FileSystemParam
		if fileSystemUpdateParamChanged(planFileSystemParam, stateFileSystemParam) {
			if planFileSystemParam != nil {
				fileSystemParam := &client.FileSystemUpdateParam{
					ThroughputMiBpsPerFileSystem: int32(planFileSystemParam.ThroughputMibpsPerFileSystem.ValueInt64()),
					FileSystemCount:              int32(planFileSystemParam.FileSystemCount.ValueInt64()),
				}

				spec := ensureSpec()
				spec.FileSystem = fileSystemParam
				updatePlan.hasUpdate = true
				updatePlan.shouldWait = true
			}
		}
	}

	return updateParam, updatePlan
}

func metricsExporterChanged(plan, state *models.MetricsExporterModel) bool {
	if plan == nil {
		return state != nil
	}
	if state == nil {
		return models.MetricsExporterHasConfig(plan)
	}
	if !prometheusExporterEqual(plan.Prometheus, state.Prometheus) {
		return true
	}
	return false
}

func prometheusExporterEqual(plan, state *models.PrometheusExporterModel) bool {
	if plan == nil || state == nil {
		return plan == nil && state == nil
	}
	if !stringAttrEqual(plan.AuthType, state.AuthType) {
		return false
	}
	if !stringAttrEqual(plan.EndPoint, state.EndPoint) {
		return false
	}
	if !stringAttrEqual(plan.PrometheusArn, state.PrometheusArn) {
		return false
	}
	if !stringAttrEqual(plan.Username, state.Username) {
		return false
	}
	if !stringAttrEqual(plan.Password, state.Password) {
		return false
	}
	if !stringAttrEqual(plan.Token, state.Token) {
		return false
	}
	if !mapAttrEqual(plan.Labels, state.Labels) {
		return false
	}
	return true
}

func buildMetricsExporterParam(model *models.MetricsExporterModel) (*client.InstanceMetricsExporterParam, bool) {
	if model == nil {
		return nil, false
	}
	exporter := client.InstanceMetricsExporterParam{}
	hasConfig := false
	if model.Prometheus != nil {
		prom, ok := buildPrometheusExporterParam(model.Prometheus)
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

func buildPrometheusExporterParam(model *models.PrometheusExporterModel) (*client.InstancePrometheusExporterParam, bool) {
	if model == nil {
		return nil, false
	}
	if !models.PrometheusExporterHasConfig(model) {
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
		labels := models.ExpandStringValueMap(model.Labels)
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

func stringAttrEqual(plan, state types.String) bool {
	if plan.IsUnknown() {
		return true
	}
	if plan.IsNull() {
		return state.IsNull() || state.IsUnknown()
	}
	if state.IsNull() || state.IsUnknown() {
		return false
	}
	return plan.ValueString() == state.ValueString()
}

func mapAttrEqual(plan, state types.Map) bool {
	if plan.IsUnknown() {
		return true
	}
	if plan.IsNull() {
		return state.IsNull() || state.IsUnknown()
	}
	if state.IsNull() || state.IsUnknown() {
		return false
	}
	return plan.Equal(state)
}

func fileSystemUpdateParamChanged(plan, state *models.FileSystemParamModel) bool {
	if plan == nil {
		return state != nil
	}
	if state == nil {
		return plan != nil
	}

	if !int64AttrEqual(plan.ThroughputMibpsPerFileSystem, state.ThroughputMibpsPerFileSystem) {
		return true
	}

	if !int64AttrEqual(plan.FileSystemCount, state.FileSystemCount) {
		return true
	}

	return false
}

func int64AttrEqual(plan, state types.Int64) bool {
	if plan.IsUnknown() {
		return true
	}
	if plan.IsNull() {
		return state.IsNull() || state.IsUnknown()
	}
	if state.IsNull() || state.IsUnknown() {
		return false
	}
	return plan.ValueInt64() == state.ValueInt64()
}

var waitForKafkaClusterToProvisionFunc = framework.WaitForKafkaClusterToProvision
