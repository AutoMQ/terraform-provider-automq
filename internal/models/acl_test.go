package models

import (
	"terraform-provider-automq/client"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestExpandKafkaACLResource(t *testing.T) {
	tests := []struct {
		acl      KafkaAclResourceModel
		expected client.KafkaAclBindingParam
	}{
		{
			acl: KafkaAclResourceModel{
				EnvironmentID:  types.StringValue("env-123"),
				KafkaInstance:  types.StringValue("kafka-123"),
				ID:             types.StringValue("acl-123"),
				ResourceType:   types.StringValue("topic"),
				ResourceName:   types.StringValue("test-topic"),
				PatternType:    types.StringValue("literal"),
				Principal:      types.StringValue("User:test-user"),
				OperationGroup: types.StringValue("read"),
				Permission:     types.StringValue("allow"),
			},
			expected: client.KafkaAclBindingParam{
				AccessControlParam: client.KafkaControlParam{
					OperationGroup: "read",
					PermissionType: "allow",
					User:           "test-user",
				},
				ResourcePatternParam: client.KafkaResourcePatternParam{
					Name:         "test-topic",
					PatternType:  "literal",
					ResourceType: "topic",
				},
			},
		},
	}

	for _, test := range tests {
		request := &client.KafkaAclBindingParam{}
		ExpandKafkaACLResource(test.acl, request)

		assert.Equal(t, test.expected.AccessControlParam.OperationGroup, request.AccessControlParam.OperationGroup)
		assert.Equal(t, test.expected.AccessControlParam.PermissionType, request.AccessControlParam.PermissionType)
		assert.Equal(t, test.expected.AccessControlParam.User, request.AccessControlParam.User)
		assert.Equal(t, test.expected.ResourcePatternParam.Name, request.ResourcePatternParam.Name)
		assert.Equal(t, test.expected.ResourcePatternParam.PatternType, request.ResourcePatternParam.PatternType)
		assert.Equal(t, test.expected.ResourcePatternParam.ResourceType, request.ResourcePatternParam.ResourceType)
	}
}

func TestParsePrincipalUser(t *testing.T) {
	tests := []struct {
		principal string
		hasDiag   bool
		expected  string
	}{
		{
			principal: "User:test-user",
			expected:  "test-user",
			hasDiag:   false,
		},
		{
			principal: "User:admin",
			expected:  "admin",
			hasDiag:   false,
		},
		{
			principal: "User:admin:admin",
			expected:  "admin:admin",
			hasDiag:   false,
		},
		{
			principal: "User:",
			expected:  "",
			hasDiag:   true,
		},
		{
			principal: "User",
			expected:  "",
			hasDiag:   true,
		},
	}
	for _, test := range tests {
		user, err := ParsePrincipalUser(test.principal)

		if test.hasDiag {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, test.expected, user)
		}
	}
}

func TestFlattenKafkaACLResource(t *testing.T) {
	tests := []struct {
		acl      *client.KafkaAclBindingVO
		expected KafkaAclResourceModel
	}{
		{
			acl: &client.KafkaAclBindingVO{
				ResourcePattern: &client.KafkaResourcePatternVO{
					ResourceType: "topic",
					Name:         "test-topic",
					PatternType:  "literal",
				},
				AccessControl: &client.KafkaAccessControlVO{
					User:           "test-user",
					OperationGroup: client.OperationGroup{Name: "read"},
					PermissionType: "allow",
				},
			},
			expected: KafkaAclResourceModel{
				ResourceType:   types.StringValue("topic"),
				ResourceName:   types.StringValue("test-topic"),
				PatternType:    types.StringValue("literal"),
				Principal:      types.StringValue("User:test-user"),
				OperationGroup: types.StringValue("read"),
				Permission:     types.StringValue("allow"),
			},
		},
	}

	for _, test := range tests {
		resource := &KafkaAclResourceModel{}
		diag := FlattenKafkaACLResource(test.acl, resource)

		assert.Nil(t, diag)

		assert.Equal(t, test.expected.ResourceType.ValueString(), resource.ResourceType.ValueString())
		assert.Equal(t, test.expected.ResourceName.ValueString(), resource.ResourceName.ValueString())
		assert.Equal(t, test.expected.PatternType.ValueString(), resource.PatternType.ValueString())
		assert.Equal(t, test.expected.Principal.ValueString(), resource.Principal.ValueString())
		assert.Equal(t, test.expected.OperationGroup.ValueString(), resource.OperationGroup.ValueString())
		assert.Equal(t, test.expected.Permission.ValueString(), resource.Permission.ValueString())
	}
}
