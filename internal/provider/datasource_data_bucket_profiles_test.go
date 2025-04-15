package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataBucketProfilesDataSource(t *testing.T) {
	envId := os.Getenv("AUTOMQ_TEST_ENV_ID")
	if envId == "" {
		t.Skip("AUTOMQ_TEST_ENV_ID must be set for this test")
	}

	envVars := getRequiredEnvVars(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataBucketProfilesDataSourceConfig(envVars),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that the attributes exist and have non-empty values
					resource.TestCheckResourceAttrSet("data.automq_data_bucket_profiles.test", "environment_id"),
					resource.TestCheckResourceAttrSet("data.automq_data_bucket_profiles.test", "profile_name"),
					// Verify the values match our environment variables
					resource.TestCheckResourceAttr("data.automq_data_bucket_profiles.test", "environment_id", envVars["AUTOMQ_TEST_ENV_ID"]),
					resource.TestCheckResourceAttr("data.automq_data_bucket_profiles.test", "profile_name", envVars["AUTOMQ_TEST_DEPLOY_PROFILE"]),
					// Verify that data_buckets list exists and has elements
					resource.TestCheckResourceAttrSet("data.automq_data_bucket_profiles.test", "data_buckets.#"),
				),
			},
		},
	})
}

func testAccDataBucketProfilesDataSourceConfig(envVars map[string]string) string {
	return `
provider "automq" {
  automq_byoc_endpoint      = "` + envVars["AUTOMQ_BYOC_ENDPOINT"] + `"
  automq_byoc_access_key_id = "` + envVars["AUTOMQ_BYOC_ACCESS_KEY_ID"] + `"
  automq_byoc_secret_key    = "` + envVars["AUTOMQ_BYOC_SECRET_KEY"] + `"
}

data "automq_data_bucket_profiles" "test" {
  environment_id = "` + envVars["AUTOMQ_TEST_ENV_ID"] + `"
  profile_name   = "` + envVars["AUTOMQ_TEST_DEPLOY_PROFILE"] + `"
}
`
}
