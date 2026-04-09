package ref

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Compile-time interface check.
var _ planmodifier.Bool = useStateUnlessResourceChangesBool{}

// UseStateUnlessResourceChangesBool is the Bool variant of
// UseStateUnlessResourceChanges.  See that function's documentation for
// the full rationale.
//
// Use this for Computed+Optional Bool fields (e.g. disable) that have a
// server-computed default but should not show a diff when no other
// attributes in the resource are changing.
func UseStateUnlessResourceChangesBool() planmodifier.Bool {
	return useStateUnlessResourceChangesBool{}
}

type useStateUnlessResourceChangesBool struct{}

func (m useStateUnlessResourceChangesBool) Description(_ context.Context) string {
	return "Preserves the prior state value unless other resource attributes are changing."
}

func (m useStateUnlessResourceChangesBool) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateUnlessResourceChangesBool) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if !req.PlanValue.IsUnknown() {
		return
	}

	planAttrs := map[string]tftypes.Value{}
	if err := req.Plan.Raw.As(&planAttrs); err != nil {
		resp.PlanValue = req.StateValue
		return
	}

	stateAttrs := map[string]tftypes.Value{}
	if err := req.State.Raw.As(&stateAttrs); err != nil {
		resp.PlanValue = req.StateValue
		return
	}

	for key, planVal := range planAttrs {
		if !planVal.IsKnown() {
			continue
		}
		stateVal, exists := stateAttrs[key]
		if !exists {
			return
		}
		if !planVal.Equal(stateVal) {
			return
		}
	}

	resp.PlanValue = req.StateValue
}
