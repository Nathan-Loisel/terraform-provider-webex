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
	// A zero-value Client built directly must still behave.
	zero := &Client{}
	if zero.retries() != DefaultMaxRetries {
		t.Fatalf("zero-value retries = %d", zero.retries())
	}
	if d := zero.backoff(0, ""); d <= 0 {
		t.Fatalf("zero-value backoff = %v", d)
	}
}
