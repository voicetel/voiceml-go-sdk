package voiceml

import (
	"context"
	"time"
)

// ExportsService lists and downloads the account's daily archives
// (/Exports).
type ExportsService struct{ c *Client }

// Export is one available daily archive.
type Export struct {
	// Date is the archive day, YYYY-MM-DD.
	Date string `json:"date"`
	// Size is the archive size in bytes.
	Size int64 `json:"size"`
	// URI is the download path for this archive.
	URI string `json:"uri"`
}

// ExportList is the GET /Exports.json response, newest first.
type ExportList struct {
	AccountSid string   `json:"account_sid"`
	URI        string   `json:"uri"`
	Exports    []Export `json:"exports"`
}

// List returns the account's available archives. GET /Exports.json.
func (s *ExportsService) List(ctx context.Context) (*ExportList, error) {
	var out ExportList
	if err := s.c.t.do(ctx, requestOpts{method: "GET", path: s.c.pathf("Exports")}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Download streams the archive for date (YYYY-MM-DD; see Date).
// GET /Exports/{Date} answers 302 to a presigned URL, which the SDK follows
// without your credentials (see RecordingsService.DownloadAudio). The caller
// must Close the returned Download's Body.
func (s *ExportsService) Download(ctx context.Context, date string) (*Download, error) {
	return s.c.t.download(ctx, requestOpts{
		method: "GET",
		path:   s.c.pathfExt("", "Exports", date),
		accept: "*/*",
	})
}

// Date formats t as the YYYY-MM-DD (UTC) date the API's date-only
// parameters and filters expect (DateCreated, StartDate, Exports dates).
func Date(t time.Time) string { return t.UTC().Format("2006-01-02") }

// Time formats t as the RFC 3339 (UTC) timestamp the API's date-time
// parameters expect (SendAt on Messages.Create).
func Time(t time.Time) string { return t.UTC().Format(time.RFC3339) }
