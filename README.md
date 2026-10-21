# 📞 VoiceML Go SDK

The official Go client for the [VoiceML REST API](https://voicetel.com/docs/api/v0.9/voiceml/) — VoiceTel's Twilio-compatible voice, messaging and answering-machine-detection API — with strongly-typed, context-aware Go.

![Version](https://img.shields.io/badge/version-0.9.4-blue)
![Go](https://img.shields.io/badge/go-1.21%2B-blue)
![License](https://img.shields.io/badge/license-MIT%20%2B%20Commons%20Clause-green)
![Typed](https://img.shields.io/badge/typed-native%20structs-blue)

## 📚 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Quickstart](#-quickstart)
- [Authentication](#-authentication)
- [Resource Reference](#-resource-reference)
- [Error Handling](#-error-handling)
- [Retries, Timeouts and Rate Limits](#-retries-timeouts-and-rate-limits)
- [Pagination](#-pagination)
- [Webhook Signatures](#-webhook-signatures)
- [Downloads](#-downloads)
- [Diagnostics](#-diagnostics)
- [Booleans and Optional Fields](#-booleans-and-optional-fields)
- [Migrating from twilio-go](#-migrating-from-twilio-go)
- [Development](#-development)
- [API Documentation](#-api-documentation)
- [Contributors](#-contributors)
- [Sponsors](#-sponsors)
- [License](#-license)

## ✨ Features

- **Complete coverage** — all 333 of the 333 operations in the VoiceML OpenAPI document, checked on every test run by a spec-driven contract test (method, path, host, auth and every request parameter).
- **Typed requests and responses** — a params struct per request and a Go struct per response. Optional request fields are pointers (`*bool`, `*string`, `*int`) so "not set" and "zero" stay distinct.
- **Context-aware** — every method takes `context.Context` first; cancellation and deadlines propagate to the HTTP layer and are returned unchanged.
- **Safe retries** — idempotent requests retry on 429/5xx and network errors; creates (POST) are never re-sent after the server may have acted on them. `Retry-After` is honoured (seconds or HTTP-date, capped).
- **Lazy pagination** that follows the server's `next_page_uri` (with its `PageToken`) — no page counting.
- **Webhook signature validation** (`X-Twilio-Signature`) with constant-time comparison.
- **Credential hygiene** — the API key never appears in `fmt`, JSON or `slog` output, and redirects to storage hosts are followed without your credentials.
- **Standard library only** — `net/http`, no third-party dependencies.

## 🚀 Installation

```bash
go get github.com/voicetel/voiceml-go-sdk
```

Requires Go 1.21 or later. `Pager.All` (range-over-func iteration) is available when building with Go 1.23 or later.

## 🏁 Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

func main() {
	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
		APIKey:     os.Getenv("VOICEML_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	call, err := c.Calls.Create(ctx, voiceml.CreateCallParams{
		To:               "+14155550100",
		From:             "+14155550101",
		URL:              voiceml.String("https://example.com/twiml"),
		MachineDetection: voiceml.String("DetectMessageEnd"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(call.Sid, call.Status)

	queues, err := c.Queues.List(ctx, voiceml.ListPageParams{PageSize: voiceml.Int(50)})
	if err != nil {
		log.Fatal(err)
	}
	for _, q := range queues.Queues {
		fmt.Println(q.FriendlyName, q.CurrentSize)
	}
}
```

## 🔑 Authentication

Authenticated endpoints use **HTTP Basic** auth: your VoiceML `AccountSid` is the username and your VoiceML API key is the password — the same header positions Twilio SDKs use. Credentials issued by Twilio do not work against VoiceML; use the AccountSid and API key from your VoiceML account.

```go
c, err := voiceml.NewClient(voiceml.ClientOptions{
	AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
	APIKey:     os.Getenv("VOICEML_API_KEY"),
})
if err != nil {
	log.Fatal(err)
}
acct, err := c.Accounts.Fetch(ctx)
if err != nil {
	log.Fatal(err)
}
fmt.Println(acct.FriendlyName, acct.Status)
```

`ClientOptions`, `Client` and `RequestValidator` redact the key when printed with `fmt`, marshalled to JSON or logged with `log/slog`.

> Don't have credentials yet? See **[voicetel.com/docs/api/v0.9/voiceml/](https://voicetel.com/docs/api/v0.9/voiceml/)** for issuance and rotation.

## 🗺️ Resource Reference

| Service | Path prefix | Covers |
| --- | --- | --- |
| `c.Accounts` | `/Accounts/{Sid}` | fetch and update the authenticated account (409 / code 20409 on a concurrent update: `IsConflict`); `RotateKey` generates a new API key (the response carries it once — store it right away) |
| `c.Balance` | `/Balance` | account balance |
| `c.Calls` | `/Calls` | create, list, fetch, update, delete; per-call recordings (list/start/fetch/update/delete), streams / SIPREC / live transcriptions (start/stop), payments, notifications, events, user-defined messages |
| `c.Conferences` | `/Conferences` | list, fetch, end; participants (add / mute / hold / kick); conference recordings (start/list/fetch/update/delete) |
| `c.Queues` | `/Queues` | CRUD; members: list, peek, dequeue (front or specific member) |
| `c.Applications` | `/Applications` | CRUD on TwiML + callback bundles, including SMS delivery-receipt callbacks |
| `c.OutgoingCallerIDs` | `/OutgoingCallerIds` | list, verify (create), fetch, update, delete |
| `c.Recordings` | `/Recordings` | account-wide list, metadata, audio download, delete; recording transcriptions (list/fetch/delete) |
| `c.Transcriptions` | `/Transcriptions` | account-wide post-recording transcriptions: list, fetch, delete |
| `c.Messages` | `/Messages` | SMS/MMS create (`MediaURL`), fetch, list, update, delete; media (list/fetch/delete); delivery feedback (`CreateFeedback`) |
| `c.IncomingPhoneNumbers` | `/IncomingPhoneNumbers` | list (incl. Local / Mobile / TollFree), create, fetch, update (voice and SMS routing), delete |
| `c.Notifications` | `/Notifications` | alert-log entries: list, fetch |
| `c.MonitorAlerts` | `/v1/Alerts` | Monitor v1 failure events (calls and messaging): list with `LogLevel` / `StartDate` / `EndDate` filters, fetch with request/response detail |
| `c.UsageRecords` | `/Usage/Records` | usage per category: custom range plus AllTime, Daily, LastMonth, Monthly, ThisMonth, Today, Yearly, Yesterday |
| `c.Exports` | `/Exports` | daily archives: list, download |
| `c.SIP` | `/SIP` | Domains, CredentialLists + Credentials, IpAccessControlLists + IpAddresses, Registrations, and the domain mappings (historical, Auth/Calls, Auth/Registrations) |
| `c.RoutesV2` | `/v2` | Inbound Processing Region for SIP domains and phone numbers |
| `c.VoiceV1` | `/v1` | BYOC trunks, connection policies + targets, dialing-permission settings, IP records, source-IP mappings |
| `c.ConversationsV1` | `/v1` (conversations host) | conversations, messages, participants, webhooks, receipts, roles, users, push credentials, configuration, addresses, services and their service-scoped resources |
| `c.MessagingV1.Services` | `/v1/Services` (messaging host) | Messaging Services (`MG…`) CRUD and their sender pool (`ListPhoneNumbers`, `CreatePhoneNumber`, `FetchPhoneNumber`, `DeletePhoneNumber`) |
| `c.AssistantsV1` | `/v1` | assistants, tools, knowledge, sessions, messages, feedback, policies |
| `c.Pricing.V1` / `.V2` | `/v1`, `/v2` | read-only voice / messaging / phone-number / trunking prices |
| `c.Diagnostics` | `/health`, `/openapi.*` | health probe and the live OpenAPI document (sent without credentials) |

Streams, SIPREC sessions and live call transcriptions are start/stop only: the server does not list or fetch them.

### 🌐 Product hosts

Two groups are served on their own host, derived automatically from the configured `BaseURL`:

| Group | Host (derived from `voiceml.voicetel.com`) |
| --- | --- |
| `c.ConversationsV1` | `conversations.voicetel.com` |
| `c.MessagingV1` | `messaging.voicetel.com` |
| everything else | `voiceml.voicetel.com` (default) |

A Messaging Service (`MG…`) and a Conversation Service (`IS…`) share the `/v1/Services` paths; the host is what disambiguates them on the wire. Derivation swaps the `voiceml` label only for `*.voicetel.com` hosts — any other `BaseURL` (a self-hosted instance, a test stub) keeps every group on that single host. Override a product host explicitly with `ClientOptions.MessagingBaseURL` / `ConversationsBaseURL`.

Every method that takes a request body accepts a typed params struct:

```go
call, err := c.Calls.Create(ctx, voiceml.CreateCallParams{
	To:   "+14155550100",
	From: "+14155550101",
	URL:  voiceml.String("https://example.com/twiml"),
})
if err != nil {
	log.Fatal(err)
}

// On a live call, open a Pay session:
payment, err := c.Calls.StartPayment(ctx, call.Sid, voiceml.CreatePaymentParams{
	ChargeAmount:   voiceml.String("19.99"),
	Currency:       voiceml.String("USD"),
	StatusCallback: voiceml.String("https://example.com/pay-status"),
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(payment.Sid)

// Send an MMS: Body, MediaURL or both.
msg, err := c.Messages.Create(ctx, voiceml.CreateMessageParams{
	To:       "+14155550100",
	From:     voiceml.String("+14155550101"),
	Body:     voiceml.String("Your receipt"),
	MediaURL: []string{"https://example.com/receipt.png"},
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(msg.Sid, msg.Status)
```

Path parameters (SIDs, phone numbers, domain names) are percent-encoded, so a value such as `+14155550100` or one containing `/` or `?` cannot change which resource a request addresses. An empty SID is rejected with `*voiceml.ValidationError` before anything is sent.

## 🚨 Error Handling

Every non-2xx response is returned as a `*voiceml.APIError`. The HTTP status drives a sentinel error that the `*APIError` wraps, so you can branch with `errors.Is`:

```go
_, err := c.Calls.Get(ctx, "CA0123456789abcdef0123456789abcdef")
switch {
case errors.Is(err, voiceml.ErrNotFound):
	// 404 — call doesn't exist or belongs to another tenant
case voiceml.IsRateLimit(err):
	// 429 that outlasted the retries (or asked for longer than MaxRetryAfter)
case voiceml.IsTimeout(err):
	// ClientOptions.Timeout elapsed
case err != nil:
	log.Fatal(err)
}
```

| Status | Sentinel | Convenience checker |
| --- | --- | --- |
| 400 | `ErrBadRequest` | — |
| 401 | `ErrAuthentication` | `IsAuthentication(err)` |
| 403 | `ErrPermissionDenied` | — |
| 404 | `ErrNotFound` | `IsNotFound(err)` |
| 405 | `ErrMethodNotAllowed` | — |
| 409 | `ErrConflict` | `IsConflict(err)` |
| 410 | `ErrGone` | — |
| 429 | `ErrRateLimit` | `IsRateLimit(err)` |
| 501 | `ErrNotImplemented` | `IsNotImplemented(err)` |
| 5xx | `ErrServer` | `IsServer(err)` |
| no response | `ErrTransport` (+ `ErrTimeout`) | `IsTransport(err)`, `IsTimeout(err)` |

The error envelope (`code`, `message`, `more_info`, `status`) is parsed into `apiErr.Code` (an `int`), `apiErr.Message`, `apiErr.MoreInfo` and `apiErr.Status`, with the raw body on `apiErr.Body`. A non-JSON error body still produces an `*APIError`. Transport errors keep the underlying `net/http` error in the chain (`errors.As(err, &netErr)` works).

```go
_, err := c.Queues.Get(ctx, "QU0123456789abcdef0123456789abcdef")
var apiErr *voiceml.APIError
if errors.As(err, &apiErr) {
	fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message, apiErr.MoreInfo, apiErr.RetryAfter)
}
```

Other error types: `*ConfigurationError` (invalid `ClientOptions`), `*ValidationError` (a request that cannot be sent, such as an empty SID or a message with neither `Body` nor `MediaURL`), `*DecodeError` (a 2xx body that could not be decoded; it wraps the `encoding/json` error) and `*PaginationError` (a next-page link the SDK refuses to follow). Context cancellation and deadlines are returned unchanged (`context.Canceled`, `context.DeadlineExceeded`).

## ⏱️ Retries, Timeouts and Rate Limits

| Request | Retried on |
| --- | --- |
| GET, HEAD, PUT, DELETE | 429, 500, 502, 503, 504, and network errors |
| POST (creates, updates) | 429, and connection failures where the request never left the client: DNS failure, connection refused, connect timeout, TLS handshake failure |

A POST is **never** retried after a 5xx, a read timeout or a connection reset, because the server may already have placed the call or sent the message. Defaults: `MaxRetries` 2, 30 s per-request `Timeout`.

- `Retry-After` is honoured as delta-seconds or an HTTP-date. A value above `voiceml.MaxRetryAfter` (60 s) is not waited out: the error is returned at once, with the requested delay in `APIError.RetryAfter`.
- Otherwise the delay is exponential with full jitter: a random wait up to 0.5 s, 1 s, 2 s … capped at 8 s.

```go
c, err := voiceml.NewClient(voiceml.ClientOptions{
	AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
	APIKey:     os.Getenv("VOICEML_API_KEY"),
	MaxRetries: voiceml.Int(5), // explicit 0 disables retries
	Timeout:    10 * time.Second,
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(c)
```

Pass a custom `*http.Client` to share a connection pool, route through a proxy, or stub for tests (its own `Timeout` then applies). The SDK uses a copy of it whose redirect policy drops `Authorization` once a redirect leaves the API host and refuses https→http redirects; your `CheckRedirect` still runs afterwards.

```go
c, err := voiceml.NewClient(voiceml.ClientOptions{
	AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
	APIKey:     os.Getenv("VOICEML_API_KEY"),
	HTTPClient: &http.Client{Timeout: 15 * time.Second},
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(c)
```

## 📄 Pagination

List methods return one page: a `*…List` struct embedding the pagination envelope (`Page`, `PageSize`, `Start`, `End`, `FirstPageURI`, `NextPageURI`, `PreviousPageURI`, `URI`). `/v1` lists carry a `Meta` envelope instead (`NextPageURL`, …).

`Iterate` methods walk every page lazily. They follow each page's `next_page_uri` verbatim — it carries the server's `PageToken` — and stop at the last page. Nothing is requested until the first `Next`, and stopping early sends nothing more:

```go
it := c.Calls.Iterate(ctx, voiceml.ListCallsParams{Status: "completed", PageSize: voiceml.Int(100)})
for it.Next() {
	call := it.Current()
	fmt.Println(call.Sid, call.Duration)
}
if err := it.Err(); err != nil {
	log.Fatal(err)
}
```

With Go 1.23 or later you can range over `Pager.All()`:

```go
for call, err := range c.Calls.Iterate(ctx, voiceml.ListCallsParams{}).All() {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(call.Sid)
}
```

To page by hand, pass the current page to `Client.NextPage`; it reports `false` at the last page. For `/v1` lists pass the list's `Meta`:

```go
page, err := c.Recordings.List(ctx, voiceml.ListRecordingsParams{PageSize: voiceml.Int(50)})
if err != nil {
	log.Fatal(err)
}
for {
	for _, r := range page.Recordings {
		fmt.Println(r.Sid)
	}
	var next voiceml.RecordingList
	more, err := c.NextPage(ctx, page, &next)
	if err != nil {
		log.Fatal(err)
	}
	if !more {
		break
	}
	page = &next
}

trunks, err := c.VoiceV1.ListByocTrunks(ctx, voiceml.V1PageParams{})
if err != nil {
	log.Fatal(err)
}
var moreTrunks voiceml.VoiceV1ByocTrunkList
if _, err := c.NextPage(ctx, trunks.Meta, &moreTrunks); err != nil {
	log.Fatal(err)
}
```

A next link is only followed on the scheme and host that served the page; any other link is refused with `*voiceml.PaginationError`, so credentials are never sent elsewhere.

## ✍️ Webhook Signatures

VoiceML signs every webhook it sends with your API key in the `X-Twilio-Signature` header. Validate it before trusting a request. Pass the public URL VoiceML called (as configured in your callback, including any query string) — behind a proxy it differs from `r.URL`:

```go
validator := voiceml.NewRequestValidator(os.Getenv("VOICEML_API_KEY"))

http.HandleFunc("/voice/status", func(w http.ResponseWriter, r *http.Request) {
	if !validator.ValidateHTTPRequest(r, "https://example.com/voice/status") {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}
	fmt.Println(r.PostForm.Get("CallSid"), r.PostForm.Get("CallStatus"))
})
```

`voiceml.ValidateRequest(apiKey, url, params, signature)` validates already-parsed parameters, and `voiceml.ComputeSignature` produces signatures for testing your handlers. The algorithm is HMAC-SHA1 over the URL (without fragment) followed, for POST, by each form parameter as key+value in sorted key order; GET webhooks sign the URL alone. Comparison is constant-time, and the URL is also tried with the default port (`:443` / `:80`) added or removed.

## 📥 Downloads

Recording audio and account exports are binary. `DownloadAudio` and `Exports.Download` stream the body; `GetAudio` reads a recording into memory (up to 1 GiB):

```go
d, err := c.Recordings.DownloadAudio(ctx, "RE0123456789abcdef0123456789abcdef")
if err != nil {
	log.Fatal(err)
}
defer d.Body.Close()
f, err := os.Create("recording.wav")
if err != nil {
	log.Fatal(err)
}
defer f.Close()
if _, err := io.Copy(f, d.Body); err != nil {
	log.Fatal(err)
}
```

Archived audio answers with a 302 to a presigned storage URL. The SDK follows it without your credentials (`Authorization` is dropped as soon as a redirect leaves the API host) and refuses https→http redirects.

## 🩺 Diagnostics

`/health` and the OpenAPI document need no credentials. Use `c.Diagnostics`, or `NewDiagnosticsService` when you have no credentials at all. Either way the request goes through your configured HTTP client, timeout and User-Agent:

```go
diag, err := voiceml.NewDiagnosticsService(voiceml.ClientOptions{})
if err != nil {
	log.Fatal(err)
}
health, err := diag.Health(ctx)
if err != nil {
	log.Fatal(err) // a 503 is an *APIError wrapping ErrServer; its Body lists the failures
}
fmt.Println(health.OK, len(health.Warnings))

spec, err := diag.OpenAPIYAML(ctx)
if err != nil {
	log.Fatal(err)
}
fmt.Println(len(spec))
```

## 🔢 Booleans and Optional Fields

Optional fields on request structs are pointers (`*bool`, `*string`, `*int`) so the SDK can distinguish "unset" from "zero". Use the helpers:

```go
params := voiceml.UpdateParticipantParams{
	Muted: voiceml.Bool(false), // sent as Muted=false
	Hold:  voiceml.Bool(true),
}
fmt.Println(voiceml.String("text"), voiceml.Int(42), voiceml.Date(time.Now()), params)
```

Booleans are sent as `true` / `false`, list parameters (`StatusCallbackEvent`, `MediaURL`, …) as repeated keys, and date-only filters as `YYYY-MM-DD` (`voiceml.Date` formats a `time.Time` in UTC).

## 🔁 Migrating from twilio-go

VoiceML's wire format is Twilio-compatible, so moving is mostly a matter of constructing a VoiceML client with your **VoiceML** AccountSid and API key (Twilio-issued credentials do not work here):

```go
c, err := voiceml.NewClient(voiceml.ClientOptions{
	AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
	APIKey:     os.Getenv("VOICEML_API_KEY"),
})
if err != nil {
	log.Fatal(err)
}

// Code that already plumbs a value called "auth token" can keep doing so;
// AuthToken is an alias for APIKey and must hold your VoiceML API key.
c, err = voiceml.NewClient(voiceml.ClientOptions{
	AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
	AuthToken:  os.Getenv("VOICEML_API_KEY"),
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(c)
```

Method calls (`c.Calls.Create(ctx, params)`, `c.Queues.List(ctx, params)`) follow the resource table above rather than a nested `client.Api…Calls.Create(params)` chain.

## 🛠️ Development

```bash
git clone https://github.com/voicetel/voiceml-go-sdk
cd voiceml-go-sdk

go build ./...
go vet ./...
test -z "$(gofmt -l .)"
go test -race ./...
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

The test suite needs no network. It includes a contract test that checks every public method against the vendored OpenAPI document in `testdata/`, and a test that compiles every Go snippet in this README. An additional fixture-conformance test runs only when `VOICEML_CONFORMANCE_FIXTURES` points at the server team's private fixture corpus; otherwise it is skipped with a message saying so.

## 📖 API Documentation

- **Reference docs:** [voicetel.com/docs/api/v0.9/voiceml/](https://voicetel.com/docs/api/v0.9/voiceml/)
- **Validator:** [voicetel.com/voiceml/validator/](https://voicetel.com/voiceml/validator/)
- **SDK catalogue:** [voicetel.com/docs/voiceml-sdks/](https://voicetel.com/docs/voiceml-sdks/)
- **Go package docs:** [pkg.go.dev/github.com/voicetel/voiceml-go-sdk](https://pkg.go.dev/github.com/voicetel/voiceml-go-sdk)
- **Changes:** [CHANGELOG.md](CHANGELOG.md)

## 🙌 Contributors

- [Michael Mavroudis](https://github.com/mavroudis) — Lead Developer

Contributions welcome. Open an issue describing the change you want to make, or send a pull request against `main`.

## 💖 Sponsors

| Sponsor | Contribution |
|---------|--------------|
| [VoiceTel Communications](https://voicetel.com) | Primary development and production hosting |

## 📄 License

MIT with the Commons Clause restriction. See [LICENSE](LICENSE) and [voicetel.com/legal/](https://voicetel.com/legal/).
