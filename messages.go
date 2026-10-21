package voiceml

import (
	"context"
	"net/url"
)

// MessagesService surfaces the /Messages REST resource — the
// Twilio-compatible SMS/MMS surface — plus each message's Media
// sub-resource.
type MessagesService struct{ c *Client }

// Message is a Twilio-compatible Message resource. Status is one of the
// Twilio-compatible message states ("queued", "sent", "delivered",
// "failed", ...); treat unknown values as opaque.
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

// CreateMessageParams is the body for POST /Messages. To is required, and
// so is at least one of Body or MediaURL (a message with MediaURL is sent
// as MMS). From falls back to the server's default sender when omitted.
// MediaURL entries are sent as repeated MediaUrl parameters. ValidityPeriod
// is in seconds; StatusCallbackMethod is "GET" or "POST". SendAt schedules
// the message for a future UTC time (RFC 3339 — see Time); it requires
// ScheduleType="fixed" (the only supported value) and MessagingServiceSid.
type CreateMessageParams struct {
	To                   string   `form:"To"`
	Body                 *string  `form:"Body"`
	MediaURL             []string `form:"MediaUrl"`
	From                 *string  `form:"From"`
	MessagingServiceSid  *string  `form:"MessagingServiceSid"`
	ValidityPeriod       *int     `form:"ValidityPeriod"`
	StatusCallback       *string  `form:"StatusCallback"`
	StatusCallbackMethod *string  `form:"StatusCallbackMethod"`
	SendAt               *string  `form:"SendAt"`
	ScheduleType         *string  `form:"ScheduleType"`
}

func (p CreateMessageParams) form() url.Values {
	v := url.Values{}
	v.Set("To", p.To)
	setStringP(v, "From", p.From)
	setStringP(v, "Body", p.Body)
	for _, u := range p.MediaURL {
		v.Add("MediaUrl", u)
	}
	setStringP(v, "MessagingServiceSid", p.MessagingServiceSid)
	setIntP(v, "ValidityPeriod", p.ValidityPeriod)
	setStringP(v, "StatusCallback", p.StatusCallback)
	setStringP(v, "StatusCallbackMethod", p.StatusCallbackMethod)
	setStringP(v, "SendAt", p.SendAt)
	setStringP(v, "ScheduleType", p.ScheduleType)
	return v
}

func (p CreateMessageParams) validate() error {
	if p.Body == nil && len(p.MediaURL) == 0 {
		return &ValidationError{Message: "CreateMessageParams needs Body or MediaURL"}
	}
	return nil
}

// Create dispatches an outbound SMS, or MMS when MediaURL is set.
// POST /Messages. It returns *ValidationError, without sending, when
// neither Body nor MediaURL is set.
func (s *MessagesService) Create(ctx context.Context, params CreateMessageParams) (*Message, error) {
	if err := params.validate(); err != nil {
		return nil, err
	}
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

// ListMessagesParams narrows GET /Messages — the Twilio-compatible filter
// set (To, From, DateSent eq/gt/lt) plus pagination.
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

// Iterate walks every message matching the filters lazily, following each
// page's next_page_uri. See Pager.
func (s *MessagesService) Iterate(ctx context.Context, params ListMessagesParams) *Pager[Message] {
	return newPager(ctx, s.c, s.c.t, requestOpts{method: "GET", path: s.c.pathf("Messages"), query: params.query()},
		func(l *MessageList) ([]Message, PageLinks) { return l.Messages, l.Page })
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

// UpdateMessageParams is the body for POST /Messages/{Sid}: Body=""
// redacts the message body; Status="canceled" cancels a message that has
// not been sent yet.
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

// --- Feedback ---

// MessageFeedback is the delivery-quality feedback echoed back for a
// Message. Outcome is "confirmed" or "unconfirmed"; treat unknown values
// as opaque.
type MessageFeedback struct {
	AccountSid  string `json:"account_sid"`
	MessageSid  string `json:"message_sid"`
	Outcome     string `json:"outcome"`
	DateCreated string `json:"date_created"`
	DateUpdated string `json:"date_updated"`
	URI         string `json:"uri"`
}

// CreateMessageFeedbackParams is the body for POST
// /Messages/{Sid}/Feedback.json. Outcome is "confirmed" or "unconfirmed";
// any other value (or none) is treated as "unconfirmed" by the server.
type CreateMessageFeedbackParams struct {
	Outcome *string `form:"Outcome"`
}

func (p CreateMessageFeedbackParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "Outcome", p.Outcome)
	return v
}

// CreateFeedback submits delivery feedback for a previously-sent Message
// and returns the echoed Feedback resource.
// POST /Messages/{Sid}/Feedback.json.
func (s *MessagesService) CreateFeedback(ctx context.Context, messageSid string, params CreateMessageFeedbackParams) (*MessageFeedback, error) {
	var out MessageFeedback
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Messages", messageSid, "Feedback"),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Media ---

// MessageMedia is one media item attached to a message (ME...). ParentSid
// is the message SID; ContentType is nil when unknown at send time.
type MessageMedia struct {
	Sid         string  `json:"sid"`
	AccountSid  string  `json:"account_sid"`
	ParentSid   string  `json:"parent_sid"`
	ContentType *string `json:"content_type"`
	DateCreated string  `json:"date_created"`
	DateUpdated string  `json:"date_updated"`
	URI         string  `json:"uri"`
}

// MessageMediaList is the paginated /Messages/{sid}/Media response.
type MessageMediaList struct {
	Page
	MediaList []MessageMedia `json:"media_list"`
}

// ListMediaParams are the filter / pagination query params for
// GET /Messages/{sid}/Media. DateCreated is a YYYY-MM-DD date (see Date).
type ListMediaParams struct {
	DateCreated   string
	DateCreatedLt string
	DateCreatedGt string
	Page          *int
	PageSize      *int
	PageToken     string
}

func (p ListMediaParams) query() url.Values {
	v := url.Values{}
	setString(v, "DateCreated", p.DateCreated)
	setString(v, "DateCreated<", p.DateCreatedLt)
	setString(v, "DateCreated>", p.DateCreatedGt)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	setString(v, "PageToken", p.PageToken)
	return v
}

// ListMedia returns a page of a message's media. GET /Messages/{sid}/Media.
func (s *MessagesService) ListMedia(ctx context.Context, messageSid string, params ListMediaParams) (*MessageMediaList, error) {
	var out MessageMediaList
	if err := s.c.t.do(ctx, requestOpts{method: "GET", path: s.c.pathf("Messages", messageSid, "Media"), query: params.query()}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IterateMedia walks every media item of a message lazily.
func (s *MessagesService) IterateMedia(ctx context.Context, messageSid string, params ListMediaParams) *Pager[MessageMedia] {
	return newPager(ctx, s.c, s.c.t, requestOpts{method: "GET", path: s.c.pathf("Messages", messageSid, "Media"), query: params.query()},
		func(l *MessageMediaList) ([]MessageMedia, PageLinks) { return l.MediaList, l.Page })
}

// FetchMedia returns one media item's metadata.
// GET /Messages/{sid}/Media/{Sid}.
func (s *MessagesService) FetchMedia(ctx context.Context, messageSid, mediaSid string) (*MessageMedia, error) {
	var out MessageMedia
	if err := s.c.t.do(ctx, requestOpts{method: "GET", path: s.c.pathf("Messages", messageSid, "Media", mediaSid)}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMedia removes a media item from a message.
// DELETE /Messages/{sid}/Media/{Sid}.
func (s *MessagesService) DeleteMedia(ctx context.Context, messageSid, mediaSid string) error {
	return s.c.t.do(ctx, requestOpts{method: "DELETE", path: s.c.pathf("Messages", messageSid, "Media", mediaSid)}, nil)
}
