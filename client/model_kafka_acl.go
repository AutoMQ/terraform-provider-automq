package client

import (
	"fmt"
	"strings"
)

// KafkaAclBindingVO struct for KafkaAclBindingVO
type KafkaAclBindingVO struct {
	AccessControl   *KafkaAccessControlVO   `json:"accessControl,omitempty"`
	ResourcePattern *KafkaResourcePatternVO `json:"resourcePattern,omitempty"`
}

// KafkaAccessControlVO struct for KafkaAccessControlVO
type KafkaAccessControlVO struct {
	User           string         `json:"user"`
	Host           *string        `json:"host,omitempty"`
	OperationGroup OperationGroup `json:"operationGroup"`
	PermissionType string         `json:"permissionType"`
}

// OperationGroup struct for OperationGroup
type OperationGroup struct {
	Name           string   `json:"name"`
	HitOperations  []string `json:"hitOperations,omitempty"`
	MissOperations []string `json:"missOperations,omitempty"`
}

// KafkaResourcePatternVO struct for KafkaResourcePatternVO
type KafkaResourcePatternVO struct {
	ResourceType string `json:"resourceType"`
	Name         string `json:"name"`
	PatternType  string `json:"patternType"`
}

// KafkaAclBindingParams struct for KafkaAclBindingParams
type KafkaAclBindingParams struct {
	Params []KafkaAclBindingParam `json:"params"`
}

// KafkaAclBindingParam struct for KafkaAclBindingParam
type KafkaAclBindingParam struct {
	AccessControlParam   KafkaControlParam         `json:"accessControlParam,omitempty"`
	ResourcePatternParam KafkaResourcePatternParam `json:"resourcePatternParam,omitempty"`
}

// KafkaControlParam struct for KafkaControlParam
type KafkaControlParam struct {
	User           string  `json:"user"`
	Host           *string `json:"host,omitempty"`
	OperationGroup string  `json:"operationGroup"`
	PermissionType string  `json:"permissionType"`
}

// KafkaResourcePatternParam struct for KafkaResourcePatternParam
type KafkaResourcePatternParam struct {
	ResourceType string `json:"resourceType"`
	Name         string `json:"name"`
	PatternType  string `json:"patternType"`
}

// PageNumResultKafkaAclBindingVO struct for PageNumResultKafkaAclBindingVO
type PageNumResultKafkaAclBindingVO struct {
	PageNum   *int32              `json:"pageNum,omitempty"`
	PageSize  *int32              `json:"pageSize,omitempty"`
	Total     *int64              `json:"total,omitempty"`
	List      []KafkaAclBindingVO `json:"list,omitempty"`
	TotalPage *int64              `json:"totalPage,omitempty"`
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
