// Pricing v1/v2 (pricing.twilio.com) resources. Read-only GETs.
//
// VoiceML has no dedicated pricing subdomain, so these live on the default host
// (voiceml.voicetel.com) under /v1 and /v2. VoiceML is NANP-only: every
// Countries list carries exactly one entry (the tenant's own country), and a
// Numbers fetch 404s for a non-NANP destination. Number path segments are
// URL-encoded (E.164 "+" becomes "%2B").
//
// Layout:
//
//	c.Pricing.V1.Voice.Countries.List / Fetch
//	c.Pricing.V1.Voice.Numbers.Fetch
//	c.Pricing.V1.Messaging.Countries.List / Fetch
//	c.Pricing.V1.PhoneNumbers.Countries.List / Fetch
//	c.Pricing.V2.Voice.Countries.List / Fetch
//	c.Pricing.V2.Voice.Numbers.Fetch
//	c.Pricing.V2.Trunking.Countries.List / Fetch
//	c.Pricing.V2.Trunking.Numbers.Fetch

package voiceml

import (
	"context"
	"net/url"
)

// ---------------------------------------------------------------------------
// Price leaves.
// ---------------------------------------------------------------------------

// PricingInboundCallPrice is a single inbound price row. NumberType is
// "local" or "toll free".
type PricingInboundCallPrice struct {
	BasePrice    *string `json:"base_price"`
	CurrentPrice *string `json:"current_price"`
	NumberType   *string `json:"number_type"`
}

// PricingOutboundCallPrice is a single outbound price row.
type PricingOutboundCallPrice struct {
	BasePrice    *string `json:"base_price"`
	CurrentPrice *string `json:"current_price"`
}

// PricingOutboundCallPriceWithOrigin is an outbound price row keyed by
// origination prefix (v2).
type PricingOutboundCallPriceWithOrigin struct {
	OriginationPrefixes []string `json:"origination_prefixes"`
	BasePrice           *string  `json:"base_price"`
	CurrentPrice        *string  `json:"current_price"`
}

// PricingOutboundPrefixPrice is an outbound price row keyed by destination
// prefix (v1).
type PricingOutboundPrefixPrice struct {
	Prefixes     []string `json:"prefixes"`
	BasePrice    *string  `json:"base_price"`
	CurrentPrice *string  `json:"current_price"`
	FriendlyName *string  `json:"friendly_name"`
}

// PricingOutboundPrefixPriceWithOrigin is an outbound price row keyed by both
// origination and destination prefixes (v2).
type PricingOutboundPrefixPriceWithOrigin struct {
	OriginationPrefixes []string `json:"origination_prefixes"`
	DestinationPrefixes []string `json:"destination_prefixes"`
	BasePrice           *string  `json:"base_price"`
	CurrentPrice        *string  `json:"current_price"`
	FriendlyName        *string  `json:"friendly_name"`
}

// PricingOutboundSMSPrice is a per-carrier outbound SMS price row.
type PricingOutboundSMSPrice struct {
	Carrier *string                   `json:"carrier"`
	Mcc     *string                   `json:"mcc"`
	Mnc     *string                   `json:"mnc"`
	Prices  []PricingInboundCallPrice `json:"prices"`
}

// PricingPhoneNumberPrice is a per-number-type phone number price row.
type PricingPhoneNumberPrice struct {
	NumberType   *string `json:"number_type"`
	BasePrice    *string `json:"base_price"`
	CurrentPrice *string `json:"current_price"`
}

// ---------------------------------------------------------------------------
// Countries list envelope.
// ---------------------------------------------------------------------------

// PricingCountryRef is one entry in a Countries list.
type PricingCountryRef struct {
	Country    *string `json:"country"`
	IsoCountry *string `json:"iso_country"`
	URL        *string `json:"url"`
}

// PricingCountriesList is the shared list envelope returned by every
// Countries.List call.
type PricingCountriesList struct {
	Countries []PricingCountryRef `json:"countries"`
	Meta      VoiceV1Meta         `json:"meta"`
}

// ---------------------------------------------------------------------------
// Pricing v1 country / number bodies.
// ---------------------------------------------------------------------------

// PricingVoiceCountry is the v1 per-country voice price body.
type PricingVoiceCountry struct {
	Country              *string                      `json:"country"`
	IsoCountry           *string                      `json:"iso_country"`
	OutboundPrefixPrices []PricingOutboundPrefixPrice `json:"outbound_prefix_prices"`
	InboundCallPrices    []PricingInboundCallPrice    `json:"inbound_call_prices"`
	PriceUnit            *string                      `json:"price_unit"`
	URL                  *string                      `json:"url"`
}

// PricingVoiceNumber is the v1 per-number voice price body.
type PricingVoiceNumber struct {
	Number            *string                   `json:"number"`
	Country           *string                   `json:"country"`
	IsoCountry        *string                   `json:"iso_country"`
	OutboundCallPrice *PricingOutboundCallPrice `json:"outbound_call_price"`
	InboundCallPrice  *PricingInboundCallPrice  `json:"inbound_call_price"`
	PriceUnit         *string                   `json:"price_unit"`
	URL               *string                   `json:"url"`
}

// PricingMessagingCountry is the v1 per-country messaging price body.
type PricingMessagingCountry struct {
	Country           *string                   `json:"country"`
	IsoCountry        *string                   `json:"iso_country"`
	OutboundSMSPrices []PricingOutboundSMSPrice `json:"outbound_sms_prices"`
	InboundSMSPrices  []PricingInboundCallPrice `json:"inbound_sms_prices"`
	PriceUnit         *string                   `json:"price_unit"`
	URL               *string                   `json:"url"`
}

// PricingPhoneNumberCountry is the v1 per-country phone number price body.
type PricingPhoneNumberCountry struct {
	Country           *string                   `json:"country"`
	IsoCountry        *string                   `json:"iso_country"`
	PhoneNumberPrices []PricingPhoneNumberPrice `json:"phone_number_prices"`
	PriceUnit         *string                   `json:"price_unit"`
	URL               *string                   `json:"url"`
}

// ---------------------------------------------------------------------------
// Pricing v2 country / number bodies.
// ---------------------------------------------------------------------------

// PricingVoiceCountryV2 is the v2 per-country voice price body.
type PricingVoiceCountryV2 struct {
	Country              *string                                `json:"country"`
	IsoCountry           *string                                `json:"iso_country"`
	OutboundPrefixPrices []PricingOutboundPrefixPriceWithOrigin `json:"outbound_prefix_prices"`
	InboundCallPrices    []PricingInboundCallPrice              `json:"inbound_call_prices"`
	PriceUnit            *string                                `json:"price_unit"`
	URL                  *string                                `json:"url"`
}

// PricingVoiceNumberV2 is the v2 per-number voice price body.
type PricingVoiceNumberV2 struct {
	DestinationNumber  *string                              `json:"destination_number"`
	OriginationNumber  *string                              `json:"origination_number"`
	Country            *string                              `json:"country"`
	IsoCountry         *string                              `json:"iso_country"`
	OutboundCallPrices []PricingOutboundCallPriceWithOrigin `json:"outbound_call_prices"`
	InboundCallPrice   *PricingInboundCallPrice             `json:"inbound_call_price"`
	PriceUnit          *string                              `json:"price_unit"`
	URL                *string                              `json:"url"`
}

// PricingTrunkingCountry is the v2 per-country trunking price body.
type PricingTrunkingCountry struct {
	Country                 *string                                `json:"country"`
	IsoCountry              *string                                `json:"iso_country"`
	TerminatingPrefixPrices []PricingOutboundPrefixPriceWithOrigin `json:"terminating_prefix_prices"`
	OriginatingCallPrices   []PricingInboundCallPrice              `json:"originating_call_prices"`
	PriceUnit               *string                                `json:"price_unit"`
	URL                     *string                                `json:"url"`
}

// PricingTrunkingNumber is the v2 per-number trunking price body.
type PricingTrunkingNumber struct {
	DestinationNumber       *string                                `json:"destination_number"`
	OriginationNumber       *string                                `json:"origination_number"`
	Country                 *string                                `json:"country"`
	IsoCountry              *string                                `json:"iso_country"`
	TerminatingPrefixPrices []PricingOutboundPrefixPriceWithOrigin `json:"terminating_prefix_prices"`
	OriginatingCallPrice    *PricingInboundCallPrice               `json:"originating_call_price"`
	PriceUnit               *string                                `json:"price_unit"`
	URL                     *string                                `json:"url"`
}

// ---------------------------------------------------------------------------
// Request params.
// ---------------------------------------------------------------------------

// PricingNumberParams are the optional query knobs a v2 Numbers.Fetch accepts.
type PricingNumberParams struct {
	// OriginationNumber narrows the quote to a specific origination E.164.
	OriginationNumber *string
}

func (p PricingNumberParams) query() url.Values {
	v := url.Values{}
	setStr(v, "OriginationNumber", p.OriginationNumber)
	return v
}

// ---------------------------------------------------------------------------
// Services.
// ---------------------------------------------------------------------------

// PricingService bundles the read-only pricing.twilio.com v1/v2 surface. Reach
// it as c.Pricing.
type PricingService struct {
	// V1 — the pricing.twilio.com/v1 surface.
	V1 *PricingV1Service
	// V2 — the pricing.twilio.com/v2 surface.
	V2 *PricingV2Service
}

// PricingV1Service groups the v1 pricing products.
type PricingV1Service struct {
	Voice        *PricingV1VoiceService
	Messaging    *PricingV1MessagingService
	PhoneNumbers *PricingV1PhoneNumbersService
}

// PricingV2Service groups the v2 pricing products.
type PricingV2Service struct {
	Voice    *PricingV2VoiceService
	Trunking *PricingV2TrunkingService
}

// PricingV1VoiceService is c.Pricing.V1.Voice.
type PricingV1VoiceService struct {
	Countries *PricingV1VoiceCountriesService
	Numbers   *PricingV1VoiceNumbersService
}

// PricingV1MessagingService is c.Pricing.V1.Messaging.
type PricingV1MessagingService struct {
	Countries *PricingV1MessagingCountriesService
}

// PricingV1PhoneNumbersService is c.Pricing.V1.PhoneNumbers.
type PricingV1PhoneNumbersService struct {
	Countries *PricingV1PhoneNumbersCountriesService
}

// PricingV2VoiceService is c.Pricing.V2.Voice.
type PricingV2VoiceService struct {
	Countries *PricingV2VoiceCountriesService
	Numbers   *PricingV2VoiceNumbersService
}

// PricingV2TrunkingService is c.Pricing.V2.Trunking.
type PricingV2TrunkingService struct {
	Countries *PricingV2TrunkingCountriesService
	Numbers   *PricingV2TrunkingNumbersService
}

// newPricingService wires the whole pricing tree onto the default transport.
func newPricingService(t *transport) *PricingService {
	return &PricingService{
		V1: &PricingV1Service{
			Voice: &PricingV1VoiceService{
				Countries: &PricingV1VoiceCountriesService{t: t},
				Numbers:   &PricingV1VoiceNumbersService{t: t},
			},
			Messaging: &PricingV1MessagingService{
				Countries: &PricingV1MessagingCountriesService{t: t},
			},
			PhoneNumbers: &PricingV1PhoneNumbersService{
				Countries: &PricingV1PhoneNumbersCountriesService{t: t},
			},
		},
		V2: &PricingV2Service{
			Voice: &PricingV2VoiceService{
				Countries: &PricingV2VoiceCountriesService{t: t},
				Numbers:   &PricingV2VoiceNumbersService{t: t},
			},
			Trunking: &PricingV2TrunkingService{
				Countries: &PricingV2TrunkingCountriesService{t: t},
				Numbers:   &PricingV2TrunkingNumbersService{t: t},
			},
		},
	}
}

// listCountries is the shared GET .../Countries helper.
func listCountries(ctx context.Context, t *transport, path string, params V1PageParams) (*PricingCountriesList, error) {
	var out PricingCountriesList
	if err := t.do(ctx, requestOpts{
		method: "GET", path: path, query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- v1 Voice --------------------------------------------------------------

// PricingV1VoiceCountriesService surfaces GET /v1/Voice/Countries[/{Iso}].
type PricingV1VoiceCountriesService struct{ t *transport }

// List returns the Countries envelope for v1 voice pricing.
func (s *PricingV1VoiceCountriesService) List(ctx context.Context, params V1PageParams) (*PricingCountriesList, error) {
	return listCountries(ctx, s.t, "/v1/Voice/Countries", params)
}

// Fetch returns the v1 voice price body for a country.
func (s *PricingV1VoiceCountriesService) Fetch(ctx context.Context, isoCountry string) (*PricingVoiceCountry, error) {
	var out PricingVoiceCountry
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Voice/Countries/" + isoCountry,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PricingV1VoiceNumbersService surfaces GET /v1/Voice/Numbers/{Number}.
type PricingV1VoiceNumbersService struct{ t *transport }

// Fetch returns the v1 voice price body for a number.
func (s *PricingV1VoiceNumbersService) Fetch(ctx context.Context, number string) (*PricingVoiceNumber, error) {
	var out PricingVoiceNumber
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Voice/Numbers/" + url.QueryEscape(number),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- v1 Messaging ----------------------------------------------------------

// PricingV1MessagingCountriesService surfaces GET /v1/Messaging/Countries[/{Iso}].
type PricingV1MessagingCountriesService struct{ t *transport }

// List returns the Countries envelope for v1 messaging pricing.
func (s *PricingV1MessagingCountriesService) List(ctx context.Context, params V1PageParams) (*PricingCountriesList, error) {
	return listCountries(ctx, s.t, "/v1/Messaging/Countries", params)
}

// Fetch returns the v1 messaging price body for a country.
func (s *PricingV1MessagingCountriesService) Fetch(ctx context.Context, isoCountry string) (*PricingMessagingCountry, error) {
	var out PricingMessagingCountry
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Messaging/Countries/" + isoCountry,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- v1 PhoneNumbers -------------------------------------------------------

// PricingV1PhoneNumbersCountriesService surfaces GET /v1/PhoneNumbers/Countries[/{Iso}].
type PricingV1PhoneNumbersCountriesService struct{ t *transport }

// List returns the Countries envelope for v1 phone number pricing.
func (s *PricingV1PhoneNumbersCountriesService) List(ctx context.Context, params V1PageParams) (*PricingCountriesList, error) {
	return listCountries(ctx, s.t, "/v1/PhoneNumbers/Countries", params)
}

// Fetch returns the v1 phone number price body for a country.
func (s *PricingV1PhoneNumbersCountriesService) Fetch(ctx context.Context, isoCountry string) (*PricingPhoneNumberCountry, error) {
	var out PricingPhoneNumberCountry
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/PhoneNumbers/Countries/" + isoCountry,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- v2 Voice --------------------------------------------------------------

// PricingV2VoiceCountriesService surfaces GET /v2/Voice/Countries[/{Iso}].
type PricingV2VoiceCountriesService struct{ t *transport }

// List returns the Countries envelope for v2 voice pricing.
func (s *PricingV2VoiceCountriesService) List(ctx context.Context, params V1PageParams) (*PricingCountriesList, error) {
	return listCountries(ctx, s.t, "/v2/Voice/Countries", params)
}

// Fetch returns the v2 voice price body for a country.
func (s *PricingV2VoiceCountriesService) Fetch(ctx context.Context, isoCountry string) (*PricingVoiceCountryV2, error) {
	var out PricingVoiceCountryV2
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v2/Voice/Countries/" + isoCountry,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PricingV2VoiceNumbersService surfaces GET /v2/Voice/Numbers/{Dest}.
type PricingV2VoiceNumbersService struct{ t *transport }

// Fetch returns the v2 voice price body for a destination number, optionally
// scoped to an origination number.
func (s *PricingV2VoiceNumbersService) Fetch(ctx context.Context, destinationNumber string, params PricingNumberParams) (*PricingVoiceNumberV2, error) {
	var out PricingVoiceNumberV2
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v2/Voice/Numbers/" + url.QueryEscape(destinationNumber), query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- v2 Trunking -----------------------------------------------------------

// PricingV2TrunkingCountriesService surfaces GET /v2/Trunking/Countries[/{Iso}].
type PricingV2TrunkingCountriesService struct{ t *transport }

// List returns the Countries envelope for v2 trunking pricing.
func (s *PricingV2TrunkingCountriesService) List(ctx context.Context, params V1PageParams) (*PricingCountriesList, error) {
	return listCountries(ctx, s.t, "/v2/Trunking/Countries", params)
}

// Fetch returns the v2 trunking price body for a country.
func (s *PricingV2TrunkingCountriesService) Fetch(ctx context.Context, isoCountry string) (*PricingTrunkingCountry, error) {
	var out PricingTrunkingCountry
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v2/Trunking/Countries/" + isoCountry,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PricingV2TrunkingNumbersService surfaces GET /v2/Trunking/Numbers/{Dest}.
type PricingV2TrunkingNumbersService struct{ t *transport }

// Fetch returns the v2 trunking price body for a destination number, optionally
// scoped to an origination number.
func (s *PricingV2TrunkingNumbersService) Fetch(ctx context.Context, destinationNumber string, params PricingNumberParams) (*PricingTrunkingNumber, error) {
	var out PricingTrunkingNumber
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v2/Trunking/Numbers/" + url.QueryEscape(destinationNumber), query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
