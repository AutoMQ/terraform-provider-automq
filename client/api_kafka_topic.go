package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	TopicPath                     = "/api/v1/instances/%s/topics"
	GetKafkaTopicPath             = "/api/v1/instances/%s/topics/%s"
	DeleteKafkaTopicPath          = "/api/v1/instances/%s/topics/%s"
	UpdateKafkaTopicConfigPath    = "/api/v1/instances/%s/topics/%s/configurations"
	UpdateKafkaTopicPartitionPath = "/api/v1/instances/%s/topics/%s/partition-counts"
)

func (c *Client) CreateKafkaTopic(ctx context.Context, instanceId string, topic TopicCreateParam) (*TopicVO, error) {
	body, err := c.Post(ctx, fmt.Sprintf(TopicPath, instanceId), topic)
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

func (c *Client) DeleteKafkaTopic(ctx context.Context, instanceId string, topicId string) error {
	_, err := c.Delete(ctx, fmt.Sprintf(DeleteKafkaTopicPath, instanceId, topicId))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateKafkaTopicConfig(ctx context.Context, instanceId string, topicId string, params TopicConfigParam) (*TopicVO, error) {
	path := fmt.Sprintf(UpdateKafkaTopicConfigPath, instanceId, topicId)
	body, err := c.Patch(ctx, path, params)
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

func (c *Client) UpdateKafkaTopicPartition(ctx context.Context, instanceId string, topicId string, partition TopicPartitionParam) error {
	path := fmt.Sprintf(UpdateKafkaTopicPartitionPath, instanceId, topicId)
	_, err := c.Patch(ctx, path, partition)
	if err != nil {
		return nil
	}
	return nil
}

func (c *Client) GetKafkaTopic(ctx context.Context, instanceId string, topicId string) (*TopicVO, error) {
	path := fmt.Sprintf(GetKafkaTopicPath, instanceId, topicId)
	body, err := c.Get(ctx, path, nil)
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
