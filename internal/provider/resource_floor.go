package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var (
	_ resource.Resource                = &FloorResource{}
	_ resource.ResourceWithImportState = &FloorResource{}
)

type FloorResource struct {
	client *client.Client
}

type FloorResourceModel struct {
	ID          types.String `tfsdk:"id"`
	LocationID  types.String `tfsdk:"location_id"`
	FloorNumber types.Int64  `tfsdk:"floor_number"`
	DisplayName types.String `tfsdk:"display_name"`
}

func NewFloorResource() resource.Resource {
	return &FloorResource{}
}

func (r *FloorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floor"
}

func (r *FloorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a floor within a Webex location.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the floor.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"location_id": schema.StringAttribute{
				Required:    true,
				Description: "The location this floor belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"floor_number": schema.Int64Attribute{
				Required:    true,
				Description: "The floor number. Can be negative (e.g. -1 for basement).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Display name for the floor.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *FloorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *FloorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FloorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	floorReq := &client.FloorRequest{
		FloorNumber: int(plan.FloorNumber.ValueInt64()),
		DisplayName: plan.DisplayName.ValueString(),
	}

	floor, err := r.client.CreateFloor(ctx, plan.LocationID.ValueString(), floorReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating floor", err.Error())
		return
	}

	r.mapToState(floor, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FloorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FloorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	floor, err := r.client.GetFloor(ctx, state.LocationID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading floor", err.Error())
		return
	}

	r.mapToState(floor, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FloorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FloorResourceModel
	var state FloorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	floorReq := &client.FloorRequest{
		FloorNumber: int(plan.FloorNumber.ValueInt64()),
		DisplayName: plan.DisplayName.ValueString(),
	}

	floor, err := r.client.UpdateFloor(ctx, state.LocationID.ValueString(), state.ID.ValueString(), floorReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating floor", err.Error())
		return
	}

	r.mapToState(floor, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FloorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FloorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFloor(ctx, state.LocationID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting floor", err.Error())
	}
}

func (r *FloorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: location_id/floor_id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("location_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *FloorResource) mapToState(f *client.Floor, model *FloorResourceModel) {
	model.ID = types.StringValue(f.ID)
	model.LocationID = types.StringValue(f.LocationID)
	model.FloorNumber = types.Int64Value(int64(f.FloorNumber))
	model.DisplayName = stringOrNull(f.DisplayName)
}
