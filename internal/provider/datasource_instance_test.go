package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKafkaInstanceDataSource(t *testing.T) {
	if os.Getenv("AUTOMQ_BYOC_ENDPOINT") == "" {
		t.Skip("Skipping test as AUTOMQ_TEST_DEPLOY_PROFILE is not set")
	}

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
		PreCheck:                 func() { testAccPreCheck(t) },
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
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "name", instanceConfig["name"].(string)),                     //nolint:forcetypeassert
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "description", instanceConfig["description"].(string)),       //nolint:forcetypeassert
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "deploy_profile", instanceConfig["deploy_profile"].(string)), //nolint:forcetypeassert
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "version", instanceConfig["version"].(string)),               //nolint:forcetypeassert

					// Verify compute specs
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "compute_specs.reserved_aku", fmt.Sprintf("%d", instanceConfig["reserved_aku"])),
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "compute_specs.networks.0.zone", instanceConfig["zone"].(string)),        //nolint:forcetypeassert
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "compute_specs.networks.0.subnets.0", instanceConfig["subnet"].(string)), //nolint:forcetypeassert

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
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "name", instanceConfig["name"].(string)), //nolint:forcetypeassert
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "id"),
				),
			},
		},
	})
}

func testAccKafkaInstanceResourceConfig(config map[string]interface{}, envVars map[string]string) string {
	environmentID, _ := config["environment_id"].(string)
	name, _ := config["name"].(string)
	description, _ := config["description"].(string)
	version, _ := config["version"].(string)
	reservedAku, _ := config["reserved_aku"].(int)
	zone, _ := config["zone"].(string)
	subnet, _ := config["subnet"].(string)

	return fmt.Sprintf(`
provider "automq" {
  automq_byoc_endpoint      = "%[1]s"
  automq_byoc_access_key_id = "%[2]s"
  automq_byoc_secret_key    = "%[3]s"
}

data "automq_deploy_profile" "test" {
  environment_id = "%[4]s"
  name          = "%[11]s"
}

data "automq_data_bucket_profiles" "test" {
  environment_id = "%[4]s"
  profile_name = data.automq_deploy_profile.test.name
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
			id = data.automq_data_bucket_profiles.test.data_buckets[0].id
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
		environmentID,
		name,
		description,
		version,
		reservedAku,
		zone,
		subnet,
		envVars["AUTOMQ_TEST_DEPLOY_PROFILE"],
	)
}

func testAccKafkaInstanceDataSourceConfigById(config map[string]interface{}, envVars map[string]string) string {
	environmentID, _ := config["environment_id"].(string)
	return fmt.Sprintf(`
%s

data "automq_kafka_instance" "test" {
  environment_id = "%s"
  id            = automq_kafka_instance.test.id
}
`,
		testAccKafkaInstanceResourceConfig(config, envVars),
		environmentID,
	)
}

func testAccKafkaInstanceDataSourceConfigByName(config map[string]interface{}, envVars map[string]string) string {
	environmentID, _ := config["environment_id"].(string)
	name, _ := config["name"].(string)
	return fmt.Sprintf(`
%s

data "automq_kafka_instance" "test" {
  environment_id = "%s"
  name          = "%s"
}
`,
		testAccKafkaInstanceResourceConfig(config, envVars),
		environmentID,
		name,
	)
}
