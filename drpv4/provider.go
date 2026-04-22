package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type fwProvider struct {
	version string
}

type providerModel struct {
	Token    types.String `tfsdk:"token"`
	Key      types.String `tfsdk:"key"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Endpoint types.String `tfsdk:"endpoint"`
}

func NewProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &fwProvider{version: version}
	}
}

func (p *fwProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "drp"
	resp.Version = p.version
}

func (p *fwProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "Granted DRP token (use instead of RS_KEY)",
				MarkdownDescription: "Granted DRP token (use instead of RS_KEY)",
			},
			"key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "The DRP user:password key",
				MarkdownDescription: "The DRP user:password key",
			},
			"username": schema.StringAttribute{
				Optional:            true,
				Description:         "The DRP user",
				MarkdownDescription: "The DRP user",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "The DRP password",
				MarkdownDescription: "The DRP password",
			},
			"endpoint": schema.StringAttribute{
				Optional:            true,
				Description:         "The DRP server URL, for example https://1.2.3.4:8092",
				MarkdownDescription: "The DRP server URL, for example https://1.2.3.4:8092",
			},
		},
	}
}

func (p *fwProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := data.Token.ValueString()
	if token == "" {
		token = os.Getenv("RS_TOKEN")
	}
	key := data.Key.ValueString()
	if key == "" {
		key = os.Getenv("RS_KEY")
	}
	username := data.Username.ValueString()
	password := data.Password.ValueString()
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("RS_ENDPOINT")
	}

	if key != "" {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) < 2 {
			resp.Diagnostics.AddError("Malformed key", "While configuring the provider, the key attribute is malformed.")
			return
		}
		username = parts[0]
		password = parts[1]
	}

	if token != "" && key != "" {
		resp.Diagnostics.AddError("Conflicting credentials", "token cannot be set together with key.")
		return
	}

	cfg := &Config{
		token:    token,
		username: username,
		password: password,
		endpoint: endpoint,
	}

	if cfg.endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing DRP Endpoint",
			"No DRP endpoint was specified; set the endpoint attribute or RS_ENDPOINT.",
		)
		return
	}
	if cfg.token == "" && cfg.username == "" {
		resp.Diagnostics.AddError(
			"Malformed DRP credentials",
			"The key, token, or username/password attributes must be provided.",
		)
		return
	}
	if cfg.username != "" && cfg.password == "" {
		resp.Diagnostics.AddError("Missing DRP password", "A password is required for the specified user.")
		return
	}

	tflog.Debug(ctx, "Attempting to connect to DRP", map[string]interface{}{"endpoint": cfg.endpoint})
	if err := cfg.validateAndConnect(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to create DRP client", err.Error())
		return
	}

	info, err := cfg.session.Info()
	if err != nil {
		resp.Diagnostics.AddError("Failed to Connect", fmt.Sprintf("Failed to fetch info for %s", cfg.endpoint))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Digital Rebar %v (features: %v)", info.Version, info.Features))
	resp.ResourceData = cfg
}

func (p *fwProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMachineResource,
		NewMachineSetPoolResource,
		NewParamResource,
		NewTemplateResource,
		NewTaskResource,
		NewStageResource,
		NewWorkflowResource,
		NewSubnetResource,
		NewReservationResource,
		NewPoolResource,
		NewProfileResource,
		NewProfileParamResource,
	}
}

func (p *fwProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
