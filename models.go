package voiceml

// Shared types used across multiple resources. Per-resource models live in
// the file that owns them (calls.go, conferences.go, ...).

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

// Page is the Twilio-shape pagination envelope embedded in every list
// response. The concrete resource list field (Calls, Conferences, ...) is
// declared in its own list type alongside this embed.
type Page struct {
	Page            int    `json:"page"`
	PageSize        int    `json:"page_size"`
	NumPages        *int   `json:"num_pages,omitempty"`
	Total           *int   `json:"total,omitempty"`
	Start           *int   `json:"start,omitempty"`
	End             *int   `json:"end,omitempty"`
	FirstPageURI    string `json:"first_page_uri,omitempty"`
	NextPageURI     string `json:"next_page_uri,omitempty"`
	PreviousPageURI string `json:"previous_page_uri,omitempty"`
	URI             string `json:"uri,omitempty"`
}

// ErrorBody is the Twilio-shape JSON payload the server returns for non-2xx
// responses. Decoded automatically into *APIError.Code / .Message — exposed
// here for callers that want to re-parse APIError.Body themselves.
type ErrorBody struct {
	Code     any    `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
	MoreInfo string `json:"more_info,omitempty"`
	Status   int    `json:"status,omitempty"`
}

// HealthFailure is one tripped check from the /health deep probe.
type HealthFailure struct {
	Check  string `json:"check"`
	Detail string `json:"detail"`
}

// HealthStatus is the parsed /health body.
type HealthStatus struct {
	OK       bool            `json:"ok"`
	Warnings []HealthFailure `json:"warnings,omitempty"`
	Failures []HealthFailure `json:"failures,omitempty"`
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
