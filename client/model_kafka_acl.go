package client

import (
	"fmt"
	"net/url"
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

func FormatKafkaAclImportIdentity(param KafkaAclBindingParam) (string, error) {
	if err := validateKafkaAclImportIdentity(param); err != nil {
		return "", err
	}
	parts := []string{
		param.AccessControlParam.User,
		param.ResourcePatternParam.ResourceType,
		param.AccessControlParam.PermissionType,
		param.ResourcePatternParam.Name,
		param.ResourcePatternParam.PatternType,
		param.AccessControlParam.OperationGroup,
	}
	for i, part := range parts {
		if part == "" {
			return "", fmt.Errorf("ACL import identity field %d must not be empty", i+1)
		}
		parts[i] = url.QueryEscape(part)
	}
	return strings.Join(parts, "|"), nil
}

func ParseKafkaAclImportIdentity(identity string) (KafkaAclBindingParam, error) {
	parts := strings.Split(identity, "|")
	if len(parts) != 6 {
		return KafkaAclBindingParam{}, fmt.Errorf("invalid ACL import identity: expected 6 fields, got %d", len(parts))
	}
	for i, part := range parts {
		decoded, err := url.QueryUnescape(part)
		if err != nil {
			return KafkaAclBindingParam{}, fmt.Errorf("invalid ACL import identity field %d: %w", i+1, err)
		}
		if decoded == "" {
			return KafkaAclBindingParam{}, fmt.Errorf("ACL import identity field %d must not be empty", i+1)
		}
		parts[i] = decoded
	}
	param := KafkaAclBindingParam{
		AccessControlParam: KafkaControlParam{
			User:           parts[0],
			PermissionType: parts[2],
			OperationGroup: parts[5],
		},
		ResourcePatternParam: KafkaResourcePatternParam{
			ResourceType: parts[1],
			Name:         parts[3],
			PatternType:  parts[4],
		},
	}
	if err := validateKafkaAclImportIdentity(param); err != nil {
		return KafkaAclBindingParam{}, err
	}
	return param, nil
}

func validateKafkaAclImportIdentity(param KafkaAclBindingParam) error {
	resourceType := param.ResourcePatternParam.ResourceType
	switch resourceType {
	case "TOPIC", "GROUP", "CLUSTER", "TRANSACTIONAL_ID":
	default:
		return fmt.Errorf("invalid ACL resource type %q", resourceType)
	}

	permissionType := param.AccessControlParam.PermissionType
	if permissionType != "ALLOW" && permissionType != "DENY" {
		return fmt.Errorf("invalid ACL permission type %q", permissionType)
	}

	patternType := param.ResourcePatternParam.PatternType
	if patternType != "LITERAL" && patternType != "PREFIXED" {
		return fmt.Errorf("invalid ACL pattern type %q", patternType)
	}

	operationGroup := param.AccessControlParam.OperationGroup
	if resourceType == "TOPIC" {
		if operationGroup != "ALL" && operationGroup != "PRODUCE" && operationGroup != "CONSUME" {
			return fmt.Errorf("invalid ACL operation group %q for TOPIC", operationGroup)
		}
	} else if operationGroup != "ALL" {
		return fmt.Errorf("ACL operation group for %s must be ALL", resourceType)
	}
	return nil
}
