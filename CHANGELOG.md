# Changelog

All notable changes to the VoiceML Go SDK are recorded here. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions
follow [Semantic Versioning](https://semver.org/). Earlier releases are
described on the [GitHub Releases](https://github.com/voicetel/voiceml-go-sdk/releases)
page.

## [0.9.5] - 2026-10-21

### Added

- Calls create (`CreateCallParams.TimeLimit`): max call duration in
  seconds. An explicit `0` or a value above `14400` is rejected
  server-side; a negative value disables the limit.
- Participants create (`CreateParticipantParams.ConferenceStatusCallback`,
  `ConferenceStatusCallbackMethod`, `ConferenceStatusCallbackEvent`): the
  participant's conference lifecycle webhook.
- Conference recordings update (`UpdateRecordingParams.PauseBehavior`):
  `skip` / `silence` with `Status="paused"`, mirroring the call-scoped
  recording update.
- Applications create/update (`ApplicationParams.SmsURL`, `SmsMethod`,
  `SmsFallbackURL`, `SmsFallbackMethod`, `PublicApplicationConnectEnabled`):
  the inbound-SMS webhook configuration and the public application connect
  toggle.
- Messages create (`CreateMessageParams.SendAt`, `ScheduleType`): schedule
  a message for a future UTC time. `SendAt` takes RFC 3339 (the new `Time`
  helper formats a `time.Time` in UTC), `ScheduleType` is `"fixed"` only,
  and both require `MessagingServiceSid`.
- Call transcriptions start (`StartTranscriptionParams.SpeechModel`,
  `EnableAutomaticPunctuation`, `IntelligenceService`).

Note: `StartTranscriptionParams.StatusCallback` is renamed
`StatusCallbackURL` (sending `StatusCallbackUrl`); the old name was never
applied server-side.

## [0.9.4] - 2026-10-21

### Added

- IncomingPhoneNumbers create/update could not set `VoiceApplicationSid`,
  `StatusCallback` or `StatusCallbackMethod`; OutgoingCallerIds create could
  not set `CallDelay`, `Extension`, `StatusCallback` or
  `StatusCallbackMethod`.

## [0.9.3] - 2026-10-18

### Breaking changes

Removed endpoints. The server answers these with 405 Method Not Allowed
(code 20004); the paths now only accept POST, so the start/stop methods stay:

- `CallsService.ListStreams` — `GET /2010-04-01/Accounts/{AccountSid}/Calls/{CallSid}/Streams.json`
- `CallsService.GetStream` — `GET /2010-04-01/Accounts/{AccountSid}/Calls/{CallSid}/Streams/{Sid}.json`
- `CallsService.ListSiprec` — `GET /2010-04-01/Accounts/{AccountSid}/Calls/{CallSid}/Siprec.json`
- `CallsService.GetSiprec` — `GET /2010-04-01/Accounts/{AccountSid}/Calls/{CallSid}/Siprec/{Sid}.json`
- `CallsService.ListTranscriptions` — `GET /2010-04-01/Accounts/{AccountSid}/Calls/{CallSid}/Transcriptions.json`
- `CallsService.GetTranscription` — `GET /2010-04-01/Accounts/{AccountSid}/Calls/{CallSid}/Transcriptions/{Sid}.json`
- The `StreamList` and `SiprecList` types are removed. `TranscriptionList` now
  lists post-recording `Transcription` items (`GET …/Transcriptions.json` and
  `GET …/Recordings/{RecordingSid}/Transcriptions.json`).

Request changes:

- `GET …/Calls.json`: `ListCallsParams.StartTimeGte` / `StartTimeLte` are
  removed; they sent `StartTime>=` / `StartTime<=`, which the API does not
  declare. Use `StartTimeGt` / `StartTimeLt` (`StartTime>` / `StartTime<`).
- `POST …/Messages.json`: `CreateMessageParams.Body` is now `*string`, and a
  message needs `Body` or `MediaURL` (otherwise `Create` returns
  `*ValidationError` without sending).
- `POST …/IncomingPhoneNumbers/{Local,Mobile,TollFree}.json`: `CreateLocal`,
  `CreateMobile` and `CreateTollFree` take `CreateTypedIncomingPhoneNumberParams`
  (PhoneNumber, VoiceURL, VoiceMethod) and no longer send `VoiceFallbackUrl` /
  `VoiceFallbackMethod`, which those endpoints do not accept. `ListLocal`,
  `ListMobile` and `ListTollFree` take `*ListIncomingPhoneNumbersParams`
  (`ListTypedIncomingPhoneNumbersParams` remains as a deprecated alias).
- `POST …/Calls/{CallSid}/Recordings/{Sid}.json`: `CallsService.UpdateRecording`
  takes `UpdateCallRecordingParams` (adds `PauseBehavior`).
  `ConferencesService.UpdateRecording` keeps `UpdateRecordingParams`.
- `POST …/Calls/{CallSid}/UserDefinedMessages.json`:
  `SendUserDefinedMessage(ctx, callSid)` drops its payload argument and sends
  an empty form instead of JSON, as the spec declares. The server answers 501.
- `GET …/SIP/Domains.json`, `…/SIP/CredentialLists.json`,
  `…/SIP/IpAccessControlLists.json` and
  `…/SIP/IpAccessControlLists/{Sid}/IpAddresses.json`: the list methods take
  `ListSIPParams` (adds `FriendlyName`, embeds `ListPageParams`).
- `POST /2010-04-01/Accounts/{Sid}.json` (new `Accounts.Update`) is sent as
  JSON so an empty `AllowCidr` list can clear the allow-list.
- Path parameters are percent-encoded (`+` → `%2B`, `/` → `%2F`, `?` → `%3F`,
  space → `%20`, …) and an empty, `.` or `..` parameter returns
  `*ValidationError` before anything is sent. Previously an empty SID was
  dropped from the path (`Calls.Get(ctx, "")` requested `GET …/Calls.json`),
  and Pricing numbers were query-escaped (space became a literal `+`).
- The `.json` suffix is always appended to JSON routes; a SID ending in
  `.json` is no longer special-cased.

Pagination:

- `Iterate` on Calls, Conferences, Queues and Recordings returns a lazy
  `*Pager[T]` instead of `([]T, error)`. It follows each page's
  `next_page_uri` verbatim (with its `PageToken`) instead of incrementing
  `Page`, which never terminated on Calls and Conferences and failed with 400
  on page 2 of Messages and Recordings.
- `Page.NumPages` and `Page.Total` are removed; the server does not send them.
- `RecordingList`, `IncomingPhoneNumbersList`, the notification list and the
  call-event list now embed `Page` (for example `RecordingList.Page` is the
  envelope, not `*int`; use `RecordingList.Page.Page` for the page number).

Response models:

- Conversations v1 objects the spec leaves untyped are `map[string]any`
  (`media` items are `[]map[string]any`) instead of `map[string]string`, so
  numbers and nested objects decode: `timers`, `links`, `bindings`,
  `delivery`, `media`, `messaging_binding`, `participant_messaging_binding`,
  `conversation_timers`, `configuration`, `auto_creation` and the service
  notification sub-objects.
- Notifications are typed: `NotificationsService.Get` and
  `CallsService.GetNotification` return `*Notification` (was
  `map[string]any`); the lists return `*NotificationList` and
  `CallsService.ListEvents` returns `*CallEventList` of `CallEvent`.
  `NotificationsList` and `EventsList` remain as deprecated aliases.
- Fields the server does not send are removed: `Stream.APIVersion`,
  `Stream.DateCreated`, `SiprecSession.APIVersion`, `SiprecSession.DateCreated`,
  `SiprecSession.ConnectorName`, `CallTranscription.APIVersion`,
  `CallTranscription.DateCreated`, `CallTranscription.LanguageCode`,
  `CallTranscription.TranscriptionEngine`, `Participant.APIVersion`,
  `QueueMember.AccountSid` and `Application.StatusCallbackEvent` (still a
  request parameter).

Errors and transport:

- `APIError.Code` is an `int` (was `any` holding `int64` or `string`;
  numeric strings are parsed). `APIError` gains `Status` and `RetryAfter`.
- A 2xx body that cannot be decoded returns `*DecodeError`, which wraps the
  `encoding/json` error, instead of an `*APIError` labelled "non-JSON".
- `ConfigurationError.Error()` is prefixed with `voiceml: `.
- POST requests are no longer retried on 500/502/503/504, read timeouts or
  connection resets (see Fixed). `Retry-After` above 60 s is returned to the
  caller instead of being waited out.
- `Diagnostics.Health` makes a single attempt.
- `NewClient` rejects a negative `Timeout` and base URLs that are not
  absolute http(s) URLs, or that carry a query, fragment or credentials.
- The SDK sends with a copy of `ClientOptions.HTTPClient` whose redirect
  policy drops `Authorization` once a redirect leaves the API host and
  refuses https→http redirects. `GetAudio` sends `Accept: audio/*, */*` and
  reads at most 1 GiB (use `DownloadAudio` to stream).

### Fixed

- Iterators never terminated (Calls, Conferences) or failed on page 2
  (Messages, Recordings): they now follow `next_page_uri`.
- `POST /Calls`, `POST /Messages` and every other create were retried on
  5xx and on client timeouts with the default settings, which could place
  duplicate calls or send duplicate messages. POST is now retried only on
  429 and when the connection could not be established.
- `Retry-After` HTTP-dates were ignored and numeric values were uncapped
  (`3600` slept an hour; `Inf` retried immediately). Both forms are parsed,
  non-finite and negative values are ignored, and the wait is capped at 60 s.
  Backoff without `Retry-After` uses full jitter (0.5 s base, 8 s cap).
- Path parameters could rewrite the request path or query (`a?Foo=1`,
  `a#frag`, `a/b`, `../`); they are now percent-encoded.
- `messages.create` could not send `MediaUrl` (MMS), `StatusCallbackMethod`
  or `ValidityPeriod`.
- IncomingPhoneNumbers create/update could not set `SmsUrl`, `SmsMethod`,
  `SmsFallbackUrl`, `SmsFallbackMethod` or `SmsApplicationSid`; the lists
  could not filter by `FriendlyName`, `Beta` or `Origin`.
- Applications create/update could not set `MessageStatusCallback` or
  `SmsStatusCallback`; `Application` lacked `message_status_callback`,
  `sms_status_callback` and `sms_url` / `sms_method` / `sms_fallback_url` /
  `sms_fallback_method`.
- `callRecordings.create` could not set `RecordingTrack`; `callRecordings.update`
  could not set `PauseBehavior`; `Recording` lacked `track`.
- SIP Domain, CredentialList, IpAccessControlList and IpAddress lists could
  not filter by `FriendlyName`.
- `IncomingPhoneNumber` lacked `subresource_uris`; `MessagingService` lacked
  `advanced_opt_out`; the notification and event lists lacked `start`,
  `end`, `first_page_uri`, `next_page_uri` and `previous_page_uri`.
- Conversations messages whose `media` or `delivery` carried numbers or
  nested objects failed to decode.
- Decode errors on 2xx responses discarded the underlying error.
- Response bodies were read without a size limit; JSON responses are now
  limited to 32 MiB and downloads stream.
- Diagnostics required credentials to construct a client although the
  endpoints are unauthenticated, and the spec fetch could only return JSON;
  `NewDiagnosticsService` needs no credentials, `OpenAPIYAML` / `OpenAPIYML`
  fetch the YAML forms, and a 503 from `/health` maps to `ErrServer`.

### Added

- 4 operations documented in the final 0.9.2 OpenAPI revision, bringing
  coverage to 333 of the 333 operations in the OpenAPI document:
  - `MonitorAlerts.List` / `MonitorAlerts.Fetch`
    (`GET /v1/Alerts`, `GET /v1/Alerts/{Sid}`) — Monitor v1 failure
    events across calls and messaging, with the `/v1` `meta` envelope
    (`Client.NextPage` applies; `LogLevel` / `StartDate` / `EndDate`
    filters)
  - `Messages.CreateFeedback`
    (`POST …/Messages/{Sid}/Feedback.json`) — delivery feedback
    (`Outcome`: `confirmed` / `unconfirmed`)
  - `Accounts.RotateKey` (`POST …/Accounts/{AccountSid}/RotateKey.json`)
    — generates a new API key; the previous key stops working
    immediately. The response (`AccountWithKey`) carries the key exactly
    once and redacts it from `String` / `GoString` output
- 36 operations:
  - `Accounts.Fetch` / `Accounts.Update` (`GET`/`POST /2010-04-01/Accounts/{Sid}.json`;
    a concurrent update returns 409, code 20409: `IsConflict`)
  - `Balance.Fetch` (`GET …/Balance.json`)
  - `OutgoingCallerIDs.List` / `Iterate` / `Create` / `Fetch` / `Update` / `Delete`
  - `Messages.ListMedia` / `IterateMedia` / `FetchMedia` / `DeleteMedia`
  - `Exports.List` / `Exports.Download` (`GET …/Exports/{Date}`, streamed)
  - `Transcriptions.List` / `Iterate` / `Fetch` / `Delete` (account-wide)
  - `Recordings.ListTranscriptions` / `IterateTranscriptions` /
    `FetchTranscription` / `DeleteTranscription`
  - `Conferences.StartRecording` (`POST …/Conferences/{ConferenceSid}/Recordings.json`)
  - `Diagnostics.OpenAPIYAML` / `OpenAPIYML`
  - `UsageRecords.List` plus `ListAllTime`, `ListDaily`, `ListLastMonth`,
    `ListMonthly`, `ListThisMonth`, `ListToday`, `ListYearly`,
    `ListYesterday` (each with an `Iterate…` twin)
  - `SIP.Registrations.List` / `Iterate`
  - `MessagingV1.Services.ListPhoneNumbers` / `CreatePhoneNumber` /
    `FetchPhoneNumber` / `DeletePhoneNumber` (messaging host)
- `Pager[T]` (`Next`, `Current`, `Err`, `PagesFetched`; `All` on Go 1.23+
  for range-over-func) and `Iterate…` methods on the /2010-04-01 lists:
  calls, applications, conferences and participants, queues and members,
  messages and media, notifications, call events, recordings (account, call
  and conference), transcriptions, incoming phone numbers (all four lists),
  outgoing caller IDs, usage records (every window), and SIP domains,
  credential lists and credentials, IP access control lists and addresses,
  and registrations. The SIP domain-mapping lists page with `Client.NextPage`.
- `Client.NextPage` to fetch the page after a list or a `/v1` `Meta`, and
  `HasNextPage` on both envelopes. Next links are followed only on the host
  that served the page (`*PaginationError` otherwise).
- Webhook signature validation: `RequestValidator`, `NewRequestValidator`,
  `ValidateRequest`, `ComputeSignature`, `SignatureHeader`.
- `NewDiagnosticsService` for health checks and spec fetches without
  credentials.
- `RecordingsService.DownloadAudio` and the `Download` type for streamed
  binary responses.
- Error types `ValidationError`, `DecodeError` and `PaginationError`;
  sentinels `ErrMethodNotAllowed`, `ErrTimeout` and `ErrInsecureRedirect`;
  helpers `IsConflict`, `IsTransport` and `IsTimeout`; `MaxRetryAfter`.
- `voiceml.Date` to format date-only filters as `YYYY-MM-DD`.
- `CreateMessageParams.MediaURL`, `ValidityPeriod`, `StatusCallbackMethod`;
  `StartRecordingParams.RecordingTrack`; `ListIncomingPhoneNumbersParams`
  `FriendlyName`, `Beta`, `Origin`; typed `Notification`, `CallEvent`,
  `Account`, `Balance`, `OutgoingCallerID`, `ValidationRequest`,
  `MessageMedia`, `Transcription`, `Export`, `UsageRecord`,
  `SIPRegistration` and `MessagingServicePhoneNumber` models.

### Security

- Credentials are never sent to another host: redirects drop
  `Authorization` once they leave the original scheme and host (including
  subdomains), https→http redirects are refused, and pagination links are
  followed only on the host that served the page.
- `ClientOptions`, `Client` and `RequestValidator` redact the API key in
  `fmt` (`%v`, `%+v`, `%#v`, `%s`), JSON and `log/slog` output.
- The default HTTP client requires TLS 1.2 or later.
- CI adds `gofmt`, `go vet`, `staticcheck` and the race detector, tests the
  `go.mod` floor (Go 1.21) and the current stable Go, and runs gosec
  (advisory, SARIF artifact) and govulncheck (fatal) in
  `.github/workflows/security-scan.yml`.

### Docs

- README rewritten: correct operation count (333 of 333), credentials
  (your VoiceML AccountSid and API key; Twilio-issued credentials do not
  work), the retry and `Retry-After` policy, pagination without `total`,
  webhook validation, downloads, diagnostics without credentials and a
  complete resource table. Every Go snippet is compiled by the test suite,
  and runnable package examples are in `example_test.go`.
- SECURITY.md lists 0.9.x as supported and documents the retry, redirect,
  pagination and redaction behaviour.
- Example phone numbers use the reserved `+1 415 555 01xx` range.
- Customer-facing wording describes the API as Twilio-compatible.
