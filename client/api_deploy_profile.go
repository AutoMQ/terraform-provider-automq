package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	GetDeployProfilePath = "/api/v1/profiles/%s"
	GetBucketProfilePath = "/api/v1/deploy-profiles/%s/data-buckets"
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
		return nil, fmt.Errorf("get deploy profile error: %s \nmaybe the profile does not exist", string(data))
	}

	return &profile, nil
}

// GetBucketProfile retrieves a bucket profile by name
func (c *Client) GetBucketProfiles(ctx context.Context, name string) (*PageNumResultBucketProfileVO, error) {
	data, err := c.Get(ctx, fmt.Sprintf(GetBucketProfilePath, name), nil)
	if err != nil {
		return nil, err
	}

	var profiles PageNumResultBucketProfileVO
	err = json.Unmarshal(data, &profiles)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling bucket profile response: %v", err)
	}

	return &profiles, nil
}
