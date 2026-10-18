package voiceml

// Shared types used across multiple resources. Per-resource models live in
// the file that owns them (calls.go, conferences.go, ...).

// CallStatus is the lifecycle state of a Call resource.
type CallStatus string

const (
	CallStatusQueued     CallStatus = "queued"
	CallStatusRinging    CallStatus = "ringing"
	CallStatusInProgress CallStatus = "in-progress"
	CallStatusCompleted  CallStatus = "completed"
	CallStatusBusy       CallStatus = "busy"
	CallStatusNoAnswer   CallStatus = "no-answer"
	CallStatusCanceled   CallStatus = "canceled"
	CallStatusFailed     CallStatus = "failed"
)

// CallDirection indicates how the call was initiated.
type CallDirection string

const (
	CallDirectionInbound      CallDirection = "inbound"
	CallDirectionOutboundAPI  CallDirection = "outbound-api"
	CallDirectionOutboundDial CallDirection = "outbound-dial"
)

// AnsweredBy describes what entity answered the call (AMD result).
type AnsweredBy string

const (
	AnsweredByHuman             AnsweredBy = "human"
	AnsweredByMachineStart      AnsweredBy = "machine_start"
	AnsweredByMachineEndBeep    AnsweredBy = "machine_end_beep"
	AnsweredByMachineEndSilence AnsweredBy = "machine_end_silence"
	AnsweredByMachineEndOther   AnsweredBy = "machine_end_other"
	AnsweredByFax               AnsweredBy = "fax"
	AnsweredByUnknown           AnsweredBy = "unknown"
)

// ConferenceStatus is the lifecycle state of a Conference resource.
type ConferenceStatus string

const (
	ConferenceStatusInit       ConferenceStatus = "init"
	ConferenceStatusInProgress ConferenceStatus = "in-progress"
	ConferenceStatusCompleted  ConferenceStatus = "completed"
)

// ParticipantStatus is the lifecycle state of a conference Participant.
type ParticipantStatus string

const (
	ParticipantStatusQueued     ParticipantStatus = "queued"
	ParticipantStatusConnecting ParticipantStatus = "connecting"
	ParticipantStatusRinging    ParticipantStatus = "ringing"
	ParticipantStatusConnected  ParticipantStatus = "connected"
	ParticipantStatusOnHold     ParticipantStatus = "on-hold"
	ParticipantStatusComplete   ParticipantStatus = "complete"
	ParticipantStatusFailed     ParticipantStatus = "failed"
	ParticipantStatusCompleted  ParticipantStatus = "completed"
)

// RecordingStatus is the lifecycle state of a Recording resource.
type RecordingStatus string

const (
	RecordingStatusInProgress RecordingStatus = "in-progress"
	RecordingStatusPaused     RecordingStatus = "paused"
	RecordingStatusStopped    RecordingStatus = "stopped"
	RecordingStatusProcessing RecordingStatus = "processing"
	RecordingStatusCompleted  RecordingStatus = "completed"
	RecordingStatusAbsent     RecordingStatus = "absent"
	RecordingStatusDeleted    RecordingStatus = "deleted"
)

// RecordingSource indicates what triggered the recording.
type RecordingSource string

const (
	RecordingSourceOutboundAPI                 RecordingSource = "OutboundAPI"
	RecordingSourceRecordVerb                  RecordingSource = "RecordVerb"
	RecordingSourceDialVerb                    RecordingSource = "DialVerb"
	RecordingSourceConference                  RecordingSource = "Conference"
	RecordingSourceTrunking                    RecordingSource = "Trunking"
	RecordingSourceStartCallRecordingAPI       RecordingSource = "StartCallRecordingAPI"
	RecordingSourceStartConferenceRecordingAPI RecordingSource = "StartConferenceRecordingAPI"
)

// HTTPMethod is the verb set the VoiceML server accepts on callback URLs.
type HTTPMethod string

const (
	MethodGET  HTTPMethod = "GET"
	MethodPOST HTTPMethod = "POST"
)

// TrackSelector picks which media legs a stream / siprec / transcription
// captures. Used by Streams, Siprec, and Transcriptions.
type TrackSelector string

const (
	TrackInbound  TrackSelector = "inbound_track"
	TrackOutbound TrackSelector = "outbound_track"
	TrackBoth     TrackSelector = "both_tracks"
)

// ErrorBody is the Twilio-compatible JSON payload the server returns for
// non-2xx responses. Decoded automatically into *APIError — exposed here for
// callers that want to re-parse APIError.Body themselves.
type ErrorBody struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info,omitempty"`
	Status   int    `json:"status"`
}

// boolPtr / intPtr / stringPtr are convenience constructors for optional
// fields on request structs. Callers can pass voiceml.Bool(true) instead of
// declaring a local variable.

// Bool returns a pointer to b. Use when constructing request structs with
// optional bool fields.
func Bool(b bool) *bool { return &b }

// Int returns a pointer to i. Use when constructing request structs with
// optional int fields.
func Int(i int) *int { return &i }

// String returns a pointer to s. Use when constructing request structs with
// optional string fields. (For required string fields, pass the value directly.)
func String(s string) *string { return &s }
