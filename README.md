# voiceml-go-sdk

Official Go SDK for the [VoiceML](https://voiceml.voicetel.com) REST API — VoiceTel's outbound voice + AMD service with a Twilio-shaped REST surface.

Wire format, auth model (HTTP Basic with `AccountSid` as username, per-tenant API key as password), error codes, and pagination envelope all match Twilio's documented Programmable Voice surface. If you've used `twilio-go`, the patterns here will feel familiar.

## Install

```bash
go get github.com/voicetel/voiceml-go-sdk
```

Requires Go 1.21+. Pure stdlib — no external dependencies.

## Quickstart

```go
package main

import (
    "context"
    "fmt"

    voiceml "github.com/voicetel/voiceml-go-sdk"
)

func main() {
    c, err := voiceml.NewClient(voiceml.ClientOptions{
        AccountSid: "AC...",
        APIKey:     "...",
    })
    if err != nil {
        panic(err)
    }

    ctx := context.Background()

    call, err := c.Calls.Create(ctx, voiceml.CreateCallParams{
        To:               "+18005551234",
        From:             "+18005550000",
        URL:              voiceml.String("https://example.com/twiml"),
        MachineDetection: voiceml.String("DetectMessageEnd"),
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(call.Sid, call.Status)

    // Pull every call across all pages.
    all, err := c.Calls.Iterate(ctx, voiceml.ListCallsParams{Status: "completed"})
    if err != nil {
        panic(err)
    }
    fmt.Printf("fetched %d calls\n", len(all))
}
```

## Resources

| Service | Path prefix | Covers |
| --- | --- | --- |
| `client.Calls` | `/Calls` | originate, fetch, terminate, update + per-call recordings, streams, siprec, transcriptions, notifications, events, user-defined messages |
| `client.Conferences` | `/Conferences` | list/fetch/end conferences, participants (mute/hold/kick), conference-scoped recordings |
| `client.Queues` | `/Queues` | create/list/update/delete queues, peek, dequeue (front or specific member) |
| `client.Applications` | `/Applications` | CRUD on stored TwiML+callback bundles |
| `client.Recordings` | `/Recordings` | account-wide list, metadata fetch, audio fetch (follows S3 redirect), delete |
| `client.Diagnostics` | `/health`, `/openapi.json` | deep liveness probe; live spec fetch (unauthenticated) |

## Errors

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

The Twilio-shape body (`code`, `message`, `more_info`, `status`) is parsed into `apiErr.Code` and `apiErr.Message`, with the raw response on `apiErr.Body`:

```go
var apiErr *voiceml.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message)
}
```

## Twilio drop-in

The same `AccountSid` + API key pair the Twilio Go SDK uses works here:

```go
// Twilio
import "github.com/twilio/twilio-go"
tw := twilio.NewRestClientWithParams(twilio.ClientParams{
    Username: "AC...", Password: "<auth_token>",
})

// VoiceML — same credentials, different host
import voiceml "github.com/voicetel/voiceml-go-sdk"
c, _ := voiceml.NewClient(voiceml.ClientOptions{
    AccountSid: "AC...", APIKey: "<api_key>",
})
```

Method shapes (`c.Calls.Create(ctx, params)`, `c.Queues.List(ctx)`) follow the resource table above rather than Twilio's nested `client.Api.V2010.Accounts(sid).Calls.Create(params)` chain.

## Pagination

List operations return a `*…List` struct with the Twilio-shape pagination envelope embedded (`Page`, `PageSize`, `Total`, `NextPageURI`, `PreviousPageURI`, …). For `/Calls`, the convenience method `Iterate` walks every page:

```go
all, err := c.Calls.Iterate(ctx, voiceml.ListCallsParams{Status: "completed"})
```

For other resources, page manually:

```go
page := 0
for {
    chunk, err := c.Queues.List(ctx)
    if err != nil || chunk.NextPageURI == "" {
        break
    }
    // ... process chunk.Queues
    page++
}
```

## Retries + timeouts

The transport retries 429/5xx and transport errors up to `MaxRetries` times with exponential backoff. `Retry-After` is honored when the server emits a numeric value. Defaults: 2 retries, 30 s per-request timeout.

```go
c, _ := voiceml.NewClient(voiceml.ClientOptions{
    AccountSid: "AC...",
    APIKey:     "...",
    MaxRetries: voiceml.Int(5),  // explicit 0 disables retries
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

## Booleans and optional fields

Optional fields on request structs are pointers (`*bool`, `*string`, `*int`) so the SDK can distinguish "unset" from "zero". Use the helpers:

```go
voiceml.Bool(true)
voiceml.String("...")
voiceml.Int(42)
```

Booleans are serialized as the literal strings `"true"` / `"false"` (Twilio convention).

## Development

```bash
go build ./...
go test ./...
go vet ./...
```

## License

MIT with the Commons Clause restriction. See [LICENSE](LICENSE) and [voicetel.com/legal/](https://voicetel.com/legal/).
