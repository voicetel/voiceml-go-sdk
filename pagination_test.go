package voiceml_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

// pageServer serves canned bodies keyed by the exact request URI and
// records every URI it receives.
type pageServer struct {
	mu    sync.Mutex
	uris  []string
	pages map[string]any
	srv   *httptest.Server
}

func newPageServer(t *testing.T, pages map[string]any) *pageServer {
	t.Helper()
	ps := &pageServer{pages: pages}
	ps.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ps.mu.Lock()
		ps.uris = append(ps.uris, r.RequestURI)
		body, ok := ps.pages[r.RequestURI]
		ps.mu.Unlock()
		if !ok {
			w.WriteHeader(404)
			_, _ = fmt.Fprintf(w, `{"code":20404,"message":"no page for %s","status":404}`, r.RequestURI)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(ps.srv.Close)
	return ps
}

func (ps *pageServer) requested() []string {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return append([]string(nil), ps.uris...)
}

func (ps *pageServer) client(t *testing.T) *voiceml.Client {
	t.Helper()
	c, err := voiceml.NewClient(voiceml.ClientOptions{AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: ps.srv.URL, MaxRetries: voiceml.Int(0)})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

var callsPath = "/2010-04-01/Accounts/" + testAccountSid + "/Calls.json"

func TestIterateFollowsNextPageURIVerbatim(t *testing.T) {
	next1 := callsPath + "?PageSize=2&Page=1&PageToken=PA" + strings.Repeat("1", 32)
	next2 := callsPath + "?PageSize=2&Page=2&PageToken=PA" + strings.Repeat("2", 32)
	ps := newPageServer(t, map[string]any{
		callsPath + "?PageSize=2": map[string]any{
			"calls": []any{callPayload("CA" + strings.Repeat("1", 32)), callPayload("CA" + strings.Repeat("2", 32))},
			"page":  0, "page_size": 2, "start": 0, "end": 1,
			"first_page_uri": callsPath + "?PageSize=2&Page=0", "next_page_uri": next1, "previous_page_uri": nil,
			"uri": callsPath + "?PageSize=2&Page=0",
		},
		next1: map[string]any{
			"calls": []any{callPayload("CA" + strings.Repeat("3", 32)), callPayload("CA" + strings.Repeat("4", 32))},
			"page":  1, "page_size": 2, "next_page_uri": next2,
		},
		next2: map[string]any{
			"calls": []any{callPayload("CA" + strings.Repeat("5", 32))},
			"page":  2, "page_size": 2, "next_page_uri": nil,
		},
	})
	it := ps.client(t).Calls.Iterate(context.Background(), voiceml.ListCallsParams{PageSize: voiceml.Int(2)})
	var sids []string
	for it.Next() {
		sids = append(sids, it.Current().Sid)
	}
	if err := it.Err(); err != nil {
		t.Fatalf("Iterate: %v", err)
	}
	if len(sids) != 5 || sids[4] != "CA"+strings.Repeat("5", 32) {
		t.Fatalf("got %d calls %v", len(sids), sids)
	}
	want := []string{callsPath + "?PageSize=2", next1, next2}
	got := ps.requested()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("requests:\n got  %q\n want %q", got, want)
	}
	if it.PagesFetched() != 3 {
		t.Fatalf("PagesFetched = %d", it.PagesFetched())
	}
}

func TestIterateIsLazyAndStopsWhenConsumerStops(t *testing.T) {
	next := callsPath + "?Page=1&PageToken=PAx"
	ps := newPageServer(t, map[string]any{
		callsPath: map[string]any{
			"calls": []any{callPayload(""), callPayload("")}, "page": 0, "page_size": 2, "next_page_uri": next,
		},
		next: map[string]any{"calls": []any{callPayload("")}, "page": 1, "page_size": 2},
	})
	it := ps.client(t).Calls.Iterate(context.Background(), voiceml.ListCallsParams{})
	if n := len(ps.requested()); n != 0 {
		t.Fatalf("Iterate sent %d requests before Next", n)
	}
	if !it.Next() {
		t.Fatalf("Next: %v", it.Err())
	}
	if n := len(ps.requested()); n != 1 {
		t.Fatalf("after one item: %d requests, want 1", n)
	}
}

func TestIterateStopsOnEmptyPage(t *testing.T) {
	// A server that always links onward from page 0 must not loop forever.
	ps := newPageServer(t, map[string]any{
		callsPath: map[string]any{"calls": []any{}, "page": 0, "page_size": 50, "next_page_uri": callsPath + "?Page=1"},
	})
	it := ps.client(t).Calls.Iterate(context.Background(), voiceml.ListCallsParams{})
	for it.Next() {
		t.Fatal("unexpected item")
	}
	if it.Err() != nil || len(ps.requested()) != 1 {
		t.Fatalf("err=%v requests=%v", it.Err(), ps.requested())
	}
}

func TestIterateRefusesSelfLink(t *testing.T) {
	ps := newPageServer(t, map[string]any{
		callsPath: map[string]any{"calls": []any{callPayload("")}, "page": 0, "page_size": 50, "next_page_uri": callsPath},
	})
	it := ps.client(t).Calls.Iterate(context.Background(), voiceml.ListCallsParams{})
	n := 0
	for it.Next() {
		n++
	}
	var pe *voiceml.PaginationError
	if n != 1 || !errors.As(it.Err(), &pe) {
		t.Fatalf("items=%d err=%v; want 1 item then *PaginationError", n, it.Err())
	}
	if len(ps.requested()) != 1 {
		t.Fatalf("requests %v", ps.requested())
	}
}

func TestIterateRefusesForeignHost(t *testing.T) {
	var foreignHits int
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		foreignHits++
	}))
	defer foreign.Close()
	ps := newPageServer(t, map[string]any{
		callsPath: map[string]any{"calls": []any{callPayload("")}, "page": 0, "page_size": 50, "next_page_uri": foreign.URL + callsPath + "?PageToken=PAx"},
	})
	it := ps.client(t).Calls.Iterate(context.Background(), voiceml.ListCallsParams{})
	n := 0
	for it.Next() {
		n++
	}
	var pe *voiceml.PaginationError
	if n != 1 || !errors.As(it.Err(), &pe) {
		t.Fatalf("items=%d err=%v", n, it.Err())
	}
	if foreignHits != 0 {
		t.Fatal("credentials were sent to a foreign host")
	}
}

func TestNextPage2010(t *testing.T) {
	next := callsPath + "?Page=1&PageToken=PAabc"
	ps := newPageServer(t, map[string]any{
		callsPath: map[string]any{"calls": []any{callPayload("")}, "page": 0, "page_size": 1, "next_page_uri": next},
		next:      map[string]any{"calls": []any{callPayload("")}, "page": 1, "page_size": 1, "next_page_uri": nil},
	})
	c := ps.client(t)
	first, err := c.Calls.List(context.Background(), voiceml.ListCallsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if !first.HasNextPage() {
		t.Fatal("HasNextPage = false")
	}
	var second voiceml.CallList
	ok, err := c.NextPage(context.Background(), first, &second)
	if err != nil || !ok || second.Page.Page != 1 {
		t.Fatalf("NextPage: ok=%v err=%v page=%+v", ok, err, second.Page)
	}
	var third voiceml.CallList
	ok, err = c.NextPage(context.Background(), second, &third)
	if err != nil || ok {
		t.Fatalf("last page: ok=%v err=%v", ok, err)
	}
	if got := ps.requested(); len(got) != 2 || got[1] != next {
		t.Fatalf("requests %q", got)
	}
}

func TestNextPageV1Meta(t *testing.T) {
	var srvURL string
	var mu sync.Mutex
	var uris []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		uris = append(uris, r.RequestURI)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("PageToken") == "" {
			_, _ = fmt.Fprintf(w, `{"ip_records":[],"meta":{"next_page_url":%q}}`, srvURL+"/v1/IpRecords?PageSize=1&PageToken=PT1")
			return
		}
		_, _ = w.Write([]byte(`{"ip_records":[],"meta":{"next_page_url":null}}`))
	}))
	defer srv.Close()
	srvURL = srv.URL
	c, err := voiceml.NewClient(voiceml.ClientOptions{AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	first, err := c.VoiceV1.ListIpRecords(context.Background(), voiceml.V1PageParams{PageSize: voiceml.Int(1)})
	if err != nil {
		t.Fatal(err)
	}
	var second voiceml.VoiceV1IpRecordList
	ok, err := c.NextPage(context.Background(), first.Meta, &second)
	if err != nil || !ok {
		t.Fatalf("NextPage: ok=%v err=%v", ok, err)
	}
	if second.Meta.HasNextPage() {
		t.Fatal("second page should be the last")
	}
	if len(uris) != 2 || uris[1] != "/v1/IpRecords?PageSize=1&PageToken=PT1" {
		t.Fatalf("requests %q", uris)
	}

	// An absolute next_page_url on another host is refused without a request.
	foreign := "https://elsewhere.example.com/v1/IpRecords?PageToken=PT2"
	forged := voiceml.VoiceV1Meta{NextPageURL: &foreign}
	ok, err = c.NextPage(context.Background(), forged, &second)
	var pe *voiceml.PaginationError
	if ok || !errors.As(err, &pe) {
		t.Fatalf("foreign next_page_url: ok=%v err=%v", ok, err)
	}
	if len(uris) != 2 {
		t.Fatalf("a request was sent for a refused link: %q", uris)
	}
}
