package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var _ datasource.DataSource = &LocationDataSource{}

type LocationDataSource struct {
	client *client.Client
}

type LocationDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	OrgID     types.String `tfsdk:"org_id"`
	Address   types.String `tfsdk:"address"`
	Address2  types.String `tfsdk:"address2"`
	City      types.String `tfsdk:"city"`
	State     types.String `tfsdk:"state"`
	PostCode  types.String `tfsdk:"post_code"`
	Country   types.String `tfsdk:"country"`
	TimeZone  types.String `tfsdk:"time_zone"`
	Latitude  types.String `tfsdk:"latitude"`
	Longitude types.String `tfsdk:"longitude"`
}

func NewLocationDataSource() datasource.DataSource {
	return &LocationDataSource{}
}

func (d *LocationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

func (d *LocationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a Webex location by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the location.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the location to look up.",
			},
			"org_id": schema.StringAttribute{
				Computed:    true,
				Description: "The organization ID.",
			},
			"address": schema.StringAttribute{
				Computed:    true,
				Description: "The street address (line 1).",
			},
			"address2": schema.StringAttribute{
				Computed:    true,
				Description: "The street address (line 2).",
			},
			"city": schema.StringAttribute{
				Computed:    true,
				Description: "The city.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The state or region.",
			},
			"post_code": schema.StringAttribute{
				Computed:    true,
				Description: "The postal code.",
			},
			"country": schema.StringAttribute{
				Computed:    true,
				Description: "The country code (ISO 3166-1).",
			},
			"time_zone": schema.StringAttribute{
				Computed:    true,
				Description: "The time zone.",
			},
			"latitude": schema.StringAttribute{
				Computed:    true,
				Description: "The latitude.",
			},
			"longitude": schema.StringAttribute{
				Computed:    true,
				Description: "The longitude.",
			},
		},
	}
}

func (d *LocationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *LocationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config LocationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	locations, err := d.client.ListLocations(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error listing locations", err.Error())
		return
	}

	var match *client.Location
	for i := range locations {
		if locations[i].Name == name {
			match = &locations[i]
			break
		}
	}

	if match == nil {
		resp.Diagnostics.AddError("Location not found", fmt.Sprintf("No location with exact name %q found", name))
		return
	}

	config.ID = types.StringValue(match.ID)
	config.OrgID = stringOrNull(match.OrgID)
	config.TimeZone = stringOrNull(match.TimeZone)
	if match.Latitude != 0 {
		config.Latitude = types.StringValue(strconv.FormatFloat(match.Latitude, 'f', -1, 64))
	} else {
		config.Latitude = types.StringNull()
	}
	if match.Longitude != 0 {
		config.Longitude = types.StringValue(strconv.FormatFloat(match.Longitude, 'f', -1, 64))
	} else {
		config.Longitude = types.StringNull()
	}

	if match.Address != nil {
		config.Address = stringOrNull(match.Address.Address1)
		config.Address2 = stringOrNull(match.Address.Address2)
		config.City = stringOrNull(match.Address.City)
		config.State = stringOrNull(match.Address.State)
		config.PostCode = stringOrNull(match.Address.PostalCode)
		config.Country = stringOrNull(match.Address.Country)
	} else {
		config.Address = types.StringNull()
		config.Address2 = types.StringNull()
		config.City = types.StringNull()
		config.State = types.StringNull()
		config.PostCode = types.StringNull()
		config.Country = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
