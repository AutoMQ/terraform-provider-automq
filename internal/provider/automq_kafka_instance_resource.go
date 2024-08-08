package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	stateCreating   = "Creating"
	stateAvailable  = "Available"
	stateChanging   = "Changing"
	stateDeleting   = "Deleting"
	stateNotFound   = "NotFound"
	stateError      = "Error"
	stateUnexpected = "Unexpected"
	stateUnknown    = "Unknown"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaInstanceResource{}
var _ resource.ResourceWithImportState = &KafkaInstanceResource{}

func NewKafkaInstanceResource() resource.Resource {
	r := &KafkaInstanceResource{}
	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)
	return r
}

// KafkaInstanceResource defines the resource implementation.
type KafkaInstanceResource struct {
	client *client.Client
	framework.WithTimeouts
}

// KafkaInstanceResourceModel describes the resource data model.
type KafkaInstanceResourceModel struct {
	InstanceID     types.String       `tfsdk:"instance_id"`
	Name           types.String       `tfsdk:"name"`
	Description    types.String       `tfsdk:"description"`
	CloudProvider  types.String       `tfsdk:"cloud_provider"`
	Region         types.String       `tfsdk:"region"`
	NetworkType    types.String       `tfsdk:"network_type"`
	Networks       []NetworkModel     `tfsdk:"networks"`
	ComputeSpecs   ComputeSpecsModel  `tfsdk:"compute_specs"`
	Config         []ConfigModel      `tfsdk:"config"`
	ACL            types.Bool         `tfsdk:"acl"`
	Integrations   []IntegrationModel `tfsdk:"integrations"`
	CreatedAt      timetypes.RFC3339  `tfsdk:"created_at"`
	LastUpdated    types.String       `tfsdk:"last_updated"`
	InstanceStatus types.String       `tfsdk:"instance_status"`
	Timeouts       timeouts.Value     `tfsdk:"timeouts"`
}

type NetworkModel struct {
	Zone   types.String `tfsdk:"zone"`
	Subnet types.String `tfsdk:"subnet"`
}

type ComputeSpecsModel struct {
	Aku     types.Int64  `tfsdk:"aku"`
	Version types.String `tfsdk:"version"`
}

type KafkaInstanceConfig struct {
	Config []ConfigModel `tfsdk:"config"`
}

type ConfigModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type IntegrationModel struct {
	IntegrationID   types.String `tfsdk:"integration_id"`
	IntegrationType types.String `tfsdk:"integration_type"`
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
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Kafka instance",
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
					"version": schema.StringAttribute{
						Optional:    true,
						Default:     stringdefault.StaticString("latest"),
						Description: "The version of the compute specs",
					},
				},
			},
			"config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The config of the Kafka instance",
				Attributes: map[string]schema.Attribute{
					"key": schema.StringAttribute{
						Required:    true,
						Description: "The key of the config",
					},
					"value": schema.StringAttribute{
						Required:    true,
						Description: "The value of the config",
					},
				},
			},
			"integrations": schema.ListNestedAttribute{
				Optional:    true,
				Description: "The integrations of the Kafka instance",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"integration_id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the integration",
						},
						"integration_type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the integration",
						},
					},
				},
			},
			"acl": schema.BoolAttribute{
				Required:    false,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The ACL of the Kafka instance",
			},
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"last_updated": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"instance_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the Kafka instance",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"Timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
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
	var instance KafkaInstanceResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	in := client.KafkaInstanceRequest{}
	ExpandKafkaInstanceResource(instance, &in)
	tflog.Debug(ctx, fmt.Sprintf("Creating new Kafka Cluster: %s", fmt.Sprintf("%v", in)))

	out, err := r.client.CreateKafkaInstance(in)
	if resp.Diagnostics.HasError() {
		return
	}
	if err != nil {
		// TODO: Standard Error
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Kafka instance, got error: %s", err))
		return
	}
	if out == nil {
		// TODO: Standard Error
		resp.Diagnostics.AddError("Client Error", "Unable to create Kafka instance, got empty response")
		return
	}

	// Flatten API response into Terraform state
	FlattenKafkaInstanceModel(out, &instance)

	instanceId := instance.InstanceID.ValueString()

	createTimeout := r.CreateTimeout(ctx, instance.Timeouts)
	if err := waitForKafkaClusterToProvision(ctx, r.client, instanceId, stateCreating, createTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}

	instance.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &instance)...)
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
	if data.InstanceID.IsNull() && data.InstanceID.ValueString() == "" {
		resp.Diagnostics.AddError("Client Error", "Instance ID is required for updating Kafka instance")
		return
	}

	instance, err := GetKafkaInstance(&data, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", data.InstanceID.ValueString(), err))
		return
	}
	if instance == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q not found", data.InstanceID.ValueString()))
		return
	}

	if instance.Status != stateAvailable {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q is not in available state", data.InstanceID.ValueString()))
		return
	}

	instanceId := data.InstanceID.ValueString()

	// Generate API request body from plan
	basicUpdate := client.InstanceBasicParam{
		DisplayName: data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}
	specUpdate := client.SpecificationUpdateParam{
		Values: make([]client.KafkaInstanceRequestValues, 1),
	}
	specUpdate.Values[0] = client.KafkaInstanceRequestValues{
		Key:   "aku",
		Value: fmt.Sprintf("%d", data.ComputeSpecs.Aku.ValueInt64()),
	}

	_, err = r.client.UpdateKafkaInstanceBasicInfo(data.InstanceID.ValueString(), basicUpdate)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", data.InstanceID.ValueString(), err))
		return
	}

	if data.ComputeSpecs.Version.ValueString() != "" {
		_, err = r.client.UpdateKafkaInstanceVersion(data.InstanceID.ValueString(), data.ComputeSpecs.Version.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", data.InstanceID.ValueString(), err))
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, data.Timeouts)

	if err := waitForKafkaClusterToProvision(ctx, r.client, instanceId, stateChanging, updateTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}

	_, err = r.client.UpdateKafkaInstanceComputeSpecs(data.InstanceID.ValueString(), specUpdate)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", data.InstanceID.ValueString(), err))
		return
	}
	if err := waitForKafkaClusterToProvision(ctx, r.client, instanceId, stateChanging, updateTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KafkaInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	instanceId := state.InstanceID.ValueString()
	instance, err := GetKafkaInstance(&state, r.client)
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	if instance == nil {
		return
	}

	if instance.Status != stateDeleting {
		err = r.client.DeleteKafkaInstance(instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Kafka instance %q, got error: %s", instanceId, err))
			return
		}
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if err := waitForKafkaClusterToDeleted(ctx, r.client, instanceId, deleteTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}
}

func GetKafkaInstance(instance *KafkaInstanceResourceModel, client *client.Client) (*client.KafkaInstanceResponse, error) {
	kafka, err := client.GetKafkaInstanceByName(instance.Name.ValueString())
	if err != nil {
		return nil, fmt.Errorf("error getting Kafka instance by name %s: %v", instance.Name.ValueString(), err)
	}
	return kafka, nil
}

func (r *KafkaInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
