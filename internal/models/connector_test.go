package models

import (
	"context"
	"terraform-provider-automq/client"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func connStrPtr(s string) *string { return &s }
func connInt32Ptr(i int32) *int32 { return &i }

func TestExpandConnectorCreate(t *testing.T) {
	ctx := context.Background()
	sensitive, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{"database.password": "secret"})
	assert.False(t, diags.HasError())
	config, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{"topics": "orders"})
	assert.False(t, diags.HasError())
	partition, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{"server": "server_01"})
	assert.False(t, diags.HasError())
	offset, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{"file": "mysql-bin.000001", "pos": "123"})
	assert.False(t, diags.HasError())

	plan := ConnectorResourceModel{
		ConnectClusterID:         types.StringValue("connect-cluster-1"),
		Name:                     types.StringValue("test-connector"),
		Description:              types.StringValue("a test connector"),
		ConnectorClass:           types.StringValue("io.example.SourceConnector"),
		TaskCount:                types.Int64Value(3),
		ConnectorConfig:          config,
		ConnectorConfigSensitive: sensitive,
		KafkaCluster: &ConnectorKafkaClusterModel{
			Security: &SecurityProtocolConfigModel{
				Protocol:      types.StringValue("SASL_PLAINTEXT"),
				Username:      types.StringValue("user1"),
				Password:      types.StringValue("pass1"),
				SaslMechanism: types.StringNull(),
			},
		},
		InitialOffsets: []InitialOffsetModel{{Partition: partition, Offset: offset}},
	}

	req, diags := ExpandConnectorCreate(plan)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, "connect-cluster-1", req.ConnectClusterId)
	assert.Equal(t, "test-connector", req.Name)
	assert.Equal(t, "io.example.SourceConnector", req.ConnectorClass)
	assert.Equal(t, int32(3), req.TaskCount)
	assert.Equal(t, "a test connector", *req.Description)
	assert.Equal(t, "SASL_PLAINTEXT", *req.KafkaCluster.SecurityProtocolConfig.Protocol)
	assert.Equal(t, "user1", *req.KafkaCluster.SecurityProtocolConfig.Username)
	assert.Equal(t, "pass1", *req.KafkaCluster.SecurityProtocolConfig.Password)
	assert.Equal(t, "orders", req.ConnectorConfig.Properties["topics"])
	assert.Equal(t, "secret", req.ConnectorConfigSensitive.Properties["database.password"])
	assert.Len(t, req.InitialOffsets, 1)
	assert.Equal(t, "server_01", req.InitialOffsets[0].Partition["server"])
	assert.Equal(t, "123", req.InitialOffsets[0].Offset["pos"])
}

func TestExpandConnectorUpdate(t *testing.T) {
	config, diags := types.MapValueFrom(context.Background(), types.StringType, map[string]string{"topics": "orders-updated"})
	assert.False(t, diags.HasError())

	plan := ConnectorResourceModel{
		Name:            types.StringValue("updated"),
		Description:     types.StringValue("new desc"),
		TaskCount:       types.Int64Value(5),
		ConnectorConfig: config,
		KafkaCluster: &ConnectorKafkaClusterModel{
			Security: &SecurityProtocolConfigModel{Protocol: types.StringValue("PLAINTEXT")},
		},
	}

	req, diags := ExpandConnectorUpdate(plan)
	assert.False(t, diags.HasError())
	assert.Equal(t, "updated", *req.Name)
	assert.Equal(t, "new desc", *req.Description)
	assert.Equal(t, int32(5), *req.TaskCount)
	assert.Equal(t, "PLAINTEXT", *req.SecurityProtocolConfig.Protocol)
	assert.Equal(t, "orders-updated", req.ConnectorConfig.Properties["topics"])
}

func TestFlattenConnector(t *testing.T) {
	now := time.Now()
	vo := &client.ConnectorVO{
		Id:               connStrPtr("conn-1"),
		ConnectClusterId: connStrPtr("connect-cluster-1"),
		Name:             connStrPtr("my-connector"),
		Description:      connStrPtr("desc"),
		State:            connStrPtr("RUNNING"),
		ConnectorType:    connStrPtr("SOURCE"),
		ConnectorClass:   connStrPtr("io.example.SourceConnector"),
		PluginId:         connStrPtr("plugin-1"),
		TaskCount:        connInt32Ptr(3),
		CreateTime:       &now,
		UpdateTime:       &now,
		ConnectorConfig:  map[string]interface{}{"topics": "orders"},
		SecurityProtocolConfig: &client.SecurityProtocolConfig{
			Protocol: connStrPtr("SASL_PLAINTEXT"),
			Username: connStrPtr("user1"),
		},
	}
	state := &ConnectorResourceModel{
		ConnectorConfigSensitive: types.MapNull(types.StringType),
	}

	diags := FlattenConnector(vo, state)
	assert.False(t, diags.HasError())
	assert.Equal(t, "conn-1", state.ID.ValueString())
	assert.Equal(t, "connect-cluster-1", state.ConnectClusterID.ValueString())
	assert.Equal(t, "my-connector", state.Name.ValueString())
	assert.Equal(t, "RUNNING", state.State.ValueString())
	assert.Equal(t, "SOURCE", state.ConnectorType.ValueString())
	assert.Equal(t, "io.example.SourceConnector", state.ConnectorClass.ValueString())
	assert.Equal(t, "plugin-1", state.PluginID.ValueString())
	assert.Equal(t, int64(3), state.TaskCount.ValueInt64())
	assert.Equal(t, "SASL_PLAINTEXT", state.KafkaCluster.Security.Protocol.ValueString())
	assert.Equal(t, "user1", state.KafkaCluster.Security.Username.ValueString())
	configVal, ok := state.ConnectorConfig.Elements()["topics"].(types.String)
	assert.True(t, ok)
	assert.Equal(t, "orders", configVal.ValueString())
}

func TestFlattenConnector_RetainsSensitiveConfig(t *testing.T) {
	existing, diags := types.MapValueFrom(context.Background(), types.StringType, map[string]string{"database.password": "secret"})
	assert.False(t, diags.HasError())
	state := &ConnectorResourceModel{ConnectorConfigSensitive: existing}

	diags = FlattenConnector(&client.ConnectorVO{Id: connStrPtr("conn-1")}, state)
	assert.False(t, diags.HasError())
	got, ok := state.ConnectorConfigSensitive.Elements()["database.password"].(types.String)
	assert.True(t, ok)
	assert.Equal(t, "secret", got.ValueString())
}

func TestFlattenConnector_Nil(t *testing.T) {
	state := &ConnectorResourceModel{}
	diags := FlattenConnector(nil, state)
	assert.True(t, diags.HasError())
}
