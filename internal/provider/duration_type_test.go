package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDurationStringValueSemanticEquals(t *testing.T) {
	tests := []struct {
		name     string
		prior    string
		new      string
		expected bool
	}{
		{name: "identical", prior: "336h", new: "336h", expected: true},
		{name: "api canonical form", prior: "336h", new: "336h0m0s", expected: true},
		{name: "different units", prior: "90m", new: "1h30m0s", expected: true},
		{name: "both never", prior: "", new: "", expected: true},
		{name: "never vs zero", prior: "", new: "0s", expected: false},
		{name: "zero vs never", prior: "0s", new: "", expected: false},
		{name: "different durations", prior: "336h", new: "1h0m0s", expected: false},
		{name: "invalid prior", prior: "two weeks", new: "336h0m0s", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			equal, diags := NewDurationStringValue(tc.prior).StringSemanticEquals(context.Background(), NewDurationStringValue(tc.new))
			assert.False(t, diags.HasError())
			assert.Equal(t, tc.expected, equal)
		})
	}
}

func TestDurationStringValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		expectErr bool
	}{
		{name: "null", value: types.StringNull()},
		{name: "unknown", value: types.StringUnknown()},
		{name: "never", value: types.StringValue("")},
		{name: "zero", value: types.StringValue("0s")},
		{name: "hours", value: types.StringValue("336h")},
		{name: "compound", value: types.StringValue("1h30m")},
		{name: "not a duration", value: types.StringValue("14d"), expectErr: true},
		{name: "negative", value: types.StringValue("-1h"), expectErr: true},
		{name: "sub-second", value: types.StringValue("1500ms"), expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			durationStringValidator{}.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("ttl"),
				ConfigValue: tc.value,
			}, resp)
			assert.Equal(t, tc.expectErr, resp.Diagnostics.HasError())
		})
	}
}
