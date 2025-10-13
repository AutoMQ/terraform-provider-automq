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
	framework.WithTimeouts
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
						MarkdownDescription: "Deployment platform for the instance. Supported values: `IAAS`, `KUBERNETES`.",
						Validators: []validator.String{
							stringvalidator.OneOf("IAAS", "KUBERNETES"),
						},
					},
					"provider": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Cloud provider identifier, e.g. `aws`.",
					},
					"region": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Region where the instance will be deployed.",
					},
					"scope": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Cloud provider scope such as account ID or organization.",
					},
					"vpc": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "VPC identifier for the deployment target.",
					},
					"domain": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Custom domain for the Kafka instance endpoints.",
					},
					"dns_zone": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "DNS zone used when creating custom records.",
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
						Description: "Inline bucket configuration replacing legacy bucket profiles.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"bucket_name": schema.StringAttribute{
									Required:    true,
									Description: "Object storage bucket name used for data.",
								},
								"provider": schema.StringAttribute{
									Optional: true,
								},
								"region": schema.StringAttribute{
									Optional: true,
								},
								"scope": schema.StringAttribute{
									Optional: true,
								},
								"credential": schema.StringAttribute{
									Optional: true,
								},
								"endpoint": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Custom endpoint for accessing the bucket (e.g., when using private object storage).",
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtMost(1),
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
						Optional: true,
					},
					"kubernetes_service_account": schema.StringAttribute{
						Optional: true,
					},
					"credential": schema.StringAttribute{
						Optional: true,
					},
					"instance_role": schema.StringAttribute{
						Optional: true,
					},
					"tenant_id": schema.StringAttribute{
						Optional: true,
					},
					"vpc_resource_group": schema.StringAttribute{
						Optional: true,
					},
					"k8s_resource_group": schema.StringAttribute{
						Optional: true,
					},
					"dns_resource_group": schema.StringAttribute{
						Optional: true,
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
								PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
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
						MarkdownDescription: "Inline metrics exporter configuration for Prometheus, CloudWatch or Kafka sinks.",
						Attributes: map[string]schema.Attribute{
							"prometheus": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"enabled":        schema.BoolAttribute{Optional: true},
									"auth_type":      schema.StringAttribute{Optional: true},
									"end_point":      schema.StringAttribute{Optional: true},
									"prometheus_arn": schema.StringAttribute{Optional: true},
									"username":       schema.StringAttribute{Optional: true},
									"password":       schema.StringAttribute{Optional: true, Sensitive: true},
									"token":          schema.StringAttribute{Optional: true, Sensitive: true},
									"labels": schema.MapAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
							"cloudwatch": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"enabled":   schema.BoolAttribute{Optional: true},
									"namespace": schema.StringAttribute{Optional: true},
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
									"sasl_password":     schema.StringAttribute{Optional: true, Sensitive: true},
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
							"keytab_file":        schema.StringAttribute{Optional: true, Sensitive: true},
							"krb5conf_file":      schema.StringAttribute{Optional: true, Sensitive: true},
						},
					},
					"s3_failover": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"enabled":            schema.BoolAttribute{Optional: true},
							"storage_type":       schema.StringAttribute{Optional: true},
							"ebs_volume_size_gb": schema.Int64Attribute{Optional: true},
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
}

func (r *KafkaInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var instance models.KafkaInstanceResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &instance)...)
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

	out, err := r.client.CreateKafkaInstance(ctx, in)
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
	if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateCreating, createTimeout); err != nil {
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
	instance, err := r.client.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			// Treat HTTP 404 Not Found status as a signal to recreate resource and return early
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
		return
	}
	// Get instance integrations
	integrations, err := r.client.ListInstanceIntegrations(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list integrations for Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
		return
	}
	// Get instance endpoints
	endpoints, err := r.client.GetInstanceEndpoints(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get endpoints for Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
		return
	}
	// Flatten API response into Terraform state
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModel(instance, &state)...)
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModelWithIntegrations(integrations, &state)...)
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModelWithEndpoints(endpoints, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.KafkaInstanceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	// check if the instance exists
	instanceId := plan.InstanceID.ValueString()
	instance, err := r.client.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	if instance == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q not found", instanceId))
		return
	}
	// check if the instance is in available state
	if *instance.State != models.StateRunning {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q is Currently in %q state, only instances in 'Running' state can be updated", instanceId, *instance.State))
		return
	}

	// Check if the basic info has changed
	if state.Name.ValueString() != plan.Name.ValueString() ||
		state.Description.ValueString() != plan.Description.ValueString() {
		// Generate API request body from plan
		basicUpdate := client.InstanceBasicParam{
			DisplayName: plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
		}
		err = r.client.UpdateKafkaInstanceBasicInfo(ctx, instanceId, basicUpdate)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q basicInfo, got error: %s", instanceId, err))
			return
		}
		// get latest info
		resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, state.Timeouts)

	// Check if the Integrations has changed
	planIntegration := models.ExpandSetValueList(plan.Features.Integrations)
	stateIntegration := models.ExpandSetValueList(state.Features.Integrations)

	// Initialize slices to track integration changes
	needAddIntegration := []string{}
	needRemoveIntegration := []string{}

	// Convert plan and state integrations to map for efficient lookup
	planMap := make(map[string]bool)
	stateMap := make(map[string]bool)

	for _, v := range planIntegration {
		planMap[v] = true
	}
	for _, v := range stateIntegration {
		stateMap[v] = true
	}

	// Find integrations that need to be added
	for integration := range planMap {
		if !stateMap[integration] {
			needAddIntegration = append(needAddIntegration, integration)
		}
	}

	// Find integrations that need to be removed
	for integration := range stateMap {
		if !planMap[integration] {
			needRemoveIntegration = append(needRemoveIntegration, integration)
		}
	}

	if len(needAddIntegration) > 0 {
		// Generate API request body from plan
		param := client.IntegrationInstanceAddParam{
			Codes: needAddIntegration,
		}
		err = r.client.AddInstanceIntergation(ctx, instanceId, &param)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add integrations for Kafka instance %q intergations, got error: %s", instanceId, err))
			return
		}
		// wait for version update
		if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateChanging, updateTimeout); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
			return
		}
		resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(needRemoveIntegration) > 0 {
		// Generate API request body from plan
		for _, integration := range needRemoveIntegration {
			err = r.client.RemoveInstanceIntergation(ctx, instanceId, integration)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove integrations for Kafka instance %q intergations, got error: %s", instanceId, err))
				return
			}
			// wait for version update
			if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateChanging, updateTimeout); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
				return
			}
			resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}
			// Save updated data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	planConfig := plan.Features.InstanceConfigs
	stateConfig := state.Features.InstanceConfigs
	// check if the config has changed
	if !models.MapsEqual(planConfig, stateConfig) {
		// Check if the plan config has removed any settings
		for name := range stateConfig.Elements() {
			if _, ok := planConfig.Elements()[name]; !ok {
				resp.Diagnostics.AddError("Config Update Error", fmt.Sprintf("Error occurred while updating Kafka Instance %q. "+
					" At present, we don't support the removal of instance settings from the 'configs' block, "+
					"meaning you can't reset to the instance's default settings. "+
					"As a workaround, you can find the default value and manually set the current value to match the default.", instanceId))
				return
			}
		}

		in := client.InstanceConfigParam{}
		in.Configs = models.ExpandStringValueMap(planConfig)

		err := r.client.UpdateKafkaInstanceConfig(ctx, instanceId, in)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q configs, got error: %s", instanceId, err))
			return
		}

		// wait for version update
		if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateChanging, updateTimeout); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
			return
		}

		state.Features.InstanceConfigs = planConfig
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Check if the Security has changed
	isCertificateChanged := false
	if plan.Features.Security != nil && instance.Features.Security != nil {
		// Check if any of the certificate fields have changed
		if !plan.Features.Security.CertificateAuthority.Equal(state.Features.Security.CertificateAuthority) ||
			!plan.Features.Security.CertificateChain.Equal(state.Features.Security.CertificateChain) ||
			!plan.Features.Security.PrivateKey.Equal(state.Features.Security.PrivateKey) {
			isCertificateChanged = true
		}
	}

	if isCertificateChanged {
		param := client.InstanceCertificateParam{
			CertificateAuthority: plan.Features.Security.CertificateAuthority.ValueString(),
			CertificateChain:     plan.Features.Security.CertificateChain.ValueString(),
			PrivateKey:           plan.Features.Security.PrivateKey.ValueString(),
		}

		// Call API to update certificate
		err := r.client.UpdateKafkaInstanceCertificate(ctx, instanceId, param)
		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to update Kafka instance %q certificate, got error: %s", instanceId, err))
			return
		}

		// Wait for certificate update to complete
		if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateChanging, updateTimeout); err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Error waiting for Kafka Cluster %q certificate update: %s", instanceId, err))
			return
		}

		// updated instance state
		state.Features.Security.PrivateKey = plan.Features.Security.PrivateKey
		state.Features.Security.CertificateChain = plan.Features.Security.CertificateChain
		state.Features.Security.CertificateAuthority = plan.Features.Security.CertificateAuthority

		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	planAKU := int32(plan.ComputeSpecs.ReservedAku.ValueInt64())
	planVersion := plan.Version.ValueString()
	planNodeGroup := make([]client.KubernetesNodeGroupParam, 0, len(plan.ComputeSpecs.KubernetesNodeGroups))
	for _, group := range plan.ComputeSpecs.KubernetesNodeGroups {
		id := group.ID.ValueString()
		planNodeGroup = append(planNodeGroup, client.KubernetesNodeGroupParam{
			Id: &id,
		})
	}

	// Check and update version if needed
	stateVersion := state.Version.ValueString()
	if planVersion != "" && planVersion != stateVersion {
		updateParam := client.InstanceUpdateParam{
			Version: &planVersion,
		}
		if err := updateInstanceAndWait(ctx, r, instanceId, updateParam, "version", updateTimeout, &state, resp); err != nil {
			return
		}
	}
	// Check and update AKU if needed
	stateAKU := *instance.Spec.ReservedAku
	if planAKU != stateAKU {
		planAKUCopy := planAKU
		updateParam := client.InstanceUpdateParam{
			Spec: &client.SpecificationUpdateParam{
				ReservedAku: &planAKUCopy,
			},
		}
		if err := updateInstanceAndWait(ctx, r, instanceId, updateParam, "aku", updateTimeout, &state, resp); err != nil {
			return
		}
	}

	// Check and update node groups if needed
	if !areNodeGroupsEqual(plan.ComputeSpecs.KubernetesNodeGroups, state.ComputeSpecs.KubernetesNodeGroups) {
		updateParam := client.InstanceUpdateParam{
			Spec: &client.SpecificationUpdateParam{
				KubernetesNodeGroups: planNodeGroup,
			},
		}
		if err := updateInstanceAndWait(ctx, r, instanceId, updateParam, "node_groups", updateTimeout, &state, resp); err != nil {
			return
		}
	}

	// Check and update file system settings for FSWAL clusters
	if plan.Features != nil && plan.Features.WalMode.ValueString() == "FSWAL" {
		planFS, planOk := extractFileSystemConfig(plan.ComputeSpecs)
		stateFS, stateOk := extractFileSystemConfig(state.ComputeSpecs)
		if !planOk {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When `features.wal_mode` is set to `FSWAL`, `compute_specs.file_system_param` must include both throughput and file system count values.",
			)
			return
		}
		if !stateOk || planFS.ThroughputMiBpsPerFileSystem != stateFS.ThroughputMiBpsPerFileSystem || planFS.FileSystemCount != stateFS.FileSystemCount {
			updateParam := client.InstanceUpdateParam{
				Spec: &client.SpecificationUpdateParam{
					FileSystem: &client.FileSystemParam{
						ThroughputMiBpsPerFileSystem: int32(planFS.ThroughputMiBpsPerFileSystem),
						FileSystemCount:              int32(planFS.FileSystemCount),
					},
				},
			}
			if err := updateInstanceAndWait(ctx, r, instanceId, updateParam, "file_system", updateTimeout, &state, resp); err != nil {
				return
			}
		}
	}
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
	instance, err := r.client.GetKafkaInstance(ctx, instanceId)
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
		err = r.client.DeleteKafkaInstance(ctx, instanceId)
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
	instance, err := r.client.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			return nil
		}
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", plan.InstanceID.ValueString(), err))}
	}
	if instance == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Kafka instance %q not found", plan.InstanceID.ValueString()))}
	}

	integrations, err := r.client.ListInstanceIntegrations(ctx, instanceId)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list integrations for Kafka instance %q, got error: %s", plan.InstanceID.ValueString(), err))}
	}

	endpoints, err := r.client.GetInstanceEndpoints(ctx, instanceId)
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

// Helper function to handle instance updates
func updateInstanceAndWait(
	ctx context.Context,
	r *KafkaInstanceResource,
	instanceId string,
	param client.InstanceUpdateParam,
	updateType string,
	timeout time.Duration,
	state *models.KafkaInstanceResourceModel,
	resp *resource.UpdateResponse,
) error {
	tflog.Debug(ctx, fmt.Sprintf("Updating Kafka instance compute specs due to changes in %s", updateType))

	var err error
	switch updateType {
	case "file_system":
		err = r.client.UpdateKafkaInstanceFileSystems(ctx, instanceId, param)
	default:
		err = r.client.UpdateKafkaInstanceComputeSpecs(ctx, instanceId, param)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update Kafka instance %q compute specs (%s), got error: %s",
				instanceId, updateType, err),
		)
		return err
	}

	if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateChanging, timeout); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Error waiting for Kafka Cluster %q compute specs update: %s", instanceId, err),
		)
		return err
	}

	resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, state)...)
	if resp.Diagnostics.HasError() {
		return fmt.Errorf("failed to read updated instance state")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return fmt.Errorf("failed to set updated instance state")
	}

	return nil
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
