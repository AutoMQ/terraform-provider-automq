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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"terraform-provider-automq/client"
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

	diags := validateInstanceContract(context.Background(), plan)
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

	diags := validateInstanceContract(context.Background(), plan)
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

	diags := validateInstanceContract(context.Background(), plan)
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

func TestSpecificationUpdateParamMatchesBackendPatchContract(t *testing.T) {
	updateType := reflect.TypeOf(client.SpecificationUpdateParam{})
	expectedFields := map[string]struct{}{
		"ReservedAku":       {},
		"ReservedNodeCount": {},
		"NodeConfig":        {},
		"FileSystem":        {},
	}

	if updateType.NumField() != len(expectedFields) {
		t.Fatalf("SpecificationUpdateParam has %d fields, want %d", updateType.NumField(), len(expectedFields))
	}

	for fieldName := range expectedFields {
		if _, ok := updateType.FieldByName(fieldName); !ok {
			t.Fatalf("SpecificationUpdateParam missing field %s", fieldName)
		}
	}

	forbiddenFields := []string{
		"PricingMode",
		"SecurityGroups",
		"Template",
		"Networks",
		"KubernetesNodeGroups",
		"DeployType",
		"DnsZone",
		"KubernetesClusterId",
		"KubernetesNamespace",
		"KubernetesServiceAccount",
		"InstanceRole",
		"DataBuckets",
	}
	for _, fieldName := range forbiddenFields {
		if _, ok := updateType.FieldByName(fieldName); ok {
			t.Fatalf("SpecificationUpdateParam must not include create-only/backend-managed field %s", fieldName)
		}
	}

	fileSystemType := reflect.TypeOf(client.FileSystemUpdateParam{})
	if _, ok := fileSystemType.FieldByName("ThroughputMiBpsPerFileSystem"); !ok {
		t.Fatalf("FileSystemUpdateParam missing ThroughputMiBpsPerFileSystem")
	}
	if _, ok := fileSystemType.FieldByName("FileSystemCount"); !ok {
		t.Fatalf("FileSystemUpdateParam missing FileSystemCount")
	}
	if _, ok := fileSystemType.FieldByName("FileSystemType"); ok {
		t.Fatalf("FileSystemUpdateParam must not include create-only file_system_type")
	}
	if _, ok := fileSystemType.FieldByName("SecurityGroups"); ok {
		t.Fatalf("FileSystemUpdateParam must not include backend-managed security_groups")
	}
}

func TestFileSystemUpdateParamChangedOnlyConsidersPatchFields(t *testing.T) {
	state := &models.FileSystemParamModel{
		FileSystemType:               types.StringValue("EFS_PROVISIONED"),
		ThroughputMibpsPerFileSystem: types.Int64Value(384),
		FileSystemCount:              types.Int64Value(1),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-state"),
		}),
	}

	planCreateOnlyChanged := &models.FileSystemParamModel{
		FileSystemType:               types.StringValue("ONTAP_V2"),
		ThroughputMibpsPerFileSystem: types.Int64Value(384),
		FileSystemCount:              types.Int64Value(1),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-plan"),
		}),
	}
	if fileSystemUpdateParamChanged(planCreateOnlyChanged, state) {
		t.Fatalf("file system type and security_groups changes must not be treated as PATCH updates")
	}

	planThroughputChanged := &models.FileSystemParamModel{
		FileSystemType:               types.StringValue("EFS_PROVISIONED"),
		ThroughputMibpsPerFileSystem: types.Int64Value(768),
		FileSystemCount:              types.Int64Value(1),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-state"),
		}),
	}
	if !fileSystemUpdateParamChanged(planThroughputChanged, state) {
		t.Fatalf("throughput_mibps_per_file_system changes must be treated as PATCH updates")
	}

	planCountChanged := &models.FileSystemParamModel{
		FileSystemType:               types.StringValue("EFS_PROVISIONED"),
		ThroughputMibpsPerFileSystem: types.Int64Value(384),
		FileSystemCount:              types.Int64Value(2),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-state"),
		}),
	}
	if !fileSystemUpdateParamChanged(planCountChanged, state) {
		t.Fatalf("file_system_count changes must be treated as PATCH updates")
	}
}

func TestBuildInstanceUpdateParamOnlyIncludesBackendPatchFields(t *testing.T) {
	state := models.KafkaInstanceResourceModel{
		InstanceID:   types.StringValue("inst-1"),
		Name:         types.StringValue("old-name"),
		Description:  types.StringValue("old-description"),
		Version:      types.StringValue("1.0.0"),
		ComputeSpecs: buildUpdateTestComputeSpecs(6, 1, "EFS_PROVISIONED", 384, 1, "sg-state", "role-state"),
	}
	plan := models.KafkaInstanceResourceModel{
		InstanceID:   types.StringValue("inst-1"),
		Name:         types.StringValue("new-name"),
		Description:  types.StringValue("new-description"),
		Version:      types.StringValue("1.1.0"),
		ComputeSpecs: buildUpdateTestComputeSpecs(12, 2, "ONTAP_V2", 768, 2, "sg-plan", "role-plan"),
	}
	updateParam, updatePlan := buildInstanceUpdateParam(plan, state)
	if !updatePlan.hasUpdate {
		t.Fatalf("expected update plan to have updates")
	}
	if !updatePlan.shouldWait {
		t.Fatalf("expected update plan to wait for asynchronous instance update")
	}
	if updateParam.Name == nil || *updateParam.Name != "new-name" {
		t.Fatalf("expected name update, got %#v", updateParam.Name)
	}
	if updateParam.Description == nil || *updateParam.Description != "new-description" {
		t.Fatalf("expected description update, got %#v", updateParam.Description)
	}
	if updateParam.Version == nil || *updateParam.Version != "1.1.0" {
		t.Fatalf("expected version update, got %#v", updateParam.Version)
	}
	if updateParam.Spec == nil {
		t.Fatalf("expected spec update")
	}
	if updateParam.Spec.ReservedAku == nil || *updateParam.Spec.ReservedAku != 12 {
		t.Fatalf("expected reserved_aku update, got %#v", updateParam.Spec.ReservedAku)
	}
	if updateParam.Spec.ReservedNodeCount == nil || *updateParam.Spec.ReservedNodeCount != 2 {
		t.Fatalf("expected reserved_node_count update, got %#v", updateParam.Spec.ReservedNodeCount)
	}
	if updateParam.Spec.FileSystem == nil {
		t.Fatalf("expected file system update")
	}
	if updateParam.Spec.FileSystem.ThroughputMiBpsPerFileSystem != 768 {
		t.Fatalf("expected file system throughput update, got %d", updateParam.Spec.FileSystem.ThroughputMiBpsPerFileSystem)
	}
	if updateParam.Spec.FileSystem.FileSystemCount != 2 {
		t.Fatalf("expected file system count update, got %d", updateParam.Spec.FileSystem.FileSystemCount)
	}
}

func TestBuildInstanceUpdateParamUsesStateForReservedAkuDiff(t *testing.T) {
	state := models.KafkaInstanceResourceModel{
		InstanceID:   types.StringValue("inst-1"),
		Name:         types.StringValue("same-name"),
		Description:  types.StringValue("same-description"),
		Version:      types.StringValue("1.0.0"),
		ComputeSpecs: buildUpdateTestComputeSpecs(6, 1, "EFS_PROVISIONED", 384, 1, "sg-state", "role-state"),
	}
	plan := models.KafkaInstanceResourceModel{
		InstanceID:   types.StringValue("inst-1"),
		Name:         types.StringValue("same-name"),
		Description:  types.StringValue("same-description"),
		Version:      types.StringValue("1.0.0"),
		ComputeSpecs: buildUpdateTestComputeSpecs(12, 1, "EFS_PROVISIONED", 384, 1, "sg-state", "role-state"),
	}
	updateParam, updatePlan := buildInstanceUpdateParam(plan, state)
	if !updatePlan.hasUpdate {
		t.Fatalf("expected reserved_aku plan/state diff to produce PATCH update")
	}
	if updateParam.Spec == nil || updateParam.Spec.ReservedAku == nil || *updateParam.Spec.ReservedAku != 12 {
		t.Fatalf("expected reserved_aku update from plan/state diff, got %#v", updateParam.Spec)
	}
}

func TestBuildInstanceUpdateParamReturnsNoUpdateForCreateOnlyChanges(t *testing.T) {
	state := models.KafkaInstanceResourceModel{
		InstanceID:   types.StringValue("inst-1"),
		Name:         types.StringValue("same-name"),
		Description:  types.StringValue("same-description"),
		Version:      types.StringValue("1.0.0"),
		ComputeSpecs: buildUpdateTestComputeSpecs(6, 1, "EFS_PROVISIONED", 384, 1, "sg-state", "role-state"),
	}
	plan := models.KafkaInstanceResourceModel{
		InstanceID:   types.StringValue("inst-1"),
		Name:         types.StringValue("same-name"),
		Description:  types.StringValue("same-description"),
		Version:      types.StringValue("1.0.0"),
		ComputeSpecs: buildUpdateTestComputeSpecs(6, 1, "ONTAP_V2", 384, 1, "sg-plan", "role-plan"),
	}
	updateParam, updatePlan := buildInstanceUpdateParam(plan, state)
	if updatePlan.hasUpdate {
		t.Fatalf("create-only/backend-managed changes must not produce PATCH update: %#v", updateParam)
	}
	if updateParam.Spec != nil || updateParam.Features != nil || updateParam.Name != nil || updateParam.Description != nil || updateParam.Version != nil {
		t.Fatalf("expected empty update payload for create-only/backend-managed changes, got %#v", updateParam)
	}
}

func TestValidateInstanceUpdateContractRejectsInstanceConfigRemoval(t *testing.T) {
	stateConfig := types.MapValueMust(types.StringType, map[string]attr.Value{
		"retained": types.StringValue("value"),
		"removed":  types.StringValue("value"),
	})
	planConfig := types.MapValueMust(types.StringType, map[string]attr.Value{
		"retained": types.StringValue("value"),
	})
	state := models.KafkaInstanceResourceModel{
		Features: &models.FeaturesModel{
			InstanceConfigs: stateConfig,
		},
	}
	plan := models.KafkaInstanceResourceModel{
		Features: &models.FeaturesModel{
			InstanceConfigs: planConfig,
		},
	}

	diags := validateInstanceUpdateContract("inst-1", plan, state)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when removing an existing instance config key")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Detail(), "removal of instance settings") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected instance config removal error, got: %v", diags)
	}
}

func buildUpdateTestComputeSpecs(reservedAKU, reservedNodeCount int64, fileSystemType string, throughput, count int64, securityGroup, instanceRole string) *models.ComputeSpecsModel {
	return &models.ComputeSpecsModel{
		ReservedAku:       types.Int64Value(reservedAKU),
		ReservedNodeCount: types.Int64Value(reservedNodeCount),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue(securityGroup),
		}),
		InstanceRole: types.StringValue(instanceRole),
		FileSystemParam: &models.FileSystemParamModel{
			FileSystemType:               types.StringValue(fileSystemType),
			ThroughputMibpsPerFileSystem: types.Int64Value(throughput),
			FileSystemCount:              types.Int64Value(count),
			SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(securityGroup),
			}),
		},
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
	requireConfiguredOnlyStringReplacement(t, instanceRoleAttr.PlanModifiers)
}

func TestGeneratedStringAttributesOnlyReplaceWhenConfigured(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttrRaw, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs attribute has unexpected type %T", s.Attributes["compute_specs"])
	}
	computeAttr := computeAttrRaw

	dnsZoneAttrRaw, ok := computeAttr.Attributes["dns_zone"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("dns_zone attribute has unexpected type %T", computeAttr.Attributes["dns_zone"])
	}
	dnsZoneAttr := dnsZoneAttrRaw
	if !dnsZoneAttr.Optional {
		t.Fatalf("dns_zone should be optional")
	}
	if !dnsZoneAttr.Computed {
		t.Fatalf("dns_zone should be computed")
	}
	if !hasStringRequiresReplace(dnsZoneAttr.PlanModifiers) {
		t.Fatalf("expected dns_zone to require replacement, modifiers: %v", dnsZoneAttr.PlanModifiers)
	}
	requireConfiguredOnlyStringReplacement(t, dnsZoneAttr.PlanModifiers)
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

func requireConfiguredOnlyStringReplacement(t *testing.T, mods []planmodifier.String) {
	t.Helper()
	var replaceModifier planmodifier.String
	for _, m := range mods {
		if strings.Contains(strings.ToLower(reflect.TypeOf(m).String()), "requiresreplace") {
			replaceModifier = m
			break
		}
	}
	if replaceModifier == nil {
		t.Fatalf("requires replace modifier missing")
	}

	nonNullPlan := tfsdk.Plan{Raw: tftypes.NewValue(tftypes.String, "plan")}
	nonNullState := tfsdk.State{Raw: tftypes.NewValue(tftypes.String, "state")}

	unconfiguredReq := planmodifier.StringRequest{
		ConfigValue: types.StringNull(),
		Plan:        nonNullPlan,
		PlanValue:   types.StringUnknown(),
		State:       nonNullState,
		StateValue:  types.StringValue("state-value"),
	}
	unconfiguredResp := planmodifier.StringResponse{}
	replaceModifier.PlanModifyString(context.Background(), unconfiguredReq, &unconfiguredResp)
	if unconfiguredResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for unconfigured string attribute: %v", unconfiguredResp.Diagnostics)
	}
	if unconfiguredResp.RequiresReplace {
		t.Fatalf("unconfigured computed string attribute must not require replacement")
	}

	configuredReq := planmodifier.StringRequest{
		ConfigValue: types.StringValue("plan-value"),
		Plan:        nonNullPlan,
		PlanValue:   types.StringValue("plan-value"),
		State:       nonNullState,
		StateValue:  types.StringValue("state-value"),
	}
	configuredResp := planmodifier.StringResponse{}
	replaceModifier.PlanModifyString(context.Background(), configuredReq, &configuredResp)
	if configuredResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for configured string attribute: %v", configuredResp.Diagnostics)
	}
	if !configuredResp.RequiresReplace {
		t.Fatalf("configured changed string attribute must require replacement")
	}
}

func hasListRequiresReplace(mods []planmodifier.List) bool {
	for _, m := range mods {
		if strings.Contains(strings.ToLower(reflect.TypeOf(m).String()), "requiresreplace") {
			return true
		}
	}
	return false
}

func requireConfiguredOnlyListReplacement(t *testing.T, mods []planmodifier.List) {
	t.Helper()
	var replaceModifier planmodifier.List
	for _, m := range mods {
		if strings.Contains(strings.ToLower(reflect.TypeOf(m).String()), "requiresreplace") {
			replaceModifier = m
			break
		}
	}
	if replaceModifier == nil {
		t.Fatalf("requires replace modifier missing")
	}

	stateValue := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-state")})
	planValue := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-plan")})
	nonNullPlan := tfsdk.Plan{Raw: tftypes.NewValue(tftypes.String, "plan")}
	nonNullState := tfsdk.State{Raw: tftypes.NewValue(tftypes.String, "state")}

	unconfiguredReq := planmodifier.ListRequest{
		ConfigValue: types.ListNull(types.StringType),
		Plan:        nonNullPlan,
		PlanValue:   types.ListUnknown(types.StringType),
		State:       nonNullState,
		StateValue:  stateValue,
	}
	unconfiguredResp := planmodifier.ListResponse{}
	replaceModifier.PlanModifyList(context.Background(), unconfiguredReq, &unconfiguredResp)
	if unconfiguredResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for unconfigured security_groups: %v", unconfiguredResp.Diagnostics)
	}
	if unconfiguredResp.RequiresReplace {
		t.Fatalf("unconfigured computed security_groups must not require replacement")
	}

	configuredReq := planmodifier.ListRequest{
		ConfigValue: planValue,
		Plan:        nonNullPlan,
		PlanValue:   planValue,
		State:       nonNullState,
		StateValue:  stateValue,
	}
	configuredResp := planmodifier.ListResponse{}
	replaceModifier.PlanModifyList(context.Background(), configuredReq, &configuredResp)
	if configuredResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for configured security_groups: %v", configuredResp.Diagnostics)
	}
	if !configuredResp.RequiresReplace {
		t.Fatalf("configured changed security_groups must require replacement")
	}
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

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when FSWAL mode is missing file_system_param")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "file_system_param configuration is required when wal_mode is FSWAL") {
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
				SecurityGroups:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-test")}),
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("EBSWAL"), // Not FSWAL
		},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when file_system_param is provided without FSWAL mode")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "file_system_param configuration is only valid when wal_mode is FSWAL") {
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
				FileSystemType:               types.StringValue("EFS_PROVISIONED"),
				ThroughputMibpsPerFileSystem: types.Int64Value(1000),
				FileSystemCount:              types.Int64Value(2),
				SecurityGroups:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-test")}),
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateInstanceContract(context.Background(), plan)
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
				SecurityGroups:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-test")}),
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when throughput_mibps_per_file_system is missing")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "throughput_mibps_per_file_system is required when wal_mode is FSWAL") {
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
				SecurityGroups:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-test")}),
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when file_system_count is missing")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "file_system_count is required when wal_mode is FSWAL") {
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
				SecurityGroups:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-test")}),
			},
		},
		// Features is nil
	}

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when file_system_param is provided without features")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "file_system_param configuration is only valid when wal_mode is FSWAL") {
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

	// Test security_groups attribute properties
	securityGroupsAttrRaw, ok := fileSystemAttr.Attributes["security_groups"].(schema.ListAttribute)
	if !ok {
		t.Fatalf("security_groups attribute has unexpected type %T", fileSystemAttr.Attributes["security_groups"])
	}
	securityGroupsAttr := securityGroupsAttrRaw
	if !securityGroupsAttr.Optional {
		t.Fatalf("security_groups should be optional")
	}
	if !securityGroupsAttr.Computed {
		t.Fatalf("security_groups should be computed")
	}
	if len(securityGroupsAttr.PlanModifiers) == 0 {
		t.Fatalf("security_groups plan modifiers missing")
	}
	if !hasListRequiresReplace(securityGroupsAttr.PlanModifiers) {
		t.Fatalf("expected security_groups to require replacement, modifiers: %v", securityGroupsAttr.PlanModifiers)
	}
	requireConfiguredOnlyListReplacement(t, securityGroupsAttr.PlanModifiers)
}
func TestSecurityGroupsValidator(t *testing.T) {
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

	// Test security_groups validator
	securityGroupsAttrRaw, ok := fileSystemAttr.Attributes["security_groups"].(schema.ListAttribute)
	if !ok {
		t.Fatalf("security_groups attribute has unexpected type %T", fileSystemAttr.Attributes["security_groups"])
	}
	securityGroupsAttr := securityGroupsAttrRaw
	if len(securityGroupsAttr.Validators) == 0 {
		t.Fatalf("security_groups validators missing")
	}

	// Test that empty list is rejected
	req := validator.ListRequest{
		ConfigValue: types.ListValueMust(types.StringType, []attr.Value{}),
		Path:        path.Root("compute_specs").AtName("file_system_param").AtName("security_groups"),
	}
	resp := validator.ListResponse{}
	securityGroupsAttr.Validators[0].ValidateList(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected validator error for empty security_groups list")
	}

	// Test that list with one element is accepted
	reqValid := validator.ListRequest{
		ConfigValue: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-12345"),
		}),
		Path: path.Root("compute_specs").AtName("file_system_param").AtName("security_groups"),
	}
	respValid := validator.ListResponse{}
	securityGroupsAttr.Validators[0].ValidateList(context.Background(), reqValid, &respValid)
	if respValid.Diagnostics.HasError() {
		t.Fatalf("validator should accept non-empty security_groups list: %v", respValid.Diagnostics)
	}

	// Test that list with multiple elements is accepted
	reqMultiple := validator.ListRequest{
		ConfigValue: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-12345"),
			types.StringValue("sg-67890"),
		}),
		Path: path.Root("compute_specs").AtName("file_system_param").AtName("security_groups"),
	}
	respMultiple := validator.ListResponse{}
	securityGroupsAttr.Validators[0].ValidateList(context.Background(), reqMultiple, &respMultiple)
	if respMultiple.Diagnostics.HasError() {
		t.Fatalf("validator should accept multiple security_groups: %v", respMultiple.Diagnostics)
	}
}

// TestValidateKafkaInstanceConfiguration_FSWALWithEmptySecurityGroups tests that
// FSWAL configuration with empty security_groups list is handled correctly.
// Note: Empty list [] is rejected by schema validator (SizeAtLeast(1)).
// Users should omit the field entirely (null) to trigger backend auto-generation.
func TestValidateKafkaInstanceConfiguration_FSWALWithEmptySecurityGroups(t *testing.T) {
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
				SecurityGroups:               types.ListValueMust(types.StringType, []attr.Value{}), // Empty list
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateInstanceContract(context.Background(), plan)
	// Note: The schema validator (listvalidator.SizeAtLeast(1)) will reject empty lists,
	// but validateInstanceContract itself should not add additional errors.
	// This test verifies the custom validation logic doesn't break with empty security_groups.

	// Check if there's an error specifically about file_system_param being ignored
	for _, d := range diags {
		if strings.Contains(d.Detail(), "file_system_param") &&
			!strings.Contains(d.Detail(), "security_groups") {
			t.Fatalf("unexpected file_system_param error with empty security_groups: %v", d)
		}
	}
}

// TestValidateKafkaInstanceConfiguration_FSWALWithNullSecurityGroups tests that
// FSWAL configuration with null security_groups is valid.
// Omitting the field (null) triggers backend auto-generation of security groups.
func TestValidateKafkaInstanceConfiguration_FSWALWithNullSecurityGroups(t *testing.T) {
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
				FileSystemType:               types.StringValue("ONTAP_V2"),
				ThroughputMibpsPerFileSystem: types.Int64Value(1000),
				FileSystemCount:              types.Int64Value(2),
				SecurityGroups:               types.ListNull(types.StringType), // Null - backend will auto-generate
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics for FSWAL with null security_groups: %v", diags)
	}
}

// TestValidateKafkaInstanceConfiguration_FSWALWithUnknownSecurityGroups tests that
// FSWAL configuration with unknown security_groups is valid during planning phase.
func TestValidateKafkaInstanceConfiguration_FSWALWithUnknownSecurityGroups(t *testing.T) {
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
				FileSystemType:               types.StringValue("EFS_PROVISIONED"),
				ThroughputMibpsPerFileSystem: types.Int64Value(1000),
				FileSystemCount:              types.Int64Value(2),
				SecurityGroups:               types.ListUnknown(types.StringType), // Unknown during planning
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics for FSWAL with unknown security_groups: %v", diags)
	}
}

// TestComputeSpecsSecurityGroupsValidator tests the compute_specs.security_groups schema validator
func TestComputeSpecsSecurityGroupsValidator(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttrRaw, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs attribute has unexpected type %T", s.Attributes["compute_specs"])
	}
	computeAttr := computeAttrRaw

	// Test security_groups attribute properties
	securityGroupsAttrRaw, ok := computeAttr.Attributes["security_groups"].(schema.ListAttribute)
	if !ok {
		t.Fatalf("security_groups attribute has unexpected type %T", computeAttr.Attributes["security_groups"])
	}
	securityGroupsAttr := securityGroupsAttrRaw
	if !securityGroupsAttr.Optional {
		t.Fatalf("security_groups should be optional")
	}
	if !securityGroupsAttr.Computed {
		t.Fatalf("security_groups should be computed")
	}
	if len(securityGroupsAttr.PlanModifiers) == 0 {
		t.Fatalf("security_groups plan modifiers missing")
	}
	if !hasListRequiresReplace(securityGroupsAttr.PlanModifiers) {
		t.Fatalf("expected security_groups to require replacement, modifiers: %v", securityGroupsAttr.PlanModifiers)
	}
	requireConfiguredOnlyListReplacement(t, securityGroupsAttr.PlanModifiers)
	if len(securityGroupsAttr.Validators) == 0 {
		t.Fatalf("security_groups validators missing")
	}

	// Test that empty list is rejected
	req := validator.ListRequest{
		ConfigValue: types.ListValueMust(types.StringType, []attr.Value{}),
		Path:        path.Root("compute_specs").AtName("security_groups"),
	}
	resp := validator.ListResponse{}
	securityGroupsAttr.Validators[0].ValidateList(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected validator error for empty security_groups list")
	}

	// Test that list with one element is accepted
	reqValid := validator.ListRequest{
		ConfigValue: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-12345"),
		}),
		Path: path.Root("compute_specs").AtName("security_groups"),
	}
	respValid := validator.ListResponse{}
	securityGroupsAttr.Validators[0].ValidateList(context.Background(), reqValid, &respValid)
	if respValid.Diagnostics.HasError() {
		t.Fatalf("validator should accept non-empty security_groups list: %v", respValid.Diagnostics)
	}

	// Test that list with multiple elements is accepted
	reqMultiple := validator.ListRequest{
		ConfigValue: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-12345"),
			types.StringValue("sg-67890"),
		}),
		Path: path.Root("compute_specs").AtName("security_groups"),
	}
	respMultiple := validator.ListResponse{}
	securityGroupsAttr.Validators[0].ValidateList(context.Background(), reqMultiple, &respMultiple)
	if respMultiple.Diagnostics.HasError() {
		t.Fatalf("validator should accept multiple security_groups: %v", respMultiple.Diagnostics)
	}
}

// TestFileSystemTypeSchema tests the file_system_type field schema properties
func TestFileSystemTypeSchema(t *testing.T) {
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

	// Test file_system_type attribute exists and is required
	fileSystemTypeAttrRaw, ok := fileSystemAttr.Attributes["file_system_type"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("file_system_type attribute has unexpected type %T", fileSystemAttr.Attributes["file_system_type"])
	}
	fileSystemTypeAttr := fileSystemTypeAttrRaw

	if !fileSystemTypeAttr.Required {
		t.Fatalf("file_system_type should be required")
	}

	// Test that validators exist
	if len(fileSystemTypeAttr.Validators) == 0 {
		t.Fatalf("file_system_type validators missing")
	}

	// Test that RequiresReplace plan modifier exists
	if len(fileSystemTypeAttr.PlanModifiers) == 0 {
		t.Fatalf("file_system_type plan modifiers missing")
	}
	if !hasStringRequiresReplace(fileSystemTypeAttr.PlanModifiers) {
		t.Fatalf("expected file_system_type to require replacement, modifiers: %v", fileSystemTypeAttr.PlanModifiers)
	}
}

// TestFileSystemTypeValidator tests that file_system_type only accepts valid values
func TestFileSystemTypeValidator(t *testing.T) {
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
	fileSystemTypeAttrRaw, ok := fileSystemAttr.Attributes["file_system_type"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("file_system_type attribute has unexpected type %T", fileSystemAttr.Attributes["file_system_type"])
	}
	fileSystemTypeAttr := fileSystemTypeAttrRaw

	// Test valid values are accepted
	validValues := []string{"EFS_PROVISIONED", "ONTAP_V2"}
	for _, value := range validValues {
		req := validator.StringRequest{
			ConfigValue: types.StringValue(value),
			Path:        path.Root("compute_specs").AtName("file_system_param").AtName("file_system_type"),
		}
		resp := validator.StringResponse{}
		fileSystemTypeAttr.Validators[0].ValidateString(context.Background(), req, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("validator should accept %s: %v", value, resp.Diagnostics)
		}
	}

	// Test invalid values are rejected
	invalidValues := []string{"INVALID", "efs", "ontap", "EFS", "ONTAP", ""}
	for _, value := range invalidValues {
		req := validator.StringRequest{
			ConfigValue: types.StringValue(value),
			Path:        path.Root("compute_specs").AtName("file_system_param").AtName("file_system_type"),
		}
		resp := validator.StringResponse{}
		fileSystemTypeAttr.Validators[0].ValidateString(context.Background(), req, &resp)
		if !resp.Diagnostics.HasError() {
			t.Fatalf("validator should reject %s", value)
		}
	}
}

// TestValidateKafkaInstanceConfiguration_FSWALMissingFileSystemType tests that
// FSWAL mode with missing file_system_type returns validation error
func TestValidateKafkaInstanceConfiguration_FSWALMissingFileSystemType(t *testing.T) {
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
				FileSystemType:               types.StringNull(), // Missing file_system_type
				ThroughputMibpsPerFileSystem: types.Int64Value(1000),
				FileSystemCount:              types.Int64Value(2),
				SecurityGroups:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-test")}),
			},
		},
		Features: &models.FeaturesModel{
			WalMode: types.StringValue("FSWAL"),
		},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected diagnostics when file_system_type is missing")
	}

	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid Configuration" &&
			strings.Contains(d.Detail(), "file_system_type is required when wal_mode is FSWAL") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error about missing file_system_type, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_UsageBasedMissingNodeCount(t *testing.T) {
	plan := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			PricingMode:       types.StringValue("UsageBased"),
			ReservedNodeCount: types.Int64Null(),
			InstanceTypes:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("m5.xlarge")}),
			Networks: []models.NetworkModel{{
				Zone:    types.StringValue("us-east-1a"),
				Subnets: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-1")}),
			}},
		},
		Features: &models.FeaturesModel{WalMode: types.StringValue("EBSWAL")},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected error when reserved_node_count is missing for UsageBased pricing")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Detail(), "reserved_node_count") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error mentioning reserved_node_count, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_UsageBasedNullNodeCountDoesNotUseState(t *testing.T) {
	plan := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			PricingMode:       types.StringValue("UsageBased"),
			ReservedNodeCount: types.Int64Null(),
			InstanceTypes:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("m5.xlarge")}),
			Networks: []models.NetworkModel{{
				Zone:    types.StringValue("us-east-1a"),
				Subnets: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-1")}),
			}},
		},
		Features: &models.FeaturesModel{WalMode: types.StringValue("EBSWAL")},
	}
	state := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			PricingMode:       types.StringValue("UsageBased"),
			ReservedNodeCount: types.Int64Value(5),
			InstanceTypes:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("m5.xlarge")}),
		},
		Features: &models.FeaturesModel{WalMode: types.StringValue("EBSWAL")},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected error when reserved_node_count is null even if state has a value: %#v", state)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Detail(), "reserved_node_count") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error mentioning reserved_node_count, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_UsageBasedMissingInstanceTypes(t *testing.T) {
	plan := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			PricingMode:       types.StringValue("UsageBased"),
			DeployType:        types.StringValue("IAAS"),
			ReservedNodeCount: types.Int64Value(5),
			InstanceTypes:     types.ListNull(types.StringType),
			Networks: []models.NetworkModel{{
				Zone:    types.StringValue("us-east-1a"),
				Subnets: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-1")}),
			}},
		},
		Features: &models.FeaturesModel{WalMode: types.StringValue("EBSWAL")},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected error when instance_types is missing for UsageBased pricing")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Detail(), "instance_types") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error mentioning instance_types, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_UsageBasedK8SAllowsMissingInstanceTypes(t *testing.T) {
	plan := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			PricingMode:         types.StringValue("UsageBased"),
			DeployType:          types.StringValue("K8S"),
			ReservedNodeCount:   types.Int64Value(5),
			InstanceTypes:       types.ListNull(types.StringType),
			KubernetesClusterID: types.StringValue("cluster-1"),
			KubernetesNodeGroups: []models.NodeGroupModel{{
				ID: types.StringValue("ng-1"),
			}},
		},
		Features: &models.FeaturesModel{WalMode: types.StringValue("EBSWAL")},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics when instance_types is omitted for UsageBased K8S: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_CommittedMissingAku(t *testing.T) {
	plan := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			PricingMode: types.StringValue("SubscriptionBased"),
			ReservedAku: types.Int64Null(),
			Networks: []models.NetworkModel{{
				Zone:    types.StringValue("us-east-1a"),
				Subnets: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-1")}),
			}},
		},
		Features: &models.FeaturesModel{WalMode: types.StringValue("EBSWAL")},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if !diags.HasError() {
		t.Fatalf("expected error when reserved_aku is missing for SubscriptionBased pricing")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Detail(), "reserved_aku") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error mentioning reserved_aku, got: %v", diags)
	}
}

func TestValidateKafkaInstanceConfiguration_UsageBasedValid(t *testing.T) {
	plan := &models.KafkaInstanceResourceModel{
		ComputeSpecs: &models.ComputeSpecsModel{
			PricingMode:       types.StringValue("UsageBased"),
			ReservedNodeCount: types.Int64Value(5),
			InstanceTypes:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("m5.xlarge")}),
			Networks: []models.NetworkModel{{
				Zone:    types.StringValue("us-east-1a"),
				Subnets: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-1")}),
			}},
		},
		Features: &models.FeaturesModel{WalMode: types.StringValue("EBSWAL")},
	}

	diags := validateInstanceContract(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics for valid UsageBased config: %v", diags)
	}
}

func TestImmutableAttributesHaveRequiresReplace_PricingMode(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttr, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs has unexpected type %T", s.Attributes["compute_specs"])
	}

	pricingModeAttr, ok := computeAttr.Attributes["pricing_mode"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("pricing_mode has unexpected type %T", computeAttr.Attributes["pricing_mode"])
	}
	if !hasStringRequiresReplace(pricingModeAttr.PlanModifiers) {
		t.Fatalf("expected pricing_mode to require replacement")
	}
}

func TestImmutableAttributesHaveRequiresReplace_InstanceTypes(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttr, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs has unexpected type %T", s.Attributes["compute_specs"])
	}

	instanceTypesAttr, ok := computeAttr.Attributes["instance_types"].(schema.ListAttribute)
	if !ok {
		t.Fatalf("instance_types has unexpected type %T", computeAttr.Attributes["instance_types"])
	}
	if !hasListRequiresReplace(instanceTypesAttr.PlanModifiers) {
		t.Fatalf("expected instance_types to require replacement")
	}
}

func TestCreateOnlyComputeAttributesHaveRequiresReplace(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttr, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs has unexpected type %T", s.Attributes["compute_specs"])
	}

	deployTypeAttr, ok := computeAttr.Attributes["deploy_type"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("deploy_type has unexpected type %T", computeAttr.Attributes["deploy_type"])
	}
	if !hasStringRequiresReplace(deployTypeAttr.PlanModifiers) {
		t.Fatalf("expected deploy_type to require replacement")
	}
	if !deployTypeAttr.Optional || !deployTypeAttr.Computed {
		t.Fatalf("deploy_type should be optional and computed to allow provider default")
	}
	if deployTypeAttr.Default == nil {
		t.Fatalf("expected deploy_type to default to IAAS")
	}

	clusterIDAttr, ok := computeAttr.Attributes["kubernetes_cluster_id"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("kubernetes_cluster_id has unexpected type %T", computeAttr.Attributes["kubernetes_cluster_id"])
	}
	if !hasStringRequiresReplace(clusterIDAttr.PlanModifiers) {
		t.Fatalf("expected kubernetes_cluster_id to require replacement")
	}
}

func TestConfiguredManagedComputeAttributesHaveRequiresReplace(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttr, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs has unexpected type %T", s.Attributes["compute_specs"])
	}

	namespaceAttr, ok := computeAttr.Attributes["kubernetes_namespace"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("kubernetes_namespace has unexpected type %T", computeAttr.Attributes["kubernetes_namespace"])
	}
	if !namespaceAttr.Optional || !namespaceAttr.Computed {
		t.Fatalf("kubernetes_namespace should be optional and computed")
	}
	requireConfiguredOnlyStringReplacement(t, namespaceAttr.PlanModifiers)

	serviceAccountAttr, ok := computeAttr.Attributes["kubernetes_service_account"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("kubernetes_service_account has unexpected type %T", computeAttr.Attributes["kubernetes_service_account"])
	}
	if !serviceAccountAttr.Optional || !serviceAccountAttr.Computed {
		t.Fatalf("kubernetes_service_account should be optional and computed")
	}
	requireConfiguredOnlyStringReplacement(t, serviceAccountAttr.PlanModifiers)

	dataBucketsAttr, ok := computeAttr.Attributes["data_buckets"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatalf("data_buckets has unexpected type %T", computeAttr.Attributes["data_buckets"])
	}
	if !dataBucketsAttr.Optional || !dataBucketsAttr.Computed {
		t.Fatalf("data_buckets should be optional and computed")
	}
	requireConfiguredOnlyListReplacement(t, dataBucketsAttr.PlanModifiers)
}

func TestPricingModeSchemaValidator(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttr, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs has unexpected type %T", s.Attributes["compute_specs"])
	}

	pricingModeAttr, ok := computeAttr.Attributes["pricing_mode"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("pricing_mode has unexpected type %T", computeAttr.Attributes["pricing_mode"])
	}

	// pricing_mode should be Optional+Computed with a default.
	if !pricingModeAttr.IsOptional() {
		t.Fatalf("expected pricing_mode to be optional")
	}
	if !pricingModeAttr.IsComputed() {
		t.Fatalf("expected pricing_mode to be computed")
	}

	// Should have validators
	if len(pricingModeAttr.Validators) == 0 {
		t.Fatalf("expected pricing_mode to have validators")
	}
}

func TestInstanceTypesSchemaValidator(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttr, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs has unexpected type %T", s.Attributes["compute_specs"])
	}

	instanceTypesAttr, ok := computeAttr.Attributes["instance_types"].(schema.ListAttribute)
	if !ok {
		t.Fatalf("instance_types has unexpected type %T", computeAttr.Attributes["instance_types"])
	}

	if !instanceTypesAttr.IsOptional() {
		t.Fatalf("expected instance_types to be optional")
	}
	if instanceTypesAttr.IsComputed() {
		t.Fatalf("expected instance_types not to be computed")
	}

	// Should have size validators (at most 1, at least 1)
	if len(instanceTypesAttr.Validators) < 2 {
		t.Fatalf("expected instance_types to have at least 2 validators, got %d", len(instanceTypesAttr.Validators))
	}
}

func TestReservedNodeCountSchema(t *testing.T) {
	s := getKafkaInstanceResourceSchema(t)
	computeAttr, ok := s.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("compute_specs has unexpected type %T", s.Attributes["compute_specs"])
	}

	nodeCountAttr, ok := computeAttr.Attributes["reserved_node_count"].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("reserved_node_count has unexpected type %T", computeAttr.Attributes["reserved_node_count"])
	}

	if !nodeCountAttr.IsOptional() {
		t.Fatalf("expected reserved_node_count to be optional")
	}

	// Should have range validator (3-100)
	if len(nodeCountAttr.Validators) == 0 {
		t.Fatalf("expected reserved_node_count to have validators")
	}
}
