package models

import (
	"terraform-provider-automq/client"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFlattenKafkaLinkPreservesSensitiveFields(t *testing.T) {
	prior := &KafkaLinkingResourceModel{
		EnvironmentID: types.StringValue("env-123"),
		SourceCluster: &KafkaLinkSourceClusterModel{
			Endpoint:                      types.StringValue("old-endpoint"),
			Password:                      types.StringValue("secret"),
			TruststoreCertificates:        types.StringValue("trust"),
			KeystoreCertificateChain:      types.StringValue("chain"),
			KeystoreKey:                   types.StringValue("key"),
			DisableEndpointIdentification: types.BoolValue(true),
		},
	}

	now := time.Now().UTC()
	link := &client.KafkaLinkVO{
		LinkID:          "link-a",
		InstanceID:      "inst-a",
		StartOffsetTime: "latest",
		GmtCreate:       &now,
		SourceCluster: &client.KafkaLinkSourceClusterVO{
			Endpoint: "bootstrap:9092",
		},
		Status:       ptr("AVAILABLE"),
		ErrorMessage: ptr(""),
	}

	var state KafkaLinkingResourceModel
	diags := FlattenKafkaLink(link, &state, prior)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.EnvironmentID.ValueString() != "env-123" {
		t.Fatalf("environment id not preserved, got %s", state.EnvironmentID.ValueString())
	}
	if state.SourceCluster.Password.ValueString() != "secret" {
		t.Fatalf("password not preserved")
	}
	if state.SourceCluster.TruststoreCertificates.ValueString() != "trust" {
		t.Fatalf("truststore not preserved")
	}
	if state.SourceCluster.KeystoreCertificateChain.ValueString() != "chain" {
		t.Fatalf("keystore chain not preserved")
	}
	if state.SourceCluster.KeystoreKey.ValueString() != "key" {
		t.Fatalf("keystore key not preserved")
	}
	if !state.SourceCluster.DisableEndpointIdentification.ValueBool() {
		t.Fatalf("disable endpoint identification not preserved")
	}
}

func TestBuildMirrorTopicUpdateParamUppercasesState(t *testing.T) {
	model := &KafkaMirrorTopicResourceModel{}
	model.State = types.StringValue("paused")
	param := BuildMirrorTopicUpdateParam(model)
	if param.State != "PAUSED" {
		t.Fatalf("expected PAUSED, got %s", param.State)
	}
}

func TestBuildMirrorGroupCreateParamUsesSource(t *testing.T) {
	model := &KafkaMirrorGroupResourceModel{}
	model.SourceGroupID = types.StringValue("group-1")
	param := BuildMirrorGroupCreateParam(model)
	if len(param.SourceGroups) != 1 || param.SourceGroups[0].ConsumerGroup != "group-1" {
		t.Fatalf("unexpected create param: %+v", param)
	}
}

func ptr(v string) *string {
	return &v
}
