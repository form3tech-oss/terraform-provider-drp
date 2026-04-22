package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.com/rackn/provision/v4/models"
)

var _ resource.Resource = (*machineResource)(nil)

type machineResource struct {
	client *Config
}

func NewMachineResource() resource.Resource {
	return &machineResource{}
}

func (r *machineResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "drp_machine"
}

func (r *machineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "Machine UUID allocated from the pool.",
				MarkdownDescription: "Machine UUID allocated from the pool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pool": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				Description:         "Pool to operate for machine actions (Machine.Pool).",
				MarkdownDescription: "Pool to operate for machine actions (Machine.Pool).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeout": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("5m"),
				Description:         "Max time string to wait for pool operations.",
				MarkdownDescription: "Max time string to wait for pool operations.",
			},
			"add_profiles": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Profiles to add to Machine.Profiles (must already exist).",
				MarkdownDescription: "Profiles to add to Machine.Profiles (must already exist).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"add_parameters": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Parameters (key: value) to add to Machine.Params.",
				MarkdownDescription: "Parameters (key: value) to add to Machine.Params.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"filters": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Selection filters (Digital Rebar format e.g. FilterVar=value).",
				MarkdownDescription: "Selection filters (Digital Rebar format e.g. FilterVar=value).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"authorized_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Sets access-keys param on the machine requested.",
				MarkdownDescription: "Sets access-keys param on the machine requested.",
			},
			"address": schema.StringAttribute{
				Computed:            true,
				Description:         "Digital Rebar Machine.Address.",
				MarkdownDescription: "Digital Rebar Machine.Address.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Digital Rebar Machine.Status / pool status.",
				MarkdownDescription: "Digital Rebar Machine.Status / pool status.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Description:         "Digital Rebar Machine.Name.",
				MarkdownDescription: "Digital Rebar Machine.Name.",
			},
		},
	}
}

func (r *machineResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

type machineResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Pool           types.String `tfsdk:"pool"`
	Timeout        types.String `tfsdk:"timeout"`
	AddProfiles    types.List   `tfsdk:"add_profiles"`
	AddParameters  types.List   `tfsdk:"add_parameters"`
	Filters        types.List   `tfsdk:"filters"`
	AuthorizedKeys types.List   `tfsdk:"authorized_keys"`
	Address        types.String `tfsdk:"address"`
	Status         types.String `tfsdk:"status"`
	Name           types.String `tfsdk:"name"`
}

func (r *machineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		return
	}
	var plan machineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool := plan.Pool.ValueString()
	if pool == "" {
		pool = "default"
	}
	timeout := plan.Timeout.ValueString()
	parms := map[string]interface{}{
		"pool/wait-timeout": timeout,
	}

	if !plan.AddProfiles.IsNull() && !plan.AddProfiles.IsUnknown() {
		profiles, d := diagListToStringSlice(ctx, plan.AddProfiles)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(profiles) > 0 {
			ifaces := make([]interface{}, len(profiles))
			for i, p := range profiles {
				ifaces[i] = p
			}
			parms["pool/add-profiles"] = ifaces
		}
	}

	parameters := map[string]interface{}{}
	if !plan.AuthorizedKeys.IsNull() && !plan.AuthorizedKeys.IsUnknown() {
		akeys, d := diagListToStringSlice(ctx, plan.AuthorizedKeys)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		accesskeys := map[string]string{}
		for i, p := range akeys {
			accesskeys[fmt.Sprintf("terraform-%d", i)] = p
		}
		if len(accesskeys) > 0 {
			parameters["access-keys"] = accesskeys
		}
	}
	if !plan.AddParameters.IsNull() && !plan.AddParameters.IsUnknown() {
		aparams, d := diagListToStringSlice(ctx, plan.AddParameters)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, p := range aparams {
			param := strings.SplitN(p, ":", 2)
			if len(param) < 2 {
				resp.Diagnostics.AddError("Invalid add_parameters entry", fmt.Sprintf("expected key:value, got %q", p))
				return
			}
			key := param[0]
			value := strings.TrimLeft(param[1], " ")
			parameters[key] = value
		}
	}
	if len(parameters) > 0 {
		parms["pool/add-parameters"] = parameters
	}

	if !plan.Filters.IsNull() && !plan.Filters.IsUnknown() {
		filters, d := diagListToStringSlice(ctx, plan.Filters)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(filters) > 0 {
			ifaces := make([]interface{}, len(filters))
			for i, f := range filters {
				ifaces[i] = f
			}
			parms["pool/filter"] = ifaces
		}
	}

	pr := []*models.PoolResult{}
	reqAPI := r.client.session.Req().Post(parms).UrlFor("pools", pool, "allocateMachines")
	if err := reqAPI.Do(&pr); err != nil {
		tflog.Debug(ctx, "allocateMachines failed", map[string]interface{}{"error": err.Error()})
		resp.Diagnostics.AddError("Allocation failed", fmt.Sprintf("Error allocating from pool %s: %s", pool, err))
		return
	}
	mc := pr[0]
	tflog.Debug(ctx, "allocated machine", map[string]interface{}{"status": mc.Status, "name": mc.Name, "uuid": mc.Uuid})

	plan.ID = types.StringValue(mc.Uuid)
	plan.Status = types.StringValue(string(mc.Status))
	plan.Name = types.StringValue(mc.Name)
	r.machineReadIntoModel(ctx, mc.Uuid, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *machineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		return
	}
	var state machineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.machineReadIntoModel(ctx, state.ID.ValueString(), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *machineResource) machineReadIntoModel(ctx context.Context, uuid string, m *machineResourceModel, diags *diag.Diagnostics) {
	if uuid == "" {
		return
	}
	mo, err := r.client.session.GetModel("machines", uuid)
	if err != nil {
		if isNotFound(err) {
			m.ID = types.StringNull()
			return
		}
		diags.AddError("Read machine failed", fmt.Sprintf("unable to get machine %s: %s", uuid, err))
		return
	}
	machineObject := mo.(*models.Machine)
	m.Status = types.StringValue(string(machineObject.PoolStatus))
	m.Address = types.StringValue(machineObject.Address.String())
	m.Name = types.StringValue(machineObject.Name)
}

func (r *machineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		return
	}
	var state machineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.machineReadIntoModel(ctx, state.ID.ValueString(), &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *machineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		return
	}
	var state machineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := state.ID.ValueString()
	if uuid == "" {
		return
	}
	pool := state.Pool.ValueString()
	if pool == "" {
		resp.Diagnostics.AddError("Delete machine failed", "pool is required to release the machine")
		return
	}

	var readState machineResourceModel
	readState.ID = state.ID
	readState.Pool = state.Pool
	readState.Timeout = state.Timeout
	readState.AddProfiles = state.AddProfiles
	readState.AddParameters = state.AddParameters
	readState.Filters = state.Filters
	readState.AuthorizedKeys = state.AuthorizedKeys
	r.machineReadIntoModel(ctx, uuid, &readState, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if readState.Status.ValueString() == "Free" {
		return
	}

	pr := []*models.PoolResult{}
	parms := map[string]interface{}{
		"pool/wait-timeout": state.Timeout.ValueString(),
		"pool/machine-list": []string{uuid},
	}
	if !state.AddProfiles.IsNull() && !state.AddProfiles.IsUnknown() {
		profiles, d := diagListToStringSlice(ctx, state.AddProfiles)
		resp.Diagnostics.Append(d...)
		if len(profiles) > 0 {
			ifaces := make([]interface{}, len(profiles))
			for i, p := range profiles {
				ifaces[i] = p
			}
			parms["pool/remove-profiles"] = ifaces
		}
	}
	parameters := []string{}
	if !state.AuthorizedKeys.IsNull() && !state.AuthorizedKeys.IsUnknown() {
		akeys, d := diagListToStringSlice(ctx, state.AuthorizedKeys)
		resp.Diagnostics.Append(d...)
		if len(akeys) > 0 {
			parameters = append(parameters, "access-keys")
		}
	}
	if !state.AddParameters.IsNull() && !state.AddParameters.IsUnknown() {
		aparams, d := diagListToStringSlice(ctx, state.AddParameters)
		resp.Diagnostics.Append(d...)
		for _, p := range aparams {
			param := strings.SplitN(p, ":", 2)
			if len(param) < 2 {
				resp.Diagnostics.AddError("Invalid add_parameters entry", fmt.Sprintf("expected key:value, got %q", p))
				return
			}
			parameters = append(parameters, param[0])
		}
	}
	if len(parameters) > 0 {
		parms["pool/remove-parameters"] = parameters
	}

	reqAPI := r.client.session.Req().Post(parms).UrlFor("pools", pool, "releaseMachines")
	if err := reqAPI.Do(&pr); err != nil {
		resp.Diagnostics.AddError("Release failed", fmt.Sprintf("Error releasing %s from pool %s: %s", uuid, pool, err))
		return
	}
	mc := pr[0]
	if mc.Status != "Free" {
		resp.Diagnostics.AddError("Release failed", fmt.Sprintf("Could not release %s from pool %s", uuid, pool))
	}
}
