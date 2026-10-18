package voiceml

import (
	"context"
	"net/url"
)

// UsageRecordsService is the /Usage/Records resource group: one record per
// usage category over a time window. List covers an explicit
// StartDate/EndDate window (all time by default); the ListXxx methods use
// the named windows.
type UsageRecordsService struct{ c *Client }

// UsageRecord is one category's usage over a window. Count, Usage and Price
// are decimal strings as the API renders them. Category is one of calls,
// calls-inbound, calls-outbound, mms, mms-inbound, mms-outbound, sms,
// sms-inbound, sms-inbound-longcode, sms-outbound, sms-outbound-longcode or
// totalprice.
type UsageRecord struct {
	AccountSid      string            `json:"account_sid"`
	APIVersion      string            `json:"api_version"`
	AsOf            string            `json:"as_of"`
	Category        string            `json:"category"`
	Count           string            `json:"count"`
	CountUnit       string            `json:"count_unit"`
	Description     string            `json:"description"`
	EndDate         string            `json:"end_date"`
	Price           string            `json:"price"`
	PriceUnit       string            `json:"price_unit"`
	StartDate       string            `json:"start_date"`
	SubresourceURIs map[string]string `json:"subresource_uris,omitempty"`
	URI             string            `json:"uri"`
	Usage           string            `json:"usage"`
	UsageUnit       string            `json:"usage_unit"`
}

// UsageRecordList is the paginated /Usage/Records response.
type UsageRecordList struct {
	Page
	UsageRecords []UsageRecord `json:"usage_records"`
}

// ListUsageRecordsParams are the query params shared by every usage window.
// StartDate / EndDate are inclusive YYYY-MM-DD dates (see Date) or offsets
// such as "-30days". IncludeSubaccounts is accepted but has no effect:
// VoiceML accounts have no subaccounts.
type ListUsageRecordsParams struct {
	Category           string
	StartDate          string
	EndDate            string
	IncludeSubaccounts *bool
	Page               *int
	PageSize           *int
	PageToken          string
}

func (p ListUsageRecordsParams) query() url.Values {
	v := url.Values{}
	setString(v, "Category", p.Category)
	setString(v, "StartDate", p.StartDate)
	setString(v, "EndDate", p.EndDate)
	setBoolP(v, "IncludeSubaccounts", p.IncludeSubaccounts)
	setIntP(v, "Page", p.Page)
	setIntP(v, "PageSize", p.PageSize)
	setString(v, "PageToken", p.PageToken)
	return v
}

func (s *UsageRecordsService) route(window string) route {
	if window == "" {
		return s.c.pathf("Usage", "Records")
	}
	return s.c.pathf("Usage", "Records", window)
}

func (s *UsageRecordsService) list(ctx context.Context, window string, params ListUsageRecordsParams) (*UsageRecordList, error) {
	var out UsageRecordList
	if err := s.c.t.do(ctx, requestOpts{method: "GET", path: s.route(window), query: params.query()}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *UsageRecordsService) iterate(ctx context.Context, window string, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return newPager(ctx, s.c, s.c.t, requestOpts{method: "GET", path: s.route(window), query: params.query()},
		func(l *UsageRecordList) ([]UsageRecord, PageLinks) { return l.UsageRecords, l.Page })
}

// List returns usage over StartDate..EndDate (all time when both are
// empty). GET /Usage/Records.json.
func (s *UsageRecordsService) List(ctx context.Context, params ListUsageRecordsParams) (*UsageRecordList, error) {
	return s.list(ctx, "", params)
}

// Iterate walks List's records lazily.
func (s *UsageRecordsService) Iterate(ctx context.Context, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return s.iterate(ctx, "", params)
}

// ListAllTime returns usage since the account was created.
// GET /Usage/Records/AllTime.json.
func (s *UsageRecordsService) ListAllTime(ctx context.Context, params ListUsageRecordsParams) (*UsageRecordList, error) {
	return s.list(ctx, "AllTime", params)
}

// IterateAllTime walks ListAllTime's records lazily.
func (s *UsageRecordsService) IterateAllTime(ctx context.Context, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return s.iterate(ctx, "AllTime", params)
}

// ListDaily returns one record per category per day.
// GET /Usage/Records/Daily.json.
func (s *UsageRecordsService) ListDaily(ctx context.Context, params ListUsageRecordsParams) (*UsageRecordList, error) {
	return s.list(ctx, "Daily", params)
}

// IterateDaily walks ListDaily's records lazily.
func (s *UsageRecordsService) IterateDaily(ctx context.Context, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return s.iterate(ctx, "Daily", params)
}

// ListLastMonth returns usage for the previous calendar month.
// GET /Usage/Records/LastMonth.json.
func (s *UsageRecordsService) ListLastMonth(ctx context.Context, params ListUsageRecordsParams) (*UsageRecordList, error) {
	return s.list(ctx, "LastMonth", params)
}

// IterateLastMonth walks ListLastMonth's records lazily.
func (s *UsageRecordsService) IterateLastMonth(ctx context.Context, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return s.iterate(ctx, "LastMonth", params)
}

// ListMonthly returns one record per category per month.
// GET /Usage/Records/Monthly.json.
func (s *UsageRecordsService) ListMonthly(ctx context.Context, params ListUsageRecordsParams) (*UsageRecordList, error) {
	return s.list(ctx, "Monthly", params)
}

// IterateMonthly walks ListMonthly's records lazily.
func (s *UsageRecordsService) IterateMonthly(ctx context.Context, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return s.iterate(ctx, "Monthly", params)
}

// ListThisMonth returns usage for the current calendar month.
// GET /Usage/Records/ThisMonth.json.
func (s *UsageRecordsService) ListThisMonth(ctx context.Context, params ListUsageRecordsParams) (*UsageRecordList, error) {
	return s.list(ctx, "ThisMonth", params)
}

// IterateThisMonth walks ListThisMonth's records lazily.
func (s *UsageRecordsService) IterateThisMonth(ctx context.Context, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return s.iterate(ctx, "ThisMonth", params)
}

// ListToday returns usage for the current day (GMT).
// GET /Usage/Records/Today.json.
func (s *UsageRecordsService) ListToday(ctx context.Context, params ListUsageRecordsParams) (*UsageRecordList, error) {
	return s.list(ctx, "Today", params)
}

// IterateToday walks ListToday's records lazily.
func (s *UsageRecordsService) IterateToday(ctx context.Context, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return s.iterate(ctx, "Today", params)
}

// ListYearly returns one record per category per year.
// GET /Usage/Records/Yearly.json.
func (s *UsageRecordsService) ListYearly(ctx context.Context, params ListUsageRecordsParams) (*UsageRecordList, error) {
	return s.list(ctx, "Yearly", params)
}

// IterateYearly walks ListYearly's records lazily.
func (s *UsageRecordsService) IterateYearly(ctx context.Context, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return s.iterate(ctx, "Yearly", params)
}

// ListYesterday returns usage for the previous day (GMT).
// GET /Usage/Records/Yesterday.json.
func (s *UsageRecordsService) ListYesterday(ctx context.Context, params ListUsageRecordsParams) (*UsageRecordList, error) {
	return s.list(ctx, "Yesterday", params)
}

// IterateYesterday walks ListYesterday's records lazily.
func (s *UsageRecordsService) IterateYesterday(ctx context.Context, params ListUsageRecordsParams) *Pager[UsageRecord] {
	return s.iterate(ctx, "Yesterday", params)
}
