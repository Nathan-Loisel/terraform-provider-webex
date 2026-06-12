package client

import (
	"context"
	"net/http"
	"net/url"
)

type Device struct {
	ID               string   `json:"id,omitempty"`
	DisplayName      string   `json:"displayName,omitempty"`
	WorkspaceID      string   `json:"workspaceId,omitempty"`
	PersonID         string   `json:"personId,omitempty"`
	OrgID            string   `json:"orgId,omitempty"`
	Capabilities     []string `json:"capabilities,omitempty"`
	Permissions      []string `json:"permissions,omitempty"`
	ConnectionStatus string   `json:"connectionStatus,omitempty"`
	Product          string   `json:"product,omitempty"`
	Type             string   `json:"type,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	IP               string   `json:"ip,omitempty"`
	MAC              string   `json:"mac,omitempty"`
	Serial           string   `json:"serial,omitempty"`
	Software         string   `json:"software,omitempty"`
	PrimarySIPURL    string   `json:"primarySipUrl,omitempty"`
	UpgradeChannel   string   `json:"upgradeChannel,omitempty"`
	Created          string   `json:"created,omitempty"`
	FirstSeen        string   `json:"firstSeen,omitempty"`
	LastSeen         string   `json:"lastSeen,omitempty"`
	LocationID       string   `json:"locationId,omitempty"`
	ManagedBy        string   `json:"managedBy,omitempty"`
	DevicePlatform   string   `json:"devicePlatform,omitempty"`
}

type DeviceListResponse struct {
	Items []Device `json:"items"`
}

type DeviceCreateRequest struct {
	MAC         string `json:"mac"`
	Model       string `json:"model"`
	WorkspaceID string `json:"workspaceId,omitempty"`
	PersonID    string `json:"personId,omitempty"`
	Password    string `json:"password,omitempty"`
}

type ActivationCodeRequest struct {
	WorkspaceID string `json:"workspaceId,omitempty"`
	PersonID    string `json:"personId,omitempty"`
	Model       string `json:"model,omitempty"`
}

type ActivationCodeResponse struct {
	Code       string `json:"code"`
	ExpiryTime string `json:"expiryTime"`
}

type DeviceTagsPatch struct {
	Op    string   `json:"op"`
	Path  string   `json:"path"`
	Value []string `json:"value,omitempty"`
}

func (c *Client) ListDevices(ctx context.Context, workspaceID string) ([]Device, error) {
	params := url.Values{}
	if workspaceID != "" {
		params.Set("workspaceId", workspaceID)
	}
	var resp DeviceListResponse
	err := c.do(ctx, http.MethodGet, "devices", params, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetDevice(ctx context.Context, id string) (*Device, error) {
	var d Device
	err := c.do(ctx, http.MethodGet, "devices/"+id, nil, nil, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (c *Client) CreateDevice(ctx context.Context, req *DeviceCreateRequest) (*Device, error) {
	var d Device
	err := c.do(ctx, http.MethodPost, "devices", nil, req, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (c *Client) DeleteDevice(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "devices/"+id, nil, nil, nil)
}

func (c *Client) CreateActivationCode(ctx context.Context, req *ActivationCodeRequest) (*ActivationCodeResponse, error) {
	var resp ActivationCodeResponse
	err := c.do(ctx, http.MethodPost, "devices/activationCode", nil, req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ModifyDeviceTags(ctx context.Context, id string, op string, tags []string) (*Device, error) {
	patch := DeviceTagsPatch{
		Op:    op,
		Path:  "tags",
		Value: tags,
	}
	var d Device
	err := c.doPatch(ctx, "devices/"+id, nil, patch, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
