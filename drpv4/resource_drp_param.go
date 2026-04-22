package drpv4

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.com/rackn/provision/v4/models"
)

var _ resource.Resource = (*paramResource)(nil)
var _ resource.ResourceWithImportState = (*paramResource)(nil)

type paramResource struct {
	client *Config
}

func NewParamResource() resource.Resource {
	return &paramResource{}
}

func (r *paramResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_param"
}

func (r *paramResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Param name.",
				MarkdownDescription: "Param name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "Param description.",
				MarkdownDescription: "Param description.",
			},
			"documentation": schema.StringAttribute{
				Optional:            true,
				Description:         "Param documentation.",
				MarkdownDescription: "Param documentation.",
			},
			"schema": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Param schema (defaults to type string).",
				MarkdownDescription: "Param schema (defaults to type string).",
			},
			"secure": schema.BoolAttribute{
				Optional:            true,
				Description:         "Whether the param is secure.",
				MarkdownDescription: "Whether the param is secure.",
			},
		},
	}
}

func (r *paramResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type paramResourceModel struct {
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Documentation types.String `tfsdk:"documentation"`
	Schema        types.Map    `tfsdk:"schema"`
	Secure        types.Bool   `tfsdk:"secure"`
}

func (r *paramResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *paramResource) expandParam(ctx context.Context, m *paramResourceModel, diags *diag.Diagnostics) *models.Param {
	schemaVal := defaultParamSchema()
	if !m.Schema.IsNull() && !m.Schema.IsUnknown() {
		var sm map[string]string
		diags.Append(m.Schema.ElementsAs(ctx, &sm, false)...)
		if diags.HasError() {
			return nil
		}
		schemaVal = stringMapToInterfaceMap(sm)
	}
	secure := false
	if !m.Secure.IsNull() && !m.Secure.IsUnknown() {
		secure = m.Secure.ValueBool()
	}
	return &models.Param{
		Name:          m.Name.ValueString(),
		Description:   m.Description.ValueString(),
		Documentation: m.Documentation.ValueString(),
		Schema:        schemaVal,
		Secure:        secure,
	}
}

func (r *paramResource) flattenParam(ctx context.Context, p *models.Param, m *paramResourceModel, diags *diag.Diagnostics) {
	m.Name = types.StringValue(p.Name)
	m.Description = mergeOptString(m.Description, p.Description)
	m.Documentation = mergeOptString(m.Documentation, p.Documentation)
	var sm map[string]string
	var err error
	if p.Schema != nil {
		if raw, ok := p.Schema.(map[string]interface{}); ok {
			sm, err = interfaceMapToStringMap(raw)
			if err != nil {
				diags.AddError("Invalid param schema", err.Error())
				return
			}
		}
	}
	if sm == nil {
		sm = map[string]string{}
	}
	if m.Schema.IsNull() || m.Schema.IsUnknown() {
		m.Schema = types.MapNull(types.StringType)
	} else {
		m.Schema = mergeOptStringMap(ctx, m.Schema, sm, diags)
	}
	m.Secure = mergeOptBool(m.Secure, p.Secure)
}

func (r *paramResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan paramResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	param := r.expandParam(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.CreateModel(param); err != nil {
		resp.Diagnostics.AddError("Create param failed", err.Error())
		return
	}
	got, err := r.client.session.GetModel("params", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read param after create failed", err.Error())
		return
	}
	r.flattenParam(ctx, got.(*models.Param), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *paramResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state paramResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	po, err := r.client.session.GetModel("params", state.Name.ValueString())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read param failed", err.Error())
		return
	}
	r.flattenParam(ctx, po.(*models.Param), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *paramResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan paramResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	param := r.expandParam(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.PutModel(param); err != nil {
		resp.Diagnostics.AddError("Update param failed", err.Error())
		return
	}
	got, err := r.client.session.GetModel("params", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read param after update failed", err.Error())
		return
	}
	r.flattenParam(ctx, got.(*models.Param), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *paramResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state paramResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.session.DeleteModel("params", state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete param failed", err.Error())
	}
}
