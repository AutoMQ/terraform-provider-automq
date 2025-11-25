package provider

// Acceptance test configuration
//
// Scenarios & required inputs:
//   VM tests (e.g., TestAccKafkaInstance_VM_Scenario)
//     - endpoint, access_key_id, secret_key, environment_id
//     - vm.zone (or AUTMQ_VM_ZONE) and at least one subnet ID (AUTMQ_VM_SUBNET_IDS / AUTMQ_VM_SUBNET_ID / AUTMQ_TEST_SUBNET_ID)
//
//   K8S tests (TestAccKafkaInstance_K8S_Scenario)
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
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"terraform-provider-automq/client"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var accConfigPath = flag.String("acc.config", "", "path to JSON file containing acceptance test configuration")

type accConfig struct {
	Endpoint       string       `json:"endpoint"`
	AccessKeyID    string       `json:"access_key_id"`
	SecretKey      string       `json:"secret_key"`
	EnvironmentID  string       `json:"environment_id"`
	Region         string       `json:"region"`
	Networks       []accNetwork `json:"networks"`
	K8S            accK8SConfig `json:"k8s"`
	Version        string       `json:"version"`
	UpgradeVersion string       `json:"upgrade_version"`
}

type accK8SConfig struct {
	ClusterID      string   `json:"cluster_id"`
	NodeGroups     []string `json:"node_groups"`
	Namespace      string   `json:"namespace"`
	ServiceAccount string   `json:"service_account"`
}

var (
	configOnce sync.Once
	configData accConfig
	configErr  error
)

type accNetwork struct {
	Zone    string   `json:"zone"`
	Subnets []string `json:"subnets"`
}

type accTableTopic struct {
	Warehouse    string
	CatalogType  string
	MetastoreURI string
}

type accMetricsExporter struct {
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
		path := strings.TrimSpace(*accConfigPath)
		if path == "" {
			configErr = fmt.Errorf("-acc.config is required for acceptance tests")
			return
		}
		configData, configErr = parseAccConfigFromFile(path)
		if configErr == nil {
			if missing := configData.missingRequired(); len(missing) > 0 {
				configErr = fmt.Errorf("missing required configuration values in acc.config: %v", missing)
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
		Endpoint       string       `json:"endpoint"`
		AccessKeyID    string       `json:"access_key_id"`
		SecretKey      string       `json:"secret_key"`
		EnvironmentID  string       `json:"environment_id"`
		Region         string       `json:"region"`
		Version        string       `json:"version"`
		UpgradeVersion string       `json:"upgrade_version"`
		Networks       []accNetwork `json:"networks"`
		K8S            accK8SConfig `json:"k8s"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return accConfig{}, fmt.Errorf("parse acc.config JSON: %w", err)
	}
	cfg := accConfig{
		Endpoint:       strings.TrimSpace(raw.Endpoint),
		AccessKeyID:    strings.TrimSpace(raw.AccessKeyID),
		SecretKey:      strings.TrimSpace(raw.SecretKey),
		EnvironmentID:  strings.TrimSpace(raw.EnvironmentID),
		Region:         strings.TrimSpace(raw.Region),
		Version:        strings.TrimSpace(raw.Version),
		UpgradeVersion: strings.TrimSpace(raw.UpgradeVersion),
		Networks:       append([]accNetwork{}, raw.Networks...),
		K8S: accK8SConfig{
			ClusterID:      strings.TrimSpace(raw.K8S.ClusterID),
			NodeGroups:     append([]string{}, raw.K8S.NodeGroups...),
			Namespace:      strings.TrimSpace(raw.K8S.Namespace),
			ServiceAccount: strings.TrimSpace(raw.K8S.ServiceAccount),
		},
	}
	return cfg.normalise(), nil
}

func (c accConfig) normalise() accConfig {
	c.Endpoint = strings.TrimSpace(c.Endpoint)
	c.AccessKeyID = strings.TrimSpace(c.AccessKeyID)
	c.SecretKey = strings.TrimSpace(c.SecretKey)
	c.EnvironmentID = strings.TrimSpace(c.EnvironmentID)
	c.Region = strings.TrimSpace(c.Region)
	c.Version = strings.TrimSpace(c.Version)
	c.UpgradeVersion = strings.TrimSpace(c.UpgradeVersion)
	c.Networks = normaliseNetworks(c.Networks)
	c.K8S.ClusterID = strings.TrimSpace(c.K8S.ClusterID)
	c.K8S.NodeGroups = trimStringSlice(c.K8S.NodeGroups)
	c.K8S.Namespace = strings.TrimSpace(c.K8S.Namespace)
	c.K8S.ServiceAccount = strings.TrimSpace(c.K8S.ServiceAccount)
	return c
}

func (c accConfig) missingRequired() []string {
	missing := make([]string, 0, 5)
	if c.Endpoint == "" {
		missing = append(missing, "endpoint")
	}
	if c.AccessKeyID == "" {
		missing = append(missing, "access_key_id")
	}
	if c.SecretKey == "" {
		missing = append(missing, "secret_key")
	}
	if c.EnvironmentID == "" {
		missing = append(missing, "environment_id")
	}
	if len(c.Networks) == 0 {
		missing = append(missing, "networks")
	}
	if c.Version == "" {
		missing = append(missing, "version")
	}
	return missing
}

func (c accConfig) requireVM(t *testing.T) {
	if len(c.vmNetworks()) == 0 {
		t.Skip("Skipping VM acceptance tests: acc.config.networks must include at least one entry with zone and subnet")
	}
}

func (c accConfig) hasK8S() bool {
	return c.K8S.ClusterID != "" && len(c.K8S.NodeGroups) > 0
}

func TestAccKafkaInstance_VM_Scenario(t *testing.T) {
	env := loadAccConfig(t)
	env.requireVM(t)
	ensureAccTimeout(t)
	scenario := buildVMInstanceScenario(env)
	runAccKafkaInstanceScenario(t, env, scenario)
}

func TestAccKafkaInstance_K8S_Scenario(t *testing.T) {
	env := loadAccConfig(t)
	if !env.hasK8S() {
		t.Skip("Skipping K8S acceptance tests: acc.config.k8s.cluster_id/node_groups not configured")
	}
	env.requireVM(t)
	ensureAccTimeout(t)
	scenario := buildK8SInstanceScenario(env)
	runAccKafkaInstanceScenario(t, env, scenario)
}

// accInstanceScenario describes a linear acceptance flow (create, update, import)
// so each resource test stays easy to follow.
type accInstanceScenario struct {
	Name        string
	RequiresVM  bool
	RequiresK8S bool
	Steps       []accInstanceScenarioStep
}

// accInstanceScenarioStep contains the config and checks for a single TF step.
type accInstanceScenarioStep struct {
	Config accInstanceConfig
	Checks []resource.TestCheckFunc
}

// runAccKafkaInstanceScenario renders each step's HCL, runs the checks, and
// verifies import. It handles shared pre-check logic (VM/K8S requirements).
func runAccKafkaInstanceScenario(t *testing.T, env accConfig, scenario accInstanceScenario) {
	t.Helper()
	t.Logf("starting %s: env=%s endpoint=%s requires_vm=%t requires_k8s=%t networks=%d k8s_node_groups=%d", scenario.Name, env.EnvironmentID, env.Endpoint, scenario.RequiresVM, scenario.RequiresK8S, len(env.Networks), len(env.K8S.NodeGroups))
	preCheck := func() {
		testAccPreCheck(t)
		if scenario.RequiresVM {
			env.requireVM(t)
		}
		if scenario.RequiresK8S && !env.hasK8S() {
			t.Skip("Skipping K8S acceptance tests: acc.config.k8s.cluster_id/node_groups not configured")
		}
	}

	steps := make([]resource.TestStep, 0, len(scenario.Steps)+1)
	totalSteps := len(scenario.Steps)
	for idx, step := range scenario.Steps {
		metricsEnabled := step.Config.MetricsExporter != nil
		stepDeployType := step.Config.DeployType
		if stepDeployType == "" {
			stepDeployType = "IAAS"
		}
		t.Logf("scenario %s step %d/%d: deploy_type=%s reserved_aku=%d networks=%d k8s_cluster=%s node_groups=%d metrics=%t", scenario.Name, idx+1, totalSteps, stepDeployType, step.Config.ReservedAKU, len(step.Config.Networks), step.Config.KubernetesClusterID, len(step.Config.KubernetesNodeGroups), metricsEnabled)
		checks := append([]resource.TestCheckFunc{testAccCheckKafkaInstanceExists()}, step.Checks...)
		steps = append(steps, resource.TestStep{
			Config: renderKafkaInstanceConfig(env, step.Config),
			Check:  resource.ComposeAggregateTestCheckFunc(checks...),
		})
	}

	t.Logf("scenario %s import verification", scenario.Name)
	steps = append(steps, resource.TestStep{
		ResourceName:      kafkaInstanceResourceName,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateIdFunc: func(state *terraform.State) (string, error) {
			rs, ok := state.RootModule().Resources[kafkaInstanceResourceName]
			if !ok {
				return "", fmt.Errorf("resource %s not found in state", kafkaInstanceResourceName)
			}
			if rs.Primary == nil || rs.Primary.ID == "" {
				return "", fmt.Errorf("resource %s has no ID", kafkaInstanceResourceName)
			}
			if env.EnvironmentID == "" {
				return "", fmt.Errorf("environment_id is not configured for import")
			}
			return fmt.Sprintf("%s@%s", env.EnvironmentID, rs.Primary.ID), nil
		},
		ImportStateVerifyIgnore: []string{
			"features.metrics_exporter.prometheus.password",
			"features.metrics_exporter.prometheus.token",
			"features.security.private_key",
			"features.instance_configs",
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 preCheck,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKafkaInstanceDestroy,
		Steps:                    steps,
	})
}

// buildVMInstanceScenario returns the canonical VM acceptance flow: creation,
// basic updates, optional version upgrade, AKU change, metrics on/off.
func buildVMInstanceScenario(env accConfig) accInstanceScenario {
	vmNetworks := env.vmNetworks()
	if len(vmNetworks) == 0 {
		panic("vmNetworks called without networks; requireVM should have skipped earlier")
	}
	suffix := generateRandomSuffix()
	base := accInstanceConfig{
		EnvironmentID: env.EnvironmentID,
		Name:          fmt.Sprintf("acc-vm-%s", suffix),
		Description:   "VM acceptance baseline",
		Version:       env.Version,
		DeployType:    "IAAS",
		ReservedAKU:   6,
		Networks:      cloneNetworks(vmNetworks),
		WalMode:       "EBSWAL",
		Security: &accSecurity{
			AuthenticationMethods:  []string{"sasl"},
			TransitEncryptionModes: []string{"plaintext"},
			DataEncryptionMode:     "NONE",
		},
	}

	steps := []accInstanceScenarioStep{
		{
			Config: base,
			Checks: []resource.TestCheckFunc{
				checkAttr("name", base.Name),
				checkAttr("description", base.Description),
				checkAttr("compute_specs.reserved_aku", fmt.Sprintf("%d", base.ReservedAKU)),
				checkAttr("features.wal_mode", base.WalMode),
			},
		},
	}

	current := cloneInstanceConfig(base)
	basicUpdate := cloneInstanceConfig(current)
	basicUpdate.Description = "VM acceptance updated"
	steps = append(steps, accInstanceScenarioStep{
		Config: basicUpdate,
		Checks: []resource.TestCheckFunc{checkAttr("description", basicUpdate.Description)},
	})
	current = basicUpdate

	if upgrade := strings.TrimSpace(env.UpgradeVersion); upgrade != "" && upgrade != current.Version {
		versionStep := cloneInstanceConfig(current)
		versionStep.Version = upgrade
		steps = append(steps, accInstanceScenarioStep{
			Config: versionStep,
			Checks: []resource.TestCheckFunc{checkAttr("version", versionStep.Version)},
		})
		current = versionStep
	}

	akuStep := cloneInstanceConfig(current)
	akuStep.ReservedAKU = current.ReservedAKU + 2
	steps = append(steps, accInstanceScenarioStep{
		Config: akuStep,
		Checks: []resource.TestCheckFunc{checkAttr("compute_specs.reserved_aku", fmt.Sprintf("%d", akuStep.ReservedAKU))},
	})
	current = akuStep

	metricsOn := cloneInstanceConfig(current)
	metricsOn.MetricsExporter = &accMetricsExporter{
		AuthType: "noauth",
		EndPoint: fmt.Sprintf("https://metrics-%s.example.com/write", suffix),
		Labels:   map[string]string{"env": "acc", "role": "vm"},
	}
	steps = append(steps, accInstanceScenarioStep{
		Config: metricsOn,
		Checks: []resource.TestCheckFunc{checkAttr("features.metrics_exporter.prometheus.endpoint", metricsOn.MetricsExporter.EndPoint)},
	})
	current = metricsOn

	metricsOff := cloneInstanceConfig(current)
	metricsOff.MetricsExporter = nil
	steps = append(steps, accInstanceScenarioStep{
		Config: metricsOff,
		Checks: []resource.TestCheckFunc{checkNoAttr("features.metrics_exporter.prometheus")},
	})

	return accInstanceScenario{
		Name:       "VM scenario",
		RequiresVM: true,
		Steps:      steps,
	}
}

// buildK8SInstanceScenario mirrors the VM flow but exercises K8S-specific
// fields (cluster ID, node groups, etc.).
func buildK8SInstanceScenario(env accConfig) accInstanceScenario {
	networks := env.k8sNetworks()
	if len(networks) == 0 {
		networks = env.vmNetworks()
	}
	base := accInstanceConfig{
		EnvironmentID:         env.EnvironmentID,
		Name:                  fmt.Sprintf("acc-k8s-%s", generateRandomSuffix()),
		Description:           "K8S acceptance baseline",
		Version:               env.Version,
		DeployType:            "K8S",
		ReservedAKU:           6,
		Networks:              cloneNetworks(networks),
		KubernetesClusterID:   env.K8S.ClusterID,
		KubernetesNodeGroups:  cloneStringSlice(env.K8S.NodeGroups),
		KubernetesNamespace:   env.K8S.Namespace,
		KubernetesServiceAcct: env.K8S.ServiceAccount,
		WalMode:               "EBSWAL",
		Security: &accSecurity{
			AuthenticationMethods:  []string{"sasl"},
			TransitEncryptionModes: []string{"plaintext"},
			DataEncryptionMode:     "NONE",
		},
	}

	steps := []accInstanceScenarioStep{
		{
			Config: base,
			Checks: []resource.TestCheckFunc{
				checkAttr("compute_specs.deploy_type", "K8S"),
				checkAttr("compute_specs.kubernetes_cluster_id", env.K8S.ClusterID),
				checkAttr("compute_specs.kubernetes_node_groups.#", fmt.Sprintf("%d", len(env.K8S.NodeGroups))),
			},
		},
	}

	current := cloneInstanceConfig(base)
	descUpdate := cloneInstanceConfig(current)
	descUpdate.Description = "K8S acceptance updated"
	descUpdate.InstanceConfigs = map[string]string{"auto.create.topics.enable": "false"}
	steps = append(steps, accInstanceScenarioStep{
		Config: descUpdate,
		Checks: []resource.TestCheckFunc{
			checkAttr("description", descUpdate.Description),
			checkAttr("features.instance_configs.auto.create.topics.enable", "false"),
		},
	})
	current = descUpdate

	if upgrade := strings.TrimSpace(env.UpgradeVersion); upgrade != "" && upgrade != current.Version {
		versionStep := cloneInstanceConfig(current)
		versionStep.Version = upgrade
		steps = append(steps, accInstanceScenarioStep{
			Config: versionStep,
			Checks: []resource.TestCheckFunc{checkAttr("version", versionStep.Version)},
		})
		current = versionStep
	}

	akuStep := cloneInstanceConfig(current)
	akuStep.ReservedAKU = current.ReservedAKU + 4
	steps = append(steps, accInstanceScenarioStep{
		Config: akuStep,
		Checks: []resource.TestCheckFunc{checkAttr("compute_specs.reserved_aku", fmt.Sprintf("%d", akuStep.ReservedAKU))},
	})
	current = akuStep

	metricsOn := cloneInstanceConfig(current)
	metricsOn.MetricsExporter = &accMetricsExporter{
		AuthType: "noauth",
		EndPoint: fmt.Sprintf("https://metrics-k8s-%s.example.com/write", generateRandomSuffix()),
		Labels:   map[string]string{"env": "acc", "mode": "k8s"},
	}
	steps = append(steps, accInstanceScenarioStep{
		Config: metricsOn,
		Checks: []resource.TestCheckFunc{checkAttr("features.metrics_exporter.prometheus.endpoint", metricsOn.MetricsExporter.EndPoint)},
	})
	current = metricsOn

	metricsOff := cloneInstanceConfig(current)
	metricsOff.MetricsExporter = nil
	steps = append(steps, accInstanceScenarioStep{
		Config: metricsOff,
		Checks: []resource.TestCheckFunc{checkNoAttr("features.metrics_exporter.prometheus")},
	})

	return accInstanceScenario{
		Name:        "K8S scenario",
		RequiresVM:  true,
		RequiresK8S: true,
		Steps:       steps,
	}
}

// newVMInstanceConfig creates the minimal IAAS instance config shared by
// topic/user/ACL acceptance tests.
func newVMInstanceConfig(env accConfig, name, description string) accInstanceConfig {
	net, ok := env.firstVMNetwork()
	if !ok {
		panic("newVMInstanceConfig called without VM networks; requireVM should have skipped earlier")
	}
	return accInstanceConfig{
		EnvironmentID: env.EnvironmentID,
		Name:          name,
		Description:   description,
		Version:       env.Version,
		DeployType:    "IAAS",
		ReservedAKU:   6,
		Networks:      []accNetwork{{Zone: net.Zone, Subnets: cloneStringSlice(net.Subnets)}},
		WalMode:       "EBSWAL",
		Security: &accSecurity{
			AuthenticationMethods:  []string{"sasl"},
			TransitEncryptionModes: []string{"plaintext"},
			DataEncryptionMode:     "NONE",
		},
	}
}

// cloneInstanceConfig deep-copies the scenario config so later steps don't
// mutate earlier ones while they still need to assert state.
func cloneInstanceConfig(src accInstanceConfig) accInstanceConfig {
	clone := src
	clone.Networks = cloneNetworks(src.Networks)
	clone.KubernetesNodeGroups = cloneStringSlice(src.KubernetesNodeGroups)
	clone.InstanceConfigs = cloneStringMap(src.InstanceConfigs)
	if src.MetricsExporter != nil {
		clone.MetricsExporter = cloneMetricsExporter(src.MetricsExporter)
	}
	if src.Security != nil {
		clone.Security = cloneSecurity(src.Security)
	}
	if src.TableTopic != nil {
		clone.TableTopic = cloneTableTopic(src.TableTopic)
	}
	return clone
}

func cloneNetworks(in []accNetwork) []accNetwork {
	if len(in) == 0 {
		return nil
	}
	out := make([]accNetwork, len(in))
	for i, n := range in {
		out[i] = accNetwork{Zone: n.Zone, Subnets: cloneStringSlice(n.Subnets)}
	}
	return out
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneMetricsExporter(in *accMetricsExporter) *accMetricsExporter {
	if in == nil {
		return nil
	}
	clone := *in
	if len(in.Labels) > 0 {
		clone.Labels = cloneStringMap(in.Labels)
	}
	return &clone
}

func cloneSecurity(in *accSecurity) *accSecurity {
	if in == nil {
		return nil
	}
	clone := *in
	clone.AuthenticationMethods = cloneStringSlice(in.AuthenticationMethods)
	clone.TransitEncryptionModes = cloneStringSlice(in.TransitEncryptionModes)
	return &clone
}

func cloneTableTopic(in *accTableTopic) *accTableTopic {
	if in == nil {
		return nil
	}
	clone := *in
	return &clone
}

func normaliseNetworks(in []accNetwork) []accNetwork {
	if len(in) == 0 {
		return nil
	}
	out := make([]accNetwork, 0, len(in))
	for _, net := range in {
		zone := strings.TrimSpace(net.Zone)
		if zone == "" {
			continue
		}
		subnets := trimStringSlice(net.Subnets)
		out = append(out, accNetwork{Zone: zone, Subnets: subnets})
	}
	return out
}

func trimStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func (c accConfig) vmNetworks() []accNetwork {
	if len(c.Networks) == 0 {
		return nil
	}
	out := make([]accNetwork, 0, len(c.Networks))
	for _, net := range c.Networks {
		if net.Zone == "" || len(net.Subnets) == 0 {
			continue
		}
		out = append(out, accNetwork{Zone: net.Zone, Subnets: cloneStringSlice(net.Subnets)})
	}
	return out
}

func (c accConfig) k8sNetworks() []accNetwork {
	if len(c.Networks) == 0 {
		return nil
	}
	out := make([]accNetwork, 0, len(c.Networks))
	for _, net := range c.Networks {
		if net.Zone == "" {
			continue
		}
		out = append(out, accNetwork{Zone: net.Zone, Subnets: cloneStringSlice(net.Subnets)})
	}
	return out
}

func (c accConfig) firstVMNetwork() (accNetwork, bool) {
	nets := c.vmNetworks()
	if len(nets) == 0 {
		return accNetwork{}, false
	}
	return nets[0], true
}

func checkAttr(path, value string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttr(kafkaInstanceResourceName, path, value)
}

func checkNoAttr(path string) resource.TestCheckFunc {
	return resource.TestCheckNoResourceAttr(kafkaInstanceResourceName, path)
}
func renderKafkaInstanceConfig(env accConfig, cfg accInstanceConfig) string {
	if cfg.Version == "" {
		cfg.Version = env.Version
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
		for i, n := range cfg.Networks {
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
			b.WriteString("      }")
			if i < len(cfg.Networks)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString("    ]\n")
	}
	if cfg.KubernetesClusterID != "" {
		fmt.Fprintf(&b, "    kubernetes_cluster_id = %q\n", cfg.KubernetesClusterID)
	}
	if len(cfg.KubernetesNodeGroups) > 0 {
		b.WriteString("    kubernetes_node_groups = [\n")
		for i, ng := range cfg.KubernetesNodeGroups {
			fmt.Fprintf(&b, "      { id = %q }", ng)
			if i < len(cfg.KubernetesNodeGroups)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
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
		if cfg.MetricsExporter.AuthType != "" {
			fmt.Fprintf(&b, "        auth_type = %q\n", cfg.MetricsExporter.AuthType)
		}
		if cfg.MetricsExporter.EndPoint != "" {
			fmt.Fprintf(&b, "        endpoint = %q\n", cfg.MetricsExporter.EndPoint)
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

const kafkaInstanceResourceName = "automq_kafka_instance.test"

func testAccCheckKafkaInstanceExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[kafkaInstanceResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", kafkaInstanceResourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Kafka instance ID is set")
		}
		return nil
	}
}

func testAccCheckKafkaInstanceDestroy(s *terraform.State) error {
	cfg := loadAccConfig(nil)
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretKey == "" {
		return fmt.Errorf("provider configuration not available for destroy check")
	}

	ctx := context.Background()
	credential := client.AuthCredentials{
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretKey,
	}
	apiClient, err := client.NewClient(ctx, cfg.Endpoint, credential)
	if err != nil {
		return fmt.Errorf("failed to create API client for destroy check: %w", err)
	}

	for name, rs := range s.RootModule().Resources {
		if rs.Type != "automq_kafka_instance" {
			continue
		}
		if rs.Primary == nil || rs.Primary.ID == "" {
			continue
		}

		instanceID := rs.Primary.ID
		envID := rs.Primary.Attributes["environment_id"]
		if envID == "" {
			envID = cfg.EnvironmentID
		}
		ctx := context.WithValue(context.Background(), client.EnvIdKey, envID)

		instance, err := apiClient.GetKafkaInstance(ctx, instanceID)
		if err != nil {
			if errResp, ok := err.(*client.ErrorResponse); ok && errResp.Code == 404 {
				continue
			}
			return fmt.Errorf("error checking if kafka instance %s was destroyed: %w", instanceID, err)
		}

		if instance != nil {
			return fmt.Errorf("kafka instance %s still exists (resource=%s)", instanceID, name)
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
