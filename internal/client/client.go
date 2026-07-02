package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func New(token string, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://webexapis.com/v1"
	}
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type APIError struct {
	StatusCode int
	Message    string
	TrackingID string
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("webex API error (HTTP %d, tracking: %s): %s", e.StatusCode, e.TrackingID, e.Message)
	if e.StatusCode == http.StatusUnauthorized {
		msg += "\n\nThis is likely caused by an expired access token. Token lifetimes:\n" +
			"  - Developer portal temporary token: 12 hours\n" +
			"  - Service App access token: 14 days\n" +
			"  - Service App refresh token: 90 days\n" +
			"If using a static token, regenerate it from the Webex Developer Portal.\n" +
			"If using OAuth (client_id + client_secret + refresh_token), your refresh token may have expired — regenerate it from your Service App's Org Authorizations page."
	}
	return msg
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func FetchAccessToken(clientID, clientSecret, refreshToken string) (string, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
	}

	resp, err := http.PostForm("https://webexapis.com/v1/access_token", data)
	if err != nil {
		return "", fmt.Errorf("requesting access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to obtain access token (HTTP %d): %s\n\n"+
			"This is likely caused by an expired refresh token (refresh tokens expire after 90 days).\n"+
			"To fix this:\n"+
			"  1. Go to your Service App page on developer.webex.com\n"+
			"  2. Under 'Org Authorizations', select your org\n"+
			"  3. Enter your Client Secret and click 'Generate Tokens'\n"+
			"  4. Copy the new refresh token and update your configuration",
			resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("token response did not contain an access token")
	}

	return tokenResp.AccessToken, nil
}

func (c *Client) do(ctx context.Context, method, path string, params url.Values, body interface{}, result interface{}) error {
	u := c.BaseURL + "/" + path
	if params != nil {
		u += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := 0; attempt <= 3; attempt++ {
		resp, err = c.HTTPClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if attempt < 1 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
				}
				if body != nil {
					b, _ := json.Marshal(body)
					bodyReader = bytes.NewReader(b)
				}
				req, _ = http.NewRequestWithContext(ctx, method, u, bodyReader)
				req.Header.Set("Authorization", "Bearer "+c.Token)
				req.Header.Set("Content-Type", "application/json")
				continue
			}
			return fmt.Errorf("executing request: %w", err)
		}
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusBadGateway &&
			resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusGatewayTimeout {
			break
		}
		_ = resp.Body.Close()
		delay := time.Duration(1<<uint(attempt)) * time.Second
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if seconds, parseErr := strconv.Atoi(ra); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		if body != nil {
			b, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(b)
		}
		req, _ = http.NewRequestWithContext(ctx, method, u, bodyReader)
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("Content-Type", "application/json")
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
			TrackingID: resp.Header.Get("TrackingID"),
		}
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

func (c *Client) doPatch(ctx context.Context, path string, params url.Values, body interface{}, result interface{}) error {
	u := c.BaseURL + "/" + path
	if params != nil {
		u += "?" + params.Encode()
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json-patch+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
			TrackingID: resp.Header.Get("TrackingID"),
		}
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}
	return nil
}

func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}
