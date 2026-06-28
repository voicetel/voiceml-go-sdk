// Voice v1 (voice.twilio.com/v1) resources: ByocTrunks, ConnectionPolicies
// + Targets, DialingPermissions Settings, SourceIpMappings, IpRecords.
//
// The /v1 namespace omits the /Accounts/{AccountSid} segment — the account is
// resolved from HTTP Basic auth. List responses carry a `meta` envelope
// (VoiceV1Meta) shared with the Conversations v1 surface.

package voiceml

import (
	"context"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Shared meta envelope for /v1 list responses (Voice v1 + Conversations v1).
// ---------------------------------------------------------------------------

// VoiceV1Meta is the pagination envelope embedded in every /v1 list response.
// Twilio's voice.twilio.com / conversations.twilio.com surfaces both ship it
// under the `meta` key.
type VoiceV1Meta struct {
	FirstPageURL    *string `json:"first_page_url,omitempty"`
	NextPageURL     *string `json:"next_page_url,omitempty"`
	PreviousPageURL *string `json:"previous_page_url,omitempty"`
	URL             *string `json:"url,omitempty"`
	Page            *int    `json:"page,omitempty"`
	PageSize        *int    `json:"page_size,omitempty"`
	Key             *string `json:"key,omitempty"`
}

// V1PageParams are the shared list-query knobs the /v1 surface accepts. The
// Twilio paginator key is `PageSize`.
type V1PageParams struct {
	PageSize *int
}

func (p V1PageParams) query() url.Values {
	v := url.Values{}
	if p.PageSize != nil {
		v.Set("PageSize", strconv.Itoa(*p.PageSize))
	}
	return v
}

// ---------------------------------------------------------------------------
// Response models.
// ---------------------------------------------------------------------------

// VoiceV1IpRecord is a standalone allowed source IP (IL...) bound to the
// authenticated account. Pair with VoiceV1SourceIpMapping to route by source.
type VoiceV1IpRecord struct {
	AccountSid       *string `json:"account_sid"`
	Sid              *string `json:"sid"`
	FriendlyName     *string `json:"friendly_name"`
	IpAddress        *string `json:"ip_address"`
	CidrPrefixLength int     `json:"cidr_prefix_length"`
	DateCreated      *string `json:"date_created"`
	DateUpdated      *string `json:"date_updated"`
	URL              *string `json:"url"`
}

// VoiceV1IpRecordList is the paginated /v1/IpRecords response.
type VoiceV1IpRecordList struct {
	IpRecords []VoiceV1IpRecord `json:"ip_records"`
	Meta      VoiceV1Meta       `json:"meta"`
}

// VoiceV1SourceIpMapping binds an IpRecord (IL...) to a SIP Domain (SD...).
// Sid is IB...
type VoiceV1SourceIpMapping struct {
	Sid          *string `json:"sid"`
	IpRecordSid  *string `json:"ip_record_sid"`
	SipDomainSid *string `json:"sip_domain_sid"`
	DateCreated  *string `json:"date_created"`
	DateUpdated  *string `json:"date_updated"`
	URL          *string `json:"url"`
}

// VoiceV1SourceIpMappingList is the paginated /v1/SourceIpMappings response.
type VoiceV1SourceIpMappingList struct {
	SourceIpMappings []VoiceV1SourceIpMapping `json:"source_ip_mappings"`
	Meta             VoiceV1Meta              `json:"meta"`
}

// VoiceV1ByocTrunk is a bring-your-own-carrier trunk (BY...). All callback
// URLs are optional and round-trip nullably.
type VoiceV1ByocTrunk struct {
	AccountSid           *string `json:"account_sid"`
	Sid                  *string `json:"sid"`
	FriendlyName         *string `json:"friendly_name"`
	VoiceURL             *string `json:"voice_url"`
	VoiceMethod          *string `json:"voice_method"`
	VoiceFallbackURL     *string `json:"voice_fallback_url"`
	VoiceFallbackMethod  *string `json:"voice_fallback_method"`
	StatusCallbackURL    *string `json:"status_callback_url"`
	StatusCallbackMethod *string `json:"status_callback_method"`
	CnamLookupEnabled    *bool   `json:"cnam_lookup_enabled"`
	ConnectionPolicySid  *string `json:"connection_policy_sid"`
	FromDomainSid        *string `json:"from_domain_sid"`
	DateCreated          *string `json:"date_created"`
	DateUpdated          *string `json:"date_updated"`
	URL                  *string `json:"url"`
}

// VoiceV1ByocTrunkList is the paginated /v1/ByocTrunks response.
type VoiceV1ByocTrunkList struct {
	ByocTrunks []VoiceV1ByocTrunk `json:"byoc_trunks"`
	Meta       VoiceV1Meta        `json:"meta"`
}

// VoiceV1ConnectionPolicy is a named bag of SIP-URI Targets (NY...).
type VoiceV1ConnectionPolicy struct {
	AccountSid   *string           `json:"account_sid"`
	Sid          *string           `json:"sid"`
	FriendlyName *string           `json:"friendly_name"`
	DateCreated  *string           `json:"date_created"`
	DateUpdated  *string           `json:"date_updated"`
	URL          *string           `json:"url"`
	Links        map[string]string `json:"links,omitempty"`
}

// VoiceV1ConnectionPolicyList is the paginated /v1/ConnectionPolicies response.
type VoiceV1ConnectionPolicyList struct {
	ConnectionPolicies []VoiceV1ConnectionPolicy `json:"connection_policies"`
	Meta               VoiceV1Meta               `json:"meta"`
}

// VoiceV1ConnectionPolicyTarget is one SIP URI target inside a
// ConnectionPolicy (NE...). Priority is lower-is-higher; Weight is the
// load-balancing weight among equal priorities.
type VoiceV1ConnectionPolicyTarget struct {
	AccountSid          *string `json:"account_sid"`
	ConnectionPolicySid *string `json:"connection_policy_sid"`
	Sid                 *string `json:"sid"`
	FriendlyName        *string `json:"friendly_name"`
	Target              *string `json:"target"`
	Priority            int     `json:"priority"`
	Weight              int     `json:"weight"`
	Enabled             *bool   `json:"enabled"`
	DateCreated         *string `json:"date_created"`
	DateUpdated         *string `json:"date_updated"`
	URL                 *string `json:"url"`
}

// VoiceV1ConnectionPolicyTargetList is the paginated nested-target response.
type VoiceV1ConnectionPolicyTargetList struct {
	Targets []VoiceV1ConnectionPolicyTarget `json:"targets"`
	Meta    VoiceV1Meta                     `json:"meta"`
}

// VoiceV1DialingPermissionsSettings is the singleton DialingPermissions
// /v1/Settings resource — fetch + update only, no list.
type VoiceV1DialingPermissionsSettings struct {
	DialingPermissionsInheritance *bool   `json:"dialing_permissions_inheritance"`
	URL                           *string `json:"url"`
}

// ---------------------------------------------------------------------------
// Request params.
// ---------------------------------------------------------------------------

// CreateVoiceV1IpRecordParams is the body for POST /v1/IpRecords. IpAddress is required.
type CreateVoiceV1IpRecordParams struct {
	IpAddress        string  `form:"IpAddress"`
	FriendlyName     *string `form:"FriendlyName"`
	CidrPrefixLength *int    `form:"CidrPrefixLength"`
}

func (p CreateVoiceV1IpRecordParams) form() url.Values {
	v := url.Values{}
	v.Set("IpAddress", p.IpAddress)
	setStr(v, "FriendlyName", p.FriendlyName)
	setInt(v, "CidrPrefixLength", p.CidrPrefixLength)
	return v
}

// UpdateVoiceV1IpRecordParams is the body for POST /v1/IpRecords/{Sid}.
// Only FriendlyName is mutable.
type UpdateVoiceV1IpRecordParams struct {
	FriendlyName *string `form:"FriendlyName"`
}

func (p UpdateVoiceV1IpRecordParams) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	return v
}

// CreateVoiceV1SourceIpMappingParams is the body for POST /v1/SourceIpMappings.
// Both fields required.
type CreateVoiceV1SourceIpMappingParams struct {
	IpRecordSid  string `form:"IpRecordSid"`
	SipDomainSid string `form:"SipDomainSid"`
}

func (p CreateVoiceV1SourceIpMappingParams) form() url.Values {
	v := url.Values{}
	v.Set("IpRecordSid", p.IpRecordSid)
	v.Set("SipDomainSid", p.SipDomainSid)
	return v
}

// UpdateVoiceV1SourceIpMappingParams is the body for POST /v1/SourceIpMappings/{Sid}.
// Only SipDomainSid is mutable.
type UpdateVoiceV1SourceIpMappingParams struct {
	SipDomainSid string `form:"SipDomainSid"`
}

func (p UpdateVoiceV1SourceIpMappingParams) form() url.Values {
	v := url.Values{}
	v.Set("SipDomainSid", p.SipDomainSid)
	return v
}

// CreateVoiceV1ByocTrunkParams is the body for POST /v1/ByocTrunks. All
// fields optional.
type CreateVoiceV1ByocTrunkParams struct {
	FriendlyName         *string `form:"FriendlyName"`
	VoiceURL             *string `form:"VoiceUrl"`
	VoiceMethod          *string `form:"VoiceMethod"`
	VoiceFallbackURL     *string `form:"VoiceFallbackUrl"`
	VoiceFallbackMethod  *string `form:"VoiceFallbackMethod"`
	StatusCallbackURL    *string `form:"StatusCallbackUrl"`
	StatusCallbackMethod *string `form:"StatusCallbackMethod"`
	CnamLookupEnabled    *bool   `form:"CnamLookupEnabled"`
	ConnectionPolicySid  *string `form:"ConnectionPolicySid"`
	FromDomainSid        *string `form:"FromDomainSid"`
}

func (p CreateVoiceV1ByocTrunkParams) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "VoiceUrl", p.VoiceURL)
	setStr(v, "VoiceMethod", p.VoiceMethod)
	setStr(v, "VoiceFallbackUrl", p.VoiceFallbackURL)
	setStr(v, "VoiceFallbackMethod", p.VoiceFallbackMethod)
	setStr(v, "StatusCallbackUrl", p.StatusCallbackURL)
	setStr(v, "StatusCallbackMethod", p.StatusCallbackMethod)
	setBool(v, "CnamLookupEnabled", p.CnamLookupEnabled)
	setStr(v, "ConnectionPolicySid", p.ConnectionPolicySid)
	setStr(v, "FromDomainSid", p.FromDomainSid)
	return v
}

// UpdateVoiceV1ByocTrunkParams is the body for POST /v1/ByocTrunks/{Sid}.
type UpdateVoiceV1ByocTrunkParams = CreateVoiceV1ByocTrunkParams

// CreateVoiceV1ConnectionPolicyParams is the body for POST /v1/ConnectionPolicies.
type CreateVoiceV1ConnectionPolicyParams struct {
	FriendlyName *string `form:"FriendlyName"`
}

func (p CreateVoiceV1ConnectionPolicyParams) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	return v
}

// UpdateVoiceV1ConnectionPolicyParams is the body for POST /v1/ConnectionPolicies/{Sid}.
type UpdateVoiceV1ConnectionPolicyParams = CreateVoiceV1ConnectionPolicyParams

// CreateVoiceV1ConnectionPolicyTargetParams is the body for POST
// /v1/ConnectionPolicies/{ConnectionPolicySid}/Targets. Target is required.
type CreateVoiceV1ConnectionPolicyTargetParams struct {
	Target       string  `form:"Target"`
	FriendlyName *string `form:"FriendlyName"`
	Priority     *int    `form:"Priority"`
	Weight       *int    `form:"Weight"`
	Enabled      *bool   `form:"Enabled"`
}

func (p CreateVoiceV1ConnectionPolicyTargetParams) form() url.Values {
	v := url.Values{}
	v.Set("Target", p.Target)
	setStr(v, "FriendlyName", p.FriendlyName)
	setInt(v, "Priority", p.Priority)
	setInt(v, "Weight", p.Weight)
	setBool(v, "Enabled", p.Enabled)
	return v
}

// UpdateVoiceV1ConnectionPolicyTargetParams is the body for POST
// /v1/ConnectionPolicies/{ConnectionPolicySid}/Targets/{Sid}.
type UpdateVoiceV1ConnectionPolicyTargetParams struct {
	Target       *string `form:"Target"`
	FriendlyName *string `form:"FriendlyName"`
	Priority     *int    `form:"Priority"`
	Weight       *int    `form:"Weight"`
	Enabled      *bool   `form:"Enabled"`
}

func (p UpdateVoiceV1ConnectionPolicyTargetParams) form() url.Values {
	v := url.Values{}
	setStr(v, "Target", p.Target)
	setStr(v, "FriendlyName", p.FriendlyName)
	setInt(v, "Priority", p.Priority)
	setInt(v, "Weight", p.Weight)
	setBool(v, "Enabled", p.Enabled)
	return v
}

// UpdateVoiceV1SettingsParams is the body for POST /v1/Settings.
type UpdateVoiceV1SettingsParams struct {
	DialingPermissionsInheritance *bool `form:"DialingPermissionsInheritance"`
}

func (p UpdateVoiceV1SettingsParams) form() url.Values {
	v := url.Values{}
	setBool(v, "DialingPermissionsInheritance", p.DialingPermissionsInheritance)
	return v
}

// ---------------------------------------------------------------------------
// Service — flat-method facade for the entire Voice v1 surface.
// ---------------------------------------------------------------------------

// VoiceV1Service exposes the voice.twilio.com/v1 endpoints. Reach it as
// c.VoiceV1. Methods are named verb+resource (Twilio-style) so the surface
// stays flat for IDE discovery.
type VoiceV1Service struct{ c *Client }

// --- IpRecords -------------------------------------------------------------

// CreateIpRecord adds a Voice v1 IpRecord.
func (s *VoiceV1Service) CreateIpRecord(ctx context.Context, params CreateVoiceV1IpRecordParams) (*VoiceV1IpRecord, error) {
	var out VoiceV1IpRecord
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/IpRecords", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListIpRecords returns a single page of the account's IpRecords.
func (s *VoiceV1Service) ListIpRecords(ctx context.Context, params V1PageParams) (*VoiceV1IpRecordList, error) {
	var out VoiceV1IpRecordList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/IpRecords", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchIpRecord retrieves an IpRecord by sid.
func (s *VoiceV1Service) FetchIpRecord(ctx context.Context, sid string) (*VoiceV1IpRecord, error) {
	var out VoiceV1IpRecord
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/IpRecords/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateIpRecord updates an IpRecord's friendly name.
func (s *VoiceV1Service) UpdateIpRecord(ctx context.Context, sid string, params UpdateVoiceV1IpRecordParams) (*VoiceV1IpRecord, error) {
	var out VoiceV1IpRecord
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/IpRecords/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteIpRecord removes an IpRecord.
func (s *VoiceV1Service) DeleteIpRecord(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/IpRecords/" + sid,
	}, nil)
}

// --- SourceIpMappings ------------------------------------------------------

// CreateSourceIpMapping binds an IpRecord to a SIP Domain.
func (s *VoiceV1Service) CreateSourceIpMapping(ctx context.Context, params CreateVoiceV1SourceIpMappingParams) (*VoiceV1SourceIpMapping, error) {
	var out VoiceV1SourceIpMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/SourceIpMappings", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSourceIpMappings returns a single page of SourceIpMappings.
func (s *VoiceV1Service) ListSourceIpMappings(ctx context.Context, params V1PageParams) (*VoiceV1SourceIpMappingList, error) {
	var out VoiceV1SourceIpMappingList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/SourceIpMappings", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchSourceIpMapping retrieves a SourceIpMapping by sid.
func (s *VoiceV1Service) FetchSourceIpMapping(ctx context.Context, sid string) (*VoiceV1SourceIpMapping, error) {
	var out VoiceV1SourceIpMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/SourceIpMappings/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSourceIpMapping re-points a SourceIpMapping at a different SIP Domain.
func (s *VoiceV1Service) UpdateSourceIpMapping(ctx context.Context, sid string, params UpdateVoiceV1SourceIpMappingParams) (*VoiceV1SourceIpMapping, error) {
	var out VoiceV1SourceIpMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/SourceIpMappings/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSourceIpMapping removes a SourceIpMapping.
func (s *VoiceV1Service) DeleteSourceIpMapping(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/SourceIpMappings/" + sid,
	}, nil)
}

// --- ByocTrunks ------------------------------------------------------------

// CreateByocTrunk adds a bring-your-own-carrier trunk.
func (s *VoiceV1Service) CreateByocTrunk(ctx context.Context, params CreateVoiceV1ByocTrunkParams) (*VoiceV1ByocTrunk, error) {
	var out VoiceV1ByocTrunk
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/ByocTrunks", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListByocTrunks returns a single page of ByocTrunks.
func (s *VoiceV1Service) ListByocTrunks(ctx context.Context, params V1PageParams) (*VoiceV1ByocTrunkList, error) {
	var out VoiceV1ByocTrunkList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/ByocTrunks", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchByocTrunk retrieves a ByocTrunk by sid.
func (s *VoiceV1Service) FetchByocTrunk(ctx context.Context, sid string) (*VoiceV1ByocTrunk, error) {
	var out VoiceV1ByocTrunk
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/ByocTrunks/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateByocTrunk mutates a ByocTrunk in place.
func (s *VoiceV1Service) UpdateByocTrunk(ctx context.Context, sid string, params UpdateVoiceV1ByocTrunkParams) (*VoiceV1ByocTrunk, error) {
	var out VoiceV1ByocTrunk
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/ByocTrunks/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteByocTrunk removes a ByocTrunk.
func (s *VoiceV1Service) DeleteByocTrunk(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/ByocTrunks/" + sid,
	}, nil)
}

// --- ConnectionPolicies ----------------------------------------------------

// CreateConnectionPolicy adds an empty ConnectionPolicy.
func (s *VoiceV1Service) CreateConnectionPolicy(ctx context.Context, params CreateVoiceV1ConnectionPolicyParams) (*VoiceV1ConnectionPolicy, error) {
	var out VoiceV1ConnectionPolicy
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/ConnectionPolicies", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListConnectionPolicies returns a single page of ConnectionPolicies.
func (s *VoiceV1Service) ListConnectionPolicies(ctx context.Context, params V1PageParams) (*VoiceV1ConnectionPolicyList, error) {
	var out VoiceV1ConnectionPolicyList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/ConnectionPolicies", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchConnectionPolicy retrieves a ConnectionPolicy by sid.
func (s *VoiceV1Service) FetchConnectionPolicy(ctx context.Context, sid string) (*VoiceV1ConnectionPolicy, error) {
	var out VoiceV1ConnectionPolicy
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/ConnectionPolicies/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateConnectionPolicy mutates a ConnectionPolicy in place.
func (s *VoiceV1Service) UpdateConnectionPolicy(ctx context.Context, sid string, params UpdateVoiceV1ConnectionPolicyParams) (*VoiceV1ConnectionPolicy, error) {
	var out VoiceV1ConnectionPolicy
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/ConnectionPolicies/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteConnectionPolicy removes a ConnectionPolicy.
func (s *VoiceV1Service) DeleteConnectionPolicy(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/ConnectionPolicies/" + sid,
	}, nil)
}

// --- ConnectionPolicy Targets (nested) -------------------------------------

// CreateConnectionPolicyTarget adds a Target to a ConnectionPolicy.
func (s *VoiceV1Service) CreateConnectionPolicyTarget(ctx context.Context, connectionPolicySid string, params CreateVoiceV1ConnectionPolicyTargetParams) (*VoiceV1ConnectionPolicyTarget, error) {
	var out VoiceV1ConnectionPolicyTarget
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/ConnectionPolicies/" + connectionPolicySid + "/Targets", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListConnectionPolicyTargets returns a single page of a policy's Targets.
func (s *VoiceV1Service) ListConnectionPolicyTargets(ctx context.Context, connectionPolicySid string, params V1PageParams) (*VoiceV1ConnectionPolicyTargetList, error) {
	var out VoiceV1ConnectionPolicyTargetList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/ConnectionPolicies/" + connectionPolicySid + "/Targets", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchConnectionPolicyTarget retrieves a Target by sid.
func (s *VoiceV1Service) FetchConnectionPolicyTarget(ctx context.Context, connectionPolicySid, sid string) (*VoiceV1ConnectionPolicyTarget, error) {
	var out VoiceV1ConnectionPolicyTarget
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/ConnectionPolicies/" + connectionPolicySid + "/Targets/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateConnectionPolicyTarget mutates a Target in place.
func (s *VoiceV1Service) UpdateConnectionPolicyTarget(ctx context.Context, connectionPolicySid, sid string, params UpdateVoiceV1ConnectionPolicyTargetParams) (*VoiceV1ConnectionPolicyTarget, error) {
	var out VoiceV1ConnectionPolicyTarget
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/ConnectionPolicies/" + connectionPolicySid + "/Targets/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteConnectionPolicyTarget removes a Target.
func (s *VoiceV1Service) DeleteConnectionPolicyTarget(ctx context.Context, connectionPolicySid, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/ConnectionPolicies/" + connectionPolicySid + "/Targets/" + sid,
	}, nil)
}

// --- DialingPermissions Settings (singleton) -------------------------------

// FetchSettings retrieves the account's DialingPermissions inheritance setting.
func (s *VoiceV1Service) FetchSettings(ctx context.Context) (*VoiceV1DialingPermissionsSettings, error) {
	var out VoiceV1DialingPermissionsSettings
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Settings",
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSettings sets the account's DialingPermissions inheritance setting.
func (s *VoiceV1Service) UpdateSettings(ctx context.Context, params UpdateVoiceV1SettingsParams) (*VoiceV1DialingPermissionsSettings, error) {
	var out VoiceV1DialingPermissionsSettings
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Settings", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
