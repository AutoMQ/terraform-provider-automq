package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) CreateKafkaInstance(kafka KafkaInstanceRequest) (*KafkaInstanceResponse, error) {
	kafkaRequest, err := json.Marshal(kafka)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.HostURL+"/api/v1/instances", strings.NewReader(string(kafkaRequest)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req, &c.Token)
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

func (c *Client) GetKafkaInstance(instanceId string) (*KafkaInstanceResponse, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/api/v1/instances/"+instanceId, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	body, err := c.doRequest(req, &c.Token)
	if err != nil {
		return nil, err
	}
	instance := KafkaInstanceResponse{}
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}
	return &instance, nil
}

func (c *Client) DeleteKafkaInstance(instanceId string) error {
	req, err := http.NewRequest("DELETE", c.HostURL+instancePath+"/"+instanceId, nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, &c.Token)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateKafkaInstanceBasicInfo(instanceId string, updateParam InstanceBasicParam) (*KafkaInstanceResponse, error) {
	return c.updateInstance(instanceId, updateParam, "/basic")
}

func (c *Client) UpdateKafkaInstanceVersion(instanceId string, version string) (*KafkaInstanceResponse, error) {
	updateParam := InstanceVersionUpgradeParam{Version: version}
	return c.updateInstance(instanceId, updateParam, "/versions/"+version)
}

func (c *Client) UpdateKafkaInstanceConfig(instanceId string, updateParam InstanceConfigParam) (*KafkaInstanceResponse, error) {
	return c.updateInstance(instanceId, updateParam, "/configurations")
}

func (c *Client) UpdateKafkaInstanceComputeSpecs(instanceId string, updateParam SpecificationUpdateParam) (*KafkaInstanceResponse, error) {
	return c.updateInstance(instanceId, updateParam, "/spec")
}

func (c *Client) updateInstance(instanceId string, updateParam interface{}, path string) (*KafkaInstanceResponse, error) {
	updateRequest, err := json.Marshal(updateParam)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PATCH", c.HostURL+instancePath+"/"+instanceId+path, strings.NewReader(string(updateRequest)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req, &c.Token)
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
