package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKafkaInstanceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKafkaInstanceResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "display_name", "test"),
				),
			},
		},
	})
}

func testAccKafkaInstanceResourceConfig() string {
	return fmt.Sprintf(`
resource "automq_kafka_instance" "test" {
  display_name   = "test"
  description    = "test"
  cloud_provider = "aliyun"
  region         = "cn-hangzhou"
  network_type   = "vpc"
  networks = [{
    zone   = "cn-hangzhou-b"
    subnet = "vsw-bp14v5eikr8wrgoqje7hr"
  }]
  compute_specs = {
    aku = "6"
  }
}
`)
}
