package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	for page := 1; ; page++ {
		body, err := c.Get(ctx, path, map[string]string{
			"userNames": userName,
			"page":      strconv.Itoa(page),
		})
		if err != nil {
			return nil, err
		}
		var userPage PageNumResultKafkaUserVO
		if err := json.Unmarshal(body, &userPage); err != nil {
			return nil, err
		}
		for _, user := range userPage.List {
			if user.Name == userName {
				return &user, nil
			}
		}
		if userPage.TotalPage == nil || int64(page) >= *userPage.TotalPage || len(userPage.List) == 0 {
			return nil, &ErrorResponse{Code: 404, ErrorMessage: "user not found"}
		}
	}
}
