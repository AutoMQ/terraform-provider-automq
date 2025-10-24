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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	ListInstanceIntegrations(ctx context.Context, instanceId string) ([]client.IntegrationVO, error)
	GetInstanceEndpoints(ctx context.Context, instanceId string) ([]client.InstanceAccessInfoVO, error)
	GetInstanceConfigs(ctx context.Context, instanceId string) ([]client.ConfigItemParam, error)
}

type defaultKafkaInstanceAPI struct {
	client *client.Client
}

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

func (a defaultKafkaInstanceAPI) ListInstanceIntegrations(ctx context.Context, instanceId string) ([]client.IntegrationVO, error) {
	return a.client.ListInstanceIntegrations(ctx, instanceId)
}

func (a defaultKafkaInstanceAPI) GetInstanceEndpoints(ctx context.Context, instanceId string) ([]client.InstanceAccessInfoVO, error) {
	return a.client.GetInstanceEndpoints(ctx, instanceId)
}

func (a defaultKafkaInstanceAPI) GetInstanceConfigs(ctx context.Context, instanceId string) ([]client.ConfigItemParam, error) {
	return a.client.GetInstanceConfigs(ctx, instanceId)
}

func (r *KafkaInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_instance"
}

func (r *KafkaInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)\n\n" +
			"Using the `automq_kafka_instance` resource type, you can create and manage Kafka instances, where each instance represents a physical cluster.\n\n" +
			"> **Note**: This provider version is only compatible with AutoMQ control plane versions 7.3.5 and later.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment, this attribute is specified during the deployment and installation process.",
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
				MarkdownDescription: "The instance description are used to differentiate the purpose of the instance. They support letters (a-z or A-Z), numbers (0-9), underscores (_), spaces( ) and hyphens (-), with a length limit of 3 to 128 characters.",
				Optional:            true,
			},
			"deploy_profile": schema.StringAttribute{
				MarkdownDescription: "(Deprecated) Deploy profile defining cloud resource configuration including VPC, Kubernetes, storage and IAM roles. Inline compute_specs should be used instead.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "deploy_profile is deprecated. Use compute_specs.* fields to describe deployments.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "The software version of AutoMQ instance. If you need to specify a version, refer to the [documentation](https://docs.automq.com/automq-cloud/release-notes) to choose the appropriate version number.",
			},
			"compute_specs": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The compute specs of the instance",
				Attributes: map[string]schema.Attribute{
					"reserved_aku": schema.Int64Attribute{
						Required:    true,
						Description: "AutoMQ defines AKU (AutoMQ Kafka Unit) to measure the scale of the cluster. Each AKU provides 20 MiB/s of read/write throughput. For more details on AKU, please refer to the [documentation](https://docs.automq.com/automq-cloud/subscriptions-and-billings/byoc-env-billings/billing-instructions-for-byoc#indicator-constraints). The currently supported AKU specifications are 6, 8, 10, 12, 14, 16, 18, 20, 22, and 24. If an invalid AKU value is set, the instance cannot be created.",
						Validators: []validator.Int64{
							int64validator.Between(3, 500),
						},
					},
					"deploy_type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Deployment platform for the instance. Supported values: `IAAS`, `K8S`.",
						Validators: []validator.String{
							stringvalidator.OneOf("IAAS", "K8S"),
						},
					},
					"provider": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Cloud provider identifier, e.g. `aws`.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"region": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Region where the instance will be deployed.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"scope": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Cloud provider scope such as account ID or organization.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"vpc": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "VPC identifier for the deployment target.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"dns_zone": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "DNS zone used when creating custom records.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"networks": schema.ListNestedAttribute{
						Optional:    true,
						Description: "To configure the network settings for an instance, you need to specify the availability zone(s) and subnet information. Currently, you can set either one availability zone or three availability zones.",
						Validators: []validator.List{
							listvalidator.UniqueValues(),
							listvalidator.SizeBetween(1, 3),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"zone": schema.StringAttribute{
									Required:      true,
									Description:   "The availability zone ID of the cloud provider.",
									PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
								},
								"subnets": schema.ListAttribute{
									Required:    true,
									Description: "Specify the subnet under the corresponding availability zone for deploying the instance. Currently, only one subnet can be set for each availability zone.",
									ElementType: types.StringType,
									Validators: []validator.List{
										listvalidator.UniqueValues(),
										listvalidator.SizeAtLeast(1),
										listvalidator.SizeAtMost(1),
									},
									PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
								},
							},
						},
					},
					"kubernetes_node_groups": schema.ListNestedAttribute{
						Optional:    true,
						Description: "Node groups (or node pools) are units for unified configuration management of physical nodes in Kubernetes. Different Kubernetes providers may use different terms for node groups. Select target node groups that must be created in advance and configured for either single-AZ or three-AZ deployment. The instance node type must meet the requirements specified in the documentation. If you select a single-AZ node group, the AutoMQ instance will be deployed in a single availability zone; if you select a three-AZ node group, the instance will be deployed across three availability zones.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required:    true,
									Description: "Node group identifier",
								},
							},
						},
						PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
					},
					"bucket_profiles": schema.ListNestedAttribute{
						Optional:           true,
						Description:        "(Deprecated) Bucket profile bindings. Use `data_buckets` instead.",
						DeprecationMessage: "bucket_profiles is deprecated. Please migrate to data_buckets.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required:    true,
									Description: "Bucket profile ID",
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.SizeAtMost(1),
						},
						PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
					},
					"data_buckets": schema.ListNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Inline bucket configuration replacing legacy bucket profiles.",
						CustomType:  types.ListType{ElemType: models.DataBucketObjectType},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"bucket_name": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Object storage bucket name used for data.",
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
							listplanmodifier.UseStateForUnknown(),
						},
					},
					"file_system_param": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "File system configuration required when WAL mode is FSWAL.",
						Attributes: map[string]schema.Attribute{
							"throughput_mibps_per_file_system": schema.Int64Attribute{
								Required:    true,
								Description: "Provisioned throughput in MiB/s for each file system.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
							"file_system_count": schema.Int64Attribute{
								Required:    true,
								Description: "Number of file systems allocated for WAL storage.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
						},
					},
					"kubernetes_cluster_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Identifier for the target Kubernetes cluster when deploy_type is KUBERNETES.",
					},
					"kubernetes_namespace": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"kubernetes_service_account": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"credential": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"instance_role": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"tenant_id": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"vpc_resource_group": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"k8s_resource_group": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"dns_resource_group": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"features": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"wal_mode": schema.StringAttribute{
						Required:    true,
						Description: "Write-Ahead Logging mode: EBSWAL (using EBS as write buffer), S3WAL (using object storage as write buffer), or FSWAL (using file systems as write buffer). Defaults to EBSWAL.",
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
					"integrations": schema.SetAttribute{
						Optional:           true,
						ElementType:        types.StringType,
						Description:        "(Deprecated) Integration identifiers previously used for metrics/table topic bindings.",
						DeprecationMessage: "integrations is deprecated. Configure metrics_exporter or table_topic blocks instead.",
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
						MarkdownDescription: "Configure Prometheus metrics scraping.",
						Attributes: map[string]schema.Attribute{
							"prometheus": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"enabled":        schema.BoolAttribute{Optional: true},
									"auth_type":      schema.StringAttribute{Optional: true},
									"end_point":      schema.StringAttribute{Optional: true},
									"prometheus_arn": schema.StringAttribute{Optional: true},
									"username":       schema.StringAttribute{Optional: true},
									"password":       schema.StringAttribute{Optional: true},
									"token":          schema.StringAttribute{Optional: true},
									"labels": schema.MapAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
							"kafka": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"enabled":           schema.BoolAttribute{Optional: true},
									"bootstrap_servers": schema.StringAttribute{Optional: true},
									"topic":             schema.StringAttribute{Optional: true},
									"collection_period": schema.Int64Attribute{Optional: true},
									"security_protocol": schema.StringAttribute{Optional: true},
									"sasl_mechanism":    schema.StringAttribute{Optional: true},
									"sasl_username":     schema.StringAttribute{Optional: true},
									"sasl_password":     schema.StringAttribute{Optional: true},
								},
							},
						},
					},
					"table_topic": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Inline table topic (Iceberg/Hive) configuration replacing legacy integration references.",
						Attributes: map[string]schema.Attribute{
							"warehouse": schema.StringAttribute{
								Required: true,
							},
							"catalog_type": schema.StringAttribute{
								Required: true,
							},
							"metastore_uri":      schema.StringAttribute{Optional: true},
							"hive_auth_mode":     schema.StringAttribute{Optional: true},
							"kerberos_principal": schema.StringAttribute{Optional: true},
							"user_principal":     schema.StringAttribute{Optional: true},
							"keytab_file":        schema.StringAttribute{Optional: true},
							"krb5conf_file":      schema.StringAttribute{Optional: true},
						},
					},
					"inbound_rules": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"listener_name": schema.StringAttribute{
									Required: true,
								},
								"cidrs": schema.ListAttribute{
									Required:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
					"extend_listeners": schema.ListNestedAttribute{
						Optional: true,
						Validators: []validator.List{
							listvalidator.SizeAtMost(5),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"listener_name": schema.StringAttribute{
									Required: true,
								},
								"security_protocol": schema.StringAttribute{Optional: true},
								"port": schema.Int64Attribute{
									Optional: true,
									Validators: []validator.Int64{
										int64validator.Between(1, 65535),
									},
								},
							},
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"last_updated": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of instance. Currently supports statuses: `Creating`, `Running`, `Deleting`, `Changing` and `Abnormal`. For definitions and limitations of each status, please refer to the [documentation](https://docs.automq.com/automq-cloud/using-automq-for-kafka/manage-instances#lifecycle).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoints": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The bootstrap endpoints of instance. AutoMQ supports multiple access protocols; therefore, the Endpoint is a list.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of endpoint",
						},
						"network_type": schema.StringAttribute{
							Computed:    true,
							Description: "The network type of endpoint. Currently support `VPC` and `INTERNET`. `VPC` type is generally used for internal network access, while `INTERNET` type is used for accessing the AutoMQ cluster from the internet.",
						},
						"protocol": schema.StringAttribute{
							Computed:    true,
							Description: "The protocol of endpoint. Currently support `PLAINTEXT` and `SASL_PLAINTEXT`.",
						},
						"mechanisms": schema.StringAttribute{
							Computed:    true,
							Description: "The supported mechanisms of endpoint. Currently support `PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512`.",
						},
						"bootstrap_servers": schema.StringAttribute{
							Computed:    true,
							Description: "The bootstrap servers of endpoint.",
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

func resolvePlannedStringValue(plan types.String, state *types.String) (string, bool) {
	if !plan.IsNull() && !plan.IsUnknown() {
		value := strings.TrimSpace(plan.ValueString())
		if value == "" {
			return "", false
		}
		return value, true
	}
	if plan.IsUnknown() && state != nil && !state.IsNull() && !state.IsUnknown() {
		value := strings.TrimSpace(state.ValueString())
		if value == "" {
			return "", false
		}
		return value, true
	}
	return "", false
}

func isStringValueSet(attr types.String) bool {
	return !attr.IsNull() && !attr.IsUnknown() && strings.TrimSpace(attr.ValueString()) != ""
}

func validateKafkaInstanceConfiguration(ctx context.Context, plan *models.KafkaInstanceResourceModel, state *models.KafkaInstanceResourceModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	if plan == nil || plan.ComputeSpecs == nil {
		return diagnostics
	}

	var stateSpecs *models.ComputeSpecsModel
	if state != nil {
		stateSpecs = state.ComputeSpecs
	}

	var stateDeploy *types.String
	if stateSpecs != nil {
		stateDeploy = &stateSpecs.DeployType
	}

	if deployType, ok := resolvePlannedStringValue(plan.ComputeSpecs.DeployType, stateDeploy); ok && strings.EqualFold(deployType, "K8S") {
		nodeGroups := plan.ComputeSpecs.KubernetesNodeGroups
		if nodeGroups == nil && stateSpecs != nil {
			nodeGroups = stateSpecs.KubernetesNodeGroups
		}
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

		var stateCluster *types.String
		if stateSpecs != nil {
			stateCluster = &stateSpecs.KubernetesClusterID
		}
		if _, ok := resolvePlannedStringValue(plan.ComputeSpecs.KubernetesClusterID, stateCluster); !ok {
			diagnostics.AddError(
				"Invalid Configuration",
				"When compute_specs.deploy_type is K8S, compute_specs.kubernetes_cluster_id must be provided.",
			)
		}
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

func (r *KafkaInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var instance models.KafkaInstanceResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &instance)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateKafkaInstanceConfiguration(ctx, &instance, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, instance.EnvironmentID.ValueString())

	// network or kubernetes node group must be set
	if len(instance.ComputeSpecs.Networks) == 0 && len(instance.ComputeSpecs.KubernetesNodeGroups) == 0 {
		resp.Diagnostics.AddError("Invalid Configuration", "At least one of the network or kubernetes node group must be set.")
		return
	}

	// Generate API request body from plan
	in := client.InstanceCreateParam{}
	if err := models.ExpandKafkaInstanceResource(instance, &in); err != nil {
		resp.Diagnostics.AddError("Model Expansion Error", fmt.Sprintf("Failed to expand Kafka instance resource: %s", err))
		return
	}
	logFields := map[string]any{
		"environment_id": instance.EnvironmentID.ValueString(),
		"name":           instance.Name.ValueString(),
		"deploy_profile": instance.DeployProfile.ValueString(),
	}
	if instance.ComputeSpecs != nil {
		logFields["reserved_aku"] = instance.ComputeSpecs.ReservedAku.ValueInt64()
	}
	tflog.Debug(ctx, "Creating new Kafka Cluster", logFields)

	out, err := r.api.CreateKafkaInstance(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Kafka instance, got error: %s", err))
		return
	}
	// Flatten API response into Terraform state
	resp.Diagnostics.Append(models.FlattenKafkaInstanceBasicModel(out, &instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceId := instance.InstanceID.ValueString()

	createTimeout := r.CreateTimeout(ctx, instance.Timeouts)
	if err := waitForKafkaClusterToProvisionFunc(ctx, r.client, instanceId, models.StateCreating, createTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}

	resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &instance)...)
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
	instance, err := r.api.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			// Treat HTTP 404 Not Found status as a signal to recreate resource and return early
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModel(instance, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if instance.State != nil && *instance.State == models.StateRunning {
		integrations, err := r.api.ListInstanceIntegrations(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list integrations for Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
			return
		}
		resp.Diagnostics.Append(models.FlattenKafkaInstanceModelWithIntegrations(integrations, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}

		endpoints, err := r.api.GetInstanceEndpoints(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get endpoints for Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
			return
		}
		resp.Diagnostics.Append(models.FlattenKafkaInstanceModelWithEndpoints(endpoints, &state)...)
	}
	if resp.Diagnostics.HasError() {
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
	resp.Diagnostics.Append(validateKafkaInstanceConfiguration(ctx, &plan, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	instanceId := plan.InstanceID.ValueString()
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

	updateTimeout := r.UpdateTimeout(ctx, state.Timeouts)

	updateParam := client.InstanceUpdateParam{}
	shouldWait := false
	hasUpdate := false
	instanceConfigsChanged := false
	certificateChanged := false

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
		hasUpdate = true
	}

	if planDesc := plan.Description.ValueString(); planDesc != state.Description.ValueString() {
		desc := planDesc
		updateParam.Description = &desc
		hasUpdate = true
	}

	if plan.Features != nil && state.Features != nil && !plan.Features.InstanceConfigs.IsUnknown() && !state.Features.InstanceConfigs.IsUnknown() {
		planConfig := plan.Features.InstanceConfigs
		stateConfig := state.Features.InstanceConfigs
		if !models.MapsEqual(planConfig, stateConfig) {
			if !stateConfig.IsNull() {
				for name := range stateConfig.Elements() {
					if _, ok := planConfig.Elements()[name]; !ok {
						resp.Diagnostics.AddError("Config Update Error", fmt.Sprintf("Error occurred while updating Kafka Instance %q. "+
							" At present, we don't support the removal of instance settings from the 'configs' block, "+
							"meaning you can't reset to the instance's default settings. "+
							"As a workaround, you can find the default value and manually set the current value to match the default.", instanceId))
						return
					}
				}
			}
			features := ensureFeatures()
			features.InstanceConfigs = models.ExpandStringValueMap(planConfig)
			hasUpdate = true
			shouldWait = true
			instanceConfigsChanged = true
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
			hasUpdate = true
			shouldWait = true
			certificateChanged = true
		}
	}

	planVersion := plan.Version.ValueString()
	if planVersion != "" && planVersion != state.Version.ValueString() {
		version := planVersion
		updateParam.Version = &version
		hasUpdate = true
		shouldWait = true
	}

	if plan.Features != nil {
		var stateMetrics *models.MetricsExporterModel
		if state.Features != nil {
			stateMetrics = state.Features.MetricsExporter
		}
		if metricsExporterChanged(plan.Features.MetricsExporter, stateMetrics) {
			exporter, hasExporter := buildMetricsExporterParam(plan.Features.MetricsExporter)
			if !hasExporter {
				if stateMetrics != nil {
					resp.Diagnostics.AddError(
						"Config Update Error",
						"Removing the metrics_exporter block is not supported. Disable exporters by setting their enabled flag to false instead.",
					)
					return
				}
			} else {
				features := ensureFeatures()
				features.MetricsExporter = exporter
				hasUpdate = true
				shouldWait = true
			}
		}
	}

	if plan.ComputeSpecs != nil && instance.Spec != nil && instance.Spec.ReservedAku != nil {
		planAKU := int32(plan.ComputeSpecs.ReservedAku.ValueInt64())
		if planAKU != *instance.Spec.ReservedAku {
			aku := planAKU
			spec := ensureSpec()
			spec.ReservedAku = &aku
			hasUpdate = true
			shouldWait = true
		}
	}

	if plan.ComputeSpecs != nil {
		var stateNodeGroups []models.NodeGroupModel
		if state.ComputeSpecs != nil {
			stateNodeGroups = state.ComputeSpecs.KubernetesNodeGroups
		}
		if !areNodeGroupsEqual(plan.ComputeSpecs.KubernetesNodeGroups, stateNodeGroups) {
			groups := make([]client.KubernetesNodeGroupParam, 0, len(plan.ComputeSpecs.KubernetesNodeGroups))
			for _, group := range plan.ComputeSpecs.KubernetesNodeGroups {
				if group.ID.IsNull() || group.ID.IsUnknown() {
					continue
				}
				id := group.ID.ValueString()
				groups = append(groups, client.KubernetesNodeGroupParam{Id: &id})
			}
			spec := ensureSpec()
			spec.KubernetesNodeGroups = groups
			hasUpdate = true
			shouldWait = true
		}
	}

	if plan.Features != nil && plan.Features.WalMode.ValueString() == "FSWAL" && plan.ComputeSpecs != nil {
		planFS, planOK := extractFileSystemConfig(plan.ComputeSpecs)
		if !planOK {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When `features.wal_mode` is set to `FSWAL`, `compute_specs.file_system_param` must include both throughput and file system count values.",
			)
			return
		}
		var stateSpecs *models.ComputeSpecsModel
		if state.ComputeSpecs != nil {
			stateSpecs = state.ComputeSpecs
		}
		stateFS, stateOK := extractFileSystemConfig(stateSpecs)
		if !stateOK || planFS != stateFS {
			spec := ensureSpec()
			spec.FileSystem = &client.FileSystemParam{
				ThroughputMiBpsPerFileSystem: int32(planFS.ThroughputMiBpsPerFileSystem),
				FileSystemCount:              int32(planFS.FileSystemCount),
			}
			hasUpdate = true
			shouldWait = true
		}
	}

	if instanceConfigsChanged && plan.Features != nil {
		if state.Features == nil {
			state.Features = &models.FeaturesModel{}
		}
		state.Features.InstanceConfigs = plan.Features.InstanceConfigs
	}

	if certificateChanged && plan.Features != nil && plan.Features.Security != nil {
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

	if hasUpdate {
		if err := r.api.UpdateKafkaInstance(ctx, instanceId, updateParam); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", instanceId, err))
			return
		}
		if shouldWait {
			if err := waitForKafkaClusterToProvisionFunc(ctx, r.client, instanceId, models.StateChanging, updateTimeout); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
				return
			}
		}
		resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
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
	if err := framework.WaitForKafkaClusterToDeleted(ctx, r.client, instanceId, deleteTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}
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

	// Set the default value for instance_configs to an empty map

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), environmentId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), instanceId)...)

	config := types.MapValueMust(types.StringType, map[string]attr.Value{})
	integration := types.SetValueMust(types.StringType, []attr.Value{})
	features := models.FeaturesModel{
		InstanceConfigs: config,
		Integrations:    integration,
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("features"), features)...)
}

func ReadKafkaInstance(ctx context.Context, r *KafkaInstanceResource, instanceId string, plan *models.KafkaInstanceResourceModel) diag.Diagnostics {
	instance, err := r.api.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			return nil
		}
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", plan.InstanceID.ValueString(), err))}
	}
	if instance == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Kafka instance %q not found", plan.InstanceID.ValueString()))}
	}

	integrations, err := r.api.ListInstanceIntegrations(ctx, instanceId)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list integrations for Kafka instance %q, got error: %s", plan.InstanceID.ValueString(), err))}
	}

	endpoints, err := r.api.GetInstanceEndpoints(ctx, instanceId)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get endpoints for Kafka instance %q, got error: %s", plan.InstanceID.ValueString(), err))}
	}

	diags := diag.Diagnostics{}
	diags.Append(models.FlattenKafkaInstanceModel(instance, plan)...)
	diags.Append(models.FlattenKafkaInstanceModelWithIntegrations(integrations, plan)...)
	diags.Append(models.FlattenKafkaInstanceModelWithEndpoints(endpoints, plan)...)
	return diags
}

// Helper function to compare node groups regardless of order
func areNodeGroupsEqual(plan, state []models.NodeGroupModel) bool {
	if len(plan) != len(state) {
		return false
	}

	// Create maps for O(1) lookup
	planMap := make(map[string]struct{}, len(plan))
	for _, group := range plan {
		planMap[group.ID.ValueString()] = struct{}{}
	}

	// Check if all state node groups exist in plan
	for _, group := range state {
		if _, exists := planMap[group.ID.ValueString()]; !exists {
			return false
		}
	}

	return true
}

func stringSetsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; !ok {
			return false
		}
	}
	return true
}

func metricsExporterChanged(plan, state *models.MetricsExporterModel) bool {
	if plan == nil {
		return state != nil
	}
	if state == nil {
		return hasMetricsExporterConfig(plan)
	}
	if !prometheusExporterEqual(plan.Prometheus, state.Prometheus) {
		return true
	}
	if !kafkaExporterEqual(plan.Kafka, state.Kafka) {
		return true
	}
	return false
}

func hasMetricsExporterConfig(model *models.MetricsExporterModel) bool {
	if model == nil {
		return false
	}
	if model.Prometheus != nil && prometheusExporterHasConfig(model.Prometheus) {
		return true
	}
	if model.Kafka != nil && kafkaExporterHasConfig(model.Kafka) {
		return true
	}
	return false
}

func prometheusExporterEqual(plan, state *models.PrometheusExporterModel) bool {
	if plan == nil || state == nil {
		return plan == nil && state == nil
	}
	if !boolAttrEqual(plan.Enabled, state.Enabled) {
		return false
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

func prometheusExporterHasConfig(model *models.PrometheusExporterModel) bool {
	if model == nil {
		return false
	}
	if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
		return true
	}
	if !model.AuthType.IsNull() && !model.AuthType.IsUnknown() {
		return true
	}
	if !model.EndPoint.IsNull() && !model.EndPoint.IsUnknown() {
		return true
	}
	if !model.PrometheusArn.IsNull() && !model.PrometheusArn.IsUnknown() {
		return true
	}
	if !model.Username.IsNull() && !model.Username.IsUnknown() {
		return true
	}
	if !model.Password.IsNull() && !model.Password.IsUnknown() {
		return true
	}
	if !model.Token.IsNull() && !model.Token.IsUnknown() {
		return true
	}
	if !model.Labels.IsNull() && !model.Labels.IsUnknown() && len(model.Labels.Elements()) > 0 {
		return true
	}
	return false
}

func kafkaExporterEqual(plan, state *models.KafkaMetricsExporterModel) bool {
	if plan == nil || state == nil {
		return plan == nil && state == nil
	}
	if !boolAttrEqual(plan.Enabled, state.Enabled) {
		return false
	}
	if !stringAttrEqual(plan.BootstrapServers, state.BootstrapServers) {
		return false
	}
	if !stringAttrEqual(plan.Topic, state.Topic) {
		return false
	}
	if !int64AttrEqual(plan.CollectionPeriod, state.CollectionPeriod) {
		return false
	}
	if !stringAttrEqual(plan.SecurityProtocol, state.SecurityProtocol) {
		return false
	}
	if !stringAttrEqual(plan.SaslMechanism, state.SaslMechanism) {
		return false
	}
	if !stringAttrEqual(plan.SaslUsername, state.SaslUsername) {
		return false
	}
	if !stringAttrEqual(plan.SaslPassword, state.SaslPassword) {
		return false
	}
	return true
}

func kafkaExporterHasConfig(model *models.KafkaMetricsExporterModel) bool {
	if model == nil {
		return false
	}
	if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
		return true
	}
	if !model.BootstrapServers.IsNull() && !model.BootstrapServers.IsUnknown() {
		return true
	}
	if !model.Topic.IsNull() && !model.Topic.IsUnknown() {
		return true
	}
	if !model.CollectionPeriod.IsNull() && !model.CollectionPeriod.IsUnknown() {
		return true
	}
	if !model.SecurityProtocol.IsNull() && !model.SecurityProtocol.IsUnknown() {
		return true
	}
	if !model.SaslMechanism.IsNull() && !model.SaslMechanism.IsUnknown() {
		return true
	}
	if !model.SaslUsername.IsNull() && !model.SaslUsername.IsUnknown() {
		return true
	}
	if !model.SaslPassword.IsNull() && !model.SaslPassword.IsUnknown() {
		return true
	}
	return false
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
	if model.Kafka != nil {
		kafka, ok := buildKafkaExporterParam(model.Kafka)
		if ok {
			exporter.Kafka = kafka
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
	if !prometheusExporterHasConfig(model) {
		return nil, false
	}
	prom := &client.InstancePrometheusExporterParam{}
	if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
		enabled := model.Enabled.ValueBool()
		prom.Enabled = &enabled
	}
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
				promLabels[i] = client.MetricsLabelParam{Name: label.Key, Value: label.Value}
			}
			prom.Labels = promLabels
		}
	}
	return prom, true
}

func buildKafkaExporterParam(model *models.KafkaMetricsExporterModel) (*client.InstanceKafkaMetricsExporterParam, bool) {
	if model == nil {
		return nil, false
	}
	if !kafkaExporterHasConfig(model) {
		return nil, false
	}
	kafka := &client.InstanceKafkaMetricsExporterParam{}
	if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
		enabled := model.Enabled.ValueBool()
		kafka.Enabled = &enabled
	}
	if !model.BootstrapServers.IsNull() && !model.BootstrapServers.IsUnknown() {
		servers := model.BootstrapServers.ValueString()
		kafka.BootstrapServers = &servers
	}
	if !model.Topic.IsNull() && !model.Topic.IsUnknown() {
		topic := model.Topic.ValueString()
		kafka.Topic = &topic
	}
	if !model.CollectionPeriod.IsNull() && !model.CollectionPeriod.IsUnknown() {
		period := int32(model.CollectionPeriod.ValueInt64())
		kafka.CollectionPeriod = &period
	}
	if !model.SecurityProtocol.IsNull() && !model.SecurityProtocol.IsUnknown() {
		protocol := model.SecurityProtocol.ValueString()
		kafka.SecurityProtocol = &protocol
	}
	if !model.SaslMechanism.IsNull() && !model.SaslMechanism.IsUnknown() {
		mechanism := model.SaslMechanism.ValueString()
		kafka.SaslMechanism = &mechanism
	}
	if !model.SaslUsername.IsNull() && !model.SaslUsername.IsUnknown() {
		username := model.SaslUsername.ValueString()
		kafka.SaslUsername = &username
	}
	if !model.SaslPassword.IsNull() && !model.SaslPassword.IsUnknown() {
		password := model.SaslPassword.ValueString()
		kafka.SaslPassword = &password
	}
	return kafka, true
}

func boolAttrEqual(plan, state types.Bool) bool {
	if plan.IsUnknown() {
		return true
	}
	if plan.IsNull() {
		return state.IsNull() || state.IsUnknown()
	}
	if state.IsNull() || state.IsUnknown() {
		return false
	}
	return plan.ValueBool() == state.ValueBool()
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

type fileSystemConfig struct {
	ThroughputMiBpsPerFileSystem int64
	FileSystemCount              int64
}

func extractFileSystemConfig(specs *models.ComputeSpecsModel) (fileSystemConfig, bool) {
	if specs == nil || specs.FileSystemParam == nil {
		return fileSystemConfig{}, false
	}
	fs := specs.FileSystemParam
	if fs.ThroughputMiBpsPerFileSystem.IsNull() || fs.ThroughputMiBpsPerFileSystem.IsUnknown() {
		return fileSystemConfig{}, false
	}
	if fs.FileSystemCount.IsNull() || fs.FileSystemCount.IsUnknown() {
		return fileSystemConfig{}, false
	}
	return fileSystemConfig{
		ThroughputMiBpsPerFileSystem: fs.ThroughputMiBpsPerFileSystem.ValueInt64(),
		FileSystemCount:              fs.FileSystemCount.ValueInt64(),
	}, true
}

var waitForKafkaClusterToProvisionFunc = framework.WaitForKafkaClusterToProvision
