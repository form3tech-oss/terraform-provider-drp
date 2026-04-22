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

var _ resource.Resource = (*profileResource)(nil)
var _ resource.ResourceWithImportState = (*profileResource)(nil)

type profileResource struct {
	client *Config
}

func NewProfileResource() resource.Resource {
	return &profileResource{}
}

func (r *profileResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_profile"
}

func (r *profileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Profile name.",
				MarkdownDescription: "Profile name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "Profile description.",
				MarkdownDescription: "Profile description.",
			},
			"meta": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Profile metadata (arbitrary string key/value pairs, e.g. UX-related flags).",
				MarkdownDescription: "Profile metadata (arbitrary string key/value pairs, e.g. UX-related flags).",
			},
		},
	}
}

func (r *profileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type profileResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Meta        types.Map    `tfsdk:"meta"`
}

func (r *profileResource) expandProfile(ctx context.Context, m *profileResourceModel, diags *diag.Diagnostics) *models.Profile {
	profile := &models.Profile{
		Name:    m.Name.ValueString(),
		DocData: newDocData(m.Description.ValueString(), ""),
	}
	if m.Meta.IsNull() || m.Meta.IsUnknown() {
		return profile
	}
	var meta map[string]string
	diags.Append(m.Meta.ElementsAs(ctx, &meta, false)...)
	if diags.HasError() {
		return nil
	}
	profile.Meta = models.Meta(meta)
	return profile
}

func (r *profileResource) flattenProfile(ctx context.Context, p *models.Profile, m *profileResourceModel, diags *diag.Diagnostics) {
	m.Name = types.StringValue(p.Name)
	m.Description = mergeOptString(m.Description, p.Description)
	var sm map[string]string
	if p.Meta != nil {
		sm = map[string]string(p.Meta)
	}
	m.Meta = mergeOptStringMap(ctx, m.Meta, sm, diags)
}

func (r *profileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *profileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan profileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	profile := r.expandProfile(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.CreateModel(profile); err != nil {
		resp.Diagnostics.AddError("Create profile failed", err.Error())
		return
	}
	res, err := r.client.session.GetModel("profiles", profile.Name)
	if err != nil {
		resp.Diagnostics.AddError("Read profile after create failed", err.Error())
		return
	}
	r.flattenProfile(ctx, res.(*models.Profile), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *profileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state profileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pr, err := r.client.session.GetModel("profiles", state.Name.ValueString())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read profile failed", err.Error())
		return
	}
	p := pr.(*models.Profile)
	r.flattenProfile(ctx, p, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *profileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan profileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	profile := r.expandProfile(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.PutModel(profile); err != nil {
		resp.Diagnostics.AddError("Update profile failed", err.Error())
		return
	}
	res, err := r.client.session.GetModel("profiles", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read profile after update failed", err.Error())
		return
	}
	r.flattenProfile(ctx, res.(*models.Profile), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *profileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state profileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.session.DeleteModel("profiles", state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete profile failed", err.Error())
	}
}
