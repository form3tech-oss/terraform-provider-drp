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

var _ resource.Resource = (*taskResource)(nil)
var _ resource.ResourceWithImportState = (*taskResource)(nil)

type taskResource struct {
	client *Config
}

func NewTaskResource() resource.Resource {
	return &taskResource{}
}

func (r *taskResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_task"
}

func taskTemplateNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"template_id": schema.StringAttribute{Optional: true, Description: "Template id."},
		"name":        schema.StringAttribute{Required: true, Description: "Template name."},
		"path":        schema.StringAttribute{Optional: true, Description: "Template path."},
		"contents":    schema.StringAttribute{Optional: true, Description: "Template contents."},
		"link":        schema.StringAttribute{Optional: true, Description: "Template link."},
		"meta": schema.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "Template meta (string map).",
		},
	}
}

func taskClaimNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"scope":    schema.StringAttribute{Optional: true, Description: "Claim scope."},
		"action":   schema.StringAttribute{Optional: true, Description: "Claim action."},
		"specific": schema.StringAttribute{Optional: true, Description: "Claim specific."},
	}
}

func (r *taskResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Task name.",
				MarkdownDescription: "Task name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "Task description.",
				MarkdownDescription: "Task description.",
			},
			"required_params": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Required params.",
			},
			"optional_params": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Optional params.",
			},
			"templates": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: taskTemplateNestedAttributes()},
				Description:  "Inline templates.",
			},
			"extra_claims": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: taskClaimNestedAttributes()},
				Description:  "Extra claims.",
			},
			"extra_roles": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Extra roles.",
			},
			"prerequisites": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Prerequisites.",
			},
		},
	}
}

func (r *taskResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type taskResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	RequiredParams types.List   `tfsdk:"required_params"`
	OptionalParams types.List   `tfsdk:"optional_params"`
	Templates      types.List   `tfsdk:"templates"`
	ExtraClaims    types.List   `tfsdk:"extra_claims"`
	ExtraRoles     types.List   `tfsdk:"extra_roles"`
	Prerequisites  types.List   `tfsdk:"prerequisites"`
}

func (r *taskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func templateInfoType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"template_id": types.StringType,
		"name":        types.StringType,
		"path":        types.StringType,
		"contents":    types.StringType,
		"link":        types.StringType,
		"meta":        types.MapType{ElemType: types.StringType},
	}}
}

func claimType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"scope":    types.StringType,
		"action":   types.StringType,
		"specific": types.StringType,
	}}
}

func (r *taskResource) expandTask(ctx context.Context, m *taskResourceModel, diags *diag.Diagnostics) *models.Task {
	task := models.Task{
		Name:           m.Name.ValueString(),
		Description:    m.Description.ValueString(),
		RequiredParams: diagListToStrings(ctx, m.RequiredParams, diags),
		OptionalParams: diagListToStrings(ctx, m.OptionalParams, diags),
		ExtraRoles:     diagListToStrings(ctx, m.ExtraRoles, diags),
		Prerequisites:  diagListToStrings(ctx, m.Prerequisites, diags),
	}
	if diags.HasError() {
		return nil
	}
	task.Templates = r.expandTaskTemplates(ctx, m.Templates, diags)
	task.ExtraClaims = r.expandClaims(ctx, m.ExtraClaims, diags)
	if diags.HasError() {
		return nil
	}
	return &task
}

func (r *taskResource) expandTaskTemplates(ctx context.Context, l types.List, diags *diag.Diagnostics) []models.TemplateInfo {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	els := l.Elements()
	if len(els) == 0 {
		return nil
	}
	out := make([]models.TemplateInfo, 0, len(els))
	for _, el := range els {
		o, ok := el.(types.Object)
		if !ok {
			diags.AddError("Invalid templates element", "expected object")
			return nil
		}
		var meta map[string]string
		mv := o.Attributes()["meta"]
		if mv != nil && !mv.IsNull() {
			mmap, ok := mv.(types.Map)
			if ok {
				diags.Append(mmap.ElementsAs(ctx, &meta, false)...)
			}
		}
		out = append(out, models.TemplateInfo{
			ID:       objectAttrString(o, "template_id"),
			Name:     objectAttrString(o, "name"),
			Path:     objectAttrString(o, "path"),
			Contents: objectAttrString(o, "contents"),
			Link:     objectAttrString(o, "link"),
			Meta:     meta,
		})
	}
	return out
}

func (r *taskResource) expandClaims(ctx context.Context, l types.List, diags *diag.Diagnostics) []*models.Claim {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	els := l.Elements()
	if len(els) == 0 {
		return nil
	}
	out := make([]*models.Claim, 0, len(els))
	for _, el := range els {
		o, ok := el.(types.Object)
		if !ok {
			diags.AddError("Invalid extra_claims element", "expected object")
			return nil
		}
		out = append(out, &models.Claim{
			Scope:    objectAttrString(o, "scope"),
			Action:   objectAttrString(o, "action"),
			Specific: objectAttrString(o, "specific"),
		})
	}
	return out
}

func (r *taskResource) flattenTaskTemplatesMerged(ctx context.Context, prior types.List, api []models.TemplateInfo, diags *diag.Diagnostics) types.List {
	if len(api) == 0 {
		if prior.IsNull() || prior.IsUnknown() {
			return types.ListNull(templateInfoType())
		}
		return types.ListNull(templateInfoType())
	}
	var priorObjs []types.Object
	if !prior.IsNull() && !prior.IsUnknown() {
		for _, el := range prior.Elements() {
			if o, ok := el.(types.Object); ok {
				priorObjs = append(priorObjs, o)
			}
		}
	}
	tplObjs := make([]types.Object, 0, len(api))
	for i, ti := range api {
		var pObj types.Object
		if i < len(priorObjs) {
			pObj = priorObjs[i]
		}
		meta := ti.Meta
		if meta == nil {
			meta = map[string]string{}
		}
		metaVal := mergeOptStringMap(ctx, priorObjMap(pObj, "meta"), meta, diags)
		attrs := map[string]attr.Value{
			"template_id": mergeOptString(priorObjString(pObj, "template_id"), ti.ID),
			"name":        types.StringValue(ti.Name),
			"path":        mergeOptString(priorObjString(pObj, "path"), ti.Path),
			"contents":    mergeOptString(priorObjString(pObj, "contents"), ti.Contents),
			"link":        mergeOptString(priorObjString(pObj, "link"), ti.Link),
			"meta":        metaVal,
		}
		obj, d := types.ObjectValue(templateInfoType().AttrTypes, attrs)
		diags.Append(d...)
		tplObjs = append(tplObjs, obj)
	}
	elems := make([]attr.Value, len(tplObjs))
	for i, o := range tplObjs {
		elems[i] = o
	}
	return types.ListValueMust(templateInfoType(), elems)
}

func (r *taskResource) flattenTaskClaimsMerged(ctx context.Context, prior types.List, api []*models.Claim, diags *diag.Diagnostics) types.List {
	if len(api) == 0 {
		if prior.IsNull() || prior.IsUnknown() {
			return types.ListNull(claimType())
		}
		return types.ListNull(claimType())
	}
	var priorObjs []types.Object
	if !prior.IsNull() && !prior.IsUnknown() {
		for _, el := range prior.Elements() {
			if o, ok := el.(types.Object); ok {
				priorObjs = append(priorObjs, o)
			}
		}
	}
	claimObjs := make([]types.Object, 0, len(api))
	for i, c := range api {
		var pObj types.Object
		if i < len(priorObjs) {
			pObj = priorObjs[i]
		}
		attrs := map[string]attr.Value{
			"scope":    mergeOptString(priorObjString(pObj, "scope"), c.Scope),
			"action":   mergeOptString(priorObjString(pObj, "action"), c.Action),
			"specific": mergeOptString(priorObjString(pObj, "specific"), c.Specific),
		}
		obj, d := types.ObjectValue(claimType().AttrTypes, attrs)
		diags.Append(d...)
		claimObjs = append(claimObjs, obj)
	}
	elems := make([]attr.Value, len(claimObjs))
	for i, o := range claimObjs {
		elems[i] = o
	}
	return types.ListValueMust(claimType(), elems)
}

func (r *taskResource) flattenTask(ctx context.Context, task *models.Task, m *taskResourceModel, diags *diag.Diagnostics) {
	m.Name = types.StringValue(task.Name)
	m.Description = mergeOptString(m.Description, task.Description)
	m.RequiredParams = mergeOptStringList(ctx, m.RequiredParams, task.RequiredParams, diags)
	m.OptionalParams = mergeOptStringList(ctx, m.OptionalParams, task.OptionalParams, diags)
	m.ExtraRoles = mergeOptStringList(ctx, m.ExtraRoles, task.ExtraRoles, diags)
	m.Prerequisites = mergeOptStringList(ctx, m.Prerequisites, task.Prerequisites, diags)
	m.Templates = r.flattenTaskTemplatesMerged(ctx, m.Templates, task.Templates, diags)
	m.ExtraClaims = r.flattenTaskClaimsMerged(ctx, m.ExtraClaims, task.ExtraClaims, diags)
}

func (r *taskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan taskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	task := r.expandTask(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.CreateModel(task); err != nil {
		resp.Diagnostics.AddError("Create task failed", err.Error())
		return
	}
	to, err := r.client.session.GetModel("tasks", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read task after create failed", err.Error())
		return
	}
	r.flattenTask(ctx, to.(*models.Task), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *taskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state taskResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	to, err := r.client.session.GetModel("tasks", state.Name.ValueString())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read task failed", err.Error())
		return
	}
	r.flattenTask(ctx, to.(*models.Task), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *taskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan taskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	task := r.expandTask(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.PutModel(task); err != nil {
		resp.Diagnostics.AddError("Update task failed", err.Error())
		return
	}
	to, err := r.client.session.GetModel("tasks", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read task after update failed", err.Error())
		return
	}
	r.flattenTask(ctx, to.(*models.Task), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *taskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state taskResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.session.DeleteModel("tasks", state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete task failed", err.Error())
	}
}
