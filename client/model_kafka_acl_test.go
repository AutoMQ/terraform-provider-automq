package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaAclImportIdentityRoundTrip(t *testing.T) {
	target := KafkaAclBindingParam{
		AccessControlParam: KafkaControlParam{
			User:           "orders@writer",
			OperationGroup: "PRODUCE",
			PermissionType: "ALLOW",
		},
		ResourcePatternParam: KafkaResourcePatternParam{
			ResourceType: "TOPIC",
			Name:         "orders|priority",
			PatternType:  "LITERAL",
		},
	}

	encoded, err := FormatKafkaAclImportIdentity(target)
	require.NoError(t, err)
	assert.Equal(t, "orders%40writer|TOPIC|ALLOW|orders%7Cpriority|LITERAL|PRODUCE", encoded)

	decoded, err := ParseKafkaAclImportIdentity(encoded)
	require.NoError(t, err)
	assert.Equal(t, target, decoded)
}

func TestParseKafkaAclImportIdentityValidatesFields(t *testing.T) {
	tests := map[string]string{
		"field count":               "writer|TOPIC|ALLOW|orders|LITERAL",
		"resource type":             "writer|BROKER|ALLOW|orders|LITERAL|ALL",
		"permission":                "writer|TOPIC|GRANT|orders|LITERAL|ALL",
		"pattern type":              "writer|TOPIC|ALLOW|orders|MATCH|ALL",
		"topic operation group":     "writer|TOPIC|ALLOW|orders|LITERAL|DELETE",
		"non-topic operation group": "writer|GROUP|ALLOW|orders|LITERAL|CONSUME",
	}

	for name, identity := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ParseKafkaAclImportIdentity(identity)
			require.Error(t, err)
		})
	}
}
