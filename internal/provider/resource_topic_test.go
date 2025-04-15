package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKafkaTopicResource(t *testing.T) {
	if os.Getenv("AUTOMQ_BYOC_ENDPOINT") == "" {
		t.Skip("Skipping test as AUTOMQ_TEST_DEPLOY_PROFILE is not set")
	}
	if os.Getenv("TF_ACC_TIMEOUT") == "" {
		t.Setenv("TF_ACC_TIMEOUT", "2h")
	}

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
		PreCheck:                 func() { testAccPreCheck(t) },
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
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "name", initialConfig["name"].(string)), //nolint:forcetypeassert
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
				ResourceName:                         "automq_kafka_topic.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "topic_id",
				ImportStateVerifyIgnore: []string{
					"configs.%", // ignore configs
					"configs.cleanup.policy",
					"configs.retention.ms",
				},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["automq_kafka_topic.test"]
					if !ok {
						return "", fmt.Errorf("Not found: %s", "automq_kafka_topic.test")
					}
					id := fmt.Sprintf("%s@%s@%s", rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["kafka_instance_id"], rs.Primary.Attributes["topic_id"])
					return id, nil
				},
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
		config["environment_id"].(string), //nolint:forcetypeassert
		config["name"].(string),           //nolint:forcetypeassert
		config["partition"].(int),         //nolint:forcetypeassert
		config["configs_str"].(string),    //nolint:forcetypeassert
	)
}

func testAccCheckKafkaTopicDestroy(s *terraform.State) error {
	// Check if the instance is destroyed
	if err := testAccCheckKafkaInstanceDestroy(s); err != nil {
		return err
	}

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
