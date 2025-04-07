package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKafkaAclResource(t *testing.T) {
	envVars := getRequiredEnvVars(t)
	suffix := generateRandomSuffix()
	username := fmt.Sprintf("testuser-%s", suffix)

	// Instance configuration
	instanceConfig := map[string]interface{}{
		"environment_id": envVars["AUTOMQ_TEST_ENV_ID"],
		"name":           fmt.Sprintf("test-acl-instance-%s", suffix),
		"description":    "Test instance for ACL testing",
		"deploy_profile": envVars["AUTOMQ_TEST_DEPLOY_PROFILE"],
		"version":        "1.3.10",
		"reserved_aku":   6,
		"zone":           envVars["AUTOMQ_TEST_ZONE"],
		"subnet":         envVars["AUTOMQ_TEST_SUBNET_ID"],
	}

	// Test cases for different ACL combinations
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

	steps := []resource.TestStep{
		{
			Config: testAccKafkaInstanceResourceConfig(instanceConfig, envVars),
		},
	}

	for _, tc := range testCases {
		steps = append(steps, resource.TestStep{
			Config: testAccKafkaInstanceResourceConfig(instanceConfig, envVars) +
				testAccKafkaAclConfigWithDependencies(
					envVars["AUTOMQ_TEST_ENV_ID"],
					username,
					tc.resourceType,
					tc.resourceName,
					tc.patternType,
					tc.operationGroup,
					tc.permission,
				),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckKafkaAclExists("automq_kafka_acl.test"),
				resource.TestCheckResourceAttr("automq_kafka_acl.test", "environment_id", envVars["AUTOMQ_TEST_ENV_ID"]),
				resource.TestCheckResourceAttrPair("automq_kafka_acl.test", "kafka_instance_id", "automq_kafka_instance.test", "id"),
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
		// PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaAclDestroy,
		Steps:                    steps,
	})
}

func testAccKafkaAclConfigWithDependencies(envId, username, resourceType, resourceName, patternType, operationGroup, permission string) string {
	userConfig := fmt.Sprintf(`
resource "automq_kafka_user" "test" {
  environment_id   = %[1]q
  kafka_instance_id = automq_kafka_instance.test.id
  username         = %[2]q
  password         = "Test123456"
}
`, envId, username)

	// Add topic resource if needed
	topicConfig := fmt.Sprintf(`
resource "automq_kafka_topic" "test" {
  environment_id   = %[1]q
  kafka_instance_id = automq_kafka_instance.test.id
  name             = "test-topic"
  partition        = 1
  configs = {
    "cleanup.policy" = "delete"
    "retention.ms" = "86400000"
  }
}
`, envId)

	// ACL resource configuration
	aclConfig := fmt.Sprintf(`
resource "automq_kafka_acl" "test" {
  environment_id   = %[1]q
  kafka_instance_id = automq_kafka_instance.test.id
  principal        = "User:${automq_kafka_user.test.username}"
  resource_type    = %[3]q
  resource_name    = %[4]q
  pattern_type     = %[5]q
  operation_group  = %[6]q
  permission       = %[7]q
}
`, envId, username, resourceType, resourceName, patternType, operationGroup, permission)

	return userConfig + topicConfig + aclConfig
}

func testAccCheckKafkaAclDestroy(s *terraform.State) error {
	testAccCheckKafkaInstanceDestroy(s)

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
