package models

import (
	"terraform-provider-automq/client"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func connStrPtr(s string) *string { return &s }
func connInt32Ptr(i int32) *int32 { return &i }

func TestExpandConnectorCreate(t *testing.T) {
	plan := ConnectorResourceModel{
		Name:                     types.StringValue("test-connector"),
		KubernetesClusterID:      types.StringValue("k8s-123"),
		PluginID:                 types.StringValue("plugin-1"),
		TaskCount:                types.Int64Value(3),
		KubernetesServiceAccount: types.StringValue("sa-test"),
		KubernetesNamespace:      types.StringValue("ns-test"),
		Description:              types.StringValue("a test connector"),
		PluginType:               types.StringValue("SOURCE"),
		ConnectorClass:           types.StringValue("com.example.Source"),
		IamRole:                  types.StringValue("arn:aws:iam::role/test"),
		Version:                  types.StringValue("1.2.0"),
		SchedulingSpec:           types.StringValue("nodeSelector:\n  zone: us-east-1a"),
		WorkerConfig:             types.MapNull(types.StringType),
		ConnectorConfig:          types.MapNull(types.StringType),
		Labels:                   types.MapNull(types.StringType),
		Capacity: &ConnectorCapacityModel{
			WorkerCount:        types.Int64Value(2),
			WorkerResourceSpec: types.StringValue("TIER2"),
		},
		KafkaCluster: &ConnectorKafkaClusterModel{
			KafkaInstanceID: types.StringValue("kf-abc"),
			Security: &SecurityProtocolConfigModel{
				SecurityProtocol: types.StringValue("SASL_PLAINTEXT"),
				Username:         types.StringValue("user1"),
				Password:         types.StringValue("pass1"),
				SaslMechanism:    types.StringNull(),
				TruststoreCerts:  types.StringNull(),
				ClientCert:       types.StringNull(),
				PrivateKey:       types.StringNull(),
			},
		},
	}

	req, diags := ExpandConnectorCreate(plan)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, "test-connector", req.Name)
	assert.Equal(t, "k8s-123", req.KubernetesClusterId)
	assert.Equal(t, "plugin-1", req.PluginId)
	assert.Equal(t, int32(3), req.TaskCount)
	assert.Equal(t, int32(2), req.Capacity.WorkerCount)
	assert.Equal(t, "TIER2", req.Capacity.WorkerResourceSpec)
	assert.Equal(t, "kf-abc", req.KafkaCluster.KafkaInstanceId)
	assert.Equal(t, "SASL_PLAINTEXT", *req.KafkaCluster.SecurityProtocolConfig.SecurityProtocol)
	assert.Equal(t, "user1", *req.KafkaCluster.SecurityProtocolConfig.Username)
	assert.Equal(t, "pass1", *req.KafkaCluster.SecurityProtocolConfig.Password)
	assert.Equal(t, "a test connector", *req.Description)
	assert.Equal(t, "SOURCE", *req.Type)
	assert.Equal(t, "com.example.Source", *req.ConnectorClass)
	assert.Equal(t, "arn:aws:iam::role/test", *req.IamRole)
	assert.Equal(t, "1.2.0", *req.Version)
	assert.Contains(t, *req.SchedulingSpec, "nodeSelector")
}

func TestExpandConnectorCreate_MissingCapacity(t *testing.T) {
	plan := ConnectorResourceModel{
		KafkaCluster: &ConnectorKafkaClusterModel{
			Security: &SecurityProtocolConfigModel{},
		},
	}
	_, diags := ExpandConnectorCreate(plan)
	assert.True(t, diags.HasError())
}

func TestExpandConnectorUpdate(t *testing.T) {
	plan := ConnectorResourceModel{
		Name:            types.StringValue("updated"),
		Description:     types.StringValue("new desc"),
		PluginID:        types.StringValue("plugin-2"),
		TaskCount:       types.Int64Value(5),
		Version:         types.StringValue("1.3.0"),
		SchedulingSpec:  types.StringValue("tolerations: []"),
		WorkerConfig:    types.MapNull(types.StringType),
		ConnectorConfig: types.MapNull(types.StringType),
		Labels:          types.MapNull(types.StringType),
		Capacity: &ConnectorCapacityModel{
			WorkerCount:        types.Int64Value(4),
			WorkerResourceSpec: types.StringValue("TIER3"),
		},
	}

	req, diags := ExpandConnectorUpdate(plan)
	assert.False(t, diags.HasError())
	assert.Equal(t, "updated", *req.Name)
	assert.Equal(t, "new desc", *req.Description)
	assert.Equal(t, int32(5), *req.TaskCount)
	assert.Equal(t, int32(4), req.Capacity.WorkerCount)
	assert.Equal(t, "1.3.0", *req.Version)
	assert.Equal(t, "tolerations: []", *req.SchedulingSpec)
}

func TestFlattenConnector(t *testing.T) {
	now := time.Now()
	vo := &client.ConnectorVO{
		Id:                       connStrPtr("conn-1"),
		Name:                     connStrPtr("my-connector"),
		Description:              connStrPtr("desc"),
		State:                    connStrPtr("RUNNING"),
		WorkerCount:              connInt32Ptr(2),
		WorkerResourceSpec:       connStrPtr("TIER2"),
		KafkaInstanceId:          connStrPtr("kf-abc"),
		KubernetesClusterId:      connStrPtr("k8s-123"),
		KubernetesNamespace:      connStrPtr("ns"),
		KubernetesServiceAccount: connStrPtr("sa"),
		IamRole:                  connStrPtr("role-arn"),
		ConnType:                 connStrPtr("SOURCE"),
		ConnClass:                connStrPtr("com.example.Source"),
		TaskCount:                connInt32Ptr(3),
		Version:                  connStrPtr("1.2.0"),
		KafkaConnectVersion:      connStrPtr("3.7.0"),
		SchedulingSpec:           connStrPtr("nodeSelector: {}"),
		CreateTime:               &now,
		UpdateTime:               &now,
		Plugin:                   &client.ConnectPluginSummaryVO{Id: connStrPtr("plugin-1")},
		Labels:                   map[string]string{"env": "test"},
		SecurityProtocolConfig: &client.SecurityProtocolConfig{
			SecurityProtocol: connStrPtr("SASL_PLAINTEXT"),
			Username:         connStrPtr("user1"),
		},
	}

	state := &ConnectorResourceModel{
		WorkerConfig:    types.MapNull(types.StringType),
		ConnectorConfig: types.MapNull(types.StringType),
		Labels:          types.MapNull(types.StringType),
	}
	diags := FlattenConnector(vo, state)
	assert.False(t, diags.HasError())
	assert.Equal(t, "conn-1", state.ID.ValueString())
	assert.Equal(t, "my-connector", state.Name.ValueString())
	assert.Equal(t, "RUNNING", state.State.ValueString())
	assert.Equal(t, int64(2), state.Capacity.WorkerCount.ValueInt64())
	assert.Equal(t, "TIER2", state.Capacity.WorkerResourceSpec.ValueString())
	assert.Equal(t, "kf-abc", state.KafkaCluster.KafkaInstanceID.ValueString())
	assert.Equal(t, "SASL_PLAINTEXT", state.KafkaCluster.Security.SecurityProtocol.ValueString())
	assert.Equal(t, "plugin-1", state.PluginID.ValueString())
	assert.Equal(t, "1.2.0", state.Version.ValueString())
	assert.Equal(t, "3.7.0", state.KafkaConnectVersion.ValueString())
	assert.Equal(t, "nodeSelector: {}", state.SchedulingSpec.ValueString())
	labelVal, ok := state.Labels.Elements()["env"].(types.String)
	assert.True(t, ok, "expected types.String for label 'env'")
	assert.Equal(t, "test", labelVal.ValueString())
}

func TestFlattenConnector_Nil(t *testing.T) {
	state := &ConnectorResourceModel{}
	diags := FlattenConnector(nil, state)
	assert.True(t, diags.HasError())
}
