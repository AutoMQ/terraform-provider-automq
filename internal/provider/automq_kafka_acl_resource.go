// automq_kafka_acl.go

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaAclResource{}
var _ resource.ResourceWithImportState = &KafkaAclResource{}

func NewKafkaAclResource() resource.Resource {
	return &KafkaAclResource{}
}

// KafkaAclResource defines the resource implementation.
type KafkaAclResource struct {
	client *http.Client
}

// KafkaAclResourceModel describes the resource data model.
type KafkaAclResourceModel struct {
	EnvironmentID  types.String `tfsdk:"environment_id"`
	KafkaInstance  types.String `tfsdk:"kafka_instance"`
	ID             types.String `tfsdk:"id"`
	ResourceType   types.String `tfsdk:"resource_type"`
	ResourceName   types.String `tfsdk:"resource_name"`
	PatternType    types.String `tfsdk:"pattern_type"`
	Principal      types.String `tfsdk:"principal"`
	OperationGroup types.String `tfsdk:"operation_group"`
	Permission     types.String `tfsdk:"permission"`
}

func (r *KafkaAclResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_acl"
}

func (r *KafkaAclResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Kafka ACL resource",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target Kafka environment",
				Required:            true,
			},
			"kafka_instance": schema.StringAttribute{
				MarkdownDescription: "Target Kafka instance ID",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Kafka instance ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_type": schema.StringAttribute{
				MarkdownDescription: "Resource type for ACL",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("TOPIC", "CONSUMERGROUP", "CLUSTER", "TRANSACTION_ID"),
				},
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: "Name of the resource for ACL",
				Required:            true,
			},
			"pattern_type": schema.StringAttribute{
				MarkdownDescription: "Pattern type for resource",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("LITERAL", "PREFIXED"),
				},
			},
			"principal": schema.StringAttribute{
				MarkdownDescription: "Principal for ACL",
				Required:            true,
			},
			"operation_group": schema.StringAttribute{
				MarkdownDescription: "Operation group for ACL",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("ALL", "PRODUCE", "CONSUME"),
				},
			},
			"permission": schema.StringAttribute{
				MarkdownDescription: "Permission type for ACL",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ALLOW"),
				Validators: []validator.String{
					stringvalidator.OneOf("ALLOW", "DENY"),
				},
			},
		},
	}
}

func (r *KafkaAclResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KafkaAclResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KafkaAclResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to create the Kafka ACL.
	// For the purposes of this example, we'll just generate an ID.

	data.ID = types.StringValue("generated-id")

	tflog.Trace(ctx, "created a Kafka ACL resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaAclResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KafkaAclResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to read the Kafka ACL details.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaAclResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KafkaAclResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to update the Kafka ACL.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaAclResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KafkaAclResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Here you would typically call an API to delete the Kafka ACL.
}

func (r *KafkaAclResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
