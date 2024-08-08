package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	client *http.Client
}

// KafkaUserResourceModel describes the resource data model.
type KafkaUserResourceModel struct {
	EnvironmentID   types.String `tfsdk:"environment_id"`
	KafkaInstanceID types.String `tfsdk:"kafka_instance_id"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
	ID              types.String `tfsdk:"id"`
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
			},
			"kafka_instance_id": schema.StringAttribute{
				MarkdownDescription: "Target Kafka instance ID",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for the Kafka user",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 64),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the Kafka user",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(8, 24),
				},
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

func (r *KafkaUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KafkaUserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to create the Kafka user.
	// For the purposes of this example, we'll just generate an ID.

	data.ID = types.StringValue("generated-id")

	tflog.Trace(ctx, "created a Kafka user resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KafkaUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to read the Kafka user details.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KafkaUserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to update the Kafka user.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KafkaUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to delete the Kafka user.
}

func (r *KafkaUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
