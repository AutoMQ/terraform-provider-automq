package models

import (
	"strings"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KafkaTopicResourceModel describes the resource data model.
type KafkaTopicResourceModel struct {
	EnvironmentID types.String `tfsdk:"environment_id"`
	KafkaInstance types.String `tfsdk:"kafka_instance_id"`
	Name          types.String `tfsdk:"name"`
	Partition     types.Int64  `tfsdk:"partition"`
	Configs       types.Map    `tfsdk:"configs"`
	TopicID       types.String `tfsdk:"topic_id"`
}

func ExpandKafkaTopicResource(topic KafkaTopicResourceModel, request *client.TopicCreateParam) {
	request.Name = topic.Name.ValueString()
	request.Partition = topic.Partition.ValueInt64()

	request.Configs = make([]client.ConfigItemParam, len(topic.Configs.Elements()))
	i := 0
	for name, value := range topic.Configs.Elements() {
		config := value.(types.String)
		request.Configs[i] = client.ConfigItemParam{
			Key:   name,
			Value: config.ValueString(),
		}
		if name == "cleanup.policy" {
			request.CompactStrategy = strings.ToUpper(config.ValueString())
		}
		i += 1
	}
	if request.CompactStrategy == "" {
		request.CompactStrategy = "DELETE"
	}
}

func FlattenKafkaTopic(topic *client.TopicVO, resource *KafkaTopicResourceModel) diag.Diagnostics {
	resource.TopicID = types.StringValue(topic.TopicId)
	resource.Name = types.StringValue(topic.Name)
	resource.Partition = types.Int64Value(int64(topic.Partition))
	return nil
}
