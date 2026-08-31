package client

import "time"

type EnvironmentCreateParam struct {
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	CloudProvider string  `json:"cloudProvider"`
	Region        string  `json:"region"`
	Scope         string  `json:"scope"`
}

type EnvironmentUpdateParam struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type EnvironmentVO struct {
	EnvID          string     `json:"envId"`
	CreatedAt      *time.Time `json:"gmtCreate,omitempty"`
	UpdatedAt      *time.Time `json:"gmtModified,omitempty"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	CloudProvider  string     `json:"cloudProvider"`
	OpsBucket      *string    `json:"opsBucket,omitempty"`
	Region         string     `json:"region"`
	Scope          string     `json:"scope"`
	OrganizationID *string    `json:"organizationId,omitempty"`
	ClientID       *string    `json:"clientId,omitempty"`
	ClientSecret   *string    `json:"clientSecret,omitempty"`
}
