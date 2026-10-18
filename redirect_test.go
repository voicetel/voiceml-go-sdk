package voiceml_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

type seenRequest struct {
	auth   string
	accept string
}

func TestRecordingAudioRedirectDropsAuthorization(t *testing.T) {
	var storage seenRequest
	var storageHits int32
	storageSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&storageHits, 1)
		storage = seenRequest{auth: r.Header.Get("Authorization"), accept: r.Header.Get("Accept")}
		w.Header().Set("Content-Type", "audio/wav")
		_, _ = w.Write([]byte("RIFFdata"))
	}))
	defer storageSrv.Close()

	var api seenRequest
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api = seenRequest{auth: r.Header.Get("Authorization"), accept: r.Header.Get("Accept")}
		http.Redirect(w, r, storageSrv.URL+"/bucket/RE1.wav?X-Amz-Signature=abc", http.StatusFound)
	}))
	defer apiSrv.Close()

	c, err := voiceml.NewClient(voiceml.ClientOptions{AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: apiSrv.URL})
	if err != nil {
		t.Fatal(err)
	}
	body, ct, err := c.Recordings.GetAudio(context.Background(), "RE1")
	if err != nil {
		t.Fatalf("GetAudio: %v", err)
	}
	if string(body) != "RIFFdata" || ct != "audio/wav" {
		t.Fatalf("got %q %q", body, ct)
	}
	if api.auth != basicAuthHeader() {
		t.Fatalf("API host Authorization = %q, want Basic credentials", api.auth)
	}
	if api.accept != "audio/*, */*" {
		t.Fatalf("Accept = %q", api.accept)
	}
	if storage.auth != "" {
		t.Fatalf("credentials forwarded to the storage host: %q", storage.auth)
	}
	if storageHits != 1 {
		t.Fatalf("storage hits = %d", storageHits)
	}
}

func TestSameOriginRedirectKeepsAuthorization(t *testing.T) {
	var final string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/moved.wav" {
			final = r.Header.Get("Authorization")
			_, _ = w.Write([]byte("ok"))
			return
		}
		http.Redirect(w, r, srv.URL+"/moved.wav", http.StatusTemporaryRedirect)
	}))
	defer srv.Close()
	c, err := voiceml.NewClient(voiceml.ClientOptions{AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	d, err := c.Recordings.DownloadAudio(context.Background(), "RE1")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(d.Body)
	_ = d.Body.Close()
	if string(b) != "ok" || final != basicAuthHeader() {
		t.Fatalf("body %q, Authorization on same-origin hop %q", b, final)
	}
}

func TestHTTPSToHTTPRedirectRefused(t *testing.T) {
	var plainHits int32
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&plainHits, 1)
	}))
	defer plain.Close()
	var tlsHits int32
	secure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tlsHits, 1)
		http.Redirect(w, r, plain.URL+"/RE1.wav", http.StatusFound)
	}))
	defer secure.Close()

	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: secure.URL, HTTPClient: secure.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Recordings.DownloadAudio(context.Background(), "RE1")
	if !errors.Is(err, voiceml.ErrInsecureRedirect) {
		t.Fatalf("want ErrInsecureRedirect, got %v", err)
	}
	if plainHits != 0 {
		t.Fatal("followed an https→http redirect")
	}
	if tlsHits != 1 {
		t.Fatalf("refused redirect was retried: %d attempts", tlsHits)
	}
}

func TestExportsDownloadFollowsRedirectWithoutCredentials(t *testing.T) {
	var storageAuth, storageAccept string
	storageSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		storageAuth, storageAccept = r.Header.Get("Authorization"), r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write([]byte("archive"))
	}))
	defer storageSrv.Close()
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, storageSrv.URL+"/exports/2026-10-16.tar.gz", http.StatusFound)
	}))
	defer apiSrv.Close()
	c, err := voiceml.NewClient(voiceml.ClientOptions{AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: apiSrv.URL})
	if err != nil {
		t.Fatal(err)
	}
	d, err := c.Exports.Download(context.Background(), "2026-10-16")
	if err != nil {
		t.Fatal(err)
	}
	defer d.Body.Close()
	b, _ := io.ReadAll(d.Body)
	if string(b) != "archive" || d.ContentType != "application/gzip" || storageAuth != "" || storageAccept != "*/*" {
		t.Fatalf("body=%q ct=%q auth=%q accept=%q", b, d.ContentType, storageAuth, storageAccept)
	}
}

func TestCustomCheckRedirectStillRuns(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elsewhere", http.StatusFound)
	}))
	defer srv.Close()
	sentinel := errors.New("caller policy")
	hc := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		atomic.AddInt32(&calls, 1)
		return sentinel
	}}
	c, err := voiceml.NewClient(voiceml.ClientOptions{AccountSid: testAccountSid, APIKey: testAPIKey, BaseURL: srv.URL, HTTPClient: hc, MaxRetries: voiceml.Int(0)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Calls.Get(context.Background(), "CA1")
	if !errors.Is(err, sentinel) || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
	if hc.CheckRedirect == nil {
		t.Fatal("the caller's client was modified")
	}
}
