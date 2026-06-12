package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var (
	_ resource.Resource                = &DeviceResource{}
	_ resource.ResourceWithImportState = &DeviceResource{}
)

type DeviceResource struct {
	client *client.Client
}

type DeviceResourceModel struct {
	ID               types.String `tfsdk:"id"`
	DisplayName      types.String `tfsdk:"display_name"`
	WorkspaceID      types.String `tfsdk:"workspace_id"`
	PersonID         types.String `tfsdk:"person_id"`
	OrgID            types.String `tfsdk:"org_id"`
	MAC              types.String `tfsdk:"mac"`
	Model            types.String `tfsdk:"model"`
	Product          types.String `tfsdk:"product"`
	Type             types.String `tfsdk:"type"`
	Serial           types.String `tfsdk:"serial"`
	Software         types.String `tfsdk:"software"`
	IP               types.String `tfsdk:"ip"`
	PrimarySIPURL    types.String `tfsdk:"primary_sip_url"`
	ConnectionStatus types.String `tfsdk:"connection_status"`
	UpgradeChannel   types.String `tfsdk:"upgrade_channel"`
	LocationID       types.String `tfsdk:"location_id"`
	ManagedBy        types.String `tfsdk:"managed_by"`
	DevicePlatform   types.String `tfsdk:"device_platform"`
	Created          types.String `tfsdk:"created"`
	FirstSeen        types.String `tfsdk:"first_seen"`
	LastSeen         types.String `tfsdk:"last_seen"`
	Tags             types.List   `tfsdk:"tags"`
	Password         types.String `tfsdk:"password"`
}

func NewDeviceResource() resource.Resource {
	return &DeviceResource{}
}

func (r *DeviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *DeviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Webex Device. Creates a phone by its MAC address in a workspace or for a person.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Computed:    true,
				Description: "Device display name.",
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				Description: "The workspace to assign the device to. Mutually exclusive with person_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"person_id": schema.StringAttribute{
				Optional:    true,
				Description: "The person to assign the device to. Mutually exclusive with workspace_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_id": schema.StringAttribute{
				Computed:    true,
				Description: "Organization ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mac": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "MAC address of the device. Required when creating a new device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Model of the device. Required when creating a new device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "SIP password for third party devices.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"product": schema.StringAttribute{
				Computed:    true,
				Description: "Product name.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Device type: roomdesk, phone, accessory, webexgo, unknown.",
			},
			"serial": schema.StringAttribute{
				Computed:    true,
				Description: "Serial number.",
			},
			"software": schema.StringAttribute{
				Computed:    true,
				Description: "Software version.",
			},
			"ip": schema.StringAttribute{
				Computed:    true,
				Description: "Current IP address.",
			},
			"primary_sip_url": schema.StringAttribute{
				Computed:    true,
				Description: "Primary SIP URL.",
			},
			"connection_status": schema.StringAttribute{
				Computed:    true,
				Description: "Connection status.",
			},
			"upgrade_channel": schema.StringAttribute{
				Computed:    true,
				Description: "Upgrade channel.",
			},
			"location_id": schema.StringAttribute{
				Computed:    true,
				Description: "Location ID.",
			},
			"managed_by": schema.StringAttribute{
				Computed:    true,
				Description: "Entity managing the device: CISCO, CUSTOMER, PARTNER.",
			},
			"device_platform": schema.StringAttribute{
				Computed:    true,
				Description: "Device platform: cisco, microsoftTeamsRoom.",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"first_seen": schema.StringAttribute{
				Computed:    true,
				Description: "First seen timestamp (ISO 8601).",
			},
			"last_seen": schema.StringAttribute{
				Computed:    true,
				Description: "Last seen timestamp (ISO 8601).",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags assigned to the device.",
			},
		},
	}
}

func (r *DeviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &client.DeviceCreateRequest{
		MAC:   plan.MAC.ValueString(),
		Model: plan.Model.ValueString(),
	}
	if !plan.WorkspaceID.IsNull() {
		createReq.WorkspaceID = plan.WorkspaceID.ValueString()
	}
	if !plan.PersonID.IsNull() {
		createReq.PersonID = plan.PersonID.ValueString()
	}
	if !plan.Password.IsNull() {
		createReq.Password = plan.Password.ValueString()
	}

	device, err := r.client.CreateDevice(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating device", err.Error())
		return
	}

	flattenDevice(ctx, device, &plan, &resp.Diagnostics)
	plan.MAC = types.StringValue(createReq.MAC)
	plan.Model = types.StringValue(createReq.Model)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	device, err := r.client.GetDevice(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading device", err.Error())
		return
	}

	mac := state.MAC
	model := state.Model
	password := state.Password

	flattenDevice(ctx, device, &state, &resp.Diagnostics)
	state.MAC = mac
	state.Model = model
	state.Password = password

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeviceResourceModel
	var state DeviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		_, err := r.client.ModifyDeviceTags(ctx, state.ID.ValueString(), "replace", tags)
		if err != nil {
			resp.Diagnostics.AddError("Error updating device tags", err.Error())
			return
		}
	}

	device, err := r.client.GetDevice(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading device after update", err.Error())
		return
	}

	flattenDevice(ctx, device, &plan, &resp.Diagnostics)
	plan.MAC = state.MAC
	plan.Model = state.Model
	plan.Password = state.Password

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDevice(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting device", err.Error())
	}
}

func (r *DeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func flattenDevice(ctx context.Context, d *client.Device, state *DeviceResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(d.ID)
	state.DisplayName = types.StringValue(d.DisplayName)
	state.OrgID = types.StringValue(d.OrgID)
	state.Created = types.StringValue(d.Created)

	state.Product = stringOrNull(d.Product)
	state.Type = stringOrNull(d.Type)
	state.Serial = stringOrNull(d.Serial)
	state.Software = stringOrNull(d.Software)
	state.IP = stringOrNull(d.IP)
	state.PrimarySIPURL = stringOrNull(d.PrimarySIPURL)
	state.ConnectionStatus = stringOrNull(d.ConnectionStatus)
	state.UpgradeChannel = stringOrNull(d.UpgradeChannel)
	state.LocationID = stringOrNull(d.LocationID)
	state.ManagedBy = stringOrNull(d.ManagedBy)
	state.DevicePlatform = stringOrNull(d.DevicePlatform)
	state.FirstSeen = stringOrNull(d.FirstSeen)
	state.LastSeen = stringOrNull(d.LastSeen)

	if d.WorkspaceID != "" {
		state.WorkspaceID = types.StringValue(d.WorkspaceID)
	}
	if d.PersonID != "" {
		state.PersonID = types.StringValue(d.PersonID)
	}

	if len(d.Tags) > 0 {
		tagValues := make([]types.String, len(d.Tags))
		for i, t := range d.Tags {
			tagValues[i] = types.StringValue(t)
		}
		tagsList, tagDiags := types.ListValueFrom(ctx, types.StringType, d.Tags)
		diags.Append(tagDiags...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
