package voiceml_test

// Wave-2 operations: Monitor v1 Alerts (list/fetch with the /v1 meta
// envelope), message delivery feedback and account key rotation (the
// rotated key is secret material and must never be formatted).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

func TestMonitorAlertsListDecodeAndNextPage(t *testing.T) {
	var srvURL string
	var mu sync.Mutex
	var uris []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		uris = append(uris, r.RequestURI)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("PageToken") == "" {
			_, _ = fmt.Fprintf(w, `{"alerts":[%s],"meta":{"page":0,"page_size":1,"next_page_url":%q}}`,
				alertPayload(), srvURL+"/v1/Alerts?PageSize=1&PageToken=PT1")
			return
		}
		_, _ = w.Write([]byte(`{"alerts":[],"meta":{"next_page_url":null}}`))
	}))
	defer srv.Close()
	srvURL = srv.URL
	c, err := voiceml.NewClient(voiceml.ClientOptions{AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	first, err := c.MonitorAlerts.List(context.Background(), voiceml.ListMonitorAlertsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Alerts) != 1 {
		t.Fatalf("alerts: %+v", first.Alerts)
	}
	a := first.Alerts[0]
	if a.Sid != "NO0123456789abcdef0123456789abcdef" || a.LogLevel != "error" || a.ErrorCode != "11200" ||
		a.ResourceSid == nil || *a.ResourceSid != "SM0123456789abcdef0123456789abcdef" ||
		a.RequestHeaders != nil || a.DateGenerated != "2026-10-17T09:41:00Z" {
		t.Fatalf("decoded alert: %+v", a)
	}
	if !first.Meta.HasNextPage() {
		t.Fatal("HasNextPage = false")
	}
	var second voiceml.MonitorAlertList
	ok, err := c.NextPage(context.Background(), first.Meta, &second)
	if err != nil || !ok {
		t.Fatalf("NextPage: ok=%v err=%v", ok, err)
	}
	if second.Meta.HasNextPage() || len(second.Alerts) != 0 {
		t.Fatalf("second page: %+v", second)
	}
	if len(uris) != 2 || uris[1] != "/v1/Alerts?PageSize=1&PageToken=PT1" {
		t.Fatalf("requests %q", uris)
	}
}

func alertPayload() string {
	return `{"sid":"NO0123456789abcdef0123456789abcdef","account_sid":"` + testAccountSid +
		`","alert_text":"HTTP retrieval failed","api_version":null,"date_created":"2026-10-17T09:40:00Z",` +
		`"date_generated":"2026-10-17T09:41:00Z","date_updated":"2026-10-17T09:41:00Z","error_code":"11200",` +
		`"log_level":"error","more_info":"https://voicetel.com/docs/errors/11200","request_method":"POST",` +
		`"request_url":"https://example.com/status","request_headers":null,"request_variables":null,` +
		`"resource_sid":"SM0123456789abcdef0123456789abcdef","response_body":null,"response_headers":null,` +
		`"service_sid":null,"url":"/v1/Alerts/NO0123456789abcdef0123456789abcdef"}`
}

func TestMonitorAlertsFetchDecodesDetail(t *testing.T) {
	body := strings.Replace(alertPayload(), `"request_headers":null`, `"request_headers":"X-Foo: bar"`, 1)
	c, _, cleanup := newClient(t, []handlerStep{jsonStep(200, json.RawMessage(body))}, nil)
	defer cleanup()
	a, err := c.MonitorAlerts.Fetch(context.Background(), "NO0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatal(err)
	}
	if a.RequestHeaders == nil || *a.RequestHeaders != "X-Foo: bar" {
		t.Fatalf("request_headers not decoded: %+v", a)
	}
}

func TestCreateFeedbackDecodes(t *testing.T) {
	c, rec, cleanup := newClient(t, []handlerStep{
		jsonStep(201, map[string]any{
			"account_sid": testAccountSid, "message_sid": "SM0123456789abcdef0123456789abcdef",
			"outcome": "confirmed", "date_created": "Fri, 17 Oct 2026 09:41:00 +0000",
			"date_updated": "Fri, 17 Oct 2026 09:41:00 +0000",
			"uri":          "/2010-04-01/Accounts/" + testAccountSid + "/Messages/SM0123456789abcdef0123456789abcdef/Feedback.json",
		}),
	}, nil)
	defer cleanup()
	fb, err := c.Messages.CreateFeedback(context.Background(), "SM0123456789abcdef0123456789abcdef",
		voiceml.CreateMessageFeedbackParams{Outcome: voiceml.String("confirmed")})
	if err != nil {
		t.Fatal(err)
	}
	if fb.Outcome != "confirmed" || fb.MessageSid != "SM0123456789abcdef0123456789abcdef" {
		t.Fatalf("feedback: %+v", fb)
	}
	assertWire(t, "feedback body", string(rec.requests[0].Body), "Outcome=confirmed")
}

func TestRotateKeyDecodesAndRedacts(t *testing.T) {
	const rotatedKey = "vk-rotated-do-not-print-9876543210"
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"sid": testAccountSid, "api_key": rotatedKey, "sip_username": "acme",
			"phone_regions": []string{"US"}, "allow_cidr": []string{}, "disabled": false,
			"stt_http_token_set": false, "tts_http_token_set": false, "stripe_secret_key_set": false,
			"pay_enabled": false, "sms_allowed_providers": []string{}, "stt_allowed_providers": []string{},
			"aws_secret_access_key_set": false, "created_at": "2026-10-17T09:41:00Z",
			"updated_at": "2026-10-17T09:41:00Z",
			"unknown_field_added_after_this_sdk_release": map[string]any{"x": 1},
		}),
	}, nil)
	defer cleanup()
	acct, err := c.Accounts.RotateKey(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if acct.APIKey != rotatedKey || acct.Sid != testAccountSid || acct.SipUsername != "acme" ||
		len(acct.AllowCidr) != 0 || acct.Disabled {
		t.Fatalf("account: %+v", acct)
	}
	for _, v := range []any{*acct, acct} {
		for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
			if out := fmt.Sprintf(verb, v); strings.Contains(out, rotatedKey) {
				t.Fatalf("rotated key leaked via %s: %s", verb, out)
			}
		}
	}
	// JSON marshalling is deliberately not redacted: the response is the
	// only delivery of the key.
	b, err := json.Marshal(acct)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), rotatedKey) {
		t.Fatal("json.Marshal must carry the rotated key (it is the delivery)")
	}

	// A 409 (code 20040) satisfies IsConflict.
	c2, _, cleanup2 := newClient(t, []handlerStep{
		jsonStep(409, map[string]any{"code": 20040, "message": "Idempotency-Key reuse", "status": 409}),
	}, nil)
	defer cleanup2()
	_, err = c2.Accounts.RotateKey(context.Background())
	if !voiceml.IsConflict(err) || !errors.Is(err, voiceml.ErrConflict) {
		t.Fatalf("409 should satisfy IsConflict, got %v", err)
	}
}
