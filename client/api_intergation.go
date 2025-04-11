package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	IntegrationPath               = "/api/v1/integrations"
	GetIntegrationPath            = "/api/v1/integrations/%s"
	PatchIntegrationPath          = "/api/v1/integrations/%s"
	InstanceIntegrationPath       = "/api/v1/instances/%s/integrations"
	RemoveInstanceIntergationPath = "/api/v1/instances/%s/integrations/%s"
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

func (c *Client) AddInstanceIntergation(ctx context.Context, instanceId string, param *IntegrationInstanceAddParam) error {
	_, err := c.Patch(ctx, InstanceIntegrationPath, param)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveInstanceIntergation(ctx context.Context, instanceId string, integrationId string) error {
	_, err := c.Delete(ctx, fmt.Sprintf(RemoveInstanceIntergationPath, instanceId, integrationId))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteIntergration(ctx context.Context, integrationId string) error {
	_, err := c.Delete(ctx, fmt.Sprintf(PatchIntegrationPath, integrationId))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListInstanceIntegrations(ctx context.Context, instanceId string) ([]IntegrationVO, error) {
	body, err := c.Get(ctx, fmt.Sprintf(InstanceIntegrationPath, instanceId), nil)
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
