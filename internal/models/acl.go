package models

import (
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KafkaAclResourceModel describes the resource data model.
type KafkaAclResourceModel struct {
	EnvironmentID  types.String `tfsdk:"environment_id"`
	KafkaInstance  types.String `tfsdk:"kafka_instance_id"`
	ID             types.String `tfsdk:"id"`
	ResourceType   types.String `tfsdk:"resource_type"`
	ResourceName   types.String `tfsdk:"resource_name"`
	PatternType    types.String `tfsdk:"pattern_type"`
	Principal      types.String `tfsdk:"principal"`
	OperationGroup types.String `tfsdk:"operation_group"`
	Permission     types.String `tfsdk:"permission"`
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

func FlattenKafkaACLResource(acl *client.KafkaAclBindingVO, resource *KafkaAclResourceModel) error {
	aclId, err := client.GenerateAclID(*acl)
	if err != nil {
		return err
	}
	resource.ID = types.StringValue(aclId)

	resource.ResourceType = types.StringValue(acl.ResourcePattern.ResourceType)
	resource.ResourceName = types.StringValue(acl.ResourcePattern.Name)
	resource.PatternType = types.StringValue(acl.ResourcePattern.PatternType)
	resource.Principal = types.StringValue("User:" + acl.AccessControl.User)
	resource.OperationGroup = types.StringValue(acl.AccessControl.OperationGroup.Name)
	resource.Permission = types.StringValue(acl.AccessControl.PermissionType)
	return nil
}
