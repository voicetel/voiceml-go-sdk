package voiceml_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

// countingTransport proves the caller's HTTP client carries diagnostics.
type countingTransport struct {
	n    int32
	next http.RoundTripper
}

func (c *countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	atomic.AddInt32(&c.n, 1)
	return c.next.RoundTrip(r)
}

func TestDiagnosticsWithoutCredentials(t *testing.T) {
	var gotAuth, gotUA, gotAccept []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		gotUA = append(gotUA, r.Header.Get("User-Agent"))
		gotAccept = append(gotAccept, r.Header.Get("Accept"))
		switch r.URL.Path {
		case "/health":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "/openapi.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"openapi":"3.1.0","info":{"version":"0.9.2"}}`))
		case "/openapi.yaml", "/openapi.yml":
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write([]byte("openapi: 3.1.0\n"))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	rt := &countingTransport{next: http.DefaultTransport}
	diag, err := voiceml.NewDiagnosticsService(voiceml.ClientOptions{
		BaseURL: srv.URL, UserAgent: "probe/1.0", HTTPClient: &http.Client{Transport: rt},
	})
	if err != nil {
		t.Fatalf("NewDiagnosticsService without credentials: %v", err)
	}
	ctx := context.Background()
	h, err := diag.Health(ctx)
	if err != nil || !h.OK || h.Warnings != nil || h.Failures != nil {
		t.Fatalf("Health: %+v %v", h, err)
	}
	spec, err := diag.OpenAPI(ctx)
	if err != nil || spec["openapi"] != "3.1.0" {
		t.Fatalf("OpenAPI: %v %v", spec, err)
	}
	for _, f := range []func(context.Context) ([]byte, error){diag.OpenAPIYAML, diag.OpenAPIYML} {
		y, err := f(ctx)
		if err != nil || string(y) != "openapi: 3.1.0\n" {
			t.Fatalf("yaml: %q %v", y, err)
		}
	}
	if rt.n != 4 {
		t.Fatalf("custom HTTPClient carried %d of 4 requests", rt.n)
	}
	for i := range gotAuth {
		if gotAuth[i] != "" {
			t.Fatalf("request %d sent Authorization", i)
		}
		if gotUA[i] != "probe/1.0" {
			t.Fatalf("request %d User-Agent %q", i, gotUA[i])
		}
	}
	if gotAccept[2] != "application/yaml, text/yaml;q=0.9, */*;q=0.1" {
		t.Fatalf("yaml Accept = %q", gotAccept[2])
	}
}

func TestHealth503IsServerErrorWithFailures(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(503)
		_, _ = w.Write([]byte(`{"ok":false,"failures":[{"check":"db","detail":"down"}]}`))
	}))
	defer srv.Close()
	diag, err := voiceml.NewDiagnosticsService(voiceml.ClientOptions{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	_, err = diag.Health(context.Background())
	var apiErr *voiceml.APIError
	if !voiceml.IsServer(err) || !errors.As(err, &apiErr) || apiErr.StatusCode != 503 {
		t.Fatalf("want 503 server error, got %v", err)
	}
	var hs voiceml.HealthStatus
	if err := json.Unmarshal(apiErr.Body, &hs); err != nil || len(hs.Failures) != 1 || hs.Failures[0].Check != "db" {
		t.Fatalf("failures not recoverable from Body: %+v %v", hs, err)
	}
	if hits != 1 {
		t.Fatalf("Health made %d attempts, want 1", hits)
	}
}

func TestDiagnosticsServiceValidatesOptions(t *testing.T) {
	_, err := voiceml.NewDiagnosticsService(voiceml.ClientOptions{BaseURL: "ftp://example.com"})
	var cfg *voiceml.ConfigurationError
	if !errors.As(err, &cfg) {
		t.Fatalf("want *ConfigurationError, got %v", err)
	}
}
