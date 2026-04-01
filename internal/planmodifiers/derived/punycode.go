package derived

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/net/idna"
)

var _ planmodifier.String = punycodeDerivedFromModifier{}

// punycodeDerivedFromModifier computes the planned value of a punycode-derived
// field (e.g. dns_name) from the planned value of its source field (e.g. name).
//
// This replaces UseStateForUnknown for client-derivable fields where the
// derived value changes when the source changes. UseStateForUnknown is
// incorrect for these fields because it preserves the old state value even
// when the source is being updated, causing a "Provider produced inconsistent
// result after apply" error.
type punycodeDerivedFromModifier struct {
	sourceAttribute string
}

func (m punycodeDerivedFromModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Computes the punycode representation from the planned value of %q", m.sourceAttribute)
}

func (m punycodeDerivedFromModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m punycodeDerivedFromModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the resource is being destroyed, do nothing.
	if req.Plan.Raw.IsNull() {
		return
	}

	// Read the source attribute's planned value.
	var sourceVal types.String
	diags := req.Plan.GetAttribute(ctx, req.Path.ParentPath().AtName(m.sourceAttribute), &sourceVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the source is unknown (e.g. from a func_call), we can't compute
	// the derived value — leave it unknown so the server value takes effect.
	if sourceVal.IsUnknown() || sourceVal.IsNull() {
		return
	}

	// Compute the punycode representation.
	// The IDNA2008 profile used by NIOS applies ToASCII.
	ascii, err := idna.Lookup.ToASCII(sourceVal.ValueString())
	if err != nil {
		// If conversion fails, fall back to the raw value.
		// This handles cases like IP-based PTR names that aren't valid IDNA.
		resp.PlanValue = sourceVal
		return
	}

	resp.PlanValue = types.StringValue(ascii)
}

// PunycodeDerivedFrom returns a plan modifier that computes the planned value
// of a punycode-derived field from the specified source attribute.
//
// Use this for fields like dns_name which are the punycode (RFC 5891 IDNA)
// representation of another field (e.g. name).
//
// This modifier is safe to use because punycode encoding is a pure,
// deterministic, client-replicable transformation.
func PunycodeDerivedFrom(sourceAttribute string) planmodifier.String {
	return punycodeDerivedFromModifier{
		sourceAttribute: sourceAttribute,
	}
}
