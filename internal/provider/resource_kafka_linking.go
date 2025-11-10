package provider

import (
	"context"
	"fmt"
	"regexp"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &KafkaLinkingResource{}
var _ resource.ResourceWithImportState = &KafkaLinkingResource{}

func NewKafkaLinkingResource() resource.Resource {
	return &KafkaLinkingResource{}
}

// KafkaLinkingResource implements automq_kafka_linking.
type KafkaLinkingResource struct {
	client *client.Client
}

func (r *KafkaLinkingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_linking"
}

func (r *KafkaLinkingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	startOffsetRegex := regexp.MustCompile(`^(latest|earliest|[0-9]+)$`)
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Kafka links for mirroring topics and consumer groups between AutoMQ instances and external Kafka clusters.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment identifier.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"instance_id": schema.StringAttribute{
				MarkdownDescription: "Kafka instance identifier that owns the link.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"link_id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the Kafka link.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"start_offset_time": schema.StringAttribute{
				MarkdownDescription: "Start offset time for mirroring. Accepted values: `latest`, `earliest`, or a Unix timestamp in milliseconds.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(startOffsetRegex, "must be `latest`, `earliest`, or a numeric timestamp"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current status of the Kafka link reported by the control plane.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp of the link.",
				Computed:            true,
				CustomType:          timetypes.RFC3339Type{},
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "Last modification timestamp of the link.",
				Computed:            true,
				CustomType:          timetypes.RFC3339Type{},
			},
			"error_message": schema.StringAttribute{
				MarkdownDescription: "Latest error message reported for the link, if any.",
				Computed:            true,
			},
			"source_cluster": schema.SingleNestedAttribute{
				MarkdownDescription: "Inline configuration for the source Kafka cluster.",
				Required:            true,
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.RequiresReplace()},
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						MarkdownDescription: "Bootstrap servers of the source Kafka cluster (host:port list).",
						Required:            true,
					},
					"security_protocol": schema.StringAttribute{
						MarkdownDescription: "Security protocol to use when connecting to the source cluster (e.g. `PLAINTEXT`, `SSL`, `SASL_SSL`).",
						Optional:            true,
					},
					"sasl_mechanism": schema.StringAttribute{
						MarkdownDescription: "SASL mechanism when SASL is enabled (e.g. `PLAIN`, `SCRAM_SHA_512`).",
						Optional:            true,
					},
					"user": schema.StringAttribute{
						MarkdownDescription: "SASL username when authentication is enabled.",
						Optional:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "SASL password for the source cluster.",
						Optional:            true,
						Sensitive:           true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"truststore_certificates": schema.StringAttribute{
						MarkdownDescription: "PEM encoded CA certificates for TLS connections.",
						Optional:            true,
						Sensitive:           true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"keystore_certificate_chain": schema.StringAttribute{
						MarkdownDescription: "PEM encoded client certificate chain for mTLS connections.",
						Optional:            true,
						Sensitive:           true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"keystore_key": schema.StringAttribute{
						MarkdownDescription: "PEM encoded private key for mTLS connections.",
						Optional:            true,
						Sensitive:           true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"disable_endpoint_identification": schema.BoolAttribute{
						MarkdownDescription: "Disable TLS endpoint identification when required by the source cluster.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
		},
	}
}

func (r *KafkaLinkingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KafkaLinkingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.KafkaLinkingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	param, diags := models.ExpandKafkaLinkCreateParam(&plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.CreateKafkaLink(ctx, plan.InstanceID.ValueString(), param)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Kafka link %q: %s", plan.LinkID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(models.FlattenKafkaLink(out, &plan, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Trace(ctx, "created kafka link resource", map[string]any{"link_id": plan.LinkID.ValueString()})
}

func (r *KafkaLinkingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.KafkaLinkingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	link, err := r.client.GetKafkaLink(ctx, state.InstanceID.ValueString(), state.LinkID.ValueString())
	if err != nil {
		if framework.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Kafka link %q: %s", state.LinkID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(models.FlattenKafkaLink(link, &state, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaLinkingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "automq_kafka_linking does not support in-place updates. Please recreate the resource after making configuration changes.")
}

func (r *KafkaLinkingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.KafkaLinkingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	if err := r.client.DeleteKafkaLink(ctx, state.InstanceID.ValueString(), state.LinkID.ValueString()); err != nil {
		if framework.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Kafka link %q: %s", state.LinkID.ValueString(), err))
		return
	}
}

func (r *KafkaLinkingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid Import ID", fmt.Sprintf("Expected <environment_id>@<instance_id>@<link_id>, got %q", req.ID)))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("link_id"), parts[2])...)
}
