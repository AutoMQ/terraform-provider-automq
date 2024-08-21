package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"

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

func (r *KafkaTopicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_topic"
}

func (r *KafkaTopicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)<br><br>`automq_kafka_topic` provides a Kafka Topic resource that enables creating and deleting Kafka Topics on a Kafka cluster on AutoMQ BYOC environment.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment, this attribute is specified during the deployment and installation process.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"kafka_instance_id": schema.StringAttribute{
				MarkdownDescription: "Target Kafka instance ID, each instance represents a kafka cluster. The instance id looks like kf-xxxxxxx.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name is the unique identifier of a topic. It can only contain letters a to z or A to z, digits 0 to 9, underscores (_), hyphens (-), and dots (.). The value contains 1 to 249 characters.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 249)},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"partition": schema.Int64Attribute{
				MarkdownDescription: "Number of partitions for the Kafka topic. The valid range is 1-1024. The number of partitions must be at least greater than the number of consumers. The default value is 16.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(16),
				Validators:          []validator.Int64{int64validator.Between(1, 1024)},
			},
			"configs": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Additional configuration for the Kafka topic. Please refer to the [documentation](https://docs.automq.com/automq-cloud/using-automq-for-kafka/restrictions#topic-level-configuration) to set the current supported custom parameters.",
				Optional:            true,
			},
			"topic_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Kafka topic identifier, this id is generated by automq.",
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
	var topic models.KafkaTopicResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &topic)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, topic.EnvironmentID.ValueString())
	// Generate API request body from plan
	in := client.TopicCreateParam{}
	models.ExpandKafkaTopicResource(topic, &in)

	instanceId := topic.KafkaInstance.ValueString()

	out, err := r.client.CreateKafkaTopic(ctx, instanceId, in)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Kafka topic %q, got error: %s", topic.Name.ValueString(), err))
	}

	resp.Diagnostics.Append(models.FlattenKafkaTopic(out, &topic)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a Kafka topic resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &topic)...)
}

func (r *KafkaTopicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.KafkaTopicResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, data.EnvironmentID.ValueString())

	topicId := data.TopicID.ValueString()
	instanceId := data.KafkaInstance.ValueString()
	out, err := r.client.GetKafkaTopic(ctx, instanceId, topicId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got error: %s", topicId, err))
	}
	if out == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got nil response", topicId))
	}

	resp.Diagnostics.Append(models.FlattenKafkaTopic(out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaTopicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.KafkaTopicResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	instanceId := plan.KafkaInstance.ValueString()
	topicId := plan.TopicID.ValueString()
	planPartition := plan.Partition.ValueInt64()
	statePartition := state.Partition.ValueInt64()
	if planPartition != statePartition {
		if planPartition < statePartition {
			resp.Diagnostics.AddError("Partition Update Error", fmt.Sprintf("Error occurred while updating Kafka TopicId %q. "+
				" At present, we don't support reducing the number of partitions for a topic. ", topicId))
			return
		}

		in := client.TopicPartitionParam{}
		in.Partition = planPartition
		err := r.client.UpdateKafkaTopicPartition(ctx, instanceId, topicId, in)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka topic %q, got error: %s", topicId, err))
		}

		resp.Diagnostics.Append(ReadKafkaTopic(ctx, r, instanceId, topicId, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

	planConfig := plan.Configs
	stateConfig := state.Configs
	// check if the configs are different
	if !models.MapsEqual(planConfig, stateConfig) {
		// check if the planConfig has removed any config
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
		in.Configs = models.ExpandStringValueMap(planConfig)

		_, err := r.client.UpdateKafkaTopicConfig(ctx, instanceId, topicId, in)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka topic %q, got error: %s", topicId, err))
		}

		resp.Diagnostics.Append(ReadKafkaTopic(ctx, r, instanceId, topicId, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *KafkaTopicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.KafkaTopicResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	topicId := state.TopicID.ValueString()
	instanceId := state.KafkaInstance.ValueString()

	err := r.client.DeleteKafkaTopic(ctx, instanceId, topicId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Kafka topic %q, got error: %s", topicId, err))
	}
}

func (r *KafkaTopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("topic_id"), req, resp)
}

func ReadKafkaTopic(ctx context.Context, r *KafkaTopicResource, instanceId, topicId string, data *models.KafkaTopicResourceModel) diag.Diagnostics {
	out, err := r.client.GetKafkaTopic(ctx, instanceId, topicId)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got error: %s", topicId, err))}
	}
	if out == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get Kafka topic %q, got nil response", topicId))}
	}
	return models.FlattenKafkaTopic(out, data)
}
