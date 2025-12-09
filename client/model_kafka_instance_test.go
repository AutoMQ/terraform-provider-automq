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
			ReservedAku: 6,
			Provider:    stringPtr("aws"),
			Region:      stringPtr("us-east-1"),
			Vpc:         stringPtr("vpc-123"),
			DataBuckets: []BucketProfileParam{{BucketName: "data-bucket"}},
		},
		Features: &InstanceFeatureParam{
			MetricsExporter: &InstanceMetricsExporterParam{
				Prometheus: &InstancePrometheusExporterParam{
					Enabled: boolPtr(true),
				},
			},
			TableTopic: &TableTopicParam{
				Warehouse:   "warehouse",
				CatalogType: "HIVE",
			},
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

	for _, key := range []string{"provider", "region", "vpc"} {
		if _, ok := spec[key]; !ok {
			t.Errorf("expected spec.%s to be present", key)
		}
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
	for _, key := range []string{"warehouse", "catalogType"} {
		if _, ok := tableTopic[key]; !ok {
			t.Errorf("expected tableTopic.%s to be present", key)
		}
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
