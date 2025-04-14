package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeployProfileDataSource(t *testing.T) {
	envVars := getRequiredEnvVars(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDeployProfileDataSourceConfig(envVars),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify basic attributes
					resource.TestCheckResourceAttr("data.automq_deploy_profile.test", "name", envVars["AUTOMQ_TEST_DEPLOY_PROFILE"]),
					resource.TestCheckResourceAttr("data.automq_deploy_profile.test", "environment_id", envVars["AUTOMQ_TEST_ENV_ID"]),

					// Verify computed fields are set
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "cloud_provider"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "region"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "vpc"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "instance_platform"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "gmt_create"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "gmt_modified"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "available"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "system"),

					// Verify ops bucket
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "ops_bucket.bucket_name"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "ops_bucket.provider"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "ops_bucket.region"),

					// Verify data buckets
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "data_buckets.#"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "data_buckets.0.id"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "data_buckets.0.bucket_name"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "data_buckets.0.provider"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "data_buckets.0.region"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "data_buckets.0.gmt_create"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "data_buckets.0.gmt_modified"),

					// Verify DNS and instance profile
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "dns_zone"),
					resource.TestCheckResourceAttrSet("data.automq_deploy_profile.test", "instance_profile"),
				),
			},
		},
	})
}

func testAccDeployProfileDataSourceConfig(envVars map[string]string) string {
	return `
provider "automq" {
  automq_byoc_endpoint      = "` + envVars["AUTOMQ_BYOC_ENDPOINT"] + `"
  automq_byoc_access_key_id = "` + envVars["AUTOMQ_BYOC_ACCESS_KEY_ID"] + `"
  automq_byoc_secret_key    = "` + envVars["AUTOMQ_BYOC_SECRET_KEY"] + `"
}

data "automq_deploy_profile" "test" {
  environment_id = "` + envVars["AUTOMQ_TEST_ENV_ID"] + `"
  name          = "` + envVars["AUTOMQ_TEST_DEPLOY_PROFILE"] + `"
}
`
}
