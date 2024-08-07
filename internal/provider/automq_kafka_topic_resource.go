package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaTopicResource{}
var _ resource.ResourceWithImportState = &KafkaTopicResource{}

func NewKafkaTopicResource() resource.Resource {
	return &KafkaTopicResource{}
}

// KafkaTopicResource defines the resource implementation.
type KafkaTopicResource struct {
	client *http.Client
}

// KafkaTopicResourceModel describes the resource data model.
type KafkaTopicResourceModel struct {
	EnvironmentID   types.String `tfsdk:"environment_id"`
	KafkaInstance   types.String `tfsdk:"kafka_instance"`
	Name            types.String `tfsdk:"name"`
	Partition       types.Int64  `tfsdk:"partition"`
	CompactStrategy types.String `tfsdk:"compact_strategy"`
	Configs         types.Map    `tfsdk:"configs"`
	TopicID         types.String `tfsdk:"topic_id"`
}

func (r *KafkaTopicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_topic"
}

func (r *KafkaTopicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Kafka Topic resource",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target Kafka environment",
				Required:            true,
			},
			"kafka_instance": schema.StringAttribute{
				MarkdownDescription: "Target Kafka instance ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Kafka topic",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 249)},
			},
			"partition": schema.Int64Attribute{
				MarkdownDescription: "Number of partitions for the Kafka topic",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(16),
				Validators:          []validator.Int64{int64validator.Between(1, 1024)},
			},
			"compact_strategy": schema.StringAttribute{
				MarkdownDescription: "Compaction strategy for the Kafka topic",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf("DELETE", "COMPACT")},
			},
			"configs": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Additional configuration for the Kafka topic",
				Optional:            true,
			},
			"topic_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Kafka topic identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *KafkaTopicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *KafkaTopicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KafkaTopicResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to create the Kafka topic.
	// For the purposes of this example, we'll just generate an ID.

	data.TopicID = types.StringValue("generated-id")

	tflog.Trace(ctx, "created a Kafka topic resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaTopicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KafkaTopicResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to read the Kafka topic details.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaTopicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KafkaTopicResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to update the Kafka topic.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaTopicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KafkaTopicResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to delete the Kafka topic.
}

func (r *KafkaTopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
