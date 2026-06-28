// SIP Trunking smoke tests — verify wire shapes for all 45 SIP operations.

package voiceml_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

const (
	sipDomainSid  = "SD11111111111111111111111111111111"
	sipCLSid      = "CL22222222222222222222222222222222"
	sipCRSid      = "CR33333333333333333333333333333333"
	sipACLSid     = "AL44444444444444444444444444444444"
	sipIPSid      = "IP55555555555555555555555555555555"
	sipMappingSid = "CL99999999999999999999999999999999"
)

func sipDomainPayload() map[string]any {
	return map[string]any{
		"sid":           sipDomainSid,
		"account_sid":   testAccountSid,
		"domain_name":   "ingress.example.com",
		"api_version":   "2010-04-01",
		"friendly_name": "ingress",
		"secure":        true,
		"date_created":  "Mon, 17 Jun 2026 12:00:00 +0000",
		"date_updated":  "Mon, 17 Jun 2026 12:00:00 +0000",
		"uri":           "/2010-04-01/Accounts/" + testAccountSid + "/SIP/Domains/" + sipDomainSid + ".json",
	}
}

func sipCredentialListPayload() map[string]any {
	return map[string]any{
		"sid":           sipCLSid,
		"account_sid":   testAccountSid,
		"friendly_name": "office-handsets",
		"date_created":  "Mon, 17 Jun 2026 12:00:00 +0000",
		"date_updated":  "Mon, 17 Jun 2026 12:00:00 +0000",
		"uri":           "/2010-04-01/Accounts/" + testAccountSid + "/SIP/CredentialLists/" + sipCLSid + ".json",
	}
}

func sipCredentialPayload() map[string]any {
	return map[string]any{
		"sid":                 sipCRSid,
		"account_sid":         testAccountSid,
		"credential_list_sid": sipCLSid,
		"username":            "alice",
		"date_created":        "Mon, 17 Jun 2026 12:00:00 +0000",
		"date_updated":        "Mon, 17 Jun 2026 12:00:00 +0000",
		"uri":                 "/2010-04-01/Accounts/" + testAccountSid + "/SIP/CredentialLists/" + sipCLSid + "/Credentials/" + sipCRSid + ".json",
	}
}

func sipIpaclPayload() map[string]any {
	return map[string]any{
		"sid":           sipACLSid,
		"account_sid":   testAccountSid,
		"friendly_name": "carrier-allowlist",
		"date_created":  "Mon, 17 Jun 2026 12:00:00 +0000",
		"date_updated":  "Mon, 17 Jun 2026 12:00:00 +0000",
		"uri":           "/2010-04-01/Accounts/" + testAccountSid + "/SIP/IpAccessControlLists/" + sipACLSid + ".json",
	}
}

func sipIpAddressPayload() map[string]any {
	return map[string]any{
		"sid":                        sipIPSid,
		"account_sid":                testAccountSid,
		"ip_access_control_list_sid": sipACLSid,
		"friendly_name":              "carrier-edge-1",
		"ip_address":                 "203.0.113.10",
		"cidr_prefix_length":         32,
		"date_created":               "Mon, 17 Jun 2026 12:00:00 +0000",
		"date_updated":               "Mon, 17 Jun 2026 12:00:00 +0000",
		"uri":                        "/2010-04-01/Accounts/" + testAccountSid + "/SIP/IpAccessControlLists/" + sipACLSid + "/IpAddresses/" + sipIPSid + ".json",
	}
}

func sipMappingPayload() map[string]any {
	return map[string]any{
		"sid":          sipMappingSid,
		"account_sid":  testAccountSid,
		"domain_sid":   sipDomainSid,
		"date_created": "Mon, 17 Jun 2026 12:00:00 +0000",
		"date_updated": "Mon, 17 Jun 2026 12:00:00 +0000",
		"uri":          "/2010-04-01/Accounts/" + testAccountSid + "/SIP/Domains/" + sipDomainSid + "/CredentialListMappings/" + sipMappingSid + ".json",
	}
}

func TestSIPDomains(t *testing.T) {
	steps := []handlerStep{
		jsonStep(200, map[string]any{"domains": []any{sipDomainPayload()}, "page": 0, "page_size": 50, "total": 1, "next_page_uri": nil, "uri": ""}),
		jsonStep(200, sipDomainPayload()),
		jsonStep(200, sipDomainPayload()),
		jsonStep(200, sipDomainPayload()),
		jsonStep(204, nil),
	}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	ctx := context.Background()

	out, err := c.SIP.Domains.List(ctx, voiceml.ListPageParams{})
	if err != nil || len(out.Domains) != 1 || out.Domains[0].DomainName != "ingress.example.com" {
		t.Fatalf("List: %v / %+v", err, out)
	}

	d, err := c.SIP.Domains.Create(ctx, voiceml.CreateSIPDomainParams{
		DomainName:   "ingress.example.com",
		FriendlyName: voiceml.String("ingress"),
		VoiceURL:     voiceml.String("https://hooks/voice"),
		VoiceMethod:  voiceml.String("POST"),
		Secure:       voiceml.Bool(true),
	})
	if err != nil || d.Sid != sipDomainSid {
		t.Fatalf("Create: %v %+v", err, d)
	}
	body, _ := url.ParseQuery(string(rec.requests[1].Body))
	if body.Get("DomainName") != "ingress.example.com" {
		t.Fatalf("Create body DomainName: %q", body.Get("DomainName"))
	}
	if body.Get("VoiceUrl") != "https://hooks/voice" || body.Get("VoiceMethod") != "POST" {
		t.Fatalf("Create body voice fields: %+v", body)
	}
	if body.Get("Secure") != "true" {
		t.Fatalf("Create body Secure: %q", body.Get("Secure"))
	}

	if _, err := c.SIP.Domains.Fetch(ctx, sipDomainSid); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if rec.requests[2].Method != "GET" || rec.requests[2].Path != "/2010-04-01/Accounts/"+testAccountSid+"/SIP/Domains/"+sipDomainSid+".json" {
		t.Fatalf("Fetch path: %+v", rec.requests[2])
	}

	if _, err := c.SIP.Domains.Update(ctx, sipDomainSid, voiceml.UpdateSIPDomainParams{FriendlyName: voiceml.String("renamed")}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	ubody, _ := url.ParseQuery(string(rec.requests[3].Body))
	if got := ubody.Get("FriendlyName"); got != "renamed" {
		t.Fatalf("Update body FriendlyName: %q", got)
	}
	if len(ubody) != 1 {
		t.Fatalf("Update body should only contain FriendlyName, got: %+v", ubody)
	}

	if err := c.SIP.Domains.Delete(ctx, sipDomainSid); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if rec.requests[4].Method != "DELETE" {
		t.Fatalf("Delete method: %s", rec.requests[4].Method)
	}
}

func TestSIPCredentialLists(t *testing.T) {
	steps := []handlerStep{
		jsonStep(200, map[string]any{"credential_lists": []any{sipCredentialListPayload()}, "page": 0, "page_size": 50, "total": 1, "next_page_uri": nil, "uri": ""}),
		jsonStep(200, sipCredentialListPayload()),
		jsonStep(200, sipCredentialListPayload()),
		jsonStep(200, sipCredentialListPayload()),
		jsonStep(204, nil),
	}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	ctx := context.Background()

	if _, err := c.SIP.CredentialLists.List(ctx, voiceml.ListPageParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if _, err := c.SIP.CredentialLists.Create(ctx, voiceml.CreateSIPCredentialListParams{FriendlyName: "office-handsets"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	body, _ := url.ParseQuery(string(rec.requests[1].Body))
	if body.Get("FriendlyName") != "office-handsets" {
		t.Fatalf("Create body: %q", body.Get("FriendlyName"))
	}

	if _, err := c.SIP.CredentialLists.Fetch(ctx, sipCLSid); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if _, err := c.SIP.CredentialLists.Update(ctx, sipCLSid, voiceml.UpdateSIPCredentialListParams{FriendlyName: voiceml.String("renamed")}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := c.SIP.CredentialLists.Delete(ctx, sipCLSid); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestSIPCredentialsNested(t *testing.T) {
	steps := []handlerStep{
		jsonStep(200, map[string]any{"credentials": []any{sipCredentialPayload()}, "page": 0, "page_size": 50, "total": 1, "next_page_uri": nil, "uri": ""}),
		jsonStep(200, sipCredentialPayload()),
		jsonStep(200, sipCredentialPayload()),
		jsonStep(200, sipCredentialPayload()),
		jsonStep(204, nil),
	}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	ctx := context.Background()

	if _, err := c.SIP.CredentialLists.ListCredentials(ctx, sipCLSid, voiceml.ListPageParams{}); err != nil {
		t.Fatalf("ListCredentials: %v", err)
	}
	expectedPath := "/2010-04-01/Accounts/" + testAccountSid + "/SIP/CredentialLists/" + sipCLSid + "/Credentials.json"
	if rec.requests[0].Path != expectedPath {
		t.Fatalf("List path: %q", rec.requests[0].Path)
	}

	if _, err := c.SIP.CredentialLists.CreateCredential(ctx, sipCLSid, voiceml.CreateSIPCredentialParams{Username: "alice", Password: "hunter2"}); err != nil {
		t.Fatalf("CreateCredential: %v", err)
	}
	body, _ := url.ParseQuery(string(rec.requests[1].Body))
	if body.Get("Username") != "alice" || body.Get("Password") != "hunter2" {
		t.Fatalf("Create body: %+v", body)
	}

	if _, err := c.SIP.CredentialLists.FetchCredential(ctx, sipCLSid, sipCRSid); err != nil {
		t.Fatalf("FetchCredential: %v", err)
	}
	if _, err := c.SIP.CredentialLists.UpdateCredential(ctx, sipCLSid, sipCRSid, voiceml.UpdateSIPCredentialParams{Password: "newpwd"}); err != nil {
		t.Fatalf("UpdateCredential: %v", err)
	}
	ubody, _ := url.ParseQuery(string(rec.requests[3].Body))
	if ubody.Get("Password") != "newpwd" || len(ubody) != 1 {
		t.Fatalf("Update body: %+v", ubody)
	}
	if err := c.SIP.CredentialLists.DeleteCredential(ctx, sipCLSid, sipCRSid); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestSIPIpAccessControlLists(t *testing.T) {
	steps := []handlerStep{
		jsonStep(200, map[string]any{"ip_access_control_lists": []any{sipIpaclPayload()}, "page": 0, "page_size": 50, "total": 1, "next_page_uri": nil, "uri": ""}),
		jsonStep(200, sipIpaclPayload()),
		jsonStep(200, sipIpaclPayload()),
		jsonStep(200, sipIpaclPayload()),
		jsonStep(204, nil),
	}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	ctx := context.Background()

	if _, err := c.SIP.IpAccessControlLists.List(ctx, voiceml.ListPageParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if _, err := c.SIP.IpAccessControlLists.Create(ctx, voiceml.CreateSIPIpAccessControlListParams{FriendlyName: "carrier-allowlist"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	body, _ := url.ParseQuery(string(rec.requests[1].Body))
	if body.Get("FriendlyName") != "carrier-allowlist" {
		t.Fatalf("Create body: %q", body.Get("FriendlyName"))
	}
	if _, err := c.SIP.IpAccessControlLists.Fetch(ctx, sipACLSid); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if _, err := c.SIP.IpAccessControlLists.Update(ctx, sipACLSid, voiceml.UpdateSIPIpAccessControlListParams{FriendlyName: voiceml.String("renamed")}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := c.SIP.IpAccessControlLists.Delete(ctx, sipACLSid); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestSIPIpAddressesNested(t *testing.T) {
	steps := []handlerStep{
		jsonStep(200, map[string]any{"ip_addresses": []any{sipIpAddressPayload()}, "page": 0, "page_size": 50, "total": 1, "next_page_uri": nil, "uri": ""}),
		jsonStep(200, sipIpAddressPayload()),
		jsonStep(200, sipIpAddressPayload()),
		jsonStep(200, sipIpAddressPayload()),
		jsonStep(204, nil),
	}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	ctx := context.Background()

	if _, err := c.SIP.IpAccessControlLists.ListIpAddresses(ctx, sipACLSid, voiceml.ListPageParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	cidr := 32
	if _, err := c.SIP.IpAccessControlLists.CreateIpAddress(ctx, sipACLSid, voiceml.CreateSIPIpAddressParams{
		FriendlyName: "carrier-edge-1", IpAddress: "203.0.113.10", CidrPrefixLength: &cidr,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	body, _ := url.ParseQuery(string(rec.requests[1].Body))
	if body.Get("FriendlyName") != "carrier-edge-1" || body.Get("IpAddress") != "203.0.113.10" || body.Get("CidrPrefixLength") != "32" {
		t.Fatalf("Create body: %+v", body)
	}
	if _, err := c.SIP.IpAccessControlLists.FetchIpAddress(ctx, sipACLSid, sipIPSid); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if _, err := c.SIP.IpAccessControlLists.UpdateIpAddress(ctx, sipACLSid, sipIPSid, voiceml.UpdateSIPIpAddressParams{IpAddress: voiceml.String("203.0.113.11")}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := c.SIP.IpAccessControlLists.DeleteIpAddress(ctx, sipACLSid, sipIPSid); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestSIPDomainHistoricalMappings(t *testing.T) {
	steps := []handlerStep{
		jsonStep(200, sipMappingPayload()), // CreateCredentialListMapping
		jsonStep(200, map[string]any{"credential_list_mappings": []any{sipMappingPayload()}, "page": 0, "page_size": 50, "total": 1, "next_page_uri": nil, "uri": ""}),
		jsonStep(200, sipMappingPayload()), // FetchCredentialListMapping
		jsonStep(204, nil),                 // DeleteCredentialListMapping
		jsonStep(200, sipMappingPayload()), // CreateIpAccessControlListMapping
	}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	ctx := context.Background()

	if _, err := c.SIP.Domains.CreateCredentialListMapping(ctx, sipDomainSid, voiceml.CreateSIPCredentialListMappingParams{CredentialListSid: sipCLSid}); err != nil {
		t.Fatalf("CreateCredentialListMapping: %v", err)
	}
	body, _ := url.ParseQuery(string(rec.requests[0].Body))
	if body.Get("CredentialListSid") != sipCLSid {
		t.Fatalf("Create body: %q", body.Get("CredentialListSid"))
	}
	expectedPath := "/2010-04-01/Accounts/" + testAccountSid + "/SIP/Domains/" + sipDomainSid + "/CredentialListMappings.json"
	if rec.requests[0].Path != expectedPath {
		t.Fatalf("Create path: %q", rec.requests[0].Path)
	}

	if _, err := c.SIP.Domains.ListCredentialListMappings(ctx, sipDomainSid, voiceml.ListPageParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if _, err := c.SIP.Domains.FetchCredentialListMapping(ctx, sipDomainSid, sipMappingSid); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if err := c.SIP.Domains.DeleteCredentialListMapping(ctx, sipDomainSid, sipMappingSid); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := c.SIP.Domains.CreateIpAccessControlListMapping(ctx, sipDomainSid, voiceml.CreateSIPIpAccessControlListMappingParams{IpAccessControlListSid: sipACLSid}); err != nil {
		t.Fatalf("CreateIpAccessControlListMapping: %v", err)
	}
	body4, _ := url.ParseQuery(string(rec.requests[4].Body))
	if body4.Get("IpAccessControlListSid") != sipACLSid {
		t.Fatalf("Create IPACL body: %q", body4.Get("IpAccessControlListSid"))
	}
}

func TestSIPDomainAuthNamespaces(t *testing.T) {
	steps := []handlerStep{
		jsonStep(200, sipMappingPayload()), // CreateAuthCallsCredentialListMapping
		jsonStep(200, sipMappingPayload()), // CreateAuthCallsIpAccessControlListMapping
		jsonStep(200, sipMappingPayload()), // CreateAuthRegistrationsCredentialListMapping
	}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	ctx := context.Background()

	if _, err := c.SIP.Domains.CreateAuthCallsCredentialListMapping(ctx, sipDomainSid, voiceml.CreateSIPCredentialListMappingParams{CredentialListSid: sipCLSid}); err != nil {
		t.Fatalf("Auth/Calls CL: %v", err)
	}
	if !strings.Contains(rec.requests[0].Path, "/Auth/Calls/CredentialListMappings") {
		t.Fatalf("Auth/Calls CL path: %q", rec.requests[0].Path)
	}

	if _, err := c.SIP.Domains.CreateAuthCallsIpAccessControlListMapping(ctx, sipDomainSid, voiceml.CreateSIPIpAccessControlListMappingParams{IpAccessControlListSid: sipACLSid}); err != nil {
		t.Fatalf("Auth/Calls ACL: %v", err)
	}
	if !strings.Contains(rec.requests[1].Path, "/Auth/Calls/IpAccessControlListMappings") {
		t.Fatalf("Auth/Calls ACL path: %q", rec.requests[1].Path)
	}

	if _, err := c.SIP.Domains.CreateAuthRegistrationsCredentialListMapping(ctx, sipDomainSid, voiceml.CreateSIPCredentialListMappingParams{CredentialListSid: sipCLSid}); err != nil {
		t.Fatalf("Auth/Registrations CL: %v", err)
	}
	if !strings.Contains(rec.requests[2].Path, "/Auth/Registrations/CredentialListMappings") {
		t.Fatalf("Auth/Registrations CL path: %q", rec.requests[2].Path)
	}
}
