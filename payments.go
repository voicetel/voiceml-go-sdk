package voiceml

import (
	"context"
	"net/url"
)

// CallPayment is the REST companion to the `<Pay>` TwiML verb. The response
// shape mirrors Twilio's deliberately-minimal payload — runtime config
// (ChargeAmount, PaymentConnector, ValidCardTypes, etc.) is captured
// server-side and not echoed back. Tenant-side BYO is binding: the account
// must have `pay_enabled = true` AND a `stripe_secret_key` set, or the call
// fails 403.
type CallPayment struct {
	Sid         string `json:"sid"`
	AccountSid  string `json:"account_sid"`
	CallSid     string `json:"call_sid"`
	APIVersion  string `json:"api_version"`
	DateCreated string `json:"date_created"`
	DateUpdated string `json:"date_updated"`
	URI         string `json:"uri"`
}

// PaymentBankAccountType narrows the BankAccountType field on a Pay session.
type PaymentBankAccountType string

const (
	PaymentBankAccountTypeConsumerChecking   PaymentBankAccountType = "consumer-checking"
	PaymentBankAccountTypeConsumerSavings    PaymentBankAccountType = "consumer-savings"
	PaymentBankAccountTypeCommercialChecking PaymentBankAccountType = "commercial-checking"
)

// PaymentInput narrows the Input field. DTMF is the only supported value today.
type PaymentInput string

const (
	PaymentInputDTMF PaymentInput = "dtmf"
)

// PaymentMethod narrows the PaymentMethod field.
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit-card"
	PaymentMethodACHDebit   PaymentMethod = "ach-debit"
)

// PaymentTokenType narrows the TokenType field.
type PaymentTokenType string

const (
	PaymentTokenTypeOneTime       PaymentTokenType = "one-time"
	PaymentTokenTypeReusable      PaymentTokenType = "reusable"
	PaymentTokenTypePaymentMethod PaymentTokenType = "payment-method"
)

// PaymentCapture narrows the Capture field on Pay-session updates — tells the
// runtime which input the user is about to type next.
type PaymentCapture string

const (
	PaymentCapturePaymentCardNumber        PaymentCapture = "payment-card-number"
	PaymentCaptureExpirationDate           PaymentCapture = "expiration-date"
	PaymentCaptureSecurityCode             PaymentCapture = "security-code"
	PaymentCapturePostalCode               PaymentCapture = "postal-code"
	PaymentCaptureBankRoutingNumber        PaymentCapture = "bank-routing-number"
	PaymentCaptureBankAccountNumber        PaymentCapture = "bank-account-number"
	PaymentCapturePaymentCardNumberMatcher PaymentCapture = "payment-card-number-matcher"
	PaymentCaptureExpirationDateMatcher    PaymentCapture = "expiration-date-matcher"
	PaymentCaptureSecurityCodeMatcher      PaymentCapture = "security-code-matcher"
	PaymentCapturePostalCodeMatcher        PaymentCapture = "postal-code-matcher"
)

// PaymentSessionStatus narrows the Status field on Pay-session updates.
type PaymentSessionStatus string

const (
	PaymentSessionStatusComplete PaymentSessionStatus = "complete"
	PaymentSessionStatusCancel   PaymentSessionStatus = "cancel"
)

// CreatePaymentParams is the body for POST /Calls/{sid}/Payments. Every
// attribute the `<Pay>` TwiML verb accepts has a counterpart here.
// IdempotencyKey is accepted and persisted for diagnostic visibility but
// replay-dedup is NOT enforced today.
type CreatePaymentParams struct {
	IdempotencyKey        *string                 `form:"IdempotencyKey"`
	StatusCallback        *string                 `form:"StatusCallback"`
	BankAccountType       *PaymentBankAccountType `form:"BankAccountType"`
	ChargeAmount          *string                 `form:"ChargeAmount"`
	Currency              *string                 `form:"Currency"`
	Description           *string                 `form:"Description"`
	Input                 *PaymentInput           `form:"Input"`
	MinPostalCodeLength   *int                    `form:"MinPostalCodeLength"`
	Parameter             *string                 `form:"Parameter"`
	PaymentConnector      *string                 `form:"PaymentConnector"`
	PaymentMethod         *PaymentMethod          `form:"PaymentMethod"`
	PostalCode            *bool                   `form:"PostalCode"`
	SecurityCode          *bool                   `form:"SecurityCode"`
	Timeout               *int                    `form:"Timeout"`
	TokenType             *PaymentTokenType       `form:"TokenType"`
	ValidCardTypes        *string                 `form:"ValidCardTypes"`
	RequireMatchingInputs *string                 `form:"RequireMatchingInputs"`
	Confirmation          *bool                   `form:"Confirmation"`
}

func (p CreatePaymentParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "IdempotencyKey", p.IdempotencyKey)
	setStringP(v, "StatusCallback", p.StatusCallback)
	if p.BankAccountType != nil {
		v.Set("BankAccountType", string(*p.BankAccountType))
	}
	setStringP(v, "ChargeAmount", p.ChargeAmount)
	setStringP(v, "Currency", p.Currency)
	setStringP(v, "Description", p.Description)
	if p.Input != nil {
		v.Set("Input", string(*p.Input))
	}
	setIntP(v, "MinPostalCodeLength", p.MinPostalCodeLength)
	setStringP(v, "Parameter", p.Parameter)
	setStringP(v, "PaymentConnector", p.PaymentConnector)
	if p.PaymentMethod != nil {
		v.Set("PaymentMethod", string(*p.PaymentMethod))
	}
	setBoolP(v, "PostalCode", p.PostalCode)
	setBoolP(v, "SecurityCode", p.SecurityCode)
	setIntP(v, "Timeout", p.Timeout)
	if p.TokenType != nil {
		v.Set("TokenType", string(*p.TokenType))
	}
	setStringP(v, "ValidCardTypes", p.ValidCardTypes)
	setStringP(v, "RequireMatchingInputs", p.RequireMatchingInputs)
	setBoolP(v, "Confirmation", p.Confirmation)
	return v
}

// UpdatePaymentParams is the body for POST /Calls/{sid}/Payments/{sid}.
// Either advance the session (Capture=...) or terminate it (Status=complete
// or Status=cancel).
type UpdatePaymentParams struct {
	IdempotencyKey *string               `form:"IdempotencyKey"`
	StatusCallback *string               `form:"StatusCallback"`
	Capture        *PaymentCapture       `form:"Capture"`
	Status         *PaymentSessionStatus `form:"Status"`
}

func (p UpdatePaymentParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "IdempotencyKey", p.IdempotencyKey)
	setStringP(v, "StatusCallback", p.StatusCallback)
	if p.Capture != nil {
		v.Set("Capture", string(*p.Capture))
	}
	if p.Status != nil {
		v.Set("Status", string(*p.Status))
	}
	return v
}

// StartPayment begins a `<Pay>` session on the live call. Returns 201 with
// the freshly-minted CallPayment. Returns 403 when the tenant is not
// pay_enabled or has no stripe_secret_key configured.
func (s *CallsService) StartPayment(ctx context.Context, callSid string, params CreatePaymentParams) (*CallPayment, error) {
	var out CallPayment
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Payments"),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdatePayment advances or terminates an existing Pay session. Status=complete
// captures the collected fields; Status=cancel aborts the session. Capture=...
// tells the runtime which input the user is about to type next.
func (s *CallsService) UpdatePayment(ctx context.Context, callSid, paymentSid string, params UpdatePaymentParams) (*CallPayment, error) {
	var out CallPayment
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Calls", callSid, "Payments", paymentSid),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
