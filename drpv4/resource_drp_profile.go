package drpv4

import (
	"context"
	"strings"

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
		},
	}
}

func (r *profileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type profileResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
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
	profile := &models.Profile{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	if err := r.client.session.CreateModel(profile); err != nil {
		resp.Diagnostics.AddError("Create profile failed", err.Error())
		return
	}
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
	state.Name = types.StringValue(p.Name)
	state.Description = types.StringValue(p.Description)
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
	profile := &models.Profile{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	if err := r.client.session.PutModel(profile); err != nil {
		resp.Diagnostics.AddError("Update profile failed", err.Error())
		return
	}
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
