package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	// DefaultMaxRetries is the number of retries after the initial attempt.
	DefaultMaxRetries = 5
	DefaultBaseDelay  = time.Second
	DefaultMaxDelay   = 30 * time.Second
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client

	// MaxRetries bounds the retries that follow a transient failure. Zero
	// means DefaultMaxRetries.
	MaxRetries int

	// Backoff tuning, kept unexported because nothing outside this package
	// has a reason to change it; the tests set it so the suite does not
	// spend real seconds asleep.
	baseDelay time.Duration
	maxDelay  time.Duration
}

func New(token string, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://webexapis.com/v1"
	}
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		MaxRetries: DefaultMaxRetries,
		baseDelay:  DefaultBaseDelay,
		maxDelay:   DefaultMaxDelay,
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
	if e.StatusCode == http.StatusTooManyRequests {
		msg += "\n\nThe request was rate limited and did not succeed after retrying.\n" +
			"Webex applies its quota per organisation, so parallel changes compete for it.\n" +
			"If this recurs, lower Terraform's concurrency, e.g. terraform apply -parallelism=2."
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
	var raw []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		raw = b
	}
	return c.send(ctx, method, c.url(path, params), "application/json", raw, result)
}

func (c *Client) doPatch(ctx context.Context, path string, params url.Values, body interface{}, result interface{}) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}
	return c.send(ctx, http.MethodPatch, c.url(path, params), "application/json-patch+json", raw, result)
}

func (c *Client) url(path string, params url.Values) string {
	u := c.BaseURL + "/" + path
	if params != nil {
		u += "?" + params.Encode()
	}
	return u
}

func (c *Client) newRequest(ctx context.Context, method, u, contentType string, body []byte) (*http.Request, error) {
	// A fresh reader per attempt: a retry cannot replay a body that the
	// previous attempt already consumed.
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, r)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

// send issues a request and retries it while the failure looks transient.
// Rate limiting is the common case rather than the exotic one: every resource
// in an apply shares a single org-wide quota, so a handful of workspaces
// changed in parallel will routinely trip a 429 that means "wait", not "stop".
func (c *Client) send(ctx context.Context, method, u, contentType string, body []byte, result interface{}) error {
	for attempt := 0; ; attempt++ {
		req, err := c.newRequest(ctx, method, u, contentType, body)
		if err != nil {
			return err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if attempt >= c.retries() {
				return fmt.Errorf("executing request: %w", err)
			}
			if err := sleep(ctx, c.backoff(attempt, "")); err != nil {
				return err
			}
			continue
		}

		// Read and close before deciding to retry: an undrained body leaks the
		// connection, and the body is also the only place the API says what
		// actually went wrong, so it has to survive into the error.
		respBody, readErr := io.ReadAll(resp.Body)
		status := resp.StatusCode
		retryAfter := resp.Header.Get("Retry-After")
		trackingID := resp.Header.Get("TrackingID")
		_ = resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("reading response body: %w", readErr)
		}

		if isRetryable(status) && attempt < c.retries() {
			if err := sleep(ctx, c.backoff(attempt, retryAfter)); err != nil {
				return err
			}
			continue
		}

		if status >= 400 {
			return &APIError{
				StatusCode: status,
				Message:    string(respBody),
				TrackingID: trackingID,
			}
		}

		if result != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("unmarshaling response: %w", err)
			}
		}
		return nil
	}
}

func isRetryable(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

func (c *Client) retries() int {
	if c.MaxRetries > 0 {
		return c.MaxRetries
	}
	return DefaultMaxRetries
}

// backoff reports how long to wait before the next attempt. Retry-After wins
// when the server sends it, since it knows when the window reopens; otherwise
// the delay doubles. The jitter matters more than it looks: resources applied
// in parallel are throttled at the same instant, so without it they would all
// wake together and trip the same limit again.
func (c *Client) backoff(attempt int, retryAfter string) time.Duration {
	limit := c.maxDelay
	if limit <= 0 {
		limit = DefaultMaxDelay
	}
	if d, ok := parseRetryAfter(retryAfter, limit); ok {
		return d
	}

	base := c.baseDelay
	if base <= 0 {
		base = DefaultBaseDelay
	}
	delay := base << uint(attempt)
	if delay <= 0 || delay > limit {
		delay = limit
	}
	return delay + time.Duration(rand.Int63n(int64(delay/2)+1))
}

// parseRetryAfter accepts both forms RFC 9110 allows: a count of seconds, or
// an HTTP date.
func parseRetryAfter(v string, limit time.Duration) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(v); err == nil {
		if seconds < 0 {
			return 0, false
		}
		return capDelay(time.Duration(seconds)*time.Second, limit), true
	}
	if t, err := http.ParseTime(v); err == nil {
		return capDelay(time.Until(t), limit), true
	}
	return 0, false
}

func capDelay(d, limit time.Duration) time.Duration {
	if d < 0 {
		return 0
	}
	if d > limit {
		return limit
	}
	return d
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}
