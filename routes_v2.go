package voiceml

import (
	"context"
	"net/url"
)

// RoutesV2SipDomain is Twilio's routes/v2 Inbound Processing Region binding
// for a SIP domain. Keyed by domain name (not the SipDomain SID); the
// account is resolved from HTTP Basic auth.
type RoutesV2SipDomain struct {
	Sid          string  `json:"sid"`
	SipDomain    string  `json:"sip_domain"`
	AccountSid   string  `json:"account_sid"`
	FriendlyName *string `json:"friendly_name"`
	VoiceRegion  *string `json:"voice_region"`
	URL          *string `json:"url"`
	DateCreated  string  `json:"date_created"`
	DateUpdated  string  `json:"date_updated"`
}

// UpdateRoutesV2SipDomainParams is the body for POST /v2/SipDomains/{SipDomain}.
// All fields optional.
type UpdateRoutesV2SipDomainParams struct {
	VoiceRegion  *string `form:"VoiceRegion"`
	FriendlyName *string `form:"FriendlyName"`
}

func (p UpdateRoutesV2SipDomainParams) form() url.Values {
	v := url.Values{}
	setStr(v, "VoiceRegion", p.VoiceRegion)
	setStr(v, "FriendlyName", p.FriendlyName)
	return v
}

// RoutesV2Service bundles the routes/v2 sub-services. Reach it as c.RoutesV2.
type RoutesV2Service struct {
	SipDomains *RoutesV2SipDomainsService
}

// RoutesV2SipDomainsService surfaces GET/POST /v2/SipDomains/{SipDomain}.
type RoutesV2SipDomainsService struct{ c *Client }

// Fetch retrieves a domain's Inbound Processing Region binding.
func (s *RoutesV2SipDomainsService) Fetch(ctx context.Context, domainName string) (*RoutesV2SipDomain, error) {
	var out RoutesV2SipDomain
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v2/SipDomains/" + domainName,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update sets a domain's voice region and/or friendly name.
func (s *RoutesV2SipDomainsService) Update(ctx context.Context, domainName string, params UpdateRoutesV2SipDomainParams) (*RoutesV2SipDomain, error) {
	var out RoutesV2SipDomain
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v2/SipDomains/" + domainName, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
