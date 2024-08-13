package client

import "time"

// IntegrationVO struct for IntegrationVO
type IntegrationVO struct {
	GmtCreate   time.Time              `json:"gmtCreate,omitempty"`
	GmtModified time.Time              `json:"gmtModified,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Code        string                 `json:"code,omitempty"`
	Name        string                 `json:"name,omitempty"`
	EndPoint    *string                `json:"endPoint,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// IntegrationParam struct for IntegrationParam
type IntegrationParam struct {
	Type     *string           `json:"type,omitempty"`
	Name     string            `json:"name" validate:"regexp=^[a-zA-Z\\\\u4e00-\\\\u9fa5][a-z0-9A-Z\\\\u4e00-\\\\u9fa5_\\\\s-]*$"`
	EndPoint string            `json:"endPoint"`
	Config   []ConfigItemParam `json:"config,omitempty"`
}

// IntegrationInstanceParam struct for IntegrationInstanceParam
type IntegrationInstanceParam struct {
	Codes    []string           `json:"codes,omitempty"`
	NewItems []IntegrationParam `json:"newItems,omitempty"`
}

// IntegrationInstanceAddParam struct for IntegrationInstanceAddParam
type IntegrationInstanceAddParam struct {
	Codes []string `json:"codes"`
}

// IntegrationUpdateParam struct for IntegrationUpdateParam
type IntegrationUpdateParam struct {
	Name     string            `json:"name" validate:"regexp=^[a-zA-Z\\\\u4e00-\\\\u9fa5][a-z0-9A-Z\\\\u4e00-\\\\u9fa5_\\\\s-]*$"`
	EndPoint string            `json:"endPoint"`
	Config   []ConfigItemParam `json:"config,omitempty"`
}

// ConfigItemParam struct for ConfigItemParam
type ConfigItemParam struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// PageNumResultIntegrationVO struct for PageNumResultIntegrationVO
type PageNumResultIntegrationVO struct {
	List []IntegrationVO `json:"list,omitempty"`
}
