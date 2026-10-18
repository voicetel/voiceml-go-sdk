package voiceml

import (
	"context"
	"net/url"
)

// TranscriptionsService is the account-scoped /Transcriptions resource
// group: post-recording transcriptions created by <Record transcribe="true">.
// They are read and delete only. Recording-scoped access lives on
// RecordingsService; live per-call transcription sessions are
// CallsService.StartTranscription / StopTranscription.
type TranscriptionsService struct{ c *Client }

// Transcription is a post-recording transcription (TR...). Status is
// "in-progress", "completed" or "failed"; TranscriptionText is nil until
// Status reaches "completed". Duration is whole seconds as a string.
type Transcription struct {
	Sid               string  `json:"sid"`
	AccountSid        string  `json:"account_sid"`
	Status            string  `json:"status"`
	RecordingSid      string  `json:"recording_sid"`
	Duration          string  `json:"duration,omitempty"`
	TranscriptionText *string `json:"transcription_text"`
	Type              string  `json:"type"`
	Price             string  `json:"price,omitempty"`
	PriceUnit         string  `json:"price_unit,omitempty"`
	APIVersion        string  `json:"api_version"`
	URI               string  `json:"uri"`
	DateCreated       string  `json:"date_created,omitempty"`
	DateUpdated       string  `json:"date_updated,omitempty"`
}

// TranscriptionList is the paginated /Transcriptions and
// /Recordings/{sid}/Transcriptions response.
type TranscriptionList struct {
	Page
	Transcriptions []Transcription `json:"transcriptions"`
}

// ListTranscriptionsParams are the filter / pagination query params for the
// transcription lists. DateCreated is a YYYY-MM-DD date (see Date).
type ListTranscriptionsParams struct {
	DateCreated   string
	DateCreatedLt string
	DateCreatedGt string
	Page          *int
	PageSize      *int
	PageToken     string
}

func (p ListTranscriptionsParams) query() url.Values {
	v := url.Values{}
	setString(v, "DateCreated", p.DateCreated)
	setString(v, "DateCreated<", p.DateCreatedLt)
	setString(v, "DateCreated>", p.DateCreatedGt)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	setString(v, "PageToken", p.PageToken)
	return v
}

// List returns a page of the account's transcriptions. GET /Transcriptions.
func (s *TranscriptionsService) List(ctx context.Context, params ListTranscriptionsParams) (*TranscriptionList, error) {
	var out TranscriptionList
	if err := s.c.t.do(ctx, requestOpts{method: "GET", path: s.c.pathf("Transcriptions"), query: params.query()}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Iterate walks every transcription on the account lazily.
func (s *TranscriptionsService) Iterate(ctx context.Context, params ListTranscriptionsParams) *Pager[Transcription] {
	return newPager(ctx, s.c, s.c.t, requestOpts{method: "GET", path: s.c.pathf("Transcriptions"), query: params.query()},
		func(l *TranscriptionList) ([]Transcription, PageLinks) { return l.Transcriptions, l.Page })
}

// Fetch returns one transcription. GET /Transcriptions/{Sid}.
func (s *TranscriptionsService) Fetch(ctx context.Context, sid string) (*Transcription, error) {
	var out Transcription
	if err := s.c.t.do(ctx, requestOpts{method: "GET", path: s.c.pathf("Transcriptions", sid)}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a transcription. DELETE /Transcriptions/{Sid}.
func (s *TranscriptionsService) Delete(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{method: "DELETE", path: s.c.pathf("Transcriptions", sid)}, nil)
}
