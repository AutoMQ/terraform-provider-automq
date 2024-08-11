package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) CreateKafkaAcls(instanceId string, param KafkaAclBindingParams) (*KafkaAclBindingVO, error) {
	aclRequest, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	localVarPath := c.HostURL + "/api/v1/instances/" + instanceId + "/acls"
	req, err := http.NewRequest("POST", localVarPath, strings.NewReader(string(aclRequest)))
	if err != nil {
		return nil, err
	}
	_, err = c.doRequest(req)
	if err != nil {
		return nil, err
	}
	aclId, err := GenerateAclID(param.Params[0])
	if err != nil {
		return nil, err
	}
	return c.GetKafkaAcls(instanceId, aclId)
}

func (c *Client) DeleteKafkaAcls(instanceId string, param KafkaAclBindingParams) error {
	aclRequest, err := json.Marshal(param)
	if err != nil {
		return err
	}
	localVarPath := c.HostURL + "/api/v1/instances/" + instanceId + "/acls/batch:delete"
	req, err := http.NewRequest("POST", localVarPath, strings.NewReader(string(aclRequest)))
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetKafkaAcls(instanceId string, aclId string) (*KafkaAclBindingVO, error) {
	localVarPath := c.HostURL + "/api/v1/instances/" + instanceId + "/acls"
	user, resourceType, permissionType, resourceName, err := ParseAclID(aclId)
	if err != nil {
		return nil, err
	}
	localVarPath = localVarPath + "?exactUser=" + user + "&resourceTypes=" + resourceType + "&permissionType=" + permissionType + "&fuzzyResourceName=" + resourceName
	req, err := http.NewRequest("GET", localVarPath, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	acl := PageNumResultKafkaAclBindingVO{}
	err = json.Unmarshal(body, &acl)
	if err != nil {
		return nil, err
	}
	if len(acl.List) == 0 {
		return nil, fmt.Errorf("acl not found")
	}
	return &acl.List[0], nil
}

func GenerateAclID(param interface{}) (string, error) {
	switch p := param.(type) {
	case KafkaAclBindingParam:
		return fmt.Sprintf("%s|%s|%s|%s",
			p.AccessControlParam.User,
			p.ResourcePatternParam.ResourceType,
			p.AccessControlParam.PermissionType,
			p.ResourcePatternParam.Name), nil
	case KafkaAclBindingVO:
		return fmt.Sprintf("%s|%s|%s|%s",
			p.AccessControl.User,
			p.ResourcePattern.ResourceType,
			p.AccessControl.PermissionType,
			p.ResourcePattern.Name), nil
	default:
		return "", fmt.Errorf("unsupported type %T", p)
	}
}

func ParseAclID(aclID string) (user string, resourceType string, permissionType string, resourceName string, err error) {
	parts := strings.Split(aclID, "|")
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("invalid aclID")
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}
