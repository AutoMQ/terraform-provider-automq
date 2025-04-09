// automq_kafka_acl.go

package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaAclResource{}

func NewKafkaAclResource() resource.Resource {
	return &KafkaAclResource{}
}

// KafkaAclResource defines the resource implementation.
type KafkaAclResource struct {
	client *client.Client
}

func (r *KafkaAclResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_acl"
}

func (r *KafkaAclResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)<br><br>`automq_kafka_acl` provides an Access Control List (ACL) Policy in AutoMQ Cluster. AutoMQ supports ACL authorization for Cluster, Topic, Consumer Group, and Transaction ID resources, and simplifies the complex API actions of Apache Kafka through Operation Groups.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment, this attribute is specified during the deployment and installation process.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"kafka_instance_id": schema.StringAttribute{
				MarkdownDescription: "Target Kafka instance ID, each instance represents a kafka cluster. The instance id looks like kf-xxxxxxx.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The Kafka ACL Resource ID is returned upon successful creation of the ACL.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_type": schema.StringAttribute{
				MarkdownDescription: "The Kafka ACL authorized resource types, currently support `CLUSTER`, `TOPIC`, `GROUP` and `TRANSACTIONAL_ID`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("TOPIC", "GROUP", "CLUSTER", "TRANSACTIONAL_ID"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: "The target resource name for Kafka ACL authorization, can be a specific resource name or a resource name prefix (when using prefix matching, only the prefix needs to be provided without ending with `*`). If only `*` is specified, it represents all resources.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"pattern_type": schema.StringAttribute{
				MarkdownDescription: "Set the resource name matching pattern, supporting `LITERAL` and `PREFIXED`. `LITERAL` represents exact matching, while `PREFIXED` represents prefix matching.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("LITERAL", "PREFIXED"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"principal": schema.StringAttribute{
				MarkdownDescription: "Set the authorized target principal, which currently supports Kafka User type principals, i.e., `User:xxxx`. Specify the Kafka user name. Principal must start with `User:` and contact with `kafka_user`.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"operation_group": schema.StringAttribute{
				MarkdownDescription: "Set the authorized operation group. For the Topic resource type, the supported operations are `ALL` (all permissions), `PRODUCE` (produce messages only), and `CONSUME` (consume messages only). For other resource types, only `ALL` (all permissions) is supported.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("ALL", "PRODUCE", "CONSUME"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"permission": schema.StringAttribute{
				MarkdownDescription: "Set the permission type, which supports `ALLOW` and `DENY`. Default value is `ALLOW`. `ALLOW` grants permission to perform the operation, while `DENY` prohibits the operation. `DENY` takes precedence over `ALLOW`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ALLOW"),
				Validators: []validator.String{
					stringvalidator.OneOf("ALLOW", "DENY"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *KafkaAclResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KafkaAclResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.KafkaAclResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	// check Param
	if plan.PatternType.ValueString() == "CLUSTER" {
		if plan.ResourceName.ValueString() != "kafka-cluster" {
			resp.Diagnostics.AddError("Invalid Resource Name", "When resource type is CLUSTER, the resource name must be kafka-cluster.")
			return
		}
	}

	instance := plan.KafkaInstance.ValueString()
	param := client.KafkaAclBindingParam{}
	// Expand the Kafka ACL resource
	models.ExpandKafkaACLResource(plan, &param)
	in := client.KafkaAclBindingParams{Params: []client.KafkaAclBindingParam{param}}
	out, err := r.client.CreateKafkaAcls(ctx, instance, in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Create Kafka ACL", err.Error())
		return
	}
	// flatten the response and set the ID to the state
	resp.Diagnostics.Append(models.FlattenKafkaACLResource(out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a Kafka ACL resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KafkaAclResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.KafkaAclResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	aclId := state.ID.ValueString()
	instance := state.KafkaInstance.ValueString()
	out, err := r.client.GetKafkaAcls(ctx, instance, aclId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Kafka ACL", err.Error())
		return
	}
	// flatten the response and set the state
	resp.Diagnostics.Append(models.FlattenKafkaACLResource(out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaAclResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *KafkaAclResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.KafkaAclResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, data.EnvironmentID.ValueString())
	param := client.KafkaAclBindingParam{}
	models.ExpandKafkaACLResource(data, &param)
	in := client.KafkaAclBindingParams{Params: []client.KafkaAclBindingParam{param}}
	instance := data.KafkaInstance.ValueString()
	err := r.client.DeleteKafkaAcls(ctx, instance, in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete Kafka ACL", err.Error())
		return
	}
}
