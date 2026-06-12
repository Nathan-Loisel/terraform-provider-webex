package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var (
	_ resource.Resource                = &WorkspaceResource{}
	_ resource.ResourceWithImportState = &WorkspaceResource{}
)

type WorkspaceResource struct {
	client *client.Client
}

type WorkspaceCallingModel struct {
	Type        types.String `tfsdk:"type"`
	PhoneNumber types.String `tfsdk:"phone_number"`
	Extension   types.String `tfsdk:"extension"`
	LocationID  types.String `tfsdk:"location_id"`
	Licenses    types.List   `tfsdk:"licenses"`
}

type WorkspaceCalendarModel struct {
	Type            types.String `tfsdk:"type"`
	EmailAddress    types.String `tfsdk:"email_address"`
	ResourceGroupID types.String `tfsdk:"resource_group_id"`
}

type WorkspaceDeviceHostedMeetingsModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	SiteURL types.String `tfsdk:"site_url"`
}

type WorkspaceResourceModel struct {
	ID                   types.String                        `tfsdk:"id"`
	OrgID                types.String                        `tfsdk:"org_id"`
	DisplayName          types.String                        `tfsdk:"display_name"`
	LocationID           types.String                        `tfsdk:"location_id"`
	FloorID              types.String                        `tfsdk:"floor_id"`
	Capacity             types.Int64                         `tfsdk:"capacity"`
	Type                 types.String                        `tfsdk:"type"`
	SIPAddress           types.String                        `tfsdk:"sip_address"`
	Created              types.String                        `tfsdk:"created"`
	Notes                types.String                        `tfsdk:"notes"`
	HotdeskingStatus     types.String                        `tfsdk:"hotdesking_status"`
	SupportedDevices     types.String                        `tfsdk:"supported_devices"`
	DevicePlatform       types.String                        `tfsdk:"device_platform"`
	Calling              *WorkspaceCallingModel              `tfsdk:"calling"`
	Calendar             *WorkspaceCalendarModel             `tfsdk:"calendar"`
	DeviceHostedMeetings *WorkspaceDeviceHostedMeetingsModel `tfsdk:"device_hosted_meetings"`
	IndoorNavigationURL  types.String                        `tfsdk:"indoor_navigation_url"`
}

func NewWorkspaceResource() resource.Resource {
	return &WorkspaceResource{}
}

func (r *WorkspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *WorkspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Webex Workspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the workspace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Organization ID. Only admin users of another organization (such as partners) may use this.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "A friendly name for the workspace.",
			},
			"location_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Location associated with the workspace. Cannot be changed once configured.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"floor_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Floor associated with the workspace.",
			},
			"capacity": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "How many people the workspace is suitable for.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Workspace type: notSet, focus, huddle, meetingRoom, open, desk, other.",
			},
			"sip_address": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SIP address for the workspace.",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp in ISO8601.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"notes": schema.StringAttribute{
				Optional:    true,
				Description: "Notes associated with the workspace.",
			},
			"hotdesking_status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Hot desking status: on, off.",
			},
			"supported_devices": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Supported devices: collaborationDevices, phones. Cannot be changed once configured.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"device_platform": schema.StringAttribute{
				Computed:    true,
				Description: "Device platform: cisco, microsoftTeamsRoom.",
			},
			"indoor_navigation_url": schema.StringAttribute{
				Optional:    true,
				Description: "URL of a map locating the workspace.",
			},
		},
		Blocks: map[string]schema.Block{
			"calling": schema.SingleNestedBlock{
				Description: "Calling configuration.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Calling type: freeCalling, hybridCalling, webexCalling, webexEdgeForDevices, thirdPartySipCalling, none.",
					},
					"phone_number": schema.StringAttribute{
						Optional:    true,
						Description: "Phone number (webexCalling only).",
					},
					"extension": schema.StringAttribute{
						Optional:    true,
						Description: "Extension (webexCalling only).",
					},
					"location_id": schema.StringAttribute{
						Optional:    true,
						Description: "Calling location ID (webexCalling only).",
					},
					"licenses": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "Webex Calling licenses assigned to this workspace.",
					},
				},
			},
			"calendar": schema.SingleNestedBlock{
				Description: "Calendar configuration.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Calendar type: none, google, microsoft.",
					},
					"email_address": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Workspace email address.",
					},
					"resource_group_id": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Calendar resource group ID.",
					},
				},
			},
			"device_hosted_meetings": schema.SingleNestedBlock{
				Description: "Device hosted meetings configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:    true,
						Description: "Enable device hosted meetings.",
					},
					"site_url": schema.StringAttribute{
						Optional:    true,
						Description: "Webex site URL for device hosted meetings.",
					},
				},
			},
		},
	}
}

func (r *WorkspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *WorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &client.WorkspaceCreateRequest{
		DisplayName: plan.DisplayName.ValueString(),
	}

	if !plan.OrgID.IsNull() && !plan.OrgID.IsUnknown() {
		createReq.OrgID = plan.OrgID.ValueString()
	}
	if !plan.LocationID.IsNull() && !plan.LocationID.IsUnknown() {
		createReq.LocationID = plan.LocationID.ValueString()
	}
	if !plan.FloorID.IsNull() && !plan.FloorID.IsUnknown() {
		createReq.FloorID = plan.FloorID.ValueString()
	}
	if !plan.Capacity.IsNull() && !plan.Capacity.IsUnknown() {
		cap := int(plan.Capacity.ValueInt64())
		createReq.Capacity = &cap
	}
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		createReq.Type = plan.Type.ValueString()
	}
	if !plan.SIPAddress.IsNull() && !plan.SIPAddress.IsUnknown() {
		createReq.SIPAddress = plan.SIPAddress.ValueString()
	}
	if !plan.Notes.IsNull() {
		createReq.Notes = plan.Notes.ValueString()
	}
	if !plan.HotdeskingStatus.IsNull() && !plan.HotdeskingStatus.IsUnknown() {
		createReq.HotdeskingStatus = plan.HotdeskingStatus.ValueString()
	}
	if !plan.SupportedDevices.IsNull() && !plan.SupportedDevices.IsUnknown() {
		createReq.SupportedDevices = plan.SupportedDevices.ValueString()
	}
	if !plan.IndoorNavigationURL.IsNull() {
		createReq.IndoorNavigation = &client.IndoorNavigationConfig{URL: plan.IndoorNavigationURL.ValueString()}
	}

	if plan.Calling != nil {
		createReq.Calling = &client.CallingConfig{
			Type: plan.Calling.Type.ValueString(),
		}
		if plan.Calling.Type.ValueString() == "webexCalling" {
			createReq.Calling.WebexCalling = &client.WebexCallingConfig{}
			if !plan.Calling.PhoneNumber.IsNull() {
				createReq.Calling.WebexCalling.PhoneNumber = plan.Calling.PhoneNumber.ValueString()
			}
			if !plan.Calling.Extension.IsNull() {
				createReq.Calling.WebexCalling.Extension = plan.Calling.Extension.ValueString()
			}
			if !plan.Calling.LocationID.IsNull() {
				createReq.Calling.WebexCalling.LocationID = plan.Calling.LocationID.ValueString()
			}
		}
	}

	if plan.Calendar != nil {
		createReq.Calendar = &client.CalendarConfig{
			Type: plan.Calendar.Type.ValueString(),
		}
		if !plan.Calendar.EmailAddress.IsNull() {
			createReq.Calendar.EmailAddress = plan.Calendar.EmailAddress.ValueString()
		}
		if !plan.Calendar.ResourceGroupID.IsNull() && !plan.Calendar.ResourceGroupID.IsUnknown() {
			createReq.Calendar.ResourceGroupID = plan.Calendar.ResourceGroupID.ValueString()
		}
	}

	if plan.DeviceHostedMeetings != nil {
		createReq.DeviceHostedMeetings = &client.DeviceHostedMeetingsConfig{
			Enabled: plan.DeviceHostedMeetings.Enabled.ValueBool(),
		}
		if !plan.DeviceHostedMeetings.SiteURL.IsNull() {
			createReq.DeviceHostedMeetings.SiteURL = plan.DeviceHostedMeetings.SiteURL.ValueString()
		}
	}

	ws, err := r.client.CreateWorkspace(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workspace", err.Error())
		return
	}

	mapWorkspaceToState(ws, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, err := r.client.GetWorkspace(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading workspace", err.Error())
		return
	}

	mapWorkspaceToState(ws, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkspaceResourceModel
	var state WorkspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &client.WorkspaceUpdateRequest{
		DisplayName: plan.DisplayName.ValueString(),
	}

	if !plan.LocationID.IsNull() && !plan.LocationID.IsUnknown() {
		updateReq.LocationID = plan.LocationID.ValueString()
	}
	if !plan.FloorID.IsNull() && !plan.FloorID.IsUnknown() {
		updateReq.FloorID = plan.FloorID.ValueString()
	}
	if !plan.Capacity.IsNull() && !plan.Capacity.IsUnknown() {
		cap := int(plan.Capacity.ValueInt64())
		updateReq.Capacity = &cap
	}
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		updateReq.Type = plan.Type.ValueString()
	}
	if !plan.SIPAddress.IsNull() && !plan.SIPAddress.IsUnknown() {
		updateReq.SIPAddress = plan.SIPAddress.ValueString()
	}
	if !plan.Notes.IsNull() {
		updateReq.Notes = plan.Notes.ValueString()
	}
	if !plan.HotdeskingStatus.IsNull() && !plan.HotdeskingStatus.IsUnknown() {
		updateReq.HotdeskingStatus = plan.HotdeskingStatus.ValueString()
	}
	if !plan.IndoorNavigationURL.IsNull() {
		updateReq.IndoorNavigation = &client.IndoorNavigationConfig{URL: plan.IndoorNavigationURL.ValueString()}
	}

	if plan.Calling != nil {
		updateReq.Calling = &client.CallingConfig{
			Type: plan.Calling.Type.ValueString(),
		}
		if plan.Calling.Type.ValueString() == "webexCalling" {
			updateReq.Calling.WebexCalling = &client.WebexCallingConfig{}
			if !plan.Calling.PhoneNumber.IsNull() {
				updateReq.Calling.WebexCalling.PhoneNumber = plan.Calling.PhoneNumber.ValueString()
			}
			if !plan.Calling.Extension.IsNull() {
				updateReq.Calling.WebexCalling.Extension = plan.Calling.Extension.ValueString()
			}
			if !plan.Calling.LocationID.IsNull() {
				updateReq.Calling.WebexCalling.LocationID = plan.Calling.LocationID.ValueString()
			}
		}
	}

	if plan.Calendar != nil {
		updateReq.Calendar = &client.CalendarConfig{
			Type: plan.Calendar.Type.ValueString(),
		}
		if !plan.Calendar.EmailAddress.IsNull() {
			updateReq.Calendar.EmailAddress = plan.Calendar.EmailAddress.ValueString()
		}
		if !plan.Calendar.ResourceGroupID.IsNull() && !plan.Calendar.ResourceGroupID.IsUnknown() {
			updateReq.Calendar.ResourceGroupID = plan.Calendar.ResourceGroupID.ValueString()
		}
	}

	if plan.DeviceHostedMeetings != nil {
		updateReq.DeviceHostedMeetings = &client.DeviceHostedMeetingsConfig{
			Enabled: plan.DeviceHostedMeetings.Enabled.ValueBool(),
		}
		if !plan.DeviceHostedMeetings.SiteURL.IsNull() {
			updateReq.DeviceHostedMeetings.SiteURL = plan.DeviceHostedMeetings.SiteURL.ValueString()
		}
	}

	ws, err := r.client.UpdateWorkspace(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating workspace", err.Error())
		return
	}

	mapWorkspaceToState(ws, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWorkspace(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting workspace", err.Error())
	}
}

func (r *WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapWorkspaceToState(ws *client.Workspace, state *WorkspaceResourceModel) {
	state.ID = types.StringValue(ws.ID)
	state.OrgID = types.StringValue(ws.OrgID)
	state.DisplayName = types.StringValue(ws.DisplayName)
	state.Created = types.StringValue(ws.Created)

	state.LocationID = stringOrNull(ws.LocationID)
	state.FloorID = stringOrNull(ws.FloorID)
	if ws.Capacity != nil {
		state.Capacity = types.Int64Value(int64(*ws.Capacity))
	} else {
		state.Capacity = types.Int64Null()
	}
	state.Type = stringOrNull(ws.Type)
	state.SIPAddress = stringOrNull(ws.SIPAddress)
	state.Notes = stringOrNull(ws.Notes)
	state.HotdeskingStatus = stringOrNull(ws.HotdeskingStatus)
	state.SupportedDevices = stringOrNull(ws.SupportedDevices)
	state.DevicePlatform = stringOrNull(ws.DevicePlatform)
	if ws.IndoorNavigation != nil && ws.IndoorNavigation.URL != "" {
		state.IndoorNavigationURL = types.StringValue(ws.IndoorNavigation.URL)
	} else {
		state.IndoorNavigationURL = types.StringNull()
	}

	if ws.Calling != nil {
		callingModel := &WorkspaceCallingModel{
			Type:     types.StringValue(ws.Calling.Type),
			Licenses: types.ListNull(types.StringType),
		}
		if ws.Calling.WebexCalling != nil {
			if ws.Calling.WebexCalling.PhoneNumber != "" {
				callingModel.PhoneNumber = types.StringValue(ws.Calling.WebexCalling.PhoneNumber)
			}
			if ws.Calling.WebexCalling.Extension != "" {
				callingModel.Extension = types.StringValue(ws.Calling.WebexCalling.Extension)
			}
			if ws.Calling.WebexCalling.LocationID != "" {
				callingModel.LocationID = types.StringValue(ws.Calling.WebexCalling.LocationID)
			}
			if len(ws.Calling.WebexCalling.Licenses) > 0 {
				licenseValues := make([]types.String, len(ws.Calling.WebexCalling.Licenses))
				for i, l := range ws.Calling.WebexCalling.Licenses {
					licenseValues[i] = types.StringValue(l)
				}
				callingModel.Licenses, _ = types.ListValueFrom(context.Background(), types.StringType, ws.Calling.WebexCalling.Licenses)
			}
		}
		state.Calling = callingModel
	}

	if ws.Calendar != nil {
		state.Calendar = &WorkspaceCalendarModel{
			Type: types.StringValue(ws.Calendar.Type),
		}
		if ws.Calendar.EmailAddress != "" {
			state.Calendar.EmailAddress = types.StringValue(ws.Calendar.EmailAddress)
		}
		if ws.Calendar.ResourceGroupID != "" {
			state.Calendar.ResourceGroupID = types.StringValue(ws.Calendar.ResourceGroupID)
		}
	}

	if ws.DeviceHostedMeetings != nil {
		state.DeviceHostedMeetings = &WorkspaceDeviceHostedMeetingsModel{
			Enabled: types.BoolValue(ws.DeviceHostedMeetings.Enabled),
		}
		if ws.DeviceHostedMeetings.SiteURL != "" {
			state.DeviceHostedMeetings.SiteURL = types.StringValue(ws.DeviceHostedMeetings.SiteURL)
		}
	}
}
