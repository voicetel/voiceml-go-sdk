// v0.9.0 surface tests — Conversations v1, Voice v1, RoutesV2 PhoneNumber.
//
// Every test stands up an httptest server, asserts on the request the SDK
// sent (method + path + form/query encoding), and feeds back a canned wire
// shape decoded into the response model. Table-driven where it pays;
// resource-specific where the body shape needs precision.

package voiceml_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

// ---------------------------------------------------------------------------
// Sids and fixtures.
// ---------------------------------------------------------------------------

const (
	convSid     = "CH" + "0123456789abcdef0123456789abcdef"
	convMsgSid  = "IM" + "0123456789abcdef0123456789abcdef"
	convPartSid = "MB" + "0123456789abcdef0123456789abcdef"
	convHookSid = "WH" + "0123456789abcdef0123456789abcdef"
	convRcptSid = "DY" + "0123456789abcdef0123456789abcdef"
	convRoleSid = "RL" + "0123456789abcdef0123456789abcdef"
	convUserSid = "US" + "0123456789abcdef0123456789abcdef"
	convCredSid = "CR" + "0123456789abcdef0123456789abcdef"
	convAddrSid = "IG" + "0123456789abcdef0123456789abcdef"
	convChatSvc = "IS" + "0123456789abcdef0123456789abcdef"
	v1IpRecSid  = "IL" + "0123456789abcdef0123456789abcdef"
	v1SipMapSid = "IB" + "0123456789abcdef0123456789abcdef"
	v1ByocSid   = "BY" + "0123456789abcdef0123456789abcdef"
	v1PolicySid = "NY" + "0123456789abcdef0123456789abcdef"
	v1TargetSid = "NE" + "0123456789abcdef0123456789abcdef"
	v1SipDomSid = "SD" + "0123456789abcdef0123456789abcdef"
	rv2PnNumber = "+18005551234"
	rv2PnSid    = "QQ" + "0000000000000000000000000000000a"
)

// ---------------------------------------------------------------------------
// Client wiring — both new services and the extended RoutesV2 are reachable.
// ---------------------------------------------------------------------------

func TestV09ClientWiring(t *testing.T) {
	c, _, done := newClient(t, nil, nil)
	defer done()
	if c.ConversationsV1 == nil {
		t.Fatal("ConversationsV1 not wired up")
	}
	if c.VoiceV1 == nil {
		t.Fatal("VoiceV1 not wired up")
	}
	if c.RoutesV2 == nil || c.RoutesV2.SipDomains == nil {
		t.Fatal("RoutesV2 / RoutesV2.SipDomains not wired up")
	}
}

// ---------------------------------------------------------------------------
// RoutesV2 PhoneNumber — fetch + update.
// ---------------------------------------------------------------------------

func rv2PhoneNumberPayload() map[string]any {
	return map[string]any{
		"sid":           rv2PnSid,
		"phone_number":  rv2PnNumber,
		"account_sid":   testAccountSid,
		"friendly_name": "main",
		"voice_region":  "us1",
		"url":           "https://example/v2/PhoneNumbers/" + rv2PnNumber,
		"date_created":  "2026-06-27T12:00:00Z",
		"date_updated":  "2026-06-27T12:00:00Z",
	}
}

func TestRoutesV2PhoneNumberFetch(t *testing.T) {
	c, rec, done := newClient(t, []handlerStep{jsonStep(200, rv2PhoneNumberPayload())}, nil)
	defer done()
	pn, err := c.RoutesV2.FetchPhoneNumber(context.Background(), rv2PnNumber)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if pn.Sid != rv2PnSid || pn.PhoneNumber != rv2PnNumber {
		t.Fatalf("payload: %+v", pn)
	}
	if pn.VoiceRegion == nil || *pn.VoiceRegion != "us1" {
		t.Fatalf("voice region: %v", pn.VoiceRegion)
	}
	if got, want := rec.requests[0].Path, "/v2/PhoneNumbers/"+rv2PnNumber; got != want {
		t.Fatalf("path: %q (want %q)", got, want)
	}
	if strings.Contains(rec.requests[0].Path, testAccountSid) {
		t.Fatalf("path leaked account sid: %q", rec.requests[0].Path)
	}
}

func TestRoutesV2PhoneNumberUpdate(t *testing.T) {
	c, rec, done := newClient(t, []handlerStep{jsonStep(200, rv2PhoneNumberPayload())}, nil)
	defer done()
	_, err := c.RoutesV2.UpdatePhoneNumber(context.Background(), rv2PnNumber, voiceml.UpdateRoutesV2PhoneNumberParams{
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

// ---------------------------------------------------------------------------
// Voice v1 — table-driven path/method tests across all six families.
// ---------------------------------------------------------------------------

func TestVoiceV1Paths(t *testing.T) {
	cases := []struct {
		name       string
		response   map[string]any
		method     string
		path       string
		invoke     func(c *voiceml.Client) error
		assertBody func(t *testing.T, body url.Values, query url.Values)
	}{
		{
			name: "CreateIpRecord",
			response: map[string]any{
				"sid": v1IpRecSid, "account_sid": testAccountSid,
				"ip_address": "203.0.113.10", "cidr_prefix_length": 32,
				"date_created": "x", "date_updated": "x", "url": "u",
			},
			method: "POST", path: "/v1/IpRecords",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.CreateIpRecord(context.Background(), voiceml.CreateVoiceV1IpRecordParams{
					IpAddress:        "203.0.113.10",
					FriendlyName:     voiceml.String("carrier-a"),
					CidrPrefixLength: voiceml.Int(24),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("IpAddress") != "203.0.113.10" || body.Get("FriendlyName") != "carrier-a" || body.Get("CidrPrefixLength") != "24" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:     "ListIpRecords",
			response: map[string]any{"ip_records": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/IpRecords",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.ListIpRecords(context.Background(), voiceml.V1PageParams{PageSize: voiceml.Int(25)})
				return err
			},
			assertBody: func(t *testing.T, _, q url.Values) {
				if q.Get("PageSize") != "25" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name:     "FetchIpRecord",
			response: map[string]any{"sid": v1IpRecSid, "account_sid": testAccountSid, "ip_address": "x", "cidr_prefix_length": 32, "date_created": "x", "date_updated": "x", "url": "u"},
			method:   "GET", path: "/v1/IpRecords/" + v1IpRecSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.FetchIpRecord(context.Background(), v1IpRecSid)
				return err
			},
		},
		{
			name:     "UpdateIpRecord",
			response: map[string]any{"sid": v1IpRecSid, "account_sid": testAccountSid, "ip_address": "x", "cidr_prefix_length": 32, "date_created": "x", "date_updated": "x", "url": "u"},
			method:   "POST", path: "/v1/IpRecords/" + v1IpRecSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.UpdateIpRecord(context.Background(), v1IpRecSid, voiceml.UpdateVoiceV1IpRecordParams{
					FriendlyName: voiceml.String("renamed"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "renamed" || len(body) != 1 {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteIpRecord",
			method: "DELETE", path: "/v1/IpRecords/" + v1IpRecSid,
			invoke: func(c *voiceml.Client) error {
				return c.VoiceV1.DeleteIpRecord(context.Background(), v1IpRecSid)
			},
		},
		{
			name: "CreateSourceIpMapping",
			response: map[string]any{
				"sid": v1SipMapSid, "ip_record_sid": v1IpRecSid, "sip_domain_sid": v1SipDomSid,
				"date_created": "x", "date_updated": "x", "url": "u",
			},
			method: "POST", path: "/v1/SourceIpMappings",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.CreateSourceIpMapping(context.Background(), voiceml.CreateVoiceV1SourceIpMappingParams{
					IpRecordSid: v1IpRecSid, SipDomainSid: v1SipDomSid,
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("IpRecordSid") != v1IpRecSid || body.Get("SipDomainSid") != v1SipDomSid {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:     "ListSourceIpMappings",
			response: map[string]any{"source_ip_mappings": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/SourceIpMappings",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.ListSourceIpMappings(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name:     "UpdateSourceIpMapping",
			response: map[string]any{"sid": v1SipMapSid, "ip_record_sid": v1IpRecSid, "sip_domain_sid": v1SipDomSid, "date_created": "x", "date_updated": "x", "url": "u"},
			method:   "POST", path: "/v1/SourceIpMappings/" + v1SipMapSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.UpdateSourceIpMapping(context.Background(), v1SipMapSid, voiceml.UpdateVoiceV1SourceIpMappingParams{
					SipDomainSid: v1SipDomSid,
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("SipDomainSid") != v1SipDomSid {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteSourceIpMapping",
			method: "DELETE", path: "/v1/SourceIpMappings/" + v1SipMapSid,
			invoke: func(c *voiceml.Client) error {
				return c.VoiceV1.DeleteSourceIpMapping(context.Background(), v1SipMapSid)
			},
		},
		{
			name: "CreateByocTrunk",
			response: map[string]any{
				"sid": v1ByocSid, "account_sid": testAccountSid,
				"date_created": "x", "date_updated": "x", "url": "u",
			},
			method: "POST", path: "/v1/ByocTrunks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.CreateByocTrunk(context.Background(), voiceml.CreateVoiceV1ByocTrunkParams{
					FriendlyName:      voiceml.String("carrier-x"),
					VoiceURL:          voiceml.String("https://example.com/hook"),
					VoiceMethod:       voiceml.String("POST"),
					CnamLookupEnabled: voiceml.Bool(true),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "carrier-x" ||
					body.Get("VoiceUrl") != "https://example.com/hook" ||
					body.Get("VoiceMethod") != "POST" ||
					body.Get("CnamLookupEnabled") != "true" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:     "ListByocTrunks",
			response: map[string]any{"byoc_trunks": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/ByocTrunks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.ListByocTrunks(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name:     "UpdateByocTrunk",
			response: map[string]any{"sid": v1ByocSid, "account_sid": testAccountSid, "date_created": "x", "date_updated": "x", "url": "u"},
			method:   "POST", path: "/v1/ByocTrunks/" + v1ByocSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.UpdateByocTrunk(context.Background(), v1ByocSid, voiceml.UpdateVoiceV1ByocTrunkParams{
					CnamLookupEnabled: voiceml.Bool(false),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("CnamLookupEnabled") != "false" || len(body) != 1 {
					t.Fatalf("partial body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteByocTrunk",
			method: "DELETE", path: "/v1/ByocTrunks/" + v1ByocSid,
			invoke: func(c *voiceml.Client) error {
				return c.VoiceV1.DeleteByocTrunk(context.Background(), v1ByocSid)
			},
		},
		{
			name: "CreateConnectionPolicy",
			response: map[string]any{
				"sid": v1PolicySid, "account_sid": testAccountSid,
				"date_created": "x", "date_updated": "x", "url": "u", "links": map[string]any{},
			},
			method: "POST", path: "/v1/ConnectionPolicies",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.CreateConnectionPolicy(context.Background(), voiceml.CreateVoiceV1ConnectionPolicyParams{
					FriendlyName: voiceml.String("origination"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "origination" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:     "ListConnectionPolicies",
			response: map[string]any{"connection_policies": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/ConnectionPolicies",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.ListConnectionPolicies(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "CreateConnectionPolicyTarget",
			response: map[string]any{
				"sid": v1TargetSid, "connection_policy_sid": v1PolicySid,
				"target": "sip:edge@example.com", "priority": 10, "weight": 10,
				"account_sid":  testAccountSid,
				"date_created": "x", "date_updated": "x", "url": "u",
			},
			method: "POST", path: "/v1/ConnectionPolicies/" + v1PolicySid + "/Targets",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.CreateConnectionPolicyTarget(context.Background(), v1PolicySid, voiceml.CreateVoiceV1ConnectionPolicyTargetParams{
					Target:   "sip:edge@example.com",
					Priority: voiceml.Int(5),
					Weight:   voiceml.Int(20),
					Enabled:  voiceml.Bool(true),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Target") != "sip:edge@example.com" ||
					body.Get("Priority") != "5" ||
					body.Get("Weight") != "20" ||
					body.Get("Enabled") != "true" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:     "ListConnectionPolicyTargets",
			response: map[string]any{"targets": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/ConnectionPolicies/" + v1PolicySid + "/Targets",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.ListConnectionPolicyTargets(context.Background(), v1PolicySid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name:   "DeleteConnectionPolicyTarget",
			method: "DELETE", path: "/v1/ConnectionPolicies/" + v1PolicySid + "/Targets/" + v1TargetSid,
			invoke: func(c *voiceml.Client) error {
				return c.VoiceV1.DeleteConnectionPolicyTarget(context.Background(), v1PolicySid, v1TargetSid)
			},
		},
		{
			name:     "FetchSettings",
			response: map[string]any{"dialing_permissions_inheritance": false, "url": "u"},
			method:   "GET", path: "/v1/Settings",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.FetchSettings(context.Background())
				return err
			},
		},
		{
			name:     "UpdateSettings",
			response: map[string]any{"dialing_permissions_inheritance": true, "url": "u"},
			method:   "POST", path: "/v1/Settings",
			invoke: func(c *voiceml.Client) error {
				_, err := c.VoiceV1.UpdateSettings(context.Background(), voiceml.UpdateVoiceV1SettingsParams{
					DialingPermissionsInheritance: voiceml.Bool(true),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("DialingPermissionsInheritance") != "true" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.response
			steps := []handlerStep{jsonStep(200, body)}
			c, rec, done := newClient(t, steps, nil)
			defer done()
			if err := tc.invoke(c); err != nil {
				t.Fatalf("invoke: %v", err)
			}
			r := rec.requests[0]
			if r.Method != tc.method {
				t.Fatalf("method: %s (want %s)", r.Method, tc.method)
			}
			if r.Path != tc.path {
				t.Fatalf("path: %s (want %s)", r.Path, tc.path)
			}
			if strings.Contains(r.Path, testAccountSid) {
				t.Fatalf("path leaked account sid: %q", r.Path)
			}
			if tc.assertBody != nil {
				formBody, _ := url.ParseQuery(string(r.Body))
				query, _ := url.ParseQuery(r.Query)
				tc.assertBody(t, formBody, query)
			}
		})
	}
}

// VoiceV1 list decode — make sure the meta envelope round-trips.
func TestVoiceV1ListDecode(t *testing.T) {
	steps := []handlerStep{jsonStep(200, map[string]any{
		"ip_records": []map[string]any{
			{
				"sid": v1IpRecSid, "account_sid": testAccountSid,
				"ip_address": "203.0.113.10", "cidr_prefix_length": 32,
				"date_created": "2026-06-27T12:00:00Z",
				"date_updated": "2026-06-27T12:00:00Z",
				"url":          "https://example/v1/IpRecords/" + v1IpRecSid,
			},
		},
		"meta": map[string]any{
			"page": 0, "page_size": 50,
			"url":            "https://example/v1/IpRecords?PageSize=50",
			"first_page_url": "https://example/v1/IpRecords?Page=0",
			"key":            "ip_records",
		},
	})}
	c, _, done := newClient(t, steps, nil)
	defer done()
	out, err := c.VoiceV1.ListIpRecords(context.Background(), voiceml.V1PageParams{})
	if err != nil {
		t.Fatalf("ListIpRecords: %v", err)
	}
	if len(out.IpRecords) != 1 {
		t.Fatalf("ip_records len: %d", len(out.IpRecords))
	}
	if out.IpRecords[0].Sid == nil || *out.IpRecords[0].Sid != v1IpRecSid {
		t.Fatalf("ip record sid: %+v", out.IpRecords[0])
	}
	if out.Meta.Page == nil || *out.Meta.Page != 0 {
		t.Fatalf("meta.page: %+v", out.Meta.Page)
	}
	if out.Meta.Key == nil || *out.Meta.Key != "ip_records" {
		t.Fatalf("meta.key: %+v", out.Meta.Key)
	}
}

// ---------------------------------------------------------------------------
// Conversations v1 — table-driven path/method tests across all 15 families.
// ---------------------------------------------------------------------------

func TestConversationsV1Paths(t *testing.T) {
	convPayload := map[string]any{
		"sid": convSid, "account_sid": testAccountSid, "state": "active",
		"attributes": "{}", "date_created": "x", "date_updated": "x", "url": "u",
	}
	msgPayload := map[string]any{
		"sid": convMsgSid, "conversation_sid": convSid, "account_sid": testAccountSid,
		"index": 0, "attributes": "{}", "date_created": "x", "date_updated": "x", "url": "u",
	}
	partPayload := map[string]any{
		"sid": convPartSid, "conversation_sid": convSid, "account_sid": testAccountSid,
		"attributes": "{}", "date_created": "x", "date_updated": "x", "url": "u",
	}
	hookPayload := map[string]any{
		"sid": convHookSid, "conversation_sid": convSid, "account_sid": testAccountSid,
		"date_created": "x", "date_updated": "x",
	}
	rcptPayload := map[string]any{
		"sid": convRcptSid, "conversation_sid": convSid, "account_sid": testAccountSid,
		"message_sid": convMsgSid, "status": "delivered", "error_code": 0,
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	rolePayload := map[string]any{
		"sid": convRoleSid, "account_sid": testAccountSid,
		"type": "conversation", "permissions": []string{"sendMessage"},
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	userPayload := map[string]any{
		"sid": convUserSid, "account_sid": testAccountSid,
		"attributes": "{}", "date_created": "x", "date_updated": "x", "url": "u",
	}
	credPayload := map[string]any{
		"sid": convCredSid, "account_sid": testAccountSid, "type": "fcm",
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	cfgPayload := map[string]any{
		"account_sid": testAccountSid, "url": "u",
	}
	cfgHookPayload := map[string]any{
		"account_sid": testAccountSid, "method": "POST", "target": "webhook", "url": "u",
	}
	addrPayload := map[string]any{
		"sid": convAddrSid, "account_sid": testAccountSid,
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	pcPayload := map[string]any{
		"account_sid": testAccountSid, "conversation_state": "active",
		"conversation_date_created": "x", "conversation_date_updated": "x",
	}
	ucPayload := map[string]any{
		"account_sid": testAccountSid, "conversation_state": "active",
		"notification_level": "default", "date_created": "x", "date_updated": "x", "url": "u",
	}
	svcPayload := map[string]any{
		"sid": convChatSvc, "account_sid": testAccountSid,
		"date_created": "x", "date_updated": "x", "url": "u",
	}

	cases := []struct {
		name       string
		response   map[string]any
		listShape  map[string]any
		method     string
		path       string
		invoke     func(c *voiceml.Client) error
		assertBody func(t *testing.T, body url.Values, query url.Values)
	}{
		// --- Conversations -------------------------------------------------
		{
			name: "CreateConversation", response: convPayload,
			method: "POST", path: "/v1/Conversations",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateConversation(context.Background(), voiceml.CreateConversationRequest{
					FriendlyName:      voiceml.String("Support"),
					State:             voiceml.String("active"),
					TimersInactive:    voiceml.String("PT1H"),
					BindingsEmailAddr: voiceml.String("support@example.com"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "Support" ||
					body.Get("State") != "active" ||
					body.Get("Timers.Inactive") != "PT1H" ||
					body.Get("Bindings.Email.Address") != "support@example.com" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListConversations", listShape: map[string]any{"conversations": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Conversations",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListConversations(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchConversation", response: convPayload,
			method: "GET", path: "/v1/Conversations/" + convSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchConversation(context.Background(), convSid)
				return err
			},
		},
		{
			name: "UpdateConversation", response: convPayload,
			method: "POST", path: "/v1/Conversations/" + convSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateConversation(context.Background(), convSid, voiceml.UpdateConversationRequest{
					State: voiceml.String("closed"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("State") != "closed" || len(body) != 1 {
					t.Fatalf("partial body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteConversation",
			method: "DELETE", path: "/v1/Conversations/" + convSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteConversation(context.Background(), convSid)
			},
		},
		// --- Messages -------------------------------------------------------
		{
			name: "CreateMessage", response: msgPayload,
			method: "POST", path: "/v1/Conversations/" + convSid + "/Messages",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateMessage(context.Background(), convSid, voiceml.CreateMessageRequest{
					Author: voiceml.String("alice"),
					Body:   voiceml.String("hello"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Author") != "alice" || body.Get("Body") != "hello" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListMessages", listShape: map[string]any{"messages": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Conversations/" + convSid + "/Messages",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListMessages(context.Background(), convSid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchMessage", response: msgPayload,
			method: "GET", path: "/v1/Conversations/" + convSid + "/Messages/" + convMsgSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchMessage(context.Background(), convSid, convMsgSid)
				return err
			},
		},
		{
			name: "UpdateMessage", response: msgPayload,
			method: "POST", path: "/v1/Conversations/" + convSid + "/Messages/" + convMsgSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateMessage(context.Background(), convSid, convMsgSid, voiceml.UpdateMessageRequest{
					Body: voiceml.String("edited"),
				})
				return err
			},
		},
		{
			name:   "DeleteMessage",
			method: "DELETE", path: "/v1/Conversations/" + convSid + "/Messages/" + convMsgSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteMessage(context.Background(), convSid, convMsgSid)
			},
		},
		// --- Participants ---------------------------------------------------
		{
			name: "CreateParticipant", response: partPayload,
			method: "POST", path: "/v1/Conversations/" + convSid + "/Participants",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateParticipant(context.Background(), convSid, voiceml.CreateParticipantRequest{
					Identity:                voiceml.String("alice"),
					MessagingBindingAddress: voiceml.String("+15551234567"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Identity") != "alice" || body.Get("MessagingBinding.Address") != "+15551234567" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListParticipants", listShape: map[string]any{"participants": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Conversations/" + convSid + "/Participants",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListParticipants(context.Background(), convSid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "UpdateParticipant", response: partPayload,
			method: "POST", path: "/v1/Conversations/" + convSid + "/Participants/" + convPartSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateParticipant(context.Background(), convSid, convPartSid, voiceml.UpdateParticipantRequest{
					LastReadMessageIndex: voiceml.Int(42),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("LastReadMessageIndex") != "42" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteParticipant",
			method: "DELETE", path: "/v1/Conversations/" + convSid + "/Participants/" + convPartSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteParticipant(context.Background(), convSid, convPartSid)
			},
		},
		// --- Receipts (list+fetch only) ------------------------------------
		{
			name: "ListMessageReceipts", listShape: map[string]any{"delivery_receipts": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Conversations/" + convSid + "/Messages/" + convMsgSid + "/Receipts",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListMessageReceipts(context.Background(), convSid, convMsgSid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchMessageReceipt", response: rcptPayload,
			method: "GET", path: "/v1/Conversations/" + convSid + "/Messages/" + convMsgSid + "/Receipts/" + convRcptSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchMessageReceipt(context.Background(), convSid, convMsgSid, convRcptSid)
				return err
			},
		},
		// --- Scoped Webhooks ------------------------------------------------
		{
			name: "CreateScopedWebhook", response: hookPayload,
			method: "POST", path: "/v1/Conversations/" + convSid + "/Webhooks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateScopedWebhook(context.Background(), convSid, voiceml.CreateScopedWebhookRequest{
					Target:              "webhook",
					ConfigurationURL:    voiceml.String("https://example.com/hook"),
					ConfigurationMethod: voiceml.String("POST"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Target") != "webhook" ||
					body.Get("Configuration.Url") != "https://example.com/hook" ||
					body.Get("Configuration.Method") != "POST" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListScopedWebhooks", listShape: map[string]any{"webhooks": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Conversations/" + convSid + "/Webhooks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListScopedWebhooks(context.Background(), convSid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "UpdateScopedWebhook", response: hookPayload,
			method: "POST", path: "/v1/Conversations/" + convSid + "/Webhooks/" + convHookSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateScopedWebhook(context.Background(), convSid, convHookSid, voiceml.UpdateScopedWebhookRequest{
					ConfigurationURL: voiceml.String("https://example.com/new"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Configuration.Url") != "https://example.com/new" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteScopedWebhook",
			method: "DELETE", path: "/v1/Conversations/" + convSid + "/Webhooks/" + convHookSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteScopedWebhook(context.Background(), convSid, convHookSid)
			},
		},
		// --- Roles ----------------------------------------------------------
		{
			name: "CreateRole", response: rolePayload,
			method: "POST", path: "/v1/Roles",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateRole(context.Background(), voiceml.CreateRoleRequest{
					FriendlyName: "agent",
					Type:         "conversation",
					Permission:   []string{"sendMessage", "leaveConversation"},
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "agent" || body.Get("Type") != "conversation" {
					t.Fatalf("body: %+v", body)
				}
				perms := body["Permission"]
				if len(perms) != 2 || perms[0] != "sendMessage" || perms[1] != "leaveConversation" {
					t.Fatalf("repeated Permission: %+v", perms)
				}
			},
		},
		{
			name: "ListRoles", listShape: map[string]any{"roles": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Roles",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListRoles(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "UpdateRole", response: rolePayload,
			method: "POST", path: "/v1/Roles/" + convRoleSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateRole(context.Background(), convRoleSid, voiceml.UpdateRoleRequest{
					Permission: []string{"sendMessage"},
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if perms := body["Permission"]; len(perms) != 1 || perms[0] != "sendMessage" {
					t.Fatalf("perms: %+v", perms)
				}
			},
		},
		{
			name:   "DeleteRole",
			method: "DELETE", path: "/v1/Roles/" + convRoleSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteRole(context.Background(), convRoleSid)
			},
		},
		// --- Users ----------------------------------------------------------
		{
			name: "CreateUser", response: userPayload,
			method: "POST", path: "/v1/Users",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateUser(context.Background(), voiceml.CreateUserRequest{
					Identity: "alice", FriendlyName: voiceml.String("Alice"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Identity") != "alice" || body.Get("FriendlyName") != "Alice" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListUsers", listShape: map[string]any{"users": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Users",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListUsers(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "UpdateUser", response: userPayload,
			method: "POST", path: "/v1/Users/" + convUserSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateUser(context.Background(), convUserSid, voiceml.UpdateUserRequest{
					FriendlyName: voiceml.String("Alice II"),
				})
				return err
			},
		},
		{
			name:   "DeleteUser",
			method: "DELETE", path: "/v1/Users/" + convUserSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteUser(context.Background(), convUserSid)
			},
		},
		// --- UserConversations ---------------------------------------------
		{
			name: "ListUserConversations", listShape: map[string]any{"conversations": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Users/" + convUserSid + "/Conversations",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListUserConversations(context.Background(), convUserSid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchUserConversation", response: ucPayload,
			method: "GET", path: "/v1/Users/" + convUserSid + "/Conversations/" + convSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchUserConversation(context.Background(), convUserSid, convSid)
				return err
			},
		},
		{
			name: "UpdateUserConversation", response: ucPayload,
			method: "POST", path: "/v1/Users/" + convUserSid + "/Conversations/" + convSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateUserConversation(context.Background(), convUserSid, convSid, voiceml.UpdateUserConversationRequest{
					NotificationLevel: voiceml.String("muted"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("NotificationLevel") != "muted" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteUserConversation",
			method: "DELETE", path: "/v1/Users/" + convUserSid + "/Conversations/" + convSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteUserConversation(context.Background(), convUserSid, convSid)
			},
		},
		// --- Credentials ----------------------------------------------------
		{
			name: "CreateCredential", response: credPayload,
			method: "POST", path: "/v1/Credentials",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateCredential(context.Background(), voiceml.CreateCredentialRequest{
					Type: "fcm", Sandbox: voiceml.Bool(true), APIKey: voiceml.String("k"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Type") != "fcm" || body.Get("Sandbox") != "true" || body.Get("ApiKey") != "k" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListCredentials", listShape: map[string]any{"credentials": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Credentials",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListCredentials(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "UpdateCredential", response: credPayload,
			method: "POST", path: "/v1/Credentials/" + convCredSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateCredential(context.Background(), convCredSid, voiceml.UpdateCredentialRequest{
					FriendlyName: voiceml.String("renamed"),
				})
				return err
			},
		},
		{
			name:   "DeleteCredential",
			method: "DELETE", path: "/v1/Credentials/" + convCredSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteCredential(context.Background(), convCredSid)
			},
		},
		// --- Configuration --------------------------------------------------
		{
			name: "FetchConfiguration", response: cfgPayload,
			method: "GET", path: "/v1/Configuration",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchConfiguration(context.Background())
				return err
			},
		},
		{
			name: "UpdateConfiguration", response: cfgPayload,
			method: "POST", path: "/v1/Configuration",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateConfiguration(context.Background(), voiceml.UpdateConfigurationRequest{
					DefaultInactiveTimer: voiceml.String("PT12H"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("DefaultInactiveTimer") != "PT12H" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "FetchConfigurationWebhook", response: cfgHookPayload,
			method: "GET", path: "/v1/Configuration/Webhooks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchConfigurationWebhook(context.Background())
				return err
			},
		},
		{
			name: "UpdateConfigurationWebhook", response: cfgHookPayload,
			method: "POST", path: "/v1/Configuration/Webhooks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateConfigurationWebhook(context.Background(), voiceml.UpdateConfigurationWebhookRequest{
					Method:  voiceml.String("POST"),
					Filters: []string{"onMessageAdded", "onConversationAdded"},
					Target:  voiceml.String("webhook"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Method") != "POST" || body.Get("Target") != "webhook" {
					t.Fatalf("body: %+v", body)
				}
				filters := body["Filters"]
				if len(filters) != 2 || filters[0] != "onMessageAdded" {
					t.Fatalf("filters: %+v", filters)
				}
			},
		},
		// --- ConfigAddress --------------------------------------------------
		{
			name: "CreateConfigAddress", response: addrPayload,
			method: "POST", path: "/v1/Configuration/Addresses",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateConfigAddress(context.Background(), voiceml.CreateConfigAddressRequest{
					Type: "sms", Address: "+15551234567",
					AutoCreationEnabled:    voiceml.Bool(true),
					AutoCreationType:       voiceml.String("webhook"),
					AutoCreationWebhookURL: voiceml.String("https://example.com/auto"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Type") != "sms" || body.Get("Address") != "+15551234567" ||
					body.Get("AutoCreation.Enabled") != "true" ||
					body.Get("AutoCreation.Type") != "webhook" ||
					body.Get("AutoCreation.WebhookUrl") != "https://example.com/auto" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListConfigAddresses", listShape: map[string]any{"addresses": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Configuration/Addresses",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListConfigAddresses(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "UpdateConfigAddress", response: addrPayload,
			method: "POST", path: "/v1/Configuration/Addresses/" + convAddrSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateConfigAddress(context.Background(), convAddrSid, voiceml.UpdateConfigAddressRequest{
					AutoCreationEnabled: voiceml.Bool(false),
				})
				return err
			},
		},
		{
			name:   "DeleteConfigAddress",
			method: "DELETE", path: "/v1/Configuration/Addresses/" + convAddrSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteConfigAddress(context.Background(), convAddrSid)
			},
		},
		// --- ParticipantConversations --------------------------------------
		{
			name: "ListParticipantConversationsByIdentity", listShape: map[string]any{"conversations": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/ParticipantConversations",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListParticipantConversations(context.Background(), voiceml.ListParticipantConversationsParams{
					Identity: voiceml.String("alice"),
				})
				return err
			},
			assertBody: func(t *testing.T, _, q url.Values) {
				if q.Get("Identity") != "alice" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name: "ListParticipantConversationsByAddress", listShape: map[string]any{"conversations": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/ParticipantConversations",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListParticipantConversations(context.Background(), voiceml.ListParticipantConversationsParams{
					Address:  voiceml.String("+15551234567"),
					PageSize: voiceml.Int(50),
				})
				return err
			},
			assertBody: func(t *testing.T, _, q url.Values) {
				if q.Get("Address") != "+15551234567" || q.Get("PageSize") != "50" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		// --- ConversationWithParticipants ----------------------------------
		{
			name: "CreateConversationWithParticipants", response: convPayload,
			method: "POST", path: "/v1/ConversationWithParticipants",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateConversationWithParticipants(context.Background(), voiceml.CreateConversationWithParticipantsRequest{
					FriendlyName: voiceml.String("group chat"),
					Participant: []string{
						`{"identity":"alice"}`,
						`{"messaging_binding":{"address":"+15551234567"}}`,
					},
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "group chat" {
					t.Fatalf("body: %+v", body)
				}
				parts := body["Participant"]
				if len(parts) != 2 || parts[0] != `{"identity":"alice"}` {
					t.Fatalf("repeated Participant: %+v", parts)
				}
			},
		},
		// --- Services -------------------------------------------------------
		{
			name: "CreateService", response: svcPayload,
			method: "POST", path: "/v1/Services",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateService(context.Background(), voiceml.CreateServiceRequest{
					FriendlyName: "main",
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "main" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListServices", listShape: map[string]any{"services": []any{}, "meta": map[string]any{}},
			method: "GET", path: "/v1/Services",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServices(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchService", response: svcPayload,
			method: "GET", path: "/v1/Services/" + convChatSvc,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchService(context.Background(), convChatSvc)
				return err
			},
		},
		{
			name:   "DeleteService",
			method: "DELETE", path: "/v1/Services/" + convChatSvc,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteService(context.Background(), convChatSvc)
			},
		},
		// --- Use pcPayload here so the linter doesn't flag it as unused. ----
		{
			name:     "FetchSingleParticipantConversation_shapeOnly",
			response: pcPayload, listShape: map[string]any{"conversations": []map[string]any{pcPayload}, "meta": map[string]any{}},
			method: "GET", path: "/v1/ParticipantConversations",
			invoke: func(c *voiceml.Client) error {
				out, err := c.ConversationsV1.ListParticipantConversations(context.Background(), voiceml.ListParticipantConversationsParams{
					Identity: voiceml.String("bob"),
				})
				if err != nil {
					return err
				}
				if len(out.Conversations) != 1 {
					t.Fatalf("decoded len: %d", len(out.Conversations))
				}
				if out.Conversations[0].ConversationState != "active" {
					t.Fatalf("conv state: %q", out.Conversations[0].ConversationState)
				}
				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var responseBody any
			switch {
			case tc.listShape != nil:
				responseBody = tc.listShape
			case tc.response != nil:
				responseBody = tc.response
			default:
				// Delete endpoints — 204 with no body.
				responseBody = nil
			}
			steps := []handlerStep{jsonStep(200, responseBody)}
			c, rec, done := newClient(t, steps, nil)
			defer done()
			if err := tc.invoke(c); err != nil {
				t.Fatalf("invoke: %v", err)
			}
			r := rec.requests[0]
			if r.Method != tc.method {
				t.Fatalf("method: %s (want %s)", r.Method, tc.method)
			}
			if r.Path != tc.path {
				t.Fatalf("path: %s (want %s)", r.Path, tc.path)
			}
			if strings.Contains(r.Path, testAccountSid) {
				t.Fatalf("path leaked account sid: %q", r.Path)
			}
			if tc.assertBody != nil {
				formBody, _ := url.ParseQuery(string(r.Body))
				query, _ := url.ParseQuery(r.Query)
				tc.assertBody(t, formBody, query)
			}
		})
	}
}

// Verify the ConversationsV1ChatService model (resource named to avoid
// clashing with the ConversationsV1Service methods struct) round-trips.
func TestConversationsV1ChatServiceDecode(t *testing.T) {
	steps := []handlerStep{jsonStep(200, map[string]any{
		"sid": convChatSvc, "account_sid": testAccountSid,
		"friendly_name": "main",
		"date_created":  "2026-06-27T12:00:00Z",
		"date_updated":  "2026-06-27T12:00:00Z",
		"url":           "https://example/v1/Services/" + convChatSvc,
		"links":         map[string]string{"conversations": "https://example/v1/Services/" + convChatSvc + "/Conversations"},
	})}
	c, _, done := newClient(t, steps, nil)
	defer done()
	out, err := c.ConversationsV1.FetchService(context.Background(), convChatSvc)
	if err != nil {
		t.Fatalf("FetchService: %v", err)
	}
	if out.Sid == nil || *out.Sid != convChatSvc {
		t.Fatalf("sid: %+v", out.Sid)
	}
	if out.FriendlyName == nil || *out.FriendlyName != "main" {
		t.Fatalf("friendly_name: %+v", out.FriendlyName)
	}
	if got := out.Links["conversations"]; !strings.HasSuffix(got, "/Conversations") {
		t.Fatalf("links: %+v", out.Links)
	}

	// Asserting the response model type without invoking is also valuable —
	// the type rename is the load-bearing detail.
	var _ *voiceml.ConversationsV1ChatService = out
}
