package voiceml

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// DefaultBaseURL is the production VoiceML API host. There is no separate
// sandbox in the spec.
const DefaultBaseURL = "https://voiceml.voicetel.com"

// DefaultTimeout is the per-request timeout when none is configured.
const DefaultTimeout = 30 * time.Second

// DefaultMaxRetries is the number of retries when none is configured. See
// ClientOptions.MaxRetries for which failures are retried. Zero disables
// retries.
const DefaultMaxRetries = 2

// MaxRetryAfter is the longest Retry-After delay the SDK will wait out. A
// 429 or 503 asking for longer is returned to the caller immediately, with
// the requested delay in APIError.RetryAfter.
const MaxRetryAfter = 60 * time.Second

const defaultUserAgent = "voiceml-go/" + Version

const (
	// maxJSONBody bounds how much of a JSON response the SDK will buffer.
	maxJSONBody = 32 << 20
	// maxErrorBody bounds how much of an error response is kept in
	// APIError.Body.
	maxErrorBody = 1 << 20
	// maxDecodeErrorBody bounds DecodeError.Body.
	maxDecodeErrorBody = 64 << 10
	// maxRedirects bounds redirect chains (matching net/http's default).
	maxRedirects = 10

	backoffBase = 500 * time.Millisecond
	backoffMax  = 8 * time.Second
)

// transport is the internal HTTP plumbing shared by every resource service.
// It is constructed by NewClient and never exposed directly.
type transport struct {
	accountSid string
	// apiKey sits behind a pointer so that even reflective formatting of a
	// transport (which bypasses String) prints an address, not the key.
	apiKey     *secret
	baseURL    string
	userAgent  string
	maxRetries int
	httpClient *http.Client

	// sleep waits between attempts; tests replace it to observe delays.
	sleep func(ctx context.Context, d time.Duration) error
	// now is the clock used to evaluate HTTP-date Retry-After values.
	now func() time.Time
}

// secret holds a credential.
type secret struct{ value string }

// String keeps the API key out of fmt output.
func (t *transport) String() string {
	return fmt.Sprintf("voiceml.transport{baseURL: %q, accountSid: %q, apiKey: [redacted]}", t.baseURL, t.accountSid)
}

// GoString keeps the API key out of %#v output.
func (t *transport) GoString() string { return t.String() }

// route is an encoded request path, or the reason one could not be built.
type route struct {
	path string
	err  error
}

// requestOpts captures the inputs to a single API call. form / json are
// mutually exclusive; if both are non-nil, json wins.
type requestOpts struct {
	method string
	path   route
	// rawURL, when set, is requested verbatim instead of baseURL+path
	// (pagination links).
	rawURL  string
	query   url.Values
	form    url.Values
	json    any
	accept  string
	noAuth  bool
	noRetry bool
}

// withBase returns a copy of t that sends to baseURL.
func (t *transport) withBase(baseURL string) *transport {
	cp := *t
	cp.baseURL = baseURL
	return &cp
}

// do executes a single API call and decodes the JSON body into out (if
// non-nil). Non-2xx responses are mapped to *APIError via the sentinel chain.
func (t *transport) do(ctx context.Context, opts requestOpts, out any) error {
	resp, err := t.send(ctx, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := readAtMost(resp.Body, maxJSONBody)
	if err != nil {
		return transportError(fmt.Errorf("read response body: %w", err), 1)
	}
	if out == nil || len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return &DecodeError{StatusCode: resp.StatusCode, Body: truncateBytes(body, maxDecodeErrorBody), Err: err}
	}
	stampOrigin(out, resp.Request.URL)
	return nil
}

// Download is a streamed binary response (recording audio, account
// exports, message media). The caller must Close Body.
type Download struct {
	// Body streams the payload. It must be closed.
	Body io.ReadCloser
	// ContentType is the declared media type ("audio/wav", ...), or
	// "application/octet-stream" when the server sent none.
	ContentType string
	// ContentLength is the declared size in bytes, or -1 when unknown.
	ContentLength int64
}

// download performs an authenticated GET and returns the streamed body.
// Redirects (for example the 302 to an S3 presigned URL) are followed under
// the client's redirect policy: Authorization is dropped as soon as the
// scheme or host changes, and https→http downgrades are refused.
func (t *transport) download(ctx context.Context, opts requestOpts) (*Download, error) {
	resp, err := t.send(ctx, opts)
	if err != nil {
		return nil, err
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &Download{Body: resp.Body, ContentType: ct, ContentLength: resp.ContentLength}, nil
}

// send runs the request with the retry policy and returns the 2xx response
// with an unread body. Every other outcome is returned as an error with the
// response body consumed and closed.
//
// Retry policy:
//   - GET, HEAD, PUT, DELETE (idempotent) retry on 429, 500, 502, 503, 504
//     and on transport errors.
//   - POST and PATCH retry only on 429 and on connection-establishment
//     failures where the request provably never left the client (DNS
//     failure, connection refused, connect timeout, TLS handshake failure).
//   - A Retry-After header (delta-seconds or HTTP-date) sets the delay; one
//     above MaxRetryAfter ends the retries and the error is returned at once.
//   - Otherwise the delay is exponential with full jitter (0.5 s base,
//     8 s cap).
func (t *transport) send(ctx context.Context, opts requestOpts) (*http.Response, error) {
	if !opts.noAuth && (t.accountSid == "" || t.apiKey == nil || t.apiKey.value == "") {
		return nil, &ConfigurationError{Message: "this client has no credentials; it can only call Diagnostics (construct it with NewClient to make authenticated calls)"}
	}
	if opts.path.err != nil {
		return nil, opts.path.err
	}
	fullURL := opts.rawURL
	if fullURL == "" {
		fullURL = t.baseURL + opts.path.path
		if len(opts.query) > 0 {
			fullURL += "?" + opts.query.Encode()
		}
	}

	var bodyBytes []byte
	contentType := ""
	switch {
	case opts.json != nil:
		b, err := json.Marshal(opts.json)
		if err != nil {
			return nil, &ValidationError{Message: fmt.Sprintf("encode JSON body: %v", err)}
		}
		bodyBytes = b
		contentType = "application/json"
	case opts.form != nil:
		bodyBytes = []byte(opts.form.Encode())
		contentType = "application/x-www-form-urlencoded"
	}
	accept := opts.accept
	if accept == "" {
		accept = "application/json"
	}

	maxRetries := t.maxRetries
	if opts.noRetry {
		maxRetries = 0
	}
	idempotent := isIdempotent(opts.method)

	for attempt := 0; ; attempt++ {
		var body io.Reader
		if bodyBytes != nil {
			body = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, opts.method, fullURL, body)
		if err != nil {
			return nil, &ValidationError{Message: fmt.Sprintf("build request: %v", err)}
		}
		if !opts.noAuth {
			t.applyAuth(req)
		}
		req.Header.Set("Accept", accept)
		req.Header.Set("User-Agent", t.userAgent)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		resp, err := t.httpClient.Do(req)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			retryable := !errors.Is(err, ErrInsecureRedirect) && (idempotent || requestNotSent(err))
			if !retryable || attempt >= maxRetries {
				return nil, transportError(err, attempt+1)
			}
			if err := t.sleep(ctx, backoff(attempt)); err != nil {
				return nil, err
			}
			continue
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		apiErr := t.responseError(resp)
		if attempt >= maxRetries || !retryableStatus(resp.StatusCode, idempotent) {
			return nil, apiErr
		}
		delay := backoff(attempt)
		if _, ok := resp.Header["Retry-After"]; ok {
			if ra, ok := parseRetryAfter(resp.Header.Get("Retry-After"), t.now()); ok {
				if ra > MaxRetryAfter {
					return nil, apiErr
				}
				delay = ra
			}
		}
		if err := t.sleep(ctx, delay); err != nil {
			return nil, err
		}
	}
}

// responseError reads (a bounded prefix of) a non-2xx response, closes it,
// and maps it to *APIError.
func (t *transport) responseError(resp *http.Response) *APIError {
	body, _ := readAtMost(resp.Body, maxErrorBody)
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	_ = resp.Body.Close()
	e := decodeError(resp.StatusCode, body)
	if ra, ok := parseRetryAfter(resp.Header.Get("Retry-After"), t.now()); ok {
		e.RetryAfter = ra
	}
	return e
}

func (t *transport) applyAuth(req *http.Request) {
	creds := t.accountSid + ":" + t.apiKey.value
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(creds)))
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	}
	return false
}

func retryableStatus(status int, idempotent bool) bool {
	switch status {
	case http.StatusTooManyRequests:
		return true
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return idempotent
	}
	return false
}

// requestNotSent reports whether err happened while establishing the
// connection, before any request bytes could reach the server.
func requestNotSent(err error) bool {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" {
		return true
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		return true
	}
	var recordErr tls.RecordHeaderError
	if errors.As(err, &recordErr) {
		return true
	}
	var alertErr tls.AlertError
	return errors.As(err, &alertErr)
}

func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// transportError wraps a request that produced no HTTP response.
func transportError(err error, attempts int) *APIError {
	chain := []error{ErrTransport}
	if isTimeout(err) {
		chain = append(chain, ErrTimeout)
	}
	chain = append(chain, err)
	noun := "attempt"
	if attempts != 1 {
		noun = "attempts"
	}
	return &APIError{
		Message: fmt.Sprintf("transport error after %d %s: %v", attempts, noun, err),
		wrapped: errors.Join(chain...),
	}
}

// parseRetryAfter accepts delta-seconds ("120", or a decimal such as "1.5")
// and HTTP-dates. Negative, non-numeric and non-finite values are ignored.
func parseRetryAfter(v string, now time.Time) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	if isDecimal(v) {
		secs, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		const maxSecs = float64(365 * 24 * 3600)
		if secs > maxSecs {
			secs = maxSecs
		}
		return time.Duration(secs * float64(time.Second)), true
	}
	when, err := http.ParseTime(v)
	if err != nil {
		return 0, false
	}
	d := when.Sub(now)
	if d < 0 {
		d = 0
	}
	return d, true
}

func isDecimal(s string) bool {
	dot := false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c >= '0' && c <= '9':
		case c == '.' && !dot && i > 0 && i < len(s)-1:
			dot = true
		default:
			return false
		}
	}
	return true
}

// backoff is the full-jitter delay before retry attempt+1: uniform in
// [0, min(backoffMax, backoffBase*2^attempt)].
func backoff(attempt int) time.Duration {
	ceiling := backoffMax
	if attempt < 5 {
		ceiling = backoffBase << uint(attempt)
		if ceiling > backoffMax {
			ceiling = backoffMax
		}
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(ceiling)+1))
	if err != nil {
		return ceiling
	}
	return time.Duration(n.Int64())
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
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

// decodeError maps a non-2xx body to *APIError. Twilio-compatible envelope:
// {"code": <int>, "message": <string>, "more_info": <string>, "status": <int>}.
// A body that is not JSON still yields an *APIError carrying the raw bytes.
func decodeError(status int, body []byte) *APIError {
	e := &APIError{StatusCode: status, Body: body, wrapped: sentinelFor(status)}
	var parsed map[string]any
	if len(body) > 0 && json.Unmarshal(body, &parsed) == nil {
		e.Code = intField(parsed["code"])
		e.Status = intField(parsed["status"])
		if m, ok := parsed["message"].(string); ok {
			e.Message = m
		}
		if mi, ok := parsed["more_info"].(string); ok {
			e.MoreInfo = mi
		}
	}
	if e.Message == "" {
		e.Message = fmt.Sprintf("HTTP %d", status)
		if text := strings.TrimSpace(string(truncateBytes(body, 200))); text != "" && parsed == nil {
			e.Message += ": " + text
		}
	}
	return e
}

// intField reads a JSON number, or a string holding an integer.
func intField(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(x)); err == nil {
			return n
		}
	}
	return 0
}

// readAtMost reads r fully, failing if it holds more than limit bytes.
func readAtMost(r io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return b, err
	}
	if int64(len(b)) > limit {
		return b[:limit], fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return b, nil
}

func truncateBytes(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[:n]
}

// newHTTPClient returns the *http.Client the SDK sends with: a copy of the
// caller's client (or a fresh one with timeout) whose redirect policy drops
// Authorization once a redirect leaves the original scheme+host and refuses
// https→http downgrades. The caller's own CheckRedirect still runs after.
func newHTTPClient(user *http.Client, timeout time.Duration) *http.Client {
	var hc http.Client
	if user != nil {
		hc = *user
	} else {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		tr.TLSClientConfig.ClientSessionCache = tls.NewLRUClientSessionCache(0)
		hc = http.Client{Timeout: timeout, Transport: tr}
	}
	hc.CheckRedirect = redirectPolicy(hc.CheckRedirect)
	return &hc
}

func redirectPolicy(next func(*http.Request, []*http.Request) error) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("voiceml: stopped after %d redirects", maxRedirects)
		}
		prev := via[len(via)-1].URL
		if strings.EqualFold(prev.Scheme, "https") && !strings.EqualFold(req.URL.Scheme, "https") {
			return fmt.Errorf("%w from %s to %s://%s", ErrInsecureRedirect, prev.Host, req.URL.Scheme, req.URL.Host)
		}
		origin := via[0].URL
		hops := append(append([]*http.Request{}, via[1:]...), req)
		for _, hop := range hops {
			if !sameOrigin(origin, hop.URL) {
				req.Header.Del("Authorization")
				break
			}
		}
		if next != nil {
			return next(req, via)
		}
		return nil
	}
}

// sameOrigin compares scheme, host and effective port.
func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) &&
		strings.EqualFold(a.Hostname(), b.Hostname()) &&
		effectivePort(a) == effectivePort(b)
}

func effectivePort(u *url.URL) string {
	if p := u.Port(); p != "" {
		return p
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return "443"
	case "http":
		return "80"
	}
	return ""
}
