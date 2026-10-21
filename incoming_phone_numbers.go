package voiceml

import (
	"context"
	"net/url"
)

// IncomingPhoneNumbersService is the /IncomingPhoneNumbers resource group —
// CRUD on the DIDs assigned to the authenticated tenant.
type IncomingPhoneNumbersService struct{ c *Client }

// IncomingPhoneNumberCapabilities is the channel-capability sub-object on an
// IncomingPhoneNumber. The server may send it as null, which decodes as all
// false.
type IncomingPhoneNumberCapabilities struct {
	Voice bool `json:"voice"`
	SMS   bool `json:"sms"`
	MMS   bool `json:"mms"`
	Fax   bool `json:"fax"`
}

// IncomingPhoneNumber is a Twilio-compatible IncomingPhoneNumber resource. `Sid`
// is the canonical PN-prefixed opaque identifier; `PhoneNumber` carries the
// E.164 form. Nullable / often-empty Twilio-compat fields are exposed as
// `*string` so callers can distinguish "server emitted empty string" from
// "field absent in this response".
type IncomingPhoneNumber struct {
	Sid          string `json:"sid"`
	AccountSid   string `json:"account_sid"`
	PhoneNumber  string `json:"phone_number"`
	FriendlyName string `json:"friendly_name,omitempty"`
	APIVersion   string `json:"api_version"`
	URI          string `json:"uri"`

	Origin *string `json:"origin,omitempty"`
	Beta   *bool   `json:"beta,omitempty"`
	Type   *string `json:"type,omitempty"`

	Capabilities IncomingPhoneNumberCapabilities `json:"capabilities"`

	VoiceURL            *string `json:"voice_url,omitempty"`
	VoiceMethod         *string `json:"voice_method,omitempty"`
	VoiceFallbackURL    *string `json:"voice_fallback_url,omitempty"`
	VoiceFallbackMethod *string `json:"voice_fallback_method,omitempty"`
	VoiceApplicationSid *string `json:"voice_application_sid,omitempty"`
	VoiceCallerIDLookup *bool   `json:"voice_caller_id_lookup,omitempty"`
	VoiceReceiveMode    *string `json:"voice_receive_mode,omitempty"`

	SMSURL            *string `json:"sms_url,omitempty"`
	SMSMethod         *string `json:"sms_method,omitempty"`
	SMSFallbackURL    *string `json:"sms_fallback_url,omitempty"`
	SMSFallbackMethod *string `json:"sms_fallback_method,omitempty"`
	SMSApplicationSid *string `json:"sms_application_sid,omitempty"`

	StatusCallback       *string `json:"status_callback,omitempty"`
	StatusCallbackMethod *string `json:"status_callback_method,omitempty"`

	TrunkSid               *string `json:"trunk_sid,omitempty"`
	AddressSid             *string `json:"address_sid,omitempty"`
	AddressRequirements    *string `json:"address_requirements,omitempty"`
	IdentitySid            *string `json:"identity_sid,omitempty"`
	BundleSid              *string `json:"bundle_sid,omitempty"`
	EmergencyStatus        *string `json:"emergency_status,omitempty"`
	EmergencyAddressSid    *string `json:"emergency_address_sid,omitempty"`
	EmergencyAddressStatus *string `json:"emergency_address_status,omitempty"`

	Status          *string           `json:"status,omitempty"`
	SubresourceURIs map[string]string `json:"subresource_uris,omitempty"`
	DateCreated     string            `json:"date_created,omitempty"`
	DateUpdated     string            `json:"date_updated,omitempty"`
}

// IncomingPhoneNumbersList is the paginated /IncomingPhoneNumbers list response.
type IncomingPhoneNumbersList struct {
	Page
	IncomingPhoneNumbers []IncomingPhoneNumber `json:"incoming_phone_numbers"`
}

// ListIncomingPhoneNumbersParams are the filter / pagination query params
// for GET /IncomingPhoneNumbers and its Local / Mobile / TollFree variants.
// PhoneNumber, when set, is an exact-match E.164 filter. Origin is "twilio"
// or "hosted".
type ListIncomingPhoneNumbersParams struct {
	PhoneNumber  string
	FriendlyName string
	Beta         *bool
	Origin       string
	Page         *int
	PageSize     *int
	PageToken    string
}

func (p ListIncomingPhoneNumbersParams) query() url.Values {
	v := url.Values{}
	setString(v, "PhoneNumber", p.PhoneNumber)
	setString(v, "FriendlyName", p.FriendlyName)
	setBoolP(v, "Beta", p.Beta)
	setString(v, "Origin", p.Origin)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	setString(v, "PageToken", p.PageToken)
	return v
}

// ListTypedIncomingPhoneNumbersParams are the query params for the
// Local / Mobile / TollFree list endpoints.
//
// Deprecated: use ListIncomingPhoneNumbersParams, which now carries the
// same filters.
type ListTypedIncomingPhoneNumbersParams = ListIncomingPhoneNumbersParams

// CreateIncomingPhoneNumberParams is the body for POST /IncomingPhoneNumbers.
// PhoneNumber is required and must be E.164 (leading `+`, 7-15 digits). The
// endpoint is idempotent on the same tenant — re-POSTing the same PhoneNumber
// rebinds its routing without erroring; returns 409 when the number is
// already claimed by a different account. The Sms* fields configure the
// inbound SMS webhook; SmsApplicationSid (AP...) routes inbound SMS to an
// Application instead. VoiceApplicationSid routes inbound voice to an
// Application; StatusCallback / StatusCallbackMethod set the number-level
// lifecycle webhook for inbound calls.
type CreateIncomingPhoneNumberParams struct {
	PhoneNumber          string  `form:"PhoneNumber"`
	VoiceURL             *string `form:"VoiceUrl"`
	VoiceMethod          *string `form:"VoiceMethod"`
	VoiceFallbackURL     *string `form:"VoiceFallbackUrl"`
	VoiceFallbackMethod  *string `form:"VoiceFallbackMethod"`
	SmsURL               *string `form:"SmsUrl"`
	SmsMethod            *string `form:"SmsMethod"`
	SmsFallbackURL       *string `form:"SmsFallbackUrl"`
	SmsFallbackMethod    *string `form:"SmsFallbackMethod"`
	SmsApplicationSid    *string `form:"SmsApplicationSid"`
	VoiceApplicationSid  *string `form:"VoiceApplicationSid"`
	StatusCallback       *string `form:"StatusCallback"`
	StatusCallbackMethod *string `form:"StatusCallbackMethod"`
}

func (p CreateIncomingPhoneNumberParams) form() url.Values {
	v := url.Values{}
	v.Set("PhoneNumber", p.PhoneNumber)
	setStringP(v, "VoiceUrl", p.VoiceURL)
	setStringP(v, "VoiceMethod", p.VoiceMethod)
	setStringP(v, "VoiceFallbackUrl", p.VoiceFallbackURL)
	setStringP(v, "VoiceFallbackMethod", p.VoiceFallbackMethod)
	setSmsRouting(v, p.SmsURL, p.SmsMethod, p.SmsFallbackURL, p.SmsFallbackMethod, p.SmsApplicationSid)
	setStatusCallbackRouting(v, p.VoiceApplicationSid, p.StatusCallback, p.StatusCallbackMethod)
	return v
}

// CreateTypedIncomingPhoneNumberParams is the body for POST
// /IncomingPhoneNumbers/{Local,Mobile,TollFree}. PhoneNumber is required.
// These endpoints accept only the primary voice webhook; configure fallback
// and SMS routing afterwards with Update.
type CreateTypedIncomingPhoneNumberParams struct {
	PhoneNumber string  `form:"PhoneNumber"`
	VoiceURL    *string `form:"VoiceUrl"`
	VoiceMethod *string `form:"VoiceMethod"`
}

func (p CreateTypedIncomingPhoneNumberParams) form() url.Values {
	v := url.Values{}
	v.Set("PhoneNumber", p.PhoneNumber)
	setStringP(v, "VoiceUrl", p.VoiceURL)
	setStringP(v, "VoiceMethod", p.VoiceMethod)
	return v
}

// UpdateIncomingPhoneNumberParams is the body for
// POST /IncomingPhoneNumbers/{sid}. Only-set-fields-touched semantics.
type UpdateIncomingPhoneNumberParams struct {
	VoiceURL             *string `form:"VoiceUrl"`
	VoiceMethod          *string `form:"VoiceMethod"`
	VoiceFallbackURL     *string `form:"VoiceFallbackUrl"`
	VoiceFallbackMethod  *string `form:"VoiceFallbackMethod"`
	SmsURL               *string `form:"SmsUrl"`
	SmsMethod            *string `form:"SmsMethod"`
	SmsFallbackURL       *string `form:"SmsFallbackUrl"`
	SmsFallbackMethod    *string `form:"SmsFallbackMethod"`
	SmsApplicationSid    *string `form:"SmsApplicationSid"`
	VoiceApplicationSid  *string `form:"VoiceApplicationSid"`
	StatusCallback       *string `form:"StatusCallback"`
	StatusCallbackMethod *string `form:"StatusCallbackMethod"`
}

func (p UpdateIncomingPhoneNumberParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "VoiceUrl", p.VoiceURL)
	setStringP(v, "VoiceMethod", p.VoiceMethod)
	setStringP(v, "VoiceFallbackUrl", p.VoiceFallbackURL)
	setStringP(v, "VoiceFallbackMethod", p.VoiceFallbackMethod)
	setSmsRouting(v, p.SmsURL, p.SmsMethod, p.SmsFallbackURL, p.SmsFallbackMethod, p.SmsApplicationSid)
	setStatusCallbackRouting(v, p.VoiceApplicationSid, p.StatusCallback, p.StatusCallbackMethod)
	return v
}

func setSmsRouting(v url.Values, smsURL, method, fallbackURL, fallbackMethod, appSid *string) {
	setStringP(v, "SmsUrl", smsURL)
	setStringP(v, "SmsMethod", method)
	setStringP(v, "SmsFallbackUrl", fallbackURL)
	setStringP(v, "SmsFallbackMethod", fallbackMethod)
	setStringP(v, "SmsApplicationSid", appSid)
}

func setStatusCallbackRouting(v url.Values, voiceAppSid, statusCallback, statusCallbackMethod *string) {
	setStringP(v, "VoiceApplicationSid", voiceAppSid)
	setStringP(v, "StatusCallback", statusCallback)
	setStringP(v, "StatusCallbackMethod", statusCallbackMethod)
}

// List returns the page of DIDs assigned to this tenant.
// GET /IncomingPhoneNumbers. Pass nil for an unfiltered first-page request.
func (s *IncomingPhoneNumbersService) List(ctx context.Context, params *ListIncomingPhoneNumbersParams) (*IncomingPhoneNumbersList, error) {
	var q url.Values
	if params != nil {
		q = params.query()
	}
	var out IncomingPhoneNumbersList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("IncomingPhoneNumbers"),
		query:  q,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Iterate walks every DID assigned to this tenant lazily. Pass nil for no
// filters. See Pager.
func (s *IncomingPhoneNumbersService) Iterate(ctx context.Context, params *ListIncomingPhoneNumbersParams) *Pager[IncomingPhoneNumber] {
	return s.iterate(ctx, "", params)
}

// Create assigns a DID to the authenticated tenant.
// POST /IncomingPhoneNumbers.
func (s *IncomingPhoneNumbersService) Create(ctx context.Context, params CreateIncomingPhoneNumberParams) (*IncomingPhoneNumber, error) {
	var out IncomingPhoneNumber
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("IncomingPhoneNumbers"),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single DID by its PN-prefixed SID.
// GET /IncomingPhoneNumbers/{sid}.
func (s *IncomingPhoneNumbersService) Get(ctx context.Context, sid string) (*IncomingPhoneNumber, error) {
	var out IncomingPhoneNumber
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("IncomingPhoneNumbers", sid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update mutates voice and SMS routing on an assigned DID.
// POST /IncomingPhoneNumbers/{sid}. Only-set-fields-touched.
func (s *IncomingPhoneNumbersService) Update(ctx context.Context, sid string, params UpdateIncomingPhoneNumberParams) (*IncomingPhoneNumber, error) {
	var out IncomingPhoneNumber
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("IncomingPhoneNumbers", sid),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *IncomingPhoneNumbersService) listTyped(ctx context.Context, kind string, params *ListIncomingPhoneNumbersParams) (*IncomingPhoneNumbersList, error) {
	var q url.Values
	if params != nil {
		q = params.query()
	}
	var out IncomingPhoneNumbersList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("IncomingPhoneNumbers", kind),
		query:  q,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *IncomingPhoneNumbersService) iterate(ctx context.Context, kind string, params *ListIncomingPhoneNumbersParams) *Pager[IncomingPhoneNumber] {
	var q url.Values
	if params != nil {
		q = params.query()
	}
	path := s.c.pathf("IncomingPhoneNumbers")
	if kind != "" {
		path = s.c.pathf("IncomingPhoneNumbers", kind)
	}
	return newPager(ctx, s.c, s.c.t, requestOpts{method: "GET", path: path, query: q},
		func(l *IncomingPhoneNumbersList) ([]IncomingPhoneNumber, PageLinks) {
			return l.IncomingPhoneNumbers, l.Page
		})
}

func (s *IncomingPhoneNumbersService) createTyped(ctx context.Context, kind string, params CreateTypedIncomingPhoneNumberParams) (*IncomingPhoneNumber, error) {
	var out IncomingPhoneNumber
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("IncomingPhoneNumbers", kind),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListLocal returns local DIDs. GET /IncomingPhoneNumbers/Local.
func (s *IncomingPhoneNumbersService) ListLocal(ctx context.Context, params *ListIncomingPhoneNumbersParams) (*IncomingPhoneNumbersList, error) {
	return s.listTyped(ctx, "Local", params)
}

// IterateLocal walks every Local DID lazily. Pass nil for no filters.
func (s *IncomingPhoneNumbersService) IterateLocal(ctx context.Context, params *ListIncomingPhoneNumbersParams) *Pager[IncomingPhoneNumber] {
	return s.iterate(ctx, "Local", params)
}

// CreateLocal assigns a local DID. POST /IncomingPhoneNumbers/Local.
func (s *IncomingPhoneNumbersService) CreateLocal(ctx context.Context, params CreateTypedIncomingPhoneNumberParams) (*IncomingPhoneNumber, error) {
	return s.createTyped(ctx, "Local", params)
}

// ListMobile returns mobile DIDs. GET /IncomingPhoneNumbers/Mobile.
func (s *IncomingPhoneNumbersService) ListMobile(ctx context.Context, params *ListIncomingPhoneNumbersParams) (*IncomingPhoneNumbersList, error) {
	return s.listTyped(ctx, "Mobile", params)
}

// IterateMobile walks every Mobile DID lazily. Pass nil for no filters.
func (s *IncomingPhoneNumbersService) IterateMobile(ctx context.Context, params *ListIncomingPhoneNumbersParams) *Pager[IncomingPhoneNumber] {
	return s.iterate(ctx, "Mobile", params)
}

// CreateMobile assigns a mobile DID. POST /IncomingPhoneNumbers/Mobile.
func (s *IncomingPhoneNumbersService) CreateMobile(ctx context.Context, params CreateTypedIncomingPhoneNumberParams) (*IncomingPhoneNumber, error) {
	return s.createTyped(ctx, "Mobile", params)
}

// ListTollFree returns toll-free DIDs. GET /IncomingPhoneNumbers/TollFree.
func (s *IncomingPhoneNumbersService) ListTollFree(ctx context.Context, params *ListIncomingPhoneNumbersParams) (*IncomingPhoneNumbersList, error) {
	return s.listTyped(ctx, "TollFree", params)
}

// IterateTollFree walks every TollFree DID lazily. Pass nil for no filters.
func (s *IncomingPhoneNumbersService) IterateTollFree(ctx context.Context, params *ListIncomingPhoneNumbersParams) *Pager[IncomingPhoneNumber] {
	return s.iterate(ctx, "TollFree", params)
}

// CreateTollFree assigns a toll-free DID. POST /IncomingPhoneNumbers/TollFree.
func (s *IncomingPhoneNumbersService) CreateTollFree(ctx context.Context, params CreateTypedIncomingPhoneNumberParams) (*IncomingPhoneNumber, error) {
	return s.createTyped(ctx, "TollFree", params)
}

// Delete releases a DID from the authenticated tenant.
// DELETE /IncomingPhoneNumbers/{sid}. Idempotent: returns nil on success or
// when the number was already gone.
func (s *IncomingPhoneNumbersService) Delete(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("IncomingPhoneNumbers", sid),
	}, nil)
}
