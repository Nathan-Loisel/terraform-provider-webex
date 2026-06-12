package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var _ datasource.DataSource = &DevicesDataSource{}

type DevicesDataSource struct {
	client *client.Client
}

type DevicesDataSourceModel struct {
	WorkspaceID types.String       `tfsdk:"workspace_id"`
	Devices     []DeviceItemModel  `tfsdk:"devices"`
}

type DeviceItemModel struct {
	ID               types.String `tfsdk:"id"`
	DisplayName      types.String `tfsdk:"display_name"`
	WorkspaceID      types.String `tfsdk:"workspace_id"`
	PersonID         types.String `tfsdk:"person_id"`
	OrgID            types.String `tfsdk:"org_id"`
	Product          types.String `tfsdk:"product"`
	Type             types.String `tfsdk:"type"`
	Serial           types.String `tfsdk:"serial"`
	ConnectionStatus types.String `tfsdk:"connection_status"`
	IP               types.String `tfsdk:"ip"`
	MAC              types.String `tfsdk:"mac"`
	Software         types.String `tfsdk:"software"`
	PrimarySIPURL    types.String `tfsdk:"primary_sip_url"`
	UpgradeChannel   types.String `tfsdk:"upgrade_channel"`
	Created          types.String `tfsdk:"created"`
	FirstSeen        types.String `tfsdk:"first_seen"`
	LastSeen         types.String `tfsdk:"last_seen"`
	LocationID       types.String `tfsdk:"location_id"`
	ManagedBy        types.String `tfsdk:"managed_by"`
	DevicePlatform   types.String `tfsdk:"device_platform"`
	Tags             types.List   `tfsdk:"tags"`
}

func NewDevicesDataSource() datasource.DataSource {
	return &DevicesDataSource{}
}

func (d *DevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devices"
}

func (d *DevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Webex devices, optionally filtered by workspace.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter devices by workspace ID.",
			},
			"devices": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of devices.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Device ID.",
						},
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "Device display name.",
						},
						"workspace_id": schema.StringAttribute{
							Computed:    true,
							Description: "Workspace ID.",
						},
						"product": schema.StringAttribute{
							Computed:    true,
							Description: "Product name.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Device type.",
						},
						"serial": schema.StringAttribute{
							Computed:    true,
							Description: "Serial number.",
						},
						"connection_status": schema.StringAttribute{
							Computed:    true,
							Description: "Connection status.",
						},
						"ip": schema.StringAttribute{
							Computed:    true,
							Description: "IP address.",
						},
						"mac": schema.StringAttribute{
							Computed:    true,
							Description: "MAC address.",
						},
						"software": schema.StringAttribute{
							Computed:    true,
							Description: "Software version.",
						},
						"person_id": schema.StringAttribute{
							Computed:    true,
							Description: "Person ID if assigned to a person.",
						},
						"org_id": schema.StringAttribute{
							Computed:    true,
							Description: "Organization ID.",
						},
						"primary_sip_url": schema.StringAttribute{
							Computed:    true,
							Description: "Primary SIP URL.",
						},
						"upgrade_channel": schema.StringAttribute{
							Computed:    true,
							Description: "Upgrade channel.",
						},
						"created": schema.StringAttribute{
							Computed:    true,
							Description: "Creation timestamp (ISO 8601).",
						},
						"first_seen": schema.StringAttribute{
							Computed:    true,
							Description: "First seen timestamp (ISO 8601).",
						},
						"last_seen": schema.StringAttribute{
							Computed:    true,
							Description: "Last seen timestamp (ISO 8601).",
						},
						"location_id": schema.StringAttribute{
							Computed:    true,
							Description: "Location ID.",
						},
						"managed_by": schema.StringAttribute{
							Computed:    true,
							Description: "Entity managing the device.",
						},
						"device_platform": schema.StringAttribute{
							Computed:    true,
							Description: "Device platform.",
						},
						"tags": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Tags assigned to the device.",
						},
					},
				},
			},
		},
	}
}

func (d *DevicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DevicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wsID := ""
	if !config.WorkspaceID.IsNull() {
		wsID = config.WorkspaceID.ValueString()
	}

	devices, err := d.client.ListDevices(ctx, wsID)
	if err != nil {
		resp.Diagnostics.AddError("Error listing devices", err.Error())
		return
	}

	config.Devices = make([]DeviceItemModel, len(devices))
	for i, dev := range devices {
		tags, diags := types.ListValueFrom(ctx, types.StringType, dev.Tags)
		if diags.HasError() {
			tags = types.ListNull(types.StringType)
		}
		config.Devices[i] = DeviceItemModel{
			ID:               types.StringValue(dev.ID),
			DisplayName:      types.StringValue(dev.DisplayName),
			WorkspaceID:      stringOrNull(dev.WorkspaceID),
			PersonID:         stringOrNull(dev.PersonID),
			OrgID:            stringOrNull(dev.OrgID),
			Product:          stringOrNull(dev.Product),
			Type:             stringOrNull(dev.Type),
			Serial:           stringOrNull(dev.Serial),
			ConnectionStatus: stringOrNull(dev.ConnectionStatus),
			IP:               stringOrNull(dev.IP),
			MAC:              stringOrNull(dev.MAC),
			Software:         stringOrNull(dev.Software),
			PrimarySIPURL:    stringOrNull(dev.PrimarySIPURL),
			UpgradeChannel:   stringOrNull(dev.UpgradeChannel),
			Created:          stringOrNull(dev.Created),
			FirstSeen:        stringOrNull(dev.FirstSeen),
			LastSeen:         stringOrNull(dev.LastSeen),
			LocationID:       stringOrNull(dev.LocationID),
			ManagedBy:        stringOrNull(dev.ManagedBy),
			DevicePlatform:   stringOrNull(dev.DevicePlatform),
			Tags:             tags,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
