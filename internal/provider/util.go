package provider

import (
	"fmt"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func isNotFoundError(err error) bool {
	condition, ok := err.(*client.ErrorResponse)
	return ok && condition.Code == 404
}

func ExpandKafkaInstanceResource(instance KafkaInstanceResourceModel, request *client.KafkaInstanceRequest) {
	request.DisplayName = instance.Name.ValueString()
	request.Description = instance.Description.ValueString()
	request.Provider = instance.CloudProvider.ValueString()
	request.Region = instance.Region.ValueString()
	request.Networks = make([]client.KafkaInstanceRequestNetwork, len(instance.Networks))
	request.Spec = client.KafkaInstanceRequestSpec{
		Template:    "aku",
		PaymentPlan: client.KafkaInstanceRequestPaymentPlan{PaymentType: "ON_DEMAND", Period: 1, Unit: "MONTH"},
		Values:      []client.KafkaInstanceRequestValues{{Key: "aku", Value: fmt.Sprintf("%d", instance.ComputeSpecs.Aku.ValueInt64())}},
	}
	request.Spec.Version = instance.ComputeSpecs.Version.ValueString()
	for i, network := range instance.Networks {
		request.Networks[i] = client.KafkaInstanceRequestNetwork{
			Zone:   network.Zone.ValueString(),
			Subnet: network.Subnet.ValueString(),
		}
	}
}

func FlattenKafkaInstanceModel(instance *client.KafkaInstanceResponse, resource *KafkaInstanceResourceModel) {
	resource.InstanceID = types.StringValue(instance.InstanceID)
	resource.Name = types.StringValue(instance.DisplayName)
	resource.Description = types.StringValue(instance.Description)
	resource.CloudProvider = types.StringValue(instance.Provider)
	resource.Region = types.StringValue(instance.Region)

	resource.Networks = flattenNetworks(instance.Networks)
	resource.ComputeSpecs = flattenComputeSpecs(instance.Spec)

	resource.CreatedAt = timetypes.NewRFC3339TimePointerValue(&instance.GmtCreate)
	resource.LastUpdated = timetypes.NewRFC3339TimePointerValue(&instance.GmtModified)

	resource.InstanceStatus = types.StringValue(instance.Status)
}

func flattenNetworks(networks []client.Network) []NetworkModel {
	networksModel := make([]NetworkModel, 0, len(networks))
	for _, network := range networks {
		zone := types.StringValue(network.Zone)
		for _, subnet := range network.Subnets {
			networksModel = append(networksModel, NetworkModel{
				Zone:   zone,
				Subnet: types.StringValue(subnet.Subnet),
			})
		}
	}
	return networksModel
}

func flattenComputeSpecs(spec client.Spec) ComputeSpecsModel {
	var aku types.Int64
	for _, value := range spec.Values {
		if value.Key == "aku" {
			aku = types.Int64Value(int64(value.Value))
			break
		}
	}
	return ComputeSpecsModel{
		Aku:     aku,
		Version: types.StringValue(spec.Version),
	}
}
