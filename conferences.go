package voiceml

import (
	"context"
	"net/url"
)

// ConferencesService is the /Conferences resource group plus participants
// and conference-scoped recordings.
type ConferencesService struct{ c *Client }

// Conference is a Twilio-shape Conference resource.
type Conference struct {
	Sid                     string            `json:"sid"`
	AccountSid              string            `json:"account_sid"`
	FriendlyName            string            `json:"friendly_name"`
	Status                  string            `json:"status"`
	Region                  string            `json:"region,omitempty"`
	APIVersion              string            `json:"api_version"`
	URI                     string            `json:"uri"`
	DateCreated             string            `json:"date_created,omitempty"`
	DateUpdated             string            `json:"date_updated,omitempty"`
	ReasonConferenceEnded   string            `json:"reason_conference_ended,omitempty"`
	CallSidEndingConference string            `json:"call_sid_ending_conference,omitempty"`
	SubresourceURIs         map[string]string `json:"subresource_uris,omitempty"`
	MemberCount             *int              `json:"member_count,omitempty"`
}

// ConferenceList is the paginated /Conferences list response.
type ConferenceList struct {
	Page
	Conferences []Conference `json:"conferences"`
}

// Participant is a single leg in a conference.
type Participant struct {
	CallSid                string `json:"call_sid"`
	ConferenceSid          string `json:"conference_sid"`
	AccountSid             string `json:"account_sid"`
	Muted                  bool   `json:"muted"`
	Hold                   bool   `json:"hold"`
	Coaching               bool   `json:"coaching"`
	CallSidToCoach         string `json:"call_sid_to_coach,omitempty"`
	QueueTime              string `json:"queue_time"`
	StartConferenceOnEnter bool   `json:"start_conference_on_enter"`
	EndConferenceOnExit    bool   `json:"end_conference_on_exit"`
	Status                 string `json:"status"`
	Label                  string `json:"label,omitempty"`
	APIVersion             string `json:"api_version"`
	URI                    string `json:"uri"`
	DateCreated            string `json:"date_created,omitempty"`
	DateUpdated            string `json:"date_updated,omitempty"`
}

// ParticipantList is the paginated /Conferences/{sid}/Participants list response.
type ParticipantList struct {
	Page
	Participants []Participant `json:"participants"`
}

// EndConferenceParams is the body for POST /Conferences/{sid}. v1 supports
// only Status="completed".
type EndConferenceParams struct {
	Status string `form:"Status"`
}

func (p EndConferenceParams) form() url.Values {
	v := url.Values{}
	v.Set("Status", p.Status)
	return v
}

// UpdateParticipantParams is the body for POST
// /Conferences/{sid}/Participants/{call_sid}. At least one of Muted / Hold
// must be set.
type UpdateParticipantParams struct {
	Muted *bool `form:"Muted"`
	Hold  *bool `form:"Hold"`
}

func (p UpdateParticipantParams) form() url.Values {
	v := url.Values{}
	setBoolP(v, "Muted", p.Muted)
	setBoolP(v, "Hold", p.Hold)
	return v
}

// ListConferencesParams are the filter / pagination query params for GET /Conferences.
type ListConferencesParams struct {
	FriendlyName string
	Status       string
	Page         *int
	PageSize     *int
}

func (p ListConferencesParams) query() url.Values {
	v := url.Values{}
	setString(v, "FriendlyName", p.FriendlyName)
	setString(v, "Status", p.Status)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	return v
}

// ListParticipantsParams are the filter / pagination query params for
// GET /Conferences/{sid}/Participants.
type ListParticipantsParams struct {
	Muted    *bool
	Hold     *bool
	Coaching *bool
	Page     *int
	PageSize *int
}

func (p ListParticipantsParams) query() url.Values {
	v := url.Values{}
	setBoolP(v, "Muted", p.Muted)
	setBoolP(v, "Hold", p.Hold)
	setBoolP(v, "Coaching", p.Coaching)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	return v
}

// List returns all conferences for this account. GET /Conferences.
func (s *ConferencesService) List(ctx context.Context, params ListConferencesParams) (*ConferenceList, error) {
	var out ConferenceList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Conferences"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a conference by SID. GET /Conferences/{sid}.
func (s *ConferencesService) Get(ctx context.Context, conferenceSid string) (*Conference, error) {
	var out Conference
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Conferences", conferenceSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// End terminates a conference. POST /Conferences/{sid}. Pass nil to default
// to Status="completed".
func (s *ConferencesService) End(ctx context.Context, conferenceSid string, params *EndConferenceParams) (*Conference, error) {
	body := EndConferenceParams{Status: "completed"}
	if params != nil {
		body = *params
	}
	var out Conference
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Conferences", conferenceSid),
		form:   body.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListParticipants returns the legs currently in a conference.
// GET /Conferences/{sid}/Participants.
func (s *ConferencesService) ListParticipants(ctx context.Context, conferenceSid string, params ListParticipantsParams) (*ParticipantList, error) {
	var out ParticipantList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Conferences", conferenceSid, "Participants"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetParticipant fetches a single participant by its call SID.
// GET /Conferences/{sid}/Participants/{call_sid}.
func (s *ConferencesService) GetParticipant(ctx context.Context, conferenceSid, callSid string) (*Participant, error) {
	var out Participant
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Conferences", conferenceSid, "Participants", callSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateParticipant mutes/unmutes or holds/unholds a participant.
// POST /Conferences/{sid}/Participants/{call_sid}.
func (s *ConferencesService) UpdateParticipant(ctx context.Context, conferenceSid, callSid string, params UpdateParticipantParams) (*Participant, error) {
	var out Participant
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Conferences", conferenceSid, "Participants", callSid),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// KickParticipant ejects a participant from a conference.
// DELETE /Conferences/{sid}/Participants/{call_sid}.
func (s *ConferencesService) KickParticipant(ctx context.Context, conferenceSid, callSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("Conferences", conferenceSid, "Participants", callSid),
	}, nil)
}

// ListRecordings returns recordings made of this conference.
// GET /Conferences/{sid}/Recordings.
func (s *ConferencesService) ListRecordings(ctx context.Context, conferenceSid string) (*RecordingList, error) {
	var out RecordingList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Conferences", conferenceSid, "Recordings"),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
