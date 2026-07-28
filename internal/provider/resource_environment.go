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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.Resource                = &EnvironmentResource{}
	_ resource.ResourceWithConfigure   = &EnvironmentResource{}
	_ resource.ResourceWithImportState = &EnvironmentResource{}
)

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

type EnvironmentResource struct {
	client *client.Client
}

func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	configuredClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = configuredClient
}

func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	immutable := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages an AutoMQ environment in the AutoMQ Cloud Control Plane.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Environment identifier assigned by AutoMQ.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Environment name.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 32)},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Environment description.",
				Validators:          []validator.String{stringvalidator.LengthAtMost(128)},
			},
			"cloud_provider": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloud provider code, such as `aws` or `gcp`.",
				PlanModifiers:       immutable,
			},
			"region": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloud region in which the environment is installed.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 32)},
				PlanModifiers:       immutable,
			},
			"scope": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloud account scope, such as an AWS account ID or Google Cloud project ID.",
				PlanModifiers:       immutable,
			},
			"product": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("kafka"),
				MarkdownDescription: "AutoMQ product deployed in the environment. Defaults to `kafka`.",
				PlanModifiers:       immutable,
			},
			"state":             schema.StringAttribute{Computed: true, MarkdownDescription: "Environment state: `Pending`, `Installed`, or `Active`."},
			"ops_bucket":        schema.StringAttribute{Computed: true, MarkdownDescription: "Operations bucket assigned to the environment."},
			"organization_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Organization identifier that owns the environment."},
			"organization_name": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization name that owns the environment."},
			"creator":           schema.StringAttribute{Computed: true, MarkdownDescription: "Identifier of the environment creator."},
			"creator_name":      schema.StringAttribute{Computed: true, MarkdownDescription: "Display name of the environment creator."},
			"stateless":         schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the environment is stateless."},
			"created_at":        schema.StringAttribute{Computed: true, MarkdownDescription: "Environment creation timestamp in RFC 3339 format."},
			"updated_at":        schema.StringAttribute{Computed: true, MarkdownDescription: "Environment update timestamp in RFC 3339 format."},
		},
	}
}

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	environment, err := r.client.CreateEnvironment(ctx, models.ExpandEnvironmentCreate(plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Environment Error", fmt.Sprintf("Unable to create environment: %s", err))
		return
	}
	if environment.EnvID == "" {
		resp.Diagnostics.AddError("Create Environment Error", "AutoMQ returned an empty environment ID.")
		return
	}
	models.FlattenEnvironment(environment, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	environment, err := r.client.GetEnvironment(ctx, state.ID.ValueString())
	if err != nil {
		if framework.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Environment Error", fmt.Sprintf("Unable to read environment %q: %s", state.ID.ValueString(), err))
		return
	}
	models.FlattenEnvironment(environment, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateEnvironment(ctx, plan.ID.ValueString(), models.ExpandEnvironmentUpdate(plan)); err != nil {
		resp.Diagnostics.AddError("Update Environment Error", fmt.Sprintf("Unable to update environment %q: %s", plan.ID.ValueString(), err))
		return
	}
	environment, err := r.client.GetEnvironment(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Environment Error", fmt.Sprintf("Unable to read environment %q after update: %s", plan.ID.ValueString(), err))
		return
	}
	models.FlattenEnvironment(environment, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEnvironment(ctx, state.ID.ValueString()); err != nil && !framework.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Delete Environment Error", fmt.Sprintf("Unable to delete environment %q: %s", state.ID.ValueString(), err))
	}
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
