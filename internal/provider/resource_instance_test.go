package provider

// Acceptance test configuration
//
// Scenarios & required inputs:
//   VM tests (e.g., TestAccKafkaInstance_VM_Update, _ImmutableFields, _SecurityCombinations)
//     - endpoint, access_key_id, secret_key, environment_id
//     - vm.zone (or AUTMQ_VM_ZONE) and at least one subnet ID (AUTMQ_VM_SUBNET_IDS / AUTMQ_VM_SUBNET_ID / AUTMQ_TEST_SUBNET_ID)
//
//   K8S tests (TestAccKafkaInstance_K8S_Basic)
//     - All VM requirements
//     - k8s.cluster_id, k8s.node_groups (comma separated), optional namespace/service_account
//
// Configuration sources (priority order):
//   1. JSON file provided via -acc.config flag (see accConfigFile struct for schema)
//   2. Environment variables listed above (legacy names fall back to AUTMQ_TEST_*)
//
// The helper loadAccConfig consolidates these sources and skips tests when required
// parameters are absent instead of using placeholder values.

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	envEndpoint            = "AUTOMQ_BYOC_ENDPOINT"
	envAccessKeyID         = "AUTOMQ_BYOC_ACCESS_KEY_ID"
	envSecretKey           = "AUTOMQ_BYOC_SECRET_KEY"
	envEnvironmentID       = "AUTOMQ_ENVIRONMENT_ID"
	envEnvironmentIDLegacy = "AUTOMQ_TEST_ENV_ID"
	envRegion              = "AUTOMQ_REGION"
	envVMZone              = "AUTOMQ_VM_ZONE"
	envVMZoneLegacy        = "AUTOMQ_TEST_ZONE"
	envVMSubnets           = "AUTOMQ_VM_SUBNET_IDS"
	envVMSubnetSingle      = "AUTOMQ_VM_SUBNET_ID"
	envVMSubnetLegacy      = "AUTOMQ_TEST_SUBNET_ID"
	envK8SClusterID        = "AUTOMQ_K8S_CLUSTER_ID"
	envK8SNodeGroups       = "AUTOMQ_K8S_NODE_GROUPS"
	envK8SNamespace        = "AUTOMQ_K8S_NAMESPACE"
	envK8SServiceAccount   = "AUTOMQ_K8S_SERVICE_ACCOUNT"
	envDeployProfileLegacy = "AUTOMQ_TEST_DEPLOY_PROFILE"
	defaultInstanceVersion = "5.2.0"
)

var accConfigPath = flag.String("acc.config", "", "path to JSON file containing acceptance test configuration")

type accConfig struct {
	Endpoint          string   `json:"endpoint"`
	AccessKeyID       string   `json:"access_key_id"`
	SecretKey         string   `json:"secret_key"`
	EnvironmentID     string   `json:"environment_id"`
	Region            string   `json:"region"`
	VMZone            string   `json:"vm_zone"`
	VMSubnets         []string `json:"vm_subnets"`
	K8SClusterID      string   `json:"k8s_cluster_id"`
	K8SNodeGroups     []string `json:"k8s_node_groups"`
	K8SNamespace      string   `json:"k8s_namespace"`
	K8SServiceAccount string   `json:"k8s_service_account"`
}

var (
	configOnce sync.Once
	configData accConfig
	configErr  error
)

type accNetwork struct {
	Zone    string
	Subnets []string
}

type accTableTopic struct {
	Warehouse    string
	CatalogType  string
	MetastoreURI string
}

type accMetricsExporter struct {
	Enabled       bool
	AuthType      string
	EndPoint      string
	PrometheusARN string
	Username      string
	Password      string
	Token         string
	Labels        map[string]string
}

type accSecurity struct {
	AuthenticationMethods  []string
	TransitEncryptionModes []string
	DataEncryptionMode     string
	CertificateAuthority   string
	CertificateChain       string
	PrivateKey             string
}

type accInstanceConfig struct {
	EnvironmentID         string
	Name                  string
	Description           string
	Version               string
	DeployType            string
	ReservedAKU           int
	Networks              []accNetwork
	KubernetesClusterID   string
	KubernetesNodeGroups  []string
	KubernetesNamespace   string
	KubernetesServiceAcct string
	InstanceRole          string
	InstanceConfigs       map[string]string
	TableTopic            *accTableTopic
	MetricsExporter       *accMetricsExporter
	Security              *accSecurity
	WalMode               string
}

func loadAccConfig(t *testing.T) accConfig {
	configOnce.Do(func() {
		if *accConfigPath != "" {
			configData, configErr = parseAccConfigFromFile(*accConfigPath)
		} else {
			configData = parseAccConfigFromEnv()
		}
		if configErr == nil {
			if missing := configData.missingRequired(); len(missing) > 0 {
				configErr = fmt.Errorf("missing required configuration values: %v", missing)
			}
		}
	})
	if configErr != nil {
		t.Skipf("Skipping acceptance tests: %v", configErr)
	}
	return configData
}

func parseAccConfigFromFile(path string) (accConfig, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return accConfig{}, fmt.Errorf("resolve acc.config path: %w", err)
	}
	data, err := os.ReadFile(abspath)
	if err != nil {
		return accConfig{}, fmt.Errorf("read acc.config file: %w", err)
	}
	var raw struct {
		Endpoint      string `json:"endpoint"`
		AccessKeyID   string `json:"access_key_id"`
		SecretKey     string `json:"secret_key"`
		EnvironmentID string `json:"environment_id"`
		Region        string `json:"region"`
		VM            struct {
			Zone    string   `json:"zone"`
			Subnets []string `json:"subnets"`
		} `json:"vm"`
		VMZone    string   `json:"vm_zone"`
		VMSubnets []string `json:"vm_subnets"`
		K8S       struct {
			ClusterID      string   `json:"cluster_id"`
			NodeGroups     []string `json:"node_groups"`
			Namespace      string   `json:"namespace"`
			ServiceAccount string   `json:"service_account"`
		} `json:"k8s"`
		K8SClusterID      string   `json:"k8s_cluster_id"`
		K8SNodeGroups     []string `json:"k8s_node_groups"`
		K8SNamespace      string   `json:"k8s_namespace"`
		K8SServiceAccount string   `json:"k8s_service_account"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return accConfig{}, fmt.Errorf("parse acc.config JSON: %w", err)
	}
	cfg := accConfig{
		Endpoint:          strings.TrimSpace(raw.Endpoint),
		AccessKeyID:       strings.TrimSpace(raw.AccessKeyID),
		SecretKey:         strings.TrimSpace(raw.SecretKey),
		EnvironmentID:     strings.TrimSpace(raw.EnvironmentID),
		Region:            strings.TrimSpace(raw.Region),
		VMZone:            strings.TrimSpace(raw.VMZone),
		VMSubnets:         append([]string{}, raw.VMSubnets...),
		K8SClusterID:      strings.TrimSpace(raw.K8SClusterID),
		K8SNodeGroups:     append([]string{}, raw.K8SNodeGroups...),
		K8SNamespace:      strings.TrimSpace(raw.K8SNamespace),
		K8SServiceAccount: strings.TrimSpace(raw.K8SServiceAccount),
	}
	if cfg.VMZone == "" {
		cfg.VMZone = strings.TrimSpace(raw.VM.Zone)
	}
	if len(cfg.VMSubnets) == 0 && len(raw.VM.Subnets) > 0 {
		cfg.VMSubnets = append([]string{}, raw.VM.Subnets...)
	}
	if cfg.K8SClusterID == "" {
		cfg.K8SClusterID = strings.TrimSpace(raw.K8S.ClusterID)
	}
	if len(cfg.K8SNodeGroups) == 0 && len(raw.K8S.NodeGroups) > 0 {
		cfg.K8SNodeGroups = append([]string{}, raw.K8S.NodeGroups...)
	}
	if cfg.K8SNamespace == "" {
		cfg.K8SNamespace = strings.TrimSpace(raw.K8S.Namespace)
	}
	if cfg.K8SServiceAccount == "" {
		cfg.K8SServiceAccount = strings.TrimSpace(raw.K8S.ServiceAccount)
	}
	return cfg.normalise(), nil
}

func parseAccConfigFromEnv() accConfig {
	cfg := accConfig{
		Endpoint:          strings.TrimSpace(os.Getenv(envEndpoint)),
		AccessKeyID:       strings.TrimSpace(os.Getenv(envAccessKeyID)),
		SecretKey:         strings.TrimSpace(os.Getenv(envSecretKey)),
		EnvironmentID:     strings.TrimSpace(os.Getenv(envEnvironmentID)),
		Region:            strings.TrimSpace(os.Getenv(envRegion)),
		VMZone:            strings.TrimSpace(os.Getenv(envVMZone)),
		K8SClusterID:      strings.TrimSpace(os.Getenv(envK8SClusterID)),
		K8SNamespace:      strings.TrimSpace(os.Getenv(envK8SNamespace)),
		K8SServiceAccount: strings.TrimSpace(os.Getenv(envK8SServiceAccount)),
	}
	if cfg.EnvironmentID == "" {
		cfg.EnvironmentID = strings.TrimSpace(os.Getenv(envEnvironmentIDLegacy))
	}
	if cfg.VMZone == "" {
		cfg.VMZone = strings.TrimSpace(os.Getenv(envVMZoneLegacy))
	}
	vmSubnetsRaw := strings.TrimSpace(os.Getenv(envVMSubnets))
	if vmSubnetsRaw == "" {
		vmSubnetsRaw = strings.TrimSpace(os.Getenv(envVMSubnetSingle))
	}
	if vmSubnetsRaw == "" {
		vmSubnetsRaw = strings.TrimSpace(os.Getenv(envVMSubnetLegacy))
	}
	cfg.VMSubnets = splitAndTrim(vmSubnetsRaw)
	cfg.K8SNodeGroups = splitAndTrim(os.Getenv(envK8SNodeGroups))
	return cfg.normalise()
}

func (c accConfig) normalise() accConfig {
	if len(c.VMSubnets) > 0 {
		c.VMSubnets = splitAndTrim(strings.Join(c.VMSubnets, ","))
	}
	if len(c.K8SNodeGroups) > 0 {
		c.K8SNodeGroups = splitAndTrim(strings.Join(c.K8SNodeGroups, ","))
	}
	return c
}

func (c accConfig) missingRequired() []string {
	missing := make([]string, 0, 4)
	if c.Endpoint == "" {
		missing = append(missing, envEndpoint)
	}
	if c.AccessKeyID == "" {
		missing = append(missing, envAccessKeyID)
	}
	if c.SecretKey == "" {
		missing = append(missing, envSecretKey)
	}
	if c.EnvironmentID == "" {
		missing = append(missing, envEnvironmentID)
	}
	return missing
}

func (c accConfig) requireVM(t *testing.T) {
	if c.VMZone == "" || len(c.VMSubnets) == 0 {
		t.Skip("Skipping VM acceptance tests: vm.zone/subnets not configured")
	}
}

func (c accConfig) hasK8S() bool {
	return c.K8SClusterID != "" && len(c.K8SNodeGroups) > 0
}

func splitAndTrim(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func TestAccKafkaInstance_VM_Update(t *testing.T) {
	env := loadAccConfig(t)
	env.requireVM(t)
	ensureAccTimeout(t)

	suffix := generateRandomSuffix()

	baseCfg := accInstanceConfig{
		EnvironmentID: env.EnvironmentID,
		Name:          fmt.Sprintf("acc-vm-%s", suffix),
		Description:   "VM acceptance baseline",
		Version:       defaultInstanceVersion,
		DeployType:    "IAAS",
		ReservedAKU:   6,
		Networks: []accNetwork{
			{
				Zone:    env.VMZone,
				Subnets: env.VMSubnets,
			},
		},
		WalMode: "EBSWAL",
		Security: &accSecurity{
			AuthenticationMethods:  []string{"sasl"},
			TransitEncryptionModes: []string{"plaintext"},
			DataEncryptionMode:     "NONE",
		},
	}

	updatedCfg := baseCfg
	updatedCfg.Name = fmt.Sprintf("acc-vm-updated-%s", suffix)
	updatedCfg.Description = "VM acceptance update"
	updatedCfg.ReservedAKU = 8
	updatedCfg.InstanceConfigs = map[string]string{
		"auto.create.topics.enable": "false",
		"num.partitions":            "3",
	}
	updatedCfg.MetricsExporter = &accMetricsExporter{
		Enabled:  true,
		EndPoint: "https://metrics.example.com/write",
		Labels: map[string]string{
			"env":  "acc",
			"team": "qa",
		},
	}

	disableMetricsCfg := updatedCfg
	disableMetricsCfg.MetricsExporter = &accMetricsExporter{Enabled: false}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); env.requireVM(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: renderKafkaInstanceConfig(env, baseCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaInstanceExists("automq_kafka_instance.test"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "name", baseCfg.Name),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "description", baseCfg.Description),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.reserved_aku", fmt.Sprintf("%d", baseCfg.ReservedAKU)),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.wal_mode", baseCfg.WalMode),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.security.authentication_methods.#", "1"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.security.data_encryption_mode", "NONE"),
				),
			},
			{
				Config: renderKafkaInstanceConfig(env, updatedCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaInstanceExists("automq_kafka_instance.test"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "name", updatedCfg.Name),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "description", updatedCfg.Description),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.reserved_aku", fmt.Sprintf("%d", updatedCfg.ReservedAKU)),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.instance_configs.num.partitions", "3"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.metrics_exporter.prometheus.enabled", "true"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.metrics_exporter.prometheus.end_point", updatedCfg.MetricsExporter.EndPoint),
				),
			},
			{
				Config: renderKafkaInstanceConfig(env, disableMetricsCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaInstanceExists("automq_kafka_instance.test"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.metrics_exporter.prometheus.enabled", "false"),
				),
			},
			{
				ResourceName:      "automq_kafka_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"features.metrics_exporter.prometheus.password",
					"features.metrics_exporter.prometheus.token",
					"features.security.private_key",
				},
			},
		},
	})
}

func TestAccKafkaInstance_K8S_Basic(t *testing.T) {
	env := loadAccConfig(t)
	if !env.hasK8S() {
		t.Skip("Skipping K8S acceptance tests: AUTMQ_K8S_* variables not configured")
	}
	ensureAccTimeout(t)

	networks := []accNetwork{{Zone: env.VMZone}}
	if len(env.VMSubnets) > 0 {
		networks[0].Subnets = env.VMSubnets
	}

	cfg := accInstanceConfig{
		EnvironmentID:         env.EnvironmentID,
		Name:                  fmt.Sprintf("acc-k8s-%s", generateRandomSuffix()),
		Description:           "K8S acceptance baseline",
		Version:               defaultInstanceVersion,
		DeployType:            "K8S",
		ReservedAKU:           6,
		Networks:              networks,
		KubernetesClusterID:   env.K8SClusterID,
		KubernetesNodeGroups:  env.K8SNodeGroups,
		KubernetesNamespace:   env.K8SNamespace,
		KubernetesServiceAcct: env.K8SServiceAccount,
		WalMode:               "S3WAL",
		Security: &accSecurity{
			AuthenticationMethods:  []string{"sasl"},
			TransitEncryptionModes: []string{"plaintext"},
			DataEncryptionMode:     "NONE",
		},
	}

	updated := cfg
	updated.ReservedAKU = 7

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: renderKafkaInstanceConfig(env, cfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaInstanceExists("automq_kafka_instance.test"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.deploy_type", "K8S"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.kubernetes_cluster_id", env.K8SClusterID),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.kubernetes_node_groups.#", fmt.Sprintf("%d", len(env.K8SNodeGroups))),
				),
			},
			{
				Config: renderKafkaInstanceConfig(env, updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.reserved_aku", fmt.Sprintf("%d", updated.ReservedAKU)),
				),
			},
		},
	})
}

func TestAccKafkaInstance_ImmutableFields(t *testing.T) {
	env := loadAccConfig(t)
	env.requireVM(t)
	ensureAccTimeout(t)

	base := accInstanceConfig{
		EnvironmentID: env.EnvironmentID,
		Name:          fmt.Sprintf("acc-immutable-%s", generateRandomSuffix()),
		Description:   "Immutable field baseline",
		Version:       defaultInstanceVersion,
		DeployType:    "IAAS",
		ReservedAKU:   6,
		Networks: []accNetwork{{
			Zone:    env.VMZone,
			Subnets: env.VMSubnets,
		}},
		WalMode:      "EBSWAL",
		InstanceRole: "role-initial",
		TableTopic: &accTableTopic{
			Warehouse:    "warehouse_immutable",
			CatalogType:  "HIVE",
			MetastoreURI: "thrift://metastore.initial:9083",
		},
		Security: &accSecurity{
			AuthenticationMethods:  []string{"sasl"},
			TransitEncryptionModes: []string{"plaintext"},
			DataEncryptionMode:     "NONE",
		},
	}

	modified := base
	modified.InstanceRole = "role-updated"
	modified.TableTopic = &accTableTopic{
		Warehouse:    "warehouse_immutable",
		CatalogType:  "HIVE",
		MetastoreURI: "thrift://metastore.updated:9083",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); env.requireVM(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: renderKafkaInstanceConfig(env, base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKafkaInstanceExists("automq_kafka_instance.test"),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "compute_specs.instance_role", base.InstanceRole),
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.table_topic.warehouse", base.TableTopic.Warehouse),
				),
			},
			{
				Config:             renderKafkaInstanceConfig(env, modified),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaInstance_SecurityCombinations(t *testing.T) {
	env := loadAccConfig(t)
	env.requireVM(t)
	ensureAccTimeout(t)

	combos := []struct {
		name     string
		security accSecurity
	}{
		{
			name: "anonymous_plaintext_none",
			security: accSecurity{
				AuthenticationMethods:  []string{"anonymous"},
				TransitEncryptionModes: []string{"plaintext"},
				DataEncryptionMode:     "NONE",
			},
		},
		{
			name: "anonymous_plaintext_cpmk",
			security: accSecurity{
				AuthenticationMethods:  []string{"anonymous"},
				TransitEncryptionModes: []string{"plaintext"},
				DataEncryptionMode:     "CPMK",
			},
		},
		{
			name: "sasl_plaintext_none",
			security: accSecurity{
				AuthenticationMethods:  []string{"sasl"},
				TransitEncryptionModes: []string{"plaintext"},
				DataEncryptionMode:     "NONE",
			},
		},
		{
			name: "sasl_plaintext_cpmk",
			security: accSecurity{
				AuthenticationMethods:  []string{"sasl"},
				TransitEncryptionModes: []string{"plaintext"},
				DataEncryptionMode:     "CPMK",
			},
		},
		{
			name: "sasl_mtls_tls_none",
			security: accSecurity{
				AuthenticationMethods:  []string{"sasl", "mtls"},
				TransitEncryptionModes: []string{"tls"},
				DataEncryptionMode:     "NONE",
			},
		},
		{
			name: "sasl_mtls_tls_cpmk",
			security: accSecurity{
				AuthenticationMethods:  []string{"sasl", "mtls"},
				TransitEncryptionModes: []string{"tls"},
				DataEncryptionMode:     "CPMK",
			},
		},
	}

	for _, combo := range combos {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			cfg := accInstanceConfig{
				EnvironmentID: env.EnvironmentID,
				Name:          fmt.Sprintf("acc-sec-%s-%s", combo.name, generateRandomSuffix()),
				Description:   fmt.Sprintf("Security combo %s", combo.name),
				Version:       defaultInstanceVersion,
				DeployType:    "IAAS",
				ReservedAKU:   6,
				Networks: []accNetwork{{
					Zone:    env.VMZone,
					Subnets: env.VMSubnets,
				}},
				WalMode:  "EBSWAL",
				Security: &combo.security,
			}

			if needsTLSCertificates(combo.security.TransitEncryptionModes) {
				ca, cert, key, err := generateTestCertificate()
				if err != nil {
					t.Fatalf("failed to generate test certificates: %v", err)
				}
				cfg.Security.CertificateAuthority = ca
				cfg.Security.CertificateChain = cert
				cfg.Security.PrivateKey = key
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t); env.requireVM(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             testAccCheckKafkaInstanceDestroy,
				Steps: []resource.TestStep{
					{
						Config: renderKafkaInstanceConfig(env, cfg),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckKafkaInstanceExists("automq_kafka_instance.test"),
							resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.security.authentication_methods.#", fmt.Sprintf("%d", len(combo.security.AuthenticationMethods))),
							resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.security.transit_encryption_modes.#", fmt.Sprintf("%d", len(combo.security.TransitEncryptionModes))),
							resource.TestCheckResourceAttr("automq_kafka_instance.test", "features.security.data_encryption_mode", combo.security.DataEncryptionMode),
						),
					},
				},
			})
		})
	}
}

func needsTLSCertificates(transit []string) bool {
	for _, v := range transit {
		if strings.EqualFold(v, "tls") {
			return true
		}
	}
	return false
}

func getRequiredEnvVars(t *testing.T) map[string]string {
	env := loadAccConfig(t)
	vars := map[string]string{
		"AUTOMQ_BYOC_ENDPOINT":      env.Endpoint,
		"AUTOMQ_BYOC_ACCESS_KEY_ID": env.AccessKeyID,
		"AUTOMQ_BYOC_SECRET_KEY":    env.SecretKey,
		"AUTOMQ_TEST_ENV_ID":        env.EnvironmentID,
		"AUTOMQ_TEST_ZONE":          env.VMZone,
	}
	if len(env.VMSubnets) > 0 {
		vars["AUTOMQ_TEST_SUBNET_ID"] = env.VMSubnets[0]
	}
	if v := strings.TrimSpace(os.Getenv(envDeployProfileLegacy)); v != "" {
		vars["AUTOMQ_TEST_DEPLOY_PROFILE"] = v
	}
	return vars
}

func renderKafkaInstanceConfig(env accConfig, cfg accInstanceConfig) string {
	if cfg.Version == "" {
		cfg.Version = defaultInstanceVersion
	}
	if cfg.WalMode == "" {
		cfg.WalMode = "EBSWAL"
	}
	if cfg.DeployType == "" {
		cfg.DeployType = "IAAS"
	}

	var b strings.Builder

	fmt.Fprintf(&b, `provider "automq" {
  automq_byoc_endpoint      = %q
  automq_byoc_access_key_id = %q
  automq_byoc_secret_key    = %q
}

`, env.Endpoint, env.AccessKeyID, env.SecretKey)

	fmt.Fprintf(&b, `resource "automq_kafka_instance" "test" {
  environment_id = %q
  name           = %q
`, cfg.EnvironmentID, cfg.Name)
	if cfg.Description != "" {
		fmt.Fprintf(&b, "  description    = %q\n", cfg.Description)
	}
	fmt.Fprintf(&b, "  version        = %q\n", cfg.Version)

	b.WriteString("  compute_specs = {\n")
	fmt.Fprintf(&b, "    reserved_aku = %d\n", cfg.ReservedAKU)
	if cfg.DeployType != "" {
		fmt.Fprintf(&b, "    deploy_type  = %q\n", cfg.DeployType)
	}
	if len(cfg.Networks) > 0 {
		b.WriteString("    networks = [\n")
		for _, n := range cfg.Networks {
			b.WriteString("      {\n")
			if n.Zone != "" {
				fmt.Fprintf(&b, "        zone = %q\n", n.Zone)
			}
			if len(n.Subnets) > 0 {
				b.WriteString("        subnets = [")
				for i, s := range n.Subnets {
					if i > 0 {
						b.WriteString(", ")
					}
					fmt.Fprintf(&b, "%q", s)
				}
				b.WriteString("]\n")
			}
			b.WriteString("      }\n")
		}
		b.WriteString("    ]\n")
	}
	if cfg.KubernetesClusterID != "" {
		fmt.Fprintf(&b, "    kubernetes_cluster_id = %q\n", cfg.KubernetesClusterID)
	}
	if len(cfg.KubernetesNodeGroups) > 0 {
		b.WriteString("    kubernetes_node_groups = [\n")
		for _, ng := range cfg.KubernetesNodeGroups {
			fmt.Fprintf(&b, "      { id = %q }\n", ng)
		}
		b.WriteString("    ]\n")
	}
	if cfg.KubernetesNamespace != "" {
		fmt.Fprintf(&b, "    kubernetes_namespace = %q\n", cfg.KubernetesNamespace)
	}
	if cfg.KubernetesServiceAcct != "" {
		fmt.Fprintf(&b, "    kubernetes_service_account = %q\n", cfg.KubernetesServiceAcct)
	}
	if cfg.InstanceRole != "" {
		fmt.Fprintf(&b, "    instance_role = %q\n", cfg.InstanceRole)
	}
	b.WriteString("  }\n\n")

	b.WriteString("  features = {\n")
	fmt.Fprintf(&b, "    wal_mode = %q\n", cfg.WalMode)

	if len(cfg.InstanceConfigs) > 0 {
		b.WriteString("    instance_configs = {\n")
		for k, v := range cfg.InstanceConfigs {
			fmt.Fprintf(&b, "      %q = %q\n", k, v)
		}
		b.WriteString("    }\n")
	}

	if cfg.MetricsExporter != nil {
		b.WriteString("    metrics_exporter = {\n")
		b.WriteString("      prometheus = {\n")
		fmt.Fprintf(&b, "        enabled = %t\n", cfg.MetricsExporter.Enabled)
		if cfg.MetricsExporter.AuthType != "" {
			fmt.Fprintf(&b, "        auth_type = %q\n", cfg.MetricsExporter.AuthType)
		}
		if cfg.MetricsExporter.EndPoint != "" {
			fmt.Fprintf(&b, "        end_point = %q\n", cfg.MetricsExporter.EndPoint)
		}
		if cfg.MetricsExporter.PrometheusARN != "" {
			fmt.Fprintf(&b, "        prometheus_arn = %q\n", cfg.MetricsExporter.PrometheusARN)
		}
		if cfg.MetricsExporter.Username != "" {
			fmt.Fprintf(&b, "        username = %q\n", cfg.MetricsExporter.Username)
		}
		if cfg.MetricsExporter.Password != "" {
			fmt.Fprintf(&b, "        password = %q\n", cfg.MetricsExporter.Password)
		}
		if cfg.MetricsExporter.Token != "" {
			fmt.Fprintf(&b, "        token = %q\n", cfg.MetricsExporter.Token)
		}
		if len(cfg.MetricsExporter.Labels) > 0 {
			b.WriteString("        labels = {\n")
			for k, v := range cfg.MetricsExporter.Labels {
				fmt.Fprintf(&b, "          %q = %q\n", k, v)
			}
			b.WriteString("        }\n")
		}
		b.WriteString("      }\n")
		b.WriteString("    }\n")
	}

	if cfg.Security != nil {
		b.WriteString("    security = {\n")
		if len(cfg.Security.AuthenticationMethods) > 0 {
			fmt.Fprintf(&b, "      authentication_methods = [%s]\n", quoteJoin(cfg.Security.AuthenticationMethods))
		}
		if len(cfg.Security.TransitEncryptionModes) > 0 {
			fmt.Fprintf(&b, "      transit_encryption_modes = [%s]\n", quoteJoin(cfg.Security.TransitEncryptionModes))
		}
		if cfg.Security.DataEncryptionMode != "" {
			fmt.Fprintf(&b, "      data_encryption_mode = %q\n", cfg.Security.DataEncryptionMode)
		}
		if cfg.Security.CertificateAuthority != "" {
			fmt.Fprintf(&b, "      certificate_authority = <<-EOT\n%s\n      EOT\n", cfg.Security.CertificateAuthority)
		}
		if cfg.Security.CertificateChain != "" {
			fmt.Fprintf(&b, "      certificate_chain = <<-EOT\n%s\n      EOT\n", cfg.Security.CertificateChain)
		}
		if cfg.Security.PrivateKey != "" {
			fmt.Fprintf(&b, "      private_key = <<-EOT\n%s\n      EOT\n", cfg.Security.PrivateKey)
		}
		b.WriteString("    }\n")
	}

	if cfg.TableTopic != nil {
		b.WriteString("    table_topic = {\n")
		fmt.Fprintf(&b, "      warehouse = %q\n", cfg.TableTopic.Warehouse)
		fmt.Fprintf(&b, "      catalog_type = %q\n", cfg.TableTopic.CatalogType)
		if cfg.TableTopic.MetastoreURI != "" {
			fmt.Fprintf(&b, "      metastore_uri = %q\n", cfg.TableTopic.MetastoreURI)
		}
		b.WriteString("    }\n")
	}

	b.WriteString("  }\n")
	b.WriteString("}\n")

	return b.String()
}

func quoteJoin(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, v := range items {
		quoted = append(quoted, fmt.Sprintf("%q", v))
	}
	return strings.Join(quoted, ", ")
}

func ensureAccTimeout(t *testing.T) {
	if os.Getenv("TF_ACC_TIMEOUT") == "" {
		t.Setenv("TF_ACC_TIMEOUT", "2h")
	}
}

func testAccCheckKafkaInstanceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Kafka instance ID is set")
		}
		return nil
	}
}

func testAccCheckKafkaInstanceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "automq_kafka_instance" && rs.Primary.ID != "" {
			return fmt.Errorf("kafka instance %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func generateRandomSuffix() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "rand"
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func generateTestCertificate() (string, string, string, error) {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}

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

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}

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

	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})

	return string(caPEM), string(serverCertPEM), string(serverKeyPEM), nil
}
