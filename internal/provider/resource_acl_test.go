package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKafkaAclResource(t *testing.T) {
	env := loadAccConfig(t)
	env.requireVM(t)
	ensureAccTimeout(t)

	suffix := generateRandomSuffix()
	instanceCfg := newVMInstanceConfig(env, fmt.Sprintf("acc-acl-instance-%s", suffix), "ACL acceptance instance")
	instanceHCL := renderKafkaInstanceConfig(env, instanceCfg)
	username := fmt.Sprintf("testuser-%s", suffix)
	topicName := fmt.Sprintf("acc-acl-topic-%s", suffix)

	userBlock := fmt.Sprintf(userResourceTemplate, env.EnvironmentID, username, "Test123456")
	topicBlock := fmt.Sprintf(topicResourceTemplate, env.EnvironmentID, topicName, 1, "86400000")

	testCases := []struct {
		name           string
		resourceType   string
		resourceName   string
		patternType    string
		operationGroup string
		permission     string
	}{
		// TOPIC tests
		{"topic_literal_all", "TOPIC", "test-topic", "LITERAL", "ALL", "ALLOW"},
		{"topic_literal_produce", "TOPIC", "test-topic", "LITERAL", "PRODUCE", "ALLOW"},
		{"topic_literal_consume", "TOPIC", "test-topic", "LITERAL", "CONSUME", "ALLOW"},
		{"topic_prefixed_all", "TOPIC", "test-t", "PREFIXED", "ALL", "ALLOW"},
		{"topic_prefixed_produce", "TOPIC", "test-t", "PREFIXED", "PRODUCE", "ALLOW"},
		{"topic_prefixed_consume", "TOPIC", "test-t", "PREFIXED", "CONSUME", "ALLOW"},

		// GROUP tests
		{"group_literal_all", "GROUP", "test-group", "LITERAL", "ALL", "ALLOW"},
		{"group_prefixed_all", "GROUP", "test-g", "PREFIXED", "ALL", "ALLOW"},

		// CLUSTER tests
		{"cluster_literal_all", "CLUSTER", "kafka-cluster", "LITERAL", "ALL", "ALLOW"},

		// TRANSACTIONAL_ID tests
		{"txn_literal_all", "TRANSACTIONAL_ID", "test-txn", "LITERAL", "ALL", "ALLOW"},
		{"txn_prefixed_all", "TRANSACTIONAL_ID", "test-t", "PREFIXED", "ALL", "ALLOW"},

		// Permission variation tests
		{"topic_literal_all_deny", "TOPIC", "test-topic-deny", "LITERAL", "ALL", "DENY"},
		{"group_literal_all_deny", "GROUP", "test-group-deny", "LITERAL", "ALL", "DENY"},
	}

	steps := make([]resource.TestStep, 0, len(testCases))
	for _, tc := range testCases {
		aclBlock := fmt.Sprintf(aclResourceTemplate,
			env.EnvironmentID,
			tc.resourceType,
			tc.resourceName,
			tc.patternType,
			tc.operationGroup,
			tc.permission,
		)
		steps = append(steps, resource.TestStep{
			Config: instanceHCL + userBlock + topicBlock + aclBlock,
			Check: resource.ComposeTestCheckFunc(
				testAccCheckKafkaAclExists("automq_kafka_acl.test"),
				resource.TestCheckResourceAttr("automq_kafka_acl.test", "resource_type", tc.resourceType),
				resource.TestCheckResourceAttr("automq_kafka_acl.test", "resource_name", tc.resourceName),
				resource.TestCheckResourceAttr("automq_kafka_acl.test", "pattern_type", tc.patternType),
				resource.TestCheckResourceAttr("automq_kafka_acl.test", "operation_group", tc.operationGroup),
				resource.TestCheckResourceAttr("automq_kafka_acl.test", "permission", tc.permission),
				resource.TestCheckResourceAttr("automq_kafka_acl.test", "principal", fmt.Sprintf("User:%s", username)),
			),
		})
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaAclDestroy,
		Steps:                    steps,
	})
}

// aclResourceTemplate renders a single ACL entry for the scenario table.
const aclResourceTemplate = `
resource "automq_kafka_acl" "test" {
  environment_id    = "%s"
  kafka_instance_id = automq_kafka_instance.test.id
  principal         = "User:${automq_kafka_user.test.username}"
  resource_type     = "%s"
  resource_name     = "%s"
  pattern_type      = "%s"
  operation_group   = "%s"
  permission        = "%s"
}
`

func testAccCheckKafkaAclDestroy(s *terraform.State) error {
	if err := testAccCheckKafkaInstanceDestroy(s); err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "automq_kafka_acl" {
			continue
		}
		// Add specific ACL destruction verification if needed
	}
	return nil
}

func testAccCheckKafkaAclExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ACL ID is set")
		}

		// Add specific ACL existence verification if needed
		return nil
	}
}
