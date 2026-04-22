package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"context"
	"fmt"

	"github.com/VictorLowther/jsonpatch2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.com/rackn/provision/v4/models"
)

var _ resource.Resource = (*machineSetPoolResource)(nil)

type machineSetPoolResource struct {
	client *Config
}

func NewMachineSetPoolResource() resource.Resource {
	return &machineSetPoolResource{}
}

func (r *machineSetPoolResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_machine_set_pool"
}

func (r *machineSetPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "Machine UUID.",
				MarkdownDescription: "Machine UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pool": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				Description:         "Which pool to add the machine to.",
				MarkdownDescription: "Which pool to add the machine to.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Machine name.",
				MarkdownDescription: "Machine name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				Computed:            true,
				Description:         "Digital Rebar Machine.Address.",
				MarkdownDescription: "Digital Rebar Machine.Address.",
			},
		},
	}
}

func (r *machineSetPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type machineSetPoolResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Pool    types.String `tfsdk:"pool"`
	Name    types.String `tfsdk:"name"`
	Address types.String `tfsdk:"address"`
}

func (r *machineSetPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan machineSetPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool := plan.Pool.ValueString()
	if pool == "" {
		pool = "default"
	}
	name := plan.Name.ValueString()

	mo, err := r.client.session.ListModel("machines", "Name", name)
	if err != nil || len(mo) != 1 {
		resp.Diagnostics.AddError("Machine lookup failed", fmt.Sprintf("unable to get machine %s", name))
		return
	}
	machineObject := mo[0].(*models.Machine)
	tflog.Debug(ctx, "resolved machine", map[string]interface{}{"uuid": machineObject.Uuid.String(), "address": machineObject.Address.String()})

	plan.ID = types.StringValue(machineObject.Uuid.String())
	plan.Address = types.StringValue(machineObject.Address.String())

	if machineObject.Pool != pool {
		patch := jsonpatch2.Patch{{Op: "replace", Path: "/Pool", Value: pool}}
		reqm := r.client.session.Req().Patch(patch).UrlFor("machines", machineObject.Uuid.String())
		mr := models.Machine{}
		if err := reqm.Do(&mr); err != nil {
			resp.Diagnostics.AddError("Set pool failed", fmt.Sprintf("error setting pool %s: %s", pool, err))
			return
		}
	}

	r.readMachineSetPool(ctx, plan.ID.ValueString(), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *machineSetPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state machineSetPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.readMachineSetPool(ctx, state.ID.ValueString(), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *machineSetPoolResource) readMachineSetPool(_ context.Context, uuid string, m *machineSetPoolResourceModel, diags *diag.Diagnostics) {
	if uuid == "" {
		return
	}
	mo, err := r.client.session.GetModel("machines", uuid)
	if err != nil {
		if isNotFound(err) {
			m.ID = types.StringNull()
			return
		}
		diags.AddError("Read failed", fmt.Sprintf("error reading machine set pool: %s", err))
		return
	}
	machineObject := mo.(*models.Machine)
	m.Address = types.StringValue(machineObject.Address.String())
}

func (r *machineSetPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var state machineSetPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.readMachineSetPool(ctx, state.ID.ValueString(), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *machineSetPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state machineSetPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid := state.ID.ValueString()
	if uuid == "" {
		return
	}

	patch := jsonpatch2.Patch{{Op: "replace", Path: "/Pool", Value: "default"}}
	reqm := r.client.session.Req().Patch(patch).UrlFor("machines", uuid)
	mr := models.Machine{}
	if err := reqm.Do(&mr); err != nil {
		resp.Diagnostics.AddError("Delete failed", fmt.Sprintf("error setting pool default: %s", err))
		return
	}
	if mr.Pool != "default" {
		resp.Diagnostics.AddError("Delete failed", fmt.Sprintf("could not set default pool for %s", uuid))
	}
}
