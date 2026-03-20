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
	ensureAccTimeout(t)

	suffix := generateRandomSuffix()
	connName := fmt.Sprintf("acc-conn-%s", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: renderConnectorConfig(env, connName, 1, 1, "TIER1"),
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
				Config: renderConnectorConfig(env, connName+"-upd", 2, 2, "TIER2"),
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

func renderConnectorConfig(env accConfig, name string, taskCount, workerCount int, workerSpec string) string {
	conn := env.Connector
	secProto := conn.SecurityProtocol
	if secProto == "" {
		secProto = "PLAINTEXT"
	}

	secFields := fmt.Sprintf("      security_protocol = %q\n", secProto)
	if conn.SecurityProtocolUser != "" {
		secFields += fmt.Sprintf("      username          = %q\n", conn.SecurityProtocolUser)
	}
	if conn.SecurityProtocolPass != "" {
		secFields += fmt.Sprintf("      password          = %q\n", conn.SecurityProtocolPass)
	}

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
	if conn.ConnCfgTopics != "" {
		connectorConfig = fmt.Sprintf(`
  connector_config = {
    "topics"         = %q
    "s3.region"      = %q
    "s3.bucket.name" = %q
    "flush.size"     = "3"
    "storage.class"  = "io.confluent.connect.s3.storage.S3Storage"
    "format.class"   = "io.confluent.connect.s3.format.json.JsonFormat"
  }
`, conn.ConnCfgTopics, conn.ConnCfgS3Region, conn.ConnCfgS3Bucket)
	}

	return fmt.Sprintf(`
provider "automq" {
  automq_byoc_endpoint      = %q
  automq_byoc_access_key_id = %q
  automq_byoc_secret_key    = %q
}

resource "automq_connector" "test" {
  environment_id             = %q
  name                       = %q
  plugin_id                  = %q
  kubernetes_cluster_id      = %q
  kubernetes_namespace       = %q
  kubernetes_service_account = %q
  task_count                 = %d
%s%s
  capacity = {
    worker_count         = %d
    worker_resource_spec = %q
  }

  kafka_cluster = {
    kafka_instance_id = %q
    security_protocol = {
%s    }
  }

  timeouts = {
    create = "30m"
    update = "30m"
    delete = "20m"
  }
}
`, env.Endpoint, env.AccessKeyID, env.SecretKey,
		env.EnvironmentID, name, conn.PluginID,
		conn.KubernetesClusterID, conn.KubernetesNamespace, conn.KubernetesServiceAcct,
		taskCount, optionalFields, connectorConfig,
		workerCount, workerSpec,
		conn.KafkaInstanceID, secFields)
}
