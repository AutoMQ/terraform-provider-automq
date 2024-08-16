package models

import (
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func ExpandStringValueList(v basetypes.ListValuable) []string {
	var output []string
	if listValue, ok := v.(basetypes.ListValue); ok {
		for _, value := range listValue.Elements() {
			if stringValue, ok := value.(types.String); ok {
				output = append(output, stringValue.ValueString())
			}
		}
	}
	return output
}

func ExpandStringValueMap(planConfig basetypes.MapValue) []client.ConfigItemParam {
	configs := make([]client.ConfigItemParam, len(planConfig.Elements()))
	i := 0
	for name, value := range planConfig.Elements() {
		config, ok := value.(types.String)
		if ok {
			configs[i] = client.ConfigItemParam{
				Key:   name,
				Value: config.ValueString(),
			}
		}
		i += 1
	}
	return configs
}

func FlattenStringValueMap(configs []client.ConfigItemParam) basetypes.MapValue {
	configMap := make(map[string]attr.Value, len(configs))
	for _, config := range configs {
		configMap[config.Key] = types.StringValue(config.Value)
	}
	return types.MapValueMust(types.StringType, configMap)
}

func MapsEqual(a, b types.Map) bool {
	if len(a.Elements()) != len(b.Elements()) {
		return false
	}
	for k, v := range a.Elements() {
		if bVal, ok := b.Elements()[k]; !ok {
			return false
		} else {
			aVal, aOK := v.(types.String)
			bVal, bOK := bVal.(types.String)

			if !aOK || !bOK {
				// if one of the values is not a string, we can't compare them
				// so we just return false.
				return false
			}

			if aVal.ValueString() != bVal.ValueString() {
				return false
			}
		}
	}
	return true
}
