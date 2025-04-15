package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIntegrationResource(t *testing.T) {
	envId := os.Getenv("AUTOMQ_TEST_ENV_ID")
	if envId == "" {
		t.Skip("AUTOMQ_TEST_ENV_ID must be set for this test")
	}
	deployProfile := os.Getenv("AUTOMQ_TEST_DEPLOY_PROFILE")
	if deployProfile == "" {
		t.Skip("AUTOMQ_TEST_DEPLOY_PROFILE must be set for this test")
	}
	if os.Getenv("TF_ACC_TIMEOUT") == "" {
		t.Setenv("TF_ACC_TIMEOUT", "2h")
	}

	// Test Prometheus Integration
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for Prometheus
			{
				Config: testAccIntegrationPrometheusConfig(envId, deployProfile),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("automq_integration.test_prometheus", "id"),
					resource.TestCheckResourceAttr("automq_integration.test_prometheus", "environment_id", envId),
					resource.TestCheckResourceAttr("automq_integration.test_prometheus", "deploy_profile", deployProfile),
					resource.TestCheckResourceAttr("automq_integration.test_prometheus", "type", "prometheusRemoteWrite"),
					resource.TestCheckResourceAttr("automq_integration.test_prometheus", "name", "test-prometheus"),
					resource.TestCheckResourceAttrSet("automq_integration.test_prometheus", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "automq_integration.test_prometheus",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccIntegrationPrometheusConfigUpdate(envId, deployProfile),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("automq_integration.test_prometheus", "name", "test-prometheus-updated"),
				),
			},
		},
	})

	// Test CloudWatch Integration
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for CloudWatch
			{
				Config: testAccIntegrationCloudWatchConfig(envId, deployProfile),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("automq_integration.test_cloudwatch", "id"),
					resource.TestCheckResourceAttr("automq_integration.test_cloudwatch", "environment_id", envId),
					resource.TestCheckResourceAttr("automq_integration.test_cloudwatch", "deploy_profile", deployProfile),
					resource.TestCheckResourceAttr("automq_integration.test_cloudwatch", "type", "cloudWatch"),
					resource.TestCheckResourceAttr("automq_integration.test_cloudwatch", "name", "test-cloudwatch"),
					resource.TestCheckResourceAttrSet("automq_integration.test_cloudwatch", "created_at"),
				),
			},
		},
	})
}

func testAccIntegrationPrometheusConfig(envId, deployProfile string) string {
	return fmt.Sprintf(`
resource "automq_integration" "test_prometheus" {
  environment_id   = %[1]q
  deploy_profile   = %[2]q
  name            = "test-prometheus"
  type            = "prometheusRemoteWrite"
  endpoint        = "http://localhost:9090/api/v1/write"
  prometheus_remote_write_config = {
    auth_type     = "basic"
    username      = "user"
    password      = "pass"
  }
}
`, envId, deployProfile)
}

func testAccIntegrationPrometheusConfigUpdate(envId, deployProfile string) string {
	return fmt.Sprintf(`
resource "automq_integration" "test_prometheus" {
  environment_id   = %[1]q
  deploy_profile   = %[2]q
  name            = "test-prometheus-updated"
  type            = "prometheusRemoteWrite"
  endpoint        = "http://localhost:9090/api/v1/write"
  prometheus_remote_write_config = {
    auth_type     = "basic"
    username      = "user"
    password      = "pass"
  }
}
`, envId, deployProfile)
}

func testAccIntegrationCloudWatchConfig(envId, deployProfile string) string {
	return fmt.Sprintf(`
resource "automq_integration" "test_cloudwatch" {
  environment_id   = %[1]q
  deploy_profile   = %[2]q
  name            = "test-cloudwatch"
  type            = "cloudWatch"
  cloudwatch_config {
    namespace     = "AutoMQ/Test"
  }
}
`, envId, deployProfile)
}
