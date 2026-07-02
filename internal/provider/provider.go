package provider

import (
	"context"
	"fmt"
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
	Token        types.String `tfsdk:"token"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	RefreshToken types.String `tfsdk:"refresh_token"`
	BaseURL      types.String `tfsdk:"base_url"`
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
		MarkdownDescription: `Terraform provider for managing Cisco Webex Workspaces, Devices, and Device Configurations.

## Authentication

Two authentication methods are supported:

**Method 1 — Static access token:** Set the ` + "`token`" + ` attribute or the ` + "`WEBEX_TOKEN`" + ` environment variable.

**Method 2 — OAuth refresh:** Set ` + "`client_id`" + `, ` + "`client_secret`" + `, and ` + "`refresh_token`" + ` (or their environment variables). The provider will automatically exchange them for a fresh access token on each run.`,
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "Webex API access token. Can also be set via the WEBEX_TOKEN environment variable. " +
					"Note: developer portal tokens expire after 12 hours; Service App access tokens expire after 14 days.",
				Optional:  true,
				Sensitive: true,
			},
			"client_id": schema.StringAttribute{
				Description: "Webex Service App Client ID. Can also be set via the WEBEX_CLIENT_ID environment variable. " +
					"Required together with client_secret and refresh_token for OAuth authentication.",
				Optional: true,
			},
			"client_secret": schema.StringAttribute{
				Description: "Webex Service App Client Secret. Can also be set via the WEBEX_CLIENT_SECRET environment variable. " +
					"Required together with client_id and refresh_token for OAuth authentication.",
				Optional:  true,
				Sensitive: true,
			},
			"refresh_token": schema.StringAttribute{
				Description: "Webex Service App Refresh Token. Can also be set via the WEBEX_REFRESH_TOKEN environment variable. " +
					"Required together with client_id and client_secret for OAuth authentication. " +
					"Refresh tokens expire after 90 days and must be regenerated from the Webex Developer Portal.",
				Optional:  true,
				Sensitive: true,
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

	// Method 1: Static access token
	token := os.Getenv("WEBEX_TOKEN")
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// Method 2: OAuth refresh (client_id + client_secret + refresh_token)
	if token == "" {
		clientID := os.Getenv("WEBEX_CLIENT_ID")
		if !config.ClientID.IsNull() {
			clientID = config.ClientID.ValueString()
		}
		clientSecret := os.Getenv("WEBEX_CLIENT_SECRET")
		if !config.ClientSecret.IsNull() {
			clientSecret = config.ClientSecret.ValueString()
		}
		refreshToken := os.Getenv("WEBEX_REFRESH_TOKEN")
		if !config.RefreshToken.IsNull() {
			refreshToken = config.RefreshToken.ValueString()
		}

		oauthFields := map[string]string{
			"client_id (WEBEX_CLIENT_ID)":         clientID,
			"client_secret (WEBEX_CLIENT_SECRET)":  clientSecret,
			"refresh_token (WEBEX_REFRESH_TOKEN)":  refreshToken,
		}

		hasAny := clientID != "" || clientSecret != "" || refreshToken != ""
		hasAll := clientID != "" && clientSecret != "" && refreshToken != ""

		if hasAny && !hasAll {
			var missing []string
			for name, val := range oauthFields {
				if val == "" {
					missing = append(missing, name)
				}
			}
			resp.Diagnostics.AddError(
				"Incomplete OAuth Configuration",
				"When using OAuth authentication, all three fields are required: client_id, client_secret, and refresh_token.\n"+
					"Missing: "+fmt.Sprintf("%v", missing),
			)
			return
		}

		if hasAll {
			var err error
			token, err = client.FetchAccessToken(clientID, clientSecret, refreshToken)
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to Obtain Access Token",
					err.Error(),
				)
				return
			}
		}
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Webex Authentication",
			"No authentication method configured. Use one of the following:\n\n"+
				"Method 1 — Static access token:\n"+
				"  Set 'token' in the provider block or the WEBEX_TOKEN environment variable.\n"+
				"  Note: developer portal tokens expire after 12 hours; Service App access tokens expire after 14 days.\n\n"+
				"Method 2 — OAuth refresh (recommended for Service Apps):\n"+
				"  Set 'client_id', 'client_secret', and 'refresh_token' in the provider block\n"+
				"  or via WEBEX_CLIENT_ID, WEBEX_CLIENT_SECRET, and WEBEX_REFRESH_TOKEN environment variables.\n"+
				"  The provider will automatically exchange them for a fresh access token.\n"+
				"  Note: refresh tokens expire after 90 days and must be regenerated from the Webex Developer Portal.",
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
