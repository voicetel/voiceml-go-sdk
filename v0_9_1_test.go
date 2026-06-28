// v0.9.1 surface tests — Assistants v1 (the /v1/Assistants product surface:
// Assistants/Tools/Knowledge/Sessions/Messages/Feedback/Policy).
//
// Each case stands up an httptest server, asserts the request the SDK sent
// (method + path + JSON-body shape + query encoding), and feeds back a
// canned wire shape decoded into the response model. Decode coverage for
// the inline-children fetch shapes + Feedback float round-trip + open
// json.RawMessage fields lives in TestAssistantsV1Decodes below.

package voiceml_test

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

// Assistants v1 sids — disambiguated suffixes so collisions with other test
// files' identifiers stay obvious (existing IL... / SD... etc. constants
// live in v0_9_0_test.go).
const (
	asstID  = "aia_asst_abc123"
	toolID  = "aia_tool_xyz789"
	knowID  = "aia_know_qwe456"
	sessID  = "sess_dlm0001"
	msgID   = "aia_msg_zzz111"
	fdbkID  = "aia_fdbk_yyy222"
	plcyID  = "aia_plcy_xxx333"
	userSid = "US" + "0123456789abcdef0123456789abcdef"
)

// ---------------------------------------------------------------------------
// Client wiring — c.AssistantsV1 is non-nil after NewClient.
// ---------------------------------------------------------------------------

func TestV091ClientWiring(t *testing.T) {
	c, _, done := newClient(t, nil, nil)
	defer done()
	if c.AssistantsV1 == nil {
		t.Fatal("AssistantsV1 not wired up")
	}
}

// ---------------------------------------------------------------------------
// Path / method / body shape — table-driven across the 30-operation surface.
// ---------------------------------------------------------------------------

func TestAssistantsV1Paths(t *testing.T) {
	// Shared minimal payloads. Each case picks one and the test asserts the
	// SDK round-trips it. Open object fields (`meta`, `policy_details`,
	// `knowledge_source_details`, `content`) are JSON-encoded literals so
	// the test only depends on `json.RawMessage` preserving raw bytes.
	asstPayload := map[string]any{
		"account_sid": testAccountSid, "id": asstID,
		"name": "support-bot", "owner": "platform", "model": "gpt-4o-mini",
		"personality_prompt": "Be helpful.",
		"customer_ai":        map[string]any{"perception_engine_enabled": true},
		"url":                "/v1/Assistants/" + asstID,
		"date_created":       "x", "date_updated": "x",
	}
	asstWithChildrenPayload := map[string]any{
		"account_sid": testAccountSid, "id": asstID, "name": "support-bot",
		"owner": "platform", "model": "gpt-4o-mini", "personality_prompt": "Be helpful.",
		"customer_ai":  map[string]any{},
		"url":          "/v1/Assistants/" + asstID,
		"date_created": "x", "date_updated": "x",
		"tools": []any{
			map[string]any{
				"id": toolID, "name": "lookup", "type": "function",
				"enabled": true, "requires_auth": false, "description": "do a thing",
				"meta": map[string]any{"x": 1}, "date_created": "x", "date_updated": "x",
			},
		},
		"knowledge": []any{
			map[string]any{
				"id": knowID, "name": "kb", "type": "documents",
				"date_created": "x", "date_updated": "x",
			},
		},
	}
	toolPayload := map[string]any{
		"id": toolID, "name": "lookup", "type": "function",
		"enabled": true, "requires_auth": false, "description": "do a thing",
		"meta": map[string]any{"k": "v"}, "date_created": "x", "date_updated": "x",
	}
	toolWithPoliciesPayload := map[string]any{
		"id": toolID, "name": "lookup", "type": "function",
		"enabled": true, "requires_auth": false, "description": "do a thing",
		"meta": map[string]any{}, "date_created": "x", "date_updated": "x",
		"policies": []any{
			map[string]any{
				"id": plcyID, "type": "function", "policy_details": map[string]any{"allow": true},
			},
		},
	}
	knowPayload := map[string]any{
		"id": knowID, "name": "kb", "type": "documents",
		"date_created": "x", "date_updated": "x",
	}
	statusPayload := map[string]any{
		"status": "ready", "last_status": "indexing",
	}
	chunkListPayload := map[string]any{
		"chunks": []any{
			map[string]any{"content": "chunk-1", "metadata": map[string]any{"page": 1}},
		},
		"meta": map[string]any{},
	}
	sessPayload := map[string]any{
		"id": sessID, "account_sid": testAccountSid, "assistant_id": asstID,
		"verified": true, "identity": "user-1",
		"date_created": "x", "date_updated": "x",
	}
	msgListPayload := map[string]any{
		"messages": []any{
			map[string]any{
				"id": msgID, "session_id": sessID, "role": "assistant",
				"content": map[string]any{"text": "hi"},
			},
		},
		"meta": map[string]any{},
	}
	sendPayload := map[string]any{
		"status": "completed", "session_id": sessID,
		"account_sid": testAccountSid, "body": "Hi! How can I help?",
		"flagged": false, "aborted": false,
	}
	fdbkPayload := map[string]any{
		"id": fdbkID, "assistant_id": asstID, "session_id": sessID,
		"message_id": msgID, "score": 0.75, "text": "good",
		"date_created": "x", "date_updated": "x",
	}

	cases := []struct {
		name        string
		response    map[string]any
		method      string
		path        string
		invoke      func(c *voiceml.Client) error
		assertBody  func(t *testing.T, body []byte)
		assertQuery func(t *testing.T, q url.Values)
	}{
		// --- Assistants ---------------------------------------------------
		{
			name: "CreateAssistant", response: asstPayload,
			method: "POST", path: "/v1/Assistants",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.CreateAssistant(context.Background(), &voiceml.CreateAssistantRequest{
					Name:              "support-bot",
					Owner:             voiceml.String("platform"),
					Model:             voiceml.String("gpt-4o-mini"),
					PersonalityPrompt: voiceml.String("Be helpful."),
					CustomerAI: &voiceml.AssistantsV1CustomerAI{
						PerceptionEngineEnabled: voiceml.Bool(true),
					},
					SegmentCredential: json.RawMessage(`{"workspace":"acme"}`),
				})
				return err
			},
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode body: %v (body=%s)", err, body)
				}
				if got["name"] != "support-bot" || got["owner"] != "platform" ||
					got["model"] != "gpt-4o-mini" || got["personality_prompt"] != "Be helpful." {
					t.Fatalf("body: %+v", got)
				}
				ai, ok := got["customer_ai"].(map[string]any)
				if !ok || ai["perception_engine_enabled"] != true {
					t.Fatalf("customer_ai missing/wrong: %+v", got["customer_ai"])
				}
				seg, ok := got["segment_credential"].(map[string]any)
				if !ok || seg["workspace"] != "acme" {
					t.Fatalf("segment_credential missing/wrong: %+v", got["segment_credential"])
				}
			},
		},
		{
			name:     "ListAssistants",
			response: map[string]any{"assistants": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/Assistants",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListAssistants(context.Background(), voiceml.ListAssistantsParams{
					PageSize:  voiceml.Int(25),
					Page:      voiceml.Int(0),
					PageToken: voiceml.String("cursor-1"),
				})
				return err
			},
			assertQuery: func(t *testing.T, q url.Values) {
				if q.Get("PageSize") != "25" || q.Get("Page") != "0" || q.Get("PageToken") != "cursor-1" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name: "FetchAssistant", response: asstWithChildrenPayload,
			method: "GET", path: "/v1/Assistants/" + asstID,
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.FetchAssistant(context.Background(), asstID)
				return err
			},
		},
		{
			name: "UpdateAssistant", response: asstPayload,
			method: "PUT", path: "/v1/Assistants/" + asstID,
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.UpdateAssistant(context.Background(), asstID, &voiceml.UpdateAssistantRequest{
					PersonalityPrompt: voiceml.String("Be VERY helpful."),
				})
				return err
			},
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got["personality_prompt"] != "Be VERY helpful." {
					t.Fatalf("body: %+v", got)
				}
				// Unset fields must not appear (omitempty).
				if _, ok := got["name"]; ok {
					t.Fatalf("name unexpectedly serialized: %+v", got)
				}
			},
		},
		{
			name:   "DeleteAssistant",
			method: "DELETE", path: "/v1/Assistants/" + asstID,
			invoke: func(c *voiceml.Client) error {
				return c.AssistantsV1.DeleteAssistant(context.Background(), asstID)
			},
		},

		// --- Tools --------------------------------------------------------
		{
			name: "CreateTool", response: toolPayload,
			method: "POST", path: "/v1/Tools",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.CreateTool(context.Background(), &voiceml.CreateToolRequest{
					Name:    "lookup",
					Type:    "function",
					Enabled: true,
					Meta:    json.RawMessage(`{"k":"v"}`),
				})
				return err
			},
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got["name"] != "lookup" || got["type"] != "function" ||
					got["enabled"] != true {
					t.Fatalf("body: %+v", got)
				}
				m, ok := got["meta"].(map[string]any)
				if !ok || m["k"] != "v" {
					t.Fatalf("meta: %+v", got["meta"])
				}
			},
		},
		{
			name:     "ListTools",
			response: map[string]any{"tools": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/Tools",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListTools(context.Background(), voiceml.ListToolsParams{
					AssistantID: voiceml.String(asstID),
					PageSize:    voiceml.Int(10),
				})
				return err
			},
			assertQuery: func(t *testing.T, q url.Values) {
				if q.Get("AssistantId") != asstID || q.Get("PageSize") != "10" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name: "FetchTool", response: toolWithPoliciesPayload,
			method: "GET", path: "/v1/Tools/" + toolID,
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.FetchTool(context.Background(), toolID)
				return err
			},
		},
		{
			name: "UpdateTool", response: toolPayload,
			method: "PUT", path: "/v1/Tools/" + toolID,
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.UpdateTool(context.Background(), toolID, &voiceml.UpdateToolRequest{
					Enabled: voiceml.Bool(false),
				})
				return err
			},
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got["enabled"] != false {
					t.Fatalf("body: %+v", got)
				}
				if len(got) != 1 {
					t.Fatalf("partial body should have only Enabled; got %+v", got)
				}
			},
		},
		{
			name:   "DeleteTool",
			method: "DELETE", path: "/v1/Tools/" + toolID,
			invoke: func(c *voiceml.Client) error {
				return c.AssistantsV1.DeleteTool(context.Background(), toolID)
			},
		},
		{
			name:     "ListAssistantTools",
			response: map[string]any{"tools": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/Assistants/" + asstID + "/Tools",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListAssistantTools(context.Background(), asstID, voiceml.V1PageParams{
					PageSize: voiceml.Int(50),
				})
				return err
			},
			assertQuery: func(t *testing.T, q url.Values) {
				if q.Get("PageSize") != "50" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name:   "AttachToolToAssistant",
			method: "POST", path: "/v1/Assistants/" + asstID + "/Tools/" + toolID,
			invoke: func(c *voiceml.Client) error {
				return c.AssistantsV1.AttachToolToAssistant(context.Background(), asstID, toolID)
			},
		},
		{
			name:   "DetachToolFromAssistant",
			method: "DELETE", path: "/v1/Assistants/" + asstID + "/Tools/" + toolID,
			invoke: func(c *voiceml.Client) error {
				return c.AssistantsV1.DetachToolFromAssistant(context.Background(), asstID, toolID)
			},
		},

		// --- Knowledge ----------------------------------------------------
		{
			name: "CreateKnowledge", response: knowPayload,
			method: "POST", path: "/v1/Knowledge",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.CreateKnowledge(context.Background(), &voiceml.CreateKnowledgeRequest{
					Name:                   "kb",
					Type:                   "documents",
					KnowledgeSourceDetails: json.RawMessage(`{"url":"s3://b/k"}`),
				})
				return err
			},
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got["name"] != "kb" || got["type"] != "documents" {
					t.Fatalf("body: %+v", got)
				}
				src, ok := got["knowledge_source_details"].(map[string]any)
				if !ok || src["url"] != "s3://b/k" {
					t.Fatalf("knowledge_source_details: %+v", got["knowledge_source_details"])
				}
			},
		},
		{
			name:     "ListKnowledge",
			response: map[string]any{"knowledge": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/Knowledge",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListKnowledge(context.Background(), voiceml.ListKnowledgeParams{
					AssistantID: voiceml.String(asstID),
				})
				return err
			},
			assertQuery: func(t *testing.T, q url.Values) {
				if q.Get("AssistantId") != asstID {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name: "FetchKnowledge", response: knowPayload,
			method: "GET", path: "/v1/Knowledge/" + knowID,
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.FetchKnowledge(context.Background(), knowID)
				return err
			},
		},
		{
			name: "UpdateKnowledge", response: knowPayload,
			method: "PUT", path: "/v1/Knowledge/" + knowID,
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.UpdateKnowledge(context.Background(), knowID, &voiceml.UpdateKnowledgeRequest{
					EmbeddingModel: voiceml.String("text-embedding-3-small"),
				})
				return err
			},
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got["embedding_model"] != "text-embedding-3-small" || len(got) != 1 {
					t.Fatalf("partial body: %+v", got)
				}
			},
		},
		{
			name:   "DeleteKnowledge",
			method: "DELETE", path: "/v1/Knowledge/" + knowID,
			invoke: func(c *voiceml.Client) error {
				return c.AssistantsV1.DeleteKnowledge(context.Background(), knowID)
			},
		},
		{
			name: "FetchKnowledgeStatus", response: statusPayload,
			method: "GET", path: "/v1/Knowledge/" + knowID + "/Status",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.FetchKnowledgeStatus(context.Background(), knowID)
				return err
			},
		},
		{
			name: "ListKnowledgeChunks", response: chunkListPayload,
			method: "GET", path: "/v1/Knowledge/" + knowID + "/Chunks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListKnowledgeChunks(context.Background(), knowID, voiceml.V1PageParams{
					PageSize: voiceml.Int(20),
				})
				return err
			},
			assertQuery: func(t *testing.T, q url.Values) {
				if q.Get("PageSize") != "20" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name:     "ListAssistantKnowledge",
			response: map[string]any{"knowledge": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/Assistants/" + asstID + "/Knowledge",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListAssistantKnowledge(context.Background(), asstID, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name:   "AttachKnowledgeToAssistant",
			method: "POST", path: "/v1/Assistants/" + asstID + "/Knowledge/" + knowID,
			invoke: func(c *voiceml.Client) error {
				return c.AssistantsV1.AttachKnowledgeToAssistant(context.Background(), asstID, knowID)
			},
		},
		{
			name:   "DetachKnowledgeFromAssistant",
			method: "DELETE", path: "/v1/Assistants/" + asstID + "/Knowledge/" + knowID,
			invoke: func(c *voiceml.Client) error {
				return c.AssistantsV1.DetachKnowledgeFromAssistant(context.Background(), asstID, knowID)
			},
		},

		// --- Sessions + Messages -----------------------------------------
		{
			name: "SendMessage", response: sendPayload,
			method: "POST", path: "/v1/Assistants/" + asstID + "/Messages",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.SendMessage(context.Background(), asstID, &voiceml.SendMessageRequest{
					Identity:  "user-1",
					Body:      "What's the status of order 42?",
					SessionID: voiceml.String(sessID),
					Mode:      voiceml.String("sync"),
				})
				return err
			},
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got["identity"] != "user-1" ||
					got["body"] != "What's the status of order 42?" ||
					got["session_id"] != sessID || got["mode"] != "sync" {
					t.Fatalf("body: %+v", got)
				}
			},
		},
		{
			name:     "ListSessions",
			response: map[string]any{"sessions": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/Sessions",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListSessions(context.Background(), voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchSession", response: sessPayload,
			method: "GET", path: "/v1/Sessions/" + sessID,
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.FetchSession(context.Background(), sessID)
				return err
			},
		},
		{
			name: "ListSessionMessages", response: msgListPayload,
			method: "GET", path: "/v1/Sessions/" + sessID + "/Messages",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListSessionMessages(context.Background(), sessID, voiceml.V1PageParams{})
				return err
			},
		},

		// --- Feedback -----------------------------------------------------
		{
			name:     "ListAssistantFeedback",
			response: map[string]any{"feedbacks": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/Assistants/" + asstID + "/Feedbacks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListAssistantFeedback(context.Background(), asstID, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "CreateFeedback", response: fdbkPayload,
			method: "POST", path: "/v1/Assistants/" + asstID + "/Feedbacks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.CreateFeedback(context.Background(), asstID, &voiceml.CreateFeedbackRequest{
					SessionID: sessID,
					MessageID: voiceml.String(msgID),
					Score:     voiceml.Float64(0.75),
					Text:      voiceml.String("good"),
				})
				return err
			},
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got["session_id"] != sessID || got["message_id"] != msgID ||
					got["text"] != "good" {
					t.Fatalf("body: %+v", got)
				}
				// JSON numbers decode as float64 — compare numerically.
				s, ok := got["score"].(float64)
				if !ok || s != 0.75 {
					t.Fatalf("score: %v (%T)", got["score"], got["score"])
				}
			},
		},

		// --- Policies -----------------------------------------------------
		{
			name:     "ListPolicies",
			response: map[string]any{"policies": []any{}, "meta": map[string]any{}},
			method:   "GET", path: "/v1/Policies",
			invoke: func(c *voiceml.Client) error {
				_, err := c.AssistantsV1.ListPolicies(context.Background(), voiceml.ListPoliciesParams{
					ToolID:      voiceml.String(toolID),
					KnowledgeID: voiceml.String(knowID),
					PageSize:    voiceml.Int(5),
				})
				return err
			},
			assertQuery: func(t *testing.T, q url.Values) {
				if q.Get("ToolId") != toolID || q.Get("KnowledgeId") != knowID || q.Get("PageSize") != "5" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []handlerStep{jsonStep(200, tc.response)}
			c, rec, done := newClient(t, steps, nil)
			defer done()
			if err := tc.invoke(c); err != nil {
				t.Fatalf("invoke: %v", err)
			}
			r := rec.requests[0]
			if r.Method != tc.method {
				t.Fatalf("method: %s (want %s)", r.Method, tc.method)
			}
			if r.Path != tc.path {
				t.Fatalf("path: %s (want %s)", r.Path, tc.path)
			}
			// /v1/ surface must NOT leak the account sid into the path; account
			// is resolved from Basic auth instead.
			if strings.Contains(r.Path, testAccountSid) {
				t.Fatalf("path leaked account sid: %q", r.Path)
			}
			// JSON bodies — assert Content-Type when a body was sent.
			if len(r.Body) > 0 {
				if got := r.Header.Get("Content-Type"); got != "application/json" {
					t.Fatalf("Content-Type: want application/json, got %q", got)
				}
			}
			if tc.assertBody != nil {
				tc.assertBody(t, r.Body)
			}
			if tc.assertQuery != nil {
				q, _ := url.ParseQuery(r.Query)
				tc.assertQuery(t, q)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Decode coverage — list envelopes, inline-children fetch shapes, Feedback's
// float Score, and json.RawMessage round-trip for the open object fields.
// ---------------------------------------------------------------------------

func TestAssistantsV1Decodes(t *testing.T) {
	// 1. Assistant list envelope round-trips one row + meta.
	t.Run("AssistantListEnvelope", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"assistants": []map[string]any{
				{
					"account_sid":        testAccountSid,
					"id":                 asstID,
					"name":               "support-bot",
					"owner":              "platform",
					"model":              "gpt-4o-mini",
					"personality_prompt": "Be helpful.",
					"customer_ai": map[string]any{
						"perception_engine_enabled":      true,
						"personalization_engine_enabled": false,
					},
					"url":          "/v1/Assistants/" + asstID,
					"date_created": "2026-06-28T12:00:00Z",
					"date_updated": "2026-06-28T12:00:00Z",
				},
			},
			"meta": map[string]any{
				"page": 0, "page_size": 50,
				"url": "/v1/Assistants?PageSize=50",
				"key": "assistants",
			},
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		out, err := c.AssistantsV1.ListAssistants(context.Background(), voiceml.ListAssistantsParams{})
		if err != nil {
			t.Fatalf("ListAssistants: %v", err)
		}
		if len(out.Assistants) != 1 {
			t.Fatalf("assistants len: %d", len(out.Assistants))
		}
		a := out.Assistants[0]
		if a.ID == nil || *a.ID != asstID {
			t.Fatalf("assistant id: %+v", a.ID)
		}
		if a.CustomerAI == nil ||
			a.CustomerAI.PerceptionEngineEnabled == nil ||
			!*a.CustomerAI.PerceptionEngineEnabled ||
			a.CustomerAI.PersonalizationEngineEnabled == nil ||
			*a.CustomerAI.PersonalizationEngineEnabled {
			t.Fatalf("customer_ai didn't round-trip: %+v", a.CustomerAI)
		}
		if out.Meta.Key == nil || *out.Meta.Key != "assistants" {
			t.Fatalf("meta.key: %+v", out.Meta.Key)
		}
	})

	// 2. FetchAssistant inlines tools + knowledge as typed slices.
	t.Run("FetchAssistantInlineChildren", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"account_sid": testAccountSid, "id": asstID,
			"name": "x", "owner": "x", "model": "x", "personality_prompt": "x",
			"customer_ai":  map[string]any{},
			"date_created": "x", "date_updated": "x",
			"tools": []any{
				map[string]any{
					"id": toolID, "name": "lookup", "type": "function",
					"enabled": true, "requires_auth": false, "description": "d",
					"meta":         map[string]any{"role": "admin"},
					"date_created": "x", "date_updated": "x",
				},
			},
			"knowledge": []any{
				map[string]any{
					"id": knowID, "name": "kb", "type": "documents",
					"date_created": "x", "date_updated": "x",
				},
			},
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		out, err := c.AssistantsV1.FetchAssistant(context.Background(), asstID)
		if err != nil {
			t.Fatalf("FetchAssistant: %v", err)
		}
		if len(out.Tools) != 1 || out.Tools[0].ID == nil || *out.Tools[0].ID != toolID {
			t.Fatalf("tools: %+v", out.Tools)
		}
		if len(out.Knowledge) != 1 || out.Knowledge[0].ID == nil || *out.Knowledge[0].ID != knowID {
			t.Fatalf("knowledge: %+v", out.Knowledge)
		}
		// `meta` round-trips as raw JSON bytes.
		var meta map[string]string
		if err := json.Unmarshal(out.Tools[0].Meta, &meta); err != nil {
			t.Fatalf("decode tool meta: %v (raw=%s)", err, out.Tools[0].Meta)
		}
		if meta["role"] != "admin" {
			t.Fatalf("tool meta: %+v", meta)
		}
	})

	// 3. FetchTool inlines policies; policy_details survives as RawMessage.
	t.Run("FetchToolWithPolicies", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"id": toolID, "name": "lookup", "type": "function",
			"enabled": true, "requires_auth": false, "description": "d",
			"meta": map[string]any{}, "date_created": "x", "date_updated": "x",
			"policies": []any{
				map[string]any{
					"id": plcyID, "type": "function", "user_sid": userSid,
					"policy_details": map[string]any{"allow": []string{"read"}},
				},
			},
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		out, err := c.AssistantsV1.FetchTool(context.Background(), toolID)
		if err != nil {
			t.Fatalf("FetchTool: %v", err)
		}
		if len(out.Policies) != 1 {
			t.Fatalf("policies: %+v", out.Policies)
		}
		p := out.Policies[0]
		if p.UserSid == nil || *p.UserSid != userSid {
			t.Fatalf("user_sid: %+v", p.UserSid)
		}
		var details map[string][]string
		if err := json.Unmarshal(p.PolicyDetails, &details); err != nil {
			t.Fatalf("decode policy_details: %v", err)
		}
		if len(details["allow"]) != 1 || details["allow"][0] != "read" {
			t.Fatalf("policy_details: %+v", details)
		}
	})

	// 4. Knowledge status payload includes optional last_status.
	t.Run("KnowledgeStatus", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"status": "ready", "last_status": "indexing",
			"date_updated": "2026-06-28T12:00:00Z",
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		s, err := c.AssistantsV1.FetchKnowledgeStatus(context.Background(), knowID)
		if err != nil {
			t.Fatalf("FetchKnowledgeStatus: %v", err)
		}
		if s.Status == nil || *s.Status != "ready" {
			t.Fatalf("status: %+v", s.Status)
		}
		if s.LastStatus == nil || *s.LastStatus != "indexing" {
			t.Fatalf("last_status: %+v", s.LastStatus)
		}
	})

	// 5. SendMessage moderation signals + body echo round-trip.
	t.Run("SendMessageResponse", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"status": "completed", "session_id": sessID,
			"account_sid": testAccountSid, "body": "Hello!",
			"flagged": false, "aborted": false,
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		resp, err := c.AssistantsV1.SendMessage(context.Background(), asstID, &voiceml.SendMessageRequest{
			Identity: "user-1", Body: "hi",
		})
		if err != nil {
			t.Fatalf("SendMessage: %v", err)
		}
		if resp.Status == nil || *resp.Status != "completed" {
			t.Fatalf("status: %+v", resp.Status)
		}
		if resp.SessionID == nil || *resp.SessionID != sessID {
			t.Fatalf("session_id: %+v", resp.SessionID)
		}
		if resp.Body == nil || *resp.Body != "Hello!" {
			t.Fatalf("body: %+v", resp.Body)
		}
		if resp.Flagged == nil || *resp.Flagged {
			t.Fatalf("flagged: %+v", resp.Flagged)
		}
	})

	// 6. Feedback float Score round-trips through *float64.
	t.Run("FeedbackFloatScore", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"id": fdbkID, "assistant_id": asstID, "session_id": sessID,
			"message_id": msgID, "score": 0.42, "text": "ok",
			"date_created": "x", "date_updated": "x",
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		f, err := c.AssistantsV1.CreateFeedback(context.Background(), asstID, &voiceml.CreateFeedbackRequest{
			SessionID: sessID,
			Score:     voiceml.Float64(0.42),
		})
		if err != nil {
			t.Fatalf("CreateFeedback: %v", err)
		}
		if f.Score == nil || *f.Score != 0.42 {
			t.Fatalf("score: %+v", f.Score)
		}
	})

	// 7. Session-message list decodes Content as raw JSON bytes.
	t.Run("SessionMessageContent", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"messages": []any{
				map[string]any{
					"id": msgID, "session_id": sessID, "role": "assistant",
					"content": map[string]any{"text": "hello", "n": 1},
				},
			},
			"meta": map[string]any{},
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		out, err := c.AssistantsV1.ListSessionMessages(context.Background(), sessID, voiceml.V1PageParams{})
		if err != nil {
			t.Fatalf("ListSessionMessages: %v", err)
		}
		if len(out.Messages) != 1 {
			t.Fatalf("messages: %+v", out.Messages)
		}
		var content map[string]any
		if err := json.Unmarshal(out.Messages[0].Content, &content); err != nil {
			t.Fatalf("decode content: %v (raw=%s)", err, out.Messages[0].Content)
		}
		if content["text"] != "hello" {
			t.Fatalf("content: %+v", content)
		}
	})
}

// ---------------------------------------------------------------------------
// HTTP Basic auth — every /v1/Assistants request carries the configured
// AccountSid + APIKey in the Authorization header (the spec's account
// resolution mechanism, since /v1/ paths omit /Accounts/{Sid}).
// ---------------------------------------------------------------------------

func TestAssistantsV1AuthHeader(t *testing.T) {
	steps := []handlerStep{jsonStep(200, map[string]any{"assistants": []any{}, "meta": map[string]any{}})}
	c, rec, done := newClient(t, steps, nil)
	defer done()
	if _, err := c.AssistantsV1.ListAssistants(context.Background(), voiceml.ListAssistantsParams{}); err != nil {
		t.Fatalf("ListAssistants: %v", err)
	}
	if got := rec.requests[0].Header.Get("Authorization"); got != basicAuthHeader() {
		t.Fatalf("Authorization: want %q, got %q", basicAuthHeader(), got)
	}
}
