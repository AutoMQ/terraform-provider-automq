package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKafkaInstanceDataSource(t *testing.T) {
	env := loadAccConfig(t)
	env.requireVM(t)
	ensureAccTimeout(t)
	suffix := generateRandomSuffix()
	net, ok := env.firstVMNetwork()
	if !ok {
		t.Skip("Skipping data source test: acc.config.networks missing zone/subnet for VM")
	}
	if len(net.Subnets) == 0 {
		t.Skip("Skipping data source test: acc.config.networks requires at least one subnet for VM")
	}
	subnet := net.Subnets[0]

	// First create an instance that we can reference
	instanceConfig := map[string]interface{}{
		"environment_id": env.EnvironmentID,
		"name":           fmt.Sprintf("test-instance-%s", suffix),
		"description":    "Test instance for data source testing",
		"version":        "1.3.10",
		"reserved_aku":   6,
		"zone":           net.Zone,
		"subnet":         subnet,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create the instance
			{
				Config: testAccKafkaInstanceResourceConfig(instanceConfig, env),
			},
			// Test reading by ID
			{
				Config: testAccKafkaInstanceDataSourceConfigById(instanceConfig, env),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify basic attributes
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "name", instanceConfig["name"].(string)),               //nolint:forcetypeassert
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "description", instanceConfig["description"].(string)), //nolint:forcetypeassert
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "version", instanceConfig["version"].(string)),         //nolint:forcetypeassert

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
				Config: testAccKafkaInstanceDataSourceConfigByName(instanceConfig, env),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.automq_kafka_instance.test", "name", instanceConfig["name"].(string)), //nolint:forcetypeassert
					resource.TestCheckResourceAttrSet("data.automq_kafka_instance.test", "id"),
				),
			},
		},
	})
}

func testAccKafkaInstanceResourceConfig(config map[string]interface{}, env accConfig) string {
	environmentID, _ := config["environment_id"].(string)
	name, _ := config["name"].(string)
	description, _ := config["description"].(string)
	version, _ := config["version"].(string)
	reservedAku, _ := config["reserved_aku"].(int)
	zone, _ := config["zone"].(string)
	subnet, _ := config["subnet"].(string)

	return fmt.Sprintf(`
provider "automq" {
  automq_byoc_endpoint      = %q
  automq_byoc_access_key_id = %q
  automq_byoc_secret_key    = %q
}

resource "automq_kafka_instance" "test" {
  environment_id = %q
  name           = %q
  description    = %q
  version        = %q

  compute_specs = {
    reserved_aku = %d
    networks = [
      {
        zone    = %q
        subnets = [%q]
      }
    ]
  }

  features = {
    wal_mode = "EBSWAL"
    security = {
      authentication_methods   = ["anonymous", "sasl"]
      transit_encryption_modes = ["plaintext"]
    }
  }
}
`,
		env.Endpoint,
		env.AccessKeyID,
		env.SecretKey,
		environmentID,
		name,
		description,
		version,
		reservedAku,
		zone,
		subnet,
	)
}

func testAccKafkaInstanceDataSourceConfigById(config map[string]interface{}, env accConfig) string {
	environmentID, _ := config["environment_id"].(string)
	return fmt.Sprintf(`
%s

data "automq_kafka_instance" "test" {
  environment_id = %q
  id             = automq_kafka_instance.test.id
}
`,
		testAccKafkaInstanceResourceConfig(config, env),
		environmentID,
	)
}

func testAccKafkaInstanceDataSourceConfigByName(config map[string]interface{}, env accConfig) string {
	environmentID, _ := config["environment_id"].(string)
	name, _ := config["name"].(string)
	return fmt.Sprintf(`
%s

data "automq_kafka_instance" "test" {
  environment_id = %q
  name           = %q
}
`,
		testAccKafkaInstanceResourceConfig(config, env),
		environmentID,
		name,
	)
}
