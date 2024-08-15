package models

import (
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func CreateConfigFromMapValue(planConfig basetypes.MapValue) []client.ConfigItemParam {
	configs := make([]client.ConfigItemParam, len(planConfig.Elements()))
	i := 0
	for name, value := range planConfig.Elements() {
		config := value.(types.String)
		configs[i] = client.ConfigItemParam{
			Key:   name,
			Value: config.ValueString(),
		}
		i += 1
	}
	return configs
}

func CreateMapFromConfigValue(configs []client.ConfigItemParam) basetypes.MapValue {
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
			aVal := v.(types.String)
			bVal := bVal.(types.String)

			if aVal.ValueString() != bVal.ValueString() {
				return false
			}
		}
	}
	return true
}
