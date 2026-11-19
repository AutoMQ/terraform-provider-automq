package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// connectClusterCapacityValidator checks relationships between known configuration
// values. Unknown objects and values are checked when the plan becomes concrete.
type connectClusterCapacityValidator struct{}

func (connectClusterCapacityValidator) Description(context.Context) string {
	return "Capacity must contain only the object selected by type; autoscaling requires ordered bounds and equal CPU targets."
}

func (v connectClusterCapacityValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (connectClusterCapacityValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	mode, _ := attrs["type"].(types.String)
	provisioned := attrs["provisioned"]
	autoscaling := attrs["autoscaling"]
	knownPresent := func(v attr.Value) bool { return v != nil && !v.IsNull() && !v.IsUnknown() }
	knownAbsent := func(v attr.Value) bool { return v == nil || v.IsNull() }
	invalid := func(detail string) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Connect Cluster capacity", detail)
	}
	if knownPresent(provisioned) && knownPresent(autoscaling) {
		invalid("Set only one of provisioned or autoscaling, matching capacity.type.")
	}
	if mode.IsNull() || mode.IsUnknown() {
		return
	}
	switch mode.ValueString() {
	case "provisioned":
		if knownAbsent(provisioned) || knownPresent(autoscaling) {
			invalid("capacity.type provisioned requires the provisioned object and excludes the autoscaling object.")
		}
	case "autoscaling":
		if knownAbsent(autoscaling) || knownPresent(provisioned) {
			invalid("capacity.type autoscaling requires the autoscaling object and excludes the provisioned object.")
		}
		if !knownPresent(autoscaling) {
			return
		}
		auto := autoscaling.(types.Object).Attributes()
		min, minOK := auto["min_worker_count"].(types.Int64)
		max, maxOK := auto["max_worker_count"].(types.Int64)
		if minOK && maxOK && !min.IsUnknown() && !max.IsUnknown() && !min.IsNull() && !max.IsNull() && max.ValueInt64() < min.ValueInt64() {
			invalid("max_worker_count must be greater than or equal to min_worker_count.")
		}
		in, out := auto["scale_in_policy"], auto["scale_out_policy"]
		if knownAbsent(in) || knownAbsent(out) {
			invalid("scale_in_policy and scale_out_policy are both required for autoscaling.")
		}
		if !knownPresent(in) || !knownPresent(out) {
			return
		}
		inCPU, inOK := in.(types.Object).Attributes()["cpu_utilization_percentage"].(types.Int64)
		outCPU, outOK := out.(types.Object).Attributes()["cpu_utilization_percentage"].(types.Int64)
		if inOK && outOK && !inCPU.IsUnknown() && !outCPU.IsUnknown() && !inCPU.IsNull() && !outCPU.IsNull() && !inCPU.Equal(outCPU) {
			invalid("scale_in_policy and scale_out_policy must use the same target CPU utilization percentage.")
		}
	}
}
