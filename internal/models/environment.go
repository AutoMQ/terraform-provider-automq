package models

import (
	"terraform-provider-automq/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	CloudProvider    types.String `tfsdk:"cloud_provider"`
	Region           types.String `tfsdk:"region"`
	Scope            types.String `tfsdk:"scope"`
	Product          types.String `tfsdk:"product"`
	State            types.String `tfsdk:"state"`
	OpsBucket        types.String `tfsdk:"ops_bucket"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	OrganizationName types.String `tfsdk:"organization_name"`
	Creator          types.String `tfsdk:"creator"`
	CreatorName      types.String `tfsdk:"creator_name"`
	Stateless        types.Bool   `tfsdk:"stateless"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func ExpandEnvironmentCreate(plan EnvironmentResourceModel) client.EnvironmentCreateParam {
	return client.EnvironmentCreateParam{
		Name:          plan.Name.ValueString(),
		Description:   optionalString(plan.Description),
		CloudProvider: plan.CloudProvider.ValueString(),
		Region:        plan.Region.ValueString(),
		Scope:         plan.Scope.ValueString(),
		Product:       plan.Product.ValueString(),
	}
}

func ExpandEnvironmentUpdate(plan EnvironmentResourceModel) client.EnvironmentUpdateParam {
	return client.EnvironmentUpdateParam{
		Name:        plan.Name.ValueString(),
		Description: optionalString(plan.Description),
	}
}

func FlattenEnvironment(environment *client.EnvironmentVO, state *EnvironmentResourceModel) {
	state.ID = types.StringValue(environment.EnvID)
	state.Name = types.StringValue(environment.Name)
	state.Description = stringPointerValue(environment.Description)
	state.CloudProvider = types.StringValue(environment.CloudProvider)
	state.Region = types.StringValue(environment.Region)
	state.Scope = types.StringValue(environment.Scope)
	state.State = types.StringValue(environment.State)
	state.OpsBucket = stringPointerValue(environment.OpsBucket)
	state.OrganizationID = stringPointerValue(environment.OrganizationID)
	state.OrganizationName = stringPointerValue(environment.OrganizationName)
	state.Creator = stringPointerValue(environment.Creator)
	state.CreatorName = stringPointerValue(environment.CreatorName)
	state.Stateless = types.BoolValue(environment.Stateless)
	if state.Product.IsNull() || state.Product.IsUnknown() {
		state.Product = types.StringValue("kafka")
	}
	if environment.CreatedAt != nil {
		state.CreatedAt = types.StringValue(environment.CreatedAt.Format(time.RFC3339))
	}
	if environment.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(environment.UpdatedAt.Format(time.RFC3339))
	}
}

func optionalString(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	result := value.ValueString()
	return &result
}

func stringPointerValue(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}
