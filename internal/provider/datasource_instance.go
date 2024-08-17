package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)<br><br>Using the `automq_kafka_instance` data source, you can manage kafka resoure within instance.",

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
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "The cloud provider of kafka instance. Currently, `aws` is supported.",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region of the Kafka instance",
				Computed:            true,
			},
			"networks": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The networks of the Kafka instance. Currently, you can get one availability zone or three availability zones.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone": schema.StringAttribute{
							Computed:    true,
							Description: "The availability zone ID of the cloud provider.",
						},
						"subnets": schema.ListAttribute{
							Computed:    true,
							Description: "The subnets under the corresponding availability zone for deploying the instance. Currently, only one subnet can be set for each availability zone.",
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.UniqueValues(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
						},
					},
				},
			},
			"compute_specs": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The compute specs of the instance, contains aku and version.",
				Attributes: map[string]schema.Attribute{
					"aku": schema.Int64Attribute{
						Computed:    true,
						Description: "AutoMQ defines AKU (AutoMQ Kafka Unit) to measure the scale of the cluster. Each AKU provides 20 MiB/s of read/write throughput. For more details on AKU, please refer to the [documentation](https://docs.automq.com/automq-cloud/subscriptions-and-billings/byoc-env-billings/billing-instructions-for-byoc).",
					},
					"version": schema.StringAttribute{
						Computed:    true,
						Description: "The software version of AutoMQ instance.",
					},
				},
			},
			"configs": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Additional configuration for the Kafka Instance. The currently supported parameters can be set by referring to the [documentation](https://docs.automq.com/automq-cloud/using-automq-for-kafka/restrictions#instance-level-configuration).",
				Computed:            true,
			},
			"integrations": schema.ListAttribute{
				Computed:    true,
				Description: "List of All Integrations Associated with the Current Instance. AutoMQ supports integration with external products like `prometheus` and `cloudwatch`, forwarding instance Metrics data to Prometheus and CloudWatch.",
				ElementType: types.StringType,
			},
			"acl": schema.BoolAttribute{
				Computed:    true,
				Description: "The ACL status the Kafka instance.",
			},
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"last_updated": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"instance_status": schema.StringAttribute{
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
							Description: "The name of the endpoint",
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
							Description: "The bootstrap servers of the endpoint.",
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

	instance := models.KafkaInstanceResourceModel{}
	var out *client.KafkaInstanceResponse
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
		}
		if out == nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Kafka instance %q not found", instanceId), err.Error())
			return
		}
		if !config.Name.IsNull() && out.DisplayName != config.Name.ValueString() {
			resp.Diagnostics.AddError(
				"Name Mismatch",
				fmt.Sprintf("The Kafka instance name '%s' does not match the expected name '%s'.", out.DisplayName, config.Name.ValueString()),
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

	instanceId := out.InstanceID
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
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModel(out, &instance, integrations, endpoints)...)
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
	models.ConvertKafkaInstanceModel(&instance, &model)
	// Update the model with the configurations
	model.Configs = models.FlattenStringValueMap(configs)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
