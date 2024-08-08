// automq_integration.go

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
var _ resource.Resource = &IntegrationResource{}
var _ resource.ResourceWithImportState = &IntegrationResource{}

func NewIntegrationResource() resource.Resource {
	return &IntegrationResource{}
}

// IntegrationResource defines the resource implementation.
type IntegrationResource struct {
	client *http.Client
}

// IntegrationResourceModel describes the resource data model.
type IntegrationResourceModel struct {
	EnvironmentID    types.String                `tfsdk:"environment_id"`
	Name             types.String                `tfsdk:"name"`
	Type             types.String                `tfsdk:"type"`
	EndPoint         types.String                `tfsdk:"endpoint"`
	ID               types.String                `tfsdk:"id"`
	KafkaConfig      KafkaIntegrationConfig      `tfsdk:"kafka_config"`
	PrometheusConfig PrometheusIntegrationConfig `tfsdk:"prometheus_config"`
}

type KafkaIntegrationConfig struct {
	SecurityProtocol string `json:"security_protocol"`
	SaslMechanism    string `json:"sasl_mechanism"`
	SaslUsername     string `json:"sasl_username"`
	SaslPassword     string `json:"sasl_password"`
}

type PrometheusIntegrationConfig struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	BearerToken string `json:"bearer_token"`
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
					stringvalidator.OneOf("Prometheus", "Kafka", "CloudWatch"),
				},
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Endpoint of the integration",
				Required:            true,
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
							stringvalidator.OneOf("PLAINTEXT", "SSL", "SASL_PLAINTEXT", "SASL_SSL"),
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
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Integration identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *IntegrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IntegrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to create the integration.
	// For the purposes of this example, we'll just generate an ID.

	data.ID = types.StringValue("generated-id")

	tflog.Trace(ctx, "created an integration resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to read the integration details.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IntegrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to update the integration.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to delete the integration.
}

func (r *IntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
