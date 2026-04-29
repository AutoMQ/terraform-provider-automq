package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	connectClusterCollectionPath = "/api/v1/connect/clusters"
	connectClusterItemPath       = "/api/v1/connect/clusters/%s"
)

func (c *Client) CreateConnectCluster(ctx context.Context, param ConnectClusterCreateParam) (*ConnectClusterVO, error) {
	body, err := c.Post(ctx, connectClusterCollectionPath, param)
	if err != nil {
		return nil, err
	}
	result := ConnectClusterVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetConnectCluster(ctx context.Context, clusterId string) (*ConnectClusterVO, error) {
	body, err := c.Get(ctx, fmt.Sprintf(connectClusterItemPath, clusterId), nil)
	if err != nil {
		return nil, err
	}
	result := ConnectClusterVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateConnectCluster(ctx context.Context, clusterId string, param ConnectClusterUpdateParam) (*ConnectClusterVO, error) {
	body, err := c.Put(ctx, fmt.Sprintf(connectClusterItemPath, clusterId), param)
	if err != nil {
		return nil, err
	}
	result := ConnectClusterVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteConnectCluster(ctx context.Context, clusterId string) error {
	_, err := c.Delete(ctx, fmt.Sprintf(connectClusterItemPath, clusterId))
	return err
}
