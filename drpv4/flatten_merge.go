package drpv4

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func priorObjString(o types.Object, key string) types.String {
	if o.IsNull() {
		return types.StringNull()
	}
	a := o.Attributes()[key]
	if v, ok := a.(types.String); ok {
		return v
	}
	return types.StringNull()
}

func priorObjMap(o types.Object, key string) types.Map {
	if o.IsNull() {
		return types.MapNull(types.StringType)
	}
	a := o.Attributes()[key]
	if v, ok := a.(types.Map); ok {
		return v
	}
	return types.MapNull(types.StringType)
}

func priorObjBool(o types.Object, key string) types.Bool {
	if o.IsNull() {
		return types.BoolNull()
	}
	a := o.Attributes()[key]
	if v, ok := a.(types.Bool); ok {
		return v
	}
	return types.BoolNull()
}

func priorObjInt64(o types.Object, key string) types.Int64 {
	if o.IsNull() {
		return types.Int64Null()
	}
	a := o.Attributes()[key]
	if v, ok := a.(types.Int64); ok {
		return v
	}
	return types.Int64Null()
}

func priorObjList(o types.Object, key string) types.List {
	if o.IsNull() {
		return types.ListNull(types.StringType)
	}
	a := o.Attributes()[key]
	if v, ok := a.(types.List); ok {
		return v
	}
	return types.ListNull(types.StringType)
}

// mergeOptString keeps null when the practitioner omitted an optional string (prior null)
// and the API returns an empty string, so post-apply refresh matches the planned value.
func mergeOptString(prior types.String, api string) types.String {
	if api != "" {
		return types.StringValue(api)
	}
	if prior.IsNull() || prior.IsUnknown() {
		return types.StringNull()
	}
	return types.StringValue("")
}

func mergeOptInt64(prior types.Int64, api int64) types.Int64 {
	if api != 0 {
		return types.Int64Value(api)
	}
	if prior.IsNull() || prior.IsUnknown() {
		return types.Int64Null()
	}
	return types.Int64Value(0)
}

func mergeOptBool(prior types.Bool, api bool) types.Bool {
	if api {
		return types.BoolValue(true)
	}
	if prior.IsNull() || prior.IsUnknown() {
		return types.BoolNull()
	}
	return types.BoolValue(false)
}

func mergeOptStringList(ctx context.Context, prior types.List, api []string, diags *diag.Diagnostics) types.List {
	if prior.IsNull() || prior.IsUnknown() {
		return types.ListNull(types.StringType)
	}
	if len(api) > 0 {
		v, d := types.ListValueFrom(ctx, types.StringType, api)
		diags.Append(d...)
		return v
	}
	v, d := types.ListValueFrom(ctx, types.StringType, []string{})
	diags.Append(d...)
	return v
}

func mergeOptStringMap(ctx context.Context, prior types.Map, api map[string]string, diags *diag.Diagnostics) types.Map {
	// Match mergeOptStringList: if the practitioner omitted the map (null/unknown), keep null
	// even when the API returns a populated map, so post-apply state matches the plan.
	if prior.IsNull() || prior.IsUnknown() {
		return types.MapNull(types.StringType)
	}
	if len(api) > 0 {
		mv, d := types.MapValueFrom(ctx, types.StringType, api)
		diags.Append(d...)
		return mv
	}
	mv, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
	diags.Append(d...)
	return mv
}
