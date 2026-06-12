package client

import (
	"context"
	"net/http"
	"net/url"
)

type ConfigurationSource struct {
	Value       interface{}            `json:"value"`
	Editability map[string]interface{} `json:"editability,omitempty"`
}

type ConfigurationValueSpace struct {
	Type      string   `json:"type,omitempty"`
	Minimum   *int     `json:"minimum,omitempty"`
	Maximum   *int     `json:"maximum,omitempty"`
	MinLength *int     `json:"minLength,omitempty"`
	MaxLength *int     `json:"maxLength,omitempty"`
	Enum      []string `json:"enum,omitempty"`
}

type DeviceConfiguration struct {
	Value      interface{}                    `json:"value"`
	Source     string                         `json:"source"`
	Sources    map[string]ConfigurationSource `json:"sources,omitempty"`
	ValueSpace *ConfigurationValueSpace       `json:"valueSpace,omitempty"`
}

type DeviceConfigurationResponse struct {
	DeviceID string                         `json:"deviceId"`
	Items    map[string]DeviceConfiguration `json:"items"`
}

type ConfigurationPatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func (c *Client) GetDeviceConfigurations(ctx context.Context, deviceID string, key string) (*DeviceConfigurationResponse, error) {
	params := url.Values{}
	params.Set("deviceId", deviceID)
	if key != "" {
		params.Set("key", key)
	}
	var resp DeviceConfigurationResponse
	err := c.do(ctx, http.MethodGet, "deviceConfigurations", params, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateDeviceConfigurations(ctx context.Context, deviceID string, ops []ConfigurationPatchOp) (*DeviceConfigurationResponse, error) {
	params := url.Values{}
	params.Set("deviceId", deviceID)
	var resp DeviceConfigurationResponse
	err := c.doPatch(ctx, "deviceConfigurations", params, ops, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
