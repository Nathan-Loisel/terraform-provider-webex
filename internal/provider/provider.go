package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var _ provider.Provider = &WebexProvider{}

type WebexProvider struct {
	version string
}

type WebexProviderModel struct {
	Token   types.String `tfsdk:"token"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WebexProvider{version: version}
	}
}

func (p *WebexProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "webex"
	resp.Version = p.version
}

func (p *WebexProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Cisco Webex Workspaces, Devices, and Device Configurations.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "Webex API access token. Can also be set via the WEBEX_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "Webex API base URL. Defaults to https://webexapis.com/v1.",
				Optional:    true,
			},
		},
	}
}

func (p *WebexProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config WebexProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := os.Getenv("WEBEX_TOKEN")
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}
	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Webex Token",
			"The Webex API token must be set in the provider configuration or via the WEBEX_TOKEN environment variable.",
		)
		return
	}

	baseURL := ""
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	c := client.New(token, baseURL)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *WebexProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewLocationResource,
		NewFloorResource,
		NewWorkspaceResource,
		NewDeviceResource,
		NewDeviceConfigurationResource,
	}
}

func (p *WebexProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDevicesDataSource,
		NewLocationDataSource,
		NewWorkspaceDataSource,
		NewWorkspacesDataSource,
		NewDeviceConfigurationsDataSource,
	}
}
