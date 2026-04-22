package drpv4

import (
	"context"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.com/rackn/provision/v4/models"
)

var _ resource.Resource = (*reservationResource)(nil)
var _ resource.ResourceWithImportState = (*reservationResource)(nil)

type reservationResource struct {
	client *Config
}

func NewReservationResource() resource.Resource {
	return &reservationResource{}
}

func (r *reservationResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_reservation"
}

func (r *reservationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Required:            true,
				Description:         "Reservation IP address.",
				MarkdownDescription: "Reservation IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description":   schema.StringAttribute{Optional: true},
			"documentation": schema.StringAttribute{Optional: true},
			"duration":      schema.Int64Attribute{Optional: true},
			"next_server":   schema.StringAttribute{Optional: true},
			"scoped": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"strategy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("MAC"),
			},
			"token": schema.StringAttribute{Required: true, Description: "Reservation token."},
			"options": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: dhcpOptionAttributes()},
			},
		},
	}
}

func (r *reservationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type reservationResourceModel struct {
	Address       types.String `tfsdk:"address"`
	Description   types.String `tfsdk:"description"`
	Documentation types.String `tfsdk:"documentation"`
	Duration      types.Int64  `tfsdk:"duration"`
	NextServer    types.String `tfsdk:"next_server"`
	Scoped        types.Bool   `tfsdk:"scoped"`
	Strategy      types.String `tfsdk:"strategy"`
	Token         types.String `tfsdk:"token"`
	Options       types.List   `tfsdk:"options"`
}

func (r *reservationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("address"), req, resp)
}

func (r *reservationResource) expandReservationOptions(ctx context.Context, l types.List, diags *diag.Diagnostics) []models.DhcpOption {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]models.DhcpOption, 0, len(l.Elements()))
	for _, el := range l.Elements() {
		o, ok := el.(types.Object)
		if !ok {
			diags.AddError("Invalid options element", "expected object")
			return nil
		}
		code := o.Attributes()["code"].(types.Int64)
		val := o.Attributes()["value"].(types.String)
		out = append(out, models.DhcpOption{
			Code:  byte(code.ValueInt64()),
			Value: val.ValueString(),
		})
	}
	return out
}

func (r *reservationResource) expandReservation(ctx context.Context, m *reservationResourceModel, diags *diag.Diagnostics) *models.Reservation {
	duration := int32(0)
	if !m.Duration.IsNull() && !m.Duration.IsUnknown() {
		duration = int32(m.Duration.ValueInt64())
	}
	res := &models.Reservation{
		Description:   m.Description.ValueString(),
		Documentation: m.Documentation.ValueString(),
		Addr:          net.ParseIP(m.Address.ValueString()),
		Duration:      duration,
		Strategy:      m.Strategy.ValueString(),
		Token:         m.Token.ValueString(),
		Options:       r.expandReservationOptions(ctx, m.Options, diags),
	}
	if !m.Scoped.IsNull() && !m.Scoped.IsUnknown() {
		res.Scoped = m.Scoped.ValueBool()
	}
	if !m.NextServer.IsNull() && m.NextServer.ValueString() != "" {
		res.NextServer = net.ParseIP(m.NextServer.ValueString())
	}
	return res
}

func (r *reservationResource) flattenReservationOptionsMerged(ctx context.Context, prior types.List, opts []models.DhcpOption, diags *diag.Diagnostics) types.List {
	if prior.IsNull() || prior.IsUnknown() {
		return types.ListNull(dhcpOptionObjType())
	}
	if len(opts) == 0 {
		return types.ListNull(dhcpOptionObjType())
	}
	objs := make([]attr.Value, 0, len(opts))
	for _, opt := range opts {
		attrs := map[string]attr.Value{
			"code":  types.Int64Value(int64(opt.Code)),
			"value": types.StringValue(opt.Value),
		}
		obj, d := types.ObjectValue(dhcpOptionObjType().AttrTypes, attrs)
		diags.Append(d...)
		objs = append(objs, obj)
	}
	return types.ListValueMust(dhcpOptionObjType(), objs)
}

func (r *reservationResource) flattenReservation(ctx context.Context, res *models.Reservation, m *reservationResourceModel, diags *diag.Diagnostics) {
	m.Description = mergeOptString(m.Description, res.Description)
	m.Documentation = mergeOptString(m.Documentation, res.Documentation)
	m.Address = types.StringValue(res.Addr.String())
	m.Duration = mergeOptInt64(m.Duration, int64(res.Duration))
	if res.NextServer != nil {
		m.NextServer = types.StringValue(res.NextServer.String())
	} else {
		m.NextServer = mergeOptString(m.NextServer, "")
	}
	m.Scoped = mergeOptBool(m.Scoped, res.Scoped)
	m.Strategy = mergeOptString(m.Strategy, res.Strategy)
	m.Token = types.StringValue(res.Token)
	m.Options = r.flattenReservationOptionsMerged(ctx, m.Options, res.Options, diags)
}

func (r *reservationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan reservationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res := r.expandReservation(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.CreateModel(res); err != nil {
		resp.Diagnostics.AddError("Create reservation failed", err.Error())
		return
	}
	got, err := r.client.session.GetModel("reservations", plan.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read reservation after create failed", err.Error())
		return
	}
	r.flattenReservation(ctx, got.(*models.Reservation), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reservationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state reservationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.session.GetModel("reservations", state.Address.ValueString())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read reservation failed", err.Error())
		return
	}
	r.flattenReservation(ctx, res.(*models.Reservation), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *reservationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan reservationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res := r.expandReservation(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.PutModel(res); err != nil {
		resp.Diagnostics.AddError("Update reservation failed", err.Error())
		return
	}
	got, err := r.client.session.GetModel("reservations", plan.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read reservation after update failed", err.Error())
		return
	}
	r.flattenReservation(ctx, got.(*models.Reservation), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reservationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state reservationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.session.DeleteModel("reservations", state.Address.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete reservation failed", err.Error())
	}
}
