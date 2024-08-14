package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	IntegrationPath              = "/api/v1/integrations"
	GetIntegrationPath           = "/api/v1/integrations/%s"
	PatchIntegrationPath         = "/api/v1/integrations/%s"
	ListInstanceIntegrationsPath = "/api/v1/instances/%s/integrations"
)

func (c *Client) CreateIntergration(ctx context.Context, param IntegrationParam) (*IntegrationVO, error) {
	body, err := c.Post(ctx, IntegrationPath, param)
	if err != nil {
		return nil, err
	}
	integration := IntegrationVO{}
	err = json.Unmarshal(body, &integration)
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

func (c *Client) GetIntergration(ctx context.Context, integrationId string) (*IntegrationVO, error) {
	body, err := c.Get(ctx, fmt.Sprintf(GetIntegrationPath, integrationId), nil)
	if err != nil {
		return nil, err
	}
	integration := IntegrationVO{}
	err = json.Unmarshal(body, &integration)
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

func (c *Client) UpdateIntergration(ctx context.Context, integrationId string, param *IntegrationParam) (*IntegrationVO, error) {
	// make sure the type is not updated
	param.Type = nil

	body, err := c.Patch(ctx, fmt.Sprintf(PatchIntegrationPath, integrationId), param)
	if err != nil {
		return nil, err
	}
	integration := IntegrationVO{}
	err = json.Unmarshal(body, &integration)
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

func (c *Client) DeleteIntergration(ctx context.Context, integrationId string) error {
	_, err := c.Delete(ctx, fmt.Sprintf(PatchIntegrationPath, integrationId))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListInstanceIntegrations(ctx context.Context, instanceId string) ([]IntegrationVO, error) {
	body, err := c.Get(ctx, fmt.Sprintf(ListInstanceIntegrationsPath, instanceId), nil)
	if err != nil {
		return nil, err
	}
	var integrations PageNumResultIntegrationVO
	err = json.Unmarshal(body, &integrations)
	if err != nil {
		return nil, err
	}
	return integrations.List, nil
}
