// automq_integration.go

package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)<br><br>AutoMQ uses `automq_integration` to describe external third-party data transmission. By creating integrations and associating them with AutoMQ instances, you can forward instance Metrics and other data to external systems. Currently supported integration types are Prometheus and CloudWatch.",

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
				MarkdownDescription: "Type of integration, currently support `kafka` and `cloudwatch`",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("prometheus", "kafka", "cloudWatch"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Endpoint of integration. When selecting Prometheus and Kafka integration, you need to configure the corresponding endpoints. For detailed configuration instructions, please refer to the [documentation](https://docs.automq.com/automq-cloud/manage-environments/byoc-environment/manage-integrations).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"kafka_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Kafka integration configurations. When Type is `kafka`, it must be set.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"security_protocol": schema.StringAttribute{
						MarkdownDescription: "Security protocol for external kafka cluster, currently support `PLAINTEXT` and `SASL_PLAINTEXT`",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("PLAINTEXT", "SASL_PLAINTEXT"),
						},
					},
					"sasl_mechanism": schema.StringAttribute{
						MarkdownDescription: "SASL mechanism for external kafka cluster, currently support `PLAIN`, `SCRAM-SHA-256` and `SCRAM-SHA-512`",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"),
						},
					},
					"sasl_username": schema.StringAttribute{
						MarkdownDescription: "SASL username for Kafka, The username and password are declared and returned when creating the kafka_user resource in AutoMQ.",
						Required:            true,
					},
					"sasl_password": schema.StringAttribute{
						MarkdownDescription: "SASL password for Kafka, The username and password are declared and returned when creating the kafka_user resource in AutoMQ.",
						Required:            true,
					},
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

	// Generate API request to create the integration.
	in := client.IntegrationParam{}
	resp.Diagnostics.Append(models.ExpandIntergationResource(&in, integration))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.CreateIntergration(ctx, in)
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

	IntegrationId := data.ID.ValueString()
	err := r.client.DeleteIntergration(ctx, IntegrationId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete integration", err.Error())
		return
	}
}

func (r *IntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
