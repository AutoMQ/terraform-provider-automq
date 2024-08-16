package models

import (
	"terraform-provider-automq/client"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestFlattenComputeSpecs(t *testing.T) {
	tests := []struct {
		name     string
		input    client.Spec
		expected ComputeSpecsModel
	}{
		{
			name: "Valid Spec",
			input: client.Spec{
				Version: "1.0.0",
				Values: []client.Value{
					{Key: "aku", Value: 4},
				},
			},
			expected: ComputeSpecsModel{
				Aku:     types.Int64Value(4),
				Version: types.StringValue("1.0.0"),
			},
		},
		{
			name: "Missing Aku",
			input: client.Spec{
				Version: "1.0.0",
				Values:  []client.Value{},
			},
			expected: ComputeSpecsModel{
				Aku:     types.Int64Value(0),
				Version: types.StringValue("1.0.0"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenComputeSpecs(tt.input)
			assert.Equal(t, tt.expected.Aku.ValueInt64(), result.Aku.ValueInt64())
			assert.Equal(t, tt.expected.Version.ValueString(), result.Version.ValueString())
		})
	}
}

func TestExpandKafkaInstanceResource(t *testing.T) {
	instance := KafkaInstanceResourceModel{
		Name:          types.StringValue("test-instance"),
		Description:   types.StringValue("test-description"),
		CloudProvider: types.StringValue("aws"),
		Region:        types.StringValue("us-west-2"),
		Networks: []NetworkModel{
			{
				Zone:    types.StringValue("zone-1"),
				Subnets: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("subnet-1")}),
			},
		},
		ComputeSpecs: &ComputeSpecsModel{
			Aku:     types.Int64Value(4),
			Version: types.StringValue("1.0.0"),
		},
		Configs:      types.MapValueMust(types.StringType, map[string]attr.Value{"config-key": types.StringValue("config-value")}),
		ACL:          types.BoolValue(true),
		Integrations: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("integration-1")}),
	}

	expected := client.KafkaInstanceRequest{
		DisplayName: "test-instance",
		Description: "test-description",
		Provider:    "aws",
		Region:      "us-west-2",
		Networks: []client.KafkaInstanceRequestNetwork{
			{
				Zone:   "zone-1",
				Subnet: "subnet-1",
			},
		},
		Spec: client.KafkaInstanceRequestSpec{
			Template:    "aku",
			PaymentPlan: client.KafkaInstanceRequestPaymentPlan{PaymentType: "ON_DEMAND", Period: 1, Unit: "MONTH"},
			Values:      []client.ConfigItemParam{{Key: "aku", Value: "4"}},
			Version:     "1.0.0",
		},
		InstanceConfig: client.InstanceConfigParam{
			Configs: []client.ConfigItemParam{{Key: "config-key", Value: "config-value"}},
		},
		Integrations: []string{"integration-1"},
		AclEnabled:   true,
	}

	request := &client.KafkaInstanceRequest{}
	ExpandKafkaInstanceResource(instance, request)

	assert.Equal(t, expected, *request)
}
