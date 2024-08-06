package client

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *Client) CreateKafkaInstance(kafka KafkaInstanceRequest) (*KafkaInstanceResponce, error) {
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

	newkafka := KafkaInstanceResponce{}
	err = json.Unmarshal(body, &newkafka)
	if err != nil {
		return nil, err
	}

	return &newkafka, nil
}

func (c *Client) GetKafkaInstance(id string) (*KafkaInstanceResponce, error) {
	req, err := http.NewRequest("GET", c.HostURL+"/api/v1/instances/"+id, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, &c.Token)
	if err != nil {
		return nil, err
	}

	kafka := KafkaInstanceResponce{}
	err = json.Unmarshal(body, &kafka)
	if err != nil {
		return nil, err
	}

	return &kafka, nil
}
