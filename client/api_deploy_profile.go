package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	GetDeployProfilePath = "/api/v1/profiles/%s"
)

// GetDeployProfile retrieves a deployment profile by name
func (c *Client) GetDeployProfile(ctx context.Context, name string) (*DeployProfileVO, error) {
	data, err := c.Get(ctx, fmt.Sprintf(GetDeployProfilePath, name), nil)
	if err != nil {
		return nil, err
	}

	var profile DeployProfileVO
	err = json.Unmarshal(data, &profile)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling deploy profile response: %v", err)
	}

	return &profile, nil
}
