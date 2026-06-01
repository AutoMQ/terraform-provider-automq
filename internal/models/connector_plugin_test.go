package models

import (
	"terraform-provider-automq/client"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandConnectorPluginCreate(t *testing.T) {
	plan := ConnectorPluginResourceModel{
		Name:           types.StringValue("my-s3-sink"),
		Version:        types.StringValue("1.0.0"),
		StorageUrl:     types.StringValue("s3://my-bucket/plugins/s3-sink-1.0.0.zip"),
		ConnectorClass: types.StringValue("io.confluent.connect.s3.S3SinkConnector"),
		Types:          types.ListValueMust(types.StringType, []attr.Value{types.StringValue("SINK")}),
		Description:    types.StringValue("S3 Sink Connector"),
	}

	got, diags := ExpandConnectorPluginCreate(plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got.Name != "my-s3-sink" {
		t.Errorf("Name = %q, want %q", got.Name, "my-s3-sink")
	}
	if got.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", got.Version, "1.0.0")
	}
	if got.StorageUrl != "s3://my-bucket/plugins/s3-sink-1.0.0.zip" {
		t.Errorf("StorageUrl = %q, want s3 URL", got.StorageUrl)
	}
	if got.ConnectorClass != "io.confluent.connect.s3.S3SinkConnector" {
		t.Errorf("ConnectorClass = %q", got.ConnectorClass)
	}
	if len(got.Types) != 1 || got.Types[0] != "SINK" {
		t.Errorf("Types = %v, want [SINK]", got.Types)
	}
	if got.Description == nil || *got.Description != "S3 Sink Connector" {
		t.Errorf("Description = %v", got.Description)
	}
}

func TestFlattenConnectorPlugin(t *testing.T) {
	now := time.Now()
	id := "conn-plugin-abc123"
	name := "my-s3-sink"
	desc := "S3 Sink Connector"
	version := "1.0.0"
	storageUrl := "s3://my-bucket/plugins/s3-sink-1.0.0.zip"
	provider := "CUSTOM"
	status := "ACTIVE"

	vo := &client.ConnectPluginVO{
		Id:          &id,
		Name:        &name,
		Description: &desc,
		Version:     &version,
		StorageUrl:  &storageUrl,
		Provider:    &provider,
		Status:      &status,
		Types:       []string{"SINK"},
		CreateTime:  &now,
		UpdateTime:  &now,
	}

	state := &ConnectorPluginResourceModel{
		Types: types.ListNull(types.StringType),
	}
	diags := FlattenConnectorPlugin(vo, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if state.ID.ValueString() != id {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), id)
	}
	if state.Name.ValueString() != name {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), name)
	}
	if state.PluginProvider.ValueString() != "CUSTOM" {
		t.Errorf("PluginProvider = %q, want CUSTOM", state.PluginProvider.ValueString())
	}
	if state.Status.ValueString() != "ACTIVE" {
		t.Errorf("Status = %q, want ACTIVE", state.Status.ValueString())
	}
	if len(state.Types.Elements()) != 1 {
		t.Errorf("Types length = %d, want 1", len(state.Types.Elements()))
	}
}
