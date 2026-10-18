package voiceml

import (
	"context"
	"net/url"
)

// MonitorAlertsService exposes the Monitor v1 Alert resource — failure
// events across calls and messaging, rendered in the Alert shape
// (/v1/Alerts). The account is resolved from HTTP Basic auth; no
// AccountSid appears in the path.
type MonitorAlertsService struct{ c *Client }

// MonitorAlert is a Monitor v1 Alert: a call-scoped error or a messaging
// delivery failure. LogLevel is "error" or "warning"; treat unknown
// values as opaque. The request/response detail fields are only carried
// by fetch responses, when the underlying event stored any.
type MonitorAlert struct {
	Sid              string  `json:"sid"`
	AccountSid       string  `json:"account_sid"`
	AlertText        string  `json:"alert_text"`
	APIVersion       *string `json:"api_version"`
	DateCreated      string  `json:"date_created"`
	DateGenerated    string  `json:"date_generated"`
	DateUpdated      string  `json:"date_updated"`
	ErrorCode        string  `json:"error_code"`
	LogLevel         string  `json:"log_level"`
	MoreInfo         string  `json:"more_info"`
	RequestMethod    string  `json:"request_method"`
	RequestURL       string  `json:"request_url"`
	RequestHeaders   *string `json:"request_headers"`
	RequestVariables *string `json:"request_variables"`
	ResourceSid      *string `json:"resource_sid"`
	ResponseBody     *string `json:"response_body"`
	ResponseHeaders  *string `json:"response_headers"`
	ServiceSid       *string `json:"service_sid"`
	URL              string  `json:"url"`
}

// MonitorAlertList is the single-page /v1/Alerts response. Meta carries
// the standard /v1 envelope; pass page.Meta to Client.NextPage to fetch
// a next page when the server links one (today it does not).
type MonitorAlertList struct {
	Alerts []MonitorAlert `json:"alerts"`
	Meta   VoiceV1Meta    `json:"meta"`
}

// ListMonitorAlertsParams narrows GET /v1/Alerts. LogLevel is "error" or
// "warning"; StartDate and EndDate are YYYY-MM-DD dates or full ISO-8601
// timestamps (see Date). Unlike the other /v1 collections, PageSize does
// limit the alerts returned.
type ListMonitorAlertsParams struct {
	LogLevel  *string
	StartDate *string
	EndDate   *string
	V1PageParams
}

func (p ListMonitorAlertsParams) query() url.Values {
	v := p.V1PageParams.query()
	setStringP(v, "LogLevel", p.LogLevel)
	setStringP(v, "StartDate", p.StartDate)
	setStringP(v, "EndDate", p.EndDate)
	return v
}

// List returns a single page of Alerts. GET /v1/Alerts.
func (s *MonitorAlertsService) List(ctx context.Context, params ListMonitorAlertsParams) (*MonitorAlertList, error) {
	var out MonitorAlertList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: apiRoute("/v1", "Alerts"), query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch retrieves one Alert, with the request/response detail the
// underlying event stored. GET /v1/Alerts/{Sid}.
func (s *MonitorAlertsService) Fetch(ctx context.Context, sid string) (*MonitorAlert, error) {
	var out MonitorAlert
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: apiRoute("/v1", "Alerts", sid),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
