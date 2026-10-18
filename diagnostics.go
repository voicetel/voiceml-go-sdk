package voiceml

import (
	"context"
	"encoding/json"
)

// DiagnosticsService calls the server-root diagnostic endpoints: /health
// and the OpenAPI document (/openapi.json, /openapi.yaml, /openapi.yml).
// The spec marks them `security: []`, so they are sent without credentials,
// through the same HTTP client, timeout, proxy and User-Agent as every other
// call. Reach it as c.Diagnostics, or build one without credentials with
// NewDiagnosticsService.
type DiagnosticsService struct{ t *transport }

// HealthFailure is one tripped check from the /health deep probe.
type HealthFailure struct {
	Check  string `json:"check"`
	Detail string `json:"detail"`
}

// HealthStatus is the parsed /health body. Warnings and Failures are
// omitted by the server when empty.
type HealthStatus struct {
	OK       bool            `json:"ok"`
	Warnings []HealthFailure `json:"warnings,omitempty"`
	Failures []HealthFailure `json:"failures,omitempty"`
}

// Health returns the parsed /health body. The endpoint answers 200 when all
// hard checks pass and 503 when any fail; a 503 is returned as an *APIError
// wrapping ErrServer whose Body holds the failure list (decode it into a
// HealthStatus). Health makes a single attempt: a 503 is the answer, not a
// transient failure.
func (s *DiagnosticsService) Health(ctx context.Context) (*HealthStatus, error) {
	var out HealthStatus
	err := s.t.do(ctx, requestOpts{method: "GET", path: apiRoute("/health"), noAuth: true, noRetry: true}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// OpenAPI fetches the live OpenAPI document (/openapi.json) as parsed JSON.
// Useful for runtime schema validation or codegen pipelines.
func (s *DiagnosticsService) OpenAPI(ctx context.Context) (map[string]any, error) {
	var raw json.RawMessage
	if err := s.t.do(ctx, requestOpts{method: "GET", path: apiRoute("/openapi.json"), noAuth: true}, &raw); err != nil {
		return nil, err
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, &DecodeError{StatusCode: 200, Body: truncateBytes(raw, maxDecodeErrorBody), Err: err}
	}
	return parsed, nil
}

// OpenAPIYAML fetches the live OpenAPI document as YAML (/openapi.yaml).
func (s *DiagnosticsService) OpenAPIYAML(ctx context.Context) ([]byte, error) {
	return s.yaml(ctx, "/openapi.yaml")
}

// OpenAPIYML fetches the live OpenAPI document as YAML from the /openapi.yml
// alias.
func (s *DiagnosticsService) OpenAPIYML(ctx context.Context) ([]byte, error) {
	return s.yaml(ctx, "/openapi.yml")
}

func (s *DiagnosticsService) yaml(ctx context.Context, path string) ([]byte, error) {
	d, err := s.t.download(ctx, requestOpts{
		method: "GET", path: apiRoute(path), noAuth: true,
		accept: "application/yaml, text/yaml;q=0.9, */*;q=0.1",
	})
	if err != nil {
		return nil, err
	}
	defer d.Body.Close()
	b, err := readAtMost(d.Body, maxJSONBody)
	if err != nil {
		return nil, transportError(err, 1)
	}
	return b, nil
}
