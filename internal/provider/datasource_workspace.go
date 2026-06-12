package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var _ datasource.DataSource = &WorkspaceDataSource{}

type WorkspaceDataSource struct {
	client *client.Client
}

type WorkspaceDSCallingModel struct {
	Type        types.String `tfsdk:"type"`
	PhoneNumber types.String `tfsdk:"phone_number"`
	Extension   types.String `tfsdk:"extension"`
	LocationID  types.String `tfsdk:"location_id"`
	Licenses    types.List   `tfsdk:"licenses"`
}

type WorkspaceDSCalendarModel struct {
	Type            types.String `tfsdk:"type"`
	EmailAddress    types.String `tfsdk:"email_address"`
	ResourceGroupID types.String `tfsdk:"resource_group_id"`
}

type WorkspaceDSDeviceHostedMeetingsModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	SiteURL types.String `tfsdk:"site_url"`
}

type WorkspaceDataSourceModel struct {
	ID                   types.String                          `tfsdk:"id"`
	DisplayName          types.String                          `tfsdk:"display_name"`
	OrgID                types.String                          `tfsdk:"org_id"`
	LocationID           types.String                          `tfsdk:"location_id"`
	FloorID              types.String                          `tfsdk:"floor_id"`
	Capacity             types.Int64                           `tfsdk:"capacity"`
	Type                 types.String                          `tfsdk:"type"`
	SIPAddress           types.String                          `tfsdk:"sip_address"`
	Created              types.String                          `tfsdk:"created"`
	Notes                types.String                          `tfsdk:"notes"`
	HotdeskingStatus     types.String                          `tfsdk:"hotdesking_status"`
	SupportedDevices     types.String                          `tfsdk:"supported_devices"`
	DevicePlatform       types.String                          `tfsdk:"device_platform"`
	IndoorNavigationURL  types.String                          `tfsdk:"indoor_navigation_url"`
	Calling              *WorkspaceDSCallingModel              `tfsdk:"calling"`
	Calendar             *WorkspaceDSCalendarModel             `tfsdk:"calendar"`
	DeviceHostedMeetings *WorkspaceDSDeviceHostedMeetingsModel `tfsdk:"device_hosted_meetings"`
}

func NewWorkspaceDataSource() datasource.DataSource {
	return &WorkspaceDataSource{}
}

func (d *WorkspaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *WorkspaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a Webex Workspace by display name.",
		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the workspace to look up.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Workspace ID.",
			},
			"org_id": schema.StringAttribute{
				Computed:    true,
				Description: "Organization ID.",
			},
			"location_id": schema.StringAttribute{
				Computed:    true,
				Description: "Location ID.",
			},
			"floor_id": schema.StringAttribute{
				Computed:    true,
				Description: "Floor ID.",
			},
			"capacity": schema.Int64Attribute{
				Computed:    true,
				Description: "Room capacity.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Workspace type.",
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
				Description: "Device platform: cisco, microsoftTeamsRoom.",
			},
			"indoor_navigation_url": schema.StringAttribute{
				Computed:    true,
				Description: "URL of a map locating the workspace.",
			},
		},
		Blocks: map[string]schema.Block{
			"calling": schema.SingleNestedBlock{
				Description: "Calling configuration.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:    true,
						Description: "Calling type.",
					},
					"phone_number": schema.StringAttribute{
						Computed:    true,
						Description: "Phone number (webexCalling only).",
					},
					"extension": schema.StringAttribute{
						Computed:    true,
						Description: "Extension (webexCalling only).",
					},
					"location_id": schema.StringAttribute{
						Computed:    true,
						Description: "Calling location ID (webexCalling only).",
					},
					"licenses": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "Webex Calling licenses.",
					},
				},
			},
			"calendar": schema.SingleNestedBlock{
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
			"device_hosted_meetings": schema.SingleNestedBlock{
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
	}
}

func (d *WorkspaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WorkspaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.DisplayName.ValueString()
	workspaces, err := d.client.ListWorkspaces(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error listing workspaces", err.Error())
		return
	}

	var match *client.Workspace
	for i := range workspaces {
		if workspaces[i].DisplayName == name {
			match = &workspaces[i]
			break
		}
	}

	if match == nil {
		resp.Diagnostics.AddError("Workspace not found", fmt.Sprintf("No workspace found with display name %q", name))
		return
	}

	config.ID = types.StringValue(match.ID)
	config.OrgID = types.StringValue(match.OrgID)
	config.DisplayName = types.StringValue(match.DisplayName)
	config.LocationID = stringOrNull(match.LocationID)
	config.FloorID = stringOrNull(match.FloorID)
	config.Type = stringOrNull(match.Type)
	config.SIPAddress = stringOrNull(match.SIPAddress)
	config.Created = stringOrNull(match.Created)
	config.Notes = stringOrNull(match.Notes)
	config.HotdeskingStatus = stringOrNull(match.HotdeskingStatus)
	config.SupportedDevices = stringOrNull(match.SupportedDevices)
	config.DevicePlatform = stringOrNull(match.DevicePlatform)

	if match.Capacity != nil {
		config.Capacity = types.Int64Value(int64(*match.Capacity))
	} else {
		config.Capacity = types.Int64Null()
	}

	if match.IndoorNavigation != nil && match.IndoorNavigation.URL != "" {
		config.IndoorNavigationURL = types.StringValue(match.IndoorNavigation.URL)
	} else {
		config.IndoorNavigationURL = types.StringNull()
	}

	if match.Calling != nil {
		callingModel := &WorkspaceDSCallingModel{
			Type:     stringOrNull(match.Calling.Type),
			Licenses: types.ListNull(types.StringType),
		}
		if match.Calling.WebexCalling != nil {
			callingModel.PhoneNumber = stringOrNull(match.Calling.WebexCalling.PhoneNumber)
			callingModel.Extension = stringOrNull(match.Calling.WebexCalling.Extension)
			callingModel.LocationID = stringOrNull(match.Calling.WebexCalling.LocationID)
			if len(match.Calling.WebexCalling.Licenses) > 0 {
				callingModel.Licenses, _ = types.ListValueFrom(ctx, types.StringType, match.Calling.WebexCalling.Licenses)
			}
		}
		config.Calling = callingModel
	}

	if match.Calendar != nil {
		config.Calendar = &WorkspaceDSCalendarModel{
			Type:            stringOrNull(match.Calendar.Type),
			EmailAddress:    stringOrNull(match.Calendar.EmailAddress),
			ResourceGroupID: stringOrNull(match.Calendar.ResourceGroupID),
		}
	}

	if match.DeviceHostedMeetings != nil {
		config.DeviceHostedMeetings = &WorkspaceDSDeviceHostedMeetingsModel{
			Enabled: types.BoolValue(match.DeviceHostedMeetings.Enabled),
			SiteURL: stringOrNull(match.DeviceHostedMeetings.SiteURL),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
