package models

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	StateCreating  = "Creating"
	StateAvailable = "Available"
	StateChanging  = "Changing"
	StateDeleting  = "Deleting"
	StateNotFound  = "NotFound"
	StateError     = "Error"
	StateUnknown   = "Unknown"
)

// KafkaInstanceResourceModel describes the resource data model.
type KafkaInstanceResourceModel struct {
	EnvironmentID  types.String       `tfsdk:"environment_id"`
	InstanceID     types.String       `tfsdk:"id"`
	Name           types.String       `tfsdk:"name"`
	Description    types.String       `tfsdk:"description"`
	CloudProvider  types.String       `tfsdk:"cloud_provider"`
	Region         types.String       `tfsdk:"region"`
	Networks       []NetworkModel     `tfsdk:"networks"`
	ComputeSpecs   *ComputeSpecsModel `tfsdk:"compute_specs"`
	Configs        types.Map          `tfsdk:"configs"`
	ACL            types.Bool         `tfsdk:"acl"`
	Integrations   types.List         `tfsdk:"integrations"`
	Endpoints      types.List         `tfsdk:"endpoints"`
	CreatedAt      timetypes.RFC3339  `tfsdk:"created_at"`
	LastUpdated    timetypes.RFC3339  `tfsdk:"last_updated"`
	InstanceStatus types.String       `tfsdk:"status"`
	Timeouts       timeouts.Value     `tfsdk:"timeouts"`
}

type KafkaInstanceModel struct {
	EnvironmentID  types.String       `tfsdk:"environment_id"`
	InstanceID     types.String       `tfsdk:"id"`
	Name           types.String       `tfsdk:"name"`
	Description    types.String       `tfsdk:"description"`
	CloudProvider  types.String       `tfsdk:"cloud_provider"`
	Region         types.String       `tfsdk:"region"`
	Networks       []NetworkModel     `tfsdk:"networks"`
	ComputeSpecs   *ComputeSpecsModel `tfsdk:"compute_specs"`
	Configs        types.Map          `tfsdk:"configs"`
	ACL            types.Bool         `tfsdk:"acl"`
	Integrations   types.List         `tfsdk:"integrations"`
	Endpoints      types.List         `tfsdk:"endpoints"`
	CreatedAt      timetypes.RFC3339  `tfsdk:"created_at"`
	LastUpdated    timetypes.RFC3339  `tfsdk:"last_updated"`
	InstanceStatus types.String       `tfsdk:"status"`
}

type InstanceAccessInfo struct {
	DisplayName      types.String `tfsdk:"display_name"`
	NetworkType      types.String `tfsdk:"network_type"`
	Protocol         types.String `tfsdk:"protocol"`
	Mechanisms       types.String `tfsdk:"mechanisms"`
	BootstrapServers types.String `tfsdk:"bootstrap_servers"`
}

type NetworkModel struct {
	Zone    types.String `tfsdk:"zone"`
	Subnets types.List   `tfsdk:"subnets"`
}

type ComputeSpecsModel struct {
	Aku     types.Int64  `tfsdk:"aku"`
	Version types.String `tfsdk:"version"`
}

func ExpandKafkaInstanceResource(instance KafkaInstanceResourceModel, request *client.KafkaInstanceRequest) {
	request.DisplayName = instance.Name.ValueString()
	request.Description = instance.Description.ValueString()
	request.Provider = mapCloudProviderValue(instance.CloudProvider)
	request.Region = instance.Region.ValueString()
	request.Networks = make([]client.KafkaInstanceRequestNetwork, len(instance.Networks))
	request.Spec = client.KafkaInstanceRequestSpec{
		Template:    "aku",
		PaymentPlan: client.KafkaInstanceRequestPaymentPlan{PaymentType: "ON_DEMAND", Period: 1, Unit: "MONTH"},
		Values:      []client.ConfigItemParam{{Key: "aku", Value: fmt.Sprintf("%d", instance.ComputeSpecs.Aku.ValueInt64())}},
	}
	request.Spec.Version = instance.ComputeSpecs.Version.ValueString()
	for i, network := range instance.Networks {
		subnetList := ExpandStringValueList(network.Subnets)
		for _, subnet := range subnetList {
			request.Networks[i] = client.KafkaInstanceRequestNetwork{
				Zone:   network.Zone.ValueString(),
				Subnet: subnet,
			}
		}
	}
	request.InstanceConfig = client.InstanceConfigParam{}
	request.InstanceConfig.Configs = ExpandStringValueMap(instance.Configs)
	request.Integrations = ExpandStringValueList(instance.Integrations)
	request.AclEnabled = instance.ACL.ValueBool()
}

func ConvertKafkaInstanceModel(resource *KafkaInstanceResourceModel, model *KafkaInstanceModel) {
	model.EnvironmentID = resource.EnvironmentID
	model.InstanceID = resource.InstanceID
	model.Name = resource.Name
	model.Description = resource.Description
	model.CloudProvider = resource.CloudProvider
	model.Region = resource.Region
	model.Networks = resource.Networks
	model.ComputeSpecs = resource.ComputeSpecs
	model.Configs = resource.Configs
	model.ACL = resource.ACL
	model.Integrations = resource.Integrations
	model.Endpoints = resource.Endpoints
	model.CreatedAt = resource.CreatedAt
	model.LastUpdated = resource.LastUpdated
	model.InstanceStatus = resource.InstanceStatus
}

func FlattenKafkaInstanceModel(instance *client.KafkaInstanceResponse, resource *KafkaInstanceResourceModel, integrations []client.IntegrationVO, endpoints []client.InstanceAccessInfoVO) diag.Diagnostics {
	resource.InstanceID = types.StringValue(instance.InstanceID)
	resource.Name = types.StringValue(instance.DisplayName)
	resource.Description = types.StringValue(instance.Description)
	resource.CloudProvider = mapCloudProvider(instance.Provider)
	resource.Region = types.StringValue(instance.Region)
	resource.ACL = types.BoolValue(instance.AclEnabled)
	networks, diag := flattenNetworks(instance.Networks)
	if diag.HasError() {
		return diag
	}
	resource.Networks = networks
	resource.ComputeSpecs = flattenComputeSpecs(instance.Spec)

	resource.CreatedAt = timetypes.NewRFC3339TimePointerValue(&instance.GmtCreate)
	resource.LastUpdated = timetypes.NewRFC3339TimePointerValue(&instance.GmtModified)

	resource.InstanceStatus = mapInstanceStatus(instance.Status)
	if integrations != nil && len(integrations) > 0 {
		integrationIds := make([]attr.Value, 0, len(integrations))
		for _, integration := range integrations {
			integrationIds = append(integrationIds, types.StringValue(integration.Code))
		}
		resource.Integrations = types.ListValueMust(types.StringType, integrationIds)
	}
	if endpoints != nil {
		diags := populateInstanceAccessInfoList(context.Background(), resource, endpoints)
		if diags.HasError() {
			return diags
		}
	}
	return nil
}

func populateInstanceAccessInfoList(ctx context.Context, data *KafkaInstanceResourceModel, in []client.InstanceAccessInfoVO) diag.Diagnostics {
	var diags diag.Diagnostics

	instanceAccessInfoList := make([]InstanceAccessInfo, len(in))

	for i, item := range in {
		instanceAccessInfoList[i] = InstanceAccessInfo{
			DisplayName:      types.StringValue(item.DisplayName),
			NetworkType:      types.StringValue(item.NetworkType),
			Protocol:         types.StringValue(item.Protocol),
			Mechanisms:       types.StringValue(item.Mechanisms),
			BootstrapServers: types.StringValue(item.BootstrapServers),
		}
	}
	data.Endpoints, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{
		"display_name":      types.StringType,
		"network_type":      types.StringType,
		"protocol":          types.StringType,
		"mechanisms":        types.StringType,
		"bootstrap_servers": types.StringType,
	}}, instanceAccessInfoList)
	return diags
}

func flattenNetworks(networks []client.Network) ([]NetworkModel, diag.Diagnostics) {
	networksModel := make([]NetworkModel, 0, len(networks))
	for _, network := range networks {
		zone := types.StringValue(network.Zone)
		subnets := make([]attr.Value, 0, len(network.Subnets))
		for _, subnet := range network.Subnets {
			subnets = append(subnets, types.StringValue(subnet.Subnet))
		}
		subnetList, diag := types.ListValue(types.StringType, subnets)
		if diag.HasError() {
			return nil, diag
		}
		networksModel = append(networksModel, NetworkModel{
			Zone:    zone,
			Subnets: subnetList,
		})
	}
	return networksModel, nil
}

func flattenComputeSpecs(spec client.Spec) *ComputeSpecsModel {
	var aku types.Int64
	for _, value := range spec.Values {
		if value.Key == "aku" {
			if intValue, ok := value.Value.(float64); ok {
				aku = types.Int64Value(int64(intValue))
			}
			break
		}
	}
	return &ComputeSpecsModel{
		Aku:     aku,
		Version: types.StringValue(spec.Version),
	}
}

func mapCloudProvider(provider string) types.String {
	switch provider {
	case "aliyun":
		return types.StringValue("alicloud")
	default:
		return types.StringValue(provider)
	}
}

func mapCloudProviderValue(provider types.String) string {
	switch provider.ValueString() {
	case "alicloud":
		return "aliyun"
	default:
		return provider.ValueString()
	}
}

func mapInstanceStatus(status string) types.String {
	switch status {
	case StateCreating:
		return types.StringValue(StateCreating)
	case StateAvailable:
		return types.StringValue("Running")
	case StateChanging:
		return types.StringValue(StateChanging)
	case StateDeleting:
		return types.StringValue(StateDeleting)
	case StateNotFound:
		return types.StringValue(StateNotFound)
	case StateError:
		return types.StringValue("Abnormal")
	default:
		return types.StringValue(StateUnknown)
	}
}
