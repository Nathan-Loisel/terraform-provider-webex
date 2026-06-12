package client

import (
	"context"
	"net/http"
	"net/url"
)

type CallingConfig struct {
	Type         string              `json:"type,omitempty"`
	WebexCalling *WebexCallingConfig `json:"webexCalling,omitempty"`
}

type WebexCallingConfig struct {
	PhoneNumber string   `json:"phoneNumber,omitempty"`
	Extension   string   `json:"extension,omitempty"`
	LocationID  string   `json:"locationId,omitempty"`
	Licenses    []string `json:"licenses,omitempty"`
}

type CalendarConfig struct {
	Type            string `json:"type,omitempty"`
	EmailAddress    string `json:"emailAddress,omitempty"`
	ResourceGroupID string `json:"resourceGroupId,omitempty"`
}

type DeviceHostedMeetingsConfig struct {
	Enabled bool   `json:"enabled,omitempty"`
	SiteURL string `json:"siteUrl,omitempty"`
}

type IndoorNavigationConfig struct {
	URL string `json:"url,omitempty"`
}

type Workspace struct {
	ID                   string                      `json:"id,omitempty"`
	OrgID                string                      `json:"orgId,omitempty"`
	LocationID           string                      `json:"locationId,omitempty"`
	FloorID              string                      `json:"floorId,omitempty"`
	DisplayName          string                      `json:"displayName,omitempty"`
	Capacity             *int                        `json:"capacity,omitempty"`
	Type                 string                      `json:"type,omitempty"`
	SIPAddress           string                      `json:"sipAddress,omitempty"`
	Created              string                      `json:"created,omitempty"`
	Calling              *CallingConfig              `json:"calling,omitempty"`
	Calendar             *CalendarConfig             `json:"calendar,omitempty"`
	Notes                string                      `json:"notes,omitempty"`
	HotdeskingStatus     string                      `json:"hotdeskingStatus,omitempty"`
	SupportedDevices     string                      `json:"supportedDevices,omitempty"`
	DeviceHostedMeetings *DeviceHostedMeetingsConfig `json:"deviceHostedMeetings,omitempty"`
	DevicePlatform       string                      `json:"devicePlatform,omitempty"`
	IndoorNavigation     *IndoorNavigationConfig     `json:"indoorNavigation,omitempty"`
}

type WorkspaceListResponse struct {
	Items []Workspace `json:"items"`
}

type WorkspaceCreateRequest struct {
	DisplayName          string                      `json:"displayName"`
	OrgID                string                      `json:"orgId,omitempty"`
	LocationID           string                      `json:"locationId,omitempty"`
	FloorID              string                      `json:"floorId,omitempty"`
	Capacity             *int                        `json:"capacity,omitempty"`
	Type                 string                      `json:"type,omitempty"`
	SIPAddress           string                      `json:"sipAddress,omitempty"`
	Calling              *CallingConfig              `json:"calling,omitempty"`
	Calendar             *CalendarConfig             `json:"calendar,omitempty"`
	Notes                string                      `json:"notes,omitempty"`
	HotdeskingStatus     string                      `json:"hotdeskingStatus,omitempty"`
	SupportedDevices     string                      `json:"supportedDevices,omitempty"`
	DeviceHostedMeetings *DeviceHostedMeetingsConfig `json:"deviceHostedMeetings,omitempty"`
	IndoorNavigation     *IndoorNavigationConfig     `json:"indoorNavigation,omitempty"`
}

type WorkspaceUpdateRequest struct {
	DisplayName          string                      `json:"displayName,omitempty"`
	LocationID           string                      `json:"locationId,omitempty"`
	FloorID              string                      `json:"floorId,omitempty"`
	Capacity             *int                        `json:"capacity,omitempty"`
	Type                 string                      `json:"type,omitempty"`
	SIPAddress           string                      `json:"sipAddress,omitempty"`
	Calling              *CallingConfig              `json:"calling,omitempty"`
	Calendar             *CalendarConfig             `json:"calendar,omitempty"`
	Notes                string                      `json:"notes,omitempty"`
	HotdeskingStatus     string                      `json:"hotdeskingStatus,omitempty"`
	DeviceHostedMeetings *DeviceHostedMeetingsConfig `json:"deviceHostedMeetings,omitempty"`
	IndoorNavigation     *IndoorNavigationConfig     `json:"indoorNavigation,omitempty"`
}

func (c *Client) ListWorkspaces(ctx context.Context, displayName string) ([]Workspace, error) {
	params := url.Values{}
	if displayName != "" {
		params.Set("displayName", displayName)
	}
	var resp WorkspaceListResponse
	err := c.do(ctx, http.MethodGet, "workspaces", params, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) CreateWorkspace(ctx context.Context, req *WorkspaceCreateRequest) (*Workspace, error) {
	var ws Workspace
	err := c.do(ctx, http.MethodPost, "workspaces", nil, req, &ws)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) GetWorkspace(ctx context.Context, id string) (*Workspace, error) {
	var ws Workspace
	err := c.do(ctx, http.MethodGet, "workspaces/"+id, nil, nil, &ws)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) UpdateWorkspace(ctx context.Context, id string, req *WorkspaceUpdateRequest) (*Workspace, error) {
	var ws Workspace
	err := c.do(ctx, http.MethodPut, "workspaces/"+id, nil, req, &ws)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) DeleteWorkspace(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "workspaces/"+id, nil, nil, nil)
}
