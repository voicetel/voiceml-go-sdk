package voiceml

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// SignatureHeader is the request header carrying a webhook's signature.
const SignatureHeader = "X-Twilio-Signature"

// RequestValidator checks the signature VoiceML puts on every webhook it
// sends (status callbacks, TwiML fetches, messaging webhooks) in the
// X-Twilio-Signature header. The signing key is your account API key.
//
// The signature is HMAC-SHA1 over the full webhook URL (without any
// #fragment) followed, for POST, by every form parameter as key+value with
// keys sorted in byte order and repeated keys kept in the order received;
// GET webhooks sign the URL alone. It is Base64 (standard, padded) encoded.
//
// A RequestValidator never includes the key in formatted output.
type RequestValidator struct {
	apiKey *secret
}

// NewRequestValidator returns a validator for webhooks signed with apiKey.
func NewRequestValidator(apiKey string) *RequestValidator {
	return &RequestValidator{apiKey: &secret{value: apiKey}}
}

// String implements fmt.Stringer without exposing the key.
func (v RequestValidator) String() string { return "voiceml.RequestValidator{apiKey: [redacted]}" }

// GoString implements fmt.GoStringer without exposing the key.
func (v RequestValidator) GoString() string { return v.String() }

// Validate reports whether signature is valid for a webhook to rawURL with
// the given POST form params (pass nil for a GET webhook). The comparison
// is constant-time. The URL is also tried with the default port added or
// removed (":443" for https, ":80" for http), since proxies commonly change
// it.
func (v *RequestValidator) Validate(rawURL string, params url.Values, signature string) bool {
	if v.apiKey == nil || v.apiKey.value == "" || signature == "" {
		return false
	}
	ok := false
	for _, candidate := range portVariants(stripFragment(rawURL)) {
		expected := computeSignature(v.apiKey.value, candidate, params)
		if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1 {
			ok = true
		}
	}
	return ok
}

// ValidateHTTPRequest validates an incoming webhook request. publicURL is
// the URL VoiceML called — the one configured as the callback, including
// its query string — which behind a proxy or load balancer differs from
// what r.URL shows. For POST requests the form body is parsed (consuming
// r.Body); the URL query is not part of the signed parameters.
func (v *RequestValidator) ValidateHTTPRequest(r *http.Request, publicURL string) bool {
	var params url.Values
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			return false
		}
		params = r.PostForm
	}
	return v.Validate(publicURL, params, r.Header.Get(SignatureHeader))
}

// ValidateRequest reports whether signature is valid for a webhook to
// rawURL with POST form params (nil for GET), signed with apiKey. See
// RequestValidator.
func ValidateRequest(apiKey, rawURL string, params url.Values, signature string) bool {
	return NewRequestValidator(apiKey).Validate(rawURL, params, signature)
}

// ComputeSignature returns the X-Twilio-Signature value for a webhook to
// rawURL with POST form params (nil for GET), signed with apiKey. Useful for
// testing your webhook handlers.
func ComputeSignature(apiKey, rawURL string, params url.Values) string {
	return computeSignature(apiKey, stripFragment(rawURL), params)
}

func computeSignature(apiKey, rawURL string, params url.Values) string {
	var b strings.Builder
	b.WriteString(rawURL)
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, val := range params[k] {
			b.WriteString(k)
			b.WriteString(val)
		}
	}
	mac := hmac.New(sha1.New, []byte(apiKey))
	mac.Write([]byte(b.String()))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func stripFragment(raw string) string {
	if i := strings.IndexByte(raw, '#'); i >= 0 {
		return raw[:i]
	}
	return raw
}

// portVariants returns raw plus raw with the scheme's default port added
// or removed. Everything other than the authority is kept byte-for-byte.
func portVariants(raw string) []string {
	out := []string{raw}
	sep := strings.Index(raw, "://")
	if sep < 0 {
		return out
	}
	scheme := strings.ToLower(raw[:sep])
	defaultPort := map[string]string{"https": "443", "http": "80"}[scheme]
	if defaultPort == "" {
		return out
	}
	rest := raw[sep+3:]
	end := strings.IndexAny(rest, "/?")
	if end < 0 {
		end = len(rest)
	}
	authority, tail := rest[:end], rest[end:]
	userinfo := ""
	if at := strings.LastIndexByte(authority, '@'); at >= 0 {
		userinfo, authority = authority[:at+1], authority[at+1:]
	}
	host, port, err := net.SplitHostPort(authority)
	var variant string
	switch {
	case err == nil && port == defaultPort:
		if strings.Contains(host, ":") {
			host = "[" + host + "]"
		}
		variant = host
	case err == nil:
		return out // a non-default port is significant
	default:
		h := strings.TrimSuffix(strings.TrimPrefix(authority, "["), "]")
		variant = net.JoinHostPort(h, defaultPort)
	}
	return append(out, raw[:sep+3]+userinfo+variant+tail)
}
