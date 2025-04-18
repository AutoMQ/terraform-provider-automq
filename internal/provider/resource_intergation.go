// automq_integration.go

package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IntegrationResource{}
var _ resource.ResourceWithImportState = &IntegrationResource{}

func NewIntegrationResource() resource.Resource {
	return &IntegrationResource{}
}

// IntegrationResource defines the resource implementation.
type IntegrationResource struct {
	client *client.Client
}

func (r *IntegrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *IntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)\n\n" +
			"Using the `automq_integration` resource type, you can describe external third-party data transmission. " +
			"By creating integrations and associating them with AutoMQ instances, you can forward instance Metrics and other data to external systems. " +
			"Currently supported integration types are Prometheus and CloudWatch.\n\n" +
			"> **Note**: This provider version is only compatible with AutoMQ control plane versions 7.3.5 and later.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment, this attribute is specified during the deployment and installation process.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The integrated name identifies different configurations and contains 3 to 64 characters, including letters a to z or a to z, digits 0 to 9, underscores (_), and hyphens (-).",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 64)},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of integration, currently supports `prometheus_remote_write`, and `cloudwatch`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("prometheusRemoteWrite", "cloudWatch"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"deploy_profile": schema.StringAttribute{
				MarkdownDescription: "Deploy profile.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Endpoint of integration. When selecting Prometheus and Kafka integration, you need to configure the corresponding endpoints. For detailed configuration instructions, please refer to the [documentation](https://docs.automq.com/automq-cloud/manage-environments/byoc-environment/manage-integrations).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"cloudwatch_config": schema.SingleNestedAttribute{
				MarkdownDescription: "CloudWatch integration configurations. When Type is `cloudwatch`, it must be set.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"namespace": schema.StringAttribute{
						MarkdownDescription: "Set cloudwatch namespace, AutoMQ will write all Metrics data under this namespace. The namespace name must contain 1 to 255 valid ASCII characters and may be alphanumeric, periods, hyphens, underscores, forward slashes, pound signs, colons, and spaces, but not all spaces.",
						Optional:            true,
					},
				},
			},
			"prometheus_remote_write_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Prometheus remote write integration configurations. When Type is `prometheus_remote_write`, it must be set.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"auth_type": schema.StringAttribute{
						MarkdownDescription: "Authentication type, currently supports `noauth`, `basic`, `bearer`, and `sigv4`.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("noauth", "basic", "bearer", "sigv4"),
						},
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "Username for basic authentication. When authType is `basic`, it must be set.",
						Optional:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "Password for basic authentication. When authType is `basic`, it must be set.",
						Optional:            true,
					},
					"bearer_token": schema.StringAttribute{
						MarkdownDescription: "Bearer token for bearer authentication. When authType is `bearer`, it must be set.",
						Optional:            true,
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Integration identifier, Used for binding and association with the instance.",
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

func (r *IntegrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var integration models.IntegrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &integration)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, integration.EnvironmentID.ValueString())

	// Generate API request to create the integration.
	in := client.IntegrationParam{}
	resp.Diagnostics.Append(models.ExpandIntergationResource(&in, integration))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.CreateIntegration(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create integration", err.Error())
		return
	}

	models.FlattenIntergrationResource(out, &integration)

	tflog.Trace(ctx, "created an integration resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &integration)...)
}

func (r *IntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.IntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, data.EnvironmentID.ValueString())

	intergationId := data.ID.ValueString()
	out, err := r.client.GetIntergration(ctx, intergationId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read integration", err.Error())
		return
	}

	models.FlattenIntergrationResource(out, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan models.IntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	// Generate API request to update the integration.
	intergationId := state.ID.ValueString()
	in := client.IntegrationParam{}
	resp.Diagnostics.Append(models.ExpandIntergationResource(&in, plan))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.UpdateIntergration(ctx, intergationId, &in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update integration", err.Error())
		return
	}
	models.FlattenIntergrationResource(out, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.IntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, data.EnvironmentID.ValueString())

	IntegrationId := data.ID.ValueString()
	err := r.client.DeleteIntergration(ctx, IntegrationId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete integration", err.Error())
		return
	}
}

func (r *IntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "@")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.Append(
			diag.NewErrorDiagnostic(
				"Invalid Import ID",
				fmt.Sprintf("The import ID must be in the format <environment_id>@<integration_id>. Got: %s", req.ID),
			),
		)
		return
	}

	environmentID := idParts[0]
	integrationId := idParts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), environmentID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), integrationId)...)
}
