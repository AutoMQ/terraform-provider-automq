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
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)<br><br>Using the `automq_kafka_instance` resource type, you can create and manage Kafka instances, where each instance represents a physical cluster.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment, this attribute is specified during the deployment and installation process.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Kafka instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Kafka instance. It can contain letters (a-z or A-Z), numbers (0-9), underscores (_), and hyphens (-), with a length limit of 3 to 64 characters.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The instance description are used to differentiate the purpose of the instance. They support letters (a-z or A-Z), numbers (0-9), underscores (_), spaces( ) and hyphens (-), with a length limit of 3 to 128 characters.",
				Optional:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "To set up a Kafka instance, you need to specify the target cloud provider environment for deployment. Currently, 'aws' is supported.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf("aws", "aws-cn", "aliyun")},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "To set up an instance, you need to specify the target region for deployment. Refer to the RegionId list provided by each cloud provider for available regions. Using AWS as an example, refer to this [documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html) to set the correct `RegionId`.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"networks": schema.ListNestedAttribute{
				Required:    true,
				Description: "To configure the network settings for an instance, you need to specify the availability zone(s) and subnet information. Currently, you can set either one availability zone or three availability zones.",
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.SizeBetween(1, 3),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone": schema.StringAttribute{
							Required:      true,
							Description:   "The availability zone ID of the cloud provider.",
							PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"subnets": schema.ListAttribute{
							Required:    true,
							Description: "Specify the subnet under the corresponding availability zone for deploying the instance. Currently, only one subnet can be set for each availability zone.",
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
				Description: "The compute specs of the instance, contains aku and version.",
				Attributes: map[string]schema.Attribute{
					"aku": schema.Int64Attribute{
						Required:    true,
						Description: "AutoMQ defines AKU (AutoMQ Kafka Unit) to measure the scale of the cluster. Each AKU provides 20 MiB/s of read/write throughput. For more details on AKU, please refer to the [documentation](https://docs.automq.com/automq-cloud/subscriptions-and-billings/byoc-env-billings/billing-instructions-for-byoc). The currently supported AKU specifications are 6, 8, 10, 12, 14, 16, 18, 20, 22, and 24. If an invalid AKU value is set, the instance cannot be created.",
					},
					"version": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The software version of AutoMQ instance. By default, there is no need to set version; the latest version will be used. If you need to specify a version, refer to the [documentation](https://docs.automq.com/automq-cloud/release-notes) to choose the appropriate version number.",
					},
				},
			},
			"configs": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Additional configuration for the Kafka Instance. The currently supported parameters can be set by referring to the [documentation](https://docs.automq.com/automq-cloud/release-notes).",
				Optional:            true,
			},
			"integrations": schema.ListAttribute{
				Optional:    true,
				Description: "Configure integration settings. AutoMQ supports integration with external products like `prometheus` and `cloudwatch`, forwarding instance Metrics data to Prometheus and CloudWatch.",
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
				Description: "Configure ACL enablement. Default is false (disabled).",
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
				MarkdownDescription: "The status of instance. Currently supports statuses: `Creating`, `Running`, `Deleting`, `Changing` and `Abnormal`. For definitions and limitations of each status, please refer to the [documentation](https://docs.automq.com/automq-cloud/using-automq-for-kafka/manage-instances#lifecycle).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoints": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The bootstrap endpoints of instance. AutoMQ supports multiple access protocols; therefore, the Endpoint is a list.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of endpoint",
						},
						"network_type": schema.StringAttribute{
							Computed:    true,
							Description: "The network type of endpoint. Currently support `VPC` and `INTERNET`. `VPC` type is generally used for internal network access, while `INTERNET` type is used for accessing the AutoMQ cluster from the internet.",
						},
						"protocol": schema.StringAttribute{
							Computed:    true,
							Description: "The protocol of endpoint. Currently support `PLAINTEXT` and `SASL_PLAINTEXT`.",
						},
						"mechanisms": schema.StringAttribute{
							Computed:    true,
							Description: "The supported mechanisms of endpoint. Currently support `PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512`.",
						},
						"bootstrap_servers": schema.StringAttribute{
							Computed:    true,
							Description: "The bootstrap servers of endpoint.",
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
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModel(out, &instance, nil, nil)...)
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
	// Flatten API response into Terraform state
	resp.Diagnostics.Append(models.FlattenKafkaInstanceModel(instance, &state, integrations, endpoints)...)
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
	return models.FlattenKafkaInstanceModel(instance, plan, integrations, endpoints)
}
