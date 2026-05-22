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
	if voiceml.Version != "0.6.4" {
		t.Fatalf("Version: want 0.6.4, got %q", voiceml.Version)
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
					"phone_number": "+18005551234",
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
			"total":          1,
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
			PhoneNumber: "+18005551234",
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
	if pn.PhoneNumber != "+18005551234" {
		t.Fatalf("phone_number: want +18005551234, got %q", pn.PhoneNumber)
	}
	if !pn.Capabilities.Voice {
		t.Fatal("capabilities.voice: want true")
	}
	if list.PageSize != 50 || list.Page != 0 {
		t.Fatalf("pagination: want page=0 size=50, got page=%d size=%d", list.Page, list.PageSize)
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
	for _, want := range []string{
		"PhoneNumber=%2B18005551234",
		"Page=0",
		"PageSize=50",
	} {
		if !strings.Contains(req.Query, want) {
			t.Errorf("query missing %q; got %q", want, req.Query)
		}
	}
}

// 20. IncomingPhoneNumbers.Create — required PhoneNumber on the wire,
// optional VoiceUrl, 201 → IncomingPhoneNumber, path ends in ".json".
func TestIncomingPhoneNumbersCreate(t *testing.T) {
	pnSid := "PN" + strings.Repeat("b", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, map[string]any{
			"sid":          pnSid,
			"account_sid":  testAccountSid,
			"phone_number": "+18005550000",
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
			PhoneNumber: "+18005550000",
			VoiceURL:    voiceml.String("https://example.com/twiml"),
			VoiceMethod: voiceml.String("POST"),
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
	body := string(req.Body)
	for _, want := range []string{
		"PhoneNumber=%2B18005550000",
		"VoiceUrl=https%3A%2F%2Fexample.com%2Ftwiml",
		"VoiceMethod=POST",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q; got %q", want, body)
		}
	}
}

// 21. IncomingPhoneNumbers.Get — single-row fetch, request path includes
// {PN-sid}.json.
func TestIncomingPhoneNumbersGet(t *testing.T) {
	pnSid := "PN" + strings.Repeat("c", 32)
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"sid":          pnSid,
			"account_sid":  testAccountSid,
			"phone_number": "+18005550001",
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
			"phone_number": "+18005550002",
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
			VoiceURL: voiceml.String("https://example.com/new"),
		})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if pn.VoiceURL == nil || *pn.VoiceURL != "https://example.com/new" {
		t.Fatalf("voice_url didn't round-trip: %+v", pn.VoiceURL)
	}
	body := string(rec.requests[0].Body)
	if !strings.Contains(body, "VoiceUrl=https%3A%2F%2Fexample.com%2Fnew") {
		t.Errorf("body missing VoiceUrl; got %q", body)
	}
	// VoiceMethod was not set on the request — must not appear on the wire.
	if strings.Contains(body, "VoiceMethod=") {
		t.Errorf("body should not contain VoiceMethod (unset); got %q", body)
	}
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
			"sid":          rSid,
			"account_sid":  testAccountSid,
			"call_sid":     "CA" + strings.Repeat("b", 32),
			"status":       "completed",
			"api_version":  "2010-04-01",
			"uri":          "/x",
			"media_url":    mediaURL,
		}),
	}, nil)
	defer cleanup()

	rec, err := c.Recordings.Get(context.Background(), rSid)
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
			"phone_number": "+18005550003",
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
			"call_sid":                 callSid,
			"conference_sid":           cfSid,
			"account_sid":              testAccountSid,
			"muted":                    false,
			"hold":                     false,
			"coaching":                 true,
			"call_sid_to_coach":        coachSid,
			"queue_time":               "12",
			"start_conference_on_enter": true,
			"end_conference_on_exit":   false,
			"status":                   "connected",
			"api_version":              "2010-04-01",
			"uri":                      "/x",
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
			"sid":          rSid,
			"account_sid":  testAccountSid,
			"call_sid":     "CA" + strings.Repeat("0", 32),
			"status":       "completed",
			"source":       "StartConferenceRecordingAPI",
			"error_code":   nil,
			"api_version":  "2010-04-01",
			"uri":          "/x",
		}),
	}, nil)
	defer cleanup()

	rec, err := c.Recordings.Get(context.Background(), rSid)
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
	q := rec.requests[0].Query
	for _, want := range []string{
		"StartTime=2026-05-01",
		"StartTime%3C=2026-05-02",
		"StartTime%3E=2026-04-30",
		"EndTime=2026-05-21",
		"EndTime%3C=2026-05-22",
		"EndTime%3E=2026-05-20",
	} {
		if !strings.Contains(q, want) {
			t.Fatalf("query %q missing %q", q, want)
		}
	}
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
	q := rec.requests[0].Query
	for _, want := range []string{
		"DateCreated=2026-05-01",
		"DateCreated%3C=2026-05-02",
		"DateCreated%3E=2026-04-30",
		"CallSid=" + callSid,
	} {
		if !strings.Contains(q, want) {
			t.Fatalf("query %q missing %q", q, want)
		}
	}
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
	body := string(rec.requests[0].Body)
	if !strings.Contains(body, "MaxSize=0") {
		t.Fatalf("body: want MaxSize=0, got %q", body)
	}
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
	if !strings.Contains(rec.requests[0].Query, "PageToken=cursor-abc123") {
		t.Fatalf("query: want PageToken, got %q", rec.requests[0].Query)
	}
}
