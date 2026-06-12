package client

import (
	"context"
	"net/http"
	"net/url"
)

type LocationAddress struct {
	Address1   string `json:"address1,omitempty"`
	Address2   string `json:"address2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	Country    string `json:"country,omitempty"`
}

type Location struct {
	ID        string           `json:"id,omitempty"`
	Name      string           `json:"name,omitempty"`
	OrgID     string           `json:"orgId,omitempty"`
	Address   *LocationAddress `json:"address,omitempty"`
	TimeZone  string           `json:"timeZone,omitempty"`
	Latitude  float64          `json:"latitude,omitempty"`
	Longitude float64          `json:"longitude,omitempty"`
}

type LocationListResponse struct {
	Items []Location `json:"items"`
}

func (c *Client) ListLocations(ctx context.Context, name string) ([]Location, error) {
	params := url.Values{}
	if name != "" {
		params.Set("name", name)
	}
	var resp LocationListResponse
	err := c.do(ctx, http.MethodGet, "locations", params, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

type LocationRequest struct {
	Name      string           `json:"name"`
	Address   *LocationAddress `json:"address,omitempty"`
	TimeZone  string           `json:"timeZone,omitempty"`
	Latitude  float64          `json:"latitude,omitempty"`
	Longitude float64          `json:"longitude,omitempty"`
}

func (c *Client) GetLocation(ctx context.Context, id string) (*Location, error) {
	var loc Location
	err := c.do(ctx, http.MethodGet, "locations/"+id, nil, nil, &loc)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (c *Client) CreateLocation(ctx context.Context, req *LocationRequest) (*Location, error) {
	var loc Location
	err := c.do(ctx, http.MethodPost, "locations", nil, req, &loc)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (c *Client) UpdateLocation(ctx context.Context, id string, req *LocationRequest) (*Location, error) {
	var loc Location
	err := c.do(ctx, http.MethodPut, "locations/"+id, nil, req, &loc)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (c *Client) DeleteLocation(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "locations/"+id, nil, nil, nil)
}
