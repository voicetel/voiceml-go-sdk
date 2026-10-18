//go:build go1.23

package voiceml_test

import (
	"context"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

func TestPagerAllRangeOverFunc(t *testing.T) {
	next := callsPath + "?Page=1&PageToken=PAall"
	ps := newPageServer(t, map[string]any{
		callsPath: map[string]any{"calls": []any{callPayload("CA" + strings.Repeat("1", 32)), callPayload("CA" + strings.Repeat("2", 32))}, "page": 0, "page_size": 2, "next_page_uri": next},
		next:      map[string]any{"calls": []any{callPayload("CA" + strings.Repeat("3", 32))}, "page": 1, "page_size": 2},
	})
	c := ps.client(t)

	var sids []string
	for call, err := range c.Calls.Iterate(context.Background(), voiceml.ListCallsParams{}).All() {
		if err != nil {
			t.Fatal(err)
		}
		sids = append(sids, call.Sid)
	}
	if len(sids) != 3 {
		t.Fatalf("got %v", sids)
	}

	// Breaking out early stops further requests.
	before := len(ps.requested())
	for range c.Calls.Iterate(context.Background(), voiceml.ListCallsParams{}).All() {
		break
	}
	if got := len(ps.requested()) - before; got != 1 {
		t.Fatalf("early break sent %d requests, want 1", got)
	}
}

func TestPagerAllYieldsError(t *testing.T) {
	ps := newPageServer(t, map[string]any{})
	var gotErr error
	for _, err := range ps.client(t).Calls.Iterate(context.Background(), voiceml.ListCallsParams{}).All() {
		gotErr = err
	}
	if !voiceml.IsNotFound(gotErr) {
		t.Fatalf("want the 404 yielded, got %v", gotErr)
	}
}
