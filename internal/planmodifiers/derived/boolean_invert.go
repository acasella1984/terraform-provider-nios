package derived

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.Bool = booleanInvertedFromModifier{}

// booleanInvertedFromModifier computes the planned value of a boolean field
// by inverting the value of its source field (e.g. disable_discovery = !enable_discovery).
//
// This replaces UseStateForUnknown for client-derivable boolean fields where
// the derived value is the logical negation of the source. UseStateForUnknown
// is incorrect for these fields because it preserves the old state value even
// when the source is being updated.
type booleanInvertedFromModifier struct {
	sourceAttribute string
}

func (m booleanInvertedFromModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Computes the boolean inverse of the planned value of %q", m.sourceAttribute)
}

func (m booleanInvertedFromModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m booleanInvertedFromModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// If the resource is being destroyed, do nothing.
	if req.Plan.Raw.IsNull() {
		return
	}

	// Read the source attribute's planned value.
	var sourceVal types.Bool
	diags := req.Plan.GetAttribute(ctx, req.Path.ParentPath().AtName(m.sourceAttribute), &sourceVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the source is unknown or null, we can't compute the derived value.
	if sourceVal.IsUnknown() || sourceVal.IsNull() {
		return
	}

	// Compute the inverse.
	resp.PlanValue = types.BoolValue(!sourceVal.ValueBool())
}

// BooleanInvertedFrom creates a plan modifier that computes the boolean
// inverse of a source attribute at plan time.
//
// For example, disable_discovery = !enable_discovery. This replicates the
// NIOS server's i2w_invert_boolean transformation, allowing Terraform to
// show the correct value in plan output rather than "(known after apply)".
func BooleanInvertedFrom(sourceAttr string) planmodifier.Bool {
	return booleanInvertedFromModifier{
		sourceAttribute: sourceAttr,
	}
}
