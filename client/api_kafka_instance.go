package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	InstancePath             = "/api/v1/instances"
	InstanceConfigPath       = "/api/v1/instances/%s/configurations"
	GetInstancePath          = "/api/v1/instances/%s"
	DeleteInstancePath       = "/api/v1/instances/%s"
	GetInstanceEndpointsPath = "/api/v1/instances/%s/endpoints"
	UpdateInstancePath       = "/api/v1/instances/%s"
)

func (c *Client) CreateKafkaInstance(ctx context.Context, kafka InstanceCreateParam) (*InstanceSummaryVO, error) {
	body, err := c.Post(ctx, InstancePath, kafka)
	if err != nil {
		return nil, err
	}
	newkafka := InstanceSummaryVO{}
	err = json.Unmarshal(body, &newkafka)
	if err != nil {
		return nil, err
	}
	return &newkafka, nil
}

func (c *Client) GetKafkaInstance(ctx context.Context, instanceId string) (*InstanceVO, error) {
	body, err := c.Get(ctx, fmt.Sprintf(GetInstancePath, instanceId), nil)
	if err != nil {
		return nil, err
	}
	instance := InstanceVO{}
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) GetKafkaInstanceByName(ctx context.Context, name string) (*InstanceVO, error) {
	queryParams := make(map[string]string)
	queryParams["keyword"] = name
	body, err := c.Get(ctx, InstancePath, queryParams)
	if err != nil {
		return nil, err
	}
	instances := PageNumResultInstanceVO{}
	err = json.Unmarshal(body, &instances)
	if err != nil {
		return nil, err
	}
	if len(instances.List) > 0 {
		for _, item := range instances.List {
			if *item.Name == name {
				return &item, nil
			}
		}
		return nil, &ErrorResponse{Code: 404, ErrorMessage: "kafka instance not found"}
	}
	return nil, &ErrorResponse{Code: 404, ErrorMessage: "kafka instance not found"}
}

func (c *Client) DeleteKafkaInstance(ctx context.Context, instanceId string) error {
	_, err := c.Delete(ctx, fmt.Sprintf(DeleteInstancePath, instanceId))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetInstanceEndpoints(ctx context.Context, instanceId string) ([]InstanceAccessInfoVO, error) {
	path := fmt.Sprintf(GetInstanceEndpointsPath, instanceId)
	body, err := c.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	endpoints := PageNumResultInstanceAccessInfoVO{}
	err = json.Unmarshal(body, &endpoints)
	if err != nil {
		return nil, err
	}
	return endpoints.List, nil
}

func (c *Client) GetInstanceConfigs(ctx context.Context, instanceId string) ([]ConfigItemParam, error) {
	path := fmt.Sprintf(InstanceConfigPath, instanceId)
	body, err := c.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	instance := PageNumResultConfigItemVO{}
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, err
	}
	return instance.List, nil
}

func (c *Client) UpdateKafkaInstance(ctx context.Context, instanceId string, updateParam InstanceUpdateParam) error {
	return c.updateInstance(ctx, instanceId, updateParam, UpdateInstancePath)
}

func (c *Client) updateInstance(ctx context.Context, instanceId string, updateParam interface{}, path string) error {
	_, err := c.Patch(ctx, fmt.Sprintf(path, instanceId), updateParam)
	if err != nil {
		return err
	}

	return nil
}
