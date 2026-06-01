package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccConnectorPluginResource_basic(t *testing.T) {
	env := loadAccConfig(t)
	if !env.hasConnectorPlugin() {
		t.Skip("Skipping connector plugin acceptance tests: acc.config.connector_plugin section not configured")
	}
	ensureAccTimeout(t)

	suffix := generateRandomSuffix()
	pluginName := fmt.Sprintf("acc-plugin-%s", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: renderConnectorPluginConfig(env, pluginName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("automq_connector_plugin.test", "name", pluginName),
					resource.TestCheckResourceAttr("automq_connector_plugin.test", "version", env.ConnectorPlugin.Version),
					resource.TestCheckResourceAttr("automq_connector_plugin.test", "connector_class", env.ConnectorPlugin.ConnectorClass),
					resource.TestCheckResourceAttrSet("automq_connector_plugin.test", "id"),
					resource.TestCheckResourceAttrSet("automq_connector_plugin.test", "status"),
					resource.TestCheckResourceAttr("automq_connector_plugin.test", "plugin_provider", "CUSTOM"),
				),
			},
			// ImportState
			{
				ResourceName:      "automq_connector_plugin.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"timeouts",
				},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["automq_connector_plugin.test"]
					if !ok {
						return "", fmt.Errorf("resource automq_connector_plugin.test not found")
					}
					return rs.Primary.Attributes["environment_id"] + "@" + rs.Primary.Attributes["id"], nil
				},
			},
		},
	})
}

func renderConnectorPluginConfig(env accConfig, name string) string {
	cp := env.ConnectorPlugin

	optionalFields := ""
	if cp.Description != "" {
		optionalFields += fmt.Sprintf("  description        = %q\n", cp.Description)
	}
	if cp.DocumentationLink != "" {
		optionalFields += fmt.Sprintf("  documentation_link = %q\n", cp.DocumentationLink)
	}

	return fmt.Sprintf(`
provider "automq" {
  automq_byoc_endpoint      = %q
  automq_byoc_access_key_id = %q
  automq_byoc_secret_key    = %q
}

resource "automq_connector_plugin" "test" {
  environment_id  = %q
  name            = %q
  version         = %q
  storage_url     = %q
  types           = %s
  connector_class = %q
%s
  timeouts = {
    create = "10m"
    delete = "10m"
  }
}
`, env.Endpoint, env.AccessKeyID, env.SecretKey,
		env.EnvironmentID, name, cp.Version, cp.StorageUrl,
		renderHCLStringList(cp.Types), cp.ConnectorClass,
		optionalFields)
}

func renderHCLStringList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	result := "["
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%q", item)
	}
	result += "]"
	return result
}
