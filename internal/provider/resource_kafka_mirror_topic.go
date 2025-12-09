package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &KafkaMirrorTopicResource{}
var _ resource.ResourceWithImportState = &KafkaMirrorTopicResource{}

func NewKafkaMirrorTopicResource() resource.Resource {
	return &KafkaMirrorTopicResource{}
}

// KafkaMirrorTopicResource manages mirrored topics for a Kafka link.
type KafkaMirrorTopicResource struct {
	client *client.Client
}

func (r *KafkaMirrorTopicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_mirror_topic"
}

func (r *KafkaMirrorTopicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage mirrored topics within a Kafka link.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"instance_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"link_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"source_topic_name": schema.StringAttribute{
				MarkdownDescription: "Topic name in the source Kafka cluster.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Desired mirror topic state (`LINKING`, `PAUSED`, or `PROMOTED`).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("LINKING", "PAUSED", "PROMOTED"),
				},
			},
			"mirror_topic_id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"mirror_topic_name": schema.StringAttribute{
				Computed: true,
			},
			"error_code": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *KafkaMirrorTopicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *KafkaMirrorTopicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.KafkaMirrorTopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	in := models.BuildMirrorTopicCreateParam(&plan)
	result, err := r.client.CreateKafkaLinkMirrorTopics(ctx, plan.InstanceID.ValueString(), plan.LinkID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create mirror topic %q: %s", plan.SourceTopicName.ValueString(), err))
		return
	}

	desiredState := ""
	if !plan.State.IsNull() && !plan.State.IsUnknown() {
		desiredState = strings.ToUpper(plan.State.ValueString())
	}
	if desiredState != "" && desiredState != "LINKING" && result != nil {
		topic := findMirrorTopicBySource(result.Topics, plan.SourceTopicName.ValueString())
		if topic != nil && topic.MirrorTopicID != nil {
			updateParam := models.BuildMirrorTopicUpdateParam(&plan)
			if err := r.client.UpdateKafkaLinkMirrorTopic(ctx, plan.InstanceID.ValueString(), plan.LinkID.ValueString(), *topic.MirrorTopicID, updateParam); err != nil {
				resp.Diagnostics.AddWarning("Failed to update state", fmt.Sprintf("Mirror topic created but failed to update state: %s", err))
			}
		}
	}

	var state models.KafkaMirrorTopicResourceModel = plan
	resp.Diagnostics.Append(r.refreshMirrorTopicState(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if !resp.Diagnostics.HasError() {
		tflog.Trace(ctx, "created kafka mirror topic", map[string]any{"source_topic": plan.SourceTopicName.ValueString()})
	}
}

func (r *KafkaMirrorTopicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.KafkaMirrorTopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	resp.Diagnostics.Append(r.refreshMirrorTopicState(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.MirrorTopicID.IsNull() && state.MirrorTopicName.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaMirrorTopicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.KafkaMirrorTopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	desiredState := ""
	if !plan.State.IsNull() && !plan.State.IsUnknown() {
		desiredState = strings.ToUpper(plan.State.ValueString())
	}
	currentState := ""
	if !state.State.IsNull() && !state.State.IsUnknown() {
		currentState = strings.ToUpper(state.State.ValueString())
	}
	if desiredState != "" && desiredState != currentState {
		topicID := state.MirrorTopicID.ValueString()
		if topicID == "" {
			resp.Diagnostics.AddError("Missing Mirror Topic ID", "Cannot update mirror topic state because the mirror_topic_id is not set.")
			return
		}
		updateParam := models.BuildMirrorTopicUpdateParam(&plan)
		if err := r.client.UpdateKafkaLinkMirrorTopic(ctx, state.InstanceID.ValueString(), state.LinkID.ValueString(), topicID, updateParam); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update mirror topic %q: %s", state.SourceTopicName.ValueString(), err))
			return
		}
	}

	resp.Diagnostics.Append(r.refreshMirrorTopicState(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaMirrorTopicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.KafkaMirrorTopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	topicID := state.MirrorTopicID.ValueString()
	if topicID == "" {
		return
	}

	if err := r.client.DeleteKafkaLinkMirrorTopic(ctx, state.InstanceID.ValueString(), state.LinkID.ValueString(), topicID); err != nil {
		if framework.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete mirror topic %q: %s", state.SourceTopicName.ValueString(), err))
	}
}

func (r *KafkaMirrorTopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid Import ID", fmt.Sprintf("Expected <environment_id>@<instance_id>@<link_id>@<source_topic_name>, got %q", req.ID)))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("link_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_topic_name"), parts[3])...)
}

func (r *KafkaMirrorTopicResource) refreshMirrorTopicState(ctx context.Context, state *models.KafkaMirrorTopicResourceModel) diag.Diagnostics {
	query := map[string]string{
		"keyword": state.SourceTopicName.ValueString(),
		"page":    strconv.Itoa(1),
		"size":    strconv.Itoa(50),
	}

	resp, err := r.client.ListKafkaLinkMirrorTopics(ctx, state.InstanceID.ValueString(), state.LinkID.ValueString(), query)
	if err != nil {
		if framework.IsNotFoundError(err) {
			state.MirrorTopicID = types.StringNull()
			state.MirrorTopicName = types.StringNull()
			state.ErrorCode = types.StringNull()
			state.State = types.StringNull()
			return nil
		}
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list mirror topics: %s", err))}
	}

	if resp == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", "Failed to list mirror topics: empty response")}
	}
	topic := findMirrorTopicBySource(resp.List, state.SourceTopicName.ValueString())
	if topic == nil {
		state.MirrorTopicID = types.StringNull()
		state.MirrorTopicName = types.StringNull()
		state.ErrorCode = types.StringNull()
		state.State = types.StringNull()
		return nil
	}

	models.FlattenMirrorTopic(topic, state, state)
	return nil
}

func findMirrorTopicBySource(topics []client.MirrorTopicVO, source string) *client.MirrorTopicVO {
	for i := range topics {
		if topics[i].SourceTopicName == source {
			return &topics[i]
		}
	}
	return nil
}
