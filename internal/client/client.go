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
	"strings"
	"time"
)

const (
	// DefaultMaxRetries is the number of retries attempted after the initial
	// request. New applies it; a Client built directly does not retry unless
	// it sets MaxRetries.
	DefaultMaxRetries = 5
	DefaultBaseDelay  = time.Second
	DefaultMaxDelay   = 30 * time.Second

	defaultTimeout = 30 * time.Second
)

// tokenEndpoint is a variable rather than a constant so the tests can point
// the exchange at a local server.
var tokenEndpoint = "https://webexapis.com/v1/access_token"

var tokenHTTPClient = &http.Client{Timeout: defaultTimeout}

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client

	// MaxRetries bounds the retries that follow a transient failure. Zero
	// disables retrying entirely; New sets DefaultMaxRetries.
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
			Timeout: defaultTimeout,
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

// FetchAccessToken exchanges a refresh token for an access token.
//
// Deprecated: prefer FetchAccessTokenContext, which can be cancelled.
func FetchAccessToken(clientID, clientSecret, refreshToken string) (string, error) {
	return FetchAccessTokenContext(context.Background(), clientID, clientSecret, refreshToken)
}

// FetchAccessTokenContext exchanges a refresh token for an access token,
// retrying transient failures the way ordinary API calls do. This is the first
// request of every run, so without retrying, one rate-limited exchange fails
// the whole apply before a single resource has been read.
//
// The exchange is a POST but replaying it is safe: a failed exchange creates
// nothing, and the refresh token is not consumed.
func FetchAccessTokenContext(ctx context.Context, clientID, clientSecret, refreshToken string) (string, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
	}

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(data.Encode()))
		if err != nil {
			return "", fmt.Errorf("creating token request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := tokenHTTPClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			if attempt >= DefaultMaxRetries {
				return "", fmt.Errorf("requesting access token: %w", err)
			}
			if err := sleep(ctx, computeBackoff(attempt, "", DefaultBaseDelay, DefaultMaxDelay)); err != nil {
				return "", err
			}
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		status := resp.StatusCode
		retryAfter := resp.Header.Get("Retry-After")
		_ = resp.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("reading token response: %w", readErr)
		}

		if isRetryable(status) && attempt < DefaultMaxRetries {
			if err := sleep(ctx, computeBackoff(attempt, retryAfter, DefaultBaseDelay, DefaultMaxDelay)); err != nil {
				return "", err
			}
			continue
		}

		if status != http.StatusOK {
			return "", fmt.Errorf("failed to obtain access token (HTTP %d): %s\n\n"+
				"This is likely caused by an expired refresh token (refresh tokens expire after 90 days).\n"+
				"To fix this:\n"+
				"  1. Go to your Service App page on developer.webex.com\n"+
				"  2. Under 'Org Authorizations', select your org\n"+
				"  3. Enter your Client Secret and click 'Generate Tokens'\n"+
				"  4. Copy the new refresh token and update your configuration",
				status, string(body))
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
	return c.send(ctx, method, c.url(path, params), "application/json", raw, isIdempotent(method), result)
}

func (c *Client) doPatch(ctx context.Context, path string, params url.Values, body interface{}, result interface{}) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}
	// PATCH is not idempotent in general, but every patch this provider sends
	// is a "replace" or "remove" against a known path, which is. Adding an op
	// that appends rather than sets would invalidate this and must pass false.
	return c.send(ctx, http.MethodPatch, c.url(path, params), "application/json-patch+json", raw, true, result)
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
//
// replayable says whether the request may be sent again after a failure that
// leaves the outcome unknown - a 502/503/504 or a dropped connection, where
// the server may well have carried the request out and only the response was
// lost. Replaying a create in that state produces a second device or
// workspace, so those failures are retried only when the caller says the
// request is safe to repeat. A 429 is exempt: it is a refusal to process, so
// nothing can have happened, and it is retried whatever the method.
func (c *Client) send(ctx context.Context, method, u, contentType string, body []byte, replayable bool, result interface{}) error {
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
			if !replayable || attempt >= c.retries() {
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

		if c.shouldRetry(status, replayable, attempt) {
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

func (c *Client) shouldRetry(status int, replayable bool, attempt int) bool {
	if attempt >= c.retries() {
		return false
	}
	if status == http.StatusTooManyRequests {
		return true
	}
	return replayable && isRetryable(status)
}

// isRetryable reports whether a status is transient. Callers must still decide
// whether the request itself is safe to repeat; see send.
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

// isIdempotent follows RFC 9110: repeating one of these has the same effect as
// issuing it once, so it is safe to replay when the outcome is unknown. POST
// is deliberately absent - a replayed create makes a duplicate.
func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut,
		http.MethodDelete, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}

func (c *Client) retries() int {
	if c.MaxRetries < 0 {
		return 0
	}
	return c.MaxRetries
}

// backoff reports how long to wait before the next attempt.
func (c *Client) backoff(attempt int, retryAfter string) time.Duration {
	base := c.baseDelay
	if base <= 0 {
		base = DefaultBaseDelay
	}
	limit := c.maxDelay
	if limit <= 0 {
		limit = DefaultMaxDelay
	}
	return computeBackoff(attempt, retryAfter, base, limit)
}

// computeBackoff honours Retry-After when the server sends it, since it knows
// when the window reopens; otherwise the delay doubles. The jitter matters more
// than it looks: resources applied in parallel are throttled at the same
// instant, so without it they would all wake together and trip the same limit
// again.
func computeBackoff(attempt int, retryAfter string, base, limit time.Duration) time.Duration {
	if d, ok := parseRetryAfter(retryAfter, limit); ok {
		return d
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
