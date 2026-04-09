package ref

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Compile-time interface check.
var _ planmodifier.Int64 = useStateUnlessResourceChangesInt64{}

// UseStateUnlessResourceChangesInt64 is the Int64 variant of
// UseStateUnlessResourceChanges.  See that function's documentation for
// the full rationale.
//
// Use this for volatile Computed-only Int64 fields (e.g. utilization_update)
// that change on every WAPI update but should not cause a perpetual diff
// when nothing else in the resource changed.
func UseStateUnlessResourceChangesInt64() planmodifier.Int64 {
	return useStateUnlessResourceChangesInt64{}
}

type useStateUnlessResourceChangesInt64 struct{}

func (m useStateUnlessResourceChangesInt64) Description(_ context.Context) string {
	return "Preserves the prior state value unless other resource attributes are changing."
}

func (m useStateUnlessResourceChangesInt64) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateUnlessResourceChangesInt64) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
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
