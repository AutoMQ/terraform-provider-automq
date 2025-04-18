package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	endpoint := os.Getenv("AUTOMQ_BYOC_ENDPOINT")
	if endpoint == "" {
		t.Skip("AUTOMQ_BYOC_ENDPOINT must be set for this test")
	}
	accessKeyId := os.Getenv("AUTOMQ_BYOC_ACCESS_KEY_ID")
	if accessKeyId == "" {
		t.Skip("AUTOMQ_TEST_BYOC_ACCESS_KEY_ID must be set for this test")
	}
	secretKey := os.Getenv("AUTOMQ_BYOC_SECRET_KEY")
	if secretKey == "" {
		t.Skip("AUTOMQ_TEST_BYOC_SECRET_KEY must be set for this test")
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
				Config: testAccIntegrationPrometheusConfig(envId, deployProfile, endpoint, accessKeyId, secretKey),
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
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["automq_integration.test_prometheus"]
					if !ok {
						return "", fmt.Errorf("Not found: %s", "automq_integration.test_prometheus")
					}
					id := fmt.Sprintf("%s@%s", rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["id"])
					return id, nil
				},
				ImportStateVerifyIgnore: []string{
					"created_at",
					"last_updated",
				},
			},
			// Update testing
			{
				Config: testAccIntegrationPrometheusConfigUpdate(envId, deployProfile, endpoint, accessKeyId, secretKey),
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
				Config: testAccIntegrationCloudWatchConfig(envId, deployProfile, endpoint, accessKeyId, secretKey),
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

func testAccIntegrationPrometheusConfig(envId, deployProfile string, endpoint string, ak string, sk string) string {
	return fmt.Sprintf(`
provider "automq" {
  automq_byoc_endpoint     = "%[3]s"
  automq_byoc_access_key_id = "%[4]s"
  automq_byoc_secret_key   = "%[5]s"
}

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
`, envId, deployProfile, endpoint, ak, sk)
}

func testAccIntegrationPrometheusConfigUpdate(envId, deployProfile string, endpoint string, ak string, sk string) string {
	return fmt.Sprintf(`
provider "automq" {
  automq_byoc_endpoint     = "%[3]s"
  automq_byoc_access_key_id = "%[4]s"
  automq_byoc_secret_key   = "%[5]s"
}

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
`, envId, deployProfile, endpoint, ak, sk)
}

func testAccIntegrationCloudWatchConfig(envId, deployProfile string, endpoint string, ak string, sk string) string {
	return fmt.Sprintf(`
provider "automq" {
  automq_byoc_endpoint     = "%[3]s"
  automq_byoc_access_key_id = "%[4]s"
  automq_byoc_secret_key   = "%[5]s"
}
  
resource "automq_integration" "test_cloudwatch" {
  environment_id   = %[1]q
  deploy_profile   = %[2]q
  name            = "test-cloudwatch"
  type            = "cloudWatch"
  cloudwatch_config =  {
    namespace     = "AutoMQ/Test"
  }
}
`, envId, deployProfile, endpoint, ak, sk)
}
