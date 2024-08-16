package models

import (
	"testing"

	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestExpandStringValueMap(t *testing.T) {
	planConfig := types.MapValueMust(types.StringType, map[string]attr.Value{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	})

	expected := []client.ConfigItemParam{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}

	result := ExpandStringValueMap(planConfig)

	assert.ElementsMatch(t, expected, result)
}

func TestFlattenStringValueMap(t *testing.T) {
	configs := []client.ConfigItemParam{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}

	expected := types.MapValueMust(types.StringType, map[string]attr.Value{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	})

	result := FlattenStringValueMap(configs)

	assert.True(t, MapsEqual(expected, result))
}

func TestMapsEqual(t *testing.T) {
	map1 := types.MapValueMust(types.StringType, map[string]attr.Value{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	})

	map2 := types.MapValueMust(types.StringType, map[string]attr.Value{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	})

	map3 := types.MapValueMust(types.StringType, map[string]attr.Value{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("different_value"),
	})

	assert.True(t, MapsEqual(map1, map2))
	assert.False(t, MapsEqual(map1, map3))
}
