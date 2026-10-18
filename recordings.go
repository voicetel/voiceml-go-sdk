package voiceml

import (
	"context"
	"fmt"
	"net/url"
)

// RecordingsService is the account-scoped /Recordings resource group plus
// recording-scoped transcriptions. Per-call recording start/stop/list lives
// on CallsService, per-conference on ConferencesService.
type RecordingsService struct{ c *Client }

// Recording is a Twilio-compatible Recording resource. Fields populated
// depend on the endpoint that produced it.
type Recording struct {
	Sid               string          `json:"sid"`
	AccountSid        string          `json:"account_sid"`
	CallSid           string          `json:"call_sid"`
	ConferenceSid     string          `json:"conference_sid,omitempty"`
	Status            RecordingStatus `json:"status"`
	Source            RecordingSource `json:"source,omitempty"`
	Channels          *int            `json:"channels,omitempty"`
	Duration          string          `json:"duration,omitempty"`
	Track             string          `json:"track,omitempty"`
	APIVersion        string          `json:"api_version,omitempty"`
	URI               string          `json:"uri,omitempty"`
	MediaURL          string          `json:"media_url,omitempty"`
	ErrorCode         *int            `json:"error_code"`
	DateCreated       string          `json:"date_created,omitempty"`
	DateUpdated       string          `json:"date_updated,omitempty"`
	StartTime         string          `json:"start_time,omitempty"`
	Price             string          `json:"price,omitempty"`
	PriceUnit         string          `json:"price_unit,omitempty"`
	EncryptionDetails map[string]any  `json:"encryption_details,omitempty"`
	SubresourceURIs   map[string]any  `json:"subresource_uris,omitempty"`
}

// RecordingList is the paginated list response for /Recordings
// (account-scoped) and the per-call / per-conference list endpoints.
type RecordingList struct {
	Page
	Recordings []Recording `json:"recordings"`
}

// StartRecordingParams is the body for POST /Calls/{sid}/Recordings.
// RecordingTrack is "inbound", "outbound" or "both".
type StartRecordingParams struct {
	RecordingMaxDuration          *int    `form:"RecordingMaxDuration"`
	RecordingChannels             *string `form:"RecordingChannels"`
	RecordingTrack                *string `form:"RecordingTrack"`
	PlayBeep                      *bool   `form:"PlayBeep"`
	RecordingStatusCallback       *string `form:"RecordingStatusCallback"`
	RecordingStatusCallbackMethod *string `form:"RecordingStatusCallbackMethod"`
	RecordingStatusCallbackEvent  *string `form:"RecordingStatusCallbackEvent"`
}

func (p StartRecordingParams) form() url.Values {
	v := url.Values{}
	setIntP(v, "RecordingMaxDuration", p.RecordingMaxDuration)
	setStringP(v, "RecordingChannels", p.RecordingChannels)
	setStringP(v, "RecordingTrack", p.RecordingTrack)
	setBoolP(v, "PlayBeep", p.PlayBeep)
	setStringP(v, "RecordingStatusCallback", p.RecordingStatusCallback)
	setStringP(v, "RecordingStatusCallbackMethod", p.RecordingStatusCallbackMethod)
	setStringP(v, "RecordingStatusCallbackEvent", p.RecordingStatusCallbackEvent)
	return v
}

// UpdateCallRecordingParams is the body for POST
// /Calls/{sid}/Recordings/{rsid}. Status takes "stopped", "paused", or
// "in-progress" (resume after pause). PauseBehavior applies with
// Status="paused": "skip" leaves the paused span out, "silence" (the
// default) keeps it as silence.
type UpdateCallRecordingParams struct {
	Status        string  `form:"Status"`
	PauseBehavior *string `form:"PauseBehavior"`
}

func (p UpdateCallRecordingParams) form() url.Values {
	v := url.Values{}
	v.Set("Status", p.Status)
	setStringP(v, "PauseBehavior", p.PauseBehavior)
	return v
}

// UpdateRecordingParams is the body for POST
// /Conferences/{sid}/Recordings/{rsid}. Status takes "stopped", "paused",
// or "in-progress" (resume after pause).
type UpdateRecordingParams struct {
	Status string `form:"Status"`
}

func (p UpdateRecordingParams) form() url.Values {
	v := url.Values{}
	v.Set("Status", p.Status)
	return v
}

// GetRecordingParams are optional query params for GET /Recordings/{sid}.
type GetRecordingParams struct {
	IncludeSoftDeleted *bool
}

func (p GetRecordingParams) query() url.Values {
	v := url.Values{}
	setBoolP(v, "IncludeSoftDeleted", p.IncludeSoftDeleted)
	return v
}

// ListRecordingsParams are the query params for GET /Recordings (account-scoped).
type ListRecordingsParams struct {
	DateCreated        string
	DateCreatedLt      string
	DateCreatedGt      string
	CallSid            string
	ConferenceSid      string
	IncludeSoftDeleted *bool
	Page               *int
	PageSize           *int
	PageToken          string
}

func (p ListRecordingsParams) query() url.Values {
	v := url.Values{}
	setString(v, "DateCreated", p.DateCreated)
	setString(v, "DateCreated<", p.DateCreatedLt)
	setString(v, "DateCreated>", p.DateCreatedGt)
	setString(v, "CallSid", p.CallSid)
	setString(v, "ConferenceSid", p.ConferenceSid)
	setBoolP(v, "IncludeSoftDeleted", p.IncludeSoftDeleted)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	setString(v, "PageToken", p.PageToken)
	return v
}

// ListCallRecordingsParams are the query params for GET
// /Calls/{sid}/Recordings and GET /Conferences/{sid}/Recordings: the
// DateCreated filters plus pagination. DateCreated is a YYYY-MM-DD date.
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

// List returns a page of recordings for this account. GET /Recordings.
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

// Iterate walks every recording matching the filters lazily, following each
// page's next_page_uri. See Pager.
func (s *RecordingsService) Iterate(ctx context.Context, params ListRecordingsParams) *Pager[Recording] {
	return newPager(ctx, s.c, s.c.t, requestOpts{method: "GET", path: s.c.pathf("Recordings"), query: params.query()},
		func(l *RecordingList) ([]Recording, PageLinks) { return l.Recordings, l.Page })
}

// Get fetches a recording's metadata. GET /Recordings/{sid}.
func (s *RecordingsService) Get(ctx context.Context, recordingSid string, params *GetRecordingParams) (*Recording, error) {
	var q url.Values
	if params != nil {
		q = params.query()
	}
	var out Recording
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Recordings", recordingSid),
		query:  q,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// maxAudioBytes bounds GetAudio's in-memory read; use DownloadAudio to
// stream larger files.
const maxAudioBytes = 1 << 30

// DownloadAudio streams the WAV audio for a recording
// (GET /Recordings/{sid}.wav). For archived recordings the server answers
// 302 to a presigned storage URL; the SDK follows it without sending your
// credentials (Authorization is dropped once the redirect leaves the API
// host) and refuses https→http redirects. On 410 Gone the audio is
// unavailable and an *APIError wrapping ErrGone is returned.
//
// The caller must Close the returned Download's Body.
func (s *RecordingsService) DownloadAudio(ctx context.Context, recordingSid string) (*Download, error) {
	return s.c.t.download(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathfExt(".wav", "Recordings", recordingSid),
		accept: "audio/*, */*",
	})
}

// GetAudio is DownloadAudio read fully into memory. It returns the bytes
// and the declared Content-Type (typically "audio/wav"), and fails for
// files over 1 GiB.
func (s *RecordingsService) GetAudio(ctx context.Context, recordingSid string) ([]byte, string, error) {
	d, err := s.DownloadAudio(ctx, recordingSid)
	if err != nil {
		return nil, "", err
	}
	defer d.Body.Close()
	b, err := readAtMost(d.Body, maxAudioBytes)
	if err != nil {
		return nil, "", transportError(fmt.Errorf("read recording audio (use DownloadAudio to stream large files): %w", err), 1)
	}
	return b, d.ContentType, nil
}

// Delete removes a recording. DELETE /Recordings/{sid}.
func (s *RecordingsService) Delete(ctx context.Context, recordingSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("Recordings", recordingSid),
	}, nil)
}

// --- Recording-scoped transcriptions ---

// ListTranscriptions returns a page of the post-recording transcriptions of
// a recording. GET /Recordings/{RecordingSid}/Transcriptions.
func (s *RecordingsService) ListTranscriptions(ctx context.Context, recordingSid string, params ListTranscriptionsParams) (*TranscriptionList, error) {
	var out TranscriptionList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Recordings", recordingSid, "Transcriptions"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// IterateTranscriptions walks every transcription of a recording lazily.
func (s *RecordingsService) IterateTranscriptions(ctx context.Context, recordingSid string, params ListTranscriptionsParams) *Pager[Transcription] {
	return newPager(ctx, s.c, s.c.t, requestOpts{method: "GET", path: s.c.pathf("Recordings", recordingSid, "Transcriptions"), query: params.query()},
		func(l *TranscriptionList) ([]Transcription, PageLinks) { return l.Transcriptions, l.Page })
}

// FetchTranscription fetches one transcription of a recording.
// GET /Recordings/{RecordingSid}/Transcriptions/{Sid}.
func (s *RecordingsService) FetchTranscription(ctx context.Context, recordingSid, transcriptionSid string) (*Transcription, error) {
	var out Transcription
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Recordings", recordingSid, "Transcriptions", transcriptionSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTranscription deletes one transcription of a recording.
// DELETE /Recordings/{RecordingSid}/Transcriptions/{Sid}.
func (s *RecordingsService) DeleteTranscription(ctx context.Context, recordingSid, transcriptionSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("Recordings", recordingSid, "Transcriptions", transcriptionSid),
	}, nil)
}
