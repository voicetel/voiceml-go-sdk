package voiceml_test

// Wire-level tests for the operations added in this release and the request
// parameters the spec declares: exact method, URI, host, auth and body
// bytes.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

type wireCall struct {
	host, method, uri, contentType, body, auth string
}

// threeHosts serves {} on the default, messaging and conversations hosts
// and records every request.
func threeHosts(t *testing.T) (*voiceml.Client, *[]wireCall) {
	t.Helper()
	var calls []wireCall
	mk := func(host string) *httptest.Server {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b := new(strings.Builder)
			buf := make([]byte, 4096)
			for {
				n, err := r.Body.Read(buf)
				b.Write(buf[:n])
				if err != nil {
					break
				}
			}
			calls = append(calls, wireCall{host, r.Method, r.RequestURI, r.Header.Get("Content-Type"), b.String(), r.Header.Get("Authorization")})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
		t.Cleanup(srv.Close)
		return srv
	}
	def, msg, conv := mk("default"), mk("messaging"), mk("conversations")
	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: testAccountSid, APIKey: testAPIKey,
		BaseURL: def.URL, MessagingBaseURL: msg.URL, ConversationsBaseURL: conv.URL, MaxRetries: voiceml.Int(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	return c, &calls
}

func TestNewOperationsWire(t *testing.T) {
	c, calls := threeHosts(t)
	ctx := context.Background()
	acct := "/2010-04-01/Accounts/" + testAccountSid
	form := "application/x-www-form-urlencoded"

	steps := []struct {
		name string
		run  func() error
		want wireCall
	}{
		{"accounts.fetch", func() error { _, err := c.Accounts.Fetch(ctx); return err },
			wireCall{"default", "GET", acct + ".json", "", "", ""}},
		{"accounts.update", func() error {
			_, err := c.Accounts.Update(ctx, voiceml.UpdateAccountParams{FriendlyName: voiceml.String("Main"), AllowCidr: []string{}})
			return err
		}, wireCall{"default", "POST", acct + ".json", "application/json", `{"AllowCidr":[],"FriendlyName":"Main"}`, ""}},
		{"balance.fetch", func() error { _, err := c.Balance.Fetch(ctx); return err },
			wireCall{"default", "GET", acct + "/Balance.json", "", "", ""}},
		{"outgoing_caller_ids.list", func() error {
			_, err := c.OutgoingCallerIDs.List(ctx, voiceml.ListOutgoingCallerIDsParams{PhoneNumber: "+14155550100", FriendlyName: "Main"})
			return err
		}, wireCall{"default", "GET", acct + "/OutgoingCallerIds.json?FriendlyName=Main&PhoneNumber=%2B14155550100", "", "", ""}},
		{"outgoing_caller_ids.create", func() error {
			_, err := c.OutgoingCallerIDs.Create(ctx, voiceml.CreateOutgoingCallerIDParams{
				PhoneNumber:          "+14155550100",
				FriendlyName:         voiceml.String("Main Office"),
				CallDelay:            voiceml.Int(5),
				Extension:            voiceml.String("123"),
				StatusCallback:       voiceml.String("https://example.com/validation"),
				StatusCallbackMethod: voiceml.String("GET"),
			})
			return err
		}, wireCall{"default", "POST", acct + "/OutgoingCallerIds.json", form, "CallDelay=5&Extension=123&FriendlyName=Main+Office&PhoneNumber=%2B14155550100&StatusCallback=https%3A%2F%2Fexample.com%2Fvalidation&StatusCallbackMethod=GET", ""}},
		{"outgoing_caller_ids.update", func() error {
			_, err := c.OutgoingCallerIDs.Update(ctx, "PN1", voiceml.UpdateOutgoingCallerIDParams{FriendlyName: voiceml.String("Desk")})
			return err
		}, wireCall{"default", "POST", acct + "/OutgoingCallerIds/PN1.json", form, "FriendlyName=Desk", ""}},
		{"outgoing_caller_ids.delete", func() error { return c.OutgoingCallerIDs.Delete(ctx, "PN1") },
			wireCall{"default", "DELETE", acct + "/OutgoingCallerIds/PN1.json", "", "", ""}},
		{"media.list", func() error {
			_, err := c.Messages.ListMedia(ctx, "MM1", voiceml.ListMediaParams{DateCreatedGt: "2026-10-01"})
			return err
		}, wireCall{"default", "GET", acct + "/Messages/MM1/Media.json?DateCreated%3E=2026-10-01", "", "", ""}},
		{"media.fetch", func() error { _, err := c.Messages.FetchMedia(ctx, "MM1", "ME1"); return err },
			wireCall{"default", "GET", acct + "/Messages/MM1/Media/ME1.json", "", "", ""}},
		{"media.delete", func() error { return c.Messages.DeleteMedia(ctx, "MM1", "ME1") },
			wireCall{"default", "DELETE", acct + "/Messages/MM1/Media/ME1.json", "", "", ""}},
		{"exports.list", func() error { _, err := c.Exports.List(ctx); return err },
			wireCall{"default", "GET", acct + "/Exports.json", "", "", ""}},
		{"transcriptions.listAccount", func() error {
			_, err := c.Transcriptions.List(ctx, voiceml.ListTranscriptionsParams{PageSize: voiceml.Int(5)})
			return err
		}, wireCall{"default", "GET", acct + "/Transcriptions.json?PageSize=5", "", "", ""}},
		{"transcription.fetch", func() error { _, err := c.Transcriptions.Fetch(ctx, "TR1"); return err },
			wireCall{"default", "GET", acct + "/Transcriptions/TR1.json", "", "", ""}},
		{"transcription.delete", func() error { return c.Transcriptions.Delete(ctx, "TR1") },
			wireCall{"default", "DELETE", acct + "/Transcriptions/TR1.json", "", "", ""}},
		{"recording.transcriptions.list", func() error {
			_, err := c.Recordings.ListTranscriptions(ctx, "RE1", voiceml.ListTranscriptionsParams{})
			return err
		}, wireCall{"default", "GET", acct + "/Recordings/RE1/Transcriptions.json", "", "", ""}},
		{"recording.transcription.fetch", func() error { _, err := c.Recordings.FetchTranscription(ctx, "RE1", "TR1"); return err },
			wireCall{"default", "GET", acct + "/Recordings/RE1/Transcriptions/TR1.json", "", "", ""}},
		{"recording.transcription.delete", func() error { return c.Recordings.DeleteTranscription(ctx, "RE1", "TR1") },
			wireCall{"default", "DELETE", acct + "/Recordings/RE1/Transcriptions/TR1.json", "", "", ""}},
		{"conferenceRecordings.create", func() error {
			_, err := c.Conferences.StartRecording(ctx, "CF1", &voiceml.StartConferenceRecordingParams{RecordingStatusCallback: voiceml.String("https://example.com/rec")})
			return err
		}, wireCall{"default", "POST", acct + "/Conferences/CF1/Recordings.json", form, "RecordingStatusCallback=https%3A%2F%2Fexample.com%2Frec", ""}},
		{"usageRecords.listDaily", func() error {
			_, err := c.UsageRecords.ListDaily(ctx, voiceml.ListUsageRecordsParams{Category: "calls", StartDate: "-30days", IncludeSubaccounts: voiceml.Bool(false)})
			return err
		}, wireCall{"default", "GET", acct + "/Usage/Records/Daily.json?Category=calls&IncludeSubaccounts=false&StartDate=-30days", "", "", ""}},
		{"usageRecords.list", func() error { _, err := c.UsageRecords.List(ctx, voiceml.ListUsageRecordsParams{}); return err },
			wireCall{"default", "GET", acct + "/Usage/Records.json", "", "", ""}},
		{"sip_registrations.list", func() error { _, err := c.SIP.Registrations.List(ctx, voiceml.ListPageParams{}); return err },
			wireCall{"default", "GET", acct + "/SIP/Registrations.json", "", "", ""}},
		{"messaging.v1.services.phoneNumbers.list", func() error {
			_, err := c.MessagingV1.Services.ListPhoneNumbers(ctx, "MG1", voiceml.ListMessagingServicePhoneNumbersParams{PageSize: voiceml.Int(20)})
			return err
		}, wireCall{"messaging", "GET", "/v1/Services/MG1/PhoneNumbers?PageSize=20", "", "", ""}},
		{"messaging.v1.services.phoneNumbers.create", func() error {
			_, err := c.MessagingV1.Services.CreatePhoneNumber(ctx, "MG1", voiceml.CreateMessagingServicePhoneNumberParams{PhoneNumberSid: "PN1"})
			return err
		}, wireCall{"messaging", "POST", "/v1/Services/MG1/PhoneNumbers", form, "PhoneNumberSid=PN1", ""}},
		{"messaging.v1.services.phoneNumbers.fetch", func() error { _, err := c.MessagingV1.Services.FetchPhoneNumber(ctx, "MG1", "PN1"); return err },
			wireCall{"messaging", "GET", "/v1/Services/MG1/PhoneNumbers/PN1", "", "", ""}},
		{"messaging.v1.services.phoneNumbers.delete", func() error { return c.MessagingV1.Services.DeletePhoneNumber(ctx, "MG1", "PN1") },
			wireCall{"messaging", "DELETE", "/v1/Services/MG1/PhoneNumbers/PN1", "", "", ""}},
		{"monitor.v1.alerts.list", func() error {
			_, err := c.MonitorAlerts.List(ctx, voiceml.ListMonitorAlertsParams{
				LogLevel: voiceml.String("error"), StartDate: voiceml.String("2026-10-01"), EndDate: voiceml.String("2026-10-18"),
				V1PageParams: voiceml.V1PageParams{PageSize: voiceml.Int(50)},
			})
			return err
		}, wireCall{"default", "GET", "/v1/Alerts?EndDate=2026-10-18&LogLevel=error&PageSize=50&StartDate=2026-10-01", "", "", ""}},
		{"monitor.v1.alerts.fetch", func() error { _, err := c.MonitorAlerts.Fetch(ctx, "NO1"); return err },
			wireCall{"default", "GET", "/v1/Alerts/NO1", "", "", ""}},
		{"messages.feedback.create", func() error {
			_, err := c.Messages.CreateFeedback(ctx, "SM1", voiceml.CreateMessageFeedbackParams{Outcome: voiceml.String("confirmed")})
			return err
		}, wireCall{"default", "POST", acct + "/Messages/SM1/Feedback.json", form, "Outcome=confirmed", ""}},
		{"accounts.rotateKey", func() error { _, err := c.Accounts.RotateKey(ctx); return err },
			wireCall{"default", "POST", acct + "/RotateKey.json", "", "", ""}},
	}
	for _, st := range steps {
		*calls = nil
		if err := st.run(); err != nil {
			t.Fatalf("%s: %v", st.name, err)
		}
		if len(*calls) != 1 {
			t.Fatalf("%s: %d requests", st.name, len(*calls))
		}
		got := (*calls)[0]
		if got.auth != basicAuthHeader() {
			t.Errorf("%s: Authorization %q", st.name, got.auth)
		}
		got.auth = ""
		if got != st.want {
			t.Errorf("%s:\n got  %+v\n want %+v", st.name, got, st.want)
		}
	}
}

func TestMessagesCreateMMSAndValidation(t *testing.T) {
	c, calls := threeHosts(t)
	ctx := context.Background()
	_, err := c.Messages.Create(ctx, voiceml.CreateMessageParams{
		To:                   "+14155550100",
		MediaURL:             []string{"https://example.com/b.png", "https://example.com/a.png"},
		ValidityPeriod:       voiceml.Int(600),
		StatusCallback:       voiceml.String("https://example.com/dlr"),
		StatusCallbackMethod: voiceml.String("POST"),
	})
	if err != nil {
		t.Fatal(err)
	}
	assertWire(t, "MMS body", (*calls)[0].body,
		"MediaUrl=https%3A%2F%2Fexample.com%2Fb.png&MediaUrl=https%3A%2F%2Fexample.com%2Fa.png&StatusCallback=https%3A%2F%2Fexample.com%2Fdlr&StatusCallbackMethod=POST&To=%2B14155550100&ValidityPeriod=600")

	*calls = nil
	_, err = c.Messages.Create(ctx, voiceml.CreateMessageParams{To: "+14155550100"})
	var ve *voiceml.ValidationError
	if !errors.As(err, &ve) || len(*calls) != 0 {
		t.Fatalf("want *ValidationError and no request, got %v (%d requests)", err, len(*calls))
	}
}

func TestIncomingPhoneNumberSmsRoutingAndTypedCreates(t *testing.T) {
	c, calls := threeHosts(t)
	ctx := context.Background()
	acct := "/2010-04-01/Accounts/" + testAccountSid
	_, err := c.IncomingPhoneNumbers.Update(ctx, "PN1", voiceml.UpdateIncomingPhoneNumberParams{
		SmsURL: voiceml.String("https://example.com/sms"), SmsMethod: voiceml.String("POST"),
		SmsFallbackURL: voiceml.String("https://example.com/sms-fb"), SmsFallbackMethod: voiceml.String("GET"),
		SmsApplicationSid: voiceml.String("AP0123456789abcdef0123456789abcdef"),
	})
	if err != nil {
		t.Fatal(err)
	}
	assertWire(t, "update body", (*calls)[0].body,
		"SmsApplicationSid=AP0123456789abcdef0123456789abcdef&SmsFallbackMethod=GET&SmsFallbackUrl=https%3A%2F%2Fexample.com%2Fsms-fb&SmsMethod=POST&SmsUrl=https%3A%2F%2Fexample.com%2Fsms")

	*calls = nil
	_, err = c.IncomingPhoneNumbers.CreateTollFree(ctx, voiceml.CreateTypedIncomingPhoneNumberParams{PhoneNumber: "+14155550100", VoiceURL: voiceml.String("https://example.com/v")})
	if err != nil {
		t.Fatal(err)
	}
	if got := (*calls)[0]; got.uri != acct+"/IncomingPhoneNumbers/TollFree.json" || got.body != "PhoneNumber=%2B14155550100&VoiceUrl=https%3A%2F%2Fexample.com%2Fv" {
		t.Fatalf("typed create: %+v", got)
	}

	*calls = nil
	_, err = c.IncomingPhoneNumbers.List(ctx, &voiceml.ListIncomingPhoneNumbersParams{FriendlyName: "main", Beta: voiceml.Bool(false), Origin: "hosted"})
	if err != nil {
		t.Fatal(err)
	}
	assertWire(t, "list query", (*calls)[0].uri, acct+"/IncomingPhoneNumbers.json?Beta=false&FriendlyName=main&Origin=hosted")
}

func TestRecordingParamsAndApplicationCallbacks(t *testing.T) {
	c, calls := threeHosts(t)
	ctx := context.Background()
	_, err := c.Calls.StartRecording(ctx, "CA1", &voiceml.StartRecordingParams{RecordingTrack: voiceml.String("inbound"), PlayBeep: voiceml.Bool(false)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Calls.UpdateRecording(ctx, "CA1", "RE1", voiceml.UpdateCallRecordingParams{Status: "paused", PauseBehavior: voiceml.String("skip")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Applications.Create(ctx, voiceml.ApplicationParams{MessageStatusCallback: voiceml.String("https://example.com/dlr"), SmsStatusCallback: voiceml.String("https://example.com/sms-dlr")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SIP.Domains.List(ctx, voiceml.ListSIPParams{FriendlyName: "edge", ListPageParams: voiceml.ListPageParams{PageSize: voiceml.Int(10)}})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Calls.SendUserDefinedMessage(ctx, "CA1"); err != nil {
		t.Fatal(err)
	}
	assertWire(t, "start recording", (*calls)[0].body, "PlayBeep=false&RecordingTrack=inbound")
	assertWire(t, "update recording", (*calls)[1].body, "PauseBehavior=skip&Status=paused")
	assertWire(t, "application", (*calls)[2].body, "MessageStatusCallback=https%3A%2F%2Fexample.com%2Fdlr&SmsStatusCallback=https%3A%2F%2Fexample.com%2Fsms-dlr")
	assertWire(t, "sip list", (*calls)[3].uri, "/2010-04-01/Accounts/"+testAccountSid+"/SIP/Domains.json?FriendlyName=edge&PageSize=10")
	if udm := (*calls)[4]; udm.contentType != "application/x-www-form-urlencoded" || udm.body != "" {
		t.Fatalf("UserDefinedMessages: %+v", udm)
	}
}

func TestAccountsUpdateConflict(t *testing.T) {
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(409, map[string]any{"code": 20409, "message": "The account was modified by another request", "status": 409}),
	}, nil)
	defer cleanup()
	_, err := c.Accounts.Update(context.Background(), voiceml.UpdateAccountParams{FriendlyName: voiceml.String("Ops")})
	var apiErr *voiceml.APIError
	if !voiceml.IsConflict(err) || !errors.As(err, &apiErr) || apiErr.Code != 20409 {
		t.Fatalf("want 409 / 20409 conflict, got %v", err)
	}
}

func TestDecodingToleratesNullsUnknownEnumsAndExtraFields(t *testing.T) {
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(200, map[string]any{
			"sid": "CA1", "account_sid": nil, "status": "status-from-the-future", "direction": nil,
			"answered_by": "", "price": nil, "subresource_uris": nil, "brand_new_field": map[string]any{"x": 1},
		}),
	}, nil)
	defer cleanup()
	call, err := c.Calls.Get(context.Background(), "CA1")
	if err != nil {
		t.Fatal(err)
	}
	if call.Status != "status-from-the-future" || call.AccountSid != "" || call.SubresourceURIs != nil {
		t.Fatalf("decoded %+v", call)
	}
}
