// Routes V2 (Inbound Processing Region) tests.

package voiceml_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

const (
	rv2DomainName = "ingress.example.com"
	rv2Sid        = "QQ00000000000000000000000000000000"
)

func rv2Payload() map[string]any {
	return map[string]any{
		"sid":           rv2Sid,
		"sip_domain":    rv2DomainName,
		"account_sid":   testAccountSid,
		"friendly_name": "ingress",
		"voice_region":  "us1",
		"url":           "https://voiceml.voicetel.com/v2/SipDomains/" + rv2DomainName,
		"date_created":  "2026-06-17T20:00:00Z",
		"date_updated":  "2026-06-17T20:00:00Z",
	}
}

func TestRoutesV2SipDomainsFetch(t *testing.T) {
	steps := []handlerStep{jsonStep(200, rv2Payload())}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	rv, err := c.RoutesV2.SipDomains.Fetch(context.Background(), rv2DomainName)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if rv.Sid != rv2Sid || rv.SipDomain != rv2DomainName {
		t.Fatalf("unexpected payload: %+v", rv)
	}
	if rv.VoiceRegion == nil || *rv.VoiceRegion != "us1" {
		t.Fatalf("voice region: %v", rv.VoiceRegion)
	}
	if rec.requests[0].Path != "/v2/SipDomains/"+rv2DomainName {
		t.Fatalf("path: %q", rec.requests[0].Path)
	}
	if strings.Contains(rec.requests[0].Path, testAccountSid) {
		t.Fatalf("path should not contain account sid: %q", rec.requests[0].Path)
	}
}

func TestRoutesV2SipDomainsUpdate(t *testing.T) {
	steps := []handlerStep{jsonStep(200, rv2Payload())}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	_, err := c.RoutesV2.SipDomains.Update(context.Background(), rv2DomainName, voiceml.UpdateRoutesV2SipDomainParams{
		VoiceRegion:  voiceml.String("ie1"),
		FriendlyName: voiceml.String("renamed"),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if rec.requests[0].Method != "POST" {
		t.Fatalf("method: %s", rec.requests[0].Method)
	}
	body, _ := url.ParseQuery(string(rec.requests[0].Body))
	if body.Get("VoiceRegion") != "ie1" || body.Get("FriendlyName") != "renamed" {
		t.Fatalf("body: %+v", body)
	}
}

func TestRoutesV2SipDomainsUpdatePartial(t *testing.T) {
	steps := []handlerStep{jsonStep(200, rv2Payload())}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	_, err := c.RoutesV2.SipDomains.Update(context.Background(), rv2DomainName, voiceml.UpdateRoutesV2SipDomainParams{
		VoiceRegion: voiceml.String("us1"),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	body, _ := url.ParseQuery(string(rec.requests[0].Body))
	if body.Get("VoiceRegion") != "us1" || len(body) != 1 {
		t.Fatalf("body should only contain VoiceRegion: %+v", body)
	}
}
