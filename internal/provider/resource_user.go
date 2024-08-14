package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaUserResource{}
var _ resource.ResourceWithImportState = &KafkaUserResource{}

func NewKafkaUserResource() resource.Resource {
	return &KafkaUserResource{}
}

// KafkaUserResource defines the resource implementation.
type KafkaUserResource struct {
	client *client.Client
}

func (r *KafkaUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_user"
}

func (r *KafkaUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Kafka User resource",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target environment ID",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"kafka_instance_id": schema.StringAttribute{
				MarkdownDescription: "Target Kafka instance ID",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for the Kafka user",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 64),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the Kafka user",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(8, 24),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Kafka user identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *KafkaUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KafkaUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.KafkaUserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	instanceId := plan.KafkaInstanceID.ValueString()
	in := client.InstanceUserCreateParam{
		Name:     plan.Username.ValueString(),
		Password: plan.Password.ValueString(),
	}

	out, err := r.client.CreateKafkaUser(ctx, instanceId, in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Kafka user", err.Error())
		return
	}

	models.FlattenKafkaUserResource(out, &plan)
	tflog.Trace(ctx, "created a Kafka user resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KafkaUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.KafkaUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceId := data.KafkaInstanceID.ValueString()
	userName := data.Username.ValueString()

	out, err := r.client.GetKafkaUser(ctx, instanceId, userName)
	if err != nil {
		if framework.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Kafka user", err.Error())
		return
	}

	models.FlattenKafkaUserResource(out, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *KafkaUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.KafkaUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteKafkaUser(ctx, data.KafkaInstanceID.ValueString(), data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete Kafka user", err.Error())
		return
	}
}

func (r *KafkaUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
