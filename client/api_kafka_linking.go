package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	kafkaLinksPath            = "/api/v1/instances/%s/kafka-links"
	kafkaLinkPath             = "/api/v1/instances/%s/kafka-links/%s"
	kafkaLinkMirrorTopicsPath = "/api/v1/instances/%s/kafka-links/%s/kafka-link-mirror-topics"
	kafkaLinkMirrorTopicPath  = "/api/v1/instances/%s/kafka-links/%s/kafka-link-mirror-topics/%s"
	kafkaLinkMirrorGroupsPath = "/api/v1/instances/%s/kafka-links/%s/kafka-link-mirror-groups"
	kafkaLinkMirrorGroupPath  = "/api/v1/instances/%s/kafka-links/%s/kafka-link-mirror-groups/%s"
)

func (c *Client) CreateKafkaLink(ctx context.Context, instanceID string, param KafkaLinkCreateParam) (*KafkaLinkVO, error) {
	body, err := c.Post(ctx, fmt.Sprintf(kafkaLinksPath, instanceID), param)
	if err != nil {
		return nil, err
	}
	link := &KafkaLinkVO{}
	if err := json.Unmarshal(body, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (c *Client) GetKafkaLink(ctx context.Context, instanceID, linkID string) (*KafkaLinkVO, error) {
	body, err := c.Get(ctx, fmt.Sprintf(kafkaLinkPath, instanceID, linkID), nil)
	if err != nil {
		return nil, err
	}
	link := &KafkaLinkVO{}
	if err := json.Unmarshal(body, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (c *Client) ListKafkaLinks(ctx context.Context, instanceID string, query map[string]string) (*PageNumResultKafkaLinkVO, error) {
	body, err := c.Get(ctx, fmt.Sprintf(kafkaLinksPath, instanceID), query)
	if err != nil {
		return nil, err
	}
	result := &PageNumResultKafkaLinkVO{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) DeleteKafkaLink(ctx context.Context, instanceID, linkID string) error {
	if _, err := c.Delete(ctx, fmt.Sprintf(kafkaLinkPath, instanceID, linkID)); err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateKafkaLinkMirrorTopics(ctx context.Context, instanceID, linkID string, param KafkaLinkMirrorTopicsCreateParam) (*MirrorTopicListVO, error) {
	path := fmt.Sprintf(kafkaLinkMirrorTopicsPath, instanceID, linkID)
	body, err := c.Post(ctx, path, param)
	if err != nil {
		return nil, err
	}
	result := &MirrorTopicListVO{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ListKafkaLinkMirrorTopics(ctx context.Context, instanceID, linkID string, query map[string]string) (*PageNumResultMirrorTopicVO, error) {
	path := fmt.Sprintf(kafkaLinkMirrorTopicsPath, instanceID, linkID)
	body, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	result := &PageNumResultMirrorTopicVO{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) UpdateKafkaLinkMirrorTopic(ctx context.Context, instanceID, linkID, topicID string, param KafkaLinkMirrorTopicsUpdateParam) error {
	path := fmt.Sprintf(kafkaLinkMirrorTopicPath, instanceID, linkID, topicID)
	if _, err := c.Patch(ctx, path, param); err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteKafkaLinkMirrorTopic(ctx context.Context, instanceID, linkID, topicID string) error {
	path := fmt.Sprintf(kafkaLinkMirrorTopicPath, instanceID, linkID, topicID)
	if _, err := c.Delete(ctx, path); err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateKafkaLinkMirrorGroups(ctx context.Context, instanceID, linkID string, param KafkaLinkMirrorGroupsCreateParam) (*MirrorConsumerGroupListVO, error) {
	path := fmt.Sprintf(kafkaLinkMirrorGroupsPath, instanceID, linkID)
	body, err := c.Post(ctx, path, param)
	if err != nil {
		return nil, err
	}
	result := &MirrorConsumerGroupListVO{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ListKafkaLinkMirrorGroups(ctx context.Context, instanceID, linkID string, query map[string]string) (*PageNumResultMirrorConsumerGroupVO, error) {
	path := fmt.Sprintf(kafkaLinkMirrorGroupsPath, instanceID, linkID)
	body, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	result := &PageNumResultMirrorConsumerGroupVO{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) DeleteKafkaLinkMirrorGroup(ctx context.Context, instanceID, linkID, groupID string) error {
	path := fmt.Sprintf(kafkaLinkMirrorGroupPath, instanceID, linkID, groupID)
	if _, err := c.Delete(ctx, path); err != nil {
		return err
	}
	return nil
}
