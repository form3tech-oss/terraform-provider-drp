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
		Name:           m.Name.ValueString(),
		Description:    m.Description.ValueString(),
		Documentation:  m.Documentation.ValueString(),
		BootEnv:        m.BootEnv.ValueString(),
		OptionalParams: diagListToStrings(ctx, m.OptionalParams, diags),
		Params:         params,
		Profiles:       diagListToStrings(ctx, m.Profiles, diags),
		Reboot:         reboot,
		RequiredParams: diagListToStrings(ctx, m.RequiredParams, diags),
		RunnerWait:     runner,
		Tasks:          diagListToStrings(ctx, m.Tasks, diags),
		Templates:      r.expandStageTemplates(ctx, m.Template, diags),
	}
}

func (r *stageResource) flattenStage(ctx context.Context, s *models.Stage, m *stageResourceModel, diags *diag.Diagnostics) {
	m.Name = types.StringValue(s.Name)
	m.Description = types.StringValue(s.Description)
	m.Documentation = types.StringValue(s.Documentation)
	m.BootEnv = types.StringValue(s.BootEnv)
	m.OptionalParams = mustListStrings(ctx, s.OptionalParams, diags)
	m.Profiles = mustListStrings(ctx, s.Profiles, diags)
	m.Reboot = types.BoolValue(s.Reboot)
	m.RequiredParams = mustListStrings(ctx, s.RequiredParams, diags)
	m.RunnerWait = types.BoolValue(s.RunnerWait)
	m.Tasks = mustListStrings(ctx, s.Tasks, diags)

	if s.Params != nil {
		sm, err := interfaceMapToStringMap(s.Params)
		if err != nil {
			diags.AddError("Invalid stage params", err.Error())
			return
		}
		mv, d := types.MapValueFrom(ctx, types.StringType, sm)
		diags.Append(d...)
		m.Params = mv
	} else {
		m.Params = types.MapNull(types.StringType)
	}

	tplObjs := make([]types.Object, 0, len(s.Templates))
	for _, ti := range s.Templates {
		meta := ti.Meta
		if meta == nil {
			meta = map[string]string{}
		}
		metaVal, d := types.MapValueFrom(ctx, types.StringType, meta)
		diags.Append(d...)
		attrs := map[string]attr.Value{
			"name":        types.StringValue(ti.Name),
			"contents":    types.StringValue(ti.Contents),
			"path":        types.StringValue(ti.Path),
			"template_id": types.StringValue(ti.ID),
			"link":        types.StringValue(ti.Link),
			"meta":        metaVal,
		}
		obj, d := types.ObjectValue(stageTemplateObjType().AttrTypes, attrs)
		diags.Append(d...)
		tplObjs = append(tplObjs, obj)
	}
	if len(tplObjs) == 0 {
		m.Template = types.ListNull(stageTemplateObjType())
	} else {
		elems := make([]attr.Value, len(tplObjs))
		for i, o := range tplObjs {
			elems[i] = o
		}
		m.Template = types.ListValueMust(stageTemplateObjType(), elems)
	}
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
