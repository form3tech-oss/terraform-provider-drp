package drpv4

import (
	"context"
	"fmt"
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

var _ resource.Resource = (*profileParamResource)(nil)
var _ resource.ResourceWithImportState = (*profileParamResource)(nil)

type profileParamResource struct {
	client *Config
}

func NewProfileParamResource() resource.Resource {
	return &profileParamResource{}
}

func (r *profileParamResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_profile_param"
}

func (r *profileParamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"profile": schema.StringAttribute{
				Required:            true,
				Description:         "Profile name.",
				MarkdownDescription: "Profile name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Param name.",
				MarkdownDescription: "Param name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Optional:            true,
				Description:         "Param value (non-secure params).",
				MarkdownDescription: "Param value (non-secure params).",
			},
			"secure_value": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "Secure param value (encrypted to profile).",
				MarkdownDescription: "Secure param value (encrypted to profile).",
			},
		},
	}
}

func (r *profileParamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type profileParamResourceModel struct {
	Profile     types.String `tfsdk:"profile"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	SecureValue types.String `tfsdk:"secure_value"`
}

func (r *profileParamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format profile/name")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("profile"), types.StringValue(parts[0]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), types.StringValue(parts[1]))...)
}

func (r *profileParamResource) upsert(ctx context.Context, m *profileParamResourceModel, diags *diag.Diagnostics) {
	profile := m.Profile.ValueString()
	name := m.Name.ValueString()
	value := m.Value.ValueString()
	secureValue := m.SecureValue.ValueString()

	if (value == "" && secureValue == "") || (value != "" && secureValue != "") {
		diags.AddError(
			"Invalid profile param",
			"Exactly one of value or secure_value must be set.",
		)
		return
	}

	if value != "" && isParamSecure(r.client, name) {
		diags.AddError(
			"Invalid profile param",
			fmt.Sprintf("Param %s is secure; use secure_value instead.", name),
		)
		return
	}

	req := r.client.session.Req().UrlFor("profiles", profile, "params", name)

	if secureValue != "" {
		sv := &models.SecureData{}
		pubkey, err := getPublicKey(r.client, profile)
		if err != nil {
			diags.AddError("Read profile pubkey failed", err.Error())
			return
		}
		if err := sv.Marshal(pubkey, secureValue); err != nil {
			diags.AddError("Marshal secure value failed", err.Error())
			return
		}
		if err := sv.Validate(); err != nil {
			diags.AddError("Validate secure value failed", err.Error())
			return
		}
		if err := req.Post(sv).Do(nil); err != nil {
			diags.AddError("Set secure profile param failed", err.Error())
			return
		}
		return
	}

	convertedValue, err := convertParamToType(r.client, name, value)
	if err != nil {
		diags.AddError("Convert param value failed", err.Error())
		return
	}
	if err := req.Post(convertedValue).Do(nil); err != nil {
		diags.AddError("Set profile param failed", err.Error())
	}
}

func (r *profileParamResource) readIntoModel(ctx context.Context, m *profileParamResourceModel, diags *diag.Diagnostics) {
	profile := m.Profile.ValueString()
	name := m.Name.ValueString()
	secureValue := m.SecureValue.ValueString()

	var p interface{}
	if err := r.client.session.Req().UrlFor("profiles", profile, "params", name).Do(&p); err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			m.Profile = types.StringNull()
			m.Name = types.StringNull()
			return
		}
		diags.AddError("Read profile param failed", err.Error())
		return
	}

	m.Name = types.StringValue(name)
	m.Profile = types.StringValue(profile)

	if secureValue == "" {
		s, err := convertParamToString(p)
		if err != nil {
			diags.AddError("Convert param to string failed", err.Error())
			return
		}
		m.Value = types.StringValue(s)
		m.SecureValue = types.StringNull()
		return
	}

	m.Value = types.StringNull()
	var securedValue string
	if err := r.client.session.Req().UrlFor("profiles", profile, "params", name).Params("decode", "true").Do(&securedValue); err != nil {
		diags.AddError("Read decoded secure param failed", err.Error())
		return
	}
	m.SecureValue = types.StringValue(securedValue)
}

func (r *profileParamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan profileParamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	r.readIntoModel(ctx, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *profileParamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state profileParamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.readIntoModel(ctx, &state, &resp.Diagnostics)
	if state.Profile.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *profileParamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan profileParamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	r.readIntoModel(ctx, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *profileParamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state profileParamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.Req().Del().UrlFor("profiles", state.Profile.ValueString(), "params", state.Name.ValueString()).Do(nil); err != nil {
		resp.Diagnostics.AddError("Delete profile param failed", err.Error())
	}
}
