package models

import (
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	IntegrationTypeCloudWatch            = "cloudWatch"
	IntegrationTypePrometheus            = "prometheus"
	IntegrationTypePrometheusRemoteWrite = "prometheusRemoteWrite"
)

// IntegrationResourceModel describes the resource data model.
type IntegrationResourceModel struct {
	EnvironmentID               types.String                            `tfsdk:"environment_id"`
	Name                        types.String                            `tfsdk:"name"`
	Type                        types.String                            `tfsdk:"type"`
	EndPoint                    types.String                            `tfsdk:"endpoint"`
	ID                          types.String                            `tfsdk:"id"`
	DeployProfile               types.String                            `tfsdk:"deploy_profile"`
	CloudWatchConfig            *CloudWatchIntegrationConfig            `tfsdk:"cloudwatch_config"`
	PrometheusRemoteWriteConfig *PrometheusRemoteWriteIntegrationConfig `tfsdk:"prometheus_remote_write_config"`
	CreatedAt                   timetypes.RFC3339                       `tfsdk:"created_at"`
	LastUpdated                 timetypes.RFC3339                       `tfsdk:"last_updated"`
}

type PrometheusRemoteWriteIntegrationConfig struct {
	AuthType    types.String `tfsdk:"auth_type"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	BearerToken types.String `tfsdk:"bearer_token"`
}

type CloudWatchIntegrationConfig struct {
	NameSpace types.String `tfsdk:"namespace"`
}

// IntegrationConfigHandler interface defines methods that each integration type must implement
type IntegrationConfigHandler interface {
	Validate(model *IntegrationResourceModel) diag.Diagnostic
	ExpandConfig(model *IntegrationResourceModel) []client.ConfigItemParam
	FlattenConfig(config map[string]interface{}, resource *IntegrationResourceModel)
}

// CloudWatchConfigHandler implements IntegrationConfigHandler
type CloudWatchConfigHandler struct{}

func (h *CloudWatchConfigHandler) Validate(model *IntegrationResourceModel) diag.Diagnostic {
	if model.CloudWatchConfig == nil {
		return diag.NewErrorDiagnostic("Missing required field", "cloudwatch_config is required for CloudWatch integration")
	}
	if model.CloudWatchConfig.NameSpace.ValueString() == "" {
		return diag.NewErrorDiagnostic("Missing required field", "namespace is required for CloudWatch integration")
	}
	return nil
}

func (h *CloudWatchConfigHandler) ExpandConfig(model *IntegrationResourceModel) []client.ConfigItemParam {
	return []client.ConfigItemParam{
		{
			Key:   "namespace",
			Value: model.CloudWatchConfig.NameSpace.ValueString(),
		},
	}
}

func (h *CloudWatchConfigHandler) FlattenConfig(config map[string]interface{}, resource *IntegrationResourceModel) {
	resource.CloudWatchConfig = &CloudWatchIntegrationConfig{}
	if v, ok := config["namespace"]; ok {
		if str, ok := v.(string); ok {
			resource.CloudWatchConfig.NameSpace = types.StringValue(str)
		}
	}
}

// PrometheusRemoteWriteConfigHandler implements IntegrationConfigHandler
type PrometheusRemoteWriteConfigHandler struct{}

func (h *PrometheusRemoteWriteConfigHandler) Validate(model *IntegrationResourceModel) diag.Diagnostic {
	if model.EndPoint.IsNull() || model.EndPoint.IsUnknown() {
		return diag.NewErrorDiagnostic("Missing required field", "endpoint is required for Prometheus Remote Write integration")
	}
	if model.PrometheusRemoteWriteConfig == nil {
		return diag.NewErrorDiagnostic("Missing required field", "prometheus_remote_write_config is required")
	}
	if model.PrometheusRemoteWriteConfig.AuthType.IsNull() || model.PrometheusRemoteWriteConfig.AuthType.IsUnknown() {
		return diag.NewErrorDiagnostic("Missing required field", "auth_type is required")
	}

	authType := model.PrometheusRemoteWriteConfig.AuthType.ValueString()
	switch authType {
	case "basic":
		if model.PrometheusRemoteWriteConfig.Username.IsNull() || model.PrometheusRemoteWriteConfig.Password.IsNull() {
			return diag.NewErrorDiagnostic("Missing required field", "username and password are required for basic auth")
		}
	case "bearer":
		if model.PrometheusRemoteWriteConfig.BearerToken.IsNull() {
			return diag.NewErrorDiagnostic("Missing required field", "bearer_token is required for bearer auth")
		}
	case "sigv4", "noauth":
		// No additional validation needed
	default:
		return diag.NewErrorDiagnostic("Invalid auth type", "auth_type must be one of: noauth, basic, bearer, sigv4")
	}
	return nil
}

func (h *PrometheusRemoteWriteConfigHandler) ExpandConfig(model *IntegrationResourceModel) []client.ConfigItemParam {
	config := []client.ConfigItemParam{{
		Key:   "authType",
		Value: model.PrometheusRemoteWriteConfig.AuthType.ValueString(),
	}}

	switch model.PrometheusRemoteWriteConfig.AuthType.ValueString() {
	case "basic":
		config = append(config,
			client.ConfigItemParam{Key: "username", Value: model.PrometheusRemoteWriteConfig.Username.ValueString()},
			client.ConfigItemParam{Key: "password", Value: model.PrometheusRemoteWriteConfig.Password.ValueString()},
		)
	case "bearer":
		config = append(config,
			client.ConfigItemParam{Key: "token", Value: model.PrometheusRemoteWriteConfig.BearerToken.ValueString()},
		)
	}
	return config
}

func (h *PrometheusRemoteWriteConfigHandler) FlattenConfig(config map[string]interface{}, resource *IntegrationResourceModel) {
	resource.PrometheusRemoteWriteConfig = &PrometheusRemoteWriteIntegrationConfig{}

	if authType, ok := config["authType"].(string); ok {
		resource.PrometheusRemoteWriteConfig.AuthType = types.StringValue(authType)

		switch authType {
		case "basic":
			if username, ok := config["username"].(string); ok {
				resource.PrometheusRemoteWriteConfig.Username = types.StringValue(username)
			}
			if password, ok := config["password"].(string); ok {
				resource.PrometheusRemoteWriteConfig.Password = types.StringValue(password)
			}
		case "bearer":
			if token, ok := config["token"].(string); ok {
				resource.PrometheusRemoteWriteConfig.BearerToken = types.StringValue(token)
			}
		}
	}
}

// getConfigHandler returns the appropriate config handler for the integration type
func getConfigHandler(integrationType string) IntegrationConfigHandler {
	switch integrationType {
	case IntegrationTypeCloudWatch:
		return &CloudWatchConfigHandler{}
	case IntegrationTypePrometheusRemoteWrite:
		return &PrometheusRemoteWriteConfigHandler{}
	default:
		return nil
	}
}

func ExpandIntergationResource(in *client.IntegrationParam, integration IntegrationResourceModel) diag.Diagnostic {
	integrationType := integration.Type.ValueString()
	in.Name = integration.Name.ValueString()
	in.Type = &integrationType
	in.EndPoint = integration.EndPoint.ValueString()
	in.Profile = integration.DeployProfile.ValueString()

	handler := getConfigHandler(integrationType)
	if handler == nil {
		return nil // 处理不需要特殊配置的类型
	}

	if diag := handler.Validate(&integration); diag != nil {
		return diag
	}

	in.Config = handler.ExpandConfig(&integration)
	return nil
}

func FlattenIntergrationResource(integration *client.IntegrationVO, resource *IntegrationResourceModel) {
	resource.ID = types.StringValue(integration.Code)
	resource.Name = types.StringValue(integration.Name)
	resource.Type = types.StringValue(integration.Type)
	resource.DeployProfile = types.StringValue(integration.Profile)
	resource.CreatedAt = timetypes.NewRFC3339TimePointerValue(&integration.GmtCreate)
	resource.LastUpdated = timetypes.NewRFC3339TimePointerValue(&integration.GmtModified)
	flattenIntergrationTypeConfig(integration.Type, integration.Config, resource)
}

func flattenIntergrationTypeConfig(iType string, config map[string]interface{}, resource *IntegrationResourceModel) {
	handler := getConfigHandler(iType)
	if handler != nil {
		handler.FlattenConfig(config, resource)
	}
}
