package voiceml

import (
	"context"
	"net/url"
)

// QueuesService is the /Queues resource group plus queue members.
type QueuesService struct{ c *Client }

// Queue is a Twilio-shape Queue resource.
type Queue struct {
	Sid             string `json:"sid"`
	AccountSid      string `json:"account_sid"`
	FriendlyName    string `json:"friendly_name"`
	CurrentSize     int    `json:"current_size"`
	MaxSize         int    `json:"max_size"`
	AverageWaitTime int    `json:"average_wait_time"`
	DateCreated     string `json:"date_created"`
	DateUpdated     string `json:"date_updated"`
	URI             string `json:"uri"`
}

// QueueList is the paginated /Queues list response.
type QueueList struct {
	Page
	Queues []Queue `json:"queues"`
}

// QueueMember is one call waiting in a queue.
type QueueMember struct {
	CallSid      string `json:"call_sid"`
	QueueSid     string `json:"queue_sid"`
	AccountSid   string `json:"account_sid"`
	DateEnqueued string `json:"date_enqueued"`
	WaitTime     int    `json:"wait_time"`
	Position     int    `json:"position"`
	URI          string `json:"uri"`
}

// QueueMemberList is the paginated /Queues/{sid}/Members list response.
type QueueMemberList struct {
	Page
	QueueMembers []QueueMember `json:"queue_members"`
}

// CreateQueueParams is the body for POST /Queues. FriendlyName is required;
// the server is idempotent on FriendlyName (creating with the same name
// returns the existing queue).
type CreateQueueParams struct {
	FriendlyName string `form:"FriendlyName"`
	MaxSize      *int   `form:"MaxSize"`
}

func (p CreateQueueParams) form() url.Values {
	v := url.Values{}
	v.Set("FriendlyName", p.FriendlyName)
	setIntP(v, "MaxSize", p.MaxSize)
	return v
}

// UpdateQueueParams is the body for POST /Queues/{sid}.
type UpdateQueueParams struct {
	FriendlyName *string `form:"FriendlyName"`
	MaxSize      *int    `form:"MaxSize"`
}

func (p UpdateQueueParams) form() url.Values {
	v := url.Values{}
	setStringP(v, "FriendlyName", p.FriendlyName)
	setIntP(v, "MaxSize", p.MaxSize)
	return v
}

// DequeueParams is the body for POST /Queues/{sid}/Members/Front and
// /Queues/{sid}/Members/{call_sid}. URL is the TwiML the dequeued call will
// execute after leaving the queue.
type DequeueParams struct {
	URL    string  `form:"Url"`
	Method *string `form:"Method"`
}

func (p DequeueParams) form() url.Values {
	v := url.Values{}
	v.Set("Url", p.URL)
	setStringP(v, "Method", p.Method)
	return v
}

// Create makes a new queue. POST /Queues.
func (s *QueuesService) Create(ctx context.Context, params CreateQueueParams) (*Queue, error) {
	var out Queue
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Queues"),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all queues for this account. GET /Queues.
func (s *QueuesService) List(ctx context.Context) (*QueueList, error) {
	var out QueueList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Queues"),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a queue by SID. GET /Queues/{sid}.
func (s *QueuesService) Get(ctx context.Context, queueSid string) (*Queue, error) {
	var out Queue
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Queues", queueSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes a queue's friendly name or max size.
// POST /Queues/{sid}.
func (s *QueuesService) Update(ctx context.Context, queueSid string, params UpdateQueueParams) (*Queue, error) {
	var out Queue
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Queues", queueSid),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a queue. The server returns 409 if the queue still has
// waiting members. DELETE /Queues/{sid}.
func (s *QueuesService) Delete(ctx context.Context, queueSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE",
		path:   s.c.pathf("Queues", queueSid),
	}, nil)
}

// ListMembers returns the calls currently waiting in a queue, in position
// order. GET /Queues/{sid}/Members.
func (s *QueuesService) ListMembers(ctx context.Context, queueSid string) (*QueueMemberList, error) {
	var out QueueMemberList
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Queues", queueSid, "Members"),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PeekFront fetches the member at the front of the queue without dequeuing.
// GET /Queues/{sid}/Members/Front.
func (s *QueuesService) PeekFront(ctx context.Context, queueSid string) (*QueueMember, error) {
	var out QueueMember
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Queues", queueSid, "Members", "Front"),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DequeueFront pops the front member and runs the supplied TwiML on its call.
// POST /Queues/{sid}/Members/Front.
func (s *QueuesService) DequeueFront(ctx context.Context, queueSid string, params DequeueParams) (*QueueMember, error) {
	var out QueueMember
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Queues", queueSid, "Members", "Front"),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMember fetches a specific waiting call from the queue.
// GET /Queues/{sid}/Members/{call_sid}.
func (s *QueuesService) GetMember(ctx context.Context, queueSid, callSid string) (*QueueMember, error) {
	var out QueueMember
	err := s.c.t.do(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathf("Queues", queueSid, "Members", callSid),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DequeueMember pops a specific call from the queue and runs TwiML on it.
// POST /Queues/{sid}/Members/{call_sid}.
func (s *QueuesService) DequeueMember(ctx context.Context, queueSid, callSid string, params DequeueParams) (*QueueMember, error) {
	var out QueueMember
	err := s.c.t.do(ctx, requestOpts{
		method: "POST",
		path:   s.c.pathf("Queues", queueSid, "Members", callSid),
		form:   params.form(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
