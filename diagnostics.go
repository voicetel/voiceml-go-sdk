package voiceml

import (
	"context"
	"encoding/json"
)

// DiagnosticsService hits the server-root diagnostic endpoints — /health and
// /openapi.json. Neither requires authentication (the spec marks them
// security: []).
type DiagnosticsService struct{ c *Client }

// Health returns the parsed /health body. The endpoint returns 200 when all
// hard checks pass and 503 when any fail (mapped to *APIError with status 503
// — the failure list is in the JSON body if you re-parse it).
func (s *DiagnosticsService) Health(ctx context.Context) (*HealthStatus, error) {
	var out HealthStatus
	err := s.c.t.unauthRequest(ctx, "GET", "/health", &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// OpenAPI fetches the live OpenAPI spec as parsed JSON. Useful for runtime
// schema validation or codegen pipelines.
func (s *DiagnosticsService) OpenAPI(ctx context.Context) (map[string]any, error) {
	var out json.RawMessage
	if err := s.c.t.unauthRequest(ctx, "GET", "/openapi.json", &out); err != nil {
		return nil, err
	}
	var parsed map[string]any
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}
