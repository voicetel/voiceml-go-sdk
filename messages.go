package voiceml

import (
	"context"
	"net/url"
)

// MessagesService surfaces the /Messages REST resource — VoiceTel's
// Twilio-compatible SMS surface, backed by the SDK 2.2 gateway.
// Outbound-only today (no MMS, no inbound webhook delivery).
type MessagesService struct{ c *Client }

// Message mirrors Twilio's Message resource on the wire. Status pins
// to "sent" on a successful SDK 2.2 dispatch and "failed" otherwise —
// there is no in-flight "queued"/"sending"/"delivered" lifecycle
// today because the gateway is fire-and-forget.
type Message struct {
	Sid                 string            `json:"sid"`
	AccountSid          string            `json:"account_sid"`
	APIVersion          string            `json:"api_version"`
	To                  string            `json:"to"`
	From                string            `json:"from"`
	Body                string            `json:"body"`
	Status              string            `json:"status"`
	NumSegments         string            `json:"num_segments"`
	NumMedia            string            `json:"num_media"`
	Direction           string            `json:"direction"`
	Price               *string           `json:"price"`
	PriceUnit           *string           `json:"price_unit"`
	ErrorCode           *int              `json:"error_code"`
	ErrorMessage        *string           `json:"error_message"`
	MessagingServiceSid *string           `json:"messaging_service_sid"`
	DateCreated         string            `json:"date_created"`
	DateUpdated         string            `json:"date_updated"`
	DateSent            *string           `json:"date_sent"`
	URI                 string            `json:"uri"`
	SubresourceURIs     map[string]string `json:"subresource_uris,omitempty"`
}

// MessageList is the paginated /Messages list response.
type MessageList struct {
	Page
	Messages []Message `json:"messages"`
}

// CreateMessageParams is the body for POST /Messages. To and Body are
// required; From falls back to the server's SMS_FROM_NUMBER when
// omitted.
type CreateMessageParams struct {
	To                  string  `form:"To"`
	Body                string  `form:"Body"`
	From                *string `form:"From"`
	MessagingServiceSid *string `form:"MessagingServiceSid"`
	StatusCallback      *string `form:"StatusCallback"`
}

func (p CreateMessageParams) form() url.Values {
	v := url.Values{}
	v.Set("To", p.To)
	v.Set("Body", p.Body)
	if p.From != nil {
		v.Set("From", *p.From)
	}
	if p.MessagingServiceSid != nil {
		v.Set("MessagingServiceSid", *p.MessagingServiceSid)
	}
	if p.StatusCallback != nil {
		v.Set("StatusCallback", *p.StatusCallback)
	}
	return v
}

// Create dispatches an outbound SMS.
func (s *MessagesService) Create(ctx context.Context, params CreateMessageParams) (*Message, error) {
	var out Message
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Messages"),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch retrieves a previously-sent Message by sid.
func (s *MessagesService) Fetch(ctx context.Context, sid string) (*Message, error) {
	var out Message
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Messages", sid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListMessagesParams narrows GET /Messages — Twilio's documented
// filter set (To, From, DateSent eq/gt/lt) plus pagination.
type ListMessagesParams struct {
	To           *string
	From         *string
	DateSent     *string
	DateSentLess *string
	DateSentMore *string
	ListPageParams
}

func (p ListMessagesParams) query() url.Values {
	v := p.ListPageParams.query()
	if p.To != nil {
		v.Set("To", *p.To)
	}
	if p.From != nil {
		v.Set("From", *p.From)
	}
	if p.DateSent != nil {
		v.Set("DateSent", *p.DateSent)
	}
	if p.DateSentLess != nil {
		v.Set("DateSent<", *p.DateSentLess)
	}
	if p.DateSentMore != nil {
		v.Set("DateSent>", *p.DateSentMore)
	}
	return v
}

// List returns a single page of Messages.
func (s *MessagesService) List(ctx context.Context, params ListMessagesParams) (*MessageList, error) {
	var out MessageList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Messages"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMessageParams is the body for POST /Messages/{Sid}. Only
// Body=""  (redaction) is honoured by the server today; Status=canceled
// returns 21610 because VoiceTel SDK 2.2 is fire-and-forget.
type UpdateMessageParams struct {
	Body   *string `form:"Body"`
	Status *string `form:"Status"`
}

func (p UpdateMessageParams) form() url.Values {
	v := url.Values{}
	if p.Body != nil {
		v.Set("Body", *p.Body)
	}
	if p.Status != nil {
		v.Set("Status", *p.Status)
	}
	return v
}

// Update mutates an existing Message — redact Body or attempt cancel.
func (s *MessagesService) Update(ctx context.Context, sid string, params UpdateMessageParams) (*Message, error) {
	var out Message
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Messages", sid),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a Message resource from the account's store.
func (s *MessagesService) Delete(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("Messages", sid),
	}, nil)
}
