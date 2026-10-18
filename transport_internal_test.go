package voiceml

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

const (
	tAccountSid = "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	tAPIKey     = "transport-test-key"
)

// countingServer answers every request with handler and counts hits.
func countingServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

// sleepRecorder replaces the transport's sleep so tests never wait.
type sleepRecorder struct {
	mu     sync.Mutex
	delays []time.Duration
}

func (s *sleepRecorder) sleep(ctx context.Context, d time.Duration) error {
	s.mu.Lock()
	s.delays = append(s.delays, d)
	s.mu.Unlock()
	return ctx.Err()
}

func testClient(t *testing.T, baseURL string, mutate func(*ClientOptions)) (*Client, *sleepRecorder) {
	t.Helper()
	opts := ClientOptions{AccountSid: tAccountSid, APIKey: tAPIKey, BaseURL: baseURL}
	if mutate != nil {
		mutate(&opts)
	}
	c, err := NewClient(opts)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	rec := &sleepRecorder{}
	c.t.sleep = rec.sleep
	return c, rec
}

func createCall(c *Client) error {
	_, err := c.Calls.Create(context.Background(), CreateCallParams{To: "+14155550100", From: "+14155550101", URL: String("https://example.com/twiml")})
	return err
}

func TestPOSTNotRetriedOn5xx(t *testing.T) {
	for _, status := range []int{500, 502, 503, 504} {
		srv, hits := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		})
		c, sleeps := testClient(t, srv.URL, nil)
		err := createCall(c)
		if !IsServer(err) {
			t.Fatalf("%d: want server error, got %v", status, err)
		}
		if got := atomic.LoadInt32(hits); got != 1 {
			t.Fatalf("%d: POST sent %d times, want exactly 1", status, got)
		}
		if len(sleeps.delays) != 0 {
			t.Fatalf("%d: slept before a retry that must not happen", status)
		}
	}
}

func TestPOSTRetriedOn429(t *testing.T) {
	var n int32
	srv2, hits2 := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&n, 1) == 1 {
			w.WriteHeader(429)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"sid":"CA1"}`))
	})
	c, sleeps := testClient(t, srv2.URL, nil)
	if err := createCall(c); err != nil {
		t.Fatalf("POST after 429: %v", err)
	}
	if got := atomic.LoadInt32(hits2); got != 2 {
		t.Fatalf("POST sent %d times, want 2 (one retry after 429)", got)
	}
	if len(sleeps.delays) != 1 || sleeps.delays[0] > backoffBase {
		t.Fatalf("backoff delays %v, want one delay in [0, %s]", sleeps.delays, backoffBase)
	}
}

func TestPOSTNotRetriedOnReadTimeout(t *testing.T) {
	release := make(chan struct{})
	srv, hits := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		<-release
		w.WriteHeader(201)
	})
	defer close(release)
	c, _ := testClient(t, srv.URL, func(o *ClientOptions) { o.Timeout = 100 * time.Millisecond })
	err := createCall(c)
	if !IsTimeout(err) || !IsTransport(err) {
		t.Fatalf("want transport timeout, got %v", err)
	}
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("cause not preserved: %v", err)
	}
	if got := atomic.LoadInt32(hits); got != 1 {
		t.Fatalf("POST sent %d times after a read timeout, want exactly 1", got)
	}
}

func TestPOSTNotRetriedOnConnectionReset(t *testing.T) {
	srv, hits := countingServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		conn, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			_ = conn.Close()
		}
	})
	c, _ := testClient(t, srv.URL, nil)
	if err := createCall(c); !IsTransport(err) {
		t.Fatalf("want transport error, got %v", err)
	}
	if got := atomic.LoadInt32(hits); got != 1 {
		t.Fatalf("POST sent %d times after a reset, want exactly 1", got)
	}
}

func TestGETRetriedOnConnectionReset(t *testing.T) {
	var n int32
	srv, hits := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&n, 1) == 1 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				_ = conn.Close()
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sid":"CA1"}`))
	})
	c, _ := testClient(t, srv.URL, nil)
	if _, err := c.Calls.Get(context.Background(), "CA1"); err != nil {
		t.Fatalf("GET after reset: %v", err)
	}
	if got := atomic.LoadInt32(hits); got != 2 {
		t.Fatalf("GET sent %d times, want 2", got)
	}
}

func TestGETRetriedOn5xx(t *testing.T) {
	for _, status := range []int{500, 502, 503, 504} {
		var n int32
		srv, hits := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
			if atomic.AddInt32(&n, 1) <= 2 {
				w.WriteHeader(status)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"sid":"CA1"}`))
		})
		c, sleeps := testClient(t, srv.URL, nil)
		if _, err := c.Calls.Get(context.Background(), "CA1"); err != nil {
			t.Fatalf("%d: GET: %v", status, err)
		}
		if got := atomic.LoadInt32(hits); got != 3 {
			t.Fatalf("%d: GET sent %d times, want 3 (MaxRetries=2)", status, got)
		}
		if len(sleeps.delays) != 2 || sleeps.delays[0] > backoffBase || sleeps.delays[1] > 2*backoffBase {
			t.Fatalf("%d: delays %v exceed the full-jitter ceilings", status, sleeps.delays)
		}
	}
}

func TestDELETERetriedOn503(t *testing.T) {
	var n int32
	srv, hits := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&n, 1) == 1 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(204)
	})
	c, _ := testClient(t, srv.URL, nil)
	if err := c.Calls.Delete(context.Background(), "CA1"); err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	if got := atomic.LoadInt32(hits); got != 2 {
		t.Fatalf("DELETE sent %d times, want 2", got)
	}
}

// dialFailOnce fails the first dial as connection-refused, so the request
// provably never reached the server; later dials succeed.
func dialFailOnce(dials *int32) func(ctx context.Context, network, addr string) (net.Conn, error) {
	var d net.Dialer
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if atomic.AddInt32(dials, 1) == 1 {
			return nil, &net.OpError{Op: "dial", Net: network, Err: syscall.ECONNREFUSED}
		}
		return d.DialContext(ctx, network, addr)
	}
}

func TestPOSTRetriedWhenConnectFails(t *testing.T) {
	srv, hits := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"sid":"CA1"}`))
	})
	var dials int32
	tr := &http.Transport{DialContext: dialFailOnce(&dials)}
	c, _ := testClient(t, srv.URL, func(o *ClientOptions) { o.HTTPClient = &http.Client{Transport: tr} })
	if err := createCall(c); err != nil {
		t.Fatalf("POST after refused connect: %v", err)
	}
	if atomic.LoadInt32(&dials) != 2 || atomic.LoadInt32(hits) != 1 {
		t.Fatalf("dials=%d hits=%d, want 2 dials and 1 delivered request", dials, *hits)
	}
}

func TestRetryAfterSeconds(t *testing.T) {
	var n int32
	srv, _ := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&n, 1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(429)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	c, sleeps := testClient(t, srv.URL, nil)
	if _, err := c.Calls.Get(context.Background(), "CA1"); err != nil {
		t.Fatal(err)
	}
	if len(sleeps.delays) != 1 || sleeps.delays[0] != 2*time.Second {
		t.Fatalf("delays %v, want [2s]", sleeps.delays)
	}
}

func TestRetryAfterHTTPDate(t *testing.T) {
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	var n int32
	srv, _ := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&n, 1) == 1 {
			w.Header().Set("Retry-After", now.Add(3*time.Second).Format(http.TimeFormat))
			w.WriteHeader(503)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	c, sleeps := testClient(t, srv.URL, nil)
	c.t.now = func() time.Time { return now }
	if _, err := c.Calls.Get(context.Background(), "CA1"); err != nil {
		t.Fatal(err)
	}
	if len(sleeps.delays) != 1 || sleeps.delays[0] != 3*time.Second {
		t.Fatalf("delays %v, want [3s]", sleeps.delays)
	}
}

func TestRetryAfterAboveCapReturnsImmediately(t *testing.T) {
	srv, hits := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"code":20429,"message":"Too Many Requests","status":429}`))
	})
	c, sleeps := testClient(t, srv.URL, nil)
	_, err := c.Calls.Get(context.Background(), "CA1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !IsRateLimit(err) {
		t.Fatalf("want rate-limit APIError, got %v", err)
	}
	if apiErr.RetryAfter != time.Hour {
		t.Fatalf("RetryAfter = %s, want 1h", apiErr.RetryAfter)
	}
	if atomic.LoadInt32(hits) != 1 || len(sleeps.delays) != 0 {
		t.Fatalf("hits=%d sleeps=%v; a Retry-After above %s must not be waited out", *hits, sleeps.delays, MaxRetryAfter)
	}
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"0", 0, true},
		{"120", 120 * time.Second, true},
		{"1.5", 1500 * time.Millisecond, true},
		{now.Add(10 * time.Second).Format(http.TimeFormat), 10 * time.Second, true},
		{now.Add(-10 * time.Second).Format(http.TimeFormat), 0, true},
		{"", 0, false},
		{"-1", 0, false},
		{"Inf", 0, false},
		{"NaN", 0, false},
		{"1e3", 0, false},
		{"soon", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseRetryAfter(tc.in, now)
		if got != tc.want || ok != tc.ok {
			t.Errorf("parseRetryAfter(%q) = %s, %v; want %s, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestBackoffFullJitterBounds(t *testing.T) {
	for attempt := 0; attempt < 8; attempt++ {
		ceiling := backoffBase << uint(attempt)
		if ceiling > backoffMax || attempt >= 5 {
			ceiling = backoffMax
		}
		var maxSeen time.Duration
		for i := 0; i < 500; i++ {
			d := backoff(attempt)
			if d < 0 || d > ceiling {
				t.Fatalf("attempt %d: delay %s outside [0, %s]", attempt, d, ceiling)
			}
			if d > maxSeen {
				maxSeen = d
			}
		}
		if maxSeen < ceiling/4 {
			t.Fatalf("attempt %d: 500 samples never exceeded %s; jitter looks degenerate", attempt, ceiling/4)
		}
	}
}

func TestCancellationDuringBackoffPropagates(t *testing.T) {
	srv, hits := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(503)
	})
	c, _ := testClient(t, srv.URL, nil)
	ctx, cancel := context.WithCancel(context.Background())
	c.t.sleep = func(ctx context.Context, _ time.Duration) error {
		cancel()
		<-ctx.Done()
		return ctx.Err()
	}
	_, err := c.Calls.Get(ctx, "CA1")
	if !errors.Is(err, context.Canceled) || err != context.Canceled {
		t.Fatalf("want context.Canceled unchanged, got %v", err)
	}
	if atomic.LoadInt32(hits) != 1 {
		t.Fatalf("hits = %d, want 1 (no retry after cancel)", *hits)
	}
}

func TestDecodeErrorKeepsCause(t *testing.T) {
	srv, _ := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sid": 42}`))
	})
	c, _ := testClient(t, srv.URL, nil)
	_, err := c.Calls.Get(context.Background(), "CA1")
	var decErr *DecodeError
	if !errors.As(err, &decErr) || decErr.StatusCode != 200 {
		t.Fatalf("want *DecodeError for 200, got %T %v", err, err)
	}
	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("decode cause not wrapped: %v", err)
	}
	if !strings.Contains(err.Error(), "could not decode") || strings.Contains(err.Error(), "non-JSON") {
		t.Fatalf("message %q should say 'could not decode'", err.Error())
	}

	srv2, _ := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html>ok</html>`))
	})
	c2, _ := testClient(t, srv2.URL, nil)
	_, err = c2.Calls.Get(context.Background(), "CA1")
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &decErr) || !errors.As(err, &syntaxErr) {
		t.Fatalf("want DecodeError wrapping *json.SyntaxError, got %v", err)
	}
	if string(decErr.Body) != "<html>ok</html>" {
		t.Fatalf("DecodeError.Body = %q", decErr.Body)
	}
}

func TestNonJSONErrorBody(t *testing.T) {
	srv, _ := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(502)
		_, _ = w.Write([]byte("<h1>Bad Gateway</h1>"))
	})
	c, _ := testClient(t, srv.URL, func(o *ClientOptions) { o.MaxRetries = Int(0) })
	_, err := c.Calls.Get(context.Background(), "CA1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !IsServer(err) {
		t.Fatalf("want server APIError, got %v", err)
	}
	if apiErr.StatusCode != 502 || string(apiErr.Body) != "<h1>Bad Gateway</h1>" || apiErr.Message != "HTTP 502: <h1>Bad Gateway</h1>" {
		t.Fatalf("got %+v", apiErr)
	}
}

func TestJSONBodyLimit(t *testing.T) {
	srv, _ := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sid":"`))
		chunk := []byte(strings.Repeat("a", 1<<20))
		for i := 0; i < (maxJSONBody>>20)+1; i++ {
			if _, err := w.Write(chunk); err != nil {
				return
			}
		}
		_, _ = w.Write([]byte(`"}`))
	})
	c, _ := testClient(t, srv.URL, nil)
	_, err := c.Calls.Get(context.Background(), "CA1")
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("want body-size error, got %v", err)
	}
}

func TestCredentialLessTransportRefusesAuthenticatedCalls(t *testing.T) {
	srv, hits := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	tr, urls, err := buildTransport(ClientOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	c := newClient(tr, urls)
	_, err = c.Calls.List(context.Background(), ListCallsParams{})
	var cfg *ConfigurationError
	if !errors.As(err, &cfg) {
		t.Fatalf("want *ConfigurationError, got %v", err)
	}
	if atomic.LoadInt32(hits) != 0 {
		t.Fatal("a credential-less client sent an authenticated request")
	}
}
