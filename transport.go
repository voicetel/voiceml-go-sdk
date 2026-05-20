package voiceml

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the production VoiceML host. There is no separate sandbox
// in the spec.
const DefaultBaseURL = "https://voiceml.voicetel.com"

// DefaultTimeout is the per-request timeout when none is configured.
const DefaultTimeout = 30 * time.Second

// DefaultMaxRetries is the number of retries for 429/5xx and transport errors
// when none is configured. Zero disables retries.
const DefaultMaxRetries = 2

const defaultUserAgent = "voiceml-go/" + Version

// retryableStatuses are the HTTP status codes that trigger a retry pass.
var retryableStatuses = map[int]struct{}{
	429: {}, 500: {}, 502: {}, 503: {}, 504: {},
}

// transport is the internal HTTP plumbing shared by every resource service.
// It is constructed by NewClient and never exposed directly.
type transport struct {
	accountSid string
	apiKey     string
	baseURL    string
	userAgent  string
	maxRetries int
	httpClient *http.Client
}

// requestOpts captures the inputs to a single API call. form / json are
// mutually exclusive; if both are non-nil, json wins.
type requestOpts struct {
	method string
	path   string
	query  url.Values
	form   url.Values
	json   any
}

// do executes a single API request and decodes the JSON body into out (if
// non-nil). Non-2xx responses are mapped to *APIError via the sentinel chain.
func (t *transport) do(ctx context.Context, opts requestOpts, out any) error {
	fullURL := t.baseURL + opts.path
	if len(opts.query) > 0 {
		// Manual encoding so that literal ">=" / "<=" in keys (StartTime>=,
		// StartTime<=) survive — url.Values.Encode would percent-encode the
		// equals sign as expected, but http.Request requires properly-escaped
		// keys too. Use rawQuery from url.Values.Encode; it handles this.
		fullURL += "?" + opts.query.Encode()
	}

	var bodyBytes []byte
	contentType := ""
	if opts.json != nil {
		b, err := json.Marshal(opts.json)
		if err != nil {
			return fmt.Errorf("voiceml: marshal json body: %w", err)
		}
		bodyBytes = b
		contentType = "application/json"
	} else if opts.form != nil {
		bodyBytes = []byte(opts.form.Encode())
		contentType = "application/x-www-form-urlencoded"
	}

	var lastErr error
	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		var body io.Reader
		if len(bodyBytes) > 0 {
			body = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, opts.method, fullURL, body)
		if err != nil {
			return fmt.Errorf("voiceml: build request: %w", err)
		}
		t.applyAuth(req)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", t.userAgent)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		resp, err := t.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if attempt >= t.maxRetries {
				return &APIError{
					StatusCode: 0,
					Message:    fmt.Sprintf("transport error after %d attempts: %s", attempt+1, err),
					wrapped:    errors.Join(ErrTransport, err),
				}
			}
			if sleepErr := sleepBackoff(ctx, attempt, nil); sleepErr != nil {
				return sleepErr
			}
			continue
		}

		if _, retry := retryableStatuses[resp.StatusCode]; retry && attempt < t.maxRetries {
			// Drain body so the connection can be reused.
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if sleepErr := sleepBackoff(ctx, attempt, resp); sleepErr != nil {
				return sleepErr
			}
			continue
		}

		return parseResponse(resp, out)
	}
	// Unreachable: the loop either returns parseResponse or returns from the
	// transport-error branch on the final attempt.
	if lastErr != nil {
		return lastErr
	}
	return errors.New("voiceml: retry loop exhausted with no result")
}

// fetchBytes follows the 302 → S3 redirect that GET /Recordings/{sid}.wav
// issues when audio has been archived. Auth is sent on the initial hop only;
// the S3 presigned URL is fetched without it.
//
// The default http.Client follows redirects and strips Authorization on
// cross-host hops automatically (Go 1.21+), so we don't need to customize
// CheckRedirect here.
func (t *transport) fetchBytes(ctx context.Context, path string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", t.baseURL+path, nil)
	if err != nil {
		return nil, "", fmt.Errorf("voiceml: build request: %w", err)
	}
	t.applyAuth(req)
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", t.userAgent)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, "", &APIError{
			StatusCode: 0,
			Message:    fmt.Sprintf("transport error: %s", err),
			wrapped:    errors.Join(ErrTransport, err),
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", parseError(resp)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("voiceml: read body: %w", err)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	return body, ct, nil
}

// unauthRequest hits an endpoint that the spec marks as `security: []`
// (currently /health and /openapi.json). No auth header, no retries.
func (t *transport) unauthRequest(ctx context.Context, method, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, t.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("voiceml: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", t.userAgent)
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return &APIError{
			StatusCode: 0,
			Message:    fmt.Sprintf("transport error: %s", err),
			wrapped:    errors.Join(ErrTransport, err),
		}
	}
	return parseResponse(resp, out)
}

func (t *transport) applyAuth(req *http.Request) {
	creds := t.accountSid + ":" + t.apiKey
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(creds)))
}

// parseResponse handles both the success and error sides of an HTTP exchange.
// On 2xx with a JSON body, it decodes into out (if non-nil).
func parseResponse(resp *http.Response, out any) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("voiceml: read body: %w", err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out == nil || len(body) == 0 {
			return nil
		}
		if err := json.Unmarshal(body, out); err != nil {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("non-JSON success response: %s", truncate(body)),
				Body:       body,
			}
		}
		return nil
	}
	return decodeError(resp.StatusCode, body)
}

// parseError is the bytes-fetch counterpart to parseResponse — we never want
// to decode the binary body into out, so we always treat the response as an
// error and let decodeError produce the right *APIError.
func parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return decodeError(resp.StatusCode, body)
}

func decodeError(status int, body []byte) error {
	var code any
	message := fmt.Sprintf("HTTP %d", status)
	var moreInfo string

	if len(body) > 0 {
		// Twilio-shape: {"code": <int|string>, "message": <string>, "more_info": <string>, ...}
		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err == nil {
			if raw, ok := parsed["code"]; ok {
				switch v := raw.(type) {
				case float64:
					code = int64(v)
				case string:
					code = v
				}
			}
			if m, ok := parsed["message"].(string); ok && m != "" {
				message = m
			}
			if mi, ok := parsed["more_info"].(string); ok {
				moreInfo = mi
			}
		}
	}
	return newAPIError(status, code, message, moreInfo, body)
}

// sleepBackoff waits before the next retry attempt. Honors a numeric
// Retry-After header when present; otherwise uses exponential backoff capped
// at 8 s. Returns ctx.Err() if the context is canceled mid-wait.
func sleepBackoff(ctx context.Context, attempt int, resp *http.Response) error {
	d := backoffDelay(attempt, resp)
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func backoffDelay(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.ParseFloat(strings.TrimSpace(ra), 64); err == nil && secs > 0 {
				return time.Duration(secs * float64(time.Second))
			}
		}
	}
	d := time.Duration(500*(1<<attempt)) * time.Millisecond
	if d > 8*time.Second {
		d = 8 * time.Second
	}
	return d
}

func truncate(b []byte) string {
	const max = 200
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "..."
}
