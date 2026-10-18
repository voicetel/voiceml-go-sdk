package voiceml

import (
	"context"
	"net/url"
)

// NotificationsService is the account-scoped /Notifications resource
// group: alert-log entries recorded when a TwiML fetch or parse fails.
type NotificationsService struct{ c *Client }

// ListNotificationsParams are query params for GET /Notifications and the
// call-scoped GET /Calls/{sid}/Notifications. Log is 0 (error) or 1
// (warning); MessageDate is a YYYY-MM-DD date (see Date).
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

// List returns a page of the account's notifications. GET /Notifications.
func (s *NotificationsService) List(ctx context.Context, params ListNotificationsParams) (*NotificationList, error) {
	var out NotificationList
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

// Iterate walks every notification on the account lazily.
func (s *NotificationsService) Iterate(ctx context.Context, params ListNotificationsParams) *Pager[Notification] {
	return newPager(ctx, s.c, s.c.t, requestOpts{method: "GET", path: s.c.pathf("Notifications"), query: params.query()},
		func(l *NotificationList) ([]Notification, PageLinks) { return l.Notifications, l.Page })
}

// Get fetches a single account notification. GET /Notifications/{Sid}.
func (s *NotificationsService) Get(ctx context.Context, notificationSid string) (*Notification, error) {
	var out Notification
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Notifications", notificationSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
