package provider

import (
	"context"
	"math"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestConnectClusterCapacitySchema(t *testing.T) {
	ctx := context.Background()
	var resp resource.SchemaResponse
	NewConnectClusterResource().Schema(ctx, resource.SchemaRequest{}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	capacity := resp.Schema.Attributes["capacity"].(schema.SingleNestedAttribute)
	assert.Len(t, capacity.Validators, 1)
	for _, mode := range []string{"provisioned", "autoscaling"} {
		nested := capacity.Attributes[mode].(schema.SingleNestedAttribute)
		assert.Empty(t, nested.PlanModifiers, "capacity changes must update in place")
		if mode == "autoscaling" {
			assert.True(t, nested.Attributes["scale_in_policy"].IsRequired())
			assert.True(t, nested.Attributes["scale_out_policy"].IsRequired())
		}
		for _, field := range []string{"worker_count", "min_worker_count", "max_worker_count"} {
			a, ok := nested.Attributes[field].(schema.Int64Attribute)
			if !ok {
				continue
			}
			for _, value := range []int64{0, 1, 256, math.MaxInt32, math.MaxInt32 + 1} {
				var result validator.Int64Response
				for _, v := range a.Validators {
					v.ValidateInt64(ctx, validator.Int64Request{Path: path.Root("capacity").AtName(mode).AtName(field), ConfigValue: types.Int64Value(value)}, &result)
				}
				assert.Equal(t, value < 1 || value > math.MaxInt32, result.Diagnostics.HasError(), "%s=%d", field, value)
			}
		}
	}
}

func TestConnectClusterCapacityRelationships(t *testing.T) {
	policyType := types.ObjectType{AttrTypes: map[string]attr.Type{"cpu_utilization_percentage": types.Int64Type}}
	autoType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"min_worker_count": types.Int64Type, "max_worker_count": types.Int64Type,
		"scale_in_policy": policyType, "scale_out_policy": policyType,
	}}
	fixedType := types.ObjectType{AttrTypes: map[string]attr.Type{"worker_count": types.Int64Type}}
	capacityType := types.ObjectType{AttrTypes: map[string]attr.Type{"type": types.StringType, "provisioned": fixedType, "autoscaling": autoType}}
	policy := func(cpu types.Int64) types.Object {
		return types.ObjectValueMust(policyType.AttrTypes, map[string]attr.Value{"cpu_utilization_percentage": cpu})
	}
	fixed := types.ObjectValueMust(fixedType.AttrTypes, map[string]attr.Value{"worker_count": types.Int64Value(2)})
	auto := func(min, max types.Int64, in, out types.Object) types.Object {
		return types.ObjectValueMust(autoType.AttrTypes, map[string]attr.Value{"min_worker_count": min, "max_worker_count": max, "scale_in_policy": in, "scale_out_policy": out})
	}
	validAuto := auto(types.Int64Value(1), types.Int64Value(256), policy(types.Int64Value(70)), policy(types.Int64Value(70)))
	capacity := func(mode types.String, provisioned, autoscaling types.Object) types.Object {
		return types.ObjectValueMust(capacityType.AttrTypes, map[string]attr.Value{"type": mode, "provisioned": provisioned, "autoscaling": autoscaling})
	}
	cases := []struct {
		name    string
		value   types.Object
		invalid bool
	}{
		{"provisioned", capacity(types.StringValue("provisioned"), fixed, types.ObjectNull(autoType.AttrTypes)), false},
		{"autoscaling min one max above 200", capacity(types.StringValue("autoscaling"), types.ObjectNull(fixedType.AttrTypes), validAuto), false},
		{"both objects", capacity(types.StringValue("autoscaling"), fixed, validAuto), true},
		{"wrong object", capacity(types.StringValue("provisioned"), types.ObjectNull(fixedType.AttrTypes), validAuto), true},
		{"missing object", capacity(types.StringValue("autoscaling"), types.ObjectNull(fixedType.AttrTypes), types.ObjectNull(autoType.AttrTypes)), true},
		{"inverted bounds", capacity(types.StringValue("autoscaling"), types.ObjectNull(fixedType.AttrTypes), auto(types.Int64Value(7), types.Int64Value(6), policy(types.Int64Value(70)), policy(types.Int64Value(70)))), true},
		{"missing CPU policy", capacity(types.StringValue("autoscaling"), types.ObjectNull(fixedType.AttrTypes), auto(types.Int64Value(1), types.Int64Value(6), types.ObjectNull(policyType.AttrTypes), policy(types.Int64Value(70)))), true},
		{"unequal CPU targets", capacity(types.StringValue("autoscaling"), types.ObjectNull(fixedType.AttrTypes), auto(types.Int64Value(1), types.Int64Value(6), policy(types.Int64Value(50)), policy(types.Int64Value(70)))), true},
		{"unknown whole object", types.ObjectUnknown(capacityType.AttrTypes), false},
		{"unknown nested object", capacity(types.StringValue("autoscaling"), types.ObjectNull(fixedType.AttrTypes), types.ObjectUnknown(autoType.AttrTypes)), false},
		{"unknown bounds and CPU", capacity(types.StringValue("autoscaling"), types.ObjectNull(fixedType.AttrTypes), auto(types.Int64Unknown(), types.Int64Unknown(), policy(types.Int64Unknown()), policy(types.Int64Value(70)))), false},
		{"unknown mode", capacity(types.StringUnknown(), types.ObjectNull(fixedType.AttrTypes), validAuto), false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var result validator.ObjectResponse
			connectClusterCapacityValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{Path: path.Root("capacity"), ConfigValue: tt.value}, &result)
			assert.Equal(t, tt.invalid, result.Diagnostics.HasError(), result.Diagnostics)
		})
	}
}
