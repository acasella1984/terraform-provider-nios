package derived_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/infobloxopen/terraform-provider-nios/internal/planmodifiers/derived"
)

// testSchema is the minimal schema used by tests.
var testSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name":     schema.StringAttribute{Required: true},
		"dns_name": schema.StringAttribute{Computed: true},
	},
}

func TestPunycodeDerivedFrom_ASCIIDomain(t *testing.T) {
	resp := runPunycodePlanModify(t, "a-record1.example.com", "old.example.com")
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %s", resp.Diagnostics.Errors())
	}
	if resp.PlanValue.ValueString() != "a-record1.example.com" {
		t.Errorf("expected a-record1.example.com, got %s", resp.PlanValue.ValueString())
	}
}

func TestPunycodeDerivedFrom_UnicodeDomain(t *testing.T) {
	resp := runPunycodePlanModify(t, "münchen.example.com", "old.example.com")
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %s", resp.Diagnostics.Errors())
	}
	expected := "xn--mnchen-3ya.example.com"
	if resp.PlanValue.ValueString() != expected {
		t.Errorf("expected %s, got %s", expected, resp.PlanValue.ValueString())
	}
}

func TestPunycodeDerivedFrom_NameChange(t *testing.T) {
	// THE KEY TEST: When `name` changes from a-record1 to a-record2,
	// dns_name should update to the new punycode value, NOT keep the old state.
	// This is the exact scenario that caused the "Provider produced inconsistent
	// result after apply" error when UseStateForUnknown was used.
	resp := runPunycodePlanModify(t, "a-record2.example.com", "a-record1.example.com")
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %s", resp.Diagnostics.Errors())
	}
	if resp.PlanValue.ValueString() != "a-record2.example.com" {
		t.Errorf("expected a-record2.example.com, got %s", resp.PlanValue.ValueString())
	}
}

func TestPunycodeDerivedFrom_SourceUnknown(t *testing.T) {
	// When the source is unknown (e.g. from func_call), dns_name stays unknown.
	mod := derived.PunycodeDerivedFrom("name")

	fwSchema := testSchema

	planVal := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name":     tftypes.String,
				"dns_name": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"name":     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"dns_name": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
	)

	req := planmodifier.StringRequest{
		Plan:       tfsdk.Plan{Raw: planVal, Schema: fwSchema},
		PlanValue:  types.StringUnknown(),
		StateValue: types.StringValue("old.example.com"),
		Path:       path.Root("dns_name"),
	}

	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	mod.PlanModifyString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %s", resp.Diagnostics.Errors())
	}
	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected unknown, got %s", resp.PlanValue.ValueString())
	}
}

// runPunycodePlanModify creates a plan with the given source name and runs the modifier.
func runPunycodePlanModify(t *testing.T, sourceName, oldDnsName string) *planmodifier.StringResponse {
	t.Helper()

	mod := derived.PunycodeDerivedFrom("name")

	fwSchema := testSchema

	planVal := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name":     tftypes.String,
				"dns_name": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"name":     tftypes.NewValue(tftypes.String, sourceName),
			"dns_name": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
	)

	req := planmodifier.StringRequest{
		Plan:       tfsdk.Plan{Raw: planVal, Schema: fwSchema},
		PlanValue:  types.StringUnknown(),
		StateValue: types.StringValue(oldDnsName),
		Path:       path.Root("dns_name"),
	}

	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	mod.PlanModifyString(context.Background(), req, resp)
	return resp
}
