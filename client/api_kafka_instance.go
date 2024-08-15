package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	InstancePath                   = "/api/v1/instances"
	InstanceConfigPath             = "/api/v1/instances/%s/configurations"
	GetInstancePath                = "/api/v1/instances/%s"
	DeleteInstancePath             = "/api/v1/instances/%s"
	ReplaceInstanceIntergationPath = "/api/v1/instances/%s/integrations"
	TurnOnInstanceAclPath          = "/api/v1/instances/%s/acls:enable"
	GetInstanceEndpointsPath       = "/api/v1/instances/%s/endpoints"
	UpdateInstanceBasicInfoPath    = "/api/v1/instances/%s/basic"
	UpdateInstanceVersionPath      = "/api/v1/instances/%s/versions/%s"
	UpdateInstanceComputeSpecsPath = "/api/v1/instances/%s/spec"
)

func (c *Client) CreateKafkaInstance(ctx context.Context, kafka KafkaInstanceRequest) (*KafkaInstanceResponse, error) {
	body, err := c.Post(ctx, InstancePath, kafka)
	if err != nil {
		return nil, err
	}
	newkafka := KafkaInstanceResponse{}
	err = json.Unmarshal(body, &newkafka)
	if err != nil {
		return nil, err
	}
	return &newkafka, nil
}

func (c *Client) GetKafkaInstance(ctx context.Context, instanceId string) (*KafkaInstanceResponse, error) {
	body, err := c.Get(ctx, fmt.Sprintf(GetInstancePath, instanceId), nil)
	if err != nil {
		return nil, err
	}
	instance := KafkaInstanceResponse{}
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) GetKafkaInstanceByName(ctx context.Context, name string) (*KafkaInstanceResponse, error) {
	queryParams := make(map[string]string)
	queryParams["keyword"] = name
	body, err := c.Get(ctx, InstancePath, queryParams)
	if err != nil {
		return nil, err
	}
	instances := KafkaInstanceResponseList{}
	err = json.Unmarshal(body, &instances)
	if err != nil {
		return nil, err
	}
	if len(instances.List) > 0 {
		for _, item := range instances.List {
			if item.DisplayName == name {
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

func (c *Client) ReplaceInstanceIntergation(ctx context.Context, instanceId string, param IntegrationInstanceParam) error {
	path := fmt.Sprintf(ReplaceInstanceIntergationPath, instanceId)
	_, err := c.Put(ctx, path, param)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) TurnOnInstanceAcl(ctx context.Context, instanceId string) error {
	path := fmt.Sprintf(TurnOnInstanceAclPath, instanceId)
	_, err := c.Post(ctx, path, nil)
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

func (c *Client) UpdateKafkaInstanceBasicInfo(ctx context.Context, instanceId string, updateParam InstanceBasicParam) (*KafkaInstanceResponse, error) {
	return c.updateInstance(ctx, instanceId, updateParam, UpdateInstanceBasicInfoPath)
}

func (c *Client) UpdateKafkaInstanceVersion(ctx context.Context, instanceId string, version string) (*KafkaInstanceResponse, error) {
	updateParam := InstanceVersionUpgradeParam{Version: version}
	return c.updateInstance(ctx, instanceId, updateParam, UpdateInstanceVersionPath)
}

func (c *Client) UpdateKafkaInstanceConfig(ctx context.Context, instanceId string, updateParam InstanceConfigParam) (*KafkaInstanceResponse, error) {
	return c.updateInstance(ctx, instanceId, updateParam, InstanceConfigPath)
}

func (c *Client) UpdateKafkaInstanceComputeSpecs(ctx context.Context, instanceId string, updateParam SpecificationUpdateParam) (*KafkaInstanceResponse, error) {
	return c.updateInstance(ctx, instanceId, updateParam, UpdateInstanceComputeSpecsPath)
}

func (c *Client) updateInstance(ctx context.Context, instanceId string, updateParam interface{}, path string) (*KafkaInstanceResponse, error) {
	body, err := c.Patch(ctx, fmt.Sprintf(path, instanceId), updateParam)
	if err != nil {
		return nil, err
	}
	instance := KafkaInstanceResponse{}
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, err
	}
	return &instance, nil
}
