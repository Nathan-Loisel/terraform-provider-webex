package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var _ datasource.DataSource = &WorkspacesDataSource{}

type WorkspacesDataSource struct {
	client *client.Client
}

type WorkspacesDataSourceModel struct {
	DisplayName types.String         `tfsdk:"display_name"`
	Workspaces  []WorkspaceItemModel `tfsdk:"workspaces"`
}

type WorkspaceItemCallingModel struct {
	Type        types.String `tfsdk:"type"`
	PhoneNumber types.String `tfsdk:"phone_number"`
	Extension   types.String `tfsdk:"extension"`
	LocationID  types.String `tfsdk:"location_id"`
	Licenses    types.List   `tfsdk:"licenses"`
}

type WorkspaceItemCalendarModel struct {
	Type            types.String `tfsdk:"type"`
	EmailAddress    types.String `tfsdk:"email_address"`
	ResourceGroupID types.String `tfsdk:"resource_group_id"`
}

type WorkspaceItemDeviceHostedMeetingsModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	SiteURL types.String `tfsdk:"site_url"`
}

type WorkspaceItemModel struct {
	ID                   types.String                            `tfsdk:"id"`
	DisplayName          types.String                            `tfsdk:"display_name"`
	OrgID                types.String                            `tfsdk:"org_id"`
	Type                 types.String                            `tfsdk:"type"`
	Capacity             types.Int64                             `tfsdk:"capacity"`
	LocationID           types.String                            `tfsdk:"location_id"`
	FloorID              types.String                            `tfsdk:"floor_id"`
	SIPAddress           types.String                            `tfsdk:"sip_address"`
	Created              types.String                            `tfsdk:"created"`
	Notes                types.String                            `tfsdk:"notes"`
	HotdeskingStatus     types.String                            `tfsdk:"hotdesking_status"`
	SupportedDevices     types.String                            `tfsdk:"supported_devices"`
	DevicePlatform       types.String                            `tfsdk:"device_platform"`
	IndoorNavigationURL  types.String                            `tfsdk:"indoor_navigation_url"`
	Calling              *WorkspaceItemCallingModel              `tfsdk:"calling"`
	Calendar             *WorkspaceItemCalendarModel             `tfsdk:"calendar"`
	DeviceHostedMeetings *WorkspaceItemDeviceHostedMeetingsModel `tfsdk:"device_hosted_meetings"`
}

func NewWorkspacesDataSource() datasource.DataSource {
	return &WorkspacesDataSource{}
}

func (d *WorkspacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspaces"
}

func (d *WorkspacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Webex Workspaces, optionally filtered by display name.",
		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter workspaces by display name (partial match).",
			},
			"workspaces": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of workspaces.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Workspace ID.",
						},
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "Display name.",
						},
						"org_id": schema.StringAttribute{
							Computed:    true,
							Description: "Organization ID.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Workspace type.",
						},
						"capacity": schema.Int64Attribute{
							Computed:    true,
							Description: "Room capacity.",
						},
						"location_id": schema.StringAttribute{
							Computed:    true,
							Description: "Location ID.",
						},
						"floor_id": schema.StringAttribute{
							Computed:    true,
							Description: "Floor ID.",
						},
						"sip_address": schema.StringAttribute{
							Computed:    true,
							Description: "SIP address.",
						},
						"created": schema.StringAttribute{
							Computed:    true,
							Description: "Creation timestamp.",
						},
						"notes": schema.StringAttribute{
							Computed:    true,
							Description: "Notes.",
						},
						"hotdesking_status": schema.StringAttribute{
							Computed:    true,
							Description: "Hot desking status.",
						},
						"supported_devices": schema.StringAttribute{
							Computed:    true,
							Description: "Supported device type.",
						},
						"device_platform": schema.StringAttribute{
							Computed:    true,
							Description: "Device platform.",
						},
						"indoor_navigation_url": schema.StringAttribute{
							Computed:    true,
							Description: "URL of a map locating the workspace.",
						},
						"calling": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Calling configuration.",
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Computed:    true,
									Description: "Calling type.",
								},
								"phone_number": schema.StringAttribute{
									Computed:    true,
									Description: "Phone number.",
								},
								"extension": schema.StringAttribute{
									Computed:    true,
									Description: "Extension.",
								},
								"location_id": schema.StringAttribute{
									Computed:    true,
									Description: "Calling location ID.",
								},
								"licenses": schema.ListAttribute{
									Computed:    true,
									ElementType: types.StringType,
									Description: "Webex Calling licenses.",
								},
							},
						},
						"calendar": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Calendar configuration.",
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Computed:    true,
									Description: "Calendar type.",
								},
								"email_address": schema.StringAttribute{
									Computed:    true,
									Description: "Calendar email address.",
								},
								"resource_group_id": schema.StringAttribute{
									Computed:    true,
									Description: "Calendar resource group ID.",
								},
							},
						},
						"device_hosted_meetings": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Device hosted meetings configuration.",
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Computed:    true,
									Description: "Whether device hosted meetings are enabled.",
								},
								"site_url": schema.StringAttribute{
									Computed:    true,
									Description: "Webex site URL.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *WorkspacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WorkspacesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := ""
	if !config.DisplayName.IsNull() {
		name = config.DisplayName.ValueString()
	}

	workspaces, err := d.client.ListWorkspaces(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error listing workspaces", err.Error())
		return
	}

	config.Workspaces = make([]WorkspaceItemModel, len(workspaces))
	for i, ws := range workspaces {
		item := WorkspaceItemModel{
			ID:               types.StringValue(ws.ID),
			DisplayName:      types.StringValue(ws.DisplayName),
			OrgID:            stringOrNull(ws.OrgID),
			Type:             stringOrNull(ws.Type),
			LocationID:       stringOrNull(ws.LocationID),
			FloorID:          stringOrNull(ws.FloorID),
			SIPAddress:       stringOrNull(ws.SIPAddress),
			Created:          stringOrNull(ws.Created),
			Notes:            stringOrNull(ws.Notes),
			HotdeskingStatus: stringOrNull(ws.HotdeskingStatus),
			SupportedDevices: stringOrNull(ws.SupportedDevices),
			DevicePlatform:   stringOrNull(ws.DevicePlatform),
		}
		if ws.Capacity != nil {
			item.Capacity = types.Int64Value(int64(*ws.Capacity))
		} else {
			item.Capacity = types.Int64Null()
		}

		if ws.IndoorNavigation != nil && ws.IndoorNavigation.URL != "" {
			item.IndoorNavigationURL = types.StringValue(ws.IndoorNavigation.URL)
		} else {
			item.IndoorNavigationURL = types.StringNull()
		}

		if ws.Calling != nil {
			callingModel := &WorkspaceItemCallingModel{
				Type:     stringOrNull(ws.Calling.Type),
				Licenses: types.ListNull(types.StringType),
			}
			if ws.Calling.WebexCalling != nil {
				callingModel.PhoneNumber = stringOrNull(ws.Calling.WebexCalling.PhoneNumber)
				callingModel.Extension = stringOrNull(ws.Calling.WebexCalling.Extension)
				callingModel.LocationID = stringOrNull(ws.Calling.WebexCalling.LocationID)
				if len(ws.Calling.WebexCalling.Licenses) > 0 {
					callingModel.Licenses, _ = types.ListValueFrom(ctx, types.StringType, ws.Calling.WebexCalling.Licenses)
				}
			}
			item.Calling = callingModel
		}

		if ws.Calendar != nil {
			item.Calendar = &WorkspaceItemCalendarModel{
				Type:            stringOrNull(ws.Calendar.Type),
				EmailAddress:    stringOrNull(ws.Calendar.EmailAddress),
				ResourceGroupID: stringOrNull(ws.Calendar.ResourceGroupID),
			}
		}

		if ws.DeviceHostedMeetings != nil {
			item.DeviceHostedMeetings = &WorkspaceItemDeviceHostedMeetingsModel{
				Enabled: types.BoolValue(ws.DeviceHostedMeetings.Enabled),
				SiteURL: stringOrNull(ws.DeviceHostedMeetings.SiteURL),
			}
		}

		config.Workspaces[i] = item
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
