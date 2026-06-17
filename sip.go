package voiceml

import (
	"context"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Response models — the wire shapes the server returns.
// ---------------------------------------------------------------------------

// SIPDomain is a SIP ingress endpoint — Twilio-compatible "SD…" resource.
// Bind a CredentialList and/or IpAccessControlList via the mapping
// sub-resources to authenticate inbound SIP traffic.
type SIPDomain struct {
	Sid                       string            `json:"sid"`
	AccountSid                string            `json:"account_sid"`
	DomainName                string            `json:"domain_name"`
	APIVersion                string            `json:"api_version"`
	FriendlyName              *string           `json:"friendly_name"`
	AuthType                  *string           `json:"auth_type"`
	VoiceURL                  *string           `json:"voice_url"`
	VoiceMethod               *string           `json:"voice_method"`
	VoiceFallbackURL          *string           `json:"voice_fallback_url"`
	VoiceFallbackMethod       *string           `json:"voice_fallback_method"`
	VoiceStatusCallbackURL    *string           `json:"voice_status_callback_url"`
	VoiceStatusCallbackMethod *string           `json:"voice_status_callback_method"`
	SipRegistration           *bool             `json:"sip_registration"`
	EmergencyCallingEnabled   *bool             `json:"emergency_calling_enabled"`
	Secure                    *bool             `json:"secure"`
	ByocTrunkSid              *string           `json:"byoc_trunk_sid"`
	EmergencyCallerSid        *string           `json:"emergency_caller_sid"`
	DateCreated               string            `json:"date_created"`
	DateUpdated               string            `json:"date_updated"`
	URI                       string            `json:"uri"`
	SubresourceURIs           map[string]string `json:"subresource_uris,omitempty"`
}

// SIPDomainList is the paginated /SIP/Domains response.
type SIPDomainList struct {
	Page
	Domains []SIPDomain `json:"domains"`
}

// SIPCredentialList is a named bag of SIP-digest credentials — "CL…".
type SIPCredentialList struct {
	Sid             string            `json:"sid"`
	AccountSid      string            `json:"account_sid"`
	FriendlyName    *string           `json:"friendly_name"`
	DateCreated     string            `json:"date_created"`
	DateUpdated     string            `json:"date_updated"`
	URI             string            `json:"uri"`
	SubresourceURIs map[string]string `json:"subresource_uris,omitempty"`
}

// SIPCredentialListList is the paginated /SIP/CredentialLists response.
type SIPCredentialListList struct {
	Page
	CredentialLists []SIPCredentialList `json:"credential_lists"`
}

// SIPCredential is a single SIP-digest username + (write-only) password — "CR…".
// Password is never round-tripped on response; use Update with a new password to
// rotate.
type SIPCredential struct {
	Sid               string `json:"sid"`
	AccountSid        string `json:"account_sid"`
	CredentialListSid string `json:"credential_list_sid"`
	Username          string `json:"username"`
	DateCreated       string `json:"date_created"`
	DateUpdated       string `json:"date_updated"`
	URI               string `json:"uri"`
}

// SIPCredentialPage is the paginated /Credentials response (named
// SipCredentialListPage in the spec — it's a page of credentials within a
// CredentialList, not a page of credential-lists, matching Twilio).
type SIPCredentialPage struct {
	Page
	Credentials []SIPCredential `json:"credentials"`
}

// SIPIpAccessControlList is a named bag of CIDR-bound IPs — "AL…".
type SIPIpAccessControlList struct {
	Sid             string            `json:"sid"`
	AccountSid      string            `json:"account_sid"`
	FriendlyName    *string           `json:"friendly_name"`
	DateCreated     string            `json:"date_created"`
	DateUpdated     string            `json:"date_updated"`
	URI             string            `json:"uri"`
	SubresourceURIs map[string]string `json:"subresource_uris,omitempty"`
}

// SIPIpAccessControlListList is the paginated /SIP/IpAccessControlLists
// response.
type SIPIpAccessControlListList struct {
	Page
	IpAccessControlLists []SIPIpAccessControlList `json:"ip_access_control_lists"`
}

// SIPIpAddress is a single CIDR-bound entry — "IP…".
type SIPIpAddress struct {
	Sid                    string `json:"sid"`
	AccountSid             string `json:"account_sid"`
	IpAccessControlListSid string `json:"ip_access_control_list_sid"`
	FriendlyName           string `json:"friendly_name"`
	IpAddress              string `json:"ip_address"`
	CidrPrefixLength       int    `json:"cidr_prefix_length"`
	DateCreated            string `json:"date_created"`
	DateUpdated            string `json:"date_updated"`
	URI                    string `json:"uri"`
}

// SIPIpAddressList is the paginated /IpAddresses response.
type SIPIpAddressList struct {
	Page
	IpAddresses []SIPIpAddress `json:"ip_addresses"`
}

// SIPDomainMapping is the round-trip shape for every domain mapping
// sub-resource (Calls / Registrations × CredentialList / IpAccessControlList).
// Sid echoes the sid of the bound resource ("CL…" for credential mappings,
// "AL…" for IP-ACL mappings); DomainSid records which domain the binding is
// attached to.
type SIPDomainMapping struct {
	Sid          string  `json:"sid"`
	AccountSid   string  `json:"account_sid"`
	FriendlyName *string `json:"friendly_name"`
	DomainSid    *string `json:"domain_sid"`
	DateCreated  string  `json:"date_created"`
	DateUpdated  string  `json:"date_updated"`
	URI          string  `json:"uri"`
}

// SIPCredentialListMappingList is the paginated mapping-list response for
// the CredentialList mapping endpoints (historical + Auth/Calls +
// Auth/Registrations namespaces).
type SIPCredentialListMappingList struct {
	Page
	CredentialListMappings []SIPDomainMapping `json:"credential_list_mappings"`
}

// SIPIpAccessControlListMappingList is the paginated mapping-list response
// for the IpAccessControlList mapping endpoints (historical + Auth/Calls;
// no registrations counterpart).
type SIPIpAccessControlListMappingList struct {
	Page
	IpAccessControlListMappings []SIPDomainMapping `json:"ip_access_control_list_mappings"`
}

// ---------------------------------------------------------------------------
// Request params — form-encoded bodies the SDK sends.
// ---------------------------------------------------------------------------

// CreateSIPDomainParams is the body for POST /SIP/Domains.json. DomainName is required.
type CreateSIPDomainParams struct {
	DomainName                string  `form:"DomainName"`
	FriendlyName              *string `form:"FriendlyName"`
	VoiceURL                  *string `form:"VoiceUrl"`
	VoiceMethod               *string `form:"VoiceMethod"`
	VoiceFallbackURL          *string `form:"VoiceFallbackUrl"`
	VoiceFallbackMethod       *string `form:"VoiceFallbackMethod"`
	VoiceStatusCallbackURL    *string `form:"VoiceStatusCallbackUrl"`
	VoiceStatusCallbackMethod *string `form:"VoiceStatusCallbackMethod"`
	SipRegistration           *bool   `form:"SipRegistration"`
	Secure                    *bool   `form:"Secure"`
	EmergencyCallingEnabled   *bool   `form:"EmergencyCallingEnabled"`
	ByocTrunkSid              *string `form:"ByocTrunkSid"`
	EmergencyCallerSid        *string `form:"EmergencyCallerSid"`
}

func setStr(v url.Values, key string, value *string) {
	if value != nil {
		v.Set(key, *value)
	}
}

func setBool(v url.Values, key string, value *bool) {
	if value != nil {
		if *value {
			v.Set(key, "true")
		} else {
			v.Set(key, "false")
		}
	}
}

func setInt(v url.Values, key string, value *int) {
	if value != nil {
		v.Set(key, strconv.Itoa(*value))
	}
}

func (p CreateSIPDomainParams) form() url.Values {
	v := url.Values{}
	v.Set("DomainName", p.DomainName)
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "VoiceUrl", p.VoiceURL)
	setStr(v, "VoiceMethod", p.VoiceMethod)
	setStr(v, "VoiceFallbackUrl", p.VoiceFallbackURL)
	setStr(v, "VoiceFallbackMethod", p.VoiceFallbackMethod)
	setStr(v, "VoiceStatusCallbackUrl", p.VoiceStatusCallbackURL)
	setStr(v, "VoiceStatusCallbackMethod", p.VoiceStatusCallbackMethod)
	setBool(v, "SipRegistration", p.SipRegistration)
	setBool(v, "Secure", p.Secure)
	setBool(v, "EmergencyCallingEnabled", p.EmergencyCallingEnabled)
	setStr(v, "ByocTrunkSid", p.ByocTrunkSid)
	setStr(v, "EmergencyCallerSid", p.EmergencyCallerSid)
	return v
}

// UpdateSIPDomainParams is the body for POST /SIP/Domains/{Sid}.json. All
// fields optional.
type UpdateSIPDomainParams struct {
	FriendlyName              *string `form:"FriendlyName"`
	VoiceURL                  *string `form:"VoiceUrl"`
	VoiceMethod               *string `form:"VoiceMethod"`
	VoiceFallbackURL          *string `form:"VoiceFallbackUrl"`
	VoiceFallbackMethod       *string `form:"VoiceFallbackMethod"`
	VoiceStatusCallbackURL    *string `form:"VoiceStatusCallbackUrl"`
	VoiceStatusCallbackMethod *string `form:"VoiceStatusCallbackMethod"`
	SipRegistration           *bool   `form:"SipRegistration"`
	Secure                    *bool   `form:"Secure"`
	EmergencyCallingEnabled   *bool   `form:"EmergencyCallingEnabled"`
	ByocTrunkSid              *string `form:"ByocTrunkSid"`
	EmergencyCallerSid        *string `form:"EmergencyCallerSid"`
}

func (p UpdateSIPDomainParams) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "VoiceUrl", p.VoiceURL)
	setStr(v, "VoiceMethod", p.VoiceMethod)
	setStr(v, "VoiceFallbackUrl", p.VoiceFallbackURL)
	setStr(v, "VoiceFallbackMethod", p.VoiceFallbackMethod)
	setStr(v, "VoiceStatusCallbackUrl", p.VoiceStatusCallbackURL)
	setStr(v, "VoiceStatusCallbackMethod", p.VoiceStatusCallbackMethod)
	setBool(v, "SipRegistration", p.SipRegistration)
	setBool(v, "Secure", p.Secure)
	setBool(v, "EmergencyCallingEnabled", p.EmergencyCallingEnabled)
	setStr(v, "ByocTrunkSid", p.ByocTrunkSid)
	setStr(v, "EmergencyCallerSid", p.EmergencyCallerSid)
	return v
}

// CreateSIPCredentialListParams is the body for POST /SIP/CredentialLists.json.
type CreateSIPCredentialListParams struct {
	FriendlyName string `form:"FriendlyName"`
}

func (p CreateSIPCredentialListParams) form() url.Values {
	v := url.Values{}
	v.Set("FriendlyName", p.FriendlyName)
	return v
}

// UpdateSIPCredentialListParams is the body for POST /SIP/CredentialLists/{Sid}.json.
type UpdateSIPCredentialListParams struct {
	FriendlyName *string `form:"FriendlyName"`
}

func (p UpdateSIPCredentialListParams) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	return v
}

// CreateSIPCredentialParams is the body for POST .../Credentials.json.
type CreateSIPCredentialParams struct {
	Username string `form:"Username"`
	Password string `form:"Password"`
}

func (p CreateSIPCredentialParams) form() url.Values {
	v := url.Values{}
	v.Set("Username", p.Username)
	v.Set("Password", p.Password)
	return v
}

// UpdateSIPCredentialParams is the body for POST .../Credentials/{Sid}.json.
// Only the password is mutable.
type UpdateSIPCredentialParams struct {
	Password string `form:"Password"`
}

func (p UpdateSIPCredentialParams) form() url.Values {
	v := url.Values{}
	v.Set("Password", p.Password)
	return v
}

// CreateSIPIpAccessControlListParams is the body for POST /SIP/IpAccessControlLists.json.
type CreateSIPIpAccessControlListParams struct {
	FriendlyName string `form:"FriendlyName"`
}

func (p CreateSIPIpAccessControlListParams) form() url.Values {
	v := url.Values{}
	v.Set("FriendlyName", p.FriendlyName)
	return v
}

// UpdateSIPIpAccessControlListParams is the body for POST .../IpAccessControlLists/{Sid}.json.
type UpdateSIPIpAccessControlListParams struct {
	FriendlyName *string `form:"FriendlyName"`
}

func (p UpdateSIPIpAccessControlListParams) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	return v
}

// CreateSIPIpAddressParams is the body for POST .../IpAddresses.json.
// CidrPrefixLength defaults server-side to 32 (single host) when omitted.
type CreateSIPIpAddressParams struct {
	FriendlyName     string `form:"FriendlyName"`
	IpAddress        string `form:"IpAddress"`
	CidrPrefixLength *int   `form:"CidrPrefixLength"`
}

func (p CreateSIPIpAddressParams) form() url.Values {
	v := url.Values{}
	v.Set("FriendlyName", p.FriendlyName)
	v.Set("IpAddress", p.IpAddress)
	setInt(v, "CidrPrefixLength", p.CidrPrefixLength)
	return v
}

// UpdateSIPIpAddressParams is the body for POST .../IpAddresses/{Sid}.json.
type UpdateSIPIpAddressParams struct {
	FriendlyName     *string `form:"FriendlyName"`
	IpAddress        *string `form:"IpAddress"`
	CidrPrefixLength *int    `form:"CidrPrefixLength"`
}

func (p UpdateSIPIpAddressParams) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "IpAddress", p.IpAddress)
	setInt(v, "CidrPrefixLength", p.CidrPrefixLength)
	return v
}

// CreateSIPCredentialListMappingParams is the body for any
// .../CredentialListMappings POST (historical / Auth/Calls / Auth/Registrations).
type CreateSIPCredentialListMappingParams struct {
	CredentialListSid string `form:"CredentialListSid"`
}

func (p CreateSIPCredentialListMappingParams) form() url.Values {
	v := url.Values{}
	v.Set("CredentialListSid", p.CredentialListSid)
	return v
}

// CreateSIPIpAccessControlListMappingParams is the body for any
// .../IpAccessControlListMappings POST (historical / Auth/Calls).
type CreateSIPIpAccessControlListMappingParams struct {
	IpAccessControlListSid string `form:"IpAccessControlListSid"`
}

func (p CreateSIPIpAccessControlListMappingParams) form() url.Values {
	v := url.Values{}
	v.Set("IpAccessControlListSid", p.IpAccessControlListSid)
	return v
}

// ---------------------------------------------------------------------------
// Services — three top-level sub-services and helpers per resource family.
// ---------------------------------------------------------------------------

// SIPService bundles the SIP Trunking sub-services. Reach it as c.SIP.
type SIPService struct {
	Domains              *SIPDomainsService
	CredentialLists      *SIPCredentialListsService
	IpAccessControlLists *SIPIpAccessControlListsService
}

// SIPDomainsService surfaces /SIP/Domains plus the four mapping endpoints
// attached to a SipDomain (historical aliases + Auth/Calls + Auth/Registrations).
type SIPDomainsService struct{ c *Client }

// SIPCredentialListsService surfaces /SIP/CredentialLists plus the per-list
// /Credentials sub-resource.
type SIPCredentialListsService struct{ c *Client }

// SIPIpAccessControlListsService surfaces /SIP/IpAccessControlLists plus the
// per-list /IpAddresses sub-resource.
type SIPIpAccessControlListsService struct{ c *Client }

// ---------------------------------------------------------------------------
// SIPDomainsService methods
// ---------------------------------------------------------------------------

// List returns a single page of SipDomains.
func (s *SIPDomainsService) List(ctx context.Context, params ListPageParams) (*SIPDomainList, error) {
	var out SIPDomainList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains"), query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create adds a SipDomain.
func (s *SIPDomainsService) Create(ctx context.Context, params CreateSIPDomainParams) (*SIPDomain, error) {
	var out SIPDomain
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "Domains"), form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch retrieves a SipDomain by sid.
func (s *SIPDomainsService) Fetch(ctx context.Context, domainSid string) (*SIPDomain, error) {
	var out SIPDomain
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update mutates a SipDomain in place.
func (s *SIPDomainsService) Update(ctx context.Context, domainSid string, params UpdateSIPDomainParams) (*SIPDomain, error) {
	var out SIPDomain
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "Domains", domainSid), form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a SipDomain.
func (s *SIPDomainsService) Delete(ctx context.Context, domainSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "Domains", domainSid),
	}, nil)
}

// --- Historical CredentialList mappings ------------------------------------

func (s *SIPDomainsService) ListCredentialListMappings(ctx context.Context, domainSid string, params ListPageParams) (*SIPCredentialListMappingList, error) {
	var out SIPCredentialListMappingList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "CredentialListMappings"),
		query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) CreateCredentialListMapping(ctx context.Context, domainSid string, params CreateSIPCredentialListMappingParams) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "Domains", domainSid, "CredentialListMappings"),
		form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) FetchCredentialListMapping(ctx context.Context, domainSid, mappingSid string) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "CredentialListMappings", mappingSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) DeleteCredentialListMapping(ctx context.Context, domainSid, mappingSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "Domains", domainSid, "CredentialListMappings", mappingSid),
	}, nil)
}

// --- Historical IpAccessControlList mappings -------------------------------

func (s *SIPDomainsService) ListIpAccessControlListMappings(ctx context.Context, domainSid string, params ListPageParams) (*SIPIpAccessControlListMappingList, error) {
	var out SIPIpAccessControlListMappingList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "IpAccessControlListMappings"),
		query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) CreateIpAccessControlListMapping(ctx context.Context, domainSid string, params CreateSIPIpAccessControlListMappingParams) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "Domains", domainSid, "IpAccessControlListMappings"),
		form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) FetchIpAccessControlListMapping(ctx context.Context, domainSid, mappingSid string) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "IpAccessControlListMappings", mappingSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) DeleteIpAccessControlListMapping(ctx context.Context, domainSid, mappingSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "Domains", domainSid, "IpAccessControlListMappings", mappingSid),
	}, nil)
}

// --- Auth/Calls/CredentialListMappings -------------------------------------

func (s *SIPDomainsService) ListAuthCallsCredentialListMappings(ctx context.Context, domainSid string, params ListPageParams) (*SIPCredentialListMappingList, error) {
	var out SIPCredentialListMappingList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Calls", "CredentialListMappings"),
		query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) CreateAuthCallsCredentialListMapping(ctx context.Context, domainSid string, params CreateSIPCredentialListMappingParams) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Calls", "CredentialListMappings"),
		form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) FetchAuthCallsCredentialListMapping(ctx context.Context, domainSid, mappingSid string) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Calls", "CredentialListMappings", mappingSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) DeleteAuthCallsCredentialListMapping(ctx context.Context, domainSid, mappingSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Calls", "CredentialListMappings", mappingSid),
	}, nil)
}

// --- Auth/Calls/IpAccessControlListMappings --------------------------------

func (s *SIPDomainsService) ListAuthCallsIpAccessControlListMappings(ctx context.Context, domainSid string, params ListPageParams) (*SIPIpAccessControlListMappingList, error) {
	var out SIPIpAccessControlListMappingList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Calls", "IpAccessControlListMappings"),
		query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) CreateAuthCallsIpAccessControlListMapping(ctx context.Context, domainSid string, params CreateSIPIpAccessControlListMappingParams) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Calls", "IpAccessControlListMappings"),
		form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) FetchAuthCallsIpAccessControlListMapping(ctx context.Context, domainSid, mappingSid string) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Calls", "IpAccessControlListMappings", mappingSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) DeleteAuthCallsIpAccessControlListMapping(ctx context.Context, domainSid, mappingSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Calls", "IpAccessControlListMappings", mappingSid),
	}, nil)
}

// --- Auth/Registrations/CredentialListMappings (no IP-ACL counterpart) -----

func (s *SIPDomainsService) ListAuthRegistrationsCredentialListMappings(ctx context.Context, domainSid string, params ListPageParams) (*SIPCredentialListMappingList, error) {
	var out SIPCredentialListMappingList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Registrations", "CredentialListMappings"),
		query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) CreateAuthRegistrationsCredentialListMapping(ctx context.Context, domainSid string, params CreateSIPCredentialListMappingParams) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Registrations", "CredentialListMappings"),
		form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) FetchAuthRegistrationsCredentialListMapping(ctx context.Context, domainSid, mappingSid string) (*SIPDomainMapping, error) {
	var out SIPDomainMapping
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Registrations", "CredentialListMappings", mappingSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPDomainsService) DeleteAuthRegistrationsCredentialListMapping(ctx context.Context, domainSid, mappingSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "Domains", domainSid, "Auth", "Registrations", "CredentialListMappings", mappingSid),
	}, nil)
}

// ---------------------------------------------------------------------------
// SIPCredentialListsService methods
// ---------------------------------------------------------------------------

func (s *SIPCredentialListsService) List(ctx context.Context, params ListPageParams) (*SIPCredentialListList, error) {
	var out SIPCredentialListList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "CredentialLists"), query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPCredentialListsService) Create(ctx context.Context, params CreateSIPCredentialListParams) (*SIPCredentialList, error) {
	var out SIPCredentialList
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "CredentialLists"), form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPCredentialListsService) Fetch(ctx context.Context, credentialListSid string) (*SIPCredentialList, error) {
	var out SIPCredentialList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "CredentialLists", credentialListSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPCredentialListsService) Update(ctx context.Context, credentialListSid string, params UpdateSIPCredentialListParams) (*SIPCredentialList, error) {
	var out SIPCredentialList
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "CredentialLists", credentialListSid), form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPCredentialListsService) Delete(ctx context.Context, credentialListSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "CredentialLists", credentialListSid),
	}, nil)
}

// --- Per-CredentialList /Credentials sub-resource --------------------------

func (s *SIPCredentialListsService) ListCredentials(ctx context.Context, credentialListSid string, params ListPageParams) (*SIPCredentialPage, error) {
	var out SIPCredentialPage
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "CredentialLists", credentialListSid, "Credentials"),
		query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPCredentialListsService) CreateCredential(ctx context.Context, credentialListSid string, params CreateSIPCredentialParams) (*SIPCredential, error) {
	var out SIPCredential
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "CredentialLists", credentialListSid, "Credentials"),
		form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPCredentialListsService) FetchCredential(ctx context.Context, credentialListSid, credentialSid string) (*SIPCredential, error) {
	var out SIPCredential
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "CredentialLists", credentialListSid, "Credentials", credentialSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPCredentialListsService) UpdateCredential(ctx context.Context, credentialListSid, credentialSid string, params UpdateSIPCredentialParams) (*SIPCredential, error) {
	var out SIPCredential
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "CredentialLists", credentialListSid, "Credentials", credentialSid),
		form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPCredentialListsService) DeleteCredential(ctx context.Context, credentialListSid, credentialSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "CredentialLists", credentialListSid, "Credentials", credentialSid),
	}, nil)
}

// ---------------------------------------------------------------------------
// SIPIpAccessControlListsService methods
// ---------------------------------------------------------------------------

func (s *SIPIpAccessControlListsService) List(ctx context.Context, params ListPageParams) (*SIPIpAccessControlListList, error) {
	var out SIPIpAccessControlListList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "IpAccessControlLists"), query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPIpAccessControlListsService) Create(ctx context.Context, params CreateSIPIpAccessControlListParams) (*SIPIpAccessControlList, error) {
	var out SIPIpAccessControlList
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "IpAccessControlLists"), form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPIpAccessControlListsService) Fetch(ctx context.Context, aclSid string) (*SIPIpAccessControlList, error) {
	var out SIPIpAccessControlList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "IpAccessControlLists", aclSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPIpAccessControlListsService) Update(ctx context.Context, aclSid string, params UpdateSIPIpAccessControlListParams) (*SIPIpAccessControlList, error) {
	var out SIPIpAccessControlList
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "IpAccessControlLists", aclSid), form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPIpAccessControlListsService) Delete(ctx context.Context, aclSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "IpAccessControlLists", aclSid),
	}, nil)
}

// --- Per-IpAccessControlList /IpAddresses sub-resource ---------------------

func (s *SIPIpAccessControlListsService) ListIpAddresses(ctx context.Context, aclSid string, params ListPageParams) (*SIPIpAddressList, error) {
	var out SIPIpAddressList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "IpAccessControlLists", aclSid, "IpAddresses"),
		query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPIpAccessControlListsService) CreateIpAddress(ctx context.Context, aclSid string, params CreateSIPIpAddressParams) (*SIPIpAddress, error) {
	var out SIPIpAddress
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "IpAccessControlLists", aclSid, "IpAddresses"),
		form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPIpAccessControlListsService) FetchIpAddress(ctx context.Context, aclSid, ipAddressSid string) (*SIPIpAddress, error) {
	var out SIPIpAddress
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: s.c.pathf("SIP", "IpAccessControlLists", aclSid, "IpAddresses", ipAddressSid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPIpAccessControlListsService) UpdateIpAddress(ctx context.Context, aclSid, ipAddressSid string, params UpdateSIPIpAddressParams) (*SIPIpAddress, error) {
	var out SIPIpAddress
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: s.c.pathf("SIP", "IpAccessControlLists", aclSid, "IpAddresses", ipAddressSid),
		form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SIPIpAccessControlListsService) DeleteIpAddress(ctx context.Context, aclSid, ipAddressSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: s.c.pathf("SIP", "IpAccessControlLists", aclSid, "IpAddresses", ipAddressSid),
	}, nil)
}
