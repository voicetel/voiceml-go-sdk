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
	"time"

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
	Method  string
	Path    string
	RawPath string
	Query   string
	Header  http.Header
	Body    []byte
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
			Method:  r.Method,
			Path:    r.URL.Path,
			RawPath: r.URL.EscapedPath(),
			Query:   r.URL.RawQuery,
			Header:  r.Header.Clone(),
			Body:    body,
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

// assertWire fails unless got is exactly want. Form bodies and query
// strings are compared as raw bytes, so "+" vs "%2B", key spelling and
// duplicate keys are all visible.
func assertWire(t *testing.T, what, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\n got  %q\n want %q", what, got, want)
	}
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
	if voiceml.Version != "0.9.5" {
		t.Fatalf("Version: want 0.9.5, got %q", voiceml.Version)
	}

	cases := []struct {
		name string
		opts voiceml.ClientOptions
		want string
	}{
		{"missing AccountSid", voiceml.ClientOptions{APIKey: testAPIKey}, "AccountSid is required"},
		{"missing APIKey", voiceml.ClientOptions{AccountSid: testAccountSid}, "APIKey is required"},
		{"both APIKey and AuthToken", voiceml.ClientOptions{
			AccountSid: testAccountSid, APIKey: testAPIKey, AuthToken: "alt",
		}, "set APIKey or AuthToken, not both"},
		{"negative MaxRetries", voiceml.ClientOptions{
			AccountSid: testAccountSid, APIKey: testAPIKey, MaxRetries: voiceml.Int(-1),
		}, "MaxRetries must be >= 0"},
		{"negative Timeout", voiceml.ClientOptions{
			AccountSid: testAccountSid, APIKey: testAPIKey, Timeout: -1,
		}, "Timeout must be >= 0"},
		{"relative BaseURL", voiceml.ClientOptions{
			AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: "voiceml.voicetel.com",
		}, `BaseURL must use http or https, got "voiceml.voicetel.com"`},
		{"BaseURL with credentials", voiceml.ClientOptions{
			AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: "https://u:p@example.com",
		}, "BaseURL must not embed credentials"},
		{"BaseURL with query", voiceml.ClientOptions{
			AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: "https://example.com/?x=1",
		}, `BaseURL must not have a query or fragment: "https://example.com/?x=1"`},
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
		c.Applications == nil || c.Recordings == nil || c.Diagnostics == nil ||
		c.IncomingPhoneNumbers == nil {
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
		To:               "+14155550100",
		From:             "+14155550101",
		URL:              voiceml.String("https://example.com/twiml"),
		MachineDetection: voiceml.String("DetectMessageEnd"),
		TimeLimit:        voiceml.Int(3600),
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
	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/Calls.json", testAccountSid)
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
	assertWire(t, "body", string(req.Body),
		"From=%2B14155550101&MachineDetection=DetectMessageEnd&TimeLimit=3600&To=%2B14155550100&Url=https%3A%2F%2Fexample.com%2Ftwiml")
}

// 4. Calls.List — only the spec's query keys: StartTime< / StartTime> (no
// StartTime>= / StartTime<=), Status, Page, PageSize.
func TestCallsListQueryParams(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"calls":     []any{callPayload("")},
			"page":      0,
			"page_size": 50,
			"uri":       "/Calls",
		}),
	}, nil)
	defer cleanup()

	page := 0
	pageSize := 10
	_, err := c.Calls.List(context.Background(), voiceml.ListCallsParams{
		Status:      "completed",
		StartTimeGt: "2026-01-01",
		StartTimeLt: "2026-12-31",
		Page:        &page,
		PageSize:    &pageSize,
	})
	if err != nil {
		t.Fatalf("Calls.List: %v", err)
	}
	// url.Values.Encode percent-encodes "<" / ">" in keys: %3C / %3E.
	assertWire(t, "query", rec.requests[0].Query,
		"Page=0&PageSize=10&StartTime%3C=2026-12-31&StartTime%3E=2026-01-01&Status=completed")
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

	assertWire(t, "body", string(rec.requests[0].Body), "Hold=false&Muted=true")
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

	assertWire(t, "body", string(rec.requests[0].Body), "Name=ws-1&Track=both_tracks&Url=wss%3A%2F%2Fexample.com%2Fws")
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

	err := c.Calls.SendUserDefinedMessage(context.Background(), sid)
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
	if apiErr.Code != 20409 {
		t.Fatalf("Code: want 20409, got %d", apiErr.Code)
	}
	if apiErr.Status != 409 {
		t.Fatalf("Status: want 409, got %d", apiErr.Status)
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

// 13. Recordings.GetAudio — returns bytes + content-type from a 200. The
// request path must end in ".wav" (not ".json" nor ".json.wav").
func TestRecordingsGetAudio(t *testing.T) {
	reSid := "RE" + strings.Repeat("e", 32)
	wavBytes := []byte{0x52, 0x49, 0x46, 0x46} // "RIFF"
	c, rec, cleanup := newClient(t, []handlerStep{
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
	gotPath := rec.requests[0].Path
	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/Recordings/%s.wav", testAccountSid, reSid)
	if gotPath != wantPath {
		t.Fatalf("path: want %q, got %q", wantPath, gotPath)
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
	assertWire(t, "body", string(rec.requests[0].Body), "FriendlyName=support&MaxSize=200")
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
	assertWire(t, "body", string(rec.requests[0].Body), "Status=completed")
}

// 16. Diagnostics.Health — unauthenticated GET /health.
func TestDiagnosticsHealth(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{"ok": true}),
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

// --- v0.5.0 additions ---

// 18. AuthToken alias: NewClient accepts AuthToken in place of APIKey and
// uses it as the HTTP Basic password on the wire.
func TestAuthTokenAlias(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, callPayload("")),
	}, func(opts *voiceml.ClientOptions) {
		opts.APIKey = ""
		opts.AuthToken = testAPIKey // reuse the constant — it's the wire password
	})
	defer cleanup()

	sid := "CA" + strings.Repeat("0", 32)
	if _, err := c.Calls.Get(context.Background(), sid); err != nil {
		t.Fatalf("Calls.Get: %v", err)
	}
	if got := rec.requests[0].Header.Get("Authorization"); got != basicAuthHeader() {
		t.Fatalf("Authorization: want %q, got %q", basicAuthHeader(), got)
	}
}

// 19. IncomingPhoneNumbers.List — query params encode, response decodes,
// request path ends in ".json", pagination envelope round-trips.
func TestIncomingPhoneNumbersList(t *testing.T) {
	pnSid := "PN" + strings.Repeat("a", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"incoming_phone_numbers": []any{
				map[string]any{
					"sid":          pnSid,
					"account_sid":  testAccountSid,
					"phone_number": "+14155550100",
					"api_version":  "2010-04-01",
					"uri": fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers/%s.json",
						testAccountSid, pnSid),
					"capabilities": map[string]any{
						"voice": true, "sms": false, "mms": false, "fax": false,
					},
				},
			},
			"page":           0,
			"page_size":      50,
			"first_page_uri": "/IncomingPhoneNumbers?Page=0&PageSize=50",
			"next_page_uri":  "",
			"uri": fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers.json",
				testAccountSid),
		}),
	}, nil)
	defer cleanup()

	page := 0
	pageSize := 50
	list, err := c.IncomingPhoneNumbers.List(context.Background(),
		&voiceml.ListIncomingPhoneNumbersParams{
			PhoneNumber: "+14155550100",
			Page:        &page,
			PageSize:    &pageSize,
		})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.IncomingPhoneNumbers) != 1 {
		t.Fatalf("len: want 1, got %d", len(list.IncomingPhoneNumbers))
	}
	pn := list.IncomingPhoneNumbers[0]
	if !strings.HasPrefix(pn.Sid, "PN") {
		t.Fatalf("expected PN-prefixed sid, got %q", pn.Sid)
	}
	if pn.PhoneNumber != "+14155550100" {
		t.Fatalf("phone_number: want +14155550100, got %q", pn.PhoneNumber)
	}
	if !pn.Capabilities.Voice {
		t.Fatal("capabilities.voice: want true")
	}
	if list.PageSize != 50 || list.Page.Page != 0 {
		t.Fatalf("pagination: want page=0 size=50, got page=%d size=%d", list.Page.Page, list.PageSize)
	}
	if list.FirstPageURI == "" {
		t.Fatal("first_page_uri missing from list envelope")
	}

	req := rec.requests[0]
	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers.json",
		testAccountSid)
	if req.Path != wantPath {
		t.Fatalf("path: want %q, got %q", wantPath, req.Path)
	}
	assertWire(t, "query", req.Query, "Page=0&PageSize=50&PhoneNumber=%2B14155550100")
}

// 20. IncomingPhoneNumbers.Create — required PhoneNumber on the wire,
// optional VoiceUrl, 201 → IncomingPhoneNumber, path ends in ".json".
func TestIncomingPhoneNumbersCreate(t *testing.T) {
	pnSid := "PN" + strings.Repeat("b", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, map[string]any{
			"sid":          pnSid,
			"account_sid":  testAccountSid,
			"phone_number": "+14155550101",
			"api_version":  "2010-04-01",
			"voice_url":    "https://example.com/twiml",
			"voice_method": "POST",
			"uri": fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers/%s.json",
				testAccountSid, pnSid),
			"capabilities": map[string]any{
				"voice": true, "sms": false, "mms": false, "fax": false,
			},
		}),
	}, nil)
	defer cleanup()

	pn, err := c.IncomingPhoneNumbers.Create(context.Background(),
		voiceml.CreateIncomingPhoneNumberParams{
			PhoneNumber:          "+14155550101",
			VoiceURL:             voiceml.String("https://example.com/twiml"),
			VoiceMethod:          voiceml.String("POST"),
			VoiceApplicationSid:  voiceml.String("AP" + strings.Repeat("a", 32)),
			StatusCallback:       voiceml.String("https://example.com/status"),
			StatusCallbackMethod: voiceml.String("GET"),
		})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if pn.Sid != pnSid {
		t.Fatalf("sid: want %q, got %q", pnSid, pn.Sid)
	}
	if pn.VoiceURL == nil || *pn.VoiceURL != "https://example.com/twiml" {
		t.Fatalf("voice_url didn't round-trip: %+v", pn.VoiceURL)
	}

	req := rec.requests[0]
	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers.json",
		testAccountSid)
	if req.Path != wantPath {
		t.Fatalf("path: want %q, got %q", wantPath, req.Path)
	}
	if req.Method != "POST" {
		t.Fatalf("method: want POST, got %s", req.Method)
	}
	assertWire(t, "body", string(req.Body), "PhoneNumber=%2B14155550101&StatusCallback=https%3A%2F%2Fexample.com%2Fstatus&StatusCallbackMethod=GET&VoiceApplicationSid=APaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&VoiceMethod=POST&VoiceUrl=https%3A%2F%2Fexample.com%2Ftwiml")
}

// 21. IncomingPhoneNumbers.Get — single-row fetch, request path includes
// {PN-sid}.json.
func TestIncomingPhoneNumbersGet(t *testing.T) {
	pnSid := "PN" + strings.Repeat("c", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"sid":          pnSid,
			"account_sid":  testAccountSid,
			"phone_number": "+14155550104",
			"api_version":  "2010-04-01",
			"uri": fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers/%s.json",
				testAccountSid, pnSid),
			"capabilities": map[string]any{
				"voice": true, "sms": false, "mms": false, "fax": false,
			},
		}),
	}, nil)
	defer cleanup()

	pn, err := c.IncomingPhoneNumbers.Get(context.Background(), pnSid)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if pn.Sid != pnSid {
		t.Fatalf("sid: want %q, got %q", pnSid, pn.Sid)
	}
	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers/%s.json",
		testAccountSid, pnSid)
	if rec.requests[0].Path != wantPath {
		t.Fatalf("path: want %q, got %q", wantPath, rec.requests[0].Path)
	}
}

// 22. IncomingPhoneNumbers.Update — only-set-fields-touched on the wire.
func TestIncomingPhoneNumbersUpdate(t *testing.T) {
	pnSid := "PN" + strings.Repeat("d", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"sid":          pnSid,
			"account_sid":  testAccountSid,
			"phone_number": "+14155550105",
			"api_version":  "2010-04-01",
			"voice_url":    "https://example.com/new",
			"uri": fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers/%s.json",
				testAccountSid, pnSid),
			"capabilities": map[string]any{
				"voice": true, "sms": false, "mms": false, "fax": false,
			},
		}),
	}, nil)
	defer cleanup()

	pn, err := c.IncomingPhoneNumbers.Update(context.Background(), pnSid,
		voiceml.UpdateIncomingPhoneNumberParams{
			VoiceURL:             voiceml.String("https://example.com/new"),
			VoiceApplicationSid:  voiceml.String("AP" + strings.Repeat("b", 32)),
			StatusCallback:       voiceml.String("https://example.com/lifecycle"),
			StatusCallbackMethod: voiceml.String("POST"),
		})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if pn.VoiceURL == nil || *pn.VoiceURL != "https://example.com/new" {
		t.Fatalf("voice_url didn't round-trip: %+v", pn.VoiceURL)
	}
	// Only the fields that were set reach the wire.
	assertWire(t, "body", string(rec.requests[0].Body), "StatusCallback=https%3A%2F%2Fexample.com%2Flifecycle&StatusCallbackMethod=POST&VoiceApplicationSid=APbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb&VoiceUrl=https%3A%2F%2Fexample.com%2Fnew")
}

// 23. IncomingPhoneNumbers.Delete — 204 No Content, returns nil.
func TestIncomingPhoneNumbersDelete(t *testing.T) {
	pnSid := "PN" + strings.Repeat("e", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(204)
		},
	}, nil)
	defer cleanup()

	if err := c.IncomingPhoneNumbers.Delete(context.Background(), pnSid); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if rec.requests[0].Method != "DELETE" {
		t.Fatalf("method: want DELETE, got %s", rec.requests[0].Method)
	}
	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers/%s.json",
		testAccountSid, pnSid)
	if rec.requests[0].Path != wantPath {
		t.Fatalf("path: want %q, got %q", wantPath, rec.requests[0].Path)
	}
}

// 24. APIError.MoreInfo is populated from the response body's `more_info`.
func TestAPIErrorMoreInfo(t *testing.T) {
	sid := "CA" + strings.Repeat("1", 32)
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(404, map[string]any{
			"code":      20404,
			"message":   "Not Found",
			"more_info": "https://voicetel.com/docs/errors/20404",
			"status":    404,
		}),
	}, nil)
	defer cleanup()

	_, err := c.Calls.Get(context.Background(), sid)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *voiceml.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError, got %T", err)
	}
	if apiErr.MoreInfo != "https://voicetel.com/docs/errors/20404" {
		t.Fatalf("MoreInfo: want docs URL, got %q", apiErr.MoreInfo)
	}
}

// 25. Recording.MediaURL round-trips JSON (spec v0.6.2 / D5).
func TestRecordingMediaURLRoundTrip(t *testing.T) {
	rSid := "RE" + strings.Repeat("a", 32)
	mediaURL := "https://media.example.com/recordings/" + rSid + ".wav"
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"sid":         rSid,
			"account_sid": testAccountSid,
			"call_sid":    "CA" + strings.Repeat("b", 32),
			"status":      "completed",
			"api_version": "2010-04-01",
			"uri":         "/x",
			"media_url":   mediaURL,
		}),
	}, nil)
	defer cleanup()

	rec, err := c.Recordings.Get(context.Background(), rSid, nil)
	if err != nil {
		t.Fatalf("Recordings.Get: %v", err)
	}
	if rec.MediaURL != mediaURL {
		t.Fatalf("MediaURL: want %q, got %q", mediaURL, rec.MediaURL)
	}
}

// 26. IncomingPhoneNumber.Type is decoded when present (spec v0.6.2 / D6).
// VoiceML emits empty string for Type but the Twilio-compat field still needs
// to round-trip when non-empty (e.g. when proxying upstream Twilio numbers).
func TestIncomingPhoneNumberTypeField(t *testing.T) {
	pnSid := "PN" + strings.Repeat("f", 32)
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"sid":          pnSid,
			"account_sid":  testAccountSid,
			"phone_number": "+14155550106",
			"api_version":  "2010-04-01",
			"uri": fmt.Sprintf("/2010-04-01/Accounts/%s/IncomingPhoneNumbers/%s.json",
				testAccountSid, pnSid),
			"type": "local",
			"capabilities": map[string]any{
				"voice": true, "sms": false, "mms": false, "fax": false,
			},
		}),
	}, nil)
	defer cleanup()

	pn, err := c.IncomingPhoneNumbers.Get(context.Background(), pnSid)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if pn.Type == nil {
		t.Fatal("Type: want non-nil pointer, got nil")
	}
	if *pn.Type != "local" {
		t.Fatalf("Type: want %q, got %q", "local", *pn.Type)
	}
}

// 27. Participant coaching fields round-trip JSON (spec v0.6.3).
func TestParticipantCoachingFields(t *testing.T) {
	cfSid := "CF" + strings.Repeat("c", 32)
	callSid := "CA" + strings.Repeat("d", 32)
	coachSid := "CA" + strings.Repeat("e", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"call_sid":                  callSid,
			"conference_sid":            cfSid,
			"account_sid":               testAccountSid,
			"muted":                     false,
			"hold":                      false,
			"coaching":                  true,
			"call_sid_to_coach":         coachSid,
			"queue_time":                "12",
			"start_conference_on_enter": true,
			"end_conference_on_exit":    false,
			"status":                    "connected",
			"api_version":               "2010-04-01",
			"uri":                       "/x",
		}),
	}, nil)
	defer cleanup()

	p, err := c.Conferences.GetParticipant(context.Background(), cfSid, callSid)
	if err != nil {
		t.Fatalf("GetParticipant: %v", err)
	}
	if !p.Coaching {
		t.Fatal("Coaching: want true")
	}
	if p.CallSidToCoach != coachSid {
		t.Fatalf("CallSidToCoach: want %q, got %q", coachSid, p.CallSidToCoach)
	}
	if p.QueueTime != "12" {
		t.Fatalf("QueueTime: want %q, got %q", "12", p.QueueTime)
	}
	if len(rec.requests) != 1 {
		t.Fatalf("requests: want 1, got %d", len(rec.requests))
	}
}

// 28. Recording.ErrorCode and StartConferenceRecordingAPI source (spec v0.6.3).
func TestRecordingErrorCodeAndSource(t *testing.T) {
	rSid := "RE" + strings.Repeat("f", 32)
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"sid":         rSid,
			"account_sid": testAccountSid,
			"call_sid":    "CA" + strings.Repeat("0", 32),
			"status":      "completed",
			"source":      "StartConferenceRecordingAPI",
			"error_code":  nil,
			"api_version": "2010-04-01",
			"uri":         "/x",
		}),
	}, nil)
	defer cleanup()

	rec, err := c.Recordings.Get(context.Background(), rSid, nil)
	if err != nil {
		t.Fatalf("Recordings.Get: %v", err)
	}
	if rec.Source != "StartConferenceRecordingAPI" {
		t.Fatalf("Source: want StartConferenceRecordingAPI, got %q", rec.Source)
	}
	if rec.ErrorCode != nil {
		t.Fatalf("ErrorCode: want nil, got %v", rec.ErrorCode)
	}
}

// 29. Calls.List sends StartTime/EndTime triple operators (spec v0.6.3).
func TestCallsListStartTimeEndTimeFilters(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{"calls": []any{}, "page": 0, "page_size": 50}),
	}, nil)
	defer cleanup()

	_, err := c.Calls.List(context.Background(), voiceml.ListCallsParams{
		StartTime:   "2026-05-01",
		StartTimeLt: "2026-05-02",
		StartTimeGt: "2026-04-30",
		EndTime:     "2026-05-21",
		EndTimeLt:   "2026-05-22",
		EndTimeGt:   "2026-05-20",
	})
	if err != nil {
		t.Fatalf("Calls.List: %v", err)
	}
	assertWire(t, "query", rec.requests[0].Query,
		"EndTime=2026-05-21&EndTime%3C=2026-05-22&EndTime%3E=2026-05-20&StartTime=2026-05-01&StartTime%3C=2026-05-02&StartTime%3E=2026-04-30")
}

// 30. Recordings.List sends DateCreated filters (spec v0.6.3).
func TestRecordingsListDateCreatedFilters(t *testing.T) {
	callSid := "CA" + strings.Repeat("1", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{"recordings": []any{}, "page": 0, "page_size": 50}),
	}, nil)
	defer cleanup()

	_, err := c.Recordings.List(context.Background(), voiceml.ListRecordingsParams{
		DateCreated:   "2026-05-01",
		DateCreatedLt: "2026-05-02",
		DateCreatedGt: "2026-04-30",
		CallSid:       callSid,
	})
	if err != nil {
		t.Fatalf("Recordings.List: %v", err)
	}
	assertWire(t, "query", rec.requests[0].Query,
		"CallSid="+callSid+"&DateCreated=2026-05-01&DateCreated%3C=2026-05-02&DateCreated%3E=2026-04-30")
}

// 31. Queues.Create accepts MaxSize=0 (unlimited, spec v0.6.3).
func TestQueuesCreateMaxSizeZero(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, map[string]any{
			"sid":               "QU" + strings.Repeat("0", 32),
			"account_sid":       testAccountSid,
			"friendly_name":     "unlimited",
			"current_size":      0,
			"max_size":          0,
			"average_wait_time": 0,
			"date_created":      "x",
			"date_updated":      "x",
			"uri":               "/x",
		}),
	}, nil)
	defer cleanup()

	maxSize := 0
	_, err := c.Queues.Create(context.Background(), voiceml.CreateQueueParams{
		FriendlyName: "unlimited",
		MaxSize:      &maxSize,
	})
	if err != nil {
		t.Fatalf("Queues.Create: %v", err)
	}
	assertWire(t, "body", string(rec.requests[0].Body), "FriendlyName=unlimited&MaxSize=0")
}

// 32. Calls.List sends PageToken (spec v0.6.4).
func TestCallsListPageToken(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{"calls": []any{}, "page": 0, "page_size": 50}),
	}, nil)
	defer cleanup()

	_, err := c.Calls.List(context.Background(), voiceml.ListCallsParams{
		PageToken: "cursor-abc123",
	})
	if err != nil {
		t.Fatalf("Calls.List: %v", err)
	}
	assertWire(t, "query", rec.requests[0].Query, "PageToken=cursor-abc123")
}

// 33. Conferences.CreateParticipant sends From/To (spec v0.6.6) plus the
// participant's conference lifecycle webhook (spec v0.9.4).
func TestConferencesCreateParticipant(t *testing.T) {
	confSid := "CF" + strings.Repeat("f", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, map[string]any{
			"call_sid":       "CA" + strings.Repeat("f", 32),
			"conference_sid": confSid,
			"account_sid":    testAccountSid,
			"status":         "queued",
			"api_version":    "2010-04-01",
			"uri":            "/x",
		}),
	}, nil)
	defer cleanup()

	_, err := c.Conferences.CreateParticipant(context.Background(), confSid, voiceml.CreateParticipantParams{
		From:                           "+14155550101",
		To:                             "+14155550100",
		ConferenceStatusCallback:       voiceml.String("https://example.com/conf"),
		ConferenceStatusCallbackMethod: voiceml.String("GET"),
		ConferenceStatusCallbackEvent:  voiceml.String("start end"),
	})
	if err != nil {
		t.Fatalf("Conferences.CreateParticipant: %v", err)
	}
	assertWire(t, "body", string(rec.requests[0].Body), "ConferenceStatusCallback=https%3A%2F%2Fexample.com%2Fconf&ConferenceStatusCallbackEvent=start+end&ConferenceStatusCallbackMethod=GET&From=%2B14155550101&To=%2B14155550100")
}

// 34. Calls.ListNotifications sends Log and MessageDate filters (spec v0.6.6).
func TestCallsListNotificationsFilters(t *testing.T) {
	callSid := "CA" + strings.Repeat("f", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"notifications": []any{},
			"page":          0,
			"page_size":     50,
			"uri":           "/Notifications",
		}),
	}, nil)
	defer cleanup()

	log := 1
	_, err := c.Calls.ListNotifications(context.Background(), callSid, voiceml.ListNotificationsParams{
		Log:           &log,
		MessageDate:   "2026-05-01",
		MessageDateLt: "2026-05-02",
		MessageDateGt: "2026-04-30",
	})
	if err != nil {
		t.Fatalf("Calls.ListNotifications: %v", err)
	}
	assertWire(t, "query", rec.requests[0].Query,
		"Log=1&MessageDate=2026-05-01&MessageDate%3C=2026-05-02&MessageDate%3E=2026-04-30")
}

// 35. Recordings.List sends IncludeSoftDeleted (spec v0.6.6).
func TestRecordingsListIncludeSoftDeleted(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{"recordings": []any{}}),
	}, nil)
	defer cleanup()

	include := true
	_, err := c.Recordings.List(context.Background(), voiceml.ListRecordingsParams{
		IncludeSoftDeleted: &include,
	})
	if err != nil {
		t.Fatalf("Recordings.List: %v", err)
	}
	assertWire(t, "query", rec.requests[0].Query, "IncludeSoftDeleted=true")
}

// --- Iterate (pagination) tests: see pagination_test.go ---

func messagePayload(sid string) map[string]any {
	if sid == "" {
		sid = "SM" + strings.Repeat("0", 32)
	}
	return map[string]any{
		"sid":          sid,
		"account_sid":  testAccountSid,
		"api_version":  "2010-04-01",
		"to":           "+14155550100",
		"from":         "+14155550101",
		"body":         "hello world",
		"status":       "sent",
		"num_segments": "1",
		"num_media":    "0",
		"direction":    "outbound-api",
		"date_created": "Mon, 19 May 2026 12:00:00 +0000",
		"date_updated": "Mon, 19 May 2026 12:00:00 +0000",
		"uri":          fmt.Sprintf("/2010-04-01/Accounts/%s/Messages/%s.json", testAccountSid, sid),
	}
}

// 41. Messages.Create — POST /Messages with To/Body/From, optionally
// scheduled via SendAt + ScheduleType (spec v0.9.4).
func TestMessagesCreate(t *testing.T) {
	sid := "SM" + strings.Repeat("1", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, messagePayload(sid)),
	}, nil)
	defer cleanup()

	msg, err := c.Messages.Create(context.Background(), voiceml.CreateMessageParams{
		To:                  "+14155550100",
		Body:                voiceml.String("hello world"),
		From:                voiceml.String("+14155550101"),
		MessagingServiceSid: voiceml.String("MG" + strings.Repeat("1", 32)),
		SendAt:              voiceml.String(voiceml.Time(time.Date(2026, 10, 22, 15, 4, 5, 0, time.UTC))),
		ScheduleType:        voiceml.String("fixed"),
	})
	if err != nil {
		t.Fatalf("Messages.Create: %v", err)
	}
	if msg.Sid != sid {
		t.Fatalf("sid: want %q, got %q", sid, msg.Sid)
	}

	req := rec.requests[0]
	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/Messages.json", testAccountSid)
	if req.Path != wantPath {
		t.Fatalf("path: want %q, got %q", wantPath, req.Path)
	}
	if req.Method != "POST" {
		t.Fatalf("method: want POST, got %s", req.Method)
	}
	assertWire(t, "body", string(req.Body), "Body=hello+world&From=%2B14155550101&MessagingServiceSid=MG11111111111111111111111111111111&ScheduleType=fixed&SendAt=2026-10-22T15%3A04%3A05Z&To=%2B14155550100")
}

// 42. Messages.Update — Body redaction sent as form field.
func TestMessagesUpdateRedact(t *testing.T) {
	sid := "SM" + strings.Repeat("2", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, messagePayload(sid)),
	}, nil)
	defer cleanup()

	_, err := c.Messages.Update(context.Background(), sid, voiceml.UpdateMessageParams{
		Body: voiceml.String(""),
	})
	if err != nil {
		t.Fatalf("Messages.Update: %v", err)
	}
	assertWire(t, "body", string(rec.requests[0].Body), "Body=")
}

// 43. Messages.List — accepts DateSent</> filters and PageToken pagination.
func TestMessagesListFilters(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"messages":  []any{messagePayload("")},
			"page":      0,
			"page_size": 50,
			"uri":       "/Messages",
		}),
	}, nil)
	defer cleanup()

	_, err := c.Messages.List(context.Background(), voiceml.ListMessagesParams{
		To:           voiceml.String("+14155550100"),
		DateSentMore: voiceml.String("2026-01-01"),
	})
	if err != nil {
		t.Fatalf("Messages.List: %v", err)
	}
	assertWire(t, "query", rec.requests[0].Query, "DateSent%3E=2026-01-01&To=%2B14155550100")
}

func callPaymentPayload(sid string) map[string]any {
	if sid == "" {
		sid = "PY" + strings.Repeat("0", 32)
	}
	return map[string]any{
		"sid":          sid,
		"account_sid":  testAccountSid,
		"call_sid":     "CA" + strings.Repeat("a", 32),
		"api_version":  "2010-04-01",
		"date_created": "Mon, 19 May 2026 12:00:00 +0000",
		"date_updated": "Mon, 19 May 2026 12:00:00 +0000",
		"uri":          fmt.Sprintf("/2010-04-01/Accounts/%s/Calls/%s/Payments/%s.json", testAccountSid, "CA"+strings.Repeat("a", 32), sid),
	}
}

// 44. Calls.StartPayment — POST /Calls/{sid}/Payments with full <Pay> body.
func TestCallsStartPayment(t *testing.T) {
	paySid := "PY" + strings.Repeat("1", 32)
	callSid := "CA" + strings.Repeat("a", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, callPaymentPayload(paySid)),
	}, nil)
	defer cleanup()

	tokenType := voiceml.PaymentTokenTypeOneTime
	method := voiceml.PaymentMethodCreditCard
	pay, err := c.Calls.StartPayment(context.Background(), callSid, voiceml.CreatePaymentParams{
		IdempotencyKey: voiceml.String("idem-1"),
		ChargeAmount:   voiceml.String("19.99"),
		Currency:       voiceml.String("USD"),
		Description:    voiceml.String("ticket"),
		PaymentMethod:  &method,
		TokenType:      &tokenType,
		PostalCode:     voiceml.Bool(false),
		SecurityCode:   voiceml.Bool(true),
	})
	if err != nil {
		t.Fatalf("StartPayment: %v", err)
	}
	if pay.Sid != paySid {
		t.Fatalf("sid: want %q, got %q", paySid, pay.Sid)
	}

	req := rec.requests[0]
	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/Calls/%s/Payments.json", testAccountSid, callSid)
	if req.Path != wantPath {
		t.Fatalf("path: want %q, got %q", wantPath, req.Path)
	}
	if req.Method != "POST" {
		t.Fatalf("method: want POST, got %s", req.Method)
	}
	assertWire(t, "body", string(req.Body),
		"ChargeAmount=19.99&Currency=USD&Description=ticket&IdempotencyKey=idem-1&PaymentMethod=credit-card&PostalCode=false&SecurityCode=true&TokenType=one-time")
}

// 45. Calls.UpdatePayment — Status=complete and Capture=... wire encoding.
func TestCallsUpdatePayment(t *testing.T) {
	paySid := "PY" + strings.Repeat("2", 32)
	callSid := "CA" + strings.Repeat("b", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(202, callPaymentPayload(paySid)),
		jsonStep(202, callPaymentPayload(paySid)),
	}, nil)
	defer cleanup()

	status := voiceml.PaymentSessionStatusComplete
	_, err := c.Calls.UpdatePayment(context.Background(), callSid, paySid, voiceml.UpdatePaymentParams{
		Status: &status,
	})
	if err != nil {
		t.Fatalf("UpdatePayment(complete): %v", err)
	}
	capture := voiceml.PaymentCaptureSecurityCode
	_, err = c.Calls.UpdatePayment(context.Background(), callSid, paySid, voiceml.UpdatePaymentParams{
		Capture: &capture,
	})
	if err != nil {
		t.Fatalf("UpdatePayment(capture): %v", err)
	}

	wantPath := fmt.Sprintf("/2010-04-01/Accounts/%s/Calls/%s/Payments/%s.json", testAccountSid, callSid, paySid)
	for i, r := range rec.requests {
		if r.Path != wantPath {
			t.Fatalf("req %d path: want %q, got %q", i, wantPath, r.Path)
		}
		if r.Method != "POST" {
			t.Fatalf("req %d method: want POST, got %s", i, r.Method)
		}
	}
	assertWire(t, "first body", string(rec.requests[0].Body), "Status=complete")
	assertWire(t, "second body", string(rec.requests[1].Body), "Capture=security-code")
}
