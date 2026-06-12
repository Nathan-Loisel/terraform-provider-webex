package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var (
	_ resource.Resource                = &LocationResource{}
	_ resource.ResourceWithImportState = &LocationResource{}
)

type LocationResource struct {
	client *client.Client
}

type LocationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	OrgID     types.String `tfsdk:"org_id"`
	Name      types.String `tfsdk:"name"`
	Address1  types.String `tfsdk:"address1"`
	Address2  types.String `tfsdk:"address2"`
	City      types.String `tfsdk:"city"`
	State     types.String `tfsdk:"state"`
	PostCode  types.String `tfsdk:"post_code"`
	Country   types.String `tfsdk:"country"`
	TimeZone  types.String `tfsdk:"time_zone"`
	Latitude  types.String `tfsdk:"latitude"`
	Longitude types.String `tfsdk:"longitude"`
}

func NewLocationResource() resource.Resource {
	return &LocationResource{}
}

func (r *LocationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

func (r *LocationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Webex location.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the location.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Computed:    true,
				Description: "The organization ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the location.",
			},
			"address1": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The primary street address.",
			},
			"address2": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Additional address line.",
			},
			"city": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The city.",
			},
			"state": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The state or region.",
			},
			"post_code": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The postal code.",
			},
			"country": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The country code (ISO 3166-1).",
			},
			"time_zone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The time zone (e.g. Europe/Oslo).",
			},
			"latitude": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The latitude.",
			},
			"longitude": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The longitude.",
			},
		},
	}
}

func (r *LocationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LocationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := r.buildRequest(&plan)
	loc, err := r.client.CreateLocation(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating location", err.Error())
		return
	}

	r.mapToState(loc, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc, err := r.client.GetLocation(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading location", err.Error())
		return
	}

	r.mapToState(loc, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LocationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LocationResourceModel
	var state LocationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := r.buildRequest(&plan)
	loc, err := r.client.UpdateLocation(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating location", err.Error())
		return
	}

	r.mapToState(loc, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLocation(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting location", err.Error())
	}
}

func (r *LocationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *LocationResource) buildRequest(model *LocationResourceModel) *client.LocationRequest {
	lr := &client.LocationRequest{
		Name: model.Name.ValueString(),
	}

	if !model.TimeZone.IsNull() && !model.TimeZone.IsUnknown() {
		lr.TimeZone = model.TimeZone.ValueString()
	}

	if !model.Latitude.IsNull() && !model.Latitude.IsUnknown() {
		if v, err := strconv.ParseFloat(model.Latitude.ValueString(), 64); err == nil {
			lr.Latitude = v
		}
	}
	if !model.Longitude.IsNull() && !model.Longitude.IsUnknown() {
		if v, err := strconv.ParseFloat(model.Longitude.ValueString(), 64); err == nil {
			lr.Longitude = v
		}
	}

	addr := &client.LocationAddress{}
	hasAddr := false
	if !model.Address1.IsNull() && !model.Address1.IsUnknown() {
		addr.Address1 = model.Address1.ValueString()
		hasAddr = true
	}
	if !model.Address2.IsNull() && !model.Address2.IsUnknown() {
		addr.Address2 = model.Address2.ValueString()
		hasAddr = true
	}
	if !model.City.IsNull() && !model.City.IsUnknown() {
		addr.City = model.City.ValueString()
		hasAddr = true
	}
	if !model.State.IsNull() && !model.State.IsUnknown() {
		addr.State = model.State.ValueString()
		hasAddr = true
	}
	if !model.PostCode.IsNull() && !model.PostCode.IsUnknown() {
		addr.PostalCode = model.PostCode.ValueString()
		hasAddr = true
	}
	if !model.Country.IsNull() && !model.Country.IsUnknown() {
		addr.Country = model.Country.ValueString()
		hasAddr = true
	}
	if hasAddr {
		lr.Address = addr
	}

	return lr
}

func (r *LocationResource) mapToState(loc *client.Location, model *LocationResourceModel) {
	model.ID = types.StringValue(loc.ID)
	model.OrgID = stringOrNull(loc.OrgID)
	model.Name = types.StringValue(loc.Name)
	model.TimeZone = stringOrNull(loc.TimeZone)

	if loc.Latitude != 0 {
		model.Latitude = types.StringValue(strconv.FormatFloat(loc.Latitude, 'f', -1, 64))
	} else {
		model.Latitude = types.StringNull()
	}
	if loc.Longitude != 0 {
		model.Longitude = types.StringValue(strconv.FormatFloat(loc.Longitude, 'f', -1, 64))
	} else {
		model.Longitude = types.StringNull()
	}

	if loc.Address != nil {
		model.Address1 = stringOrNull(loc.Address.Address1)
		model.Address2 = stringOrNull(loc.Address.Address2)
		model.City = stringOrNull(loc.Address.City)
		model.State = stringOrNull(loc.Address.State)
		model.PostCode = stringOrNull(loc.Address.PostalCode)
		model.Country = stringOrNull(loc.Address.Country)
	} else {
		model.Address1 = types.StringNull()
		model.Address2 = types.StringNull()
		model.City = types.StringNull()
		model.State = types.StringNull()
		model.PostCode = types.StringNull()
		model.Country = types.StringNull()
	}
}
