package models

import (
	"testing"

	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestExpandKafkaTopicResource(t *testing.T) {
	config1, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"delete.retention.ms": types.StringValue("86400000"),
		"cleanup.policy":      types.StringValue("compact"),
	})

	config2, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"retention.ms": types.StringValue("3600000"),
	})

	tests := []struct {
		input    KafkaTopicResourceModel
		expected client.TopicCreateParam
	}{
		{
			input: KafkaTopicResourceModel{
				EnvironmentID: types.StringValue("env-123"),
				KafkaInstance: types.StringValue("kf-123"),
				Name:          types.StringValue("test-topic"),
				Partition:     types.Int64Value(3),
				Configs:       config1,
			},
			expected: client.TopicCreateParam{
				Name:            "test-topic",
				Partition:       3,
				CompactStrategy: "COMPACT",
				Configs: []client.ConfigItemParam{
					{Key: testStringPtr("cleanup.policy"), Value: testStringPtr("compact")},
					{Key: testStringPtr("delete.retention.ms"), Value: testStringPtr("86400000")},
				},
			},
		},
		{
			input: KafkaTopicResourceModel{
				EnvironmentID: types.StringValue("env-456"),
				KafkaInstance: types.StringValue("kf-456"),
				Name:          types.StringValue("another-topic"),
				Partition:     types.Int64Value(1),
				Configs:       config2,
			},
			expected: client.TopicCreateParam{
				Name:            "another-topic",
				Partition:       1,
				CompactStrategy: "DELETE",
				Configs: []client.ConfigItemParam{
					{Key: testStringPtr("retention.ms"), Value: testStringPtr("3600000")},
				},
			},
		},
	}

	for _, test := range tests {
		request := &client.TopicCreateParam{}
		ExpandKafkaTopicResource(test.input, request)

		assert.Equal(t, test.expected.Name, request.Name)
		assert.Equal(t, test.expected.Partition, request.Partition)
		assert.Equal(t, test.expected.CompactStrategy, request.CompactStrategy)
		assert.ElementsMatch(t, test.expected.Configs, request.Configs)
	}
}

func TestFlattenKafkaTopic(t *testing.T) {
	tests := []struct {
		input    *client.TopicVO
		expected KafkaTopicResourceModel
	}{
		{
			input: &client.TopicVO{
				TopicId:   "topic-123",
				Name:      "test-topic",
				Partition: 3,
			},
			expected: KafkaTopicResourceModel{
				TopicID:   types.StringValue("topic-123"),
				Name:      types.StringValue("test-topic"),
				Partition: types.Int64Value(3),
			},
		},
		{
			input: &client.TopicVO{
				TopicId:   "topic-456",
				Name:      "another-topic",
				Partition: 1,
			},
			expected: KafkaTopicResourceModel{
				TopicID:   types.StringValue("topic-456"),
				Name:      types.StringValue("another-topic"),
				Partition: types.Int64Value(1),
			},
		},
	}

	for _, test := range tests {
		resource := &KafkaTopicResourceModel{}
		diag := FlattenKafkaTopic(test.input, resource)

		assert.Nil(t, diag)
		assert.Equal(t, test.expected.TopicID.ValueString(), resource.TopicID.ValueString())
		assert.Equal(t, test.expected.Name.ValueString(), resource.Name.ValueString())
		assert.Equal(t, test.expected.Partition.ValueInt64(), resource.Partition.ValueInt64())
	}
}
