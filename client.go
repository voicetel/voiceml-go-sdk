// Package voiceml is the official Go SDK for the VoiceML REST API —
// VoiceTel's voice, messaging and AMD service with a Twilio-compatible REST
// surface.
//
// The API follows a Twilio-compatible wire format: the same resource paths,
// form-encoded parameters, error envelope and pagination envelope. Code
// written against twilio-go patterns ports over with little change, but
// VoiceML credentials are your VoiceML AccountSid plus a VoiceML API key;
// credentials issued by Twilio do not authenticate here.
//
// Quickstart:
//
//	c, err := voiceml.NewClient(voiceml.ClientOptions{
//		AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
//		APIKey:     os.Getenv("VOICEML_API_KEY"),
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	ctx := context.Background()
//	call, err := c.Calls.Create(ctx, voiceml.CreateCallParams{
//		To:               "+14155550100",
//		From:             "+14155550101",
//		URL:              voiceml.String("https://example.com/twiml"),
//		MachineDetection: voiceml.String("DetectMessageEnd"),
//	})
//
// See the package examples and README for pagination, error handling,
// webhook signature validation and diagnostics.
package voiceml

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ClientOptions configures a *Client. AccountSid and exactly one of APIKey /
// AuthToken are required; the rest have sensible defaults.
//
// ClientOptions redacts APIKey and AuthToken when formatted with fmt,
// marshalled to JSON or logged with log/slog.
type ClientOptions struct {
	// AccountSid is your VoiceML account identifier ("AC" + 32 hex chars).
	// Sent as the HTTP Basic auth username on every authenticated request.
	AccountSid string

	// APIKey is your VoiceML API key. Sent as the HTTP Basic auth password.
	// Set this OR AuthToken — never both.
	APIKey string

	// AuthToken is an alias for APIKey for code migrating from twilio-go
	// that already plumbs a value called "auth token". It must hold your
	// VoiceML API key; Twilio-issued auth tokens do not work. Set this OR
	// APIKey — never both.
	AuthToken string

	// BaseURL overrides the default server URL. Mostly useful for tests.
	// Defaults to DefaultBaseURL ("https://voiceml.voicetel.com"). Must be
	// an absolute http or https URL without query, fragment or user info.
	BaseURL string

	// MessagingBaseURL overrides the host the Messaging Service group
	// (c.MessagingV1) targets. When empty it is derived from BaseURL by
	// swapping the "voiceml" label for "messaging" on *.voicetel.com hosts
	// (e.g. messaging.voicetel.com), and falls back to BaseURL for any other
	// host. Set this to reach Messaging Service on a custom self-hosted subdomain.
	MessagingBaseURL string

	// ConversationsBaseURL overrides the host the Conversations group
	// (c.ConversationsV1) targets. When empty it is derived from BaseURL by
	// swapping the "voiceml" label for "conversations" on *.voicetel.com hosts
	// (e.g. conversations.voicetel.com), and falls back to BaseURL for any
	// other host.
	ConversationsBaseURL string

	// Timeout is the per-request timeout. Defaults to DefaultTimeout (30 s).
	// Ignored if HTTPClient is provided — set the timeout on the passed client.
	Timeout time.Duration

	// MaxRetries is the number of retries after the first attempt. Nil
	// defaults to DefaultMaxRetries (2); 0 disables retries; values < 0
	// cause NewClient to return *ConfigurationError.
	//
	// GET, HEAD, PUT and DELETE are retried on 429, 500, 502, 503, 504 and
	// on transport errors. POST (and PATCH) are retried only on 429 and on
	// connection failures where the request was never sent (DNS failure,
	// connection refused, connect timeout, TLS handshake failure), so a
	// create is never sent twice after the server may have acted on it.
	// Retry-After (seconds or HTTP-date) is honoured up to MaxRetryAfter;
	// otherwise the delay is exponential with full jitter (0.5 s base, 8 s
	// cap).
	MaxRetries *int

	// UserAgent overrides the User-Agent header. Defaults to "voiceml-go/<version>".
	UserAgent string

	// HTTPClient is an optional *http.Client to use for all requests. When
	// nil, the SDK constructs one with the configured Timeout. The SDK uses
	// a copy whose redirect policy drops Authorization once a redirect
	// leaves the API host and refuses https→http redirects; your own
	// CheckRedirect still runs afterwards.
	HTTPClient *http.Client
}

const redacted = "[redacted]"

func redact(s string) string {
	if s == "" {
		return ""
	}
	return redacted
}

// String implements fmt.Stringer without exposing the API key.
func (o ClientOptions) String() string {
	return fmt.Sprintf("voiceml.ClientOptions{AccountSid: %q, APIKey: %q, AuthToken: %q, BaseURL: %q, MessagingBaseURL: %q, ConversationsBaseURL: %q, Timeout: %s, UserAgent: %q}",
		o.AccountSid, redact(o.APIKey), redact(o.AuthToken), o.BaseURL, o.MessagingBaseURL, o.ConversationsBaseURL, o.Timeout, o.UserAgent)
}

// GoString implements fmt.GoStringer (%#v) without exposing the API key.
func (o ClientOptions) GoString() string { return o.String() }

// MarshalJSON emits the options with APIKey and AuthToken redacted.
func (o ClientOptions) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		AccountSid           string `json:"account_sid,omitempty"`
		APIKey               string `json:"api_key,omitempty"`
		AuthToken            string `json:"auth_token,omitempty"`
		BaseURL              string `json:"base_url,omitempty"`
		MessagingBaseURL     string `json:"messaging_base_url,omitempty"`
		ConversationsBaseURL string `json:"conversations_base_url,omitempty"`
		Timeout              string `json:"timeout,omitempty"`
		UserAgent            string `json:"user_agent,omitempty"`
	}{o.AccountSid, redact(o.APIKey), redact(o.AuthToken), o.BaseURL, o.MessagingBaseURL, o.ConversationsBaseURL, o.Timeout.String(), o.UserAgent})
}

// LogValue implements slog.LogValuer without exposing the API key.
func (o ClientOptions) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("account_sid", o.AccountSid),
		slog.String("api_key", redact(o.APIKey)),
		slog.String("auth_token", redact(o.AuthToken)),
		slog.String("base_url", o.BaseURL),
	)
}

// Client is the entry point to the SDK. Each public field is a resource
// service that exposes the operations under that resource group.
//
// Construct via NewClient. Clients are safe for concurrent use; share one
// across goroutines rather than constructing per request.
type Client struct {
	// AccountSid is the account whose resources this client targets.
	AccountSid string
	// BaseURL is the resolved default server URL (trailing slashes stripped).
	BaseURL string
	// MessagingBaseURL is the resolved host the Messaging Service group targets.
	MessagingBaseURL string
	// ConversationsBaseURL is the resolved host the Conversations group targets.
	ConversationsBaseURL string

	t *transport

	// Accounts — fetch and update the authenticated account.
	Accounts *AccountsService

	// Balance — the account balance.
	Balance *BalanceService

	// Calls — /Calls plus per-call recordings, streams, siprec,
	// transcriptions, payments, notifications, events, and user-defined
	// messages.
	Calls *CallsService

	// Conferences — list/get/end conferences; participant mute/hold/kick;
	// conference-scoped recordings.
	Conferences *ConferencesService

	// Queues — CRUD on queues; peek and dequeue (front or specific member).
	Queues *QueuesService

	// Applications — CRUD on stored TwiML+callback bundles.
	Applications *ApplicationsService

	// OutgoingCallerIDs — verified caller IDs.
	OutgoingCallerIDs *OutgoingCallerIDsService

	// Recordings — account-wide list, metadata fetch, audio download (the
	// 302 to a presigned URL is followed without credentials), delete, and
	// recording-scoped transcriptions.
	Recordings *RecordingsService

	// Transcriptions — account-wide post-recording transcriptions.
	Transcriptions *TranscriptionsService

	// Exports — daily account archives.
	Exports *ExportsService

	// UsageRecords — usage and price totals per category and time window.
	UsageRecords *UsageRecordsService

	// Diagnostics — /health and the OpenAPI document. Sent without
	// credentials. Use NewDiagnosticsService when you have no credentials.
	Diagnostics *DiagnosticsService

	// IncomingPhoneNumbers — CRUD on DIDs assigned to the authenticated tenant.
	IncomingPhoneNumbers *IncomingPhoneNumbersService

	// Notifications — account-scoped alert-log entries.
	Notifications *NotificationsService

	// MonitorAlerts — Monitor v1 Alerts: failure events across calls and
	// messaging (/v1/Alerts).
	MonitorAlerts *MonitorAlertsService

	// Messages — Twilio-compatible SMS/MMS resource plus message media.
	Messages *MessagesService

	// SIP — Twilio-compatible SIP Trunking (Domains, CredentialLists,
	// IpAccessControlLists, Registrations, plus the domain mapping
	// namespaces).
	SIP *SIPService

	// RoutesV2 — routes/v2 Inbound Processing Region API.
	RoutesV2 *RoutesV2Service

	// ConversationsV1 — the Conversations /v1 surface: stateful messaging
	// threads, messages, participants, scoped webhooks, roles, users, push
	// credentials, account configuration + addresses, services. Sent to
	// ConversationsBaseURL.
	ConversationsV1 *ConversationsV1Service

	// VoiceV1 — the Voice /v1 surface: ByocTrunks, ConnectionPolicies
	// (+ nested Targets), DialingPermissions Settings, SourceIpMappings, IpRecords.
	VoiceV1 *VoiceV1Service

	// AssistantsV1 — the AI-Assistants surface under /v1/Assistants:
	// Assistants + attached Tools and Knowledge, Sessions + Messages,
	// Feedback, and Tool/Knowledge-scoped Policies. JSON request bodies, no
	// Account SID in path (account resolved from HTTP Basic auth).
	AssistantsV1 *AssistantsV1Service

	// MessagingV1 — the Messaging /v1 surface: Messaging Services (MG...)
	// and their sender pools under /v1/Services. Sent to MessagingBaseURL,
	// which is what disambiguates a Messaging Service from a Conversation
	// Service (they share the /v1/Services path shape).
	MessagingV1 *MessagingV1Service

	// Pricing — read-only voice / messaging / phone-number / trunking price
	// lookups (v1 and v2) on the default host.
	Pricing *PricingService
}

// String implements fmt.Stringer; the API key is never included.
func (c Client) String() string {
	return fmt.Sprintf("voiceml.Client{AccountSid: %q, BaseURL: %q}", c.AccountSid, c.BaseURL)
}

// GoString implements fmt.GoStringer; the API key is never included.
func (c Client) GoString() string { return c.String() }

// NewClient constructs a *Client. Returns *ConfigurationError if AccountSid
// is missing, if neither APIKey nor AuthToken is set, if both are set
// simultaneously, if MaxRetries or Timeout is negative, or if a base URL is
// not an absolute http(s) URL.
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
	t, urls, err := buildTransport(opts)
	if err != nil {
		return nil, err
	}
	t.accountSid = opts.AccountSid
	t.apiKey = &secret{value: apiKey}
	return newClient(t, urls), nil
}

// NewDiagnosticsService returns a DiagnosticsService that needs no
// credentials, for health probes and spec fetches before (or without)
// provisioning an account. AccountSid, APIKey and AuthToken are ignored;
// BaseURL, HTTPClient, Timeout, MaxRetries and UserAgent apply as they do
// for NewClient.
func NewDiagnosticsService(opts ClientOptions) (*DiagnosticsService, error) {
	t, _, err := buildTransport(opts)
	if err != nil {
		return nil, err
	}
	return &DiagnosticsService{t: t}, nil
}

type productURLs struct{ base, messaging, conversations string }

func buildTransport(opts ClientOptions) (*transport, productURLs, error) {
	if opts.MaxRetries != nil && *opts.MaxRetries < 0 {
		return nil, productURLs{}, &ConfigurationError{Message: "MaxRetries must be >= 0"}
	}
	if opts.Timeout < 0 {
		return nil, productURLs{}, &ConfigurationError{Message: "Timeout must be >= 0"}
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	for _, u := range []struct{ name, value string }{
		{"BaseURL", baseURL}, {"MessagingBaseURL", opts.MessagingBaseURL}, {"ConversationsBaseURL", opts.ConversationsBaseURL},
	} {
		if u.value == "" {
			continue
		}
		if err := validateBaseURL(u.value); err != nil {
			return nil, productURLs{}, &ConfigurationError{Message: u.name + " " + err.Error()}
		}
	}

	maxRetries := DefaultMaxRetries
	if opts.MaxRetries != nil {
		maxRetries = *opts.MaxRetries
	}
	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	defaultURL, messagingURL, conversationsURL := resolveProductBaseURLs(
		strings.TrimRight(baseURL, "/"), opts.MessagingBaseURL, opts.ConversationsBaseURL,
	)
	t := &transport{
		baseURL:    defaultURL,
		userAgent:  userAgent,
		maxRetries: maxRetries,
		httpClient: newHTTPClient(opts.HTTPClient, timeout),
		sleep:      sleepContext,
		now:        time.Now,
	}
	return t, productURLs{defaultURL, messagingURL, conversationsURL}, nil
}

func validateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("is not a valid URL: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("must use http or https, got %q", raw)
	}
	if u.Host == "" {
		return fmt.Errorf("has no host: %q", raw)
	}
	if u.User != nil {
		return fmt.Errorf("must not embed credentials")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("must not have a query or fragment: %q", raw)
	}
	return nil
}

func newClient(t *transport, urls productURLs) *Client {
	c := &Client{
		AccountSid:           t.accountSid,
		BaseURL:              urls.base,
		MessagingBaseURL:     urls.messaging,
		ConversationsBaseURL: urls.conversations,
		t:                    t,
	}
	// Conversations and Messaging Service ride their own product subdomains
	// (they share the /v1/Services path shape — the host is what
	// disambiguates them). Each is a copy of the default transport pinned
	// to its product host; the underlying *http.Client is shared.
	messagingT := t.withBase(urls.messaging)
	conversationsT := t.withBase(urls.conversations)

	c.Accounts = &AccountsService{c: c}
	c.Balance = &BalanceService{c: c}
	c.Calls = &CallsService{c: c}
	c.Conferences = &ConferencesService{c: c}
	c.Queues = &QueuesService{c: c}
	c.Applications = &ApplicationsService{c: c}
	c.OutgoingCallerIDs = &OutgoingCallerIDsService{c: c}
	c.Recordings = &RecordingsService{c: c}
	c.Transcriptions = &TranscriptionsService{c: c}
	c.Exports = &ExportsService{c: c}
	c.UsageRecords = &UsageRecordsService{c: c}
	c.Diagnostics = &DiagnosticsService{t: t}
	c.IncomingPhoneNumbers = &IncomingPhoneNumbersService{c: c}
	c.Notifications = &NotificationsService{c: c}
	c.MonitorAlerts = &MonitorAlertsService{c: c}
	c.Messages = &MessagesService{c: c}
	c.SIP = &SIPService{
		Domains:              &SIPDomainsService{c: c},
		CredentialLists:      &SIPCredentialListsService{c: c},
		IpAccessControlLists: &SIPIpAccessControlListsService{c: c},
		Registrations:        &SIPRegistrationsService{c: c},
	}
	c.RoutesV2 = &RoutesV2Service{
		SipDomains: &RoutesV2SipDomainsService{c: c},
		c:          c,
	}
	c.ConversationsV1 = &ConversationsV1Service{t: conversationsT}
	c.VoiceV1 = &VoiceV1Service{c: c}
	c.AssistantsV1 = &AssistantsV1Service{c: c}
	c.MessagingV1 = &MessagingV1Service{
		Services: &MessagingV1ServicesService{t: messagingT},
	}
	c.Pricing = newPricingService(t)
	return c
}

// pathf builds an /2010-04-01/Accounts/{AccountSid}/... path from literal
// and caller-supplied segments (each percent-encoded) and appends the
// ".json" suffix every such JSON route carries in the spec.
func (c *Client) pathf(parts ...string) route {
	return c.pathfExt(".json", parts...)
}

// pathfExt is pathf with an explicit suffix: ".wav" for recording audio, ""
// for routes without one (Exports/{Date}).
func (c *Client) pathfExt(ext string, parts ...string) route {
	return joinRoute("/2010-04-01/Accounts", ext, append([]string{c.AccountSid}, parts...)...)
}

// apiRoute builds a /v1 or /v2 path (no suffix) from segments.
func apiRoute(prefix string, segs ...string) route {
	return joinRoute(prefix, "", segs...)
}

// joinRoute joins prefix and the percent-encoded segments, then appends
// suffix. A segment that is empty, "." or ".." would silently address a
// different resource, so it is rejected with *ValidationError.
func joinRoute(prefix, suffix string, segs ...string) route {
	var b strings.Builder
	b.WriteString(prefix)
	for _, s := range segs {
		if s == "" || s == "." || s == ".." {
			return route{err: &ValidationError{Message: fmt.Sprintf("invalid path parameter %q: SIDs and names must be non-empty and not \".\" or \"..\"", s)}}
		}
		b.WriteByte('/')
		b.WriteString(escapePathSegment(s))
	}
	b.WriteString(suffix)
	return route{path: b.String()}
}

// escapePathSegment percent-encodes every byte outside the RFC 3986
// unreserved set (A-Z a-z 0-9 - . _ ~) as %XX of its UTF-8 encoding, so
// "+" becomes %2B, " " %20, "/" %2F, "?" %3F, "#" %23 and "%" %25.
func escapePathSegment(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || ('0' <= c && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~' {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&0x0F])
	}
	return b.String()
}
