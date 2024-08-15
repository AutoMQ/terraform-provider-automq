package models

import (
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	IntegrationTypeCloudWatch = "cloudWatch"
	IntegrationTypeKafka      = "kafka"
	IntegrationTypePrometheus = "prometheus"

	SecurityProtocolSASLPlain = "SASL_PLAINTEXT"
	SecurityProtocolPlainText = "PLAINTEXT"
)

// IntegrationResourceModel describes the resource data model.
type IntegrationResourceModel struct {
	EnvironmentID    types.String                 `tfsdk:"environment_id"`
	Name             types.String                 `tfsdk:"name"`
	Type             types.String                 `tfsdk:"type"`
	EndPoint         types.String                 `tfsdk:"endpoint"`
	ID               types.String                 `tfsdk:"id"`
	KafkaConfig      *KafkaIntegrationConfig      `tfsdk:"kafka_config"`
	PrometheusConfig *PrometheusIntegrationConfig `tfsdk:"prometheus_config"`
	CloudWatchConfig *CloudWatchIntegrationConfig `tfsdk:"cloudwatch_config"`
	CreatedAt        timetypes.RFC3339            `tfsdk:"created_at"`
	LastUpdated      timetypes.RFC3339            `tfsdk:"last_updated"`
}

type KafkaIntegrationConfig struct {
	SecurityProtocol types.String `tfsdk:"security_protocol"`
	SaslMechanism    types.String `tfsdk:"sasl_mechanism"`
	SaslUsername     types.String `tfsdk:"sasl_username"`
	SaslPassword     types.String `tfsdk:"sasl_password"`
}

type PrometheusIntegrationConfig struct {
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	BearerToken types.String `tfsdk:"bearer_token"`
}

type CloudWatchIntegrationConfig struct {
	NameSpace types.String `tfsdk:"namespace"`
}

func ExpandIntergationResource(in *client.IntegrationParam, integration IntegrationResourceModel) diag.Diagnostic {
	integrationType := integration.Type.ValueString()
	in.Name = integration.Name.ValueString()
	in.Type = &integrationType
	if integrationType == IntegrationTypeCloudWatch {
		in.Name = integration.Name.ValueString()
		if integration.CloudWatchConfig == nil {
			return diag.NewErrorDiagnostic("Missing required field", "cloud_watch_config is required for CloudWatch integration")
		}
		if integration.CloudWatchConfig.NameSpace.ValueString() == "" {
			return diag.NewErrorDiagnostic("Missing required field", "namespace is required for CloudWatch integration")
		}
		in.Config = []client.ConfigItemParam{
			{
				Key:   "namespace",
				Value: integration.CloudWatchConfig.NameSpace.ValueString(),
			},
		}
	} else if integrationType == IntegrationTypeKafka {
		in.Name = integration.Name.ValueString()
		if integration.EndPoint.IsNull() || integration.EndPoint.IsUnknown() {
			return diag.NewErrorDiagnostic("Missing required field", "endpoint is required for Kafka integration")
		}
		in.EndPoint = integration.EndPoint.ValueString()
		if integration.KafkaConfig == nil {
			return diag.NewErrorDiagnostic("Missing required field", "kafka_config is required for Kafka integration")
		}
		if integration.KafkaConfig.SecurityProtocol.ValueString() == "" {
			return diag.NewErrorDiagnostic("Missing required field", "security_protocol is required for Kafka integration")
		}
		if integration.KafkaConfig.SecurityProtocol.ValueString() == SecurityProtocolSASLPlain {
			if integration.KafkaConfig.SaslMechanism.ValueString() == "" {
				return diag.NewErrorDiagnostic("Missing required field", "sasl_mechanism is required for Kafka integration")
			}
			if integration.KafkaConfig.SaslUsername.ValueString() == "" {
				return diag.NewErrorDiagnostic("Missing required field", "sasl_username is required for Kafka integration")
			}
			if integration.KafkaConfig.SaslPassword.ValueString() == "" {
				return diag.NewErrorDiagnostic("Missing required field", "sasl_password is required for Kafka integration")
			}
			in.Config = []client.ConfigItemParam{
				{
					Key:   "security_protocol",
					Value: integration.KafkaConfig.SecurityProtocol.ValueString(),
				},
				{
					Key:   "sasl_mechanism",
					Value: integration.KafkaConfig.SaslMechanism.ValueString(),
				},
				{
					Key:   "sasl_username",
					Value: integration.KafkaConfig.SaslUsername.ValueString(),
				},
				{
					Key:   "sasl_password",
					Value: integration.KafkaConfig.SaslPassword.ValueString(),
				},
			}
		} else if integration.KafkaConfig.SecurityProtocol.ValueString() == SecurityProtocolPlainText {
			in.Config = []client.ConfigItemParam{
				{
					Key:   "security_protocol",
					Value: integration.KafkaConfig.SecurityProtocol.ValueString(),
				},
			}
		}
	} else if integrationType == IntegrationTypePrometheus {
		in.Name = integration.Name.ValueString()
		if integration.EndPoint.IsNull() || integration.EndPoint.IsUnknown() {
			return diag.NewErrorDiagnostic("Missing required field", "endpoint is required for Prometheus integration")
		}
		in.EndPoint = integration.EndPoint.ValueString()
	}
	return nil
}

func FlattenIntergrationResource(integration *client.IntegrationVO, resource *IntegrationResourceModel) {
	resource.ID = types.StringValue(integration.Code)
	resource.Name = types.StringValue(integration.Name)
	resource.Type = types.StringValue(integration.Type)
	resource.CreatedAt = timetypes.NewRFC3339TimePointerValue(&integration.GmtCreate)
	resource.LastUpdated = timetypes.NewRFC3339TimePointerValue(&integration.GmtModified)
	flattenIntergrationTypeConfig(integration.Type, integration.Config, resource)
}

func flattenIntergrationTypeConfig(iType string, config map[string]interface{}, resource *IntegrationResourceModel) {
	if iType == IntegrationTypeKafka {
		flattenKafkaConfig(config, resource)
		return
	} else if iType == IntegrationTypeCloudWatch {
		flattenCloudWatchConfig(config, resource)
		return
	} else if iType == IntegrationTypePrometheus {
		return
	}
}

func flattenKafkaConfig(config map[string]interface{}, resource *IntegrationResourceModel) {
	resource.KafkaConfig = &KafkaIntegrationConfig{}
	if v, ok := config["securityProtocol"]; ok {
		resource.KafkaConfig.SecurityProtocol = types.StringValue(v.(string))
	}
	if v, ok := config["saslMechanism"]; ok {
		resource.KafkaConfig.SaslMechanism = types.StringValue(v.(string))
	}
	if v, ok := config["saslUsername"]; ok {
		resource.KafkaConfig.SaslUsername = types.StringValue(v.(string))
	}
	if v, ok := config["saslPassword"]; ok {
		resource.KafkaConfig.SaslPassword = types.StringValue(v.(string))
	}
}

func flattenCloudWatchConfig(config map[string]interface{}, resource *IntegrationResourceModel) {
	resource.CloudWatchConfig = &CloudWatchIntegrationConfig{}
	if v, ok := config["namespace"]; ok {
		resource.CloudWatchConfig.NameSpace = types.StringValue(v.(string))
	}
}
