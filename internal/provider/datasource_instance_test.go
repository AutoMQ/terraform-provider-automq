package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKafkaInstanceDataSource(t *testing.T) {
	envVars := getRequiredEnvVars(t)
	suffix := generateRandomSuffix()

	// First create an instance that we can reference
	instanceConfig := map[string]interface{}{
		"environment_id": envVars["AUTOMQ_TEST_ENV_ID"],
		"name":           fmt.Sprintf("test-instance-%s", suffix),
		"description":    "Test instance for data source testing",
		"version":        "1.3.10",
		"reserved_aku":   6,
		"zone":           envVars["AUTOMQ_TEST_ZONE"],
		"subnet":         envVars["AUTOMQ_TEST_SUBNET_ID"],
		"deploy_profile": envVars["AUTOMQ_TEST_DEPLOY_PROFILE"],
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the instance
			{
				Config: testAccKafkaInstanceResourceConfig(instanceConfig, envVars),
			},
			// Test reading by ID
			{
				Config: testAccKafkaInstanceDataSourceConfigById(instanceConfig, envVars),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify basic attributes
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "name", instanceConfig["name"].(string)),
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "description", instanceConfig["description"].(string)),
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "deploy_profile", instanceConfig["deploy_profile"].(string)),
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "version", instanceConfig["version"].(string)),

					// Verify compute specs
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "compute_specs.reserved_aku", fmt.Sprintf("%d", instanceConfig["reserved_aku"])),
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "compute_specs.networks.0.zone", instanceConfig["zone"].(string)),
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "compute_specs.networks.0.subnets.0", instanceConfig["subnet"].(string)),

					// Verify computed fields are set
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "status"),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "last_updated"),

					// Verify features
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "features.wal_mode"),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "features.security.authentication_methods.#"),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "features.security.transit_encryption_modes.#"),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "features.security.data_encryption_mode"),

					// Verify endpoints
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "endpoints.#"),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "endpoints.0.display_name"),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "endpoints.0.network_type"),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "endpoints.0.protocol"),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "endpoints.0.bootstrap_servers"),
				),
			},
			// Test reading by name
			{
				Config: testAccKafkaInstanceDataSourceConfigByName(instanceConfig, envVars),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "name", instanceConfig["name"].(string)),
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "id"),
				),
			},
		},
	})
}

func testAccKafkaInstanceResourceConfig(config map[string]interface{}, envVars map[string]string) string {
	return fmt.Sprintf(`
provider "automq" {
  automq_byoc_endpoint      = "%[1]s"
  automq_byoc_access_key_id = "%[2]s"
  automq_byoc_secret_key    = "%[3]s"
}

data "automq_deploy_profile" "test" {
  environment_id = "`+envVars["AUTOMQ_TEST_ENV_ID"]+`"
  name          = "`+envVars["AUTOMQ_TEST_DEPLOY_PROFILE"]+`"
}

resource "automq_kafka_instance" "test" {
  environment_id = "%[4]s"
  name          = "%[5]s"
  description   = "%[6]s"
  deploy_profile = data.automq_deploy_profile.test.name
  version       = "%[7]s"

  compute_specs = {
	reserved_aku = %[8]d
	networks = [
	  {
		zone    = "%[9]s"
		subnets = ["%[10]s"]
	  }
	]
	bucket_profiles = [
		{
			id = data.automq_deploy_profile.test.data_buckets[0].id
		}
	]
  }
  features = {
	wal_mode = "EBSWAL"
	security = {
	  authentication_methods = ["anonymous", "sasl"]
	  transit_encryption_modes = ["plaintext"]
 	}
  }
}
`,
		envVars["AUTOMQ_BYOC_ENDPOINT"],
		envVars["AUTOMQ_BYOC_ACCESS_KEY_ID"],
		envVars["AUTOMQ_BYOC_SECRET_KEY"],
		config["environment_id"].(string),
		config["name"].(string),
		config["description"].(string),
		config["version"].(string),
		config["reserved_aku"].(int),
		config["zone"].(string),
		config["subnet"].(string),
	)
}

func testAccKafkaInstanceDataSourceConfigById(config map[string]interface{}, envVars map[string]string) string {
	return fmt.Sprintf(`
%s

data "automq_kafka_instance" "test" {
  environment_id = "%s"
  id            = automq_kafka_instance.test.id
}
`,
		testAccKafkaInstanceResourceConfig(config, envVars),
		config["environment_id"].(string),
	)
}

func testAccKafkaInstanceDataSourceConfigByName(config map[string]interface{}, envVars map[string]string) string {
	return fmt.Sprintf(`
%s

data "automq_kafka_instance" "test" {
  environment_id = "%s"
  name          = "%s"
}
`,
		testAccKafkaInstanceResourceConfig(config, envVars),
		config["environment_id"].(string),
		config["name"].(string),
	)
}
