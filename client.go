// Package voiceml is the official Go SDK for the VoiceML REST API — VoiceTel's
// outbound voice + AMD service with a Twilio-compatible REST surface.
//
// The wire format, authentication model, error codes, and pagination envelope
// all match Twilio's documented Programmable Voice surface. If you've used
// twilio-go, the patterns here will feel familiar.
//
// Quickstart:
//
//	import (
//		"context"
//		"fmt"
//		"github.com/voicetel/voiceml-go-sdk"
//	)
//
//	func main() {
//		c, err := voiceml.NewClient(voiceml.ClientOptions{
//			AccountSid: "AC...",
//			APIKey:     "...",
//		})
//		if err != nil {
//			panic(err)
//		}
//		ctx := context.Background()
//		call, err := c.Calls.Create(ctx, voiceml.CreateCallParams{
//			To:               "+18005551234",
//			From:             "+18005550000",
//			URL:              voiceml.String("https://example.com/twiml"),
//			MachineDetection: voiceml.String("DetectMessageEnd"),
//		})
//		if err != nil {
//			panic(err)
//		}
//		fmt.Println(call.Sid, call.Status)
//	}
package voiceml

import (
	"crypto/tls"
	"net/http"
	"strings"
	"time"
)

// ClientOptions configures a *Client. AccountSid and exactly one of APIKey /
// AuthToken are required; the rest have sensible defaults.
type ClientOptions struct {
	// AccountSid is the Twilio-format account identifier ("AC" + 32 hex chars).
	// Sent as the HTTP Basic auth username on every authenticated request.
	AccountSid string

	// APIKey is the per-tenant secret. Sent as the HTTP Basic auth password.
	// Set this OR AuthToken — never both.
	APIKey string

	// AuthToken is a Twilio-ergonomic alias for APIKey. Useful for code paths
	// migrating from twilio-go that already plumb a value called "auth token".
	// Set this OR APIKey — never both. If only AuthToken is set, it is used as
	// the HTTP Basic auth password.
	AuthToken string

	// BaseURL overrides the default server URL. Mostly useful for tests.
	// Defaults to DefaultBaseURL ("https://voiceml.voicetel.com").
	BaseURL string

	// Timeout is the per-request timeout. Defaults to DefaultTimeout (30 s).
	// Ignored if HTTPClient is provided — set the timeout on the passed client.
	Timeout time.Duration

	// MaxRetries is the number of retries for 429/5xx and transport errors.
	// Nil defaults to DefaultMaxRetries (2). Set explicitly to 0 to disable
	// retries; values < 0 cause NewClient to return *ConfigurationError.
	MaxRetries *int

	// UserAgent overrides the User-Agent header. Defaults to "voiceml-go/<version>".
	UserAgent string

	// HTTPClient is an optional *http.Client to use for all requests. When nil,
	// the SDK constructs one with the configured Timeout.
	HTTPClient *http.Client
}

// Client is the entry point to the SDK. Each public field is a resource
// service that exposes the operations under that resource group.
//
// Construct via NewClient. Clients are safe for concurrent use; share one
// across goroutines rather than constructing per request.
type Client struct {
	// AccountSid is the account whose resources this client targets.
	AccountSid string
	// BaseURL is the resolved server URL (with trailing slashes stripped).
	BaseURL string

	t *transport

	// Calls — /Calls plus per-call recordings, streams, siprec, transcriptions,
	// notifications, events, and user-defined messages.
	Calls *CallsService

	// Conferences — list/get/end conferences; participant mute/hold/kick;
	// conference-scoped recordings.
	Conferences *ConferencesService

	// Queues — CRUD on queues; peek and dequeue (front or specific member).
	Queues *QueuesService

	// Applications — CRUD on stored TwiML+callback bundles.
	Applications *ApplicationsService

	// Recordings — account-wide list, metadata fetch, audio fetch (follows
	// the 302 → S3 presigned URL), delete.
	Recordings *RecordingsService

	// Diagnostics — /health and /openapi.json. Unauthenticated.
	Diagnostics *DiagnosticsService

	// IncomingPhoneNumbers — CRUD on DIDs assigned to the authenticated tenant.
	IncomingPhoneNumbers *IncomingPhoneNumbersService

	// Notifications — account-scoped compat stubs (always empty).
	Notifications *NotificationsService

	// Messages — Twilio-compatible SMS resource backed by VoiceTel's
	// SDK 2.2 gateway. Outbound-only today.
	Messages *MessagesService
}

// NewClient constructs a *Client. Returns *ConfigurationError if AccountSid
// is missing, if neither APIKey nor AuthToken is set, if both are set
// simultaneously, or if MaxRetries is negative.
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.AccountSid == "" {
		return nil, &ConfigurationError{Message: "AccountSid is required"}
	}
	if opts.APIKey != "" && opts.AuthToken != "" {
		return nil, &ConfigurationError{Message: "set APIKey or AuthToken, not both"}
	}
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = opts.AuthToken
	}
	if apiKey == "" {
		return nil, &ConfigurationError{Message: "APIKey is required"}
	}
	if opts.MaxRetries != nil && *opts.MaxRetries < 0 {
		return nil, &ConfigurationError{Message: "MaxRetries must be >= 0"}
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	maxRetries := DefaultMaxRetries
	if opts.MaxRetries != nil {
		maxRetries = *opts.MaxRetries
	}

	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		timeout := opts.Timeout
		if timeout == 0 {
			timeout = DefaultTimeout
		}
		tr := http.DefaultTransport.(*http.Transport).Clone()
		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = &tls.Config{}
		}
		tr.TLSClientConfig.ClientSessionCache = tls.NewLRUClientSessionCache(0)
		httpClient = &http.Client{Timeout: timeout, Transport: tr}
	}

	t := &transport{
		accountSid: opts.AccountSid,
		apiKey:     apiKey,
		baseURL:    baseURL,
		userAgent:  userAgent,
		maxRetries: maxRetries,
		httpClient: httpClient,
	}

	c := &Client{
		AccountSid: opts.AccountSid,
		BaseURL:    baseURL,
		t:          t,
	}
	c.Calls = &CallsService{c: c}
	c.Conferences = &ConferencesService{c: c}
	c.Queues = &QueuesService{c: c}
	c.Applications = &ApplicationsService{c: c}
	c.Recordings = &RecordingsService{c: c}
	c.Diagnostics = &DiagnosticsService{c: c}
	c.IncomingPhoneNumbers = &IncomingPhoneNumbersService{c: c}
	c.Notifications = &NotificationsService{c: c}
	c.Messages = &MessagesService{c: c}
	return c, nil
}

// pathf builds a path under /2010-04-01/Accounts/{AccountSid}/... by joining
// the supplied segments with a single slash and appending the Twilio-style
// ".json" suffix. Empty segments are skipped so callers can pass conditional
// sub-resource names.
//
// Callers that need a non-".json" representation (e.g. recording audio at
// ".wav") should use pathfExt instead — pathf is for the default JSON shape.
func (c *Client) pathf(parts ...string) string {
	return c.pathfExt(".json", parts...)
}

// pathfExt is the explicit-extension form of pathf. Used by GetAudio to ask
// for ".wav" instead of ".json". Pass "" to skip the suffix entirely (no
// current callers do that, but it keeps the helper general).
func (c *Client) pathfExt(ext string, parts ...string) string {
	out := "/2010-04-01/Accounts/" + c.AccountSid
	for _, p := range parts {
		if p == "" {
			continue
		}
		out += "/" + p
	}
	if ext != "" && !strings.HasSuffix(out, ext) {
		out += ext
	}
	return out
}
