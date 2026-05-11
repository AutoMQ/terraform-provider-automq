package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccConnectorResource_basic(t *testing.T) {
	env := loadAccConfig(t)
	if !env.hasConnector() {
		t.Skip("Skipping connector acceptance tests: acc.config.connector section not configured")
	}
	if env.K8S.ClusterID == "" {
		t.Skip("Skipping connector acceptance tests: acc.config.k8s.cluster_id not configured")
	}
	env.requireVM(t)
	ensureAccTimeout(t)

	suffix := generateRandomSuffix()

	// 1. VM instance for the connector to attach to
	instanceCfg := newVMInstanceConfig(env, fmt.Sprintf("acc-conn-inst-%s", suffix), "Connector acceptance instance")
	instanceHCL := renderKafkaInstanceConfig(env, instanceCfg)

	// 2. Kafka user for SASL_PLAINTEXT auth
	connUser := fmt.Sprintf("conn-user-%s", suffix)
	connPass := "ConnTest123!"

	// 3. Topic that the connector will consume/produce
	topicName := fmt.Sprintf("acc-conn-topic-%s", suffix)

	// 4. Prerequisite HCL: user + topic + ACLs (TOPIC ALL, GROUP ALL, CLUSTER ALL)
	prereqHCL := renderConnectorPrereqs(env.EnvironmentID, connUser, connPass, topicName)

	connName := fmt.Sprintf("acc-conn-%s", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaInstanceDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: instanceHCL + prereqHCL + renderConnectorConfig(env, connName, connUser, connPass, topicName, 1, 1, "TIER1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("automq_connector.test", "name", connName),
					resource.TestCheckResourceAttr("automq_connector.test", "task_count", "1"),
					resource.TestCheckResourceAttr("automq_connector.test", "capacity.worker_count", "1"),
					resource.TestCheckResourceAttr("automq_connector.test", "capacity.worker_resource_spec", "TIER1"),
					resource.TestCheckResourceAttrSet("automq_connector.test", "id"),
					resource.TestCheckResourceAttrSet("automq_connector.test", "state"),
				),
			},
			// ImportState
			{
				ResourceName:      "automq_connector.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"kafka_cluster.security_protocol.password",
					"timeouts",
				},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["automq_connector.test"]
					if !ok {
						return "", fmt.Errorf("resource automq_connector.test not found")
					}
					return rs.Primary.Attributes["environment_id"] + "@" + rs.Primary.Attributes["id"], nil
				},
			},
			// Update name + task_count + capacity
			{
				Config: instanceHCL + prereqHCL + renderConnectorConfig(env, connName+"-upd", connUser, connPass, topicName, 2, 2, "TIER2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("automq_connector.test", "name", connName+"-upd"),
					resource.TestCheckResourceAttr("automq_connector.test", "task_count", "2"),
					resource.TestCheckResourceAttr("automq_connector.test", "capacity.worker_count", "2"),
					resource.TestCheckResourceAttr("automq_connector.test", "capacity.worker_resource_spec", "TIER2"),
				),
			},
		},
	})
}

// renderConnectorPrereqs generates the HCL for user, topic, and ACLs that the
// connector needs before it can start successfully.
func renderConnectorPrereqs(envID, username, password, topicName string) string {
	return fmt.Sprintf(`
resource "automq_kafka_user" "conn" {
  environment_id    = %q
  kafka_instance_id = automq_kafka_instance.test.id
  username          = %q
  password          = %q
}

resource "automq_kafka_topic" "conn" {
  environment_id    = %q
  kafka_instance_id = automq_kafka_instance.test.id
  name              = %q
  partition         = 1
  configs = {
    "cleanup.policy" = "delete"
    "retention.ms"   = "86400000"
  }
}

resource "automq_kafka_acl" "conn_topic" {
  environment_id    = %q
  kafka_instance_id = automq_kafka_instance.test.id
  principal         = "User:%s"
  resource_type     = "TOPIC"
  resource_name     = %q
  pattern_type      = "LITERAL"
  operation_group   = "ALL"
  permission        = "ALLOW"
  depends_on        = [automq_kafka_user.conn]
}

resource "automq_kafka_acl" "conn_group" {
  environment_id    = %q
  kafka_instance_id = automq_kafka_instance.test.id
  principal         = "User:%s"
  resource_type     = "GROUP"
  resource_name     = "connect"
  pattern_type      = "PREFIXED"
  operation_group   = "ALL"
  permission        = "ALLOW"
  depends_on        = [automq_kafka_user.conn]
}

resource "automq_kafka_acl" "conn_cluster" {
  environment_id    = %q
  kafka_instance_id = automq_kafka_instance.test.id
  principal         = "User:%s"
  resource_type     = "CLUSTER"
  resource_name     = "kafka-cluster"
  pattern_type      = "LITERAL"
  operation_group   = "ALL"
  permission        = "ALLOW"
  depends_on        = [automq_kafka_user.conn]
}

resource "automq_kafka_acl" "conn_txn" {
  environment_id    = %q
  kafka_instance_id = automq_kafka_instance.test.id
  principal         = "User:%s"
  resource_type     = "TRANSACTIONAL_ID"
  resource_name     = "connect"
  pattern_type      = "PREFIXED"
  operation_group   = "ALL"
  permission        = "ALLOW"
  depends_on        = [automq_kafka_user.conn]
}
`, envID, username, password,
		envID, topicName,
		envID, username, topicName,
		envID, username,
		envID, username,
		envID, username)
}

func renderConnectorConfig(env accConfig, name, saslUser, saslPass, topicName string, taskCount, workerCount int, workerSpec string) string {
	conn := env.Connector

	optionalFields := ""
	if conn.PluginType != "" {
		optionalFields += fmt.Sprintf("  plugin_type                  = %q\n", conn.PluginType)
	}
	if conn.ConnectorClass != "" {
		optionalFields += fmt.Sprintf("  connector_class              = %q\n", conn.ConnectorClass)
	}
	if conn.IamRole != "" {
		optionalFields += fmt.Sprintf("  iam_role                     = %q\n", conn.IamRole)
	}

	connectorConfig := ""
	if conn.ConnCfgS3Bucket != "" {
		connectorConfig = fmt.Sprintf(`
  connector_config = {
    "topics"         = %q
    "s3.region"      = %q
    "s3.bucket.name" = %q
    "flush.size"     = "3"
    "storage.class"  = "io.confluent.connect.s3.storage.S3Storage"
    "format.class"   = "io.confluent.connect.s3.format.json.JsonFormat"
  }
`, topicName, conn.ConnCfgS3Region, conn.ConnCfgS3Bucket)
	}

	return fmt.Sprintf(`
resource "automq_connector" "test" {
  environment_id             = %q
  name                       = %q
  plugin_id                  = %q
  kubernetes_cluster_id      = %q
  kubernetes_namespace       = "default"
  kubernetes_service_account = "default"
  task_count                 = %d
%s%s
  capacity = {
    worker_count         = %d
    worker_resource_spec = %q
  }

  kafka_cluster = {
    kafka_instance_id = automq_kafka_instance.test.id
    security_protocol = {
      security_protocol = "SASL_PLAINTEXT"
      username          = %q
      password          = %q
    }
  }

  timeouts = {
    create = "30m"
    update = "30m"
    delete = "20m"
  }

  depends_on = [
    automq_kafka_user.conn,
    automq_kafka_topic.conn,
    automq_kafka_acl.conn_topic,
    automq_kafka_acl.conn_group,
    automq_kafka_acl.conn_cluster,
    automq_kafka_acl.conn_txn,
  ]
}
`, env.EnvironmentID, name, conn.PluginID,
		env.K8S.ClusterID,
		taskCount, optionalFields, connectorConfig,
		workerCount, workerSpec,
		saslUser, saslPass)
}
