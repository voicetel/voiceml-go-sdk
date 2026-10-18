package voiceml

import (
	"errors"
	"fmt"
	"time"
)

// VoiceMLError is the base interface implemented by every error this SDK
// returns. Callers that want to handle "any SDK error" should type-assert
// against this; callers that need to branch on HTTP status family should
// inspect *APIError or use errors.Is with the ErrXxx sentinels below.
//
// Context cancellation is the one exception: when the caller's context is
// canceled or its deadline passes, the SDK returns ctx.Err() unchanged.
type VoiceMLError interface {
	error
	voicemlError()
}

// ConfigurationError is returned when the client is constructed with missing
// or invalid options (e.g. no AccountSid, negative MaxRetries, a BaseURL that
// is not an absolute http(s) URL), or when a credential-less diagnostics
// client is asked to make an authenticated call.
type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string { return "voiceml: " + e.Message }
func (e *ConfigurationError) voicemlError() {}

// ValidationError is returned before any request is sent when a call's
// arguments cannot form a valid request: an empty, "." or ".." path segment
// (for example an empty SID), or a body the API would reject outright (a
// message with neither Body nor MediaURL).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return "voiceml: " + e.Message }
func (e *ValidationError) voicemlError() {}

// DecodeError is returned when a 2xx response body could not be decoded into
// the result type. Err is the underlying encoding/json error, so
// errors.As(err, new(*json.UnmarshalTypeError)) works; Body holds up to the
// first 64 KiB of the response.
type DecodeError struct {
	StatusCode int
	Body       []byte
	Err        error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("voiceml: could not decode HTTP %d response: %v", e.StatusCode, e.Err)
}

// Unwrap exposes the underlying decode error.
func (e *DecodeError) Unwrap() error { return e.Err }

func (e *DecodeError) voicemlError() {}

// PaginationError is returned when a page's next link cannot be followed:
// it points at a host other than the one that served the page (the SDK never
// sends credentials elsewhere), it is not a URL the SDK can resolve, or it
// repeats the page that was just fetched.
type PaginationError struct {
	Message string
}

func (e *PaginationError) Error() string { return "voiceml: " + e.Message }
func (e *PaginationError) voicemlError() {}

// Sentinel error values for use with errors.Is. Every *APIError wraps one of
// these (keyed off HTTP status code) so callers can write:
//
//	if errors.Is(err, voiceml.ErrNotFound) { ... }
var (
	ErrBadRequest       = errors.New("voiceml: bad request")
	ErrAuthentication   = errors.New("voiceml: authentication failed")
	ErrPermissionDenied = errors.New("voiceml: permission denied")
	ErrNotFound         = errors.New("voiceml: not found")
	ErrMethodNotAllowed = errors.New("voiceml: method not allowed")
	ErrConflict         = errors.New("voiceml: conflict")
	ErrGone             = errors.New("voiceml: gone")
	ErrRateLimit        = errors.New("voiceml: rate limit exceeded")
	ErrNotImplemented   = errors.New("voiceml: not implemented")
	ErrServer           = errors.New("voiceml: server error")

	// ErrTransport marks a request that failed without an HTTP response
	// (DNS, connect, TLS, reset, timeout). The underlying net/http error is
	// also in the chain.
	ErrTransport = errors.New("voiceml: transport error")
	// ErrTimeout additionally marks transport errors caused by a timeout
	// (ClientOptions.Timeout or a net.Error reporting Timeout()).
	ErrTimeout = errors.New("voiceml: request timed out")
	// ErrInsecureRedirect marks a redirect the SDK refused to follow because
	// it downgraded from https to http.
	ErrInsecureRedirect = errors.New("voiceml: refused insecure redirect")
)

// APIError is the error type returned for non-2xx responses and for
// transport failures. The HTTP status determines which sentinel
// (ErrNotFound, ErrRateLimit, ...) it wraps; the parsed Twilio-compatible
// error envelope populates Code, Message, MoreInfo and Status.
//
// Inspect:
//   - StatusCode  HTTP status code from the server; 0 for transport failures.
//   - Code        Numeric error code from the body's `code` field; 0 if absent.
//   - Message     Human-readable message (the body's `message`, or "HTTP <status>").
//   - MoreInfo    Documentation URL from the body's `more_info` field. Empty if absent.
//   - Status      The body's `status` field (normally equal to StatusCode); 0 if absent.
//   - RetryAfter  Delay requested by a Retry-After header (delta-seconds or
//     HTTP-date); 0 when the response carried none.
//   - Body        Raw bytes of the response body (up to 1 MiB), useful for
//     non-JSON failures.
type APIError struct {
	StatusCode int
	Code       int
	Message    string
	MoreInfo   string
	Status     int
	RetryAfter time.Duration
	Body       []byte
	wrapped    error
}

func (e *APIError) Error() string {
	if e.StatusCode == 0 {
		return "voiceml: " + e.Message
	}
	if e.Code != 0 {
		return fmt.Sprintf("voiceml: HTTP %d (code %d): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("voiceml: HTTP %d: %s", e.StatusCode, e.Message)
}

// Unwrap exposes the sentinel (ErrNotFound, ErrRateLimit, ...) so errors.Is
// works. For transport failures it also exposes the underlying error.
func (e *APIError) Unwrap() error { return e.wrapped }

func (e *APIError) voicemlError() {}

// IsNotFound reports whether err is or wraps a 404 APIError.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsConflict reports whether err is or wraps a 409 APIError (for example the
// optimistic-lock rejection, code 20409, from Accounts.Update).
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }

// IsRateLimit reports whether err is or wraps a 429 APIError.
func IsRateLimit(err error) bool { return errors.Is(err, ErrRateLimit) }

// IsAuthentication reports whether err is or wraps a 401 APIError.
func IsAuthentication(err error) bool { return errors.Is(err, ErrAuthentication) }

// IsNotImplemented reports whether err is or wraps a 501 APIError.
func IsNotImplemented(err error) bool { return errors.Is(err, ErrNotImplemented) }

// IsServer reports whether err is or wraps a 5xx APIError.
func IsServer(err error) bool { return errors.Is(err, ErrServer) }

// IsTransport reports whether err is a request that failed without an HTTP
// response (DNS, connect, TLS, reset, timeout).
func IsTransport(err error) bool { return errors.Is(err, ErrTransport) }

// IsTimeout reports whether err is a transport failure caused by a timeout.
func IsTimeout(err error) bool { return errors.Is(err, ErrTimeout) }

func sentinelFor(status int) error {
	switch status {
	case 400:
		return ErrBadRequest
	case 401:
		return ErrAuthentication
	case 403:
		return ErrPermissionDenied
	case 404:
		return ErrNotFound
	case 405:
		return ErrMethodNotAllowed
	case 409:
		return ErrConflict
	case 410:
		return ErrGone
	case 429:
		return ErrRateLimit
	case 501:
		return ErrNotImplemented
	}
	if status >= 500 && status < 600 {
		return ErrServer
	}
	return nil
}
