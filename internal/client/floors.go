package client

import (
	"context"
	"fmt"
	"net/http"
)

type Floor struct {
	ID          string `json:"id,omitempty"`
	LocationID  string `json:"locationId,omitempty"`
	FloorNumber int    `json:"floorNumber"`
	DisplayName string `json:"displayName,omitempty"`
}

type FloorListResponse struct {
	Items []Floor `json:"items"`
}

type FloorRequest struct {
	FloorNumber int    `json:"floorNumber"`
	DisplayName string `json:"displayName,omitempty"`
}

func (c *Client) ListFloors(ctx context.Context, locationID string) ([]Floor, error) {
	var resp FloorListResponse
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("locations/%s/floors", locationID), nil, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetFloor(ctx context.Context, locationID, floorID string) (*Floor, error) {
	var f Floor
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("locations/%s/floors/%s", locationID, floorID), nil, nil, &f)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (c *Client) CreateFloor(ctx context.Context, locationID string, req *FloorRequest) (*Floor, error) {
	var f Floor
	err := c.do(ctx, http.MethodPost, fmt.Sprintf("locations/%s/floors", locationID), nil, req, &f)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (c *Client) UpdateFloor(ctx context.Context, locationID, floorID string, req *FloorRequest) (*Floor, error) {
	var f Floor
	err := c.do(ctx, http.MethodPut, fmt.Sprintf("locations/%s/floors/%s", locationID, floorID), nil, req, &f)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (c *Client) DeleteFloor(ctx context.Context, locationID, floorID string) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("locations/%s/floors/%s", locationID, floorID), nil, nil, nil)
}
