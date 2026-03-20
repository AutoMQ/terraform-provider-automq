package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	connectorCollectionPath = "/api/v1/connectors"
	connectorItemPath       = "/api/v1/connectors/%s"
	connectorPausePath      = "/api/v1/connectors/%s:pause"
	connectorResumePath     = "/api/v1/connectors/%s:resume"
)

func (c *Client) CreateConnector(ctx context.Context, param ConnectorCreateParam) (*ConnectorVO, error) {
	body, err := c.Post(ctx, connectorCollectionPath, param)
	if err != nil {
		return nil, err
	}
	result := ConnectorVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetConnector(ctx context.Context, connectorId string) (*ConnectorVO, error) {
	path := fmt.Sprintf(connectorItemPath, connectorId)
	body, err := c.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	result := ConnectorVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateConnector(ctx context.Context, connectorId string, param ConnectorUpdateParam) (*ConnectorVO, error) {
	path := fmt.Sprintf(connectorItemPath, connectorId)
	body, err := c.Put(ctx, path, param)
	if err != nil {
		return nil, err
	}
	result := ConnectorVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteConnector(ctx context.Context, connectorId string) error {
	_, err := c.Delete(ctx, fmt.Sprintf(connectorItemPath, connectorId))
	return err
}

func (c *Client) PauseConnector(ctx context.Context, connectorId string) (*ConnectorVO, error) {
	body, err := c.Post(ctx, fmt.Sprintf(connectorPausePath, connectorId), struct{}{})
	if err != nil {
		return nil, err
	}
	result := ConnectorVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ResumeConnector(ctx context.Context, connectorId string) (*ConnectorVO, error) {
	body, err := c.Post(ctx, fmt.Sprintf(connectorResumePath, connectorId), struct{}{})
	if err != nil {
		return nil, err
	}
	result := ConnectorVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
