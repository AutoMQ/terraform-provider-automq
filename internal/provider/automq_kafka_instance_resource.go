package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	stateNotFound   = "NotFound"
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
	InstanceID    types.String       `tfsdk:"instance_id"`
	Name          types.String       `tfsdk:"name"`
	Description   types.String       `tfsdk:"description"`
	CloudProvider types.String       `tfsdk:"cloud_provider"`
	Region        types.String       `tfsdk:"region"`
	NetworkType   types.String       `tfsdk:"network_type"`
	Networks      []NetworkModel     `tfsdk:"networks"`
	ComputeSpecs  ComputeSpecsModel  `tfsdk:"compute_specs"`
	Config        []ConfigModel      `tfsdk:"config"`
	ACL           types.Bool         `tfsdk:"acl"`
	Integrations  []IntegrationModel `tfsdk:"integrations"`
	LastUpdated   types.String       `tfsdk:"last_updated"`
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
			"last_updated": schema.StringAttribute{
				Computed: true,
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
		DisplayName: data.Name.ValueString(),
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

	if !data.ComputeSpecs.Version.IsUnknown() && !data.ComputeSpecs.Version.IsNull() {
		request.Spec.Version = data.ComputeSpecs.Version.ValueString()
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

	if err := waitForKafkaClusterToProvision(ctx, r.client, instanceId, data.CloudProvider.ValueString(), stateCreating); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

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

	if err := waitForKafkaClusterToProvision(ctx, r.client, instanceId, data.CloudProvider.ValueString(), stateChanging); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}

	_, err = r.client.UpdateKafkaInstanceComputeSpecs(data.InstanceID.ValueString(), specUpdate)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", data.InstanceID.ValueString(), err))
		return
	}
	if err := waitForKafkaClusterToProvision(ctx, r.client, instanceId, data.CloudProvider.ValueString(), stateChanging); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
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

	instance, err := GetKafkaInstance(&data, r.client)
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", data.InstanceID.ValueString(), err))
		return
	}
	if instance == nil {
		return
	}

	instanceId := data.InstanceID.ValueString()

	if instance.Status != stateDeleting {
		err = r.client.DeleteKafkaInstance(instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Kafka instance %q, got error: %s", data.InstanceID.ValueString(), err))
			return
		}
	}

	if err := waitForKafkaClusterToDeleted(ctx, r.client, instanceId, data.CloudProvider.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}
}

func GetKafkaInstance(data *KafkaInstanceResourceModel, client *client.Client) (*client.KafkaInstanceResponse, error) {
	if data.InstanceID.IsNull() && !data.Name.IsNull() {
		kafka, err := client.GetKafkaInstanceByName(data.Name.ValueString())
		if err != nil {
			return nil, fmt.Errorf("error getting Kafka instance by name %s: %v", data.Name.ValueString(), err)
		}
		return kafka, nil
	}
	if !data.InstanceID.IsNull() {
		kafka, err := client.GetKafkaInstance(data.InstanceID.ValueString())
		if err != nil {
			return nil, err
		}
		return kafka, nil
	}
	return nil, fmt.Errorf("both Kafka instance ID and name are null")
}

func (r *KafkaInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
