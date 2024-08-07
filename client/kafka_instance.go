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
		if err.(*ErrorResponse).Code == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("error doing request: %v", err)
	}
	instance := KafkaInstanceResponse{}
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}
	return &instance, nil
}

func (c *Client) GetKafkaInstanceByName(name string) (*KafkaInstanceResponse, error) {
	req, err := http.NewRequest("GET", c.HostURL+instancePath+"?keyword="+name, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req, &c.Token)
	if err != nil {
		return nil, err
	}
	kafka := KafkaInstanceResponse{}
	err = json.Unmarshal(body, &kafka)
	klist := KafkaInstanceResponseList{}

	err = json.Unmarshal(body, &klist)
	if err != nil {
		return nil, err
	}

	var result KafkaInstanceResponse
	for _, instance := range klist.List {
		if instance.DisplayName == name {
			result = instance
			return &result, nil
		}
	}
	return nil, fmt.Errorf("Kafka instance with name %s not found", name)
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
