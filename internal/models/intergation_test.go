package models

import (
	"testing"
	"time"

	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

const (
	cloudWatchType = IntegrationTypeCloudWatch

	securityProtocol = "SASL_PLAINTEXT"
	saslMechanism    = "PLAIN"
	saslUsername     = "user1"
	saslPassword     = "pass1"
	testNamespace    = "test-namespace"

	promRemoteWriteType = IntegrationTypePrometheusRemoteWrite

	testAuthType    = "basic"
	testUsername    = "admin"
	testPassword    = "password"
	testBearerToken = "test-token"
)

func TestFlattenIntergationResource(t *testing.T) {
	testEndpoint := "http://localhost:9090/api/v1/write"
	tests := []struct {
		integration *client.IntegrationVO
		expected    IntegrationResourceModel
	}{
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
		{
			integration: &client.IntegrationVO{
				Code:        "789",
				Name:        "test-prometheus-remote-write",
				Type:        promRemoteWriteType,
				EndPoint:    &testEndpoint,
				GmtCreate:   time.Now(),
				GmtModified: time.Now(),
				Config: map[string]interface{}{
					"authType": testAuthType,
					"username": testUsername,
					"password": testPassword,
				},
			},
			expected: IntegrationResourceModel{
				ID:       types.StringValue("789"),
				Name:     types.StringValue("test-prometheus-remote-write"),
				Type:     types.StringValue(promRemoteWriteType),
				EndPoint: types.StringValue(testEndpoint),
				PrometheusRemoteWriteConfig: &PrometheusRemoteWriteIntegrationConfig{
					AuthType: types.StringValue(testAuthType),
					Username: types.StringValue(testUsername),
					Password: types.StringValue(testPassword),
				},
			},
		},
		{
			integration: &client.IntegrationVO{
				Code:        "101",
				Name:        "test-prometheus-remote-write-bearer",
				Type:        promRemoteWriteType,
				EndPoint:    &testEndpoint,
				GmtCreate:   time.Now(),
				GmtModified: time.Now(),
				Config: map[string]interface{}{
					"authType": "bearer",
					"token":    testBearerToken,
				},
			},
			expected: IntegrationResourceModel{
				ID:       types.StringValue("101"),
				Name:     types.StringValue("test-prometheus-remote-write-bearer"),
				Type:     types.StringValue(promRemoteWriteType),
				EndPoint: types.StringValue(testEndpoint),
				PrometheusRemoteWriteConfig: &PrometheusRemoteWriteIntegrationConfig{
					AuthType:    types.StringValue("bearer"),
					BearerToken: types.StringValue(testBearerToken),
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
		if test.expected.CloudWatchConfig != nil {
			assert.Equal(t, test.expected.CloudWatchConfig.NameSpace.ValueString(), resource.CloudWatchConfig.NameSpace.ValueString())
		}
		if test.expected.PrometheusRemoteWriteConfig != nil {
			assert.Equal(t, test.expected.PrometheusRemoteWriteConfig.AuthType.ValueString(), resource.PrometheusRemoteWriteConfig.AuthType.ValueString())
			if test.expected.PrometheusRemoteWriteConfig.AuthType.ValueString() == "basic" {
				assert.Equal(t, test.expected.PrometheusRemoteWriteConfig.Username.ValueString(), resource.PrometheusRemoteWriteConfig.Username.ValueString())
				assert.Equal(t, test.expected.PrometheusRemoteWriteConfig.Password.ValueString(), resource.PrometheusRemoteWriteConfig.Password.ValueString())
			} else if test.expected.PrometheusRemoteWriteConfig.AuthType.ValueString() == "bearer" {
				assert.Equal(t, test.expected.PrometheusRemoteWriteConfig.BearerToken.ValueString(), resource.PrometheusRemoteWriteConfig.BearerToken.ValueString())
			}
		}
	}
}

func TestExpandIntergationResource(t *testing.T) {
	cloudwatch := cloudWatchType
	promRemoteWrite := promRemoteWriteType
	testEndpoint := "http://localhost:9090/api/v1/write"

	tests := []struct {
		integration IntegrationResourceModel
		expected    client.IntegrationParam
	}{
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
		{
			integration: IntegrationResourceModel{
				Name:     types.StringValue("test-prometheus-remote-write"),
				Type:     types.StringValue(promRemoteWrite),
				EndPoint: types.StringValue(testEndpoint),
				PrometheusRemoteWriteConfig: &PrometheusRemoteWriteIntegrationConfig{
					AuthType: types.StringValue(testAuthType),
					Username: types.StringValue(testUsername),
					Password: types.StringValue(testPassword),
				},
			},
			expected: client.IntegrationParam{
				Name:     "test-prometheus-remote-write",
				Type:     &promRemoteWrite,
				EndPoint: testEndpoint,
				Config: []client.ConfigItemParam{
					{Key: "authType", Value: testAuthType},
					{Key: "username", Value: testUsername},
					{Key: "password", Value: testPassword},
				},
			},
		},
		{
			integration: IntegrationResourceModel{
				Name:     types.StringValue("test-prometheus-remote-write-bearer"),
				Type:     types.StringValue(promRemoteWrite),
				EndPoint: types.StringValue(testEndpoint),
				PrometheusRemoteWriteConfig: &PrometheusRemoteWriteIntegrationConfig{
					AuthType:    types.StringValue("bearer"),
					BearerToken: types.StringValue(testBearerToken),
				},
			},
			expected: client.IntegrationParam{
				Name:     "test-prometheus-remote-write-bearer",
				Type:     &promRemoteWrite,
				EndPoint: testEndpoint,
				Config: []client.ConfigItemParam{
					{Key: "authType", Value: "bearer"},
					{Key: "token", Value: testBearerToken},
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
