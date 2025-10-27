package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &KafkaInstanceDataSource{}

func NewKafkaInstanceDataSource() datasource.DataSource {
	ds := &KafkaInstanceDataSource{}
	return ds
}

// KafkaInstanceDataSource defines the resource implementation.
type KafkaInstanceDataSource struct {
	client *client.Client
}

func (r *KafkaInstanceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_instance"
}

func (r *KafkaInstanceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)\n\n" +
			"Using the `automq_kafka_instance` data source, you can manage kafka resoure within instance.\n\n" +
			"> **Note**: This provider version is only compatible with AutoMQ control plane versions 7.3.5 and later.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment, this attribute is specified during the deployment and installation process.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the Kafka instance.",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Kafka instance. It can contain letters (a-z or A-Z), numbers (0-9), underscores (_), and hyphens (-), with a length limit of 3 to 64 characters.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The instance description are used to differentiate the purpose of the instance. They support letters (a-z or A-Z), numbers (0-9), underscores (_), spaces( ) and hyphens (-), with a length limit of 3 to 128 characters.",
				Computed:            true,
			},
			"deploy_profile": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The software version of AutoMQ instance. By default, there is no need to set version; the latest version will be used. If you need to specify a version, refer to the [documentation](https://docs.automq.com/automq-cloud/release-notes) to choose the appropriate version number.",
			},
			"compute_specs": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The compute specs of the instance",
				Attributes: map[string]schema.Attribute{
					"reserved_aku": schema.Int64Attribute{
						Computed:    true,
						Description: "AutoMQ defines AKU (AutoMQ Kafka Unit) to measure the scale of the cluster. Each AKU provides 20 MiB/s of read/write throughput. For more details on AKU, please refer to the [documentation](https://docs.automq.com/automq-cloud/subscriptions-and-billings/byoc-env-billings/billing-instructions-for-byoc#indicator-constraints). The currently supported AKU specifications are 6, 8, 10, 12, 14, 16, 18, 20, 22, and 24. If an invalid AKU value is set, the instance cannot be created.",
					},
					"deploy_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Deployment platform for the instance.",
					},
					"provider":                   schema.StringAttribute{Computed: true},
					"region":                     schema.StringAttribute{Computed: true},
					"scope":                      schema.StringAttribute{Computed: true},
					"vpc":                        schema.StringAttribute{Computed: true},
					"dns_zone":                   schema.StringAttribute{Computed: true},
					"kubernetes_cluster_id":      schema.StringAttribute{Computed: true},
					"kubernetes_namespace":       schema.StringAttribute{Computed: true},
					"kubernetes_service_account": schema.StringAttribute{Computed: true},
					"security_group":             schema.StringAttribute{Computed: true},
					"instance_role":              schema.StringAttribute{Computed: true},
					"networks": schema.ListNestedAttribute{
						Computed:    true,
						Description: "To configure the network settings for an instance, you need to specify the availability zone(s) and subnet information. Currently, you can set either one availability zone or three availability zones.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"zone": schema.StringAttribute{
									Computed:    true,
									Description: "The availability zone ID of the cloud provider.",
								},
								"subnets": schema.ListAttribute{
									Computed:    true,
									Description: "Specify the subnet under the corresponding availability zone for deploying the instance. Currently, only one subnet can be set for each availability zone.",
									ElementType: types.StringType,
								},
							},
						},
					},
					"kubernetes_node_groups": schema.ListNestedAttribute{
						Computed:    true,
						Description: "Kubernetes node groups configuration",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:    true,
									Description: "Node group ID",
								},
							},
						},
					},
					"bucket_profiles": schema.ListNestedAttribute{
						Computed:           true,
						Description:        "(Deprecated) Bucket profile bindings.",
						DeprecationMessage: "bucket_profiles is deprecated. Consult data_buckets instead.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:    true,
									Description: "Bucket profile ID",
								},
							},
						},
					},
					"data_buckets": schema.ListNestedAttribute{
						Computed:    true,
						Description: "Inline bucket configuration replacing legacy bucket profiles.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"bucket_name": schema.StringAttribute{Computed: true},
							},
						},
					},
					"file_system_param": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "File system configuration for FSWAL clusters.",
						Attributes: map[string]schema.Attribute{
							"throughput_mibps_per_file_system": schema.Int64Attribute{
								Computed:    true,
								Description: "Provisioned throughput in MiB/s for each file system.",
							},
							"file_system_count": schema.Int64Attribute{
								Computed:    true,
								Description: "Number of file systems allocated for WAL storage.",
							},
						},
					},
				},
			},
			"features": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"wal_mode": schema.StringAttribute{
						Computed:    true,
						Description: "Write-Ahead Logging mode: EBSWAL (using EBS as write buffer), S3WAL (using object storage as write buffer), or FSWAL (using file systems as write buffer). Defaults to EBSWAL.",
					},
					"instance_configs": schema.MapAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "Additional configuration for the Kafka Instance. The currently supported parameters can be set by referring to the [documentation](https://docs.automq.com/automq-cloud/using-automq-for-kafka/restrictions#instance-level-configuration).",
						Computed:            true,
					},
					"integrations": schema.SetAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "(Deprecated) Integration identifiers returned for compatibility.",
					},
					"metrics_exporter": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Metrics exporter configuration for Prometheus or Kafka sinks.",
						Attributes: map[string]schema.Attribute{
							"prometheus": schema.SingleNestedAttribute{
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"enabled":        schema.BoolAttribute{Computed: true},
									"auth_type":      schema.StringAttribute{Computed: true},
									"end_point":      schema.StringAttribute{Computed: true},
									"prometheus_arn": schema.StringAttribute{Computed: true},
									"username":       schema.StringAttribute{Computed: true},
									"password":       schema.StringAttribute{Computed: true, Sensitive: true},
									"token":          schema.StringAttribute{Computed: true, Sensitive: true},
									"labels": schema.MapAttribute{
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
							"kafka": schema.SingleNestedAttribute{
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"enabled":           schema.BoolAttribute{Computed: true},
									"bootstrap_servers": schema.StringAttribute{Computed: true},
									"topic":             schema.StringAttribute{Computed: true},
									"collection_period": schema.Int64Attribute{Computed: true},
									"security_protocol": schema.StringAttribute{Computed: true},
									"sasl_mechanism":    schema.StringAttribute{Computed: true},
									"sasl_username":     schema.StringAttribute{Computed: true},
									"sasl_password":     schema.StringAttribute{Computed: true, Sensitive: true},
								},
							},
						},
					},
					"table_topic": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Table topic configuration (warehouse/catalog settings).",
						Attributes: map[string]schema.Attribute{
							"warehouse":          schema.StringAttribute{Computed: true},
							"catalog_type":       schema.StringAttribute{Computed: true},
							"metastore_uri":      schema.StringAttribute{Computed: true},
							"hive_auth_mode":     schema.StringAttribute{Computed: true},
							"kerberos_principal": schema.StringAttribute{Computed: true},
							"user_principal":     schema.StringAttribute{Computed: true},
							"keytab_file":        schema.StringAttribute{Computed: true, Sensitive: true},
							"krb5conf_file":      schema.StringAttribute{Computed: true, Sensitive: true},
						},
					},
					"inbound_rules": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"listener_name": schema.StringAttribute{Computed: true},
								"cidrs": schema.ListAttribute{
									Computed:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
					"extend_listeners": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"listener_name":     schema.StringAttribute{Computed: true},
								"security_protocol": schema.StringAttribute{Computed: true},
								"port":              schema.Int64Attribute{Computed: true},
							},
						},
					},
					"security": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"authentication_methods": schema.SetAttribute{
								Computed:    true,
								ElementType: types.StringType,
								Description: "Authentication methods: anonymous (anonymous access), sasl (SASL user auth), mtls (TLS cert auth). Defaults to anonymous.",
							},
							"transit_encryption_modes": schema.SetAttribute{
								Computed:    true,
								ElementType: types.StringType,
								Description: "Transit encryption modes: plaintext (unencrypted) or tls (TLS encrypted). Defaults to plaintext.",
							},
							"data_encryption_mode": schema.StringAttribute{
								Computed:    true,
								Description: "Data encryption mode: NONE (no encryption), CPMK (cloud-managed KMS), BYOK (custom KMS key)",
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
			},
			"endpoints": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The bootstrap endpoints of instance. AutoMQ supports multiple access protocols; therefore, the Endpoint is a list.",
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
	}
}

func (r *KafkaInstanceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *KafkaInstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.KafkaInstanceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, config.EnvironmentID.ValueString())

	instance := models.KafkaInstanceResourceModel{}
	var out *client.InstanceVO
	var err error

	if !config.InstanceID.IsNull() {
		instanceId := config.InstanceID.ValueString()
		out, err = r.client.GetKafkaInstance(ctx, instanceId)
		if err != nil {
			if framework.IsNotFoundError(err) {
				resp.Diagnostics.AddError(fmt.Sprintf("Kafka instance %q not found", instanceId), err.Error())
				return
			}
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", instanceId, err))
			return
		}
		if out == nil {
			resp.Diagnostics.AddError("Kafka instance not found", fmt.Sprintf("Kafka instance %q returned nil", instanceId))
			return
		}
		if !config.Name.IsNull() && *out.Name != config.Name.ValueString() {
			resp.Diagnostics.AddError(
				"Name Mismatch",
				fmt.Sprintf("The Kafka instance name '%s' does not match the expected name '%s'.", *out.Name, config.Name.ValueString()),
			)
		}
	} else if !config.Name.IsNull() {
		instanceName := config.Name.ValueString()
		out, err = r.client.GetKafkaInstanceByName(ctx, instanceName)
		if err != nil {
			if framework.IsNotFoundError(err) {
				resp.Diagnostics.AddError(fmt.Sprintf("Kafka instance %q not found", instanceName), err.Error())
				return
			}
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", instanceName, err))
		}
		if out == nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Kafka instance %q not found", instanceName), err.Error())
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Either 'id' or 'name' must be provided.",
		)
		return
	}

	instanceId := *out.InstanceId
	// Get instance integrations
	integrations, err := r.client.ListInstanceIntegrations(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list integrations for Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	// Get instance endpoints
	endpoints, err := r.client.GetInstanceEndpoints(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get endpoints for Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	model := models.KafkaInstanceModel{}
	// Flatten API response into Terraform state
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModel(out, &instance)...)
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModelWithIntegrations(integrations, &instance)...)
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModelWithEndpoints(endpoints, &instance)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get instance configurations
	configs, err := r.client.GetInstanceConfigs(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get configurations for Kafka instance %q, got error: %s", instanceId, err))
		return
	}

	// Convert API response into data source model
	if err := models.ConvertKafkaInstanceModel(&instance, &model); err != nil {
		resp.Diagnostics.AddError("Conversion Error", fmt.Sprintf("Failed to convert Kafka instance model: %s", err))
		return
	}
	// Update the model with the configurations
	model.Features.InstanceConfigs = models.FlattenStringValueMap(configs)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
