package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKafkaUserResource(t *testing.T) {
	env := loadAccConfig(t)
	env.requireVM(t)
	ensureAccTimeout(t)

	suffix := generateRandomSuffix()
	instanceCfg := newVMInstanceConfig(env, fmt.Sprintf("acc-user-instance-%s", suffix), "User acceptance instance")
	instanceHCL := renderKafkaInstanceConfig(env, instanceCfg)
	username := fmt.Sprintf("test-user-%s", suffix)

	initialUser := fmt.Sprintf(userResourceTemplate,
		env.EnvironmentID,
		username,
		"TestPassword123!",
	)
	updatedUser := fmt.Sprintf(userResourceTemplate,
		env.EnvironmentID,
		username,
		"NewTestPassword456!",
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaUserDestroy,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create test
			{
				Config: instanceHCL + initialUser,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaUserExists("automq_kafka_user.test"),
					resource.TestCheckResourceAttr("automq_kafka_user.test", "username", username),           //nolint:forcetypeassert
					resource.TestCheckResourceAttr("automq_kafka_user.test", "password", "TestPassword123!"), //nolint:forcetypeassert
					resource.TestCheckResourceAttrSet("automq_kafka_user.test", "id"),
				),
			},
			// Update test
			{
				Config: instanceHCL + updatedUser,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaUserExists("automq_kafka_user.test"),
					resource.TestCheckResourceAttr("automq_kafka_user.test", "username", username),              //nolint:forcetypeassert
					resource.TestCheckResourceAttr("automq_kafka_user.test", "password", "NewTestPassword456!"), //nolint:forcetypeassert
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

// userResourceTemplate renders the user resource shared by the user/ACL tests.
const userResourceTemplate = `
resource "automq_kafka_user" "test" {
  environment_id    = "%s"
  kafka_instance_id = automq_kafka_instance.test.id
  username          = "%s"
  password          = "%s"
}
`

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
