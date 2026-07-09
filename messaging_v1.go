// Messaging v1 (messaging.twilio.com/v1) resources: the Messaging Service
// (MG...) CRUD surface under /v1/Services.
//
// A Messaging Service shares the /v1/Services path shape with the Conversations
// Service (IS...); the two are disambiguated on the wire by host
// (messaging.voicetel.com vs conversations.voicetel.com). This SDK routes the
// whole c.MessagingV1 group at the messaging host automatically — see hosts.go.
// Only the Messaging Service has an update verb (POST /v1/Services/{sid}), so
// there is no path collision with the Conversation Service.
//
// The /v1 namespace omits /Accounts/{AccountSid}; the account is resolved from
// HTTP Basic auth. The list response shares the VoiceV1Meta envelope declared
// in voice_v1.go.

package voiceml

import (
	"context"
	"net/url"
)

// ---------------------------------------------------------------------------
// Response models.
// ---------------------------------------------------------------------------

// MessagingService is a Messaging Service (MG...). The feature-toggle fields
// (StickySender, MmsConverter, ...) are accept-and-echo on VoiceML; the
// service's operative role is gating scheduled sends (a real
// messaging_service_sid is required on POST /Messages when send_at /
// schedule_type is set).
type MessagingService struct {
	Sid                       *string `json:"sid"`
	AccountSid                *string `json:"account_sid"`
	FriendlyName              *string `json:"friendly_name"`
	DateCreated               *string `json:"date_created"`
	DateUpdated               *string `json:"date_updated"`
	InboundRequestURL         *string `json:"inbound_request_url"`
	InboundMethod             *string `json:"inbound_method"`
	FallbackURL               *string `json:"fallback_url"`
	FallbackMethod            *string `json:"fallback_method"`
	StatusCallback            *string `json:"status_callback"`
	StickySender              *bool   `json:"sticky_sender"`
	MmsConverter              *bool   `json:"mms_converter"`
	SmartEncoding             *bool   `json:"smart_encoding"`
	ScanMessageContent        *string `json:"scan_message_content"`
	FallbackToLongCode        *bool   `json:"fallback_to_long_code"`
	AreaCodeGeomatch          *bool   `json:"area_code_geomatch"`
	SynchronousValidation     *bool   `json:"synchronous_validation"`
	ValidityPeriod            *int    `json:"validity_period"`
	URL                       *string `json:"url"`
	Usecase                   *string `json:"usecase"`
	UseInboundWebhookOnNumber *bool   `json:"use_inbound_webhook_on_number"`
}

// MessagingServiceList is the paginated /v1/Services response on the messaging
// host.
type MessagingServiceList struct {
	Services []MessagingService `json:"services"`
	Meta     VoiceV1Meta        `json:"meta"`
}

// ---------------------------------------------------------------------------
// Request params.
// ---------------------------------------------------------------------------

// CreateMessagingServiceParams is the body for POST /v1/Services (messaging
// host). FriendlyName is required; every other field is optional.
type CreateMessagingServiceParams struct {
	FriendlyName              string
	InboundRequestURL         *string
	InboundMethod             *string
	FallbackURL               *string
	FallbackMethod            *string
	StatusCallback            *string
	StickySender              *bool
	MmsConverter              *bool
	SmartEncoding             *bool
	ScanMessageContent        *string
	FallbackToLongCode        *bool
	AreaCodeGeomatch          *bool
	SynchronousValidation     *bool
	ValidityPeriod            *int
	Usecase                   *string
	UseInboundWebhookOnNumber *bool
}

func (p CreateMessagingServiceParams) form() url.Values {
	v := url.Values{}
	v.Set("FriendlyName", p.FriendlyName)
	setMessagingServiceOptional(v, messagingServiceOptional{
		InboundRequestURL:         p.InboundRequestURL,
		InboundMethod:             p.InboundMethod,
		FallbackURL:               p.FallbackURL,
		FallbackMethod:            p.FallbackMethod,
		StatusCallback:            p.StatusCallback,
		StickySender:              p.StickySender,
		MmsConverter:              p.MmsConverter,
		SmartEncoding:             p.SmartEncoding,
		ScanMessageContent:        p.ScanMessageContent,
		FallbackToLongCode:        p.FallbackToLongCode,
		AreaCodeGeomatch:          p.AreaCodeGeomatch,
		SynchronousValidation:     p.SynchronousValidation,
		ValidityPeriod:            p.ValidityPeriod,
		Usecase:                   p.Usecase,
		UseInboundWebhookOnNumber: p.UseInboundWebhookOnNumber,
	})
	return v
}

// UpdateMessagingServiceParams is the body for POST /v1/Services/{sid}
// (messaging host). Every field is optional.
type UpdateMessagingServiceParams struct {
	FriendlyName              *string
	InboundRequestURL         *string
	InboundMethod             *string
	FallbackURL               *string
	FallbackMethod            *string
	StatusCallback            *string
	StickySender              *bool
	MmsConverter              *bool
	SmartEncoding             *bool
	ScanMessageContent        *string
	FallbackToLongCode        *bool
	AreaCodeGeomatch          *bool
	SynchronousValidation     *bool
	ValidityPeriod            *int
	Usecase                   *string
	UseInboundWebhookOnNumber *bool
}

func (p UpdateMessagingServiceParams) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setMessagingServiceOptional(v, messagingServiceOptional{
		InboundRequestURL:         p.InboundRequestURL,
		InboundMethod:             p.InboundMethod,
		FallbackURL:               p.FallbackURL,
		FallbackMethod:            p.FallbackMethod,
		StatusCallback:            p.StatusCallback,
		StickySender:              p.StickySender,
		MmsConverter:              p.MmsConverter,
		SmartEncoding:             p.SmartEncoding,
		ScanMessageContent:        p.ScanMessageContent,
		FallbackToLongCode:        p.FallbackToLongCode,
		AreaCodeGeomatch:          p.AreaCodeGeomatch,
		SynchronousValidation:     p.SynchronousValidation,
		ValidityPeriod:            p.ValidityPeriod,
		Usecase:                   p.Usecase,
		UseInboundWebhookOnNumber: p.UseInboundWebhookOnNumber,
	})
	return v
}

// messagingServiceOptional holds the config fields shared by the create and
// update bodies so their PascalCase wire names are declared once.
type messagingServiceOptional struct {
	InboundRequestURL         *string
	InboundMethod             *string
	FallbackURL               *string
	FallbackMethod            *string
	StatusCallback            *string
	StickySender              *bool
	MmsConverter              *bool
	SmartEncoding             *bool
	ScanMessageContent        *string
	FallbackToLongCode        *bool
	AreaCodeGeomatch          *bool
	SynchronousValidation     *bool
	ValidityPeriod            *int
	Usecase                   *string
	UseInboundWebhookOnNumber *bool
}

func setMessagingServiceOptional(v url.Values, o messagingServiceOptional) {
	setStr(v, "InboundRequestUrl", o.InboundRequestURL)
	setStr(v, "InboundMethod", o.InboundMethod)
	setStr(v, "FallbackUrl", o.FallbackURL)
	setStr(v, "FallbackMethod", o.FallbackMethod)
	setStr(v, "StatusCallback", o.StatusCallback)
	setBool(v, "StickySender", o.StickySender)
	setBool(v, "MmsConverter", o.MmsConverter)
	setBool(v, "SmartEncoding", o.SmartEncoding)
	setStr(v, "ScanMessageContent", o.ScanMessageContent)
	setBool(v, "FallbackToLongCode", o.FallbackToLongCode)
	setBool(v, "AreaCodeGeomatch", o.AreaCodeGeomatch)
	setBool(v, "SynchronousValidation", o.SynchronousValidation)
	setInt(v, "ValidityPeriod", o.ValidityPeriod)
	setStr(v, "Usecase", o.Usecase)
	setBool(v, "UseInboundWebhookOnNumber", o.UseInboundWebhookOnNumber)
}

// ---------------------------------------------------------------------------
// Services.
// ---------------------------------------------------------------------------

// MessagingV1Service bundles the messaging.twilio.com/v1 sub-services. Reach it
// as c.MessagingV1.
type MessagingV1Service struct {
	// Services — CRUD on /v1/Services Messaging Services (MG...).
	Services *MessagingV1ServicesService
}

// MessagingV1ServicesService surfaces the /v1/Services CRUD verbs on the
// messaging host.
type MessagingV1ServicesService struct{ t *transport }

// Create adds a Messaging Service.
func (s *MessagingV1ServicesService) Create(ctx context.Context, params CreateMessagingServiceParams) (*MessagingService, error) {
	var out MessagingService
	if err := s.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a single page of Messaging Services.
func (s *MessagingV1ServicesService) List(ctx context.Context, params V1PageParams) (*MessagingServiceList, error) {
	var out MessagingServiceList
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch retrieves a Messaging Service by sid.
func (s *MessagingV1ServicesService) Fetch(ctx context.Context, sid string) (*MessagingService, error) {
	var out MessagingService
	if err := s.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update mutates a Messaging Service in place.
func (s *MessagingV1ServicesService) Update(ctx context.Context, sid string, params UpdateMessagingServiceParams) (*MessagingService, error) {
	var out MessagingService
	if err := s.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a Messaging Service.
func (s *MessagingV1ServicesService) Delete(ctx context.Context, sid string) error {
	return s.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Services/" + sid,
	}, nil)
}
