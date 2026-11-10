package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable                    = DurationStringType{}
	_ basetypes.StringValuableWithSemanticEquals = DurationStringValue{}
	_ validator.String                           = durationStringValidator{}
)

// DurationStringType is a string type holding a Go duration, where "" means unset.
type DurationStringType struct {
	basetypes.StringType
}

func (t DurationStringType) Equal(o attr.Type) bool {
	other, ok := o.(DurationStringType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t DurationStringType) String() string {
	return "DurationStringType"
}

func (t DurationStringType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return DurationStringValue{StringValue: in}, nil
}

func (t DurationStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	return DurationStringValue{StringValue: stringValue}, nil
}

func (t DurationStringType) ValueType(_ context.Context) attr.Value {
	return DurationStringValue{}
}

// DurationStringValue is a string value compared as a duration, so "336h" equals "336h0m0s".
type DurationStringValue struct {
	basetypes.StringValue
}

func NewDurationStringValue(value string) DurationStringValue {
	return DurationStringValue{StringValue: basetypes.NewStringValue(value)}
}

func (v DurationStringValue) Equal(o attr.Value) bool {
	other, ok := o.(DurationStringValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v DurationStringValue) Type(_ context.Context) attr.Type {
	return DurationStringType{}
}

func (v DurationStringValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(DurationStringValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("Expected value type %T but got %T. Please report this issue to the provider developers.", v, newValuable),
		)
		return false, diags
	}

	return durationStringsEqual(v.ValueString(), newValue.ValueString()), diags
}

// durationStringsEqual compares two duration strings; "" (never) only equals "".
func durationStringsEqual(a, b string) bool {
	if a == "" || b == "" {
		return a == b
	}

	da, err := time.ParseDuration(a)
	if err != nil {
		return false
	}
	db, err := time.ParseDuration(b)
	if err != nil {
		return false
	}

	return da == db
}

type durationStringValidator struct{}

func (v durationStringValidator) Description(_ context.Context) string {
	return `value must be "" or a non-negative Go duration in whole seconds`
}

func (v durationStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v durationStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Duration", fmt.Sprintf("%q is not a valid Go duration (e.g. \"336h\"): %s", value, err))
		return
	}

	if d < 0 || d%time.Second != 0 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Duration", fmt.Sprintf("%q must be a non-negative duration in whole seconds", value))
	}
}
