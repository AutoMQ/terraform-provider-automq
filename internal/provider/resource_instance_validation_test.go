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

func TestMetricsExporterAuthTypeSchema(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	featuresAttrRaw, ok := s.Attributes["features"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("features attribute has unexpected type %T", s.Attributes["features"])
	}
	featuresAttr := featuresAttrRaw
	metricsAttrRaw, ok := featuresAttr.Attributes["metrics_exporter"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("metrics_exporter attribute has unexpected type %T", featuresAttr.Attributes["metrics_exporter"])
	}
	metricsAttr := metricsAttrRaw
	prometheusAttrRaw, ok := metricsAttr.Attributes["prometheus"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("prometheus attribute has unexpected type %T", metricsAttr.Attributes["prometheus"])
	}
	prometheusAttr := prometheusAttrRaw
	authAttrRaw, ok := prometheusAttr.Attributes["auth_type"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("auth_type attribute has unexpected type %T", prometheusAttr.Attributes["auth_type"])
	}
	authAttr := authAttrRaw
	if !authAttr.Required {
		t.Fatalf("auth_type should be required")
	}
	if len(authAttr.Validators) == 0 {
		t.Fatalf("auth_type validators missing")
	}
}

func TestWalModeValidatorRejectsUnsupportedValue(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	featuresAttrRaw, ok := s.Attributes["features"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("features attribute has unexpected type %T", s.Attributes["features"])
	}
	featuresAttr := featuresAttrRaw
	walAttrRaw, ok := featuresAttr.Attributes["wal_mode"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("wal_mode attribute has unexpected type %T", featuresAttr.Attributes["wal_mode"])
	}
	walAttr := walAttrRaw
	if len(walAttr.Validators) == 0 {
		t.Fatalf("wal_mode validators missing")
	}

	// Test that an unsupported value is rejected
	req := validator.StringRequest{
		ConfigValue: types.StringValue("INVALID_WAL"),
		Path:        path.Root("features").AtName("wal_mode"),
	}
	resp := validator.StringResponse{}
	walAttr.Validators[0].ValidateString(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected validator error for unsupported wal mode")
	}

	// Test that all supported values are accepted
	supportedValues := []string{"EBSWAL", "S3WAL", "FSWAL"}
	for _, value := range supportedValues {
		respOk := validator.StringResponse{}
		walAttr.Validators[0].ValidateString(context.Background(), validator.StringRequest{
			ConfigValue: types.StringValue(value),
			Path:        req.Path,
		}, &respOk)
		if respOk.Diagnostics.HasError() {
			t.Fatalf("validator should accept %s: %v", value, respOk.Diagnostics)
		}
	}
}

func TestImmutableAttributesHaveRequiresReplace(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	featuresAttrRaw, ok := s.Attributes["features"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("features attribute has unexpected type %T", s.Attributes["features"])
	}
	featuresAttr := featuresAttrRaw
	computeAttrRaw, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs attribute has unexpected type %T", s.Attributes["compute_specs"])
	}
	computeAttr := computeAttrRaw

	// features.table_topic
	tableTopicAttrRaw, ok := featuresAttr.Attributes["table_topic"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("table_topic attribute has unexpected type %T", featuresAttr.Attributes["table_topic"])
	}
	tableTopicAttr := tableTopicAttrRaw
	if !hasObjectRequiresReplace(tableTopicAttr.PlanModifiers) {
		t.Fatalf("expected table_topic to require replacement, modifiers: %v", tableTopicAttr.PlanModifiers)
	}
	// compute_specs.instance_role
	instanceRoleAttrRaw, ok := computeAttr.Attributes["instance_role"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("instance_role attribute has unexpected type %T", computeAttr.Attributes["instance_role"])
	}
	instanceRoleAttr := instanceRoleAttrRaw
	if !hasStringRequiresReplace(instanceRoleAttr.PlanModifiers) {
		t.Fatalf("expected instance_role to require replacement, modifiers: %v", instanceRoleAttr.PlanModifiers)
	}
}

func getKafkaInstanceResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	resIface := NewKafkaInstanceResource()
	res, ok := resIface.(*KafkaInstanceResource)
	if !ok {
		t.Fatalf("NewKafkaInstanceResource returned unexpected type %T", resIface)
	}
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

func TestValidateKafkaInstanceConfiguration_FSWALMissingFileSystem(t *testing.T) {
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
			// FileSystemParam is intentionally nil
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateKafkaInstanceConfiguration(context.Background(), plan, nil)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when FSWAL mode is missing file_system_param")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "当 wal_mode 为 FSWAL 时，必须提供 file_system_param 配置") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error about missing file_system_param for FSWAL, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_FileSystemWithoutFSWAL(t *testing.T) {
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
			FileSystemParam: &models.FileSystemParamModel{
				ThroughputMibpsPerFileSystem: types.Int64Value(1000),
				FileSystemCount:              types.Int64Value(2),
				SecurityGroup:                types.StringValue("sg-test"),
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("EBSWAL"), // Not FSWAL
		},
	}

	diags := validateKafkaInstanceConfiguration(context.Background(), plan, nil)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when file_system_param is provided without FSWAL mode")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "file_system_param 配置仅在 wal_mode 为 FSWAL 时有效") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error about file_system_param only valid with FSWAL, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_FSWALValid(t *testing.T) {
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
			FileSystemParam: &models.FileSystemParamModel{
				ThroughputMibpsPerFileSystem: types.Int64Value(1000),
				FileSystemCount:              types.Int64Value(2),
				SecurityGroup:                types.StringValue("sg-test"),
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateKafkaInstanceConfiguration(context.Background(), plan, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics for valid FSWAL configuration: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_FSWALMissingThroughput(t *testing.T) {
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
			FileSystemParam: &models.FileSystemParamModel{
				ThroughputMibpsPerFileSystem: types.Int64Null(), // Missing required field
				FileSystemCount:              types.Int64Value(2),
				SecurityGroup:                types.StringValue("sg-test"),
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateKafkaInstanceConfiguration(context.Background(), plan, nil)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when throughput_mibps_per_file_system is missing")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "file_system_param.throughput_mibps_per_file_system 是必填字段") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error about missing throughput_mibps_per_file_system, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_FSWALMissingFileSystemCount(t *testing.T) {
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
			FileSystemParam: &models.FileSystemParamModel{
				ThroughputMibpsPerFileSystem: types.Int64Value(1000),
				FileSystemCount:              types.Int64Null(), // Missing required field
				SecurityGroup:                types.StringValue("sg-test"),
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateKafkaInstanceConfiguration(context.Background(), plan, nil)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when file_system_count is missing")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "file_system_param.file_system_count 是必填字段") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error about missing file_system_count, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_FileSystemWithoutFeatures(t *testing.T) {
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
			FileSystemParam: &models.FileSystemParamModel{
				ThroughputMibpsPerFileSystem: types.Int64Value(1000),
				FileSystemCount:              types.Int64Value(2),
				SecurityGroup:                types.StringValue("sg-test"),
			},
		},
		// Features is nil
	}

	diags := validateKafkaInstanceConfiguration(context.Background(), plan, nil)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when file_system_param is provided without features")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "file_system_param 配置仅在 wal_mode 为 FSWAL 时有效") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error about file_system_param only valid with FSWAL, got: %v", diags)
	}
}

func TestFileSystemParamValidators(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttrRaw, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs attribute has unexpected type %T", s.Attributes["compute_specs"])
	}
	computeAttr := computeAttrRaw
	fileSystemAttrRaw, ok := computeAttr.Attributes["file_system_param"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("file_system_param attribute has unexpected type %T", computeAttr.Attributes["file_system_param"])
	}
	fileSystemAttr := fileSystemAttrRaw

	// Test throughput_mibps_per_file_system validator
	throughputAttrRaw, ok := fileSystemAttr.Attributes["throughput_mibps_per_file_system"].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("throughput_mibps_per_file_system attribute has unexpected type %T", fileSystemAttr.Attributes["throughput_mibps_per_file_system"])
	}
	throughputAttr := throughputAttrRaw
	if !throughputAttr.Required {
		t.Fatalf("throughput_mibps_per_file_system should be required")
	}
	if len(throughputAttr.Validators) == 0 {
		t.Fatalf("throughput_mibps_per_file_system validators missing")
	}

	// Test that zero and negative values are rejected
	invalidValues := []int64{0, -1, -100}
	for _, value := range invalidValues {
		req := validator.Int64Request{
			ConfigValue: types.Int64Value(value),
			Path:        path.Root("compute_specs").AtName("file_system_param").AtName("throughput_mibps_per_file_system"),
		}
		resp := validator.Int64Response{}
		throughputAttr.Validators[0].ValidateInt64(context.Background(), req, &resp)
		if !resp.Diagnostics.HasError() {
			t.Fatalf("expected validator error for throughput value %d", value)
		}
	}

	// Test that positive values are accepted
	validValues := []int64{1, 100, 1000, 5000}
	for _, value := range validValues {
		req := validator.Int64Request{
			ConfigValue: types.Int64Value(value),
			Path:        path.Root("compute_specs").AtName("file_system_param").AtName("throughput_mibps_per_file_system"),
		}
		resp := validator.Int64Response{}
		throughputAttr.Validators[0].ValidateInt64(context.Background(), req, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("validator should accept throughput value %d: %v", value, resp.Diagnostics)
		}
	}

	// Test file_system_count validator
	countAttrRaw, ok := fileSystemAttr.Attributes["file_system_count"].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("file_system_count attribute has unexpected type %T", fileSystemAttr.Attributes["file_system_count"])
	}
	countAttr := countAttrRaw
	if !countAttr.Required {
		t.Fatalf("file_system_count should be required")
	}
	if len(countAttr.Validators) == 0 {
		t.Fatalf("file_system_count validators missing")
	}

	// Test that zero and negative values are rejected for file_system_count
	for _, value := range invalidValues {
		req := validator.Int64Request{
			ConfigValue: types.Int64Value(value),
			Path:        path.Root("compute_specs").AtName("file_system_param").AtName("file_system_count"),
		}
		resp := validator.Int64Response{}
		countAttr.Validators[0].ValidateInt64(context.Background(), req, &resp)
		if !resp.Diagnostics.HasError() {
			t.Fatalf("expected validator error for file_system_count value %d", value)
		}
	}

	// Test that positive values are accepted for file_system_count
	for _, value := range validValues {
		req := validator.Int64Request{
			ConfigValue: types.Int64Value(value),
			Path:        path.Root("compute_specs").AtName("file_system_param").AtName("file_system_count"),
		}
		resp := validator.Int64Response{}
		countAttr.Validators[0].ValidateInt64(context.Background(), req, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("validator should accept file_system_count value %d: %v", value, resp.Diagnostics)
		}
	}

	// Test security_group attribute properties
	securityGroupAttrRaw, ok := fileSystemAttr.Attributes["security_group"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("security_group attribute has unexpected type %T", fileSystemAttr.Attributes["security_group"])
	}
	securityGroupAttr := securityGroupAttrRaw
	if !securityGroupAttr.Optional {
		t.Fatalf("security_group should be optional")
	}
	if !securityGroupAttr.Computed {
		t.Fatalf("security_group should be computed")
	}
	if len(securityGroupAttr.PlanModifiers) == 0 {
		t.Fatalf("security_group plan modifiers missing")
	}
	if !hasStringRequiresReplace(securityGroupAttr.PlanModifiers) {
		t.Fatalf("expected security_group to require replacement, modifiers: %v", securityGroupAttr.PlanModifiers)
	}
}
