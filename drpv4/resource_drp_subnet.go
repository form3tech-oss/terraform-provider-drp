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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.com/rackn/provision/v4/models"
)

var _ resource.Resource = (*subnetResource)(nil)
var _ resource.ResourceWithImportState = (*subnetResource)(nil)

type subnetResource struct {
	client *Config
}

func NewSubnetResource() resource.Resource {
	return &subnetResource{}
}

func (r *subnetResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_subnet"
}

func dhcpOptionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"code": schema.Int64Attribute{Required: true, Description: "DHCP option code."},
		"value": schema.StringAttribute{
			Required:            true,
			Description:         "DHCP option value.",
			MarkdownDescription: "DHCP option value.",
		},
	}
}

func (r *subnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Subnet name.",
				MarkdownDescription: "Subnet name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description":   schema.StringAttribute{Optional: true},
			"documentation": schema.StringAttribute{Optional: true},
			"enabled":       schema.BoolAttribute{Optional: true},
			"subnet":        schema.StringAttribute{Required: true, Description: "CIDR or subnet mask."},
			"active_start":  schema.StringAttribute{Required: true},
			"active_end":    schema.StringAttribute{Required: true},
			"active_lease_time": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(60),
			},
			"next_server": schema.StringAttribute{Optional: true},
			"only_reservations": schema.BoolAttribute{
				Optional: true,
			},
			"options": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: dhcpOptionAttributes()},
			},
			"pickers": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"proxy": schema.BoolAttribute{Optional: true},
			"reserved_lease_time": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(7200),
			},
			"strategy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("MAC"),
			},
			"unmanaged": schema.BoolAttribute{Optional: true},
		},
	}
}

func (r *subnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type subnetResourceModel struct {
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Documentation     types.String `tfsdk:"documentation"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	Subnet            types.String `tfsdk:"subnet"`
	ActiveStart       types.String `tfsdk:"active_start"`
	ActiveEnd         types.String `tfsdk:"active_end"`
	ActiveLeaseTime   types.Int64  `tfsdk:"active_lease_time"`
	NextServer        types.String `tfsdk:"next_server"`
	OnlyReservations  types.Bool   `tfsdk:"only_reservations"`
	Options           types.List   `tfsdk:"options"`
	Pickers           types.List   `tfsdk:"pickers"`
	Proxy             types.Bool   `tfsdk:"proxy"`
	ReservedLeaseTime types.Int64  `tfsdk:"reserved_lease_time"`
	Strategy          types.String `tfsdk:"strategy"`
	Unmanaged         types.Bool   `tfsdk:"unmanaged"`
}

func dhcpOptionObjType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"code":  types.Int64Type,
		"value": types.StringType,
	}}
}

func (r *subnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *subnetResource) expandSubnetOptions(ctx context.Context, l types.List, diags *diag.Diagnostics) []models.DhcpOption {
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

func (r *subnetResource) expandSubnet(ctx context.Context, m *subnetResourceModel, diags *diag.Diagnostics) *models.Subnet {
	enabled := false
	if !m.Enabled.IsNull() && !m.Enabled.IsUnknown() {
		enabled = m.Enabled.ValueBool()
	}
	only := false
	if !m.OnlyReservations.IsNull() && !m.OnlyReservations.IsUnknown() {
		only = m.OnlyReservations.ValueBool()
	}
	proxy := false
	if !m.Proxy.IsNull() && !m.Proxy.IsUnknown() {
		proxy = m.Proxy.ValueBool()
	}
	unmanaged := false
	if !m.Unmanaged.IsNull() && !m.Unmanaged.IsUnknown() {
		unmanaged = m.Unmanaged.ValueBool()
	}
	var next net.IP
	if !m.NextServer.IsNull() && m.NextServer.ValueString() != "" {
		next = net.ParseIP(m.NextServer.ValueString())
	}
	return &models.Subnet{
		Name:              m.Name.ValueString(),
		Description:       m.Description.ValueString(),
		Documentation:     m.Documentation.ValueString(),
		Enabled:           enabled,
		Subnet:            m.Subnet.ValueString(),
		ActiveStart:       net.ParseIP(m.ActiveStart.ValueString()),
		ActiveEnd:         net.ParseIP(m.ActiveEnd.ValueString()),
		ActiveLeaseTime:   int32(m.ActiveLeaseTime.ValueInt64()),
		NextServer:        next,
		OnlyReservations:  only,
		Options:           r.expandSubnetOptions(ctx, m.Options, diags),
		Pickers:           diagListToStrings(ctx, m.Pickers, diags),
		Proxy:             proxy,
		ReservedLeaseTime: int32(m.ReservedLeaseTime.ValueInt64()),
		Strategy:          m.Strategy.ValueString(),
		Unmanaged:         unmanaged,
	}
}

func (r *subnetResource) flattenSubnetOptions(ctx context.Context, opts []models.DhcpOption, diags *diag.Diagnostics) types.List {
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

func (r *subnetResource) flattenSubnet(ctx context.Context, s *models.Subnet, m *subnetResourceModel, diags *diag.Diagnostics) {
	m.Name = types.StringValue(s.Name)
	m.Description = types.StringValue(s.Description)
	m.Documentation = types.StringValue(s.Documentation)
	m.Enabled = types.BoolValue(s.Enabled)
	m.Subnet = types.StringValue(s.Subnet)
	m.ActiveStart = types.StringValue(s.ActiveStart.String())
	m.ActiveEnd = types.StringValue(s.ActiveEnd.String())
	m.ActiveLeaseTime = types.Int64Value(int64(s.ActiveLeaseTime))
	if s.NextServer != nil {
		m.NextServer = types.StringValue(s.NextServer.String())
	} else {
		m.NextServer = types.StringNull()
	}
	m.OnlyReservations = types.BoolValue(s.OnlyReservations)
	m.Options = r.flattenSubnetOptions(ctx, s.Options, diags)
	m.Pickers = mustListStrings(ctx, s.Pickers, diags)
	m.Proxy = types.BoolValue(s.Proxy)
	m.ReservedLeaseTime = types.Int64Value(int64(s.ReservedLeaseTime))
	m.Strategy = types.StringValue(s.Strategy)
	m.Unmanaged = types.BoolValue(s.Unmanaged)
}

func (r *subnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sub := r.expandSubnet(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.CreateModel(sub); err != nil {
		resp.Diagnostics.AddError("Create subnet failed", err.Error())
		return
	}
	r.flattenSubnet(ctx, sub, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state subnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.session.GetModel("subnets", state.Name.ValueString())
	if err != nil {
		if strings.HasSuffix(err.Error(), "Not Found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read subnet failed", err.Error())
		return
	}
	r.flattenSubnet(ctx, res.(*models.Subnet), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var plan subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sub := r.expandSubnet(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.session.PutModel(sub); err != nil {
		resp.Diagnostics.AddError("Update subnet failed", err.Error())
		return
	}
	r.flattenSubnet(ctx, sub, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state subnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.session.DeleteModel("subnets", state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete subnet failed", err.Error())
	}
}
