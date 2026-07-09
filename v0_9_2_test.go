// Wire-shape tests for the v0.9.2 surface: per-product host routing, Messaging
// Service (/v1/Services on the messaging host), and Pricing v1/v2 (read-only on
// the default host). Mirrors the Python tests in tests/unit/test_v0_9_2.py.
//
// Messaging Service must ride the messaging host (that host is what
// disambiguates it from Conversation Service on the shared /v1/Services path).
// Pricing rides the default host. Host derivation itself is unit-tested
// directly in hosts_internal_test.go.

package voiceml_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

// newServer stands up a recorded httptest server that walks the given steps.
func newServer(t *testing.T, steps []handlerStep) (*recorder, *httptest.Server) {
	t.Helper()
	rec, handler := newRecorder(t, steps)
	return rec, httptest.NewServer(handler)
}

// parseForm decodes an application/x-www-form-urlencoded request body.
func parseForm(body []byte) url.Values {
	v, _ := url.ParseQuery(string(body))
	return v
}

func messagingServicePayload(sid string) map[string]any {
	return map[string]any{
		"sid":                 sid,
		"account_sid":         testAccountSid,
		"friendly_name":       "alerts",
		"inbound_request_url": "https://example.com/in",
		"sticky_sender":       true,
		"date_created":        "2026-07-08T00:00:00Z",
		"date_updated":        "2026-07-08T00:00:00Z",
		"url":                 "https://messaging.voicetel.com/v1/Services/" + sid,
	}
}

func messagingMeta() map[string]any {
	return map[string]any{
		"url":       "https://messaging.voicetel.com/v1/Services",
		"page":      0,
		"page_size": 50,
		"key":       "services",
	}
}

// TestProductHostsDerivedFromDefault confirms the client resolves the two
// product hosts from a *.voicetel.com base URL without any explicit override.
func TestProductHostsDerivedFromDefault(t *testing.T) {
	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: testAccountSid,
		APIKey:     testAPIKey,
		BaseURL:    "https://voiceml.voicetel.com",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.BaseURL != "https://voiceml.voicetel.com" {
		t.Fatalf("BaseURL: %q", c.BaseURL)
	}
	if c.MessagingBaseURL != "https://messaging.voicetel.com" {
		t.Fatalf("MessagingBaseURL: %q", c.MessagingBaseURL)
	}
	if c.ConversationsBaseURL != "https://conversations.voicetel.com" {
		t.Fatalf("ConversationsBaseURL: %q", c.ConversationsBaseURL)
	}
}

// TestV092ResourcesWired confirms the new resource groups are reachable.
func TestV092ResourcesWired(t *testing.T) {
	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: testAccountSid,
		APIKey:     testAPIKey,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.MessagingV1 == nil || c.MessagingV1.Services == nil {
		t.Fatal("MessagingV1.Services not wired")
	}
	if c.Pricing == nil ||
		c.Pricing.V1.Voice.Countries == nil ||
		c.Pricing.V1.Voice.Numbers == nil ||
		c.Pricing.V1.Messaging.Countries == nil ||
		c.Pricing.V1.PhoneNumbers.Countries == nil ||
		c.Pricing.V2.Voice.Countries == nil ||
		c.Pricing.V2.Voice.Numbers == nil ||
		c.Pricing.V2.Trunking.Countries == nil ||
		c.Pricing.V2.Trunking.Numbers == nil {
		t.Fatal("Pricing tree not fully wired")
	}
}

// TestMessagingServiceCRUDOnMessagingHost drives all five verbs and asserts
// every request lands on the messaging host (not the default one).
func TestMessagingServiceCRUDOnMessagingHost(t *testing.T) {
	sid := "MG" + strings.Repeat("1", 32)
	msgRec, msgSrv := newServer(t, []handlerStep{
		jsonStep(201, messagingServicePayload(sid)),
		jsonStep(200, map[string]any{
			"services": []any{messagingServicePayload(sid)},
			"meta":     messagingMeta(),
		}),
		jsonStep(200, messagingServicePayload(sid)),
		jsonStep(200, messagingServicePayload(sid)),
		jsonStep(204, nil),
	})
	defer msgSrv.Close()
	// The default host must receive nothing — zero steps means the recorder
	// fatals if any request arrives here.
	defRec, defSrv := newServer(t, nil)
	defer defSrv.Close()

	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid:       testAccountSid,
		APIKey:           testAPIKey,
		BaseURL:          defSrv.URL,
		MessagingBaseURL: msgSrv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx := context.Background()

	created, err := c.MessagingV1.Services.Create(ctx, voiceml.CreateMessagingServiceParams{
		FriendlyName:      "alerts",
		InboundRequestURL: voiceml.String("https://example.com/in"),
		StickySender:      voiceml.Bool(true),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Sid == nil || *created.Sid != sid || !strings.HasPrefix(*created.Sid, "MG") {
		t.Fatalf("created sid: %+v", created.Sid)
	}

	listed, err := c.MessagingV1.Services.List(ctx, voiceml.V1PageParams{PageSize: voiceml.Int(25)})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(listed.Services) != 1 {
		t.Fatalf("list len: %d", len(listed.Services))
	}

	fetched, err := c.MessagingV1.Services.Fetch(ctx, sid)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if fetched.Sid == nil || *fetched.Sid != sid {
		t.Fatalf("fetched sid: %+v", fetched.Sid)
	}

	updated, err := c.MessagingV1.Services.Update(ctx, sid, voiceml.UpdateMessagingServiceParams{
		FriendlyName: voiceml.String("renamed"),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Sid == nil || *updated.Sid != sid {
		t.Fatalf("updated sid: %+v", updated.Sid)
	}

	if err := c.MessagingV1.Services.Delete(ctx, sid); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if len(defRec.requests) != 0 {
		t.Fatalf("default host received %d requests, want 0", len(defRec.requests))
	}
	if len(msgRec.requests) != 5 {
		t.Fatalf("messaging host received %d requests, want 5", len(msgRec.requests))
	}

	// Verify verbs, paths, and bodies on the messaging host.
	if msgRec.requests[0].Method != "POST" || msgRec.requests[0].Path != "/v1/Services" {
		t.Fatalf("create req: %+v", msgRec.requests[0])
	}
	createBody := parseForm(msgRec.requests[0].Body)
	if createBody.Get("FriendlyName") != "alerts" ||
		createBody.Get("InboundRequestUrl") != "https://example.com/in" ||
		createBody.Get("StickySender") != "true" {
		t.Fatalf("create body: %v", createBody)
	}
	if msgRec.requests[1].Method != "GET" || !strings.Contains(msgRec.requests[1].Query, "PageSize=25") {
		t.Fatalf("list req: %+v", msgRec.requests[1])
	}
	if msgRec.requests[2].Path != "/v1/Services/"+sid {
		t.Fatalf("fetch path: %q", msgRec.requests[2].Path)
	}
	if msgRec.requests[3].Method != "POST" || msgRec.requests[3].Path != "/v1/Services/"+sid {
		t.Fatalf("update req: %+v", msgRec.requests[3])
	}
	updateBody := parseForm(msgRec.requests[3].Body)
	if updateBody.Get("FriendlyName") != "renamed" || len(updateBody) != 1 {
		t.Fatalf("update body should only carry FriendlyName: %v", updateBody)
	}
	if msgRec.requests[4].Method != "DELETE" {
		t.Fatalf("delete method: %s", msgRec.requests[4].Method)
	}
}

// TestMessagingServiceHostOverride confirms an explicit MessagingBaseURL wins
// over derivation even when the default host is a self-hosted one.
func TestMessagingServiceHostOverride(t *testing.T) {
	msgRec, msgSrv := newServer(t, []handlerStep{
		jsonStep(200, map[string]any{"services": []any{}, "meta": messagingMeta()}),
	})
	defer msgSrv.Close()

	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid:       testAccountSid,
		APIKey:           testAPIKey,
		BaseURL:          "https://pbx.acme.com",
		MessagingBaseURL: msgSrv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.MessagingV1.Services.List(context.Background(), voiceml.V1PageParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(msgRec.requests) != 1 {
		t.Fatalf("messaging host received %d requests, want 1", len(msgRec.requests))
	}
}

// TestConversationsRoutedToConversationsHost confirms the Conversations group
// is dispatched at the conversations host after the v0.9.2 routing change.
func TestConversationsRoutedToConversationsHost(t *testing.T) {
	convRec, convSrv := newServer(t, []handlerStep{
		jsonStep(200, map[string]any{
			"conversations": []any{},
			"meta":          map[string]any{"page": 0, "page_size": 50},
		}),
	})
	defer convSrv.Close()
	defRec, defSrv := newServer(t, nil)
	defer defSrv.Close()

	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid:           testAccountSid,
		APIKey:               testAPIKey,
		BaseURL:              defSrv.URL,
		ConversationsBaseURL: convSrv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.ConversationsV1.ListConversations(context.Background(), voiceml.ListConversationsParams{}); err != nil {
		t.Fatalf("ListConversations: %v", err)
	}
	if len(defRec.requests) != 0 {
		t.Fatalf("default host received %d requests, want 0", len(defRec.requests))
	}
	if len(convRec.requests) != 1 {
		t.Fatalf("conversations host received %d requests, want 1", len(convRec.requests))
	}
	if convRec.requests[0].Path != "/v1/Conversations" {
		t.Fatalf("conversations path: %q", convRec.requests[0].Path)
	}
}

// TestPricingV1VoiceCountriesAndNumber exercises the v1 voice pricing reads on
// the default host and confirms the number segment is URL-encoded on the wire.
func TestPricingV1VoiceCountriesAndNumber(t *testing.T) {
	countries := map[string]any{
		"countries": []any{
			map[string]any{
				"country":     "United States",
				"iso_country": "US",
				"url":         "https://voiceml.voicetel.com/v1/Voice/Countries/US",
			},
		},
		"meta": map[string]any{"page": 0, "page_size": 50},
	}
	country := map[string]any{
		"country":     "United States",
		"iso_country": "US",
		"outbound_prefix_prices": []any{
			map[string]any{
				"prefixes":      []any{"1"},
				"base_price":    "0.013",
				"current_price": "0.013",
				"friendly_name": "United States & Canada",
			},
		},
		"inbound_call_prices": []any{
			map[string]any{"base_price": "0.0085", "current_price": "0.0085", "number_type": "local"},
		},
		"price_unit": "USD",
		"url":        "https://voiceml.voicetel.com/v1/Voice/Countries/US",
	}
	number := map[string]any{
		"number":              "+18005551234",
		"country":             "United States",
		"iso_country":         "US",
		"outbound_call_price": map[string]any{"base_price": "0.013", "current_price": "0.013"},
		"inbound_call_price":  map[string]any{"base_price": "0.0085", "current_price": "0.0085", "number_type": "toll free"},
		"price_unit":          "USD",
		"url":                 "https://voiceml.voicetel.com/v1/Voice/Numbers/+18005551234",
	}
	rec, srv := newServer(t, []handlerStep{
		jsonStep(200, countries),
		jsonStep(200, country),
		jsonStep(200, number),
	})
	defer srv.Close()

	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: testAccountSid,
		APIKey:     testAPIKey,
		BaseURL:    srv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx := context.Background()

	listed, err := c.Pricing.V1.Voice.Countries.List(ctx, voiceml.V1PageParams{})
	if err != nil {
		t.Fatalf("Countries.List: %v", err)
	}
	if len(listed.Countries) != 1 || listed.Countries[0].IsoCountry == nil || *listed.Countries[0].IsoCountry != "US" {
		t.Fatalf("countries: %+v", listed.Countries)
	}

	fetched, err := c.Pricing.V1.Voice.Countries.Fetch(ctx, "US")
	if err != nil {
		t.Fatalf("Countries.Fetch: %v", err)
	}
	if len(fetched.OutboundPrefixPrices) != 1 || len(fetched.OutboundPrefixPrices[0].Prefixes) != 1 ||
		fetched.OutboundPrefixPrices[0].Prefixes[0] != "1" {
		t.Fatalf("country body: %+v", fetched)
	}

	num, err := c.Pricing.V1.Voice.Numbers.Fetch(ctx, "+18005551234")
	if err != nil {
		t.Fatalf("Numbers.Fetch: %v", err)
	}
	if num.InboundCallPrice == nil || num.InboundCallPrice.NumberType == nil ||
		*num.InboundCallPrice.NumberType != "toll free" {
		t.Fatalf("number body: %+v", num)
	}

	// Number segment must be percent-encoded on the wire (+ -> %2B).
	if rec.requests[2].RawPath != "/v1/Voice/Numbers/%2B18005551234" {
		t.Fatalf("number segment not URL-encoded: raw=%q", rec.requests[2].RawPath)
	}
	if rec.requests[2].Path != "/v1/Voice/Numbers/+18005551234" {
		t.Fatalf("decoded number path: %q", rec.requests[2].Path)
	}
}

// TestPricingV2VoiceNumberWithOrigination confirms the OriginationNumber query
// knob is sent and percent-encoded.
func TestPricingV2VoiceNumberWithOrigination(t *testing.T) {
	payload := map[string]any{
		"destination_number": "+18005551234",
		"origination_number": "+15551112222",
		"country":            "United States",
		"iso_country":        "US",
		"outbound_call_prices": []any{
			map[string]any{"origination_prefixes": []any{"1"}, "base_price": "0.013", "current_price": "0.013"},
		},
		"inbound_call_price": map[string]any{"base_price": "0.0085", "current_price": "0.0085", "number_type": "local"},
		"price_unit":         "USD",
		"url":                "https://voiceml.voicetel.com/v2/Voice/Numbers/+18005551234",
	}
	rec, srv := newServer(t, []handlerStep{jsonStep(200, payload)})
	defer srv.Close()

	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: testAccountSid,
		APIKey:     testAPIKey,
		BaseURL:    srv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	got, err := c.Pricing.V2.Voice.Numbers.Fetch(context.Background(), "+18005551234", voiceml.PricingNumberParams{
		OriginationNumber: voiceml.String("+15551112222"),
	})
	if err != nil {
		t.Fatalf("Numbers.Fetch: %v", err)
	}
	if got.OriginationNumber == nil || *got.OriginationNumber != "+15551112222" {
		t.Fatalf("origination number: %+v", got.OriginationNumber)
	}
	if !strings.Contains(rec.requests[0].Query, "OriginationNumber=%2B15551112222") {
		t.Fatalf("query missing encoded OriginationNumber: %q", rec.requests[0].Query)
	}
	if rec.requests[0].RawPath != "/v2/Voice/Numbers/%2B18005551234" {
		t.Fatalf("dest number not URL-encoded: %q", rec.requests[0].RawPath)
	}
}

// TestPricingV2TrunkingCountry exercises a v2 trunking country fetch.
func TestPricingV2TrunkingCountry(t *testing.T) {
	payload := map[string]any{
		"country":     "United States",
		"iso_country": "US",
		"terminating_prefix_prices": []any{
			map[string]any{
				"origination_prefixes": []any{"1"},
				"destination_prefixes": []any{"1"},
				"base_price":           "0.013",
				"current_price":        "0.013",
				"friendly_name":        "US",
			},
		},
		"originating_call_prices": []any{
			map[string]any{"base_price": "0.0085", "current_price": "0.0085", "number_type": "local"},
		},
		"price_unit": "USD",
		"url":        "https://voiceml.voicetel.com/v2/Trunking/Countries/US",
	}
	rec, srv := newServer(t, []handlerStep{jsonStep(200, payload)})
	defer srv.Close()

	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: testAccountSid,
		APIKey:     testAPIKey,
		BaseURL:    srv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	got, err := c.Pricing.V2.Trunking.Countries.Fetch(context.Background(), "US")
	if err != nil {
		t.Fatalf("Countries.Fetch: %v", err)
	}
	if len(got.TerminatingPrefixPrices) != 1 || got.TerminatingPrefixPrices[0].FriendlyName == nil ||
		*got.TerminatingPrefixPrices[0].FriendlyName != "US" {
		t.Fatalf("trunking country body: %+v", got)
	}
	if rec.requests[0].Path != "/v2/Trunking/Countries/US" {
		t.Fatalf("path: %q", rec.requests[0].Path)
	}
}

// TestPricingV1MessagingCountriesList confirms a v1 messaging countries list on
// the default host.
func TestPricingV1MessagingCountriesList(t *testing.T) {
	rec, srv := newServer(t, []handlerStep{
		jsonStep(200, map[string]any{"countries": []any{}, "meta": map[string]any{"page": 0}}),
	})
	defer srv.Close()

	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: testAccountSid,
		APIKey:     testAPIKey,
		BaseURL:    srv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	listed, err := c.Pricing.V1.Messaging.Countries.List(context.Background(), voiceml.V1PageParams{})
	if err != nil {
		t.Fatalf("Countries.List: %v", err)
	}
	if len(listed.Countries) != 0 {
		t.Fatalf("countries: %+v", listed.Countries)
	}
	if rec.requests[0].Path != "/v1/Messaging/Countries" {
		t.Fatalf("path: %q", rec.requests[0].Path)
	}
}
