package ref

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// isUnknown returns true if the value is an unknown tftypes value.
// After tftypes.Value.As(&map[string]interface{}), unknown values are
// represented as tftypes.UnknownValue.
func isUnknown(v interface{}) bool {
	if v == nil {
		return false
	}
	// tftypes.Value.As() populates map values as tftypes.Value for complex
	// types and Go primitives for simple types.  An unknown value remains
	// as a tftypes.Value with IsKnown() == false.
	if tv, ok := v.(tftypes.Value); ok {
		return !tv.IsKnown()
	}
	return false
}

// deepEqual compares two values produced by tftypes.Value.As().
// Uses reflect.DeepEqual which handles nil, primitives, slices,
// and nested maps correctly.
func deepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
