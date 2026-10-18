package voiceml

import (
	"context"
	"net/url"
)

// OutgoingCallerIDsService is the /OutgoingCallerIds resource group:
// verified numbers the account may present as caller ID.
type OutgoingCallerIDsService struct{ c *Client }

// OutgoingCallerID is a verified caller ID (PN...).
type OutgoingCallerID struct {
	Sid          string `json:"sid"`
	AccountSid   string `json:"account_sid"`
	FriendlyName string `json:"friendly_name"`
	PhoneNumber  string `json:"phone_number"`
	DateCreated  string `json:"date_created"`
	DateUpdated  string `json:"date_updated"`
	URI          string `json:"uri"`
}

// OutgoingCallerIDList is the paginated /OutgoingCallerIds response.
type OutgoingCallerIDList struct {
	Page
	OutgoingCallerIDs []OutgoingCallerID `json:"outgoing_caller_ids"`
}

// ValidationRequest is returned by Create: the server places a verification
// call to PhoneNumber, and the person answering enters ValidationCode.
type ValidationRequest struct {
	AccountSid     string `json:"account_sid"`
	PhoneNumber    string `json:"phone_number"`
	FriendlyName   string `json:"friendly_name"`
	ValidationCode string `json:"validation_code"`
	CallSid        string `json:"call_sid"`
}

// ListOutgoingCallerIDsParams are the filter / pagination query params for
// GET /OutgoingCallerIds.
type ListOutgoingCallerIDsParams struct {
	PhoneNumber  string
	FriendlyName string
	Page         *int
	PageSize     *int
	PageToken    string
}

func (p ListOutgoingCallerIDsParams) query() url.Values {
	v := url.Values{}
	setString(v, "PhoneNumber", p.PhoneNumber)
	setString(v, "FriendlyName", p.FriendlyName)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	setString(v, "PageToken", p.PageToken)
	return v
}

// CreateOutgoingCallerIDParams is the body for POST /OutgoingCallerIds.
// PhoneNumber (E.164) is required.
type CreateOutgoingCallerIDParams struct {
	PhoneNumber  string  `form:"PhoneNumber"`
	FriendlyName *string `form:"FriendlyName"`
}

func (p CreateOutgoingCallerIDParams) form() url.Values {
	v := url.Values{}
	v.Set("PhoneNumber", p.PhoneNumber)
	setStringP(v, "FriendlyName", p.FriendlyName)
	return v
}

// UpdateOutgoingCallerIDParams is the body for POST /OutgoingCallerIds/{Sid}.
type UpdateOutgoingCallerIDParams struct {
	FriendlyName *string `form:"FriendlyName"`
}

func (p UpdateOutgoingCallerIDParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "FriendlyName", p.FriendlyName)
	return v
}

// List returns a page of verified caller IDs. GET /OutgoingCallerIds.
func (s *OutgoingCallerIDsService) List(ctx context.Context, params ListOutgoingCallerIDsParams) (*OutgoingCallerIDList, error) {
	var out OutgoingCallerIDList
	if err := s.c.t.do(ctx, requestOpts{method: "GET", path: s.c.pathf("OutgoingCallerIds"), query: params.query()}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Iterate walks every verified caller ID lazily.
func (s *OutgoingCallerIDsService) Iterate(ctx context.Context, params ListOutgoingCallerIDsParams) *Pager[OutgoingCallerID] {
	return newPager(ctx, s.c, s.c.t, requestOpts{method: "GET", path: s.c.pathf("OutgoingCallerIds"), query: params.query()},
		func(l *OutgoingCallerIDList) ([]OutgoingCallerID, PageLinks) { return l.OutgoingCallerIDs, l.Page })
}

// Create starts verification of a caller ID. POST /OutgoingCallerIds.
func (s *OutgoingCallerIDsService) Create(ctx context.Context, params CreateOutgoingCallerIDParams) (*ValidationRequest, error) {
	var out ValidationRequest
	if err := s.c.t.do(ctx, requestOpts{method: "POST", path: s.c.pathf("OutgoingCallerIds"), form: params.form()}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Fetch returns a verified caller ID. GET /OutgoingCallerIds/{Sid}.
func (s *OutgoingCallerIDsService) Fetch(ctx context.Context, sid string) (*OutgoingCallerID, error) {
	var out OutgoingCallerID
	if err := s.c.t.do(ctx, requestOpts{method: "GET", path: s.c.pathf("OutgoingCallerIds", sid)}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update renames a verified caller ID. POST /OutgoingCallerIds/{Sid}.
func (s *OutgoingCallerIDsService) Update(ctx context.Context, sid string, params UpdateOutgoingCallerIDParams) (*OutgoingCallerID, error) {
	var out OutgoingCallerID
	if err := s.c.t.do(ctx, requestOpts{method: "POST", path: s.c.pathf("OutgoingCallerIds", sid), form: params.form()}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a verified caller ID. DELETE /OutgoingCallerIds/{Sid}.
func (s *OutgoingCallerIDsService) Delete(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{method: "DELETE", path: s.c.pathf("OutgoingCallerIds", sid)}, nil)
}
