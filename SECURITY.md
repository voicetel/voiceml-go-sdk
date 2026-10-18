# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| 0.9.x   | ✅        |
| < 0.9   | ❌        |

## Reporting a Vulnerability

Please **do not** open a public issue for security vulnerabilities.

Use GitHub's private vulnerability reporting for this repository:
**Security → Report a vulnerability** (or
<https://github.com/voicetel/voiceml-go-sdk/security/advisories/new>).

Include, where possible:

- A description of the issue and its impact
- Steps to reproduce or a proof of concept
- Affected version(s) and configuration

You can expect an acknowledgement within a few business days. Please
allow reasonable time for a fix before any public disclosure.

## Scope Notes

This SDK constructs authenticated HTTP requests to the VoiceML REST
API. What it does, and what remains the consumer's responsibility:

- **Credentials.** The AccountSid and API key travel only in the HTTP
  Basic `Authorization` header. `ClientOptions`, `Client` and
  `RequestValidator` redact the key in `fmt`, JSON and `log/slog`
  output. Do not log the `Authorization` header yourself, and keep the
  key out of source control (load it from a secret manager or the
  environment).
- **Redirects.** Recording audio and exports answer with a redirect to a
  presigned storage URL. The SDK drops `Authorization` as soon as a
  redirect leaves the original scheme and host, refuses https→http
  redirects, and stops after 10 hops. If you supply your own
  `HTTPClient`, the SDK applies this policy to a copy of it; your own
  `CheckRedirect` still runs afterwards.
- **Pagination links** are followed only on the scheme and host that
  served the page, so a crafted `next_page_uri` / `next_page_url` cannot
  send credentials to another host.
- **Retries.** GET, HEAD, PUT and DELETE are retried on 429, 500, 502,
  503, 504 and network errors. POST is retried only on 429 and on
  connection failures where the request never left the client (DNS
  failure, connection refused, connect timeout, TLS handshake failure);
  it is never re-sent after a 5xx, a read timeout or a reset, so a call
  or message is not duplicated by the SDK. `Retry-After` is honoured up
  to 60 seconds; a longer delay is returned to the caller as an error.
  The default is two retries; `MaxRetries: voiceml.Int(0)` disables them.
- **Path parameters** are percent-encoded, and empty, `.` or `..`
  segments are rejected, so caller-supplied SIDs cannot rewrite the
  request path or query.
- **Response bodies.** JSON responses are read up to 32 MiB; larger
  bodies are rejected. Binary downloads are streamed (`DownloadAudio`,
  `Exports.Download`).
- **Webhooks.** Verify inbound webhooks with `RequestValidator` (or
  `ValidateRequest`), which compares signatures in constant time. The
  signature scheme is HMAC-SHA1, as the server defines it.
- **TLS.** The default client uses `crypto/tls` defaults with a
  TLS 1.2 minimum. If you supply a custom `HTTPClient`, you are
  responsible for matching that posture.

Out of scope: vulnerabilities in `github.com/voicetel/voiceml-go-sdk`
caused by a forked / vendored copy that has been modified.
