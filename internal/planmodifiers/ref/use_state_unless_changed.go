package ref

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Compile-time interface check.
var _ planmodifier.String = useStateUnlessResourceChanges{}

// UseStateUnlessResourceChanges returns a plan modifier that preserves the
// prior state value for a Computed attribute (like "ref") UNLESS other
// attributes in the resource have planned changes.
//
// Without this modifier a Computed-only field shows "(known after apply)" on
// every plan, which Terraform treats as a diff even when nothing meaningful
// changed.  With plain UseStateForUnknown the value is always preserved, but
// that causes "Provider produced inconsistent result after apply" when the
// server returns a new value (e.g. after a rename that changes the WAPI ref).
//
// This modifier solves both problems: it compares every *known* planned
// attribute against the corresponding state attribute.  If all match, the
// resource is unchanged and the state value is preserved.  If any differ,
// the field is left unknown so Terraform accepts whatever the server returns.
func UseStateUnlessResourceChanges() planmodifier.String {
	return useStateUnlessResourceChanges{}
}

type useStateUnlessResourceChanges struct{}

func (m useStateUnlessResourceChanges) Description(_ context.Context) string {
	return "Preserves the prior state value unless other resource attributes are changing."
}

func (m useStateUnlessResourceChanges) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateUnlessResourceChanges) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// On create there is no prior state — leave as unknown so Terraform
	// accepts whatever the server returns.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// If the plan already has a concrete value, don't override it.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Extract the plan and state as maps of tftypes.Value keyed by
	// attribute name.  The top-level object is always "known" even when
	// individual attributes inside are unknown.
	planAttrs := map[string]tftypes.Value{}
	if err := req.Plan.Raw.As(&planAttrs); err != nil {
		// Can't decode — safe fallback: preserve state value.
		resp.PlanValue = req.StateValue
		return
	}

	stateAttrs := map[string]tftypes.Value{}
	if err := req.State.Raw.As(&stateAttrs); err != nil {
		resp.PlanValue = req.StateValue
		return
	}

	// Walk planned attributes.  For each known value, compare against
	// state.  Skip unknown values (other Computed-only fields whose plan
	// value hasn't been resolved yet).  If any known value differs from
	// state, another attribute changed and this field might also change.
	for key, planVal := range planAttrs {
		if !planVal.IsKnown() {
			continue
		}
		stateVal, exists := stateAttrs[key]
		if !exists {
			// Attribute in plan not in state — something changed.
			return
		}
		if !planVal.Equal(stateVal) {
			// A known attribute differs from state — leave unknown.
			return
		}
	}

	// All known planned attributes match state — nothing changed.
	resp.PlanValue = req.StateValue
}
