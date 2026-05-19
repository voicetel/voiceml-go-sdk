// Package voiceml smoke tests. Every test stands up an httptest server,
// inspects the request the SDK sends, and feeds back a canned response.
// No external dependencies; pure stdlib.

package voiceml_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

const (
	testAccountSid = "AC" + "ffffffffffffffffffffffffffffffff"
	testAPIKey     = "secret-key-1234"
)

// recorder captures the requests received by an httptest server so a test
// can assert on path / method / headers / body after the fact.
type recorder struct {
	requests []capturedRequest
}

type capturedRequest struct {
	Method string
	Path   string
	Query  string
	Header http.Header
	Body   []byte
}

// newRecorder returns a *recorder and an http.Handler. The handler walks
// `responses` in order — each is either a status+body or a function that
// builds a response from the captured request.
func newRecorder(t *testing.T, responses []handlerStep) (*recorder, http.Handler) {
	t.Helper()
	rec := &recorder{}
	i := 0
	return rec, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		rec.requests = append(rec.requests, capturedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.RawQuery,
			Header: r.Header.Clone(),
			Body:   body,
		})
		if i >= len(responses) {
			t.Fatalf("recorder: out of responses after %d requests", i)
		}
		step := responses[i]
		i++
		step(w, r)
	})
}

type handlerStep func(w http.ResponseWriter, r *http.Request)

func jsonStep(status int, body any) handlerStep {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body == nil {
			return
		}
		b, _ := json.Marshal(body)
		_, _ = w.Write(b)
	}
}

func plainStep(status int, body string, headers map[string]string) handlerStep {
	return func(w http.ResponseWriter, _ *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// newClient builds a Client pointed at a recorded test server. Returns the
// client + recorder + cleanup func.
func newClient(t *testing.T, responses []handlerStep, mutate func(*voiceml.ClientOptions)) (*voiceml.Client, *recorder, func()) {
	t.Helper()
	rec, handler := newRecorder(t, responses)
	srv := httptest.NewServer(handler)
	opts := voiceml.ClientOptions{
		AccountSid: testAccountSid,
		APIKey:     testAPIKey,
		BaseURL:    srv.URL,
	}
	if mutate != nil {
		mutate(&opts)
	}
	c, err := voiceml.NewClient(opts)
	if err != nil {
		srv.Close()
		t.Fatalf("NewClient: %v", err)
	}
	return c, rec, srv.Close
}

func basicAuthHeader() string {
	creds := testAccountSid + ":" + testAPIKey
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))
}

func callPayload(sid string) map[string]any {
	if sid == "" {
		sid = "CA" + strings.Repeat("0", 32)
	}
	return map[string]any{
		"sid":          sid,
		"account_sid":  testAccountSid,
		"api_version":  "2010-04-01",
		"status":       "queued",
		"direction":    "outbound-api",
		"date_created": "Mon, 19 May 2026 12:00:00 +0000",
		"date_updated": "Mon, 19 May 2026 12:00:00 +0000",
		"uri":          fmt.Sprintf("/2010-04-01/Accounts/%s/Calls/%s.json", testAccountSid, sid),
	}
}

// 1. Module surface — version + required options.
func TestModuleSurface(t *testing.T) {
	if voiceml.Version != "0.4.0" {
		t.Fatalf("Version: want 0.4.0, got %q", voiceml.Version)
	}

	cases := []struct {
		name string
		opts voiceml.ClientOptions
		want string
	}{
		{"missing AccountSid", voiceml.ClientOptions{APIKey: testAPIKey}, "AccountSid is required"},
		{"missing APIKey", voiceml.ClientOptions{AccountSid: testAccountSid}, "APIKey is required"},
		{"negative MaxRetries", voiceml.ClientOptions{
			AccountSid: testAccountSid, APIKey: testAPIKey, MaxRetries: voiceml.Int(-1),
		}, "MaxRetries must be >= 0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := voiceml.NewClient(tc.opts)
			if err == nil {
				t.Fatal("expected error")
			}
			var cfg *voiceml.ConfigurationError
			if !errors.As(err, &cfg) {
				t.Fatalf("want *ConfigurationError, got %T", err)
			}
			if cfg.Message != tc.want {
				t.Fatalf("want %q, got %q", tc.want, cfg.Message)
			}
		})
	}
}

// 2. Default base URL — when BaseURL is unset, the client uses production.
func TestDefaultBaseURL(t *testing.T) {
	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: testAccountSid,
		APIKey:     testAPIKey,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.BaseURL != "https://voiceml.voicetel.com" {
		t.Fatalf("BaseURL: want production, got %q", c.BaseURL)
	}
	if c.AccountSid != testAccountSid {
		t.Fatalf("AccountSid: want %q, got %q", testAccountSid, c.AccountSid)
	}
	// All resource services wired up.
	if c.Calls == nil || c.Conferences == nil || c.Queues == nil ||
		c.Applications == nil || c.Recordings == nil || c.Diagnostics == nil {
		t.Fatal("not all services wired up on Client")
	}
}

// 3. Calls.Create — form body fields, Basic auth, URL path.
func TestCallsCreate(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, callPayload("")),
	}, nil)
	defer cleanup()

	call, err := c.Calls.Create(context.Background(), voiceml.CreateCallParams{
		To:               "+18005551234",
		From:             "+18005550000",
		URL:              voiceml.String("https://example.com/twiml"),
		MachineDetection: voiceml.String("DetectMessageEnd"),
	})
	if err != nil {
		t.Fatalf("Calls.Create: %v", err)
	}
	if !strings.HasPrefix(call.Sid, "CA") {
		t.Fatalf("expected CA-prefixed sid, got %q", call.Sid)
	}

	if len(rec.requests) != 1 {
		t.Fatalf("want 1 request, got %d", len(rec.requests))
	}
	req := rec.requests[0]
	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/Calls", testAccountSid)
	if req.Path != wantPath {
		t.Fatalf("path: want %q, got %q", wantPath, req.Path)
	}
	if req.Method != "POST" {
		t.Fatalf("method: want POST, got %s", req.Method)
	}
	if got := req.Header.Get("Authorization"); got != basicAuthHeader() {
		t.Fatalf("Authorization: want %q, got %q", basicAuthHeader(), got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type: want form-urlencoded, got %q", got)
	}
	body := string(req.Body)
	for _, want := range []string{
		"To=%2B18005551234",
		"From=%2B18005550000",
		"Url=https%3A%2F%2Fexample.com%2Ftwiml",
		"MachineDetection=DetectMessageEnd",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q; got %q", want, body)
		}
	}
}

// 4. Calls.List — Twilio-shape query params including StartTime>= / <=.
func TestCallsListQueryParams(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"calls":     []any{callPayload("")},
			"page":      0,
			"page_size": 50,
			"total":     1,
			"uri":       "/Calls",
		}),
	}, nil)
	defer cleanup()

	page := 0
	pageSize := 10
	_, err := c.Calls.List(context.Background(), voiceml.ListCallsParams{
		Status:       "completed",
		StartTimeGte: "2026-01-01",
		StartTimeLte: "2026-12-31",
		Page:         &page,
		PageSize:     &pageSize,
	})
	if err != nil {
		t.Fatalf("Calls.List: %v", err)
	}

	q := rec.requests[0].Query
	// url.Values.Encode percent-encodes ">" / "<" / "=" in keys:
	// ">" → %3E, "<" → %3C, "=" → %3D.
	wants := []string{
		"Status=completed",
		"StartTime%3E%3D=2026-01-01",
		"StartTime%3C%3D=2026-12-31",
		"Page=0",
		"PageSize=10",
	}
	for _, w := range wants {
		if !strings.Contains(q, w) {
			t.Errorf("query missing %q; got %q", w, q)
		}
	}
}

// 5. Boolean encoding — Muted=true, Hold=false.
func TestParticipantBooleanEncoding(t *testing.T) {
	cfSid := "CF" + strings.Repeat("5", 32)
	callSid := "CA" + strings.Repeat("4", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"call_sid":                  callSid,
			"conference_sid":            cfSid,
			"account_sid":               testAccountSid,
			"muted":                     true,
			"hold":                      false,
			"start_conference_on_enter": true,
			"end_conference_on_exit":    false,
			"status":                    "connected",
			"api_version":               "2010-04-01",
			"uri":                       "/x",
		}),
	}, nil)
	defer cleanup()

	_, err := c.Conferences.UpdateParticipant(context.Background(), cfSid, callSid, voiceml.UpdateParticipantParams{
		Muted: voiceml.Bool(true),
		Hold:  voiceml.Bool(false),
	})
	if err != nil {
		t.Fatalf("UpdateParticipant: %v", err)
	}

	body := string(rec.requests[0].Body)
	if !strings.Contains(body, "Muted=true") {
		t.Errorf("body missing Muted=true; got %q", body)
	}
	if !strings.Contains(body, "Hold=false") {
		t.Errorf("body missing Hold=false; got %q", body)
	}
}

// 6. Streams.Start — Url + Track + Name.
func TestStreamsStart(t *testing.T) {
	callSid := "CA" + strings.Repeat("6", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, map[string]any{
			"sid":         "MZ" + strings.Repeat("7", 32),
			"account_sid": testAccountSid,
			"call_sid":    callSid,
			"status":      "in-progress",
			"api_version": "2010-04-01",
			"uri":         "/x",
		}),
	}, nil)
	defer cleanup()

	track := string(voiceml.TrackBoth)
	_, err := c.Calls.StartStream(context.Background(), callSid, voiceml.StartStreamParams{
		URL:   "wss://example.com/ws",
		Track: &track,
		Name:  voiceml.String("ws-1"),
	})
	if err != nil {
		t.Fatalf("StartStream: %v", err)
	}

	body := string(rec.requests[0].Body)
	for _, want := range []string{
		"Url=wss%3A%2F%2Fexample.com%2Fws",
		"Track=both_tracks",
		"Name=ws-1",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q; got %q", want, body)
		}
	}
}

// 7. Error mapping: 401 → ErrAuthentication.
func TestErrorMapping401(t *testing.T) {
	sid := "CA" + strings.Repeat("8", 32)
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(401, map[string]any{"code": 20003, "message": "Authentication Error", "status": 401}),
	}, nil)
	defer cleanup()

	_, err := c.Calls.Get(context.Background(), sid)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, voiceml.ErrAuthentication) {
		t.Fatalf("want ErrAuthentication, got %v", err)
	}
	if !voiceml.IsAuthentication(err) {
		t.Fatal("IsAuthentication(err) = false")
	}
	var apiErr *voiceml.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Fatalf("StatusCode: want 401, got %d", apiErr.StatusCode)
	}
}

// 8. Error mapping: 404 → ErrNotFound.
func TestErrorMapping404(t *testing.T) {
	sid := "CA" + strings.Repeat("9", 32)
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(404, map[string]any{"code": 20404, "message": "Not Found", "status": 404}),
	}, nil)
	defer cleanup()

	_, err := c.Calls.Get(context.Background(), sid)
	if !errors.Is(err, voiceml.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if !voiceml.IsNotFound(err) {
		t.Fatal("IsNotFound(err) = false")
	}
}

// 9. Error mapping: 429 → ErrRateLimit. No retry when MaxRetries=0.
func TestErrorMapping429NoRetry(t *testing.T) {
	sid := "CA" + strings.Repeat("a", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(429, map[string]any{"code": 20429, "message": "Too Many", "status": 429}),
	}, func(opts *voiceml.ClientOptions) {
		opts.MaxRetries = voiceml.Int(0)
	})
	defer cleanup()

	_, err := c.Calls.Get(context.Background(), sid)
	if !errors.Is(err, voiceml.ErrRateLimit) {
		t.Fatalf("want ErrRateLimit, got %v", err)
	}
	if !voiceml.IsRateLimit(err) {
		t.Fatal("IsRateLimit(err) = false")
	}
	if len(rec.requests) != 1 {
		t.Fatalf("expected exactly 1 attempt (no retry), got %d", len(rec.requests))
	}
}

// 10. Error mapping: 501 → ErrNotImplemented (UserDefinedMessages).
func TestErrorMapping501(t *testing.T) {
	sid := "CA" + strings.Repeat("b", 32)
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(501, map[string]any{"code": 20501, "message": "Not Implemented", "status": 501}),
	}, nil)
	defer cleanup()

	err := c.Calls.SendUserDefinedMessage(context.Background(), sid, map[string]any{"hello": "world"})
	if !errors.Is(err, voiceml.ErrNotImplemented) {
		t.Fatalf("want ErrNotImplemented, got %v", err)
	}
	if !voiceml.IsNotImplemented(err) {
		t.Fatal("IsNotImplemented(err) = false")
	}
}

// 11. Error mapping: 409 → catch-all APIError with Code 20409.
func TestErrorMapping409(t *testing.T) {
	sid := "QU" + strings.Repeat("c", 32)
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(409, map[string]any{"code": 20409, "message": "Queue still has waiting members", "status": 409}),
	}, nil)
	defer cleanup()

	err := c.Queues.Delete(context.Background(), sid)
	var apiErr *voiceml.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError, got %T", err)
	}
	if apiErr.StatusCode != 409 {
		t.Fatalf("StatusCode: want 409, got %d", apiErr.StatusCode)
	}
	codeInt, ok := apiErr.Code.(int64)
	if !ok {
		t.Fatalf("Code: want int64, got %T (%v)", apiErr.Code, apiErr.Code)
	}
	if codeInt != 20409 {
		t.Fatalf("Code: want 20409, got %d", codeInt)
	}
	if !errors.Is(err, voiceml.ErrConflict) {
		t.Fatal("err should wrap ErrConflict")
	}
}

// 12. Retry policy: 503 then 200 succeeds when MaxRetries=1.
func TestRetry503Then200(t *testing.T) {
	sid := "CA" + strings.Repeat("d", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		plainStep(503, "upstream busy", nil),
		jsonStep(200, callPayload(sid)),
	}, func(opts *voiceml.ClientOptions) {
		opts.MaxRetries = voiceml.Int(1)
	})
	defer cleanup()

	call, err := c.Calls.Get(context.Background(), sid)
	if err != nil {
		t.Fatalf("Calls.Get: %v", err)
	}
	if call.Sid != sid {
		t.Fatalf("sid: want %q, got %q", sid, call.Sid)
	}
	if len(rec.requests) != 2 {
		t.Fatalf("expected 2 attempts (1 retry), got %d", len(rec.requests))
	}
}

// 13. Recordings.GetAudio — returns bytes + content-type from a 200.
func TestRecordingsGetAudio(t *testing.T) {
	reSid := "RE" + strings.Repeat("e", 32)
	wavBytes := []byte{0x52, 0x49, 0x46, 0x46} // "RIFF"
	c, _, cleanup := newClient(t, []handlerStep{
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "audio/wav")
			w.WriteHeader(200)
			_, _ = w.Write(wavBytes)
		},
	}, nil)
	defer cleanup()

	body, ct, err := c.Recordings.GetAudio(context.Background(), reSid)
	if err != nil {
		t.Fatalf("GetAudio: %v", err)
	}
	if ct != "audio/wav" {
		t.Fatalf("content-type: want audio/wav, got %q", ct)
	}
	if string(body) != string(wavBytes) {
		t.Fatalf("body: want %x, got %x", wavBytes, body)
	}
}

// 14. Queues.Create — FriendlyName + MaxSize fields.
func TestQueuesCreate(t *testing.T) {
	quSid := "QU" + strings.Repeat("3", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, map[string]any{
			"sid":               quSid,
			"account_sid":       testAccountSid,
			"friendly_name":     "support",
			"current_size":      0,
			"max_size":          200,
			"average_wait_time": 0,
			"date_created":      "x",
			"date_updated":      "x",
			"uri":               "/x",
		}),
	}, nil)
	defer cleanup()

	maxSize := 200
	q, err := c.Queues.Create(context.Background(), voiceml.CreateQueueParams{
		FriendlyName: "support",
		MaxSize:      &maxSize,
	})
	if err != nil {
		t.Fatalf("Queues.Create: %v", err)
	}
	if q.Sid != quSid {
		t.Fatalf("sid: want %q, got %q", quSid, q.Sid)
	}
	body := string(rec.requests[0].Body)
	for _, want := range []string{"FriendlyName=support", "MaxSize=200"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q; got %q", want, body)
		}
	}
}

// 15. Conferences.End — defaults to Status=completed when nil is passed.
func TestConferencesEndDefault(t *testing.T) {
	cfSid := "CF" + strings.Repeat("2", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"sid":           cfSid,
			"account_sid":   testAccountSid,
			"friendly_name": "x",
			"status":        "completed",
			"api_version":   "2010-04-01",
			"uri":           "/x",
		}),
	}, nil)
	defer cleanup()

	conf, err := c.Conferences.End(context.Background(), cfSid, nil)
	if err != nil {
		t.Fatalf("Conferences.End: %v", err)
	}
	if conf.Status != "completed" {
		t.Fatalf("status: want completed, got %q", conf.Status)
	}
	body := string(rec.requests[0].Body)
	if !strings.Contains(body, "Status=completed") {
		t.Errorf("body missing Status=completed; got %q", body)
	}
}

// 16. Diagnostics.Health — unauthenticated GET /health.
func TestDiagnosticsHealth(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{"ok": true, "warnings": []any{}, "failures": []any{}}),
	}, nil)
	defer cleanup()

	h, err := c.Diagnostics.Health(context.Background())
	if err != nil {
		t.Fatalf("Diagnostics.Health: %v", err)
	}
	if !h.OK {
		t.Fatal("ok: want true")
	}
	req := rec.requests[0]
	if req.Path != "/health" {
		t.Fatalf("path: want /health, got %q", req.Path)
	}
	if got := req.Header.Get("Authorization"); got != "" {
		t.Fatalf("Authorization: want empty (unauth), got %q", got)
	}
}

// 17. Recording.GetAudio: 410 Gone → ErrGone.
func TestRecordingsGetAudio410(t *testing.T) {
	reSid := "RE" + strings.Repeat("f", 32)
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(410, map[string]any{"code": 20410, "message": "Audio is gone", "status": 410}),
	}, nil)
	defer cleanup()

	_, _, err := c.Recordings.GetAudio(context.Background(), reSid)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, voiceml.ErrGone) {
		t.Fatalf("want ErrGone, got %v", err)
	}
}
