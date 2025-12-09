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

	request.Configs = make([]client.ConfigItemParam, 0, len(topic.Configs.Elements()))
	for name, value := range topic.Configs.Elements() {
		if config, ok := value.(types.String); ok {
			key := name
			val := config.ValueString()
			request.Configs = append(request.Configs, client.ConfigItemParam{
				Key:   &key,
				Value: &val,
			})
			if name == "cleanup.policy" {
				request.CompactStrategy = strings.ToUpper(val)
			}
		}
	}
	if request.CompactStrategy == "" {
		request.CompactStrategy = "DELETE"
	}
}

func FlattenKafkaTopic(topic *client.TopicVO, resource *KafkaTopicResourceModel) diag.Diagnostics {
	resource.TopicID = types.StringValue(topic.TopicId)
	resource.Name = types.StringValue(topic.Name)
	resource.Partition = types.Int64Value(topic.Partition)
	return nil
}
