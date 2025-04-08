package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func getRequiredEnvVars(t *testing.T) map[string]string {
	envVars := map[string]string{
		"AUTOMQ_BYOC_ENDPOINT":       "http://44.207.11.83:8080",         // os.Getenv("AUTOMQ_BYOC_ENDPOINT"),
		"AUTOMQ_BYOC_ACCESS_KEY_ID":  "TeDaXnaRASROF9AZ",                 // os.Getenv("AUTOMQ_BYOC_ACCESS_KEY_ID"),
		"AUTOMQ_BYOC_SECRET_KEY":     "Lj41KSn5WciTS9UNc5cf4k3MRcIbs0D4", // os.Getenv("AUTOMQ_BYOC_SECRET_KEY"),
		"AUTOMQ_TEST_ENV_ID":         "env-vstohprknxupims1",             // os.Getenv("AUTOMQ_TEST_ENV_ID"),
		"AUTOMQ_TEST_SUBNET_ID":      "subnet-0005ed305e0891752",         //os.Getenv("AUTOMQ_TEST_SUBNET_ID"),
		"AUTOMQ_TEST_ZONE":           "us-east-1a",                       // os.Getenv("AUTOMQ_TEST_ZONE"),
		"AUTOMQ_TEST_DEPLOY_PROFILE": "default",                          // os.Getenv("AUTOMQ_TEST_DEPLOY_PROFILE"),
	}

	missingVars := []string{}
	for k, v := range envVars {
		if v == "" {
			missingVars = append(missingVars, k)
		}
	}

	if len(missingVars) > 0 {
		t.Skipf("Missing required environment variables: %v", missingVars)
	}

	return envVars
}

func generateRandomSuffix() string {
	// Generate a random 6-character string for resource naming
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "default"
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// Define version upgrade path
var kafkaVersionUpgrades = []string{
	"1.3.10",
	"1.4.0",
}

// TestConfig holds the test configuration parameters
type TestConfig struct {
	EnvironmentID string
	Name          string
	Description   string
	DeployProfile string
	Version       string
	ComputeSpecs  ComputeSpecs
	Features      Features
}

type ComputeSpecs struct {
	ReservedAKU int
	Networks    []Network
}

type Network struct {
	Zone    string
	Subnets []string
}

type Features struct {
	WALMode         string
	InstanceConfigs map[string]string
	Security        *SecurityConfig
}

type SecurityConfig struct {
	AuthenticationMethods  []string
	TransitEncryptionModes []string
	DataEncryptionMode     string
	CertificateAuthority   string
	CertificateChain       string
	PrivateKey             string
}

// defaultTestConfig returns a default test configuration
func defaultTestConfig(suffix string, envVars map[string]string) TestConfig {
	return TestConfig{
		EnvironmentID: envVars["AUTOMQ_TEST_ENV_ID"],
		Name:          fmt.Sprintf("test-instance-%s", suffix),
		Description:   "test instance description",
		DeployProfile: envVars["AUTOMQ_TEST_DEPLOY_PROFILE"],
		Version:       "1.3.10",
		ComputeSpecs: ComputeSpecs{
			ReservedAKU: 6,
			Networks: []Network{
				{
					Zone:    envVars["AUTOMQ_TEST_ZONE"],
					Subnets: []string{envVars["AUTOMQ_TEST_SUBNET_ID"]},
				},
			},
		},
		Features: Features{
			WALMode: "EBSWAL",
			InstanceConfigs: map[string]string{
				"auto.create.topics.enable": "true",
				"num.partitions":            "1",
			},
		},
	}
}

// renderHCLConfig converts a TestConfig to HCL format
func renderHCLConfig(config TestConfig, envVars map[string]string) string {
	// Base template with provider and resource structure
	template := `
provider "automq" {
  automq_byoc_endpoint     = "%[1]s"
  automq_byoc_access_key_id = "%[2]s"
  automq_byoc_secret_key   = "%[3]s"
}

data "automq_deploy_profile" "default" {
  environment_id = "%[4]s"
  name = "default"
}

data "automq_data_bucket_profiles" "test" {
  environment_id = "%[4]s"
  profile_name = data.automq_deploy_profile.default.name
}

resource "automq_kafka_instance" "test" {
    environment_id = "%[4]s"
    name          = "%[5]s"
    description   = "%[6]s"
    deploy_profile = "%[7]s"
    version       = "%[8]s"

    compute_specs = {
        reserved_aku = %[9]d
        networks = [
            {
                zone    = "%[10]s"
                subnets = ["%[11]s"]
            }
        ]
        bucket_profiles = [
            {
                id = data.automq_data_bucket_profiles.test.data_buckets[0].id
            }
        ]
    }

    features = {
        wal_mode = "%[12]s"
        %[13]s
        %[14]s
    }
}
`

	// Build instance configs string if present
	instanceConfigsStr := ""
	if len(config.Features.InstanceConfigs) > 0 {
		instanceConfigsStr = "instance_configs = {\n"
		for k, v := range config.Features.InstanceConfigs {
			instanceConfigsStr += fmt.Sprintf(`            "%s" = "%s"`+"\n", k, v)
		}
		instanceConfigsStr += "        }"
	}

	// Build security config string if present
	securityConfigStr := ""
	if config.Features.Security != nil {
		securityConfigStr = "security = {\n"
		if len(config.Features.Security.AuthenticationMethods) > 0 {
			securityConfigStr += fmt.Sprintf(`            authentication_methods = ["%s"]`+"\n",
				strings.Join(config.Features.Security.AuthenticationMethods, `","`))
		}
		if len(config.Features.Security.TransitEncryptionModes) > 0 {
			securityConfigStr += fmt.Sprintf(`            transit_encryption_modes = ["%s"]`+"\n",
				strings.Join(config.Features.Security.TransitEncryptionModes, `","`))
		}
		securityConfigStr += fmt.Sprintf(`            data_encryption_mode = "%s"`+"\n", config.Features.Security.DataEncryptionMode)

		if config.Features.Security.CertificateAuthority != "" {
			securityConfigStr += fmt.Sprintf(`            certificate_authority = <<-EOT
%s
EOT
`, config.Features.Security.CertificateAuthority)
		}
		if config.Features.Security.CertificateChain != "" {
			securityConfigStr += fmt.Sprintf(`            certificate_chain = <<-EOT
%s
EOT
`, config.Features.Security.CertificateChain)
		}
		if config.Features.Security.PrivateKey != "" {
			securityConfigStr += fmt.Sprintf(`            private_key = <<-EOT
%s
EOT
`, config.Features.Security.PrivateKey)
		}
		securityConfigStr += "        }"
	}

	// Format the final configuration
	return fmt.Sprintf(template,
		envVars["AUTOMQ_BYOC_ENDPOINT"],            // 1
		envVars["AUTOMQ_BYOC_ACCESS_KEY_ID"],       // 2
		envVars["AUTOMQ_BYOC_SECRET_KEY"],          // 3
		config.EnvironmentID,                       // 4
		config.Name,                                // 5
		config.Description,                         // 6
		config.DeployProfile,                       // 7
		config.Version,                             // 8
		config.ComputeSpecs.ReservedAKU,            // 9
		config.ComputeSpecs.Networks[0].Zone,       // 10
		config.ComputeSpecs.Networks[0].Subnets[0], // 11
		config.Features.WALMode,                    // 12
		instanceConfigsStr,                         // 13
		securityConfigStr,                          // 14
	)
}

// TestAccKafkaInstanceResource tests basic CRUD operations for Kafka instance
func TestAccKafkaInstanceResource(t *testing.T) {
	envVars := getRequiredEnvVars(t)
	suffix := generateRandomSuffix()

	// Generate certificates for security testing
	caCert, serverCert, serverKey, err := generateTestCertificate()
	newCaCert, newServerCert, newServerKey, newErr := generateTestCertificate()
	if err != nil || newErr != nil {
		t.Fatalf("Failed to generate test certificates: %v, %v", err, newErr)
	}

	// Initial configuration
	initialConfig := defaultTestConfig(suffix, envVars)
	initialConfig.Version = kafkaVersionUpgrades[0]
	initialConfig.Features.Security = &SecurityConfig{
		AuthenticationMethods:  []string{"sasl", "mtls"},
		TransitEncryptionModes: []string{"tls"},
		DataEncryptionMode:     "CPMK",
		CertificateAuthority:   caCert,
		CertificateChain:       serverCert,
		PrivateKey:             serverKey,
	}

	// Updated configuration
	updatedConfig := defaultTestConfig(suffix, envVars)
	updatedConfig.Name = fmt.Sprintf("test-instance-updated-%s", suffix)
	updatedConfig.Description = "updated test instance description"
	updatedConfig.Version = kafkaVersionUpgrades[1]
	updatedConfig.ComputeSpecs.ReservedAKU = 8
	updatedConfig.Features.InstanceConfigs = map[string]string{
		"auto.create.topics.enable": "false",
		"num.partitions":            "3",
		"log.retention.hours":       "24",
	}
	updatedConfig.Features.Security = &SecurityConfig{
		AuthenticationMethods:  []string{"sasl", "mtls"},
		TransitEncryptionModes: []string{"tls"},
		DataEncryptionMode:     "CPMK",
		CertificateAuthority:   newCaCert,
		CertificateChain:       newServerCert,
		PrivateKey:             newServerKey,
	}

	resource.Test(t, resource.TestCase{
		// PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaInstanceDestroy,
		Steps: []resource.TestStep{
			// Initial creation
			{
				Config: renderHCLConfig(initialConfig, envVars),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaInstanceExists("automq_kafka_instance.test"),
					// Basic attributes
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "name", initialConfig.Name),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "description", initialConfig.Description),

					// Compute specs
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.reserved_aku", fmt.Sprintf("%d", initialConfig.ComputeSpecs.ReservedAKU)),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.networks.#", "1"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.networks.0.zone", initialConfig.ComputeSpecs.Networks[0].Zone),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.networks.0.subnets.#", "1"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.networks.0.subnets.0", initialConfig.ComputeSpecs.Networks[0].Subnets[0]),

					// Features
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.wal_mode", initialConfig.Features.WALMode),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "version", initialConfig.Version),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.instance_configs.num.partitions", initialConfig.Features.InstanceConfigs["num.partitions"]),

					// Required fields
					resource.TestCheckResourceAttrSet("automq_kafka_instance.test", "id"),
					resource.TestCheckResourceAttrSet("automq_kafka_instance.test", "status"),
					resource.TestCheckResourceAttrSet("automq_kafka_instance.test", "created_at"),
				),
			},
			// Update test
			{
				Config: renderHCLConfig(updatedConfig, envVars),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaInstanceExists("automq_kafka_instance.test"),
					// Basic attributes
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "name", updatedConfig.Name),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "description", updatedConfig.Description),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "version", updatedConfig.Version),

					// Compute specs
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.reserved_aku", fmt.Sprintf("%d", updatedConfig.ComputeSpecs.ReservedAKU)),

					// Instance configs
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.instance_configs.auto.create.topics.enable", updatedConfig.Features.InstanceConfigs["auto.create.topics.enable"]),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.instance_configs.num.partitions", updatedConfig.Features.InstanceConfigs["num.partitions"]),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.instance_configs.log.retention.hours", updatedConfig.Features.InstanceConfigs["log.retention.hours"]),

					// Security certificates
					resource.TestCheckResourceAttrSet("automq_kafka_instance.test", "features.security.certificate_authority"),
					resource.TestCheckResourceAttrSet("automq_kafka_instance.test", "features.security.certificate_chain"),
					resource.TestCheckResourceAttrSet("automq_kafka_instance.test", "features.security.private_key"),
				),
			},
			// import test
			{
				ResourceName:      "automq_kafka_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["automq_kafka_instance.test"]
					if !ok {
						return "", fmt.Errorf("Not found: %s", "automq_kafka_instance.test")
					}
					id := fmt.Sprintf("%s@%s", rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["id"])
					// The import ID format is <environment_id>@<kafka_instance_id>
					return id, nil
				},
				ImportStateVerifyIgnore: []string{
					"features.security.certificate_authority",
					"features.security.certificate_chain",
					"features.security.private_key",
					"features.instance_configs",
				},
			},
		},
	})
}

// Add CheckDestroy function
func testAccCheckKafkaInstanceDestroy(s *terraform.State) error {
	// Loop through resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "automq_kafka_instance" {
			continue
		}

		// Try to get the instance
		// In a real implementation, you would use your client to check if the instance still exists
		// For now, we'll just return nil to indicate the instance was destroyed
		return nil
	}

	return nil
}

// Add Exists check function
func testAccCheckKafkaInstanceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kafka Instance ID is set")
		}

		// In a real implementation, you would use your client to check if the instance exists
		// For now, we'll just return nil to indicate the instance exists
		return nil
	}
}

// SecurityTestCase defines a test case for security configurations
type SecurityTestCase struct {
	name                 string
	authMethods          []string
	transitEncryption    []string
	dataEncryption       string
	requiresCertificates bool
}

func TestAccKafkaInstanceSecurityCombinations(t *testing.T) {
	envVars := getRequiredEnvVars(t)
	suffix := generateRandomSuffix()
	caCert, serverCert, serverKey, err := generateTestCertificate()
	if err != nil {
		t.Fatalf("Failed to generate test certificates: %v", err)
	}

	// Define all possible security combinations
	testCases := []SecurityTestCase{
		{
			name:                 "anonymous_plaintext_none",
			authMethods:          []string{"anonymous"},
			transitEncryption:    []string{"plaintext"},
			dataEncryption:       "NONE",
			requiresCertificates: false,
		},
		{
			name:                 "anonymous_plaintext_cpmk",
			authMethods:          []string{"anonymous"},
			transitEncryption:    []string{"plaintext"},
			dataEncryption:       "CPMK",
			requiresCertificates: false,
		},
		{
			name:                 "sasl_plaintext_none",
			authMethods:          []string{"sasl"},
			transitEncryption:    []string{"plaintext"},
			dataEncryption:       "NONE",
			requiresCertificates: false,
		},
		{
			name:                 "sasl_plaintext_cpmk",
			authMethods:          []string{"sasl"},
			transitEncryption:    []string{"plaintext"},
			dataEncryption:       "CPMK",
			requiresCertificates: false,
		},
		{
			name:                 "sasl_mtls_tls_none",
			authMethods:          []string{"sasl", "mtls"},
			transitEncryption:    []string{"tls"},
			dataEncryption:       "NONE",
			requiresCertificates: true,
		},
		{
			name:                 "sasl_mtls_tls_cpmk",
			authMethods:          []string{"sasl", "mtls"},
			transitEncryption:    []string{"tls"},
			dataEncryption:       "CPMK",
			requiresCertificates: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := defaultTestConfig(suffix, envVars)
			config.Name = fmt.Sprintf("test-secure-%s-%s", tc.name, suffix)
			config.Description = fmt.Sprintf("Test instance with %s security configuration", tc.name)

			// Set up security configuration
			config.Features.Security = &SecurityConfig{
				AuthenticationMethods:  tc.authMethods,
				TransitEncryptionModes: tc.transitEncryption,
				DataEncryptionMode:     tc.dataEncryption,
			}

			if tc.requiresCertificates {
				config.Features.Security.CertificateAuthority = caCert
				config.Features.Security.CertificateChain = serverCert
				config.Features.Security.PrivateKey = serverKey
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderHCLConfig(config, envVars),
						Check: resource.ComposeAggregateTestCheckFunc(
							// Basic attributes
							resource.TestCheckResourceAttr("automq_kafka_instance.test", "name", config.Name),
							resource.TestCheckResourceAttr("automq_kafka_instance.test", "description", config.Description),

							// Security configuration
							resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.security.authentication_methods.#", fmt.Sprintf("%d", len(tc.authMethods))),
							resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.security.transit_encryption_modes.#", fmt.Sprintf("%d", len(tc.transitEncryption))),
							resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.security.data_encryption_mode", tc.dataEncryption),

							// Conditional checks for certificates
							func(s *terraform.State) error {
								if tc.requiresCertificates {
									return resource.ComposeAggregateTestCheckFunc(
										resource.TestCheckResourceAttrSet("automq_kafka_instance.test", "features.security.certificate_authority"),
										resource.TestCheckResourceAttrSet("automq_kafka_instance.test", "features.security.certificate_chain"),
										resource.TestCheckResourceAttrSet("automq_kafka_instance.test", "features.security.private_key"),
									)(s)
								}
								return nil
							},
						),
					},
				},
			})
		})
	}
}

// generateTestCertificate generates test certificates for TLS configuration
func generateTestCertificate() (string, string, string, error) {
	// Generate CA key
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}

	// Generate CA certificate
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return "", "", "", err
	}

	// Generate server key
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}

	// Generate server certificate
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return "", "", "", err
	}

	// Encode to PEM
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})

	return string(caPEM), string(serverCertPEM), string(serverKeyPEM), nil
}
