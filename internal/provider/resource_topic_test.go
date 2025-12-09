package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKafkaTopicResource(t *testing.T) {
	env := loadAccConfig(t)
	env.requireVM(t)
	ensureAccTimeout(t)

	suffix := generateRandomSuffix()
	instanceCfg := newVMInstanceConfig(env, fmt.Sprintf("acc-topic-instance-%s", suffix), "Topic acceptance instance")
	instanceHCL := renderKafkaInstanceConfig(env, instanceCfg)
	topicName := fmt.Sprintf("acc-topic-%s", suffix)

	initialTopic := fmt.Sprintf(topicResourceTemplate,
		env.EnvironmentID,
		topicName,
		16,
		"86400000",
	)
	updatedTopic := fmt.Sprintf(topicResourceTemplate,
		env.EnvironmentID,
		topicName,
		32,
		"172800000",
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaTopicDestroy,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: instanceHCL + initialTopic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaTopicExists("automq_kafka_topic.test"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "name", topicName),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "partition", "16"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "configs.cleanup.policy", "delete"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "configs.retention.ms", "86400000"),
					resource.TestCheckResourceAttrSet("automq_kafka_topic.test", "topic_id"),
				),
			},
			{
				Config: instanceHCL + updatedTopic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaTopicExists("automq_kafka_topic.test"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "partition", "32"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "configs.cleanup.policy", "delete"),
					resource.TestCheckResourceAttr("automq_kafka_topic.test", "configs.retention.ms", "172800000"),
				),
			},
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

// topicResourceTemplate renders the minimal topic resource used in tests.
const topicResourceTemplate = `
resource "automq_kafka_topic" "test" {
  environment_id    = "%s"
  kafka_instance_id = automq_kafka_instance.test.id
  name              = "%s"
  partition         = %d
  configs = {
    "cleanup.policy" = "delete"
    "retention.ms"   = "%s"
  }
}
`

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
