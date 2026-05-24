package voiceml

import (
	"context"
	"net/url"
)

// NotificationsService is the account-scoped /Notifications resource group
// (compat stubs — always empty list, fetch returns 404).
type NotificationsService struct{ c *Client }

// ListNotificationsParams are query params for GET /Notifications and the
// call-scoped GET /Calls/{sid}/Notifications compat stubs.
type ListNotificationsParams struct {
	Page          *int
	PageSize      *int
	PageToken     string
	Log           *int
	MessageDate   string
	MessageDateLt string
	MessageDateGt string
}

func (p ListNotificationsParams) query() url.Values {
	v := url.Values{}
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	setString(v, "PageToken", p.PageToken)
	setIntP(v, "Log", p.Log)
	setString(v, "MessageDate", p.MessageDate)
	setString(v, "MessageDate<", p.MessageDateLt)
	setString(v, "MessageDate>", p.MessageDateGt)
	return v
}

// List hits the account-wide compat stub at /Notifications.
func (s *NotificationsService) List(ctx context.Context, params ListNotificationsParams) (*NotificationsList, error) {
	var out NotificationsList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Notifications"),
		query:  params.query(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single account notification. Always 404 today (compat stub).
func (s *NotificationsService) Get(ctx context.Context, notificationSid string) (map[string]any, error) {
	var out map[string]any
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Notifications", notificationSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
