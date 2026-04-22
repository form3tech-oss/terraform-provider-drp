package drpv4

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func objectAttrString(o types.Object, key string) string {
	v := o.Attributes()[key]
	if v == nil || v.IsNull() || v.IsUnknown() {
		return ""
	}
	sv, ok := v.(types.String)
	if !ok {
		return ""
	}
	return sv.ValueString()
}

func diagListToStrings(ctx context.Context, l types.List, diags *diag.Diagnostics) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var out []string
	diags.Append(l.ElementsAs(ctx, &out, false)...)
	return out
}

func diagListToStringSlice(ctx context.Context, l types.List) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if l.IsNull() || l.IsUnknown() {
		return nil, diags
	}
	var out []string
	diags.Append(l.ElementsAs(ctx, &out, false)...)
	return out, diags
}
