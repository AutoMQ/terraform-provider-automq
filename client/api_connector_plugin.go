package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	pluginCollectionPath = "/api/v1/connect-plugins"
	pluginItemPath       = "/api/v1/connect-plugins/%s"
)

func (c *Client) CreateConnectPlugin(ctx context.Context, param ConnectPluginCreateParam) (*ConnectPluginVO, error) {
	body, err := c.Post(ctx, pluginCollectionPath, param)
	if err != nil {
		return nil, err
	}
	result := ConnectPluginVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetConnectPlugin(ctx context.Context, pluginId string) (*ConnectPluginVO, error) {
	path := fmt.Sprintf(pluginItemPath, pluginId)
	body, err := c.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, &ErrorResponse{Code: 404, ErrorMessage: fmt.Sprintf("plugin %s returned empty response", pluginId)}
	}
	result := ConnectPluginVO{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteConnectPlugin(ctx context.Context, pluginId string) error {
	_, err := c.Delete(ctx, fmt.Sprintf(pluginItemPath, pluginId))
	return err
}
