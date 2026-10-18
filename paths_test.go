package voiceml_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

// uriRecorder answers {} and records each raw request URI.
func uriRecorder(t *testing.T) (*voiceml.Client, *[]string) {
	t.Helper()
	var uris []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uris = append(uris, r.RequestURI)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	c, err := voiceml.NewClient(voiceml.ClientOptions{AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: srv.URL, MaxRetries: voiceml.Int(0)})
	if err != nil {
		t.Fatal(err)
	}
	return c, &uris
}

func TestPathSegmentsArePercentEncoded(t *testing.T) {
	c, uris := uriRecorder(t)
	acct := "/2010-04-01/Accounts/" + testAccountSid
	cases := []struct {
		sid  string
		want string
	}{
		{"a?Foo=1", "a%3FFoo%3D1"},
		{"a#frag", "a%23frag"},
		{"a/b", "a%2Fb"},
		{"../../../Calls", "..%2F..%2F..%2FCalls"},
		{"50%", "50%25"},
		{"a b", "a%20b"},
		{"+14155550100", "%2B14155550100"},
		{"alice@example.com", "alice%40example.com"},
		{"café", "caf%C3%A9"},
		{"x.json", "x.json"},
		{"A-Z_a~z.0", "A-Z_a~z.0"},
	}
	for _, tc := range cases {
		*uris = nil
		if _, err := c.Calls.Get(context.Background(), tc.sid); err != nil {
			t.Fatalf("Calls.Get(%q): %v", tc.sid, err)
		}
		if got, want := (*uris)[0], acct+"/Calls/"+tc.want+".json"; got != want {
			t.Errorf("Calls.Get(%q) -> %q, want %q", tc.sid, got, want)
		}
		*uris = nil
		if _, err := c.ConversationsV1.FetchUser(context.Background(), tc.sid); err != nil {
			t.Fatalf("FetchUser(%q): %v", tc.sid, err)
		}
		if got, want := (*uris)[0], "/v1/Users/"+tc.want; got != want {
			t.Errorf("FetchUser(%q) -> %q, want %q", tc.sid, got, want)
		}
	}
}

func TestPricingNumberPathEncoding(t *testing.T) {
	c, uris := uriRecorder(t)
	if _, err := c.Pricing.V1.Voice.Numbers.Fetch(context.Background(), "+1 415 555 0100"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Pricing.V2.Voice.Numbers.Fetch(context.Background(), "+14155550100", voiceml.PricingNumberParams{OriginationNumber: voiceml.String("+14155550101")}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.RoutesV2.FetchPhoneNumber(context.Background(), "+14155550100"); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/v1/Voice/Numbers/%2B1%20415%20555%200100",
		"/v2/Voice/Numbers/%2B14155550100?OriginationNumber=%2B14155550101",
		"/v2/PhoneNumbers/%2B14155550100",
	}
	if strings.Join(*uris, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got %q\nwant %q", *uris, want)
	}
}

func TestJSONSuffixComesFromTheRoute(t *testing.T) {
	c, uris := uriRecorder(t)
	acct := "/2010-04-01/Accounts/" + testAccountSid
	_, _ = c.Recordings.Get(context.Background(), "Twilio.CURRENT", nil)
	_, _, _ = c.Recordings.GetAudio(context.Background(), "RE1")
	if d, err := c.Exports.Download(context.Background(), "2026-10-16"); err == nil {
		_ = d.Body.Close()
	}
	_, _ = c.Accounts.Fetch(context.Background())
	want := []string{
		acct + "/Recordings/Twilio.CURRENT.json",
		acct + "/Recordings/RE1.wav",
		acct + "/Exports/2026-10-16",
		acct + ".json",
	}
	if strings.Join(*uris, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got %q\nwant %q", *uris, want)
	}
}

func TestInvalidPathSegmentsAreRejectedBeforeSending(t *testing.T) {
	c, uris := uriRecorder(t)
	for _, sid := range []string{"", ".", ".."} {
		_, err := c.Calls.Get(context.Background(), sid)
		var ve *voiceml.ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("Calls.Get(%q): want *ValidationError, got %v", sid, err)
		}
		err = c.ConversationsV1.DeleteUser(context.Background(), sid)
		if !errors.As(err, &ve) {
			t.Errorf("DeleteUser(%q): want *ValidationError, got %v", sid, err)
		}
	}
	if len(*uris) != 0 {
		t.Fatalf("requests were sent: %q", *uris)
	}
}
