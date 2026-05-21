package voiceml

import (
	"context"
	"net/url"
	"strconv"
)

// CallsService is the /Calls resource group plus per-call sub-resources
// (Recordings, Streams, Siprec, Transcriptions, Notifications, Events,
// UserDefinedMessages).
type CallsService struct{ c *Client }

// Call is a Twilio-shape Call resource.
type Call struct {
	Sid             string            `json:"sid"`
	AccountSid      string            `json:"account_sid"`
	APIVersion      string            `json:"api_version"`
	To              string            `json:"to,omitempty"`
	ToFormatted     string            `json:"to_formatted,omitempty"`
	From            string            `json:"from,omitempty"`
	FromFormatted   string            `json:"from_formatted,omitempty"`
	ParentCallSid   string            `json:"parent_call_sid,omitempty"`
	CallerName      string            `json:"caller_name,omitempty"`
	ForwardedFrom   string            `json:"forwarded_from,omitempty"`
	Status          string            `json:"status"`
	Direction       string            `json:"direction"`
	AnsweredBy      string            `json:"answered_by,omitempty"`
	StartTime       string            `json:"start_time,omitempty"`
	EndTime         string            `json:"end_time,omitempty"`
	Duration        string            `json:"duration,omitempty"`
	Price           string            `json:"price,omitempty"`
	PriceUnit       string            `json:"price_unit,omitempty"`
	PhoneNumberSid  string            `json:"phone_number_sid,omitempty"`
	Annotation      string            `json:"annotation,omitempty"`
	GroupSid        string            `json:"group_sid,omitempty"`
	QueueTime       string            `json:"queue_time,omitempty"`
	TrunkSid        string            `json:"trunk_sid,omitempty"`
	DateCreated     string            `json:"date_created"`
	DateUpdated     string            `json:"date_updated"`
	URI             string            `json:"uri"`
	SubresourceURIs map[string]string `json:"subresource_uris,omitempty"`
}

// CallList is the paginated /Calls list response.
type CallList struct {
	Page
	Calls []Call `json:"calls"`
}

// CreateCallParams is the body for POST /Calls. To and From are required;
// set exactly one of URL / Twiml / ApplicationSid (Twiml wins if multiple
// are set — Twilio's documented precedence).
type CreateCallParams struct {
	To   string `form:"To"`
	From string `form:"From"`

	URL                                *string  `form:"Url"`
	Method                             *string  `form:"Method"`
	Twiml                              *string  `form:"Twiml"`
	ApplicationSid                     *string  `form:"ApplicationSid"`
	FallbackURL                        *string  `form:"FallbackUrl"`
	FallbackMethod                     *string  `form:"FallbackMethod"`
	StatusCallback                     *string  `form:"StatusCallback"`
	StatusCallbackMethod               *string  `form:"StatusCallbackMethod"`
	StatusCallbackEvent                []string `form:"StatusCallbackEvent"`
	MachineDetection                   *string  `form:"MachineDetection"`
	MachineDetectionTimeout            *int     `form:"MachineDetectionTimeout"`
	MachineDetectionSpeechThreshold    *int     `form:"MachineDetectionSpeechThreshold"`
	MachineDetectionSpeechEndThreshold *int     `form:"MachineDetectionSpeechEndThreshold"`
	MachineDetectionSilenceTimeout     *int     `form:"MachineDetectionSilenceTimeout"`
	AsyncAmdStatusCallback             *string  `form:"AsyncAmdStatusCallback"`
	AsyncAmdStatusCallbackMethod       *string  `form:"AsyncAmdStatusCallbackMethod"`
	Record                             *bool    `form:"Record"`
	RecordingStatusCallback            *string  `form:"RecordingStatusCallback"`
	RecordingStatusCallbackMethod      *string  `form:"RecordingStatusCallbackMethod"`
	RecordingStatusCallbackEvent       *string  `form:"RecordingStatusCallbackEvent"`
	RecordingChannels                  *string  `form:"RecordingChannels"`
	RecordingTrack                     *string  `form:"RecordingTrack"`
	Trim                               *string  `form:"Trim"`
	Timeout                            *int     `form:"Timeout"`
	SendDigits                         *string  `form:"SendDigits"`
	CallerID                           *string  `form:"CallerId"`
	CallReason                         *string  `form:"CallReason"`
	SipAuthUsername                    *string  `form:"SipAuthUsername"`
	SipAuthPassword                    *string  `form:"SipAuthPassword"`
	Byoc                               *string  `form:"Byoc"`
	AsyncAmd                           *bool    `form:"AsyncAmd"`
	CallToken                          *string  `form:"CallToken"`
}

func (p CreateCallParams) form() url.Values {
	v := url.Values{}
	v.Set("To", p.To)
	v.Set("From", p.From)
	setStringP(v, "Url", p.URL)
	setStringP(v, "Method", p.Method)
	setStringP(v, "Twiml", p.Twiml)
	setStringP(v, "ApplicationSid", p.ApplicationSid)
	setStringP(v, "FallbackUrl", p.FallbackURL)
	setStringP(v, "FallbackMethod", p.FallbackMethod)
	setStringP(v, "StatusCallback", p.StatusCallback)
	setStringP(v, "StatusCallbackMethod", p.StatusCallbackMethod)
	for _, s := range p.StatusCallbackEvent {
		v.Add("StatusCallbackEvent", s)
	}
	setStringP(v, "MachineDetection", p.MachineDetection)
	setIntP(v, "MachineDetectionTimeout", p.MachineDetectionTimeout)
	setIntP(v, "MachineDetectionSpeechThreshold", p.MachineDetectionSpeechThreshold)
	setIntP(v, "MachineDetectionSpeechEndThreshold", p.MachineDetectionSpeechEndThreshold)
	setIntP(v, "MachineDetectionSilenceTimeout", p.MachineDetectionSilenceTimeout)
	setStringP(v, "AsyncAmdStatusCallback", p.AsyncAmdStatusCallback)
	setStringP(v, "AsyncAmdStatusCallbackMethod", p.AsyncAmdStatusCallbackMethod)
	setBoolP(v, "Record", p.Record)
	setStringP(v, "RecordingStatusCallback", p.RecordingStatusCallback)
	setStringP(v, "RecordingStatusCallbackMethod", p.RecordingStatusCallbackMethod)
	setStringP(v, "RecordingStatusCallbackEvent", p.RecordingStatusCallbackEvent)
	setStringP(v, "RecordingChannels", p.RecordingChannels)
	setStringP(v, "RecordingTrack", p.RecordingTrack)
	setStringP(v, "Trim", p.Trim)
	setIntP(v, "Timeout", p.Timeout)
	setStringP(v, "SendDigits", p.SendDigits)
	setStringP(v, "CallerId", p.CallerID)
	setStringP(v, "CallReason", p.CallReason)
	setStringP(v, "SipAuthUsername", p.SipAuthUsername)
	setStringP(v, "SipAuthPassword", p.SipAuthPassword)
	setStringP(v, "Byoc", p.Byoc)
	setBoolP(v, "AsyncAmd", p.AsyncAmd)
	setStringP(v, "CallToken", p.CallToken)
	return v
}

// UpdateCallParams is the body for POST /Calls/{sid}. Three flows on the same
// endpoint:
//   - Status="completed"|"canceled" — terminate the call.
//   - Twiml=<inline> — execute inline TwiML on the live call.
//   - URL=<...> — fetch new TwiML and execute it on the live call.
type UpdateCallParams struct {
	Status               *string  `form:"Status"`
	Twiml                *string  `form:"Twiml"`
	URL                  *string  `form:"Url"`
	Method               *string  `form:"Method"`
	FallbackURL          *string  `form:"FallbackUrl"`
	FallbackMethod       *string  `form:"FallbackMethod"`
	StatusCallback       *string  `form:"StatusCallback"`
	StatusCallbackMethod *string  `form:"StatusCallbackMethod"`
	StatusCallbackEvent  []string `form:"StatusCallbackEvent"`
}

func (p UpdateCallParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "Status", p.Status)
	setStringP(v, "Twiml", p.Twiml)
	setStringP(v, "Url", p.URL)
	setStringP(v, "Method", p.Method)
	setStringP(v, "FallbackUrl", p.FallbackURL)
	setStringP(v, "FallbackMethod", p.FallbackMethod)
	setStringP(v, "StatusCallback", p.StatusCallback)
	setStringP(v, "StatusCallbackMethod", p.StatusCallbackMethod)
	for _, s := range p.StatusCallbackEvent {
		v.Add("StatusCallbackEvent", s)
	}
	return v
}

// ListCallsParams are the filter / pagination query params for GET /Calls.
// StartTimeGte / StartTimeLte map to the legacy Twilio query keys
// "StartTime>=" / "StartTime<=" on the wire.
type ListCallsParams struct {
	To            string
	From          string
	Status        string
	ParentCallSid string
	StartTime     string
	StartTimeLt   string
	StartTimeGt   string
	EndTime       string
	EndTimeLt     string
	EndTimeGt     string
	StartTimeGte  string
	StartTimeLte  string
	Page          *int
	PageSize      *int
}

func (p ListCallsParams) query() url.Values {
	v := url.Values{}
	setString(v, "To", p.To)
	setString(v, "From", p.From)
	setString(v, "Status", p.Status)
	setString(v, "ParentCallSid", p.ParentCallSid)
	setString(v, "StartTime", p.StartTime)
	setString(v, "StartTime<", p.StartTimeLt)
	setString(v, "StartTime>", p.StartTimeGt)
	setString(v, "EndTime", p.EndTime)
	setString(v, "EndTime<", p.EndTimeLt)
	setString(v, "EndTime>", p.EndTimeGt)
	setString(v, "StartTime>=", p.StartTimeGte)
	setString(v, "StartTime<=", p.StartTimeLte)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	return v
}

// ListPageParams are shared Page / PageSize query params for stub list endpoints.
type ListPageParams struct {
	Page     *int
	PageSize *int
}

func (p ListPageParams) query() url.Values {
	v := url.Values{}
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	return v
}

// Create originates a new outbound call. POST /Calls.
func (s *CallsService) Create(ctx context.Context, params CreateCallParams) (*Call, error) {
	var out Call
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls"),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a page of calls matching the supplied filters. GET /Calls.
func (s *CallsService) List(ctx context.Context, params ListCallsParams) (*CallList, error) {
	var out CallList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single call by SID. GET /Calls/{sid}.
func (s *CallsService) Get(ctx context.Context, callSid string) (*Call, error) {
	var out Call
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update mutates a live call — terminate it, redirect it to new TwiML, or
// execute inline TwiML. POST /Calls/{sid}.
func (s *CallsService) Update(ctx context.Context, callSid string, params UpdateCallParams) (*Call, error) {
	var out Call
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a completed call. DELETE /Calls/{sid}.
func (s *CallsService) Delete(ctx context.Context, callSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("Calls", callSid),
	}, nil)
}

// --- Call-scoped Recordings ---

// ListRecordings returns recordings made on this call. GET /Calls/{sid}/Recordings.
func (s *CallsService) ListRecordings(ctx context.Context, callSid string, params ListCallRecordingsParams) (*RecordingList, error) {
	var out RecordingList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Recordings"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// StartRecording begins a new recording on this live call. POST /Calls/{sid}/Recordings.
// Pass nil for default options.
func (s *CallsService) StartRecording(ctx context.Context, callSid string, params *StartRecordingParams) (*Recording, error) {
	var form url.Values
	if params != nil {
		form = params.form()
	}
	var out Recording
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Recordings"),
		form:   form,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRecording fetches metadata for a recording on this call.
// GET /Calls/{sid}/Recordings/{rsid}.
func (s *CallsService) GetRecording(ctx context.Context, callSid, recordingSid string) (*Recording, error) {
	var out Recording
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Recordings", recordingSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateRecording changes the state of an in-progress recording on this call:
// pause, resume, or stop. POST /Calls/{sid}/Recordings/{rsid}.
func (s *CallsService) UpdateRecording(ctx context.Context, callSid, recordingSid string, params UpdateRecordingParams) (*Recording, error) {
	var out Recording
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Recordings", recordingSid),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteRecording removes a recording from this call.
// DELETE /Calls/{sid}/Recordings/{rsid}.
func (s *CallsService) DeleteRecording(ctx context.Context, callSid, recordingSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("Calls", callSid, "Recordings", recordingSid),
	}, nil)
}

// --- Streams ---

// Stream is a media-stream session on a call (REST equivalent of <Start><Stream>).
type Stream struct {
	Sid         string `json:"sid"`
	AccountSid  string `json:"account_sid"`
	CallSid     string `json:"call_sid"`
	Name        string `json:"name,omitempty"`
	Status      string `json:"status"`
	APIVersion  string `json:"api_version"`
	URI         string `json:"uri"`
	DateCreated string `json:"date_created,omitempty"`
	DateUpdated string `json:"date_updated,omitempty"`
}

// StreamList is the paginated /Calls/{sid}/Streams list response.
type StreamList struct {
	Page
	Streams []Stream `json:"streams"`
}

// StartStreamParams is the body for POST /Calls/{sid}/Streams. URL is the
// wss:// endpoint that receives media frames.
type StartStreamParams struct {
	URL                  string  `form:"Url"`
	Track                *string `form:"Track"`
	Name                 *string `form:"Name"`
	StatusCallback       *string `form:"StatusCallback"`
	StatusCallbackMethod *string `form:"StatusCallbackMethod"`
}

func (p StartStreamParams) form() url.Values {
	v := url.Values{}
	v.Set("Url", p.URL)
	setStringP(v, "Track", p.Track)
	setStringP(v, "Name", p.Name)
	setStringP(v, "StatusCallback", p.StatusCallback)
	setStringP(v, "StatusCallbackMethod", p.StatusCallbackMethod)
	return v
}

// StopStreamParams is the body for POST /Calls/{sid}/Streams/{sid}. Only
// "stopped" is accepted.
type StopStreamParams struct {
	Status string `form:"Status"`
}

func (p StopStreamParams) form() url.Values {
	v := url.Values{}
	v.Set("Status", p.Status)
	return v
}

// ListStreams returns streams on this call. GET /Calls/{sid}/Streams.
func (s *CallsService) ListStreams(ctx context.Context, callSid string) (*StreamList, error) {
	var out StreamList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Streams"),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// StartStream opens a new wss media stream on the live call.
// POST /Calls/{sid}/Streams.
func (s *CallsService) StartStream(ctx context.Context, callSid string, params StartStreamParams) (*Stream, error) {
	var out Stream
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Streams"),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetStream fetches a stream's current state.
// GET /Calls/{sid}/Streams/{sid}.
func (s *CallsService) GetStream(ctx context.Context, callSid, streamSid string) (*Stream, error) {
	var out Stream
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Streams", streamSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// StopStream closes an active stream. If params is nil, "stopped" is sent.
// POST /Calls/{sid}/Streams/{sid}.
func (s *CallsService) StopStream(ctx context.Context, callSid, streamSid string, params *StopStreamParams) (*Stream, error) {
	body := StopStreamParams{Status: "stopped"}
	if params != nil {
		body = *params
	}
	var out Stream
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Streams", streamSid),
		form:   body.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// --- SIPREC ---

// SiprecSession is a SIPREC recording session on a call (REST equivalent of <Start><Siprec>).
type SiprecSession struct {
	Sid           string `json:"sid"`
	AccountSid    string `json:"account_sid"`
	CallSid       string `json:"call_sid"`
	Name          string `json:"name,omitempty"`
	ConnectorName string `json:"connector_name,omitempty"`
	Status        string `json:"status"`
	APIVersion    string `json:"api_version"`
	URI           string `json:"uri"`
	DateCreated   string `json:"date_created,omitempty"`
	DateUpdated   string `json:"date_updated,omitempty"`
}

// SiprecList is the paginated /Calls/{sid}/Siprec list response.
type SiprecList struct {
	Page
	Siprec []SiprecSession `json:"siprec"`
}

// StartSiprecParams is the body for POST /Calls/{sid}/Siprec.
type StartSiprecParams struct {
	Name                 *string `form:"Name"`
	ConnectorName        *string `form:"ConnectorName"`
	Track                *string `form:"Track"`
	StatusCallback       *string `form:"StatusCallback"`
	StatusCallbackMethod *string `form:"StatusCallbackMethod"`
}

func (p StartSiprecParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "Name", p.Name)
	setStringP(v, "ConnectorName", p.ConnectorName)
	setStringP(v, "Track", p.Track)
	setStringP(v, "StatusCallback", p.StatusCallback)
	setStringP(v, "StatusCallbackMethod", p.StatusCallbackMethod)
	return v
}

// StopSiprecParams is the body for POST /Calls/{sid}/Siprec/{sid}. Clears
// VoiceML's session tracking only — the SRS recording continues until call
// hangup (documented mod_siprec limitation).
type StopSiprecParams struct {
	Status string `form:"Status"`
}

func (p StopSiprecParams) form() url.Values {
	v := url.Values{}
	v.Set("Status", p.Status)
	return v
}

// ListSiprec returns SIPREC sessions on this call. GET /Calls/{sid}/Siprec.
func (s *CallsService) ListSiprec(ctx context.Context, callSid string) (*SiprecList, error) {
	var out SiprecList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Siprec"),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// StartSiprec opens a new SIPREC session on the live call.
// POST /Calls/{sid}/Siprec. Pass nil for default options.
func (s *CallsService) StartSiprec(ctx context.Context, callSid string, params *StartSiprecParams) (*SiprecSession, error) {
	var form url.Values
	if params != nil {
		form = params.form()
	}
	var out SiprecSession
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Siprec"),
		form:   form,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSiprec fetches a SIPREC session's current state.
// GET /Calls/{sid}/Siprec/{sid}.
func (s *CallsService) GetSiprec(ctx context.Context, callSid, siprecSid string) (*SiprecSession, error) {
	var out SiprecSession
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Siprec", siprecSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// StopSiprec stops tracking a SIPREC session. Pass nil to default to
// Status="stopped". POST /Calls/{sid}/Siprec/{sid}.
func (s *CallsService) StopSiprec(ctx context.Context, callSid, siprecSid string, params *StopSiprecParams) (*SiprecSession, error) {
	body := StopSiprecParams{Status: "stopped"}
	if params != nil {
		body = *params
	}
	var out SiprecSession
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Siprec", siprecSid),
		form:   body.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Transcriptions ---

// CallTranscription is a live per-call transcription session. Transcript
// events stream via StatusCallback rather than this resource.
type CallTranscription struct {
	Sid                 string `json:"sid"`
	AccountSid          string `json:"account_sid"`
	CallSid             string `json:"call_sid"`
	Name                string `json:"name,omitempty"`
	LanguageCode        string `json:"language_code,omitempty"`
	TranscriptionEngine string `json:"transcription_engine,omitempty"`
	Status              string `json:"status"`
	APIVersion          string `json:"api_version"`
	URI                 string `json:"uri"`
	DateCreated         string `json:"date_created,omitempty"`
	DateUpdated         string `json:"date_updated,omitempty"`
}

// TranscriptionList is the paginated /Calls/{sid}/Transcriptions list response.
type TranscriptionList struct {
	Page
	Transcriptions []CallTranscription `json:"transcriptions"`
}

// StartTranscriptionParams is the body for POST /Calls/{sid}/Transcriptions.
type StartTranscriptionParams struct {
	Name                 *string `form:"Name"`
	Track                *string `form:"Track"`
	LanguageCode         *string `form:"LanguageCode"`
	TranscriptionEngine  *string `form:"TranscriptionEngine"`
	ProfanityFilter      *bool   `form:"ProfanityFilter"`
	PartialResults       *bool   `form:"PartialResults"`
	Hints                *string `form:"Hints"`
	StatusCallback       *string `form:"StatusCallback"`
	StatusCallbackMethod *string `form:"StatusCallbackMethod"`
	StatusCallbackEvents *string `form:"StatusCallbackEvents"`
}

func (p StartTranscriptionParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "Name", p.Name)
	setStringP(v, "Track", p.Track)
	setStringP(v, "LanguageCode", p.LanguageCode)
	setStringP(v, "TranscriptionEngine", p.TranscriptionEngine)
	setBoolP(v, "ProfanityFilter", p.ProfanityFilter)
	setBoolP(v, "PartialResults", p.PartialResults)
	setStringP(v, "Hints", p.Hints)
	setStringP(v, "StatusCallback", p.StatusCallback)
	setStringP(v, "StatusCallbackMethod", p.StatusCallbackMethod)
	setStringP(v, "StatusCallbackEvents", p.StatusCallbackEvents)
	return v
}

// StopTranscriptionParams is the body for POST /Calls/{sid}/Transcriptions/{sid}.
type StopTranscriptionParams struct {
	Status string `form:"Status"`
}

func (p StopTranscriptionParams) form() url.Values {
	v := url.Values{}
	v.Set("Status", p.Status)
	return v
}

// ListTranscriptions returns transcription sessions on this call.
// GET /Calls/{sid}/Transcriptions.
func (s *CallsService) ListTranscriptions(ctx context.Context, callSid string) (*TranscriptionList, error) {
	var out TranscriptionList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Transcriptions"),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// StartTranscription opens a new live transcription session.
// POST /Calls/{sid}/Transcriptions. Pass nil for default options.
func (s *CallsService) StartTranscription(ctx context.Context, callSid string, params *StartTranscriptionParams) (*CallTranscription, error) {
	var form url.Values
	if params != nil {
		form = params.form()
	}
	var out CallTranscription
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Transcriptions"),
		form:   form,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTranscription fetches a transcription session's current state.
// GET /Calls/{sid}/Transcriptions/{sid}.
func (s *CallsService) GetTranscription(ctx context.Context, callSid, transcriptionSid string) (*CallTranscription, error) {
	var out CallTranscription
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Transcriptions", transcriptionSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// StopTranscription ends a transcription session. Pass nil to default to
// Status="stopped". POST /Calls/{sid}/Transcriptions/{sid}.
func (s *CallsService) StopTranscription(ctx context.Context, callSid, transcriptionSid string, params *StopTranscriptionParams) (*CallTranscription, error) {
	body := StopTranscriptionParams{Status: "stopped"}
	if params != nil {
		body = *params
	}
	var out CallTranscription
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Transcriptions", transcriptionSid),
		form:   body.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Notifications / Events (compat stubs — always empty) ---

// NotificationsList is the always-empty compat stub for /Calls/{sid}/Notifications.
type NotificationsList struct {
	Notifications []any  `json:"notifications"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
	Total         int    `json:"total"`
	URI           string `json:"uri,omitempty"`
}

// EventsList is the always-empty compat stub for /Calls/{sid}/Events.
// The canonical event source is the customer's StatusCallback URL.
type EventsList struct {
	Events   []any  `json:"events"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Total    int    `json:"total"`
	URI      string `json:"uri,omitempty"`
}

// ListNotifications hits the compat stub at /Calls/{sid}/Notifications.
// Always returns an empty list when the call exists.
func (s *CallsService) ListNotifications(ctx context.Context, callSid string, params ListPageParams) (*NotificationsList, error) {
	var out NotificationsList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Notifications"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEvents hits the compat stub at /Calls/{sid}/Events. Always returns an
// empty list when the call exists.
func (s *CallsService) ListEvents(ctx context.Context, callSid string, params ListPageParams) (*EventsList, error) {
	var out EventsList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Calls", callSid, "Events"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SendUserDefinedMessage forwards a JSON payload to
// POST /Calls/{sid}/UserDefinedMessages. The server returns 501 — this method
// exists only to surface a clean *APIError (errors.Is(err, ErrNotImplemented))
// rather than make callers discover the missing endpoint at runtime.
func (s *CallsService) SendUserDefinedMessage(ctx context.Context, callSid string, payload any) error {
	return s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "UserDefinedMessages"),
		json:   payload,
	}, nil)
}

// Iterate walks every page of /Calls matching the supplied filters and
// returns the collected slice. Use for small-to-medium result sets; for very
// large pulls, drive pagination manually via List(...).NextPageURI.
func (s *CallsService) Iterate(ctx context.Context, params ListCallsParams) ([]Call, error) {
	out := []Call{}
	page := 0
	if params.Page != nil {
		page = *params.Page
	}
	for {
		params.Page = &page
		chunk, err := s.List(ctx, params)
		if err != nil {
			return nil, err
		}
		out = append(out, chunk.Calls...)
		if chunk.NextPageURI == "" || len(chunk.Calls) == 0 {
			return out, nil
		}
		page++
	}
}

// --- helpers shared across param structs ---

func setString(v url.Values, key, value string) {
	if value != "" {
		v.Set(key, value)
	}
}

func setStringP(v url.Values, key string, value *string) {
	if value != nil {
		v.Set(key, *value)
	}
}

func setIntP(v url.Values, key string, value *int) {
	if value != nil {
		v.Set(key, strconv.Itoa(*value))
	}
}

func setBoolP(v url.Values, key string, value *bool) {
	if value == nil {
		return
	}
	if *value {
		v.Set(key, "true")
	} else {
		v.Set(key, "false")
	}
}
