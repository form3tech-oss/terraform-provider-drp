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

var _ resource.Resource = (*stageResource)(nil)
var _ resource.ResourceWithImportState = (*stageResource)(nil)

type stageResource struct {
	client *Config
}

func NewStageResource() resource.Resource {
	return &stageResource{}
}

func (r *stageResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_stage"
}

func stageTemplateNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name":        schema.StringAttribute{Required: true, Description: "Template name."},
		"contents":    schema.StringAttribute{Optional: true, Description: "Template content."},
		"path":        schema.StringAttribute{Optional: true, Description: "Template path."},
		"template_id": schema.StringAttribute{Optional: true, Description: "Template ID."},
		"link":        schema.StringAttribute{Optional: true, Description: "Template link."},
		"meta": schema.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "Template meta.",
		},
	}
}

func (r *stageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Stage name.",
				MarkdownDescription: "Stage name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description":   schema.StringAttribute{Optional: true, Description: "Stage description."},
			"documentation": schema.StringAttribute{Optional: true, Description: "Stage documentation."},
			"bootenv":       schema.StringAttribute{Optional: true, Description: "Bootenv name."},
			"optional_params": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Optional params.",
			},
			"params": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Stage params (string values).",
			},
			"profiles": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Profiles.",
			},
			"reboot": schema.BoolAttribute{Optional: true, Description: "Reboot flag."},
			"required_params": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Required params.",
			},
			"runner_wait": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Runner wait (from DRP).",
				MarkdownDescription: "Runner wait (from DRP).",
			},
			"tasks": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Tasks.",
			},
			"template": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: stageTemplateNestedAttributes()},
				Description:  "Stage templates.",
			},
		},
	}
}

func (r *stageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type stageResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Documentation  types.String `tfsdk:"documentation"`
	BootEnv        types.String `tfsdk:"bootenv"`
	OptionalParams types.List   `tfsdk:"optional_params"`
	Params         types.Map    `tfsdk:"params"`
	Profiles       types.List   `tfsdk:"profiles"`
	Reboot         types.Bool   `tfsdk:"reboot"`
	RequiredParams types.List   `tfsdk:"required_params"`
	RunnerWait     types.Bool   `tfsdk:"runner_wait"`
	Tasks          types.List   `tfsdk:"tasks"`
	Template       types.List   `tfsdk:"template"`
}

func (r *stageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func stageTemplateObjType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":        types.StringType,
		"contents":    types.StringType,
		"path":        types.StringType,
		"template_id": types.StringType,
		"link":        types.StringType,
		"meta":        types.MapType{ElemType: types.StringType},
	}}
}

func (r *stageResource) expandStageTemplates(ctx context.Context, l types.List, diags *diag.Diagnostics) []models.TemplateInfo {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	els := l.Elements()
	out := make([]models.TemplateInfo, 0, len(els))
	for _, el := range els {
		o, ok := el.(types.Object)
		if !ok {
			diags.AddError("Invalid template element", "expected object")
			return nil
		}
		var meta map[string]string
		mv := o.Attributes()["meta"]
		if mv != nil && !mv.IsNull() {
			if mmap, ok := mv.(types.Map); ok {
				diags.Append(mmap.ElementsAs(ctx, &meta, false)...)
			}
		}
		out = append(out, models.TemplateInfo{
			Name:     objectAttrString(o, "name"),
			Contents: objectAttrString(o, "contents"),
			Path:     objectAttrString(o, "path"),
			ID:       objectAttrString(o, "template_id"),
			Link:     objectAttrString(o, "link"),
			Meta:     meta,
		})
	}
	return out
}

func (r *stageResource) expandStage(ctx context.Context, m *stageResourceModel, diags *diag.Diagnostics) *models.Stage {
	reboot := false
	if !m.Reboot.IsNull() && !m.Reboot.IsUnknown() {
		reboot = m.Reboot.ValueBool()
	}
	runner := false
	if !m.RunnerWait.IsNull() && !m.RunnerWait.IsUnknown() {
		runner = m.RunnerWait.ValueBool()
	}
	var params map[string]interface{}
	if !m.Params.IsNull() && !m.Params.IsUnknown() {
		var sm map[string]string
		diags.Append(m.Params.ElementsAs(ctx, &sm, false)...)
		if diags.HasError() {
			return nil
		}
		params = stringMapToInterfaceMap(sm)
	}
	return &models.Stage{
		Name:    m.Name.ValueString(),
		DocData: newDocData(m.Description.ValueString(), m.Documentation.ValueString()),
		ParamData: models.ParamData{
			Params: params,
		},
		ProfileData: models.ProfileData{
			Profiles: diagListToStrings(ctx, m.Profiles, diags),
		},
		BootEnv:        m.BootEnv.ValueString(),
		OptionalParams: diagListToStrings(ctx, m.OptionalParams, diags),
		Reboot:         reboot,
		RequiredParams: diagListToStrings(ctx, m.RequiredParams, diags),
		RunnerWait:     runner,
		Tasks:          diagListToStrings(ctx, m.Tasks, diags),
		Templates:      r.expandStageTemplates(ctx, m.Template, diags),
	}
}

func (r *stageResource) flattenStageTemplatesMerged(ctx context.Context, prior types.List, api []models.TemplateInfo, diags *diag.Diagnostics) types.List {
	if len(api) == 0 {
		if prior.IsNull() || prior.IsUnknown() {
			return types.ListNull(stageTemplateObjType())
		}
		return types.ListNull(stageTemplateObjType())
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
		pObj := types.ObjectNull(stageTemplateObjType().AttrTypes)
		if i < len(priorObjs) {
			pObj = priorObjs[i]
		}
		meta := ti.Meta
		if meta == nil {
			meta = map[string]string{}
		}
		metaVal := mergeOptStringMap(ctx, priorObjMap(pObj, "meta"), meta, diags)
		attrs := map[string]attr.Value{
			"name":        types.StringValue(ti.Name),
			"contents":    mergeOptString(priorObjString(pObj, "contents"), ti.Contents),
			"path":        mergeOptString(priorObjString(pObj, "path"), ti.Path),
			"template_id": mergeOptString(priorObjString(pObj, "template_id"), ti.ID),
			"link":        mergeOptString(priorObjString(pObj, "link"), ti.Link),
			"meta":        metaVal,
		}
		obj, d := types.ObjectValue(stageTemplateObjType().AttrTypes, attrs)
		diags.Append(d...)
		tplObjs = append(tplObjs, obj)
	}
	elems := make([]attr.Value, len(tplObjs))
	for i, o := range tplObjs {
		elems[i] = o
	}
	return types.ListValueMust(stageTemplateObjType(), elems)
}

func (r *stageResource) flattenStage(ctx context.Context, s *models.Stage, m *stageResourceModel, diags *diag.Diagnostics) {
	m.Name = types.StringValue(s.Name)
	m.Description = mergeOptString(m.Description, s.Description)
	m.Documentation = mergeOptString(m.Documentation, s.Documentation)
	m.BootEnv = mergeOptString(m.BootEnv, s.BootEnv)
	m.OptionalParams = mergeOptStringList(ctx, m.OptionalParams, s.OptionalParams, diags)
	m.Profiles = mergeOptStringList(ctx, m.Profiles, s.Profiles, diags)
	m.Reboot = mergeOptBool(m.Reboot, s.Reboot)
	m.RequiredParams = mergeOptStringList(ctx, m.RequiredParams, s.RequiredParams, diags)
	m.RunnerWait = mergeOptBool(m.RunnerWait, s.RunnerWait)
	m.Tasks = mergeOptStringList(ctx, m.Tasks, s.Tasks, diags)

	var sm map[string]string
	if s.Params != nil {
		var err error
		sm, err = interfaceMapToStringMap(s.Params)
		if err != nil {
			diags.AddError("Invalid stage params", err.Error())
			return
		}
	}
	m.Params = mergeOptStringMap(ctx, m.Params, sm, diags)

	m.Template = r.flattenStageTemplatesMerged(ctx, m.Template, s.Templates, diags)
}

func (r *stageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan stageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stage := r.expandStage(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.CreateModel(stage); err != nil {
		resp.Diagnostics.AddError("Create stage failed", err.Error())
		return
	}
	res, err := r.client.session.GetModel("stages", stage.Name)
	if err != nil {
		resp.Diagnostics.AddError("Read stage after create failed", err.Error())
		return
	}
	r.flattenStage(ctx, res.(*models.Stage), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state stageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.session.GetModel("stages", state.Name.ValueString())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read stage failed", err.Error())
		return
	}
	r.flattenStage(ctx, res.(*models.Stage), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *stageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan stageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	stage := r.expandStage(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.PutModel(stage); err != nil {
		resp.Diagnostics.AddError("Update stage failed", err.Error())
		return
	}
	res, err := r.client.session.GetModel("stages", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read stage after update failed", err.Error())
		return
	}
	r.flattenStage(ctx, res.(*models.Stage), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state stageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.session.DeleteModel("stages", state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete stage failed", err.Error())
	}
}
