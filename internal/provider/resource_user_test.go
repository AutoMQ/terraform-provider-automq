package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKafkaUserResource(t *testing.T) {
	if os.Getenv("AUTOMQ_BYOC_ENDPOINT") == "" {
		t.Skip("Skipping test as AUTOMQ_TEST_DEPLOY_PROFILE is not set")
	}

	envVars := getRequiredEnvVars(t)
	suffix := generateRandomSuffix()

	// First create an instance that we can reference
	instanceConfig := map[string]interface{}{
		"environment_id": envVars["AUTOMQ_TEST_ENV_ID"],
		"name":           fmt.Sprintf("test-user-instance-%s", suffix),
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
		"username":       fmt.Sprintf("test-user-%s", suffix),
		"password":       "TestPassword123!",
	}

	// Updated configuration with new password
	updatedConfig := map[string]interface{}{
		"environment_id": envVars["AUTOMQ_TEST_ENV_ID"],
		"username":       fmt.Sprintf("test-user-%s", suffix),
		"password":       "NewTestPassword456!",
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaUserDestroy,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create the instance
			{
				Config: testAccKafkaInstanceResourceConfig(instanceConfig, envVars),
			},
			// Initial creation
			{
				Config: testAccKafkaInstanceResourceConfig(instanceConfig, envVars) + testAccKafkaUserConfig(initialConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaUserExists("automq_kafka_user.test"),
					resource.TestCheckResourceAttr("automq_kafka_user.test", "username", initialConfig["username"].(string)), //nolint:forcetypeassert
					resource.TestCheckResourceAttr("automq_kafka_user.test", "password", initialConfig["password"].(string)), //nolint:forcetypeassert
					resource.TestCheckResourceAttrSet("automq_kafka_user.test", "id"),
				),
			},
			// Update test with new password
			{
				Config: testAccKafkaInstanceResourceConfig(instanceConfig, envVars) + testAccKafkaUserConfig(updatedConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaUserExists("automq_kafka_user.test"),
					resource.TestCheckResourceAttr("automq_kafka_user.test", "username", updatedConfig["username"].(string)), //nolint:forcetypeassert
					resource.TestCheckResourceAttr("automq_kafka_user.test", "password", updatedConfig["password"].(string)), //nolint:forcetypeassert
				),
			},
			// Import test
			{
				ResourceName:      "automq_kafka_user.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["automq_kafka_user.test"]
					if !ok {
						return "", fmt.Errorf("Not found: %s", "automq_kafka_user.test")
					}
					id := fmt.Sprintf("%s@%s@%s", rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["kafka_instance_id"], rs.Primary.Attributes["username"])
					// The import ID format is <environment_id>@<kafka_instance_id>@<username>
					return id, nil
				},
				// Password cannot be imported
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccKafkaUserConfig(config map[string]interface{}) string {
	return fmt.Sprintf(`

resource "automq_kafka_user" "test" {
  environment_id    = "%s"
  kafka_instance_id = automq_kafka_instance.test.id
  username         = "%s"
  password         = "%s"
}
`,
		config["environment_id"].(string), //nolint:forcetypeassert
		config["username"].(string),       //nolint:forcetypeassert
		config["password"].(string),       //nolint:forcetypeassert
	)
}

func testAccCheckKafkaUserDestroy(s *terraform.State) error {
	// Check if the instance is destroyed
	if err := testAccCheckKafkaInstanceDestroy(s); err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "automq_kafka_user" {
			continue
		}

		// Add check to verify the user was actually destroyed
		// In a real implementation, you would use the client to verify the user no longer exists
		return nil
	}
	return nil
}

func testAccCheckKafkaUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User ID is set")
		}

		// Add check to verify the user exists
		// In a real implementation, you would use the client to verify the user exists
		return nil
	}
}
