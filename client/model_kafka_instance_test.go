package client

import (
	"encoding/json"
	"testing"
)

func TestInstanceCreateParamMarshalMatchesNewContract(t *testing.T) {
	param := InstanceCreateParam{
		Name:    "example",
		Version: "1.0.0",
		Spec: SpecificationParam{
			ReservedAku:         6,
			KubernetesLBSubnets: []string{"subnet-1"},
			ScheduleSpec:        stringPtr("nodeSelector: {}"),
			Provider:            stringPtr("aws"),
			Region:              stringPtr("us-east-1"),
			Vpc:                 stringPtr("vpc-123"),
			DataBuckets:         []BucketProfileParam{{BucketName: "data-bucket"}},
			FileSystem: &FileSystemParam{
				FileSystemType:               stringPtr("EFS_PROVISIONED"),
				ThroughputMiBpsPerFileSystem: 100,
				FileSystemCount:              1,
				SubnetIds:                    []string{"subnet-a", "subnet-b", "subnet-c"},
			},
		},
		Features: &InstanceFeatureParam{
			MetricsExporter: &InstanceMetricsExporterParam{
				Prometheus: &InstancePrometheusExporterParam{
					Enabled: boolPtr(true),
				},
			},
			TableTopic: &TableTopicParam{
				Enabled:     boolPtr(true),
				Warehouse:   "warehouse",
				CatalogType: "HIVE",
			},
			SchemaRegistryEnabled: boolPtr(true),
		},
	}

	encoded, err := json.Marshal(param)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, ok := payload["deployProfile"]; ok {
		t.Errorf("deployProfile should be omitted for empty values")
	}

	spec, ok := payload["spec"].(map[string]any)
	if !ok {
		t.Fatalf("spec field missing or wrong type: %T", payload["spec"])
	}
	if spec["scheduleSpec"] != "nodeSelector: {}" {
		t.Errorf("expected spec.scheduleSpec to be present, got %v", spec["scheduleSpec"])
	}
	if subnets, ok := spec["kubernetesLoadBalancerSubnets"].([]any); !ok || len(subnets) != 1 || subnets[0] != "subnet-1" {
		t.Errorf("expected spec.kubernetesLoadBalancerSubnets to be present, got %v", spec["kubernetesLoadBalancerSubnets"])
	}

	for _, key := range []string{"provider", "region", "vpc"} {
		if _, ok := spec[key]; !ok {
			t.Errorf("expected spec.%s to be present", key)
		}
	}
	fileSystem, ok := spec["fileSystemForFsWal"].(map[string]any)
	if !ok {
		t.Fatalf("spec.fileSystemForFsWal missing or wrong type: %T", spec["fileSystemForFsWal"])
	}
	fileSystemSubnets, ok := fileSystem["subnetIds"].([]any)
	if !ok || len(fileSystemSubnets) != 3 ||
		fileSystemSubnets[0] != "subnet-a" ||
		fileSystemSubnets[1] != "subnet-b" ||
		fileSystemSubnets[2] != "subnet-c" {
		t.Fatalf("expected spec.fileSystemForFsWal.subnetIds, got %v", fileSystem["subnetIds"])
	}

	features, ok := payload["features"].(map[string]any)
	if !ok {
		t.Fatalf("features field missing or wrong type: %T", payload["features"])
	}

	metrics, ok := features["metricsExporter"].(map[string]any)
	if !ok || metrics["prometheus"] == nil {
		t.Errorf("expected metricsExporter.prometheus to be present")
	}

	tableTopic, ok := features["tableTopic"].(map[string]any)
	if !ok {
		t.Fatalf("tableTopic missing or wrong type")
	}
	for _, key := range []string{"enabled", "warehouse", "catalogType"} {
		if _, ok := tableTopic[key]; !ok {
			t.Errorf("expected tableTopic.%s to be present", key)
		}
	}
	if tableTopic["enabled"] != true {
		t.Fatalf("expected tableTopic.enabled=true, got %v", tableTopic["enabled"])
	}

	if features["schemaRegistryEnabled"] != true {
		t.Fatalf("expected schemaRegistryEnabled=true, got %v", features["schemaRegistryEnabled"])
	}
	if _, ok := features["schemaRegistry"]; ok {
		t.Fatalf("schemaRegistry object should not be exposed in API payload")
	}
}

func TestInstanceVOUnmarshalFileSystemSubnetIds(t *testing.T) {
	const payload = `{
		"instanceId": "inst-1",
		"spec": {
			"fileSystemForFsWal": {
				"fileSystemType": "EFS_PROVISIONED",
				"subnetIds": ["subnet-a", "subnet-b", "subnet-c"]
			}
		}
	}`

	var instance InstanceVO
	if err := json.Unmarshal([]byte(payload), &instance); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if instance.Spec == nil || instance.Spec.FileSystem == nil {
		t.Fatalf("expected spec.fileSystemForFsWal to be decoded")
	}
	got := instance.Spec.FileSystem.SubnetIds
	if len(got) != 3 || got[0] != "subnet-a" || got[1] != "subnet-b" || got[2] != "subnet-c" {
		t.Fatalf("unexpected file system subnet IDs: %v", got)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
