package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

const (
	environmentCollectionPath = "/api/v1/environments"
	environmentItemPath       = "/api/v1/environments/%s"
)

func (c *Client) CreateEnvironment(ctx context.Context, param EnvironmentCreateParam) (*EnvironmentVO, error) {
	body, err := c.Post(ctx, environmentCollectionPath, param)
	if err != nil {
		return nil, err
	}
	environment := EnvironmentVO{}
	if err := json.Unmarshal(body, &environment); err != nil {
		return nil, err
	}
	return &environment, nil
}

func (c *Client) GetEnvironment(ctx context.Context, environmentID string) (*EnvironmentVO, error) {
	path := fmt.Sprintf(environmentItemPath, url.PathEscape(environmentID))
	body, err := c.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	environment := EnvironmentVO{}
	if err := json.Unmarshal(body, &environment); err != nil {
		return nil, err
	}
	return &environment, nil
}

func (c *Client) UpdateEnvironment(ctx context.Context, environmentID string, param EnvironmentUpdateParam) error {
	path := fmt.Sprintf(environmentItemPath, url.PathEscape(environmentID))
	_, err := c.Patch(ctx, path, param)
	return err
}

func (c *Client) DeleteEnvironment(ctx context.Context, environmentID string) error {
	path := fmt.Sprintf(environmentItemPath, url.PathEscape(environmentID))
	_, err := c.Delete(ctx, path)
	return err
}
