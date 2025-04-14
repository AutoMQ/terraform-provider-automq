// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"automq": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	envVars := []string{
		"AUTOMQ_BYOC_ENDPOINT",
		"AUTOMQ_BYOC_ACCESS_KEY_ID",
		"AUTOMQ_BYOC_SECRET_KEY",
		"AUTOMQ_TEST_ENV_ID",
		"AUTOMQ_TEST_SUBNET_ID",
		"AUTOMQ_TEST_ZONE",
		"AUTOMQ_TEST_DEPLOY_PROFILE",
	}

	for _, v := range envVars {
		if os.Getenv(v) == "" {
			t.Fatalf("%s must be set for acceptance tests", v)
		}
	}
}
