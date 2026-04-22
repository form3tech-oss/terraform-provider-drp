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

var _ resource.Resource = (*workflowResource)(nil)
var _ resource.ResourceWithImportState = (*workflowResource)(nil)

type workflowResource struct {
	client *Config
}

func NewWorkflowResource() resource.Resource {
	return &workflowResource{}
}

func (r *workflowResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_workflow"
}

func (r *workflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Workflow name.",
				MarkdownDescription: "Workflow name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "Workflow description.",
				MarkdownDescription: "Workflow description.",
			},
			"documentation": schema.StringAttribute{
				Optional:            true,
				Description:         "Workflow documentation.",
				MarkdownDescription: "Workflow documentation.",
			},
			"stages": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Ordered list of stage names.",
				MarkdownDescription: "Ordered list of stage names.",
			},
		},
	}
}

func (r *workflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type workflowResourceModel struct {
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Documentation types.String `tfsdk:"documentation"`
	Stages        types.List   `tfsdk:"stages"`
}

func (r *workflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *workflowResource) expandWorkflow(ctx context.Context, m *workflowResourceModel, diags *diag.Diagnostics) *models.Workflow {
	stages := diagListToStrings(ctx, m.Stages, diags)
	return &models.Workflow{
		Name:    m.Name.ValueString(),
		DocData: newDocData(m.Description.ValueString(), m.Documentation.ValueString()),
		Stages:  stages,
	}
}

func (r *workflowResource) flattenWorkflow(ctx context.Context, wf *models.Workflow, m *workflowResourceModel, diags *diag.Diagnostics) {
	m.Name = types.StringValue(wf.Name)
	m.Description = mergeOptString(m.Description, wf.Description)
	m.Documentation = mergeOptString(m.Documentation, wf.Documentation)
	m.Stages = mergeOptStringList(ctx, m.Stages, wf.Stages, diags)
}

func (r *workflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan workflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	wf := r.expandWorkflow(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.CreateModel(wf); err != nil {
		resp.Diagnostics.AddError("Create workflow failed", err.Error())
		return
	}
	got, err := r.client.session.GetModel("workflows", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read workflow after create failed", err.Error())
		return
	}
	r.flattenWorkflow(ctx, got.(*models.Workflow), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state workflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.session.GetModel("workflows", state.Name.ValueString())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read workflow failed", err.Error())
		return
	}
	r.flattenWorkflow(ctx, res.(*models.Workflow), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan workflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	wf := r.expandWorkflow(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.PutModel(wf); err != nil {
		resp.Diagnostics.AddError("Update workflow failed", err.Error())
		return
	}
	got, err := r.client.session.GetModel("workflows", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read workflow after update failed", err.Error())
		return
	}
	r.flattenWorkflow(ctx, got.(*models.Workflow), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state workflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.session.DeleteModel("workflows", state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete workflow failed", err.Error())
	}
}
