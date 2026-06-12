package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var _ datasource.DataSource = &DeviceConfigurationsDataSource{}

type DeviceConfigurationsDataSource struct {
	client *client.Client
}

type DeviceConfigurationsDataSourceModel struct {
	DeviceID       types.String `tfsdk:"device_id"`
	KeyFilter      types.String `tfsdk:"key_filter"`
	Configurations types.Map    `tfsdk:"configurations"`
	Defaults       types.Map    `tfsdk:"defaults"`
	Sources        types.Map    `tfsdk:"sources"`
}

func NewDeviceConfigurationsDataSource() datasource.DataSource {
	return &DeviceConfigurationsDataSource{}
}

func (d *DeviceConfigurationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_configurations"
}

func (d *DeviceConfigurationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads all device configurations (xAPI settings) for a Webex device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:    true,
				Description: "The device to read configurations from.",
			},
			"key_filter": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by key prefix. Supports wildcards (e.g., 'Audio.*', 'Standby.*').",
			},
			"configurations": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Map of configuration key to current effective value.",
			},
			"defaults": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Map of configuration key to factory default value.",
			},
			"sources": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Map of configuration key to value source (default or configured).",
			},
		},
	}
}

func (d *DeviceConfigurationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *DeviceConfigurationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DeviceConfigurationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := config.DeviceID.ValueString()
	keyFilter := ""
	if !config.KeyFilter.IsNull() {
		keyFilter = config.KeyFilter.ValueString()
	}

	configResp, err := d.client.GetDeviceConfigurations(ctx, deviceID, keyFilter)
	if err != nil {
		resp.Diagnostics.AddError("Error reading device configurations", err.Error())
		return
	}

	values := make(map[string]string)
	defaults := make(map[string]string)
	sources := make(map[string]string)

	for key, cfg := range configResp.Items {
		values[key] = fmt.Sprintf("%v", cfg.Value)
		sources[key] = cfg.Source

		if cfg.Sources != nil {
			if defSource, ok := cfg.Sources["default"]; ok {
				defaults[key] = fmt.Sprintf("%v", defSource.Value)
			}
		}
	}

	valuesMap, diags := types.MapValueFrom(ctx, types.StringType, values)
	resp.Diagnostics.Append(diags...)
	defaultsMap, diags := types.MapValueFrom(ctx, types.StringType, defaults)
	resp.Diagnostics.Append(diags...)
	sourcesMap, diags := types.MapValueFrom(ctx, types.StringType, sources)
	resp.Diagnostics.Append(diags...)

	config.Configurations = valuesMap
	config.Defaults = defaultsMap
	config.Sources = sourcesMap

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
