# 📞 VoiceML Go SDK

The official Go client for the [VoiceML REST API](https://voicetel.com/docs/api/v0.6/voiceml/) — Twilio-compatible outbound voice and answering-machine-detection from VoiceTel, with strongly-typed, context-aware Go.

![Version](https://img.shields.io/badge/version-0.7.1.1-blue)
![Go](https://img.shields.io/badge/go-1.21%2B-blue)
![License](https://img.shields.io/badge/license-MIT%20%2B%20Commons%20Clause-green)
![Tests](https://img.shields.io/badge/tests-46%20unit-brightgreen)
![Typed](https://img.shields.io/badge/typed-native%20structs-blue)

## 📚 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Quickstart](#-quickstart)
- [Authentication](#-authentication)
- [Resource Reference](#-resource-reference)
- [Error Handling](#-error-handling)
- [Pagination](#-pagination)
- [Booleans and Optional Fields](#-booleans-and-optional-fields)
- [Migration from twilio-go](#-migration-from-twilio-go)
- [Rate Limits](#-rate-limits)
- [Development](#-development)
- [API Documentation](#-api-documentation)
- [Contributors](#-contributors)
- [Sponsors](#-sponsors)
- [License](#-license)

## ✨ Features

### 🛡️ Strongly Typed End-to-End
- **Native Go structs** for every one of the 81 API operations across 9 resource families — request params encoded directly, responses decoded into typed structs.
- **Pointer types for optional request fields** (`*bool`, `*string`, `*int`) — distinguish "not set" from "zero" cleanly when PATCH-ing.
- **Context-aware throughout.** Every method takes `context.Context` as the first argument; cancel and timeouts propagate down to the HTTP layer.
- **Twilio-compatible wire shapes** — `AccountSid`, `From`, `To`, status callbacks, pagination envelopes — match what Twilio's Programmable Voice API documents.

### 🔁 Production-Grade Transport
- Built on `net/http` — pure standard library, no third-party dependencies.
- **Automatic retry** with exponential backoff on 429 / 5xx and transport errors — honors numeric `Retry-After` headers.
- **Configurable timeout** per client (default 30 s).
- **HTTP Basic auth** with `AccountSid:APIKey` — exactly what the Twilio SDK uses, so existing credentials work unchanged.
- **Structured `*APIError`** with status-driven sentinels (`ErrNotFound`, `ErrRateLimit`, …) usable with `errors.Is` / `errors.As`.

### 📞 Complete API Coverage
- **Calls** — originate, fetch, terminate, update + per-call recordings, streams, siprec, transcriptions, notifications, events, user-defined messages, and the `/Calls/{sid}/Payments` lifecycle (Pay TwiML companion).
- **Conferences** — list, fetch, end conferences, plus participants (mute / hold / kick) and conference-scoped recordings.
- **Queues** — create, list, update, delete, peek, dequeue (front or specific member).
- **Applications** — CRUD on stored TwiML + callback bundles.
- **Recordings** — account-wide list, metadata fetch, audio fetch (follows S3 redirect), delete.
- **Messages** — create, fetch, list (To/From/DateSent filters + pagination), update (Body redaction; Status=canceled), delete.
- **IncomingPhoneNumbers** — list, fetch, update.
- **Notifications** — fetch, list.
- **Diagnostics** — `/health` deep probe, `/openapi.json` live spec fetch.

### 🧪 Tested
- **46 unit tests** with `httptest`-based fakes exercising every service and every error path.
- **Race-detector clean** (`go test -race ./...`).
- **`go vet` and `gofmt` clean.**
- **Integration test suite** that runs against a callBroadcast / VoiceML instance — gated by env vars, safe for CI.

### 📦 Clean Distribution
- Zero codegen footprint — every byte hand-written.
- Single module (`github.com/voicetel/voiceml-go-sdk`); install with `go get`.
- No external dependencies beyond the standard library.

## 🚀 Installation

```bash
go get github.com/voicetel/voiceml-go-sdk
```

Requires Go 1.21 or later.

## 🏁 Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"

    voiceml "github.com/voicetel/voiceml-go-sdk"
)

func main() {
    c, err := voiceml.NewClient(voiceml.ClientOptions{
        AccountSid: "AC...",
        APIKey:     "...",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    call, err := c.Calls.Create(ctx, voiceml.CreateCallParams{
        To:               "+18005551234",
        From:             "+18005550000",
        URL:              voiceml.String("https://example.com/twiml"),
        MachineDetection: voiceml.String("DetectMessageEnd"),
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(call.Sid, call.Status)

    queues, err := c.Queues.List(ctx)
    if err != nil {
        log.Fatal(err)
    }
    for _, q := range queues.Queues {
        fmt.Println(q.FriendlyName, q.CurrentSize)
    }
}
```

## 🔑 Authentication

Every endpoint uses **HTTP Basic** with your `AccountSid` as the username and your per-tenant API key as the password — identical to Twilio's auth shape, so credentials issued for Twilio code work here unchanged.

```go
c, err := voiceml.NewClient(voiceml.ClientOptions{
    AccountSid: "AC...",
    APIKey:     "...",
})
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
health, err := c.Diagnostics.Health(ctx) // uses your AccountSid + key on every call
if err != nil {
    log.Fatal(err)
}
fmt.Println(health.Status)
```

> Don't have credentials yet? See **[voicetel.com/docs/api/v0.6/voiceml/](https://voicetel.com/docs/api/v0.6/voiceml/)** for issuance and rotation.

## 🗺️ Resource Reference

| Service | Path prefix | Covers |
| --- | --- | --- |
| `client.Calls` | `/Calls` | originate, fetch, list, terminate, update + per-call recordings, streams, siprec, transcriptions, notifications, events, payments |
| `client.Conferences` | `/Conferences` | list, fetch, end conferences; participants (mute / hold / kick); conference-scoped recordings |
| `client.Queues` | `/Queues` | create, list, update, delete; peek, dequeue (front or specific member) |
| `client.Applications` | `/Applications` | CRUD on TwiML + callback bundles |
| `client.Recordings` | `/Recordings` | account-wide list, metadata, audio fetch (follows S3 redirect), delete |
| `client.Messages` | `/Messages` | create, fetch, list, update, delete; To/From/DateSent filters; Body redaction; Status=canceled |
| `client.IncomingPhoneNumbers` | `/IncomingPhoneNumbers` | list, fetch, update |
| `client.Notifications` | `/Notifications` | fetch, list |
| `client.Diagnostics` | `/health`, `/openapi.json` | deep liveness probe; live spec fetch (unauthenticated) |

Every method that takes a request body accepts a typed params struct:

```go
ctx := context.Background()

call, err := c.Calls.Create(ctx, voiceml.CreateCallParams{
    To:   "+18005551234",
    From: "+18005550000",
    URL:  voiceml.String("https://example.com/twiml"),
})
if err != nil {
    log.Fatal(err)
}

// On a live call, open a Pay session:
session, err := c.Calls.StartPayment(ctx, call.Sid, voiceml.StartPaymentParams{
    IdempotencyKey: "order-482917",
    StatusCallback: voiceml.String("https://example.com/pay-status"),
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(session.Sid, session.Status)
```

## 🚨 Error Handling

Every non-2xx response is returned as a `*voiceml.APIError`. The HTTP status drives a sentinel error that the `*APIError` wraps, so you can branch with `errors.Is`:

```go
_, err := c.Calls.Get(ctx, sid)
if errors.Is(err, voiceml.ErrNotFound) {
    // 404 — call doesn't exist or belongs to another tenant
}
if errors.Is(err, voiceml.ErrRateLimit) {
    // 429 — back off and retry
}
```

| Status | Sentinel | Convenience checker |
| --- | --- | --- |
| 400 | `ErrBadRequest` | — |
| 401 | `ErrAuthentication` | `IsAuthentication(err)` |
| 403 | `ErrPermissionDenied` | — |
| 404 | `ErrNotFound` | `IsNotFound(err)` |
| 409 | `ErrConflict` | — |
| 410 | `ErrGone` | — |
| 429 | `ErrRateLimit` | `IsRateLimit(err)` |
| 501 | `ErrNotImplemented` | `IsNotImplemented(err)` |
| 5xx | `ErrServer` | `IsServer(err)` |
| network | `ErrTransport` | — |

The Twilio-compatible body (`code`, `message`, `more_info`, `status`) is parsed into `apiErr.Code`, `apiErr.Message`, and `apiErr.MoreInfo`, with the raw response on `apiErr.Body`:

```go
var apiErr *voiceml.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message, apiErr.MoreInfo)
}
```

## 📄 Pagination

List operations return a `*…List` struct with a Twilio-compatible pagination envelope embedded (`Page`, `PageSize`, `Total`, `NextPageURI`, `PreviousPageURI`, …). For `/Calls`, the convenience method `Iterate` walks every page transparently:

```go
all, err := c.Calls.Iterate(ctx, voiceml.ListCallsParams{Status: "completed"})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("fetched %d calls\n", len(all))
```

For other resources, page manually using the envelope:

```go
ctx := context.Background()
page := 0
for {
    chunk, err := c.Queues.List(ctx)
    if err != nil {
        log.Fatal(err)
    }
    for _, q := range chunk.Queues {
        fmt.Println(q.FriendlyName, q.CurrentSize)
    }
    if chunk.NextPageURI == "" {
        break
    }
    page++
}
```

## 🔢 Booleans and Optional Fields

Optional fields on request structs are pointers (`*bool`, `*string`, `*int`) so the SDK can distinguish "unset" from "zero". Use the helpers:

```go
voiceml.Bool(true)
voiceml.String("...")
voiceml.Int(42)
```

Booleans are serialized as the literal strings `"true"` / `"false"` (Twilio convention).

## 🔁 Migration from twilio-go

The same `AccountSid` + API key pair the Twilio Go SDK uses works here:

```go
// Before — Twilio
import "github.com/twilio/twilio-go"
tw := twilio.NewRestClientWithParams(twilio.ClientParams{
    Username: "AC...", Password: "<auth_token>",
})

// After — VoiceML (Twilio-compatible), same credentials, different host
import voiceml "github.com/voicetel/voiceml-go-sdk"
c, _ := voiceml.NewClient(voiceml.ClientOptions{
    AccountSid: "AC...", APIKey: "<api_key>",
})

// Migrating from twilio-go? AuthToken is accepted as an alias for APIKey
// so existing wiring just works:
c, _ = voiceml.NewClient(voiceml.ClientOptions{
    AccountSid: "AC...", AuthToken: "<api_key>",
})
```

Method shapes (`c.Calls.Create(ctx, params)`, `c.Queues.List(ctx)`) follow the resource table above rather than Twilio's nested `client.Api.V2010.Accounts(sid).Calls.Create(params)` chain — flatter, fewer keystrokes, same wire format on the way out.

## ⏱️ Rate Limits

VoiceML applies per-tenant rate limits at the edge. The transport automatically retries 429/5xx and transport errors up to `MaxRetries` times with exponential backoff. `Retry-After` is honored when the server emits a numeric value. Defaults: 2 retries, 30 s per-request timeout.

```go
c, _ := voiceml.NewClient(voiceml.ClientOptions{
    AccountSid: "AC...",
    APIKey:     "...",
    MaxRetries: voiceml.Int(5), // explicit 0 disables retries
    Timeout:    10 * time.Second,
})
```

Pass a custom `*http.Client` to share a connection pool, route through a proxy, or stub for tests:

```go
c, _ := voiceml.NewClient(voiceml.ClientOptions{
    AccountSid: "AC...",
    APIKey:     "...",
    HTTPClient: myClient,
})
```

## 🛠️ Development

```bash
git clone https://github.com/voicetel/voiceml-go-sdk
cd voiceml-go-sdk

# Build + unit tests (fast, no network)
go build ./...
go test ./...

# Race detector + vet
go test -race ./...
go vet ./...

# Integration tests (live, read-only against a configured VoiceML instance)
cp .env.example .env  # fill in VOICEML_ACCOUNT_SID / VOICEML_API_KEY / VOICEML_BASE_URL
go test -tags=integration ./...
```

## 📖 API Documentation

- **Reference docs:** [voicetel.com/docs/api/v0.6/voiceml/](https://voicetel.com/docs/api/v0.6/voiceml/)
- **Validator:** [voicetel.com/voiceml/validator/](https://voicetel.com/voiceml/validator/)
- **SDK catalogue:** [voicetel.com/docs/voiceml-sdks/](https://voicetel.com/docs/voiceml-sdks/)
- **Go package docs:** [pkg.go.dev/github.com/voicetel/voiceml-go-sdk](https://pkg.go.dev/github.com/voicetel/voiceml-go-sdk)

## 🙌 Contributors

- [Michael Mavroudis](https://github.com/mavroudis) — Lead Developer

Contributions welcome. Open an issue describing the change you want to make, or send a pull request against `main`.

## 💖 Sponsors

| Sponsor | Contribution |
|---------|--------------|
| [VoiceTel Communications](https://voicetel.com) | Primary development and production hosting |

## 📄 License

MIT with the Commons Clause restriction. See [LICENSE](LICENSE) and [voicetel.com/legal/](https://voicetel.com/legal/).
