package provider

import (
	"context"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	testingresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestKafkaInstanceReservedAKUThreePlansSuccessfully(t *testing.T) {
	testingresource.UnitTest(t, testingresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []testingresource.TestStep{{
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
			Config: `
provider "automq" {
  automq_byoc_endpoint      = "http://localhost:8080"
  automq_byoc_access_key_id = "test-access-key"
  automq_byoc_secret_key    = "test-secret-key"
}

resource "automq_kafka_instance" "test" {
  environment_id = "env-test"
  name            = "aku-three"
  deploy_profile  = "default"
  version         = "1.4.0"

  compute_specs = {
    reserved_aku = 3
    networks = [{
      zone    = "us-east-1a"
      subnets = ["subnet-test"]
    }]
    bucket_profiles = [{
      id = "bucket-profile-test"
    }]
  }

  features = {
    wal_mode = "EBSWAL"
    security = {
      authentication_methods   = ["anonymous"]
      transit_encryption_modes = ["plaintext"]
    }
  }
}
`,
		}},
	})
}

func TestKafkaInstanceReservedAKUHasNoValidation(t *testing.T) {
	var schemaResponse frameworkresource.SchemaResponse
	(&KafkaInstanceResource{}).Schema(context.Background(), frameworkresource.SchemaRequest{}, &schemaResponse)

	computeSpecs, ok := schemaResponse.Schema.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("compute_specs is not a single nested attribute")
	}
	reservedAKU, ok := computeSpecs.Attributes["reserved_aku"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("reserved_aku is not an int64 attribute")
	}
	if len(reservedAKU.Validators) != 0 {
		t.Fatalf("expected reserved_aku to have no validators, got %d", len(reservedAKU.Validators))
	}
}
