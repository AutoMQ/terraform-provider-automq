package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetKafkaAclReturnsExactLogicalMatch(t *testing.T) {
	wildcardHost := "*"
	specificHost := "10.0.0.1"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/instances/kf-test/acls", r.URL.Path)
		assert.Equal(t, "orders-writer", r.URL.Query().Get("exactUser"))
		assert.Equal(t, "TOPIC", r.URL.Query().Get("resourceTypes"))
		assert.Equal(t, "ALLOW", r.URL.Query().Get("permissionType"))
		assert.Equal(t, "orders", r.URL.Query().Get("fuzzyResourceName"))

		response := PageNumResultKafkaAclBindingVO{
			List: []KafkaAclBindingVO{
				newTestKafkaAcl(specificHost, "orders", "LITERAL", "PRODUCE"),
				newTestKafkaAcl(wildcardHost, "orders-copy", "LITERAL", "PRODUCE"),
				newTestKafkaAcl(wildcardHost, "orders", "PREFIXED", "PRODUCE"),
				newTestKafkaAcl(wildcardHost, "orders", "LITERAL", "CONSUME"),
				newTestKafkaAcl(wildcardHost, "orders", "LITERAL", "PRODUCE"),
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer server.Close()

	api, err := NewClient(context.Background(), server.URL, AuthCredentials{AccessKeyID: "key", SecretAccessKey: "secret"})
	require.NoError(t, err)
	api.MaxRetries = 0

	target := KafkaAclBindingParam{
		AccessControlParam: KafkaControlParam{
			User:           "orders-writer",
			OperationGroup: "PRODUCE",
			PermissionType: "ALLOW",
		},
		ResourcePatternParam: KafkaResourcePatternParam{
			ResourceType: "TOPIC",
			Name:         "orders",
			PatternType:  "LITERAL",
		},
	}
	ctx := context.WithValue(context.Background(), EnvIdKey, "env-test")

	actual, err := api.GetKafkaAcl(ctx, "kf-test", target)
	require.NoError(t, err)
	require.NotNil(t, actual)
	assert.Equal(t, "PRODUCE", actual.AccessControl.OperationGroup.Name)
	assert.Equal(t, "LITERAL", actual.ResourcePattern.PatternType)
	assert.Equal(t, "orders", actual.ResourcePattern.Name)
}

func TestGetKafkaAclFollowsPagination(t *testing.T) {
	totalPages := int64(2)
	wildcardHost := "*"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PageNumResultKafkaAclBindingVO{TotalPage: &totalPages}
		switch r.URL.Query().Get("page") {
		case "1":
			response.List = []KafkaAclBindingVO{
				newTestKafkaAcl(wildcardHost, "orders-copy", "LITERAL", "PRODUCE"),
			}
		case "2":
			response.List = []KafkaAclBindingVO{
				newTestKafkaAcl(wildcardHost, "orders", "LITERAL", "PRODUCE"),
			}
		default:
			t.Fatalf("unexpected page query: %q", r.URL.Query().Get("page"))
		}
		require.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer server.Close()

	api, err := NewClient(context.Background(), server.URL, AuthCredentials{AccessKeyID: "key", SecretAccessKey: "secret"})
	require.NoError(t, err)
	api.MaxRetries = 0
	target := KafkaAclBindingParam{
		AccessControlParam:   KafkaControlParam{User: "orders-writer", OperationGroup: "PRODUCE", PermissionType: "ALLOW"},
		ResourcePatternParam: KafkaResourcePatternParam{ResourceType: "TOPIC", Name: "orders", PatternType: "LITERAL"},
	}
	ctx := context.WithValue(context.Background(), EnvIdKey, "env-test")

	actual, err := api.GetKafkaAcl(ctx, "kf-test", target)
	require.NoError(t, err)
	require.NotNil(t, actual)
	assert.Equal(t, "orders", actual.ResourcePattern.Name)
}

func TestGetKafkaAclRejectsAmbiguousMatches(t *testing.T) {
	wildcardHost := "*"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := PageNumResultKafkaAclBindingVO{
			List: []KafkaAclBindingVO{
				newTestKafkaAcl(wildcardHost, "orders", "LITERAL", "PRODUCE"),
				newTestKafkaAcl(wildcardHost, "orders", "LITERAL", "PRODUCE"),
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer server.Close()

	api, err := NewClient(context.Background(), server.URL, AuthCredentials{AccessKeyID: "key", SecretAccessKey: "secret"})
	require.NoError(t, err)
	api.MaxRetries = 0
	target := KafkaAclBindingParam{
		AccessControlParam:   KafkaControlParam{User: "orders-writer", OperationGroup: "PRODUCE", PermissionType: "ALLOW"},
		ResourcePatternParam: KafkaResourcePatternParam{ResourceType: "TOPIC", Name: "orders", PatternType: "LITERAL"},
	}
	ctx := context.WithValue(context.Background(), EnvIdKey, "env-test")

	actual, err := api.GetKafkaAcl(ctx, "kf-test", target)
	require.ErrorContains(t, err, "multiple ACLs matched")
	assert.Nil(t, actual)
}

func newTestKafkaAcl(host, resourceName, patternType, operationGroup string) KafkaAclBindingVO {
	return KafkaAclBindingVO{
		AccessControl: &KafkaAccessControlVO{
			User:           "orders-writer",
			Host:           &host,
			OperationGroup: OperationGroup{Name: operationGroup},
			PermissionType: "ALLOW",
		},
		ResourcePattern: &KafkaResourcePatternVO{
			ResourceType: "TOPIC",
			Name:         resourceName,
			PatternType:  patternType,
		},
	}
}
