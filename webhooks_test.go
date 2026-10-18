package voiceml_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

const webhookKey = "test-api-key-0123456789"

func v1Params() url.Values {
	return url.Values{
		"CallSid":    {"CA0123456789abcdef0123456789abcdef"},
		"From":       {"+14155550100"},
		"To":         {"+14155550101"},
		"Digits":     {"1234"},
		"AccountSid": {"ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"},
	}
}

func TestWebhookSignatureVectors(t *testing.T) {
	cases := []struct {
		name   string
		url    string
		params url.Values
		sig    string
	}{
		{"V1 POST", "https://example.com/voice/answer?case=1", v1Params(), "RweoLyWqNC7H8+MLXPb7EPz18OI="},
		{"V1 POST with :443", "https://example.com:443/voice/answer?case=1", v1Params(), "xE0r2wPa6d1ktQeCFpPyINmTNYg="},
		{"V3 repeated keys", "https://example.com/hook", url.Values{
			"To": {"+14155550101"}, "Event": {"answered", "initiated"},
		}, "jCnpBgyg45aZYwUie5fmq4EYUic="},
		{"V4 GET", "https://example.com/voice/answer?CallSid=CA0123456789abcdef0123456789abcdef&From=%2B14155550100", nil, "9I4odWBHhF7TMIkIrNtPEPjD9fU="},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := voiceml.ComputeSignature(webhookKey, tc.url, tc.params); got != tc.sig {
				t.Fatalf("ComputeSignature = %q, want %q", got, tc.sig)
			}
			if !voiceml.ValidateRequest(webhookKey, tc.url, tc.params, tc.sig) {
				t.Fatal("ValidateRequest rejected a valid signature")
			}
		})
	}
}

func TestWebhookSignaturePortNormalisation(t *testing.T) {
	// The signature over the port-less URL validates against the URL with
	// :443 added, and vice versa.
	if !voiceml.ValidateRequest(webhookKey, "https://example.com:443/voice/answer?case=1", v1Params(), "RweoLyWqNC7H8+MLXPb7EPz18OI=") {
		t.Fatal("port-less signature rejected for :443 URL")
	}
	if !voiceml.ValidateRequest(webhookKey, "https://example.com/voice/answer?case=1", v1Params(), "xE0r2wPa6d1ktQeCFpPyINmTNYg=") {
		t.Fatal(":443 signature rejected for port-less URL")
	}
	// http and :80 behave the same way.
	sig := voiceml.ComputeSignature(webhookKey, "http://example.com/hook", nil)
	if !voiceml.ValidateRequest(webhookKey, "http://example.com:80/hook", nil, sig) {
		t.Fatal("http :80 variant rejected")
	}
	// A non-default port is significant.
	if voiceml.ValidateRequest(webhookKey, "https://example.com:8443/voice/answer?case=1", v1Params(), "RweoLyWqNC7H8+MLXPb7EPz18OI=") {
		t.Fatal("signature accepted for a different port")
	}
}

func TestWebhookSignatureFragmentIgnored(t *testing.T) {
	if !voiceml.ValidateRequest(webhookKey, "https://example.com/voice/answer?case=1#section", v1Params(), "RweoLyWqNC7H8+MLXPb7EPz18OI=") {
		t.Fatal("fragment should be stripped before signing")
	}
}

func TestWebhookSignatureNegative(t *testing.T) {
	tampered := v1Params()
	tampered.Set("Digits", "9999")
	if voiceml.ValidateRequest(webhookKey, "https://example.com/voice/answer?case=1", tampered, "RweoLyWqNC7H8+MLXPb7EPz18OI=") {
		t.Fatal("tampered value accepted")
	}
	if voiceml.ValidateRequest("wrong-key", "https://example.com/voice/answer?case=1", v1Params(), "RweoLyWqNC7H8+MLXPb7EPz18OI=") {
		t.Fatal("wrong key accepted")
	}
	// Reordering repeated values changes the signature.
	if voiceml.ValidateRequest(webhookKey, "https://example.com/hook", url.Values{
		"To": {"+14155550101"}, "Event": {"initiated", "answered"},
	}, "jCnpBgyg45aZYwUie5fmq4EYUic=") {
		t.Fatal("reordered repeated values accepted")
	}
	if voiceml.ValidateRequest(webhookKey, "https://example.com/hook", nil, "") {
		t.Fatal("empty signature accepted")
	}
	if voiceml.ValidateRequest("", "https://example.com/hook", nil, voiceml.ComputeSignature("", "https://example.com/hook", nil)) {
		t.Fatal("empty key accepted")
	}
}

func TestRequestValidatorHTTPRequest(t *testing.T) {
	v := voiceml.NewRequestValidator(webhookKey)
	form := v1Params()
	req := httptest.NewRequest(http.MethodPost, "http://internal:8080/voice/answer?case=1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(voiceml.SignatureHeader, "RweoLyWqNC7H8+MLXPb7EPz18OI=")
	if !v.ValidateHTTPRequest(req, "https://example.com/voice/answer?case=1") {
		t.Fatal("valid POST webhook rejected")
	}

	get := httptest.NewRequest(http.MethodGet, "http://internal/voice/answer?CallSid=CA0123456789abcdef0123456789abcdef&From=%2B14155550100", nil)
	get.Header.Set(voiceml.SignatureHeader, "9I4odWBHhF7TMIkIrNtPEPjD9fU=")
	if !v.ValidateHTTPRequest(get, "https://example.com/voice/answer?CallSid=CA0123456789abcdef0123456789abcdef&From=%2B14155550100") {
		t.Fatal("valid GET webhook rejected")
	}
}

func TestRequestValidatorRedactsKey(t *testing.T) {
	v := voiceml.NewRequestValidator(webhookKey)
	for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
		if out := fmt.Sprintf(verb, v); strings.Contains(out, webhookKey) {
			t.Fatalf("%s leaked the key: %s", verb, out)
		}
	}
}
