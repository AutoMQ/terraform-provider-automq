package client

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *Client) CreateKafkaTopic(instanceId string, topic TopicCreateParam) (*TopicVO, error) {
	topicRequest, err := json.Marshal(topic)
	if err != nil {
		return nil, err
	}
	localVarPath := c.HostURL + "/api/v1/instances/" + instanceId + "/topics"

	req, err := http.NewRequest("POST", localVarPath, strings.NewReader(string(topicRequest)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	newtopic := TopicVO{}
	err = json.Unmarshal(body, &newtopic)
	if err != nil {
		return nil, err
	}
	return &newtopic, nil
}

func (c *Client) DeleteKafkaTopic(instanceId string, topicId string) error {
	localVarPath := c.HostURL + "/api/v1/instances/" + instanceId + "/topics/" + topicId

	req, err := http.NewRequest("DELETE", localVarPath, nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateKafkaTopicConfig(instanceId string, topicId string, topic TopicConfigParam) (*TopicVO, error) {
	topicRequest, err := json.Marshal(topic)
	if err != nil {
		return nil, err
	}
	localVarPath := c.HostURL + "/api/v1/instances/" + instanceId + "/topics/" + topicId + "/configurations"

	req, err := http.NewRequest("PATCH", localVarPath, strings.NewReader(string(topicRequest)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	newtopic := TopicVO{}
	err = json.Unmarshal(body, &newtopic)
	if err != nil {
		return nil, err
	}
	return &newtopic, nil
}

func (c *Client) UpdateKafkaTopicPartition(instanceId string, topicId string, partition TopicPartitionParam) error {
	partitionRequest, err := json.Marshal(partition)
	if err != nil {
		return nil
	}
	localVarPath := c.HostURL + "/api/v1/instances/" + instanceId + "/topics/" + topicId + "/partition-counts"

	req, err := http.NewRequest("PATCH", localVarPath, strings.NewReader(string(partitionRequest)))
	if err != nil {
		return nil
	}
	_, err = c.doRequest(req)
	if err != nil {
		return nil
	}
	return nil
}

func (c *Client) GetKafkaTopic(instanceId string, topicId string) (*TopicVO, error) {
	localVarPath := c.HostURL + "/api/v1/instances/" + instanceId + "/topics/" + topicId

	req, err := http.NewRequest("GET", localVarPath, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	topic := TopicVO{}
	err = json.Unmarshal(body, &topic)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}
