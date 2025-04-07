package models

import (
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KafkaUserResourceModel describes the resource data model.
type KafkaUserResourceModel struct {
	EnvironmentID   types.String `tfsdk:"environment_id"`
	KafkaInstanceID types.String `tfsdk:"kafka_instance_id"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
	ID              types.String `tfsdk:"id"`
}

func FlattenKafkaUserResource(user *client.KafkaUserVO, resource *KafkaUserResourceModel) {
	resource.Username = types.StringValue(user.Name)
	if user.Password != "" {
		resource.Password = types.StringValue(user.Password)
	}
	// id: {environment_id}@{instance_id}@{username}
	resource.ID = types.StringValue(resource.EnvironmentID.String() + "@" + resource.KafkaInstanceID.String() + "@" + resource.Username.String())
}
