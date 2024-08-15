package models

import (
	"testing"
	"time"

	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

const (
	kafkaType      = IntegrationTypeKafka
	cloudWatchType = IntegrationTypeCloudWatch

	securityProtocol = "SASL_PLAINTEXT"
	saslMechanism    = "PLAIN"
	saslUsername     = "user1"
	saslPassword     = "pass1"
	testNamespace    = "test-namespace"
)

func TestFlattenIntergationResource(t *testing.T) {
	tests := []struct {
		integration *client.IntegrationVO
		expected    IntegrationResourceModel
	}{
		{
			integration: &client.IntegrationVO{
				Code:        "123",
				Name:        "test-kafka",
				Type:        kafkaType,
				GmtCreate:   time.Now(),
				GmtModified: time.Now(),
				Config: map[string]interface{}{
					"securityProtocol": securityProtocol,
					"saslMechanism":    saslMechanism,
					"saslUsername":     saslUsername,
					"saslPassword":     saslPassword,
				},
			},
			expected: IntegrationResourceModel{
				ID:   types.StringValue("123"),
				Name: types.StringValue("test-kafka"),
				Type: types.StringValue(kafkaType),
				KafkaConfig: &KafkaIntegrationConfig{
					SecurityProtocol: types.StringValue(securityProtocol),
					SaslMechanism:    types.StringValue(saslMechanism),
					SaslUsername:     types.StringValue(saslUsername),
					SaslPassword:     types.StringValue(saslPassword),
				},
			},
		},
		{
			integration: &client.IntegrationVO{
				Code:        "456",
				Name:        "test-cloudwatch",
				Type:        cloudWatchType,
				GmtCreate:   time.Now(),
				GmtModified: time.Now(),
				Config: map[string]interface{}{
					"namespace": testNamespace,
				},
			},
			expected: IntegrationResourceModel{
				ID:   types.StringValue("456"),
				Name: types.StringValue("test-cloudwatch"),
				Type: types.StringValue(cloudWatchType),
				CloudWatchConfig: &CloudWatchIntegrationConfig{
					NameSpace: types.StringValue(testNamespace),
				},
			},
		},
	}

	for _, test := range tests {
		resource := &IntegrationResourceModel{}
		FlattenIntergrationResource(test.integration, resource)

		assert.Equal(t, test.expected.ID.ValueString(), resource.ID.ValueString())
		assert.Equal(t, test.expected.Name.ValueString(), resource.Name.ValueString())
		assert.Equal(t, test.expected.Type.ValueString(), resource.Type.ValueString())
		if test.expected.KafkaConfig != nil {
			assert.Equal(t, test.expected.KafkaConfig.SecurityProtocol.ValueString(), resource.KafkaConfig.SecurityProtocol.ValueString())
			assert.Equal(t, test.expected.KafkaConfig.SaslMechanism.ValueString(), resource.KafkaConfig.SaslMechanism.ValueString())
			assert.Equal(t, test.expected.KafkaConfig.SaslUsername.ValueString(), resource.KafkaConfig.SaslUsername.ValueString())
			assert.Equal(t, test.expected.KafkaConfig.SaslPassword.ValueString(), resource.KafkaConfig.SaslPassword.ValueString())
		}
		if test.expected.CloudWatchConfig != nil {
			assert.Equal(t, test.expected.CloudWatchConfig.NameSpace.ValueString(), resource.CloudWatchConfig.NameSpace.ValueString())
		}
	}
}

func TestExpandIntergationResource(t *testing.T) {
	cloudwatch := cloudWatchType
	kafka := kafkaType

	tests := []struct {
		integration IntegrationResourceModel
		expected    client.IntegrationParam
	}{
		{
			integration: IntegrationResourceModel{
				Name:     types.StringValue("test-kafka"),
				Type:     types.StringValue(kafka),
				EndPoint: types.StringValue("http://localhost:9092"),
				KafkaConfig: &KafkaIntegrationConfig{
					SecurityProtocol: types.StringValue("SASL_PLAINTEXT"),
					SaslMechanism:    types.StringValue("PLAIN"),
					SaslUsername:     types.StringValue("user1"),
					SaslPassword:     types.StringValue("pass1"),
				},
			},
			expected: client.IntegrationParam{
				Name:     "test-kafka",
				Type:     &kafka,
				EndPoint: "http://localhost:9092",
				Config: []client.ConfigItemParam{
					{Key: "security_protocol", Value: "SASL_PLAINTEXT"},
					{Key: "sasl_mechanism", Value: "PLAIN"},
					{Key: "sasl_username", Value: "user1"},
					{Key: "sasl_password", Value: "pass1"},
				},
			},
		},
		{
			integration: IntegrationResourceModel{
				Name: types.StringValue("test-cloudwatch"),
				Type: types.StringValue(cloudwatch),
				CloudWatchConfig: &CloudWatchIntegrationConfig{
					NameSpace: types.StringValue("test-namespace"),
				},
			},
			expected: client.IntegrationParam{
				Name: "test-cloudwatch",
				Type: &cloudwatch,
				Config: []client.ConfigItemParam{
					{Key: "namespace", Value: "test-namespace"},
				},
			},
		},
	}

	for _, test := range tests {
		in := client.IntegrationParam{}
		diag := ExpandIntergationResource(&in, test.integration)

		assert.Nil(t, diag)
		assert.Equal(t, test.expected.Name, in.Name)
		assert.Equal(t, test.expected.Type, in.Type)
		assert.Equal(t, test.expected.EndPoint, in.EndPoint)
		assert.Equal(t, test.expected.Config, in.Config)
	}
}
