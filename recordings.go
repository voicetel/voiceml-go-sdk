package voiceml

import (
	"context"
	"net/url"
)

// RecordingsService is the account-scoped /Recordings resource group.
// Per-call recording start/stop/list lives on CallsService.
type RecordingsService struct{ c *Client }

// Recording is a Twilio-shape Recording resource. Fields populated depend on
// the endpoint that produced it — call-scoped and conference-scoped lists may
// omit pagination fields.
type Recording struct {
	Sid               string         `json:"sid"`
	AccountSid        string         `json:"account_sid"`
	CallSid           string         `json:"call_sid"`
	ConferenceSid     string         `json:"conference_sid,omitempty"`
	Status            string         `json:"status"`
	Source            string         `json:"source,omitempty"`
	Channels          *int           `json:"channels,omitempty"`
	Duration          string         `json:"duration,omitempty"`
	APIVersion        string         `json:"api_version,omitempty"`
	URI               string         `json:"uri,omitempty"`
	MediaURL          string         `json:"media_url,omitempty"`
	ErrorCode         *int           `json:"error_code"`
	DateCreated       string         `json:"date_created,omitempty"`
	DateUpdated       string         `json:"date_updated,omitempty"`
	StartTime         string         `json:"start_time,omitempty"`
	Price             string         `json:"price,omitempty"`
	PriceUnit         string         `json:"price_unit,omitempty"`
	EncryptionDetails map[string]any `json:"encryption_details,omitempty"`
	SubresourceURIs   map[string]any `json:"subresource_uris,omitempty"`
}

// RecordingList is the list response for /Recordings (account-scoped) and the
// per-call / per-conference list endpoints. The latter two currently return
// only the Recordings slice — the pagination fields will be zero.
type RecordingList struct {
	Recordings      []Recording `json:"recordings"`
	Page            *int        `json:"page,omitempty"`
	PageSize        *int        `json:"page_size,omitempty"`
	Total           *int        `json:"total,omitempty"`
	NumPages        *int        `json:"num_pages,omitempty"`
	FirstPageURI    string      `json:"first_page_uri,omitempty"`
	NextPageURI     string      `json:"next_page_uri,omitempty"`
	PreviousPageURI string      `json:"previous_page_uri,omitempty"`
	URI             string      `json:"uri,omitempty"`
}

// StartRecordingParams is the body for POST /Calls/{sid}/Recordings.
type StartRecordingParams struct {
	RecordingMaxDuration          *int    `form:"RecordingMaxDuration"`
	RecordingChannels             *string `form:"RecordingChannels"`
	PlayBeep                      *bool   `form:"PlayBeep"`
	RecordingStatusCallback       *string `form:"RecordingStatusCallback"`
	RecordingStatusCallbackMethod *string `form:"RecordingStatusCallbackMethod"`
	RecordingStatusCallbackEvent  *string `form:"RecordingStatusCallbackEvent"`
}

func (p StartRecordingParams) form() url.Values {
	v := url.Values{}
	setIntP(v, "RecordingMaxDuration", p.RecordingMaxDuration)
	setStringP(v, "RecordingChannels", p.RecordingChannels)
	setBoolP(v, "PlayBeep", p.PlayBeep)
	setStringP(v, "RecordingStatusCallback", p.RecordingStatusCallback)
	setStringP(v, "RecordingStatusCallbackMethod", p.RecordingStatusCallbackMethod)
	setStringP(v, "RecordingStatusCallbackEvent", p.RecordingStatusCallbackEvent)
	return v
}

// UpdateRecordingParams is the body for POST /Calls/{sid}/Recordings/{rsid}.
// Status takes "stopped", "paused", or "in-progress" (resume after pause).
type UpdateRecordingParams struct {
	Status string `form:"Status"`
}

func (p UpdateRecordingParams) form() url.Values {
	v := url.Values{}
	v.Set("Status", p.Status)
	return v
}

// ListRecordingsParams are the query params for GET /Recordings (account-scoped).
type ListRecordingsParams struct {
	DateCreated   string
	DateCreatedLt string
	DateCreatedGt string
	CallSid       string
	ConferenceSid string
	Page          *int
	PageSize      *int
	PageToken     string
}

func (p ListRecordingsParams) query() url.Values {
	v := url.Values{}
	setString(v, "DateCreated", p.DateCreated)
	setString(v, "DateCreated<", p.DateCreatedLt)
	setString(v, "DateCreated>", p.DateCreatedGt)
	setString(v, "CallSid", p.CallSid)
	setString(v, "ConferenceSid", p.ConferenceSid)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	setString(v, "PageToken", p.PageToken)
	return v
}

// ListCallRecordingsParams are the query params for GET /Calls/{sid}/Recordings.
type ListCallRecordingsParams struct {
	DateCreated   string
	DateCreatedLt string
	DateCreatedGt string
	Page          *int
	PageSize      *int
	PageToken     string
}

func (p ListCallRecordingsParams) query() url.Values {
	v := url.Values{}
	setString(v, "DateCreated", p.DateCreated)
	setString(v, "DateCreated<", p.DateCreatedLt)
	setString(v, "DateCreated>", p.DateCreatedGt)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	setString(v, "PageToken", p.PageToken)
	return v
}

// List returns recordings for this account. GET /Recordings.
func (s *RecordingsService) List(ctx context.Context, params ListRecordingsParams) (*RecordingList, error) {
	var out RecordingList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Recordings"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a recording's metadata. GET /Recordings/{sid}.
func (s *RecordingsService) Get(ctx context.Context, recordingSid string) (*Recording, error) {
	var out Recording
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Recordings", recordingSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAudio fetches the WAV audio for a recording, transparently following the
// 302 → S3 presigned URL the server emits for archived recordings.
//
// Returns the bytes and the Content-Type the server (or S3) declared
// (typically "audio/wav"). On 410 Gone the audio is unavailable and an
// *APIError wrapping ErrGone is returned.
func (s *RecordingsService) GetAudio(ctx context.Context, recordingSid string) ([]byte, string, error) {
	return s.c.t.fetchBytes(ctx, s.c.pathfExt(".wav", "Recordings", recordingSid))
}

// Delete removes a recording. DELETE /Recordings/{sid}.
func (s *RecordingsService) Delete(ctx context.Context, recordingSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("Recordings", recordingSid),
	}, nil)
}
