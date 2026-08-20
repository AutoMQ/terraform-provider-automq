package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

func (c *Client) GetKafkaAcl(ctx context.Context, instanceId string, target KafkaAclBindingParam) (*KafkaAclBindingVO, error) {
	const pageSize = 100
	matches := make([]KafkaAclBindingVO, 0, 1)
	queryParams := map[string]string{
		"exactUser":         target.AccessControlParam.User,
		"resourceTypes":     target.ResourcePatternParam.ResourceType,
		"permissionType":    target.AccessControlParam.PermissionType,
		"fuzzyResourceName": target.ResourcePatternParam.Name,
		"size":              strconv.Itoa(pageSize),
	}
	for pageNumber := 1; ; pageNumber++ {
		queryParams["page"] = strconv.Itoa(pageNumber)
		body, err := c.Get(ctx, fmt.Sprintf(KafkaAclPath, instanceId), queryParams)
		if err != nil {
			return nil, err
		}
		var page PageNumResultKafkaAclBindingVO
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		for i := range page.List {
			if kafkaAclMatches(page.List[i], target) {
				matches = append(matches, page.List[i])
			}
		}
		if page.TotalPage != nil {
			if int64(pageNumber) >= *page.TotalPage {
				break
			}
			continue
		}
		if len(page.List) < pageSize {
			break
		}
	}
	switch len(matches) {
	case 0:
		return nil, &ErrorResponse{Code: 404, ErrorMessage: "acl not found"}
	case 1:
		return &matches[0], nil
	default:
		return nil, fmt.Errorf("multiple ACLs matched the requested identity")
	}
}

func kafkaAclMatches(candidate KafkaAclBindingVO, target KafkaAclBindingParam) bool {
	if candidate.AccessControl == nil || candidate.ResourcePattern == nil {
		return false
	}
	host := candidate.AccessControl.Host
	if host != nil && *host != "" && *host != "*" {
		return false
	}
	return candidate.AccessControl.User == target.AccessControlParam.User &&
		candidate.AccessControl.PermissionType == target.AccessControlParam.PermissionType &&
		candidate.AccessControl.OperationGroup.Name == target.AccessControlParam.OperationGroup &&
		candidate.ResourcePattern.ResourceType == target.ResourcePatternParam.ResourceType &&
		candidate.ResourcePattern.Name == target.ResourcePatternParam.Name &&
		candidate.ResourcePattern.PatternType == target.ResourcePatternParam.PatternType
}
