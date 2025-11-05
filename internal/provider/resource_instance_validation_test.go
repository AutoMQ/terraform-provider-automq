package provider

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-automq/internal/models"
)

func TestValidateKafkaInstanceConfiguration_K8SMissingCluster(t *testing.T) {
	plan := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			ReservedAku: types.Int64Value(6),
			DeployType:  types.StringValue("K8S"),
			Networks: []models.NetworkModel{{
				Zone:    types.StringValue("cn-test-1"),
				Subnets: types.ListNull(types.StringType),
			}},
			KubernetesNodeGroups: []models.NodeGroupModel{{
				ID: types.StringValue("ng-1"),
			}},
		},
	}

	diags := validateKafkaInstanceConfiguration(context.Background(), plan, nil)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when kubernetes_cluster_id is missing")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "kubernetes_cluster_id") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error mentioning kubernetes_cluster_id, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_K8SValid(t *testing.T) {
	plan := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			ReservedAku:          types.Int64Value(6),
			DeployType:           types.StringValue("K8S"),
			KubernetesClusterID:  types.StringValue("cluster-1"),
			KubernetesNodeGroups: []models.NodeGroupModel{{ID: types.StringValue("ng-1")}},
			Networks: []models.NetworkModel{{
				Zone: types.StringValue("cn-test-1"),
				Subnets: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("subnet-1"),
				}),
			}},
		},
		Features: &models.FeaturesModel{WalMode: types.StringValue("EBSWAL")},
	}

	diags := validateKafkaInstanceConfiguration(context.Background(), plan, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_DataBucketsMissingName(t *testing.T) {
	plan := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			ReservedAku: types.Int64Value(6),
			DeployType:  types.StringValue("IAAS"),
			Networks: []models.NetworkModel{{
				Zone: types.StringValue("cn-test-1"),
				Subnets: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("subnet-1"),
				}),
			}},
			DataBuckets: types.ListValueMust(
				models.DataBucketObjectType,
				[]attr.Value{
					types.ObjectValueMust(
						models.DataBucketObjectType.AttrTypes,
						map[string]attr.Value{
							"bucket_name": types.StringNull(),
						},
					),
				},
			),
		},
	}

	diags := validateKafkaInstanceConfiguration(context.Background(), plan, nil)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics for missing bucket_name")
	}

	found := false
	for _, d := range diags {
		if strings.Contains(d.Detail(), "compute_specs.data_buckets") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected data bucket error, got %v", diags)
	}
}

func TestWalModeValidatorRejectsUnsupportedValue(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	featuresAttr := s.Attributes["features"].(schema.SingleNestedAttribute)
	walAttr := featuresAttr.Attributes["wal_mode"].(schema.StringAttribute)
	if len(walAttr.Validators) == 0 {
		t.Fatalf("wal_mode validators missing")
	}

	req := validator.StringRequest{
		ConfigValue: types.StringValue("FSWAL"),
		Path:        path.Root("features").AtName("wal_mode"),
	}
	resp := validator.StringResponse{}
	walAttr.Validators[0].ValidateString(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected validator error for unsupported wal mode")
	}

	respOk := validator.StringResponse{}
	walAttr.Validators[0].ValidateString(context.Background(), validator.StringRequest{
		ConfigValue: types.StringValue("S3WAL"),
		Path:        req.Path,
	}, &respOk)
	if respOk.Diagnostics.HasError() {
		t.Fatalf("validator should accept S3WAL: %v", respOk.Diagnostics)
	}
}

func TestImmutableAttributesHaveRequiresReplace(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	featuresAttr := s.Attributes["features"].(schema.SingleNestedAttribute)
	computeAttr := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)

	// features.table_topic
	tableTopicAttr := featuresAttr.Attributes["table_topic"].(schema.SingleNestedAttribute)
	if !hasObjectRequiresReplace(tableTopicAttr.PlanModifiers) {
		t.Fatalf("expected table_topic to require replacement, modifiers: %v", tableTopicAttr.PlanModifiers)
	}
	// compute_specs.instance_role
	instanceRoleAttr := computeAttr.Attributes["instance_role"].(schema.StringAttribute)
	if !hasStringRequiresReplace(instanceRoleAttr.PlanModifiers) {
		t.Fatalf("expected instance_role to require replacement, modifiers: %v", instanceRoleAttr.PlanModifiers)
	}
}

func getKafkaInstanceResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	res := NewKafkaInstanceResource().(*KafkaInstanceResource)
	resp := resource.SchemaResponse{}
	res.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("failed to get schema: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func hasObjectRequiresReplace(mods []planmodifier.Object) bool {
	for _, m := range mods {
		if strings.Contains(strings.ToLower(reflect.TypeOf(m).String()), "requiresreplace") {
			return true
		}
	}
	return false
}

func hasStringRequiresReplace(mods []planmodifier.String) bool {
	for _, m := range mods {
		if strings.Contains(strings.ToLower(reflect.TypeOf(m).String()), "requiresreplace") {
			return true
		}
	}
	return false
}
