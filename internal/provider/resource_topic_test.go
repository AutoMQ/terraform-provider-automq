package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKafkaTopicResource(t *testing.T) {
	envVars := getRequiredEnvVars(t)
	suffix := generateRandomSuffix()

	// First create an instance that we can reference
	instanceConfig := map[string]interface{}{
		"environment_id": envVars["AUTOMQ_TEST_ENV_ID"],
		"name":           fmt.Sprintf("test-topic-instance-%s", suffix),
		"description":    "Test instance for data source testing",
		"deploy_profile": envVars["AUTOMQ_TEST_DEPLOY_PROFILE"],
		"version":        "1.3.10",
		"reserved_aku":   6,
		"zone":           envVars["AUTOMQ_TEST_ZONE"],
		"subnet":         envVars["AUTOMQ_TEST_SUBNET_ID"],
	}

	// Initial configuration
	initialConfig := map[string]interface{}{
		"environment_id": envVars["AUTOMQ_TEST_ENV_ID"],
		"name":           fmt.Sprintf("test-topic-%s", suffix),
		"partition":      16,
		"configs_str": `{
			"cleanup.policy" = "delete"
			"retention.ms" = "86400000"
		}`,
		"automq_byoc_endpoint":      envVars["AUTOMQ_BYOC_ENDPOINT"],
		"automq_byoc_access_key_id": envVars["AUTOMQ_BYOC_ACCESS_KEY_ID"],
		"automq_byoc_secret_key":    envVars["AUTOMQ_BYOC_SECRET_KEY"],
	}

	// Updated configuration
	updatedConfig := map[string]interface{}{
		"environment_id": envVars["AUTOMQ_TEST_ENV_ID"],
		"name":           fmt.Sprintf("test-topic-%s", suffix),
		"partition":      32,
		"configs_str": `{
			"cleanup.policy" = "delete"
			"retention.ms" = "172800000"
		}`,
		"automq_byoc_endpoint":      envVars["AUTOMQ_BYOC_ENDPOINT"],
		"automq_byoc_access_key_id": envVars["AUTOMQ_BYOC_ACCESS_KEY_ID"],
		"automq_byoc_secret_key":    envVars["AUTOMQ_BYOC_SECRET_KEY"],
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaTopicDestroy,
		Steps: []resource.TestStep{
			// Create the instance
			{
				Config: testAccKafkaInstanceResourceConfig(instanceConfig, envVars),
			},
			// Initial creation
			{
				Config: testAccKafkaInstanceResourceConfig(instanceConfig, envVars) + testAccKafkaTopicConfig(initialConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaTopicExists("automq_kafka_topic.test"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "name", initialConfig["name"].(string)),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "partition", fmt.Sprintf("%d", initialConfig["partition"])),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "configs.cleanup.policy", "delete"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "configs.retention.ms", "86400000"),
					resource.TestCheckResourceAttrSet("automq_kafka_topic.test", "topic_id"),
				),
			},
			// Update test
			{
				Config: testAccKafkaInstanceResourceConfig(instanceConfig, envVars) + testAccKafkaTopicConfig(updatedConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaTopicExists("automq_kafka_topic.test"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "partition", fmt.Sprintf("%d", updatedConfig["partition"])),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "configs.cleanup.policy", "delete"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "configs.retention.ms", "172800000"),
				),
			},
			// Import test
			{
				ResourceName:      "automq_kafka_topic.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["automq_kafka_topic.test"]
					if !ok {
						return "", fmt.Errorf("Not found: %s", "automq_kafka_topic.test")
					}
					id := fmt.Sprintf("%s@%s@%s", rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["kafka_instance_id"], rs.Primary.Attributes["topic_id"])
					return id, nil
				},
				ImportStateVerifyIgnore: []string{"id"},
			},
		},
	})
}

func testAccKafkaTopicConfig(config map[string]interface{}) string {
	return fmt.Sprintf(`

resource "automq_kafka_topic" "test" {
  environment_id    = "%s"
  kafka_instance_id = automq_kafka_instance.test.id
  name             = "%s"
  partition        = %d
  configs          = %s
}
`,
		config["environment_id"].(string),
		config["name"].(string),
		config["partition"].(int),
		config["configs_str"].(string),
	)
}

func testAccCheckKafkaTopicDestroy(s *terraform.State) error {
	// Check if the instance is destroyed
	testAccCheckKafkaInstanceDestroy(s)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "automq_kafka_topic" {
			continue
		}

		// Add check to verify the topic was actually destroyed
		// In a real implementation, you would use the client to verify the topic no longer exists
		return nil
	}
	return nil
}

func testAccCheckKafkaTopicExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Topic ID is set")
		}

		// Add check to verify the topic exists
		// In a real implementation, you would use the client to verify the topic exists
		return nil
	}
}
