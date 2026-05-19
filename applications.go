package voiceml

import (
	"context"
	"net/url"
)

// ApplicationsService is the /Applications resource group — CRUD on stored
// TwiML+callback bundles dispatched by `<Dial><Application>`.
type ApplicationsService struct{ c *Client }

// Application is a stored TwiML+callback bundle.
type Application struct {
	Sid                  string `json:"sid"`
	AccountSid           string `json:"account_sid"`
	FriendlyName         string `json:"friendly_name"`
	APIVersion           string `json:"api_version"`
	VoiceURL             string `json:"voice_url"`
	VoiceMethod          string `json:"voice_method,omitempty"`
	VoiceFallbackURL     string `json:"voice_fallback_url,omitempty"`
	VoiceFallbackMethod  string `json:"voice_fallback_method,omitempty"`
	VoiceCallerIDLookup  bool   `json:"voice_caller_id_lookup"`
	StatusCallback       string `json:"status_callback,omitempty"`
	StatusCallbackMethod string `json:"status_callback_method,omitempty"`
	StatusCallbackEvent  string `json:"status_callback_event,omitempty"`
	DateCreated          string `json:"date_created"`
	DateUpdated          string `json:"date_updated"`
	URI                  string `json:"uri"`
}

// ApplicationList is the paginated /Applications list response.
type ApplicationList struct {
	Page
	Applications []Application `json:"applications"`
}

// ApplicationParams is shared between create and update — every field
// optional, only set ones are sent. For create, the spec does not require
// any particular field; the server applies defaults.
type ApplicationParams struct {
	FriendlyName         *string `form:"FriendlyName"`
	VoiceURL             *string `form:"VoiceUrl"`
	VoiceMethod          *string `form:"VoiceMethod"`
	VoiceFallbackURL     *string `form:"VoiceFallbackUrl"`
	VoiceFallbackMethod  *string `form:"VoiceFallbackMethod"`
	VoiceCallerIDLookup  *bool   `form:"VoiceCallerIdLookup"`
	StatusCallback       *string `form:"StatusCallback"`
	StatusCallbackMethod *string `form:"StatusCallbackMethod"`
	StatusCallbackEvent  *string `form:"StatusCallbackEvent"`
}

func (p ApplicationParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "FriendlyName", p.FriendlyName)
	setStringP(v, "VoiceUrl", p.VoiceURL)
	setStringP(v, "VoiceMethod", p.VoiceMethod)
	setStringP(v, "VoiceFallbackUrl", p.VoiceFallbackURL)
	setStringP(v, "VoiceFallbackMethod", p.VoiceFallbackMethod)
	setBoolP(v, "VoiceCallerIdLookup", p.VoiceCallerIDLookup)
	setStringP(v, "StatusCallback", p.StatusCallback)
	setStringP(v, "StatusCallbackMethod", p.StatusCallbackMethod)
	setStringP(v, "StatusCallbackEvent", p.StatusCallbackEvent)
	return v
}

// Create makes a new application. POST /Applications.
func (s *ApplicationsService) Create(ctx context.Context, params ApplicationParams) (*Application, error) {
	var out Application
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Applications"),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all applications for this account. GET /Applications.
func (s *ApplicationsService) List(ctx context.Context) (*ApplicationList, error) {
	var out ApplicationList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Applications"),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches an application by SID. GET /Applications/{sid}.
func (s *ApplicationsService) Get(ctx context.Context, applicationSid string) (*Application, error) {
	var out Application
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Applications", applicationSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes one or more fields on an existing application.
// POST /Applications/{sid}. Only set fields are touched.
func (s *ApplicationsService) Update(ctx context.Context, applicationSid string, params ApplicationParams) (*Application, error) {
	var out Application
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Applications", applicationSid),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes an application. DELETE /Applications/{sid}.
func (s *ApplicationsService) Delete(ctx context.Context, applicationSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("Applications", applicationSid),
	}, nil)
}
