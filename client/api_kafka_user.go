package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	KafkaUserPath       = "/api/v1/instances/%s/users"
	DeleteKafkaUserPath = "/api/v1/instances/%s/users/%s"
)

// CreateUser creates a new user
func (c *Client) CreateKafkaUser(ctx context.Context, instanceId string, user InstanceUserCreateParam) (*KafkaUserVO, error) {
	path := fmt.Sprintf(KafkaUserPath, instanceId)
	body, err := c.Post(ctx, path, user)
	if err != nil {
		return nil, err
	}
	newuser := KafkaUserVO{}
	err = json.Unmarshal(body, &newuser)
	if err != nil {
		return nil, err
	}
	return &newuser, nil
}

func (c *Client) DeleteKafkaUser(ctx context.Context, instanceId string, userName string) error {
	path := fmt.Sprintf(DeleteKafkaUserPath, instanceId, userName)
	_, err := c.Delete(ctx, path)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetKafkaUser(ctx context.Context, instanceId string, userName string) (*KafkaUserVO, error) {
	path := fmt.Sprintf(KafkaUserPath, instanceId)
	queryParams := make(map[string]string)
	queryParams["userNames"] = userName
	body, err := c.Get(ctx, path, queryParams)
	if err != nil {
		return nil, err
	}
	userPage := PageNumResultKafkaUserVO{}
	err = json.Unmarshal(body, &userPage)
	if err != nil {
		return nil, err
	}
	if len(userPage.List) == 0 {
		return nil, &ErrorResponse{Code: 404, ErrorMessage: "user not found"}
	}
	user := userPage.List[0]
	return &user, nil
}
