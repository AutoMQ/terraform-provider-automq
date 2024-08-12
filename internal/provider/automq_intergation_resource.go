// automq_integration.go

package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
var _ resource.Resource = &IntegrationResource{}
var _ resource.ResourceWithImportState = &IntegrationResource{}

func NewIntegrationResource() resource.Resource {
	return &IntegrationResource{}
}

// IntegrationResource defines the resource implementation.
type IntegrationResource struct {
	client *client.Client
}

// IntegrationResourceModel describes the resource data model.
type IntegrationResourceModel struct {
	EnvironmentID    types.String                 `tfsdk:"environment_id"`
	Name             types.String                 `tfsdk:"name"`
	Type             types.String                 `tfsdk:"type"`
	EndPoint         types.String                 `tfsdk:"endpoint"`
	ID               types.String                 `tfsdk:"id"`
	KafkaConfig      *KafkaIntegrationConfig      `tfsdk:"kafka_config"`
	PrometheusConfig *PrometheusIntegrationConfig `tfsdk:"prometheus_config"`
	CloudWatchConfig *CloudWatchIntegrationConfig `tfsdk:"cloudwatch_config"`
	CreatedAt        timetypes.RFC3339            `tfsdk:"created_at"`
	LastUpdated      timetypes.RFC3339            `tfsdk:"last_updated"`
}

type KafkaIntegrationConfig struct {
	SecurityProtocol types.String `tfsdk:"security_protocol"`
	SaslMechanism    types.String `tfsdk:"sasl_mechanism"`
	SaslUsername     types.String `tfsdk:"sasl_username"`
	SaslPassword     types.String `tfsdk:"sasl_password"`
}

type PrometheusIntegrationConfig struct {
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	BearerToken types.String `tfsdk:"bearer_token"`
}

type CloudWatchIntegrationConfig struct {
	NameSpace types.String `tfsdk:"namespace"`
}

func (r *IntegrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *IntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Integration resource",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target environment ID",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the integration",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 64)},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the integration",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("prometheus", "kafka", "cloudWatch"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Endpoint of the integration",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"kafka_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Kafka configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"security_protocol": schema.StringAttribute{
						MarkdownDescription: "Security protocol for Kafka",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("PLAINTEXT", "SASL_PLAINTEXT"),
						},
					},
					"sasl_mechanism": schema.StringAttribute{
						MarkdownDescription: "SASL mechanism for Kafka",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"),
						},
					},
					"sasl_username": schema.StringAttribute{
						MarkdownDescription: "SASL username for Kafka",
						Required:            true,
					},
					"sasl_password": schema.StringAttribute{
						MarkdownDescription: "SASL password for Kafka",
						Required:            true,
					},
				},
			},
			"prometheus_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Prometheus",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						MarkdownDescription: "Username",
						Optional:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "Password",
						Optional:            true,
					},
					"bearer_token": schema.StringAttribute{
						MarkdownDescription: "Bearer token",
						Optional:            true,
					},
				},
			},
			"cloudwatch_config": schema.SingleNestedAttribute{
				MarkdownDescription: "CloudWatch",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"namespace": schema.StringAttribute{
						MarkdownDescription: "Namespace",
						Optional:            true,
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Integration identifier",
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
	var integration IntegrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &integration)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request to create the integration.
	in := client.IntegrationParam{}
	resp.Diagnostics.Append(ExpandIntergationResource(&in, integration))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.CreateIntergration(in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create integration", err.Error())
		return
	}

	FlattenIntergrationResource(out, &integration)

	tflog.Trace(ctx, "created an integration resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &integration)...)
}

func (r *IntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	intergationId := data.ID.ValueString()
	out, err := r.client.GetIntergration(intergationId)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Failed to read integration", err.Error())
		return
	}
	FlattenIntergrationResource(out, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan IntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Generate API request to update the integration.
	intergationId := state.ID.ValueString()
	in := client.IntegrationParam{}
	resp.Diagnostics.Append(ExpandIntergationResource(&in, plan))
	in.Type = nil
	out, err := r.client.UpdateIntergration(intergationId, &in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update integration", err.Error())
		return
	}
	FlattenIntergrationResource(out, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	IntegrationId := data.ID.ValueString()
	err := r.client.DeleteIntergration(IntegrationId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete integration", err.Error())
		return
	}
}

func (r *IntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
