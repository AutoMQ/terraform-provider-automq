package models

import (
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectorPluginResourceModel struct {
	EnvironmentID     types.String      `tfsdk:"environment_id"`
	ID                types.String      `tfsdk:"id"`
	Name              types.String      `tfsdk:"name"`
	Version           types.String      `tfsdk:"version"`
	StorageUrl        types.String      `tfsdk:"storage_url"`
	Types             types.List        `tfsdk:"types"`
	ConnectorClass    types.String      `tfsdk:"connector_class"`
	Description       types.String      `tfsdk:"description"`
	DocumentationLink types.String      `tfsdk:"documentation_link"`
	PluginProvider    types.String      `tfsdk:"plugin_provider"`
	Status            types.String      `tfsdk:"status"`
	CreatedAt         timetypes.RFC3339 `tfsdk:"created_at"`
	UpdatedAt         timetypes.RFC3339 `tfsdk:"updated_at"`
	Timeouts          timeouts.Value    `tfsdk:"timeouts"`
}

// ExpandConnectorPluginCreate converts the Terraform plan into an API create request.
func ExpandConnectorPluginCreate(plan ConnectorPluginResourceModel) (*client.ConnectPluginCreateParam, diag.Diagnostics) {
	var diags diag.Diagnostics

	request := &client.ConnectPluginCreateParam{
		Name:           plan.Name.ValueString(),
		Version:        plan.Version.ValueString(),
		StorageUrl:     plan.StorageUrl.ValueString(),
		ConnectorClass: plan.ConnectorClass.ValueString(),
		Types:          ExpandStringValueList(plan.Types),
	}
	if s := cpOptStr(plan.Description); s != nil {
		request.Description = s
	}
	if s := cpOptStr(plan.DocumentationLink); s != nil {
		request.DocumentationLink = s
	}
	return request, diags
}

// FlattenConnectorPlugin maps an API response into the Terraform state model.
func FlattenConnectorPlugin(vo *client.ConnectPluginVO, state *ConnectorPluginResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if vo == nil {
		diags.AddError("Invalid connector plugin", "nil connector plugin received")
		return diags
	}
	state.ID = cpToStr(vo.Id)
	state.Name = cpToStr(vo.Name)
	state.Description = cpToStr(vo.Description)
	state.DocumentationLink = cpToStr(vo.DocumentationLink)
	state.Version = cpToStr(vo.Version)
	state.StorageUrl = cpToStr(vo.StorageUrl)
	state.PluginProvider = cpToStr(vo.Provider)
	state.Status = cpToStr(vo.Status)

	// connector_class: prefer explicit field, fall back to sinkConnectorClasses/sourceConnectorClasses
	if vo.ConnectorClass != nil && *vo.ConnectorClass != "" {
		state.ConnectorClass = types.StringValue(*vo.ConnectorClass)
	} else if len(vo.SinkConnectorClasses) > 0 {
		state.ConnectorClass = types.StringValue(vo.SinkConnectorClasses[0])
	} else if len(vo.SourceConnectorClasses) > 0 {
		state.ConnectorClass = types.StringValue(vo.SourceConnectorClasses[0])
	}

	if len(vo.Types) > 0 {
		elems := make([]attr.Value, len(vo.Types))
		for i, t := range vo.Types {
			elems[i] = types.StringValue(t)
		}
		state.Types = types.ListValueMust(types.StringType, elems)
	}

	if vo.CreateTime != nil {
		state.CreatedAt = timetypes.NewRFC3339TimePointerValue(vo.CreateTime)
	}
	if vo.UpdateTime != nil {
		state.UpdatedAt = timetypes.NewRFC3339TimePointerValue(vo.UpdateTime)
	}
	return diags
}

// ---------------------------------------------------------------------------
// Internal helpers (prefixed cp to avoid collision with connector helpers)
// ---------------------------------------------------------------------------

func cpOptStr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func cpToStr(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}
