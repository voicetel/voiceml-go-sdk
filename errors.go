package voiceml

import (
	"errors"
	"fmt"
)

// VoiceMLError is the base interface implemented by every error this SDK
// returns. Callers that want to handle "any SDK error" should type-assert
// against this; callers that need to branch on HTTP status family should
// inspect *APIError or use errors.Is with the ErrXxx sentinels below.
type VoiceMLError interface {
	error
	voicemlError()
}

// ConfigurationError is returned when the client is constructed with missing
// or invalid options (e.g. no AccountSid, negative MaxRetries).
type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string { return e.Message }
func (e *ConfigurationError) voicemlError() {}

// Sentinel error values for use with errors.Is. Every *APIError wraps one of
// these (keyed off HTTP status code) so callers can write:
//
//	if errors.Is(err, voiceml.ErrNotFound) { ... }
var (
	ErrBadRequest       = errors.New("voiceml: bad request")
	ErrAuthentication   = errors.New("voiceml: authentication failed")
	ErrPermissionDenied = errors.New("voiceml: permission denied")
	ErrNotFound         = errors.New("voiceml: not found")
	ErrConflict         = errors.New("voiceml: conflict")
	ErrGone             = errors.New("voiceml: gone")
	ErrRateLimit        = errors.New("voiceml: rate limit exceeded")
	ErrNotImplemented   = errors.New("voiceml: not implemented")
	ErrServer           = errors.New("voiceml: server error")
	ErrTransport        = errors.New("voiceml: transport error")
)

// APIError is the single error type returned for non-2xx responses. The
// HTTP status determines which sentinel (ErrNotFound, ErrRateLimit, ...) it
// wraps; the parsed Twilio-shape body populates Code and Message.
//
// Inspect:
//   - StatusCode  HTTP status code from the server.
//   - Code        Numeric/string code from the body's `code` field (Twilio convention).
//   - Message     Human-readable message (from the body's `message`, or "HTTP <status>").
//   - Body        Raw bytes of the response body, useful for non-JSON failures.
type APIError struct {
	StatusCode int
	Code       any // numeric (int64) or string per Twilio convention; nil if absent
	Message    string
	Body       []byte
	wrapped    error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("voiceml: HTTP %d: %s", e.StatusCode, e.Message)
}

// Unwrap exposes the sentinel (ErrNotFound, ErrRateLimit, ...) so errors.Is works.
func (e *APIError) Unwrap() error { return e.wrapped }

func (e *APIError) voicemlError() {}

// IsNotFound reports whether err is or wraps a 404 APIError.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsRateLimit reports whether err is or wraps a 429 APIError.
func IsRateLimit(err error) bool { return errors.Is(err, ErrRateLimit) }

// IsAuthentication reports whether err is or wraps a 401 APIError.
func IsAuthentication(err error) bool { return errors.Is(err, ErrAuthentication) }

// IsNotImplemented reports whether err is or wraps a 501 APIError.
func IsNotImplemented(err error) bool { return errors.Is(err, ErrNotImplemented) }

// IsServer reports whether err is or wraps a 5xx APIError.
func IsServer(err error) bool { return errors.Is(err, ErrServer) }

// newAPIError builds an *APIError from a parsed response and pins the right
// sentinel onto its Unwrap chain.
func newAPIError(status int, code any, message string, body []byte) *APIError {
	if message == "" {
		message = fmt.Sprintf("HTTP %d", status)
	}
	return &APIError{
		StatusCode: status,
		Code:       code,
		Message:    message,
		Body:       body,
		wrapped:    sentinelFor(status),
	}
}

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
