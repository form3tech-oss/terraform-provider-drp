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

var _ resource.Resource = (*templateResource)(nil)
var _ resource.ResourceWithImportState = (*templateResource)(nil)

type templateResource struct {
	client *Config
}

type templatePutResult struct {
	Template  *models.Template
	Available bool
	Errors    []string
}

func NewTemplateResource() resource.Resource {
	return &templateResource{}
}

func (r *templateResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_template"
}

func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"template_id": schema.StringAttribute{
				Required:            true,
				Description:         "Template id.",
				MarkdownDescription: "Template id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "Template description.",
				MarkdownDescription: "Template description.",
			},
			"contents": schema.StringAttribute{
				Optional:            true,
				Description:         "Template contents.",
				MarkdownDescription: "Template contents.",
			},
			"start_delimiter": schema.StringAttribute{
				Optional:            true,
				Description:         "Template start delimiter.",
				MarkdownDescription: "Template start delimiter.",
			},
			"end_delimiter": schema.StringAttribute{
				Optional:            true,
				Description:         "Template end delimiter.",
				MarkdownDescription: "Template end delimiter.",
			},
		},
	}
}

func (r *templateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type templateResourceModel struct {
	TemplateID     types.String `tfsdk:"template_id"`
	Description    types.String `tfsdk:"description"`
	Contents       types.String `tfsdk:"contents"`
	StartDelimiter types.String `tfsdk:"start_delimiter"`
	EndDelimiter   types.String `tfsdk:"end_delimiter"`
}

func (r *templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("template_id"), req, resp)
}

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	template := models.Template{
		ID: plan.TemplateID.ValueString(),
		DescData: models.DescData{
			Description: plan.Description.ValueString(),
		},
		Contents:       plan.Contents.ValueString(),
		StartDelimiter: plan.StartDelimiter.ValueString(),
		EndDelimiter:   plan.EndDelimiter.ValueString(),
	}
	template.Validate()
	if template.Error() != "" {
		resp.Diagnostics.AddError("Template validation failed", template.Error())
		return
	}

	reqAPI := r.client.session.Req().Post(template).UrlFor("templates")
	if err := reqAPI.Do(&template); err != nil {
		resp.Diagnostics.AddError("Create template failed", err.Error())
		return
	}
	if template.Error() != "" {
		resp.Diagnostics.AddError("Create template failed", template.Error())
		return
	}

	r.flattenTemplate(ctx, &template, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	to, err := r.client.session.GetModel("templates", state.TemplateID.ValueString())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read template failed", err.Error())
		return
	}
	r.flattenTemplate(ctx, to.(*models.Template), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *templateResource) flattenTemplate(_ context.Context, t *models.Template, m *templateResourceModel, _ *diag.Diagnostics) {
	m.TemplateID = types.StringValue(t.ID)
	m.Description = types.StringValue(t.Description)
	m.Contents = types.StringValue(t.Contents)
	m.StartDelimiter = types.StringValue(t.StartDelimiter)
	m.EndDelimiter = types.StringValue(t.EndDelimiter)
}

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	template := models.Template{
		ID: plan.TemplateID.ValueString(),
		DescData: models.DescData{
			Description: plan.Description.ValueString(),
		},
		Contents:       plan.Contents.ValueString(),
		StartDelimiter: plan.StartDelimiter.ValueString(),
		EndDelimiter:   plan.EndDelimiter.ValueString(),
	}
	template.Validate()
	if template.Error() != "" {
		resp.Diagnostics.AddError("Template validation failed", template.Error())
		return
	}

	var putResult templatePutResult
	reqAPI := r.client.session.Req().Put(template).UrlFor("templates", plan.TemplateID.ValueString())
	if err := reqAPI.Do(&putResult); err != nil {
		resp.Diagnostics.AddError("Update template failed", err.Error())
		return
	}

	to, err := r.client.session.GetModel("templates", plan.TemplateID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read template after update failed", err.Error())
		return
	}
	r.flattenTemplate(ctx, to.(*models.Template), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.session.DeleteModel("templates", state.TemplateID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete template failed", err.Error())
	}
}
