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
			"> **Note**: This provider version is only compatible with AutoMQ control plane versions 8.0 and later.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment identifier (e.g. `env-xxxxx`). Find this on the AutoMQ console System Settings page.",
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
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The software version of AutoMQ instance. By default, there is no need to set version; the latest version will be used. If you need to specify a version, refer to the [documentation](https://docs.automq.com/automq-cloud/release-notes) to choose the appropriate version number.",
			},
			"compute_specs": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The compute specs of the instance",
				Attributes: map[string]schema.Attribute{
					"reserved_aku": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "AKU (AutoMQ Kafka Unit) defines the cluster scale. Each AKU provides up to 30 MiB/s write or 60 MiB/s read throughput. For sizing guidance, refer to the [billing documentation](https://docs.automq.com/automq-cloud/subscriptions-and-billings/byoc-env-billings/billing-instructions-for-byoc#indicator-constraints).",
					},
					"deploy_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Deployment platform for the instance.",
					},
					"dns_zone":                   schema.StringAttribute{Computed: true, MarkdownDescription: "DNS zone used when creating custom records."},
					"kubernetes_cluster_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Identifier for the target Kubernetes cluster."},
					"kubernetes_namespace":       schema.StringAttribute{Computed: true, MarkdownDescription: "Kubernetes namespace for the instance deployment."},
					"kubernetes_service_account": schema.StringAttribute{Computed: true, MarkdownDescription: "Kubernetes service account for the instance pods."},
					"instance_role":              schema.StringAttribute{Computed: true, MarkdownDescription: "IAM role ARN for the Kafka instance."},
					"networks": schema.ListNestedAttribute{
						Computed:            true,
						MarkdownDescription: "To configure the network settings for an instance, you need to specify the availability zone(s) and subnet information. Currently, you can set either one availability zone or three availability zones.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"zone": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Cloud provider availability zone ID (e.g. `us-east-1a` for AWS).",
								},
								"subnets": schema.ListAttribute{
									Computed:            true,
									MarkdownDescription: "Specify the subnet under the corresponding availability zone for deploying the instance. Currently, only one subnet can be set for each availability zone.",
									ElementType:         types.StringType,
								},
							},
						},
					},
					"kubernetes_node_groups": schema.ListNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Kubernetes node groups configuration",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Node group ID",
								},
							},
						},
					},
					"data_buckets": schema.ListNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Inline bucket configuration replacing legacy bucket profiles.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"bucket_name": schema.StringAttribute{Computed: true, MarkdownDescription: "Object storage bucket name used for data."},
							},
						},
					},
					"security_groups": schema.ListAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						MarkdownDescription: "Security groups for the instance",
					},
					"file_system_param": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "File system configuration for FSWAL mode",
						Attributes: map[string]schema.Attribute{
							"throughput_mibps_per_file_system": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "Throughput in MiBps per file system",
							},
							"file_system_count": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "Number of file systems",
							},
							"security_groups": schema.ListAttribute{
								ElementType:         types.StringType,
								Computed:            true,
								MarkdownDescription: "Security groups for file systems",
							},
						},
					},
				},
			},
			"features": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Feature configuration for the Kafka instance.",
				Attributes: map[string]schema.Attribute{
					"wal_mode": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Write-Ahead Logging mode: EBSWAL (using EBS as write buffer), S3WAL (using object storage as write buffer), or FSWAL (using file system as write buffer). Defaults to EBSWAL.",
					},
					"instance_configs": schema.MapAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "Additional configuration for the Kafka Instance. The currently supported parameters can be set by referring to the [documentation](https://docs.automq.com/automq-cloud/using-automq-for-kafka/restrictions#instance-level-configuration).",
						Computed:            true,
					},
					"metrics_exporter": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Prometheus Remote Write metrics exporter configuration.",
						Attributes: map[string]schema.Attribute{
							"prometheus": schema.SingleNestedAttribute{
								Computed:            true,
								MarkdownDescription: "Prometheus Remote Write configuration for exporting metrics.",
								Attributes: map[string]schema.Attribute{
									"auth_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Authentication type. Values: `noauth`, `basic`, `bearer`, or `sigv4`.",
									},
									"endpoint": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Prometheus Remote Write endpoint URL.",
									},
									"prometheus_arn": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "AWS Managed Service for Prometheus workspace ARN (required for `sigv4` auth).",
									},
									"username": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Username for HTTP Basic authentication.",
									},
									"password": schema.StringAttribute{
										Computed:            true,
										Sensitive:           true,
										MarkdownDescription: "Password for HTTP Basic authentication.",
									},
									"token": schema.StringAttribute{
										Computed:            true,
										Sensitive:           true,
										MarkdownDescription: "Bearer token for authentication.",
									},
									"labels": schema.MapAttribute{
										Computed:            true,
										MarkdownDescription: "Custom labels attached to exported metrics.",
										ElementType:         types.StringType,
									},
								},
							},
							"kafka": schema.SingleNestedAttribute{
								Computed:            true,
								MarkdownDescription: "Kafka metrics exporter configuration.",
								Attributes: map[string]schema.Attribute{
									"enabled":           schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the Kafka metrics exporter is enabled."},
									"bootstrap_servers": schema.StringAttribute{Computed: true, MarkdownDescription: "Bootstrap servers for the metrics Kafka cluster."},
									"topic":             schema.StringAttribute{Computed: true, MarkdownDescription: "Kafka topic for metrics data."},
									"collection_period": schema.Int64Attribute{Computed: true, MarkdownDescription: "Metrics collection period in seconds."},
									"security_protocol": schema.StringAttribute{Computed: true, MarkdownDescription: "Security protocol for the metrics Kafka cluster."},
									"sasl_mechanism":    schema.StringAttribute{Computed: true, MarkdownDescription: "SASL mechanism for the metrics Kafka cluster."},
									"sasl_username":     schema.StringAttribute{Computed: true, MarkdownDescription: "SASL username for the metrics Kafka cluster."},
									"sasl_password":     schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "SASL password for the metrics Kafka cluster."},
								},
							},
						},
					},
					"table_topic": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Table topic configuration (warehouse/catalog settings).",
						Attributes: map[string]schema.Attribute{
							"warehouse":          schema.StringAttribute{Computed: true, MarkdownDescription: "Warehouse location for table data."},
							"catalog_type":       schema.StringAttribute{Computed: true, MarkdownDescription: "Catalog type: `s3Table`, `glue`, or `hive`."},
							"metastore_uri":      schema.StringAttribute{Computed: true, MarkdownDescription: "Hive Metastore endpoint (for `hive` catalog)."},
							"hive_auth_mode":     schema.StringAttribute{Computed: true, MarkdownDescription: "Authentication mode for Hive Metastore."},
							"kerberos_principal": schema.StringAttribute{Computed: true, MarkdownDescription: "Kerberos principal for Hive Metastore server."},
							"user_principal":     schema.StringAttribute{Computed: true, MarkdownDescription: "Kerberos user principal for authentication."},
							"keytab_file":        schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Kerberos keytab file content."},
							"krb5conf_file":      schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Kerberos krb5.conf file content."},
						},
					},
					"security": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Security configuration for the Kafka instance.",
						Attributes: map[string]schema.Attribute{
							"authentication_methods": schema.SetAttribute{
								Computed:            true,
								ElementType:         types.StringType,
								MarkdownDescription: "Authentication methods: anonymous (anonymous access), sasl (SASL user auth), mtls (TLS cert auth). Defaults to anonymous.",
							},
							"transit_encryption_modes": schema.SetAttribute{
								Computed:            true,
								ElementType:         types.StringType,
								MarkdownDescription: "Transit encryption modes: plaintext (unencrypted) or tls (TLS encrypted). Defaults to plaintext.",
							},
							"data_encryption_mode": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Data encryption mode: NONE (no encryption), CPMK (cloud-managed KMS), BYOK (custom KMS key)",
							},
							"tls_hostname_validation_enabled": schema.BoolAttribute{
								Computed:            true,
								MarkdownDescription: "Whether TLS hostname validation is enabled for broker certificates.",
							},
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
			},
			"endpoints": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The bootstrap endpoints of instance. AutoMQ supports multiple access protocols; therefore, the Endpoint is a list.",
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
	// Get instance endpoints
	endpoints, err := r.client.GetInstanceEndpoints(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get endpoints for Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	model := models.KafkaInstanceModel{}
	// Flatten API response into Terraform state
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModel(ctx, out, &instance)...)
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
