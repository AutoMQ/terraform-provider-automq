package provider

import (
	"fmt"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func isNotFoundError(err error) bool {
	condition, ok := err.(*client.ErrorResponse)
	return ok && condition.Code == 404
}

func mapsEqual(a, b types.Map) bool {
	if len(a.Elements()) != len(b.Elements()) {
		return false
	}
	for k, v := range a.Elements() {
		if bVal, ok := b.Elements()[k]; !ok {
			return false
		} else {
			aVal := v.(types.String)
			bVal := bVal.(types.String)

			if aVal.ValueString() != bVal.ValueString() {
				return false
			}
		}
	}
	return true
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
	request.Integrations = make([]string, len(instance.Integrations))
	for i, integration := range instance.Integrations {
		request.Integrations[i] = integration.IntegrationID.String()
	}
}

func FlattenKafkaInstanceModel(instance *client.KafkaInstanceResponse, resource *KafkaInstanceResourceModel) {
	resource.InstanceID = types.StringValue(instance.InstanceID)
	resource.Name = types.StringValue(instance.DisplayName)
	resource.Description = types.StringValue(instance.Description)
	resource.CloudProvider = types.StringValue(instance.Provider)
	resource.Region = types.StringValue(instance.Region)
	resource.ACL = types.BoolValue(instance.AclEnabled)

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

func ExpandKafkaTopicResource(topic KafkaTopicResourceModel, request *client.TopicCreateParam) {
	request.Name = topic.Name.ValueString()
	request.Partition = topic.Partition.ValueInt64()
	request.CompactStrategy = topic.CompactStrategy.ValueString()
	request.Configs = make([]client.ConfigItemParam, len(topic.Configs.Elements()))
	i := 0
	for name, value := range topic.Configs.Elements() {
		config := value.(types.String)
		request.Configs[i] = client.ConfigItemParam{
			Key:   name,
			Value: config.ValueString(),
		}
		i += 1
	}
}

func FlattenKafkaTopic(topic *client.TopicVO, resource *KafkaTopicResourceModel) diag.Diagnostics {
	resource.TopicID = types.StringValue(topic.TopicId)
	resource.Name = types.StringValue(topic.Name)
	resource.Partition = types.Int64Value(int64(topic.Partition))
	// config, diags := flattenTopicConfigs(topic.Configs, resource)
	// resource.Configs = config
	return nil
}

func flattenTopicConfigs(configs map[string]interface{}, resource *KafkaTopicResourceModel) (types.Map, diag.Diagnostics) {
	attrs := make(map[string]attr.Value, len(configs))

	for k, v := range configs {
		attrs[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	return types.MapValue(types.StringType, attrs)
}

func FlattenKafkaUserResource(user client.KafkaUserVO, resource *KafkaUserResourceModel) {
	resource.Username = types.StringValue(user.Name)
	if user.Password != "" {
		resource.Password = types.StringValue(user.Password)
	}
	resource.ID = types.StringValue(user.Name)
}

func ExpandKafkaACLResource(acl KafkaAclResourceModel, request *client.KafkaAclBindingParam) {
	request.AccessControlParam = client.KafkaControlParam{}
	request.ResourcePatternParam = client.KafkaResourcePatternParam{}
	request.AccessControlParam.OperationGroup = acl.OperationGroup.ValueString()
	request.AccessControlParam.PermissionType = acl.Permission.ValueString()

	request.AccessControlParam.User = ParsePrincipalUser(acl.Principal.ValueString())

	request.ResourcePatternParam.Name = acl.ResourceName.ValueString()
	request.ResourcePatternParam.PatternType = acl.PatternType.ValueString()
	request.ResourcePatternParam.ResourceType = acl.ResourceType.ValueString()
}

func ParsePrincipalUser(principal string) string {
	if condition := principal[:5]; condition == "User:" {
		return principal[5:]
	}
	return principal
}

func FlattenKafkaACLResource(acl *client.KafkaAclBindingVO, resource *KafkaAclResourceModel) {
	aclId, err := client.GenerateAclID(*acl)
	if err != nil {
		return
	}
	resource.ID = types.StringValue(aclId)

	resource.ResourceType = types.StringValue(acl.ResourcePattern.ResourceType)
	resource.ResourceName = types.StringValue(acl.ResourcePattern.Name)
	resource.PatternType = types.StringValue(acl.ResourcePattern.PatternType)
	resource.Principal = types.StringValue("User:" + acl.AccessControl.User)
	resource.OperationGroup = types.StringValue(acl.AccessControl.OperationGroup.Name)
	resource.Permission = types.StringValue(acl.AccessControl.PermissionType)
}
