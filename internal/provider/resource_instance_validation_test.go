package provider

import (
	"context"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestKafkaInstanceReservedAKUValidation(t *testing.T) {
	var schemaResponse frameworkresource.SchemaResponse
	(&KafkaInstanceResource{}).Schema(context.Background(), frameworkresource.SchemaRequest{}, &schemaResponse)

	computeSpecs, ok := schemaResponse.Schema.Attributes["compute_specs"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("compute_specs is not a single nested attribute")
	}
	reservedAKU, ok := computeSpecs.Attributes["reserved_aku"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("reserved_aku is not an int64 attribute")
	}
	if len(reservedAKU.Validators) != 1 {
		t.Fatalf("expected one reserved_aku validator, got %d", len(reservedAKU.Validators))
	}

	testCases := []struct {
		value int64
		valid bool
	}{
		{value: 3, valid: false},
		{value: 6, valid: true},
		{value: 8, valid: true},
		{value: 10, valid: true},
		{value: 12, valid: true},
		{value: 14, valid: true},
		{value: 15, valid: false},
		{value: 16, valid: true},
		{value: 18, valid: true},
		{value: 20, valid: true},
		{value: 22, valid: true},
		{value: 24, valid: true},
		{value: 26, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(types.Int64Value(testCase.value).String(), func(t *testing.T) {
			request := validator.Int64Request{ConfigValue: types.Int64Value(testCase.value)}
			var response validator.Int64Response
			reservedAKU.Validators[0].ValidateInt64(context.Background(), request, &response)

			if testCase.valid && response.Diagnostics.HasError() {
				t.Fatalf("expected %d AKU to be valid, got: %v", testCase.value, response.Diagnostics)
			}
			if !testCase.valid && !response.Diagnostics.HasError() {
				t.Fatalf("expected %d AKU to be rejected", testCase.value)
			}
		})
	}
}
