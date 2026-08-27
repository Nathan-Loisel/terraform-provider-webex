package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testClient points at srv and collapses the backoff so the suite runs in
// milliseconds instead of sleeping through real retry delays.
func testClient(srv *httptest.Server) *Client {
	return &Client{
		BaseURL:    srv.URL,
		Token:      "test-token",
		HTTPClient: srv.Client(),
		MaxRetries: DefaultMaxRetries,
		baseDelay:  time.Millisecond,
		maxDelay:   5 * time.Millisecond,
	}
}

func TestDoRetriesRateLimit(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, `{"id":"ok"}`)
	}))
	defer srv.Close()

	var out struct {
		ID string `json:"id"`
	}
	if err := testClient(srv).do(context.Background(), http.MethodGet, "devices", nil, nil, &out); err != nil {
		t.Fatalf("expected the third attempt to succeed, got %v", err)
	}
	if out.ID != "ok" {
		t.Fatalf("body not decoded: %+v", out)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

// The regression this change exists for: UpdateDeviceConfigurations goes
// through doPatch, which used to fire exactly once and surface the 429 to the
// practitioner as a failed apply.
func TestDoPatchRetriesRateLimit(t *testing.T) {
	var calls int32
	var mu sync.Mutex
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(b))
		mu.Unlock()
		if ct := r.Header.Get("Content-Type"); ct != "application/json-patch+json" {
			t.Errorf("wrong content type on retry: %q", ct)
		}
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, `{"deviceId":"d1"}`)
	}))
	defer srv.Close()

	ops := []ConfigurationPatchOp{{Op: "replace", Path: "/x", Value: 1}}
	var out DeviceConfigurationResponse
	if err := testClient(srv).doPatch(context.Background(), "deviceConfigurations", nil, ops, &out); err != nil {
		t.Fatalf("expected doPatch to retry and succeed, got %v", err)
	}
	if out.DeviceID != "d1" {
		t.Fatalf("body not decoded: %+v", out)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
	mu.Lock()
	defer mu.Unlock()
	// Every attempt must carry the payload; a replayed request whose body was
	// already drained would silently PATCH nothing.
	for i, b := range bodies {
		if !strings.Contains(b, `"op":"replace"`) {
			t.Fatalf("attempt %d sent an empty body: %q", i+1, b)
		}
	}
}

// A 502/503/504 leaves the outcome unknown: Webex may have carried the request
// out and lost the response. Replaying a POST there would create a second
// device or workspace.
func TestPostIsNotReplayedWhenOutcomeIsUnknown(t *testing.T) {
	for _, status := range []int{http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var calls int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&calls, 1)
				w.WriteHeader(status)
			}))
			defer srv.Close()

			err := testClient(srv).do(context.Background(), http.MethodPost, "devices", nil, map[string]string{"a": "b"}, nil)
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := atomic.LoadInt32(&calls); got != 1 {
				t.Fatalf("POST was replayed %d times after %d", got-1, status)
			}
		})
	}
}

// A dropped connection is the same unknown-outcome problem as a 502.
func TestPostIsNotReplayedOnNetworkError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		// Hijack and close without writing a response so the client sees a
		// transport error rather than a status code.
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Errorf("hijack: %v", err)
			return
		}
		_ = conn.Close()
	}))
	defer srv.Close()

	if err := testClient(srv).do(context.Background(), http.MethodPost, "devices", nil, map[string]string{"a": "b"}, nil); err == nil {
		t.Fatal("expected an error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("POST was replayed %d times after a network error", got-1)
	}
}

// 429 is the exception: the request was refused outright, so nothing can have
// happened and even a create is safe to send again.
func TestPostIsRetriedOnRateLimit(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, `{"id":"created"}`)
	}))
	defer srv.Close()

	var out struct {
		ID string `json:"id"`
	}
	if err := testClient(srv).do(context.Background(), http.MethodPost, "devices", nil, map[string]string{"a": "b"}, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ID != "created" {
		t.Fatalf("id = %q", out.ID)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestIdempotentMethodsAreReplayedOnServerError(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			var calls int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if atomic.AddInt32(&calls, 1) < 2 {
					w.WriteHeader(http.StatusBadGateway)
					return
				}
				_, _ = io.WriteString(w, `{}`)
			}))
			defer srv.Close()

			if err := testClient(srv).do(context.Background(), method, "devices", nil, nil, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := atomic.LoadInt32(&calls); got != 2 {
				t.Fatalf("%s attempted %d times, want 2", method, got)
			}
		})
	}
}

func TestIsIdempotent(t *testing.T) {
	for _, m := range []string{
		http.MethodGet, http.MethodHead, http.MethodPut,
		http.MethodDelete, http.MethodOptions, http.MethodTrace,
	} {
		if !isIdempotent(m) {
			t.Errorf("%s should be idempotent", m)
		}
	}
	// PATCH is not idempotent in general; doPatch opts in explicitly because
	// every op this provider sends is a replace or a remove.
	for _, m := range []string{http.MethodPost, http.MethodPatch} {
		if isIdempotent(m) {
			t.Errorf("%s must not be treated as idempotent", m)
		}
	}
}

// Exhausting the retries must still produce the API's own explanation. The
// previous loop closed the body before the final read, so the caller got
// "read on closed response body" instead of the 429 and its tracking ID.
func TestExhaustedRetriesReturnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("TrackingID", "ROUTERGW_abc")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"message":"GET failed: 429 Too Many Requests"}`)
	}))
	defer srv.Close()

	c := testClient(srv)
	c.MaxRetries = 2
	err := c.do(context.Background(), http.MethodGet, "devices", nil, nil, nil)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
	if apiErr.TrackingID != "ROUTERGW_abc" {
		t.Fatalf("tracking id lost: %q", apiErr.TrackingID)
	}
	if !strings.Contains(apiErr.Message, "Too Many Requests") {
		t.Fatalf("response body lost: %q", apiErr.Message)
	}
	if !strings.Contains(apiErr.Error(), "-parallelism") {
		t.Fatalf("429 hint missing from: %s", apiErr.Error())
	}
}

func TestNonRetryableStatusFailsImmediately(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var calls int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&calls, 1)
				w.WriteHeader(status)
			}))
			defer srv.Close()

			err := testClient(srv).do(context.Background(), http.MethodGet, "devices", nil, nil, nil)
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := atomic.LoadInt32(&calls); got != 1 {
				t.Fatalf("status %d was retried %d times", status, got-1)
			}
			if status == http.StatusNotFound && !IsNotFound(err) {
				t.Fatal("IsNotFound stopped recognising 404")
			}
		})
	}
}

// MaxRetries is a real switch: zero means one attempt, and a negative value is
// clamped rather than read as "retry forever".
func TestMaxRetriesZeroDisablesRetrying(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := testClient(srv)
	for _, n := range []int{0, -1} {
		atomic.StoreInt32(&calls, 0)
		c.MaxRetries = n
		if err := c.do(context.Background(), http.MethodGet, "devices", nil, nil, nil); err == nil {
			t.Fatalf("MaxRetries=%d: expected an error", n)
		}
		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Fatalf("MaxRetries=%d made %d attempts, want 1", n, got)
		}
	}
}

func TestRetryAfterHeaderIsHonoured(t *testing.T) {
	var mu sync.Mutex
	var first time.Time
	var gap time.Duration
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if atomic.AddInt32(&calls, 1) == 1 {
			first = time.Now()
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		gap = time.Since(first)
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	c := testClient(srv)
	c.maxDelay = 2 * time.Second // do not cap the header below its own value
	if err := c.do(context.Background(), http.MethodGet, "devices", nil, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The header asked for a second; the computed backoff would have been ~1ms.
	mu.Lock()
	defer mu.Unlock()
	if gap < 900*time.Millisecond {
		t.Fatalf("Retry-After ignored, retried after %v", gap)
	}
}

func TestRetryAfterAcceptsHTTPDate(t *testing.T) {
	limit := 30 * time.Second
	d, ok := parseRetryAfter(time.Now().Add(3*time.Second).UTC().Format(http.TimeFormat), limit)
	if !ok {
		t.Fatal("HTTP-date form not parsed")
	}
	if d < time.Second || d > 4*time.Second {
		t.Fatalf("unexpected delay %v", d)
	}
	// A date already in the past means retry now, not travel backwards.
	past, ok := parseRetryAfter(time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat), limit)
	if !ok || past != 0 {
		t.Fatalf("stale date gave %v (ok=%v)", past, ok)
	}
	if _, ok := parseRetryAfter("not-a-date", limit); ok {
		t.Fatal("garbage Retry-After should fall back to the computed backoff")
	}
}

func TestBackoffGrowsAndIsCapped(t *testing.T) {
	c := &Client{baseDelay: time.Second, maxDelay: 10 * time.Second}
	for attempt := 0; attempt < 8; attempt++ {
		d := c.backoff(attempt, "")
		if d <= 0 {
			t.Fatalf("attempt %d gave a non-positive delay %v", attempt, d)
		}
		// Jitter adds up to half the delay on top of the cap.
		if d > 15*time.Second {
			t.Fatalf("attempt %d blew past the cap: %v", attempt, d)
		}
	}
}

func TestContextCancellationStopsRetrying(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := testClient(srv)
	c.baseDelay = time.Hour
	c.maxDelay = time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := c.do(ctx, http.MethodGet, "devices", nil, nil, nil)
	if err == nil {
		t.Fatal("expected a context error")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("backoff did not observe context cancellation")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected to stop after 1 attempt, got %d", got)
	}
}

func TestQueryParametersSurviveRetries(t *testing.T) {
	var mu sync.Mutex
	var seen []string
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.URL.Query().Get("deviceId"))
		mu.Unlock()
		if atomic.AddInt32(&calls, 1) < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	params := url.Values{}
	params.Set("deviceId", "d1")
	if err := testClient(srv).do(context.Background(), http.MethodGet, "deviceConfigurations", params, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	for i, v := range seen {
		if v != "d1" {
			t.Fatalf("attempt %d lost the query string: %q", i+1, v)
		}
	}
}

func TestNewAppliesRetryDefaults(t *testing.T) {
	c := New("t", "")
	if c.retries() != DefaultMaxRetries {
		t.Fatalf("retries = %d", c.retries())
	}
	if c.baseDelay != DefaultBaseDelay || c.maxDelay != DefaultMaxDelay {
		t.Fatalf("backoff defaults not set: %v / %v", c.baseDelay, c.maxDelay)
	}
	if c.HTTPClient.Timeout != defaultTimeout {
		t.Fatalf("timeout = %v", c.HTTPClient.Timeout)
	}
	// A Client built directly opts in to retrying rather than acquiring it by
	// surprise, but its backoff must still be usable if it does.
	zero := &Client{}
	if zero.retries() != 0 {
		t.Fatalf("zero-value retries = %d, want 0", zero.retries())
	}
	if d := zero.backoff(0, ""); d <= 0 {
		t.Fatalf("zero-value backoff = %v", d)
	}
}

// withTokenEndpoint points the OAuth exchange at srv for the duration of a test.
func withTokenEndpoint(t *testing.T, srv *httptest.Server) {
	t.Helper()
	prevEndpoint, prevClient := tokenEndpoint, tokenHTTPClient
	tokenEndpoint, tokenHTTPClient = srv.URL, srv.Client()
	t.Cleanup(func() {
		tokenEndpoint, tokenHTTPClient = prevEndpoint, prevClient
	})
}

// The token exchange is the first request of every run. It used to be a bare
// http.PostForm - no context, no timeout, no retry - so a single rate-limited
// exchange failed the whole apply before a resource had been read.
func TestFetchAccessTokenRetriesRateLimit(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("form not parseable: %v", err)
		}
		if got := r.PostForm.Get("refresh_token"); got != "rt" {
			t.Errorf("refresh token not replayed: %q", got)
		}
		if atomic.AddInt32(&calls, 1) < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, `{"access_token":"at","expires_in":1209600}`)
	}))
	defer srv.Close()
	withTokenEndpoint(t, srv)

	tok, err := FetchAccessTokenContext(context.Background(), "id", "secret", "rt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "at" {
		t.Fatalf("token = %q", tok)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestFetchAccessTokenRespectsContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	withTokenEndpoint(t, srv)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	if _, err := FetchAccessTokenContext(ctx, "id", "secret", "rt"); err == nil {
		t.Fatal("expected a context error")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("token backoff ignored context cancellation")
	}
}

// An expired refresh token is a 400, not something to retry, and the guidance
// on regenerating it has to survive.
func TestFetchAccessTokenDoesNotRetryBadRequest(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
	}))
	defer srv.Close()
	withTokenEndpoint(t, srv)

	_, err := FetchAccessTokenContext(context.Background(), "id", "secret", "rt")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("invalid_grant retried %d times", got-1)
	}
	if !strings.Contains(err.Error(), "90 days") {
		t.Fatalf("refresh-token guidance lost: %v", err)
	}
}

// The deprecated wrapper must keep working for anything still calling it.
func TestFetchAccessTokenWrapper(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"access_token":"at"}`)
	}))
	defer srv.Close()
	withTokenEndpoint(t, srv)

	tok, err := FetchAccessToken("id", "secret", "rt")
	if err != nil || tok != "at" {
		t.Fatalf("token = %q, err = %v", tok, err)
	}
}
