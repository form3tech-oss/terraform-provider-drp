package drpv4

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.com/rackn/provision/v4/models"
)

var _ resource.Resource = (*poolResource)(nil)
var _ resource.ResourceWithImportState = (*poolResource)(nil)

type poolResource struct {
	client *Config
}

func NewPoolResource() resource.Resource {
	return &poolResource{}
}

func (r *poolResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_pool"
}

func poolActionNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"add_parameters": schema.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "Parameters to add (values are strings; typed per param definition).",
		},
		"add_profiles": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"remove_parameters": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"remove_profiles": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"workflow": schema.StringAttribute{Optional: true},
	}
}

func (r *poolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	actionObj := schema.NestedAttributeObject{Attributes: poolActionNestedAttributes()}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"pool_id": schema.StringAttribute{
				Required:            true,
				Description:         "Pool id.",
				MarkdownDescription: "Pool id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description":   schema.StringAttribute{Optional: true},
			"documentation": schema.StringAttribute{Optional: true},
			"parent_pool":   schema.StringAttribute{Optional: true},
			"allocate_actions": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: actionObj,
			},
			"release_actions": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: actionObj,
			},
			"enter_actions": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: actionObj,
			},
			"exit_actions": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: actionObj,
			},
			"autofill": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"acquire_pool": schema.StringAttribute{Optional: true},
						"create_parameters": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						"max_free":     schema.Int64Attribute{Optional: true},
						"min_free":     schema.Int64Attribute{Optional: true},
						"return_pool":  schema.StringAttribute{Optional: true},
						"use_autofill": schema.BoolAttribute{Optional: true},
					},
				},
			},
		},
	}
}

func (r *poolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type poolResourceModel struct {
	PoolID          types.String `tfsdk:"pool_id"`
	Description     types.String `tfsdk:"description"`
	Documentation   types.String `tfsdk:"documentation"`
	ParentPool      types.String `tfsdk:"parent_pool"`
	AllocateActions types.List   `tfsdk:"allocate_actions"`
	ReleaseActions  types.List   `tfsdk:"release_actions"`
	EnterActions    types.List   `tfsdk:"enter_actions"`
	ExitActions     types.List   `tfsdk:"exit_actions"`
	Autofill        types.List   `tfsdk:"autofill"`
}

func poolActionObjType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"add_parameters":    types.MapType{ElemType: types.StringType},
		"add_profiles":      types.ListType{ElemType: types.StringType},
		"remove_parameters": types.ListType{ElemType: types.StringType},
		"remove_profiles":   types.ListType{ElemType: types.StringType},
		"workflow":          types.StringType,
	}}
}

func autofillObjType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"acquire_pool":      types.StringType,
		"create_parameters": types.MapType{ElemType: types.StringType},
		"max_free":          types.Int64Type,
		"min_free":          types.Int64Type,
		"return_pool":       types.StringType,
		"use_autofill":      types.BoolType,
	}}
}

func (r *poolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("pool_id"), req, resp)
}

func (r *poolResource) expandPoolActions(ctx context.Context, l types.List, diags *diag.Diagnostics) *models.PoolTransitionActions {
	if l.IsNull() || l.IsUnknown() || len(l.Elements()) == 0 {
		return nil
	}
	o, ok := l.Elements()[0].(types.Object)
	if !ok {
		diags.AddError("Invalid pool actions", "expected object")
		return nil
	}
	attrs := o.Attributes()
	workflow := ""
	if w := attrs["workflow"]; w != nil && !w.IsNull() {
		if sv, ok := w.(types.String); ok {
			workflow = sv.ValueString()
		}
	}
	action := &models.PoolTransitionActions{
		AddParameters:    map[string]interface{}{},
		AddProfiles:      objectListStrings(ctx, attrs["add_profiles"], diags),
		RemoveParameters: objectListStrings(ctx, attrs["remove_parameters"], diags),
		RemoveProfiles:   objectListStrings(ctx, attrs["remove_profiles"], diags),
		Workflow:         workflow,
	}
	if diags.HasError() {
		return nil
	}
	ap := attrs["add_parameters"]
	if ap != nil && !ap.IsNull() {
		mmap, ok := ap.(types.Map)
		if ok {
			var sm map[string]string
			diags.Append(mmap.ElementsAs(ctx, &sm, false)...)
			for k, v := range sm {
				val, err := convertParamToType(r.client, k, v)
				if err != nil {
					diags.AddError("Invalid add_parameters value", err.Error())
					return nil
				}
				action.AddParameters[k] = val
			}
		}
	}
	return action
}

func objectListStrings(ctx context.Context, v attr.Value, diags *diag.Diagnostics) []string {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return nil
	}
	lv, ok := v.(types.List)
	if !ok {
		return nil
	}
	return diagListToStrings(ctx, lv, diags)
}

func (r *poolResource) expandPoolAutofill(ctx context.Context, l types.List, diags *diag.Diagnostics) *models.PoolAutoFill {
	if l.IsNull() || l.IsUnknown() || len(l.Elements()) == 0 {
		return nil
	}
	o, ok := l.Elements()[0].(types.Object)
	if !ok {
		diags.AddError("Invalid autofill", "expected object")
		return nil
	}
	attrs := o.Attributes()
	af := &models.PoolAutoFill{}
	if v := attrs["acquire_pool"]; v != nil && !v.IsNull() {
		if sv, ok := v.(types.String); ok {
			af.AcquirePool = sv.ValueString()
		}
	}
	if v := attrs["return_pool"]; v != nil && !v.IsNull() {
		if sv, ok := v.(types.String); ok {
			af.ReturnPool = sv.ValueString()
		}
	}
	if v := attrs["max_free"]; v != nil && !v.IsNull() {
		if iv, ok := v.(types.Int64); ok {
			af.MaxFree = int32(iv.ValueInt64())
		}
	}
	if v := attrs["min_free"]; v != nil && !v.IsNull() {
		if iv, ok := v.(types.Int64); ok {
			af.MinFree = int32(iv.ValueInt64())
		}
	}
	if v := attrs["use_autofill"]; v != nil && !v.IsNull() {
		if bv, ok := v.(types.Bool); ok {
			af.UseAutoFill = bv.ValueBool()
		}
	}
	cp := attrs["create_parameters"]
	if cp != nil && !cp.IsNull() {
		if mmap, ok := cp.(types.Map); ok {
			var sm map[string]string
			diags.Append(mmap.ElementsAs(ctx, &sm, false)...)
			af.CreateParameters = stringMapToInterfaceMap(sm)
		}
	}
	return af
}

func (r *poolResource) expandPool(ctx context.Context, m *poolResourceModel, diags *diag.Diagnostics) *models.Pool {
	return &models.Pool{
		Id:              m.PoolID.ValueString(),
		Description:     m.Description.ValueString(),
		Documentation:   m.Documentation.ValueString(),
		ParentPool:      m.ParentPool.ValueString(),
		AllocateActions: r.expandPoolActions(ctx, m.AllocateActions, diags),
		ReleaseActions:  r.expandPoolActions(ctx, m.ReleaseActions, diags),
		EnterActions:    r.expandPoolActions(ctx, m.EnterActions, diags),
		ExitActions:     r.expandPoolActions(ctx, m.ExitActions, diags),
		AutoFill:        r.expandPoolAutofill(ctx, m.Autofill, diags),
	}
}

func (r *poolResource) flattenPoolActions(ctx context.Context, actions *models.PoolTransitionActions, diags *diag.Diagnostics) types.List {
	if actions == nil {
		return types.ListNull(poolActionObjType())
	}
	addParams := map[string]string{}
	for k, v := range actions.AddParameters {
		s, err := convertParamToString(v)
		if err != nil {
			diags.AddError("Flatten pool actions", err.Error())
			return types.ListNull(poolActionObjType())
		}
		addParams[k] = s
	}
	apMap, d := types.MapValueFrom(ctx, types.StringType, addParams)
	diags.Append(d...)
	apl, d := types.ListValueFrom(ctx, types.StringType, actions.AddProfiles)
	diags.Append(d...)
	rpl, d := types.ListValueFrom(ctx, types.StringType, actions.RemoveParameters)
	diags.Append(d...)
	rpr, d := types.ListValueFrom(ctx, types.StringType, actions.RemoveProfiles)
	diags.Append(d...)
	attrs := map[string]attr.Value{
		"add_parameters":    apMap,
		"add_profiles":      apl,
		"remove_parameters": rpl,
		"remove_profiles":   rpr,
		"workflow":          types.StringValue(actions.Workflow),
	}
	obj, d := types.ObjectValue(poolActionObjType().AttrTypes, attrs)
	diags.Append(d...)
	return types.ListValueMust(poolActionObjType(), []attr.Value{obj})
}

func (r *poolResource) flattenPoolAutofill(ctx context.Context, af *models.PoolAutoFill, diags *diag.Diagnostics) types.List {
	if af == nil {
		return types.ListNull(autofillObjType())
	}
	var cp types.Map
	if af.CreateParameters != nil {
		sm := make(map[string]string)
		for k, v := range af.CreateParameters {
			s, err := convertParamToString(v)
			if err != nil {
				diags.AddError("Flatten autofill create_parameters", err.Error())
				return types.ListNull(autofillObjType())
			}
			sm[k] = s
		}
		mv, d := types.MapValueFrom(ctx, types.StringType, sm)
		diags.Append(d...)
		cp = mv
	} else {
		cp = types.MapNull(types.StringType)
	}
	attrs := map[string]attr.Value{
		"acquire_pool":      types.StringValue(af.AcquirePool),
		"create_parameters": cp,
		"max_free":          types.Int64Value(int64(af.MaxFree)),
		"min_free":          types.Int64Value(int64(af.MinFree)),
		"return_pool":       types.StringValue(af.ReturnPool),
		"use_autofill":      types.BoolValue(af.UseAutoFill),
	}
	obj, d := types.ObjectValue(autofillObjType().AttrTypes, attrs)
	diags.Append(d...)
	return types.ListValueMust(autofillObjType(), []attr.Value{obj})
}

func (r *poolResource) flattenPool(ctx context.Context, p *models.Pool, m *poolResourceModel, diags *diag.Diagnostics) {
	m.PoolID = types.StringValue(p.Id)
	m.Description = types.StringValue(p.Description)
	m.Documentation = types.StringValue(p.Documentation)
	m.ParentPool = types.StringValue(p.ParentPool)
	m.AllocateActions = r.flattenPoolActions(ctx, p.AllocateActions, diags)
	m.ReleaseActions = r.flattenPoolActions(ctx, p.ReleaseActions, diags)
	m.EnterActions = r.flattenPoolActions(ctx, p.EnterActions, diags)
	m.ExitActions = r.flattenPoolActions(ctx, p.ExitActions, diags)
	m.Autofill = r.flattenPoolAutofill(ctx, p.AutoFill, diags)
}

func (r *poolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan poolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool := r.expandPool(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.CreateModel(pool); err != nil {
		resp.Diagnostics.AddError("Create pool failed", err.Error())
		return
	}
	r.flattenPool(ctx, pool, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *poolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state poolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool, err := r.client.session.GetModel("pools", state.PoolID.ValueString())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read pool failed", err.Error())
		return
	}
	r.flattenPool(ctx, pool.(*models.Pool), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *poolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan poolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool := r.expandPool(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.PutModel(pool); err != nil {
		resp.Diagnostics.AddError("Update pool failed", err.Error())
		return
	}
	r.flattenPool(ctx, pool, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *poolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state poolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.session.DeleteModel("pools", state.PoolID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete pool failed", err.Error())
	}
}
