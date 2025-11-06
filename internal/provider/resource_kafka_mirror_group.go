package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &KafkaMirrorGroupResource{}
var _ resource.ResourceWithImportState = &KafkaMirrorGroupResource{}

func NewKafkaMirrorGroupResource() resource.Resource {
	return &KafkaMirrorGroupResource{}
}

// KafkaMirrorGroupResource manages mirrored consumer groups on a Kafka link.
type KafkaMirrorGroupResource struct {
	client *client.Client
}

func (r *KafkaMirrorGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_mirror_group"
}

func (r *KafkaMirrorGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage mirrored consumer groups within a Kafka link.",
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
			"source_group_id": schema.StringAttribute{
				MarkdownDescription: "Consumer group identifier in the source Kafka cluster.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"mirror_group_id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"error_code": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *KafkaMirrorGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KafkaMirrorGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.KafkaMirrorGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	in := models.BuildMirrorGroupCreateParam(&plan)
	result, err := r.client.CreateKafkaLinkMirrorGroups(ctx, plan.InstanceID.ValueString(), plan.LinkID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create mirror group %q: %s", plan.SourceGroupID.ValueString(), err))
		return
	}

	var state models.KafkaMirrorGroupResourceModel = plan
	if result != nil {
		group := findMirrorGroupBySource(result.Groups, plan.SourceGroupID.ValueString())
		if group != nil {
			models.FlattenMirrorConsumerGroup(group, &state, &state)
		}
	}

	resp.Diagnostics.Append(r.refreshMirrorGroupState(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if !resp.Diagnostics.HasError() {
		tflog.Trace(ctx, "created kafka mirror group", map[string]any{"source_group": plan.SourceGroupID.ValueString()})
	}
}

func (r *KafkaMirrorGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.KafkaMirrorGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	resp.Diagnostics.Append(r.refreshMirrorGroupState(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.MirrorGroupID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaMirrorGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "automq_kafka_mirror_group does not support in-place updates. Recreate the resource to apply changes.")
}

func (r *KafkaMirrorGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.KafkaMirrorGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	groupID := state.MirrorGroupID.ValueString()
	if groupID == "" {
		return
	}

	if err := r.client.DeleteKafkaLinkMirrorGroup(ctx, state.InstanceID.ValueString(), state.LinkID.ValueString(), groupID); err != nil {
		if framework.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete mirror group %q: %s", state.SourceGroupID.ValueString(), err))
	}
}

func (r *KafkaMirrorGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid Import ID", fmt.Sprintf("Expected <environment_id>@<instance_id>@<link_id>@<source_group_id>, got %q", req.ID)))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("link_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_group_id"), parts[3])...)
}

func (r *KafkaMirrorGroupResource) refreshMirrorGroupState(ctx context.Context, state *models.KafkaMirrorGroupResourceModel) diag.Diagnostics {
	query := map[string]string{
		"keyword": state.SourceGroupID.ValueString(),
		"page":    strconv.Itoa(1),
		"size":    strconv.Itoa(50),
	}

	resp, err := r.client.ListKafkaLinkMirrorGroups(ctx, state.InstanceID.ValueString(), state.LinkID.ValueString(), query)
	if err != nil {
		if framework.IsNotFoundError(err) {
			state.MirrorGroupID = types.StringNull()
			state.State = types.StringNull()
			state.ErrorCode = types.StringNull()
			return nil
		}
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list mirror groups: %s", err))}
	}

	if resp == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", "Failed to list mirror groups: empty response")}
	}

	group := findMirrorGroupBySource(resp.List, state.SourceGroupID.ValueString())
	if group == nil {
		state.MirrorGroupID = types.StringNull()
		state.State = types.StringNull()
		state.ErrorCode = types.StringNull()
		return nil
	}

	models.FlattenMirrorConsumerGroup(group, state, state)
	return nil
}

func findMirrorGroupBySource(groups []client.MirrorConsumerGroupVO, source string) *client.MirrorConsumerGroupVO {
	for i := range groups {
		if groups[i].SourceGroupID == source {
			return &groups[i]
		}
	}
	return nil
}
