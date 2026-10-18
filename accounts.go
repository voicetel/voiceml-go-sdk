package voiceml

import (
	"context"
	"fmt"
)

// AccountsService fetches and updates the authenticated account and
// rotates its API key (/2010-04-01/Accounts/{AccountSid}.json and
// …/RotateKey.json). An account can only read and update its own record.
type AccountsService struct{ c *Client }

// AccountStatus is the lifecycle state of an Account.
type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusClosed    AccountStatus = "closed"
)

// Account is the authenticated account. Type is "Trial" or "Full".
type Account struct {
	Sid             string            `json:"sid"`
	FriendlyName    string            `json:"friendly_name"`
	Status          AccountStatus     `json:"status"`
	Type            string            `json:"type"`
	OwnerAccountSid string            `json:"owner_account_sid"`
	DateCreated     string            `json:"date_created"`
	DateUpdated     string            `json:"date_updated"`
	URI             string            `json:"uri"`
	SubresourceURIs map[string]string `json:"subresource_uris,omitempty"`
}

// UpdateAccountParams is the body for POST /Accounts/{AccountSid}.json.
// Only set fields are changed. VoiceMethod / VoiceFallbackMethod take "GET"
// or "POST".
//
// AllowCidr replaces the IP allow-list (one CIDR prefix per entry). A nil
// slice leaves it unchanged; an empty, non-nil slice clears it (no
// restriction).
type UpdateAccountParams struct {
	FriendlyName        *string
	VoiceURL            *string
	VoiceMethod         *string
	VoiceFallbackURL    *string
	VoiceFallbackMethod *string
	SttFsVendor         *string
	SttHTTPURL          *string
	SttHTTPToken        *string
	TtsHTTPURL          *string
	TtsHTTPToken        *string
	TtsEngineName       *string
	TtsDefaultVoice     *string
	TtsDefaultLanguage  *string
	AwsBucket           *string
	AwsRegion           *string
	AwsPrefix           *string
	AwsAccessKeyID      *string
	AwsSecretAccessKey  *string
	AwsKmsKeyID         *string
	AllowCidr           []string
}

// body renders the update as the JSON object the endpoint accepts. JSON
// (rather than a form) is used so an empty AllowCidr list can be expressed.
func (p UpdateAccountParams) body() map[string]any {
	b := map[string]any{}
	put := func(key string, v *string) {
		if v != nil {
			b[key] = *v
		}
	}
	put("FriendlyName", p.FriendlyName)
	put("VoiceUrl", p.VoiceURL)
	put("VoiceMethod", p.VoiceMethod)
	put("VoiceFallbackUrl", p.VoiceFallbackURL)
	put("VoiceFallbackMethod", p.VoiceFallbackMethod)
	put("SttFsVendor", p.SttFsVendor)
	put("SttHttpUrl", p.SttHTTPURL)
	put("SttHttpToken", p.SttHTTPToken)
	put("TtsHttpUrl", p.TtsHTTPURL)
	put("TtsHttpToken", p.TtsHTTPToken)
	put("TtsEngineName", p.TtsEngineName)
	put("TtsDefaultVoice", p.TtsDefaultVoice)
	put("TtsDefaultLanguage", p.TtsDefaultLanguage)
	put("AwsBucket", p.AwsBucket)
	put("AwsRegion", p.AwsRegion)
	put("AwsPrefix", p.AwsPrefix)
	put("AwsAccessKeyId", p.AwsAccessKeyID)
	put("AwsSecretAccessKey", p.AwsSecretAccessKey)
	put("AwsKmsKeyId", p.AwsKmsKeyID)
	if p.AllowCidr != nil {
		b["AllowCidr"] = p.AllowCidr
	}
	return b
}

// Fetch returns the authenticated account. GET /Accounts/{AccountSid}.json.
func (s *AccountsService) Fetch(ctx context.Context) (*Account, error) {
	var out Account
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   joinRoute("/2010-04-01/Accounts", ".json", s.c.AccountSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes the authenticated account and returns the updated record.
// POST /Accounts/{AccountSid}.json.
//
// A concurrent modification is rejected with 409 (code 20409); the error
// satisfies IsConflict / errors.Is(err, ErrConflict). Refetch and retry.
func (s *AccountsService) Update(ctx context.Context, params UpdateAccountParams) (*Account, error) {
	var out Account
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   joinRoute("/2010-04-01/Accounts", ".json", s.c.AccountSid),
		json:   params.body(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AccountWithKey is the authenticated account's configuration view plus
// the freshly generated API key, returned by RotateKey. No read endpoint
// exposes the key again: store APIKey right away — requests signed with
// the previous key fail with 401 from the moment rotation succeeds.
//
// String and GoString redact APIKey; JSON marshalling does not, because
// the response is the only delivery of the key.
type AccountWithKey struct {
	Sid                   string   `json:"sid"`
	APIKey                string   `json:"api_key"`
	SipUsername           string   `json:"sip_username"`
	Name                  *string  `json:"name,omitempty"`
	PhoneRegions          []string `json:"phone_regions"`
	AllowCidr             []string `json:"allow_cidr"`
	Disabled              bool     `json:"disabled"`
	SttFsVendor           *string  `json:"stt_fs_vendor,omitempty"`
	SttHTTPURL            *string  `json:"stt_http_url,omitempty"`
	SttHTTPTokenSet       bool     `json:"stt_http_token_set"`
	TtsHTTPURL            *string  `json:"tts_http_url,omitempty"`
	TtsHTTPTokenSet       bool     `json:"tts_http_token_set"`
	TtsEngineName         *string  `json:"tts_engine_name,omitempty"`
	TtsDefaultVoice       *string  `json:"tts_default_voice,omitempty"`
	TtsDefaultLanguage    *string  `json:"tts_default_language,omitempty"`
	StripeSecretKeySet    bool     `json:"stripe_secret_key_set"`
	PayEnabled            bool     `json:"pay_enabled"`
	SmsProvider           *string  `json:"sms_provider,omitempty"`
	SmsAllowedProviders   []string `json:"sms_allowed_providers"`
	SttAllowedProviders   []string `json:"stt_allowed_providers"`
	SttProvider           *string  `json:"stt_provider,omitempty"`
	VoiceURL              *string  `json:"voice_url,omitempty"`
	VoiceMethod           *string  `json:"voice_method,omitempty"`
	VoiceFallbackURL      *string  `json:"voice_fallback_url,omitempty"`
	VoiceFallbackMethod   *string  `json:"voice_fallback_method,omitempty"`
	AwsBucket             *string  `json:"aws_bucket,omitempty"`
	AwsRegion             *string  `json:"aws_region,omitempty"`
	AwsPrefix             *string  `json:"aws_prefix,omitempty"`
	AwsAccessKeyID        *string  `json:"aws_access_key_id,omitempty"`
	AwsSecretAccessKeySet bool     `json:"aws_secret_access_key_set"`
	AwsKmsKeyID           *string  `json:"aws_kms_key_id,omitempty"`
	CreatedAt             string   `json:"created_at"`
	UpdatedAt             string   `json:"updated_at"`
}

// String implements fmt.Stringer; APIKey is redacted.
func (a AccountWithKey) String() string {
	return fmt.Sprintf("voiceml.AccountWithKey{Sid: %q, APIKey: %q}", a.Sid, redact(a.APIKey))
}

// GoString implements fmt.GoStringer; APIKey is redacted.
func (a AccountWithKey) GoString() string { return a.String() }

// RotateKey generates a new API key for the authenticated account and
// returns the account view plus the new key. POST
// /2010-04-01/Accounts/{AccountSid}/RotateKey.json. It takes no body.
//
// The previous key stops working immediately, and no read endpoint
// exposes the new one — store AccountWithKey.APIKey right away. A 409
// (code 20040, Idempotency-Key reuse) satisfies IsConflict /
// errors.Is(err, ErrConflict).
func (s *AccountsService) RotateKey(ctx context.Context) (*AccountWithKey, error) {
	var out AccountWithKey
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("RotateKey"),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// BalanceService reads the account balance.
type BalanceService struct{ c *Client }

// Balance is the account balance. VoiceML reports a static zero.
type Balance struct {
	AccountSid string `json:"account_sid"`
	Balance    string `json:"balance"`
	Currency   string `json:"currency"`
}

// Fetch returns the account balance. GET /Accounts/{AccountSid}/Balance.json.
func (s *BalanceService) Fetch(ctx context.Context) (*Balance, error) {
	var out Balance
	if err := s.c.t.do(ctx, requestOpts{method: "GET", path: s.c.pathf("Balance")}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
