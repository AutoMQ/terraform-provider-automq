package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaInstanceResource{}
var _ resource.ResourceWithConfigure = &KafkaInstanceResource{}
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

func (r *KafkaInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_instance"
}

func (r *KafkaInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AutoMQ Kafka instance resource",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target Kafka environment",
				Required:            true,
			},
			"id": schema.StringAttribute{
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
				Validators:          []validator.String{stringvalidator.OneOf("aws", "aws-cn", "aliyun")},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region of the Kafka instance",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"networks": schema.ListNestedAttribute{
				Required:    true,
				Description: "The networks of the Kafka instance",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone": schema.StringAttribute{
							Required:      true,
							Description:   "The zone of the network",
							PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"subnets": schema.ListAttribute{
							Required:    true,
							Description: "The subnets of the network",
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.UniqueValues(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
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
						Computed:    true,
						Description: "The version of the compute specs",
					},
				},
			},
			"configs": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Additional configuration for the Kafka topic",
				Optional:            true,
				Computed:            true,
			},
			"integrations": schema.ListAttribute{
				Optional:    true,
				Description: "The integrations of the Kafka instance",
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.SizeAtMost(1),
				},
			},
			"acl": schema.BoolAttribute{
				Optional:    true,
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
			"endpoints": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The endpoints of the Kafka instance",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "The display name of the endpoint",
						},
						"network_type": schema.StringAttribute{
							Computed:    true,
							Description: "The network type of the endpoint",
						},
						"protocol": schema.StringAttribute{
							Computed:    true,
							Description: "The protocol of the endpoint",
						},
						"mechanisms": schema.StringAttribute{
							Computed:    true,
							Description: "The mechanisms of the endpoint",
						},
						"bootstrap_servers": schema.StringAttribute{
							Computed:    true,
							Description: "The bootstrap servers of the endpoint",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
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
	var instance models.KafkaInstanceResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	in := client.KafkaInstanceRequest{}
	models.ExpandKafkaInstanceResource(instance, &in)
	tflog.Debug(ctx, fmt.Sprintf("Creating new Kafka Cluster: %s", fmt.Sprintf("%v", in)))

	out, err := r.client.CreateKafkaInstance(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Kafka instance, got error: %s", err))
		return
	}
	// Flatten API response into Terraform state
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModel(out, &instance, nil, nil, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceId := instance.InstanceID.ValueString()

	createTimeout := r.CreateTimeout(ctx, instance.Timeouts)
	if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateCreating, createTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}

	resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &instance)...)
}

func (r *KafkaInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.KafkaInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceId := state.InstanceID.ValueString()
	instance, err := r.client.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			// Treat HTTP 404 Not Found status as a signal to recreate resource and return early
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
	}
	// Get instance integrations
	integrations, err := r.client.ListInstanceIntegrations(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list integrations for Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
		return
	}
	// Get instance endpoints
	endpoints, err := r.client.GetInstanceEndpoints(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get endpoints for Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
		return
	}
	// Get instance configurations
	configs, err := r.client.GetInstanceConfigs(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get configurations for Kafka instance %q, got error: %s", state.InstanceID.ValueString(), err))
		return
	}

	// Flatten API response into Terraform state
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModel(instance, &state, integrations, endpoints, configs)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KafkaInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.KafkaInstanceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if the instance exists
	instanceId := plan.InstanceID.ValueString()
	instance, err := r.client.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	if instance == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q not found", instanceId))
		return
	}
	// check if the instance is in available state
	if instance.Status != models.StateAvailable {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Kafka instance %q is not in available state", instanceId))
		return
	}

	// Check if the basic info has changed
	if state.Name.ValueString() != plan.Name.ValueString() ||
		state.Description.ValueString() != plan.Description.ValueString() {
		// Generate API request body from plan
		basicUpdate := client.InstanceBasicParam{
			DisplayName: plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
		}
		_, err = r.client.UpdateKafkaInstanceBasicInfo(ctx, instanceId, basicUpdate)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", instanceId, err))
			return
		}
		// get latest info
		resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
	// Check if the ACL has changed
	if state.ACL.ValueBool() != plan.ACL.ValueBool() {

		if state.ACL.ValueBool() && plan.ACL.ValueBool() {
			resp.Diagnostics.AddError("Unsupported Operation", "Turning off ACL is not supported")
			return
		}

		err = r.client.TurnOnInstanceAcl(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to turn on ACL for Kafka instance %q, got error: %s", instanceId, err))
			return
		}
		// get latest info
		resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

	// Check if the Integrations has changed
	var isIntegrationChanged bool
	for _, i := range plan.Integrations.Elements() {
		found := false
		for _, integration := range state.Integrations.Elements() {
			if integration == i {
				found = true
				break
			}
		}
		if !found {
			isIntegrationChanged = true
			break
		}
	}
	if isIntegrationChanged {
		// Generate API request body from plan
		param := client.IntegrationInstanceParam{
			Codes: models.ExpandStringValueList(plan.Integrations),
		}
		err = r.client.ReplaceInstanceIntergation(ctx, instanceId, param)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update integrations for Kafka instance %q, got error: %s", instanceId, err))
			return
		}

		resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

	updateTimeout := r.UpdateTimeout(ctx, state.Timeouts)

	planConfig := plan.Configs
	stateConfig := state.Configs
	// check if the config has changed
	if !models.MapsEqual(planConfig, stateConfig) {
		// Check if the plan config has removed any settings
		for name := range stateConfig.Elements() {
			if _, ok := planConfig.Elements()[name]; !ok {
				resp.Diagnostics.AddError("Config Update Error", fmt.Sprintf("Error occurred while updating Kafka Instance %q. "+
					" At present, we don't support the removal of topic settings from the 'configs' block, "+
					"meaning you can't reset to the instance's default settings. "+
					"As a workaround, you can find the default value and manually set the current value to match the default.", instanceId))
				return
			}
		}

		in := client.InstanceConfigParam{}
		in.Configs = models.CreateConfigFromMapValue(planConfig)

		_, err := r.client.UpdateKafkaInstanceConfig(ctx, instanceId, in)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", instanceId, err))
		}

		// wait for version update
		if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateChanging, updateTimeout); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
			return
		}

		resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

	// Check if the compute specs (version) has changed
	planVersion := plan.ComputeSpecs.Version.ValueString()
	stateVersion := state.ComputeSpecs.Version.ValueString()
	if planVersion != "" && planVersion != stateVersion {
		_, err = r.client.UpdateKafkaInstanceVersion(ctx, state.InstanceID.ValueString(), planVersion)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", instanceId, err))
			return
		}
		// wait for version update
		if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateChanging, updateTimeout); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
			return
		}
		// get latest info
		resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

	stateAKU := state.ComputeSpecs.Aku.ValueInt64()
	planAKU := plan.ComputeSpecs.Aku.ValueInt64()
	if stateAKU != planAKU {
		// Generate API request body from plan
		specUpdate := client.SpecificationUpdateParam{
			Values: make([]client.ConfigItemParam, 1),
		}
		specUpdate.Values[0] = client.ConfigItemParam{
			Key:   "aku",
			Value: fmt.Sprintf("%d", planAKU),
		}
		_, err = r.client.UpdateKafkaInstanceComputeSpecs(ctx, instanceId, specUpdate)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kafka instance %q, got error: %s", instanceId, err))
			return
		}
		// wait for aku update
		if err := framework.WaitForKafkaClusterToProvision(ctx, r.client, instanceId, models.StateChanging, updateTimeout); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
			return
		}
		// get latest info
		resp.Diagnostics.Append(ReadKafkaInstance(ctx, r, instanceId, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Save updated data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *KafkaInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.KafkaInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	instanceId := state.InstanceID.ValueString()
	instance, err := r.client.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", instanceId, err))
		return
	}
	if instance == nil {
		return
	}

	if instance.Status != models.StateDeleting {
		err = r.client.DeleteKafkaInstance(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Kafka instance %q, got error: %s", instanceId, err))
			return
		}
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if err := framework.WaitForKafkaClusterToDeleted(ctx, r.client, instanceId, deleteTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for Kafka Cluster %q to provision: %s", instanceId, err))
		return
	}
}

func (r *KafkaInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func ReadKafkaInstance(ctx context.Context, r *KafkaInstanceResource, instanceId string, plan *models.KafkaInstanceResourceModel) diag.Diagnostics {
	instance, err := r.client.GetKafkaInstance(ctx, instanceId)
	if err != nil {
		if framework.IsNotFoundError(err) {
			return nil
		}
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get Kafka instance %q, got error: %s", plan.InstanceID.ValueString(), err))}
	}
	if instance == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Kafka instance %q not found", plan.InstanceID.ValueString()))}
	}

	integrations, err := r.client.ListInstanceIntegrations(ctx, instanceId)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list integrations for Kafka instance %q, got error: %s", plan.InstanceID.ValueString(), err))}
	}

	endpoints, err := r.client.GetInstanceEndpoints(ctx, instanceId)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get endpoints for Kafka instance %q, got error: %s", plan.InstanceID.ValueString(), err))}
	}
	// Get instance configurations
	configs, err := r.client.GetInstanceConfigs(ctx, instanceId)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to get configurations for Kafka instance %q, got error: %s", plan.InstanceID.ValueString(), err))}
	}

	return models.FlattenKafkaInstanceModel(instance, plan, integrations, endpoints, configs)
}
