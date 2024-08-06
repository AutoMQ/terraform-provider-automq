package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	stateCreating   = "Creating"
	stateAvailable  = "Available"
	stateChanging   = "Changing"
	stateDeleting   = "Deleting"
	stateError      = "Error"
	stateUnexpected = "Unexpected"
	stateUnknown    = "Unknown"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaInstanceResource{}
var _ resource.ResourceWithImportState = &KafkaInstanceResource{}

func NewKafkaInstanceResource() resource.Resource {
	return &KafkaInstanceResource{}
}

// KafkaInstanceResource defines the resource implementation.
type KafkaInstanceResource struct {
	client *client.Client
}

// KafkaInstanceResourceModel describes the resource data model.
type KafkaInstanceResourceModel struct {
	InstanceID    types.String      `tfsdk:"instance_id"`
	DisplayName   types.String      `tfsdk:"display_name"`
	Description   types.String      `tfsdk:"description"`
	CloudProvider types.String      `tfsdk:"cloud_provider"`
	Region        types.String      `tfsdk:"region"`
	NetworkType   types.String      `tfsdk:"network_type"`
	Networks      []NetworkModel    `tfsdk:"networks"`
	ComputeSpecs  ComputeSpecsModel `tfsdk:"compute_specs"`
}

type NetworkModel struct {
	Zone   types.String `tfsdk:"zone"`
	Subnet types.String `tfsdk:"subnet"`
}

type ComputeSpecsModel struct {
	Aku types.Int64 `tfsdk:"aku"`
}

func (r *KafkaInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_instance"
}

func (r *KafkaInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AutoMQ Kafka instance resource",

		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Kafka instance",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the Kafka instance",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the Kafka instance",
				Optional:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "The cloud provider of the Kafka instance",
				Required:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region of the Kafka instance",
				Required:            true,
			},
			"network_type": schema.StringAttribute{
				MarkdownDescription: "The network type of the Kafka instance",
				Required:            true,
			},
			"networks": schema.ListNestedAttribute{
				Required:    true,
				Description: "The networks of the Kafka instance",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone": schema.StringAttribute{
							Required:    true,
							Description: "The zone of the network",
						},
						"subnet": schema.StringAttribute{
							Required:    true,
							Description: "The subnetId of the network",
						},
					},
				},
			},
			"compute_specs": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The compute specs of the Kafka instance",
				Attributes: map[string]schema.Attribute{
					"aku": schema.Int64Attribute{
						Required:    true,
						Description: "The template of the compute specs",
					},
				},
			},
		},
	}
}

func (r *KafkaInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *KafkaInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KafkaInstanceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	request := client.KafkaInstanceRequest{
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
		Provider:    data.CloudProvider.ValueString(),
		Region:      data.Region.ValueString(),
		Networks:    make([]client.KafkaInstanceRequestNetwork, len(data.Networks)),
		Spec: client.KafkaInstanceRequestSpec{
			Template:    "aku",
			PaymentPlan: client.KafkaInstanceRequestPaymentPlan{PaymentType: "ON_DEMAND", Period: 1, Unit: "MONTH"},
			Values:      []client.KafkaInstanceRequestValues{{Key: "aku", Value: fmt.Sprintf("%d", data.ComputeSpecs.Aku.ValueInt64())}},
		},
	}
	for i, network := range data.Networks {
		request.Networks[i] = client.KafkaInstanceRequestNetwork{
			Zone:   network.Zone.ValueString(),
			Subnet: network.Subnet.ValueString(),
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating new Kafka Cluster: %s", fmt.Sprintf("%v", request)))

	apiResp, err := r.client.CreateKafkaInstance(request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Kafka instance, got error: %s", err))
		return
	}

	instanceId := apiResp.InstanceID
	data.InstanceID = types.StringValue(instanceId)

	if err := waitForKafkaClusterToProvision(ctx, r.client, instanceId, data.CloudProvider.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, createDescriptiveError(err)))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KafkaInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KafkaInstanceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KafkaInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *KafkaInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
