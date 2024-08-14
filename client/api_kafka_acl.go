package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	KafkaAclPath  = "/api/v1/instances/%s/acls"
	DeleteAclPath = "/api/v1/instances/%s/acls/batch:delete"
)

func (c *Client) CreateKafkaAcls(ctx context.Context, instanceId string, param KafkaAclBindingParams) (*KafkaAclBindingVO, error) {
	body, err := c.Post(ctx, fmt.Sprintf(KafkaAclPath, instanceId), param)
	if err != nil {
		return nil, err
	}
	acl := PageNumResultKafkaAclBindingVO{}
	err = json.Unmarshal(body, &acl)
	if err != nil {
		return nil, err
	}
	if len(acl.List) == 0 {
		return nil, &ErrorResponse{Code: 404, ErrorMessage: "acl not found"}
	}
	return &acl.List[0], nil
}

func (c *Client) DeleteKafkaAcls(ctx context.Context, instanceId string, param KafkaAclBindingParams) error {
	_, err := c.Post(ctx, fmt.Sprintf(DeleteAclPath, instanceId), param)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetKafkaAcls(ctx context.Context, instanceId string, aclId string) (*KafkaAclBindingVO, error) {
	localVarPath := fmt.Sprintf(KafkaAclPath, instanceId)
	user, resourceType, permissionType, resourceName, err := ParseAclID(aclId)
	if err != nil {
		return nil, err
	}
	queryParams := make(map[string]string)
	queryParams["exactUser"] = user
	queryParams["resourceTypes"] = resourceType
	queryParams["permissionType"] = permissionType
	queryParams["fuzzyResourceName"] = resourceName

	body, err := c.Get(ctx, localVarPath, queryParams)
	if err != nil {
		return nil, err
	}
	acl := PageNumResultKafkaAclBindingVO{}
	err = json.Unmarshal(body, &acl)
	if err != nil {
		return nil, err
	}
	if len(acl.List) == 0 {
		return nil, &ErrorResponse{Code: 404, ErrorMessage: "acl not found"}
	}
	return &acl.List[0], nil
}
