package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	client *client.Client
}

// KafkaTopicResourceModel describes the resource data model.
type KafkaTopicResourceModel struct {
	EnvironmentID   types.String      `tfsdk:"environment_id"`
	KafkaInstance   types.String      `tfsdk:"kafka_instance"`
	Name            types.String      `tfsdk:"name"`
	Partition       types.Int64       `tfsdk:"partition"`
	CompactStrategy types.String      `tfsdk:"compact_strategy"`
	Configs         types.Map         `tfsdk:"configs"`
	TopicID         types.String      `tfsdk:"topic_id"`
	CreatedAt       timetypes.RFC3339 `tfsdk:"created_at"`
	LastUpdated     timetypes.RFC3339 `tfsdk:"last_updated"`
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
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"last_updated": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (r *KafkaTopicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KafkaTopicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var topic KafkaTopicResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &topic)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	in := client.TopicCreateParam{}
	ExpandKafkaTopicResource(topic, &in)

	instanceId := topic.KafkaInstance.ValueString()

	out, err := r.client.CreateKafkaTopic(instanceId, in)
	if err != nil {
		if isNotFoundError(err) {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got error: %s", topic, err))
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got error: %s", topic, err))
	}
	if out == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got nil response", topic))
	}
	resp.Diagnostics.Append(FlattenKafkaTopic(out, &topic)...)
	if resp.Diagnostics.HasError() {
		return
	}

	now := time.Now()
	topic.CreatedAt = timetypes.NewRFC3339TimePointerValue(&now)
	topic.LastUpdated = timetypes.NewRFC3339TimePointerValue(&now)

	tflog.Trace(ctx, "created a Kafka topic resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &topic)...)
}

func (r *KafkaTopicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KafkaTopicResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	topicId := data.TopicID.ValueString()
	instanceId := data.KafkaInstance.ValueString()
	out, err := r.client.GetKafkaTopic(instanceId, topicId)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got error: %s", topicId, err))
	}
	if out == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got nil response", topicId))
	}
	resp.Diagnostics.Append(FlattenKafkaTopic(out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ReadKafkaTopic(r *KafkaTopicResource, instanceId, topicId string, data *KafkaTopicResourceModel) diag.Diagnostics {
	out, err := r.client.GetKafkaTopic(instanceId, topicId)
	if err != nil {
		if isNotFoundError(err) {
			return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got error: %s", topicId, err))}
		}
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got error: %s", topicId, err))}
	}
	if out == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got nil response", topicId))}
	}
	return FlattenKafkaTopic(out, data)
}

func (r *KafkaTopicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan KafkaTopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state KafkaTopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceId := plan.KafkaInstance.ValueString()
	topicId := plan.TopicID.ValueString()
	planPartition := plan.Partition.ValueInt64()
	statePartition := state.Partition.ValueInt64()
	if planPartition != statePartition {
		in := client.TopicPartitionParam{}
		in.Partition = planPartition
		_, err := r.client.UpdateKafkaTopicPartition(instanceId, topicId, in)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka topic %q, got error: %s", topicId, err))
		}
	}
	planConfig := plan.Configs
	stateConfig := state.Configs
	// diff planConfig and stateConfig
	if !mapsEqual(planConfig, stateConfig) {
		// 检查是否有被删除的配置
		for name := range stateConfig.Elements() {
			if _, ok := planConfig.Elements()[name]; !ok {
				resp.Diagnostics.AddError("Config Update Error", fmt.Sprintf("Error occurred while updating Kafka TopicId %q. "+
					" At present, we don't support the removal of topic settings from the 'configs' block, "+
					"meaning you can't reset to the topic's default settings. "+
					"As a workaround, you can find the default value and manually set the current value to match the default.", topicId))
				return
			}
		}
		in := client.TopicConfigParam{}
		in.Configs = make([]client.ConfigItemParam, len(planConfig.Elements()))
		i := 0
		for name, value := range planConfig.Elements() {
			config := value.(types.String)
			in.Configs[i] = client.ConfigItemParam{
				Key:   name,
				Value: config.ValueString(),
			}
			i += 1
		}
		_, err := r.client.UpdateKafkaTopicConfig(instanceId, topicId, in)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka topic %q, got error: %s", topicId, err))
		}
	}
	resp.Diagnostics.Append(ReadKafkaTopic(r, instanceId, topicId, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	now := time.Now()
	plan.LastUpdated = timetypes.NewRFC3339TimePointerValue(&now)
	plan.CreatedAt = state.CreatedAt

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KafkaTopicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KafkaTopicResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	topicId := state.TopicID.ValueString()
	instanceId := state.KafkaInstance.ValueString()

	err := r.client.DeleteKafkaTopic(instanceId, topicId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Kafka topic %q, got error: %s", topicId, err))
	}
}

func (r *KafkaTopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
