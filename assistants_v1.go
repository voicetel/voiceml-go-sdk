// Assistants v1 (assistants.voiceml.voicetel.com / Twilio AI-Assistants
// product surface, /v1/Assistants and friends) — 7 resource families, 30
// operations under /v1/, no /Accounts/{AccountSid} segment (account is
// resolved from HTTP Basic auth).
//
// Resources:
//
//   - Assistant       (aia_asst_*)   CRUD + fetch-with-tools-and-knowledge
//   - Tool            (aia_tool_*)   CRUD + attach/detach + assistant-scoped list
//   - Knowledge       (aia_know_*)   CRUD + status + chunks +
//                                    attach/detach + assistant-scoped list
//   - Session         (sess_*)       list / fetch / list-messages, plus
//                                    POST /v1/Assistants/{id}/Messages
//                                    (creates or resumes a Session)
//   - Message         (aia_msg_*)    Session-scoped list (read-only)
//   - Feedback        (aia_fdbk_*)   Assistant-scoped create + list
//   - Policy          (aia_plcy_*)   list (filterable by ToolId / KnowledgeId)
//
// Wire format is JSON request bodies (unlike the Conversations v1 /
// Voice v1 form-encoded surfaces). Object-typed fields where the spec is
// open (`type: object` without a fixed property list) are exposed as
// json.RawMessage so callers can encode/decode arbitrary nested shapes
// without forcing an `any` / `interface{}` into the public API.

package voiceml

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Helpers local to the Assistants v1 surface.
// ---------------------------------------------------------------------------

// Float64 returns a pointer to f. Use when constructing request structs with
// optional float64 fields (e.g. AssistantsV1 Feedback score).
func Float64(f float64) *float64 { return &f }

// ---------------------------------------------------------------------------
// Response models.
// ---------------------------------------------------------------------------

// AssistantsV1CustomerAI is the Customer-AI sub-object embedded in Assistant
// payloads. The spec names two well-known booleans; pointers distinguish
// "explicitly off" (`false`) from "unset" (`nil`).
type AssistantsV1CustomerAI struct {
	PerceptionEngineEnabled      *bool `json:"perception_engine_enabled,omitempty"`
	PersonalizationEngineEnabled *bool `json:"personalization_engine_enabled,omitempty"`
}

// AssistantsV1Assistant is the list/CRUD shape for an Assistant (aia_asst_*).
// The fetch-by-id endpoint returns AssistantsV1AssistantWithToolsAndKnowledge
// instead; this leaner shape backs list rows, create, and update responses.
type AssistantsV1Assistant struct {
	AccountSid        *string                 `json:"account_sid"`
	ID                *string                 `json:"id"`
	Name              *string                 `json:"name"`
	Owner             *string                 `json:"owner"`
	Model             *string                 `json:"model"`
	PersonalityPrompt *string                 `json:"personality_prompt"`
	CustomerAI        *AssistantsV1CustomerAI `json:"customer_ai,omitempty"`
	URL               *string                 `json:"url"`
	DateCreated       *string                 `json:"date_created"`
	DateUpdated       *string                 `json:"date_updated"`
}

// AssistantsV1AssistantList is the paginated /v1/Assistants response.
type AssistantsV1AssistantList struct {
	Assistants []AssistantsV1Assistant `json:"assistants"`
	Meta       VoiceV1Meta             `json:"meta"`
}

// AssistantsV1AssistantWithToolsAndKnowledge is the fetch-by-id payload:
// the Assistant scalar fields plus its attached Tools and Knowledge inline.
type AssistantsV1AssistantWithToolsAndKnowledge struct {
	AccountSid        *string                 `json:"account_sid"`
	ID                *string                 `json:"id"`
	Name              *string                 `json:"name"`
	Owner             *string                 `json:"owner"`
	Model             *string                 `json:"model"`
	PersonalityPrompt *string                 `json:"personality_prompt"`
	CustomerAI        *AssistantsV1CustomerAI `json:"customer_ai,omitempty"`
	URL               *string                 `json:"url"`
	Tools             []AssistantsV1Tool      `json:"tools,omitempty"`
	Knowledge         []AssistantsV1Knowledge `json:"knowledge,omitempty"`
	DateCreated       *string                 `json:"date_created"`
	DateUpdated       *string                 `json:"date_updated"`
}

// AssistantsV1Tool is the list/CRUD shape for a Tool (aia_tool_*).
// `Meta` is the spec's open `meta: object` — exposed as json.RawMessage so
// callers can decode it into a type they control.
type AssistantsV1Tool struct {
	AccountSid   *string         `json:"account_sid,omitempty"`
	ID           *string         `json:"id"`
	Name         *string         `json:"name"`
	Type         *string         `json:"type"`
	Description  *string         `json:"description"`
	Enabled      *bool           `json:"enabled"`
	RequiresAuth *bool           `json:"requires_auth"`
	Meta         json.RawMessage `json:"meta,omitempty"`
	URL          *string         `json:"url,omitempty"`
	DateCreated  *string         `json:"date_created"`
	DateUpdated  *string         `json:"date_updated"`
}

// AssistantsV1ToolList is the paginated /v1/Tools response.
type AssistantsV1ToolList struct {
	Tools []AssistantsV1Tool `json:"tools"`
	Meta  VoiceV1Meta        `json:"meta"`
}

// AssistantsV1ToolWithPolicies is the fetch-by-id payload: the Tool scalar
// fields plus its inline Policy list.
type AssistantsV1ToolWithPolicies struct {
	AccountSid   *string              `json:"account_sid,omitempty"`
	ID           *string              `json:"id"`
	Name         *string              `json:"name"`
	Type         *string              `json:"type"`
	Description  *string              `json:"description"`
	Enabled      *bool                `json:"enabled"`
	RequiresAuth *bool                `json:"requires_auth"`
	Meta         json.RawMessage      `json:"meta,omitempty"`
	URL          *string              `json:"url,omitempty"`
	Policies     []AssistantsV1Policy `json:"policies,omitempty"`
	DateCreated  *string              `json:"date_created"`
	DateUpdated  *string              `json:"date_updated"`
}

// AssistantsV1Knowledge is a knowledge source (aia_know_*) — an ingestible
// document collection an Assistant can be grounded on.
type AssistantsV1Knowledge struct {
	AccountSid             *string         `json:"account_sid,omitempty"`
	ID                     *string         `json:"id"`
	Name                   *string         `json:"name"`
	Type                   *string         `json:"type"`
	Description            *string         `json:"description,omitempty"`
	Status                 *string         `json:"status,omitempty"`
	EmbeddingModel         *string         `json:"embedding_model,omitempty"`
	KnowledgeSourceDetails json.RawMessage `json:"knowledge_source_details,omitempty"`
	URL                    *string         `json:"url,omitempty"`
	DateCreated            *string         `json:"date_created"`
	DateUpdated            *string         `json:"date_updated"`
}

// AssistantsV1KnowledgeList is the paginated /v1/Knowledge response. Note
// the spec keys it as the singular `knowledge` (not "knowledges").
type AssistantsV1KnowledgeList struct {
	Knowledge []AssistantsV1Knowledge `json:"knowledge"`
	Meta      VoiceV1Meta             `json:"meta"`
}

// AssistantsV1KnowledgeStatus is the read-only ingestion status for a
// Knowledge source. Status / LastStatus walk an ingestion state machine
// (e.g. queued → indexing → ready).
type AssistantsV1KnowledgeStatus struct {
	AccountSid  *string `json:"account_sid,omitempty"`
	Status      *string `json:"status"`
	LastStatus  *string `json:"last_status,omitempty"`
	DateUpdated *string `json:"date_updated,omitempty"`
}

// AssistantsV1KnowledgeChunk is a single ingested chunk row — the GET
// /v1/Knowledge/{id}/Chunks read-only listing. Metadata is open-shaped.
type AssistantsV1KnowledgeChunk struct {
	AccountSid  *string         `json:"account_sid,omitempty"`
	Content     *string         `json:"content,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	DateCreated *string         `json:"date_created,omitempty"`
	DateUpdated *string         `json:"date_updated,omitempty"`
}

// AssistantsV1KnowledgeChunkList is the paginated /v1/Knowledge/{id}/Chunks response.
type AssistantsV1KnowledgeChunkList struct {
	Chunks []AssistantsV1KnowledgeChunk `json:"chunks"`
	Meta   VoiceV1Meta                  `json:"meta"`
}

// AssistantsV1Session is a stateful chat session between a caller-identified
// user and an Assistant. Sessions are created implicitly by SendMessage.
type AssistantsV1Session struct {
	ID          *string `json:"id"`
	AccountSid  *string `json:"account_sid,omitempty"`
	AssistantID *string `json:"assistant_id,omitempty"`
	Verified    *bool   `json:"verified,omitempty"`
	Identity    *string `json:"identity,omitempty"`
	DateCreated *string `json:"date_created,omitempty"`
	DateUpdated *string `json:"date_updated,omitempty"`
}

// AssistantsV1SessionList is the paginated /v1/Sessions response.
type AssistantsV1SessionList struct {
	Sessions []AssistantsV1Session `json:"sessions"`
	Meta     VoiceV1Meta           `json:"meta"`
}

// AssistantsV1Message is one record in a Session's message history.
// Content + Meta are open-shaped (provider-specific payloads).
type AssistantsV1Message struct {
	ID          *string         `json:"id"`
	AccountSid  *string         `json:"account_sid,omitempty"`
	AssistantID *string         `json:"assistant_id,omitempty"`
	SessionID   *string         `json:"session_id,omitempty"`
	Identity    *string         `json:"identity,omitempty"`
	Role        *string         `json:"role,omitempty"`
	Content     json.RawMessage `json:"content,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	DateCreated *string         `json:"date_created,omitempty"`
	DateUpdated *string         `json:"date_updated,omitempty"`
}

// AssistantsV1MessageList is the paginated /v1/Sessions/{id}/Messages response.
type AssistantsV1MessageList struct {
	Messages []AssistantsV1Message `json:"messages"`
	Meta     VoiceV1Meta           `json:"meta"`
}

// AssistantsV1SendMessageResponse is the result of POST
// /v1/Assistants/{id}/Messages. `Body` carries the assistant's reply when
// Mode requests a synchronous response; `Status` reports the result state
// (e.g. "completed", "queued"). Flagged/Aborted are moderation signals.
type AssistantsV1SendMessageResponse struct {
	Status     *string `json:"status"`
	Flagged    *bool   `json:"flagged,omitempty"`
	Aborted    *bool   `json:"aborted,omitempty"`
	SessionID  *string `json:"session_id"`
	AccountSid *string `json:"account_sid"`
	Body       *string `json:"body,omitempty"`
	Error      *string `json:"error,omitempty"`
}

// AssistantsV1Feedback is a per-Session / per-Message quality signal.
// Score is in [0, 1]; the spec exposes it as a float.
type AssistantsV1Feedback struct {
	ID          *string  `json:"id"`
	AssistantID *string  `json:"assistant_id"`
	AccountSid  *string  `json:"account_sid,omitempty"`
	UserSid     *string  `json:"user_sid,omitempty"`
	SessionID   *string  `json:"session_id"`
	MessageID   *string  `json:"message_id"`
	Score       *float64 `json:"score"`
	Text        *string  `json:"text"`
	DateCreated *string  `json:"date_created"`
	DateUpdated *string  `json:"date_updated"`
}

// AssistantsV1FeedbackList is the paginated /v1/Assistants/{id}/Feedbacks response.
type AssistantsV1FeedbackList struct {
	Feedbacks []AssistantsV1Feedback `json:"feedbacks"`
	Meta      VoiceV1Meta            `json:"meta"`
}

// AssistantsV1Policy is a Tool- or Knowledge-scoped access policy. The
// schema is open (`policy_details: object`) so the payload is surfaced as
// json.RawMessage and decoded by the caller against the policy type.
type AssistantsV1Policy struct {
	ID            *string         `json:"id,omitempty"`
	Name          *string         `json:"name,omitempty"`
	Description   *string         `json:"description,omitempty"`
	AccountSid    *string         `json:"account_sid,omitempty"`
	UserSid       *string         `json:"user_sid,omitempty"`
	Type          *string         `json:"type"`
	PolicyDetails json.RawMessage `json:"policy_details"`
	DateCreated   *string         `json:"date_created,omitempty"`
	DateUpdated   *string         `json:"date_updated,omitempty"`
}

// AssistantsV1PolicyList is the paginated /v1/Policies response.
type AssistantsV1PolicyList struct {
	Policies []AssistantsV1Policy `json:"policies"`
	Meta     VoiceV1Meta          `json:"meta"`
}

// ---------------------------------------------------------------------------
// Request bodies (all JSON; pointer fields = optional → omitted when nil).
// ---------------------------------------------------------------------------

// CreateAssistantRequest is the body for POST /v1/Assistants. `Name` is the
// only spec-required field. SegmentCredential is open-shaped → json.RawMessage.
type CreateAssistantRequest struct {
	Name              string                  `json:"name"`
	Owner             *string                 `json:"owner,omitempty"`
	PersonalityPrompt *string                 `json:"personality_prompt,omitempty"`
	Model             *string                 `json:"model,omitempty"`
	CustomerAI        *AssistantsV1CustomerAI `json:"customer_ai,omitempty"`
	SegmentCredential json.RawMessage         `json:"segment_credential,omitempty"`
}

// UpdateAssistantRequest is the body for PUT /v1/Assistants/{id}. All fields
// are optional; the server merges the set ones onto the existing row.
type UpdateAssistantRequest struct {
	Name              *string                 `json:"name,omitempty"`
	Owner             *string                 `json:"owner,omitempty"`
	PersonalityPrompt *string                 `json:"personality_prompt,omitempty"`
	Model             *string                 `json:"model,omitempty"`
	CustomerAI        *AssistantsV1CustomerAI `json:"customer_ai,omitempty"`
	SegmentCredential json.RawMessage         `json:"segment_credential,omitempty"`
}

// ListAssistantsParams is the query for GET /v1/Assistants. The endpoint
// accepts PageSize plus the cursor-style Page / PageToken knobs.
type ListAssistantsParams struct {
	PageSize  *int
	Page      *int
	PageToken *string
}

func (p ListAssistantsParams) query() url.Values {
	v := url.Values{}
	if p.PageSize != nil {
		v.Set("PageSize", strconv.Itoa(*p.PageSize))
	}
	if p.Page != nil {
		v.Set("Page", strconv.Itoa(*p.Page))
	}
	if p.PageToken != nil {
		v.Set("PageToken", *p.PageToken)
	}
	return v
}

// CreateToolRequest is the body for POST /v1/Tools. `Name`, `Type`, and
// `Enabled` are spec-required. `AssistantID` optionally attaches the new
// Tool to an Assistant at creation time.
type CreateToolRequest struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Enabled     bool            `json:"enabled"`
	AssistantID *string         `json:"assistant_id,omitempty"`
	Description *string         `json:"description,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
}

// UpdateToolRequest is the body for PUT /v1/Tools/{id}.
type UpdateToolRequest struct {
	Name        *string         `json:"name,omitempty"`
	Type        *string         `json:"type,omitempty"`
	Enabled     *bool           `json:"enabled,omitempty"`
	Description *string         `json:"description,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
}

// ListToolsParams is the query for GET /v1/Tools. `AssistantID` narrows to
// Tools attached to the given Assistant.
type ListToolsParams struct {
	AssistantID *string
	PageSize    *int
}

func (p ListToolsParams) query() url.Values {
	v := url.Values{}
	if p.AssistantID != nil {
		v.Set("AssistantId", *p.AssistantID)
	}
	if p.PageSize != nil {
		v.Set("PageSize", strconv.Itoa(*p.PageSize))
	}
	return v
}

// CreateKnowledgeRequest is the body for POST /v1/Knowledge. `Name` and
// `Type` are required; the rest configure ingestion shape and source.
type CreateKnowledgeRequest struct {
	Name                   string          `json:"name"`
	Type                   string          `json:"type"`
	AssistantID            *string         `json:"assistant_id,omitempty"`
	Description            *string         `json:"description,omitempty"`
	EmbeddingModel         *string         `json:"embedding_model,omitempty"`
	KnowledgeSourceDetails json.RawMessage `json:"knowledge_source_details,omitempty"`
}

// UpdateKnowledgeRequest is the body for PUT /v1/Knowledge/{id}.
type UpdateKnowledgeRequest struct {
	Name                   *string         `json:"name,omitempty"`
	Type                   *string         `json:"type,omitempty"`
	Description            *string         `json:"description,omitempty"`
	EmbeddingModel         *string         `json:"embedding_model,omitempty"`
	KnowledgeSourceDetails json.RawMessage `json:"knowledge_source_details,omitempty"`
}

// ListKnowledgeParams is the query for GET /v1/Knowledge. `AssistantID`
// narrows to Knowledge attached to the given Assistant.
type ListKnowledgeParams struct {
	AssistantID *string
	PageSize    *int
}

func (p ListKnowledgeParams) query() url.Values {
	v := url.Values{}
	if p.AssistantID != nil {
		v.Set("AssistantId", *p.AssistantID)
	}
	if p.PageSize != nil {
		v.Set("PageSize", strconv.Itoa(*p.PageSize))
	}
	return v
}

// SendMessageRequest is the body for POST /v1/Assistants/{id}/Messages.
// `Identity` and `Body` are required. `SessionID` resumes an existing chat
// session; omit it to start a new one. `Mode` controls sync vs. async reply.
type SendMessageRequest struct {
	Identity  string  `json:"identity"`
	Body      string  `json:"body"`
	SessionID *string `json:"session_id,omitempty"`
	Webhook   *string `json:"webhook,omitempty"`
	Mode      *string `json:"mode,omitempty"`
}

// CreateFeedbackRequest is the body for POST /v1/Assistants/{id}/Feedbacks.
// `SessionID` is the only spec-required field; Score is in [0, 1].
type CreateFeedbackRequest struct {
	SessionID string   `json:"session_id"`
	MessageID *string  `json:"message_id,omitempty"`
	Score     *float64 `json:"score,omitempty"`
	Text      *string  `json:"text,omitempty"`
}

// ListPoliciesParams is the query for GET /v1/Policies. Either ToolID or
// KnowledgeID narrows to policies scoped to a specific Tool or Knowledge.
type ListPoliciesParams struct {
	ToolID      *string
	KnowledgeID *string
	PageSize    *int
}

func (p ListPoliciesParams) query() url.Values {
	v := url.Values{}
	if p.ToolID != nil {
		v.Set("ToolId", *p.ToolID)
	}
	if p.KnowledgeID != nil {
		v.Set("KnowledgeId", *p.KnowledgeID)
	}
	if p.PageSize != nil {
		v.Set("PageSize", strconv.Itoa(*p.PageSize))
	}
	return v
}

// ---------------------------------------------------------------------------
// Service — flat-method facade for the entire Assistants v1 surface.
// ---------------------------------------------------------------------------

// AssistantsV1Service exposes the /v1/Assistants product surface. Reach it
// as c.AssistantsV1. Methods are named verb+resource (CreateAssistant,
// ListTools, AttachToolToAssistant, ...) so the surface stays flat for
// IDE discovery.
type AssistantsV1Service struct{ c *Client }

// --- Assistants ------------------------------------------------------------

// CreateAssistant creates an Assistant. `Name` is required.
func (s *AssistantsV1Service) CreateAssistant(ctx context.Context, params *CreateAssistantRequest) (*AssistantsV1Assistant, error) {
	var out AssistantsV1Assistant
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Assistants", json: params,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAssistants returns a single page of Assistants.
func (s *AssistantsV1Service) ListAssistants(ctx context.Context, params ListAssistantsParams) (*AssistantsV1AssistantList, error) {
	var out AssistantsV1AssistantList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Assistants", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchAssistant retrieves an Assistant by id, with its attached Tools and
// Knowledge inlined into the response.
func (s *AssistantsV1Service) FetchAssistant(ctx context.Context, assistantID string) (*AssistantsV1AssistantWithToolsAndKnowledge, error) {
	var out AssistantsV1AssistantWithToolsAndKnowledge
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Assistants/" + assistantID,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAssistant mutates an Assistant in place. The spec uses PUT (not
// POST) for the Assistants v1 update endpoints.
func (s *AssistantsV1Service) UpdateAssistant(ctx context.Context, assistantID string, params *UpdateAssistantRequest) (*AssistantsV1Assistant, error) {
	var out AssistantsV1Assistant
	if err := s.c.t.do(ctx, requestOpts{
		method: "PUT", path: "/v1/Assistants/" + assistantID, json: params,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAssistant removes an Assistant.
func (s *AssistantsV1Service) DeleteAssistant(ctx context.Context, assistantID string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Assistants/" + assistantID,
	}, nil)
}

// --- Tools -----------------------------------------------------------------

// CreateTool creates a Tool. `Name`, `Type`, and `Enabled` are required;
// AssistantID optionally attaches the new Tool at creation time.
func (s *AssistantsV1Service) CreateTool(ctx context.Context, params *CreateToolRequest) (*AssistantsV1Tool, error) {
	var out AssistantsV1Tool
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Tools", json: params,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTools returns a single page of Tools, optionally filtered by AssistantID.
func (s *AssistantsV1Service) ListTools(ctx context.Context, params ListToolsParams) (*AssistantsV1ToolList, error) {
	var out AssistantsV1ToolList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Tools", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchTool retrieves a Tool by id, with its scoped Policies inlined.
func (s *AssistantsV1Service) FetchTool(ctx context.Context, toolID string) (*AssistantsV1ToolWithPolicies, error) {
	var out AssistantsV1ToolWithPolicies
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Tools/" + toolID,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTool mutates a Tool in place.
func (s *AssistantsV1Service) UpdateTool(ctx context.Context, toolID string, params *UpdateToolRequest) (*AssistantsV1Tool, error) {
	var out AssistantsV1Tool
	if err := s.c.t.do(ctx, requestOpts{
		method: "PUT", path: "/v1/Tools/" + toolID, json: params,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTool removes a Tool.
func (s *AssistantsV1Service) DeleteTool(ctx context.Context, toolID string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Tools/" + toolID,
	}, nil)
}

// ListAssistantTools returns a page of Tools attached to a specific
// Assistant. The endpoint is /v1/Assistants/{id}/Tools — equivalent to
// ListTools with AssistantID set, but the SDK exposes both for parity
// with the spec's two operation IDs.
func (s *AssistantsV1Service) ListAssistantTools(ctx context.Context, assistantID string, params V1PageParams) (*AssistantsV1ToolList, error) {
	var out AssistantsV1ToolList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Assistants/" + assistantID + "/Tools", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AttachToolToAssistant attaches an existing Tool to an Assistant. 204 success.
func (s *AssistantsV1Service) AttachToolToAssistant(ctx context.Context, assistantID, toolID string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Assistants/" + assistantID + "/Tools/" + toolID,
	}, nil)
}

// DetachToolFromAssistant detaches a Tool from an Assistant. 204 success.
func (s *AssistantsV1Service) DetachToolFromAssistant(ctx context.Context, assistantID, toolID string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Assistants/" + assistantID + "/Tools/" + toolID,
	}, nil)
}

// --- Knowledge -------------------------------------------------------------

// CreateKnowledge creates a Knowledge source. `Name` and `Type` are required.
func (s *AssistantsV1Service) CreateKnowledge(ctx context.Context, params *CreateKnowledgeRequest) (*AssistantsV1Knowledge, error) {
	var out AssistantsV1Knowledge
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Knowledge", json: params,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListKnowledge returns a single page of Knowledge sources, optionally
// filtered by AssistantID.
func (s *AssistantsV1Service) ListKnowledge(ctx context.Context, params ListKnowledgeParams) (*AssistantsV1KnowledgeList, error) {
	var out AssistantsV1KnowledgeList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Knowledge", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchKnowledge retrieves a Knowledge source by id.
func (s *AssistantsV1Service) FetchKnowledge(ctx context.Context, knowledgeID string) (*AssistantsV1Knowledge, error) {
	var out AssistantsV1Knowledge
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Knowledge/" + knowledgeID,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateKnowledge mutates a Knowledge source in place.
func (s *AssistantsV1Service) UpdateKnowledge(ctx context.Context, knowledgeID string, params *UpdateKnowledgeRequest) (*AssistantsV1Knowledge, error) {
	var out AssistantsV1Knowledge
	if err := s.c.t.do(ctx, requestOpts{
		method: "PUT", path: "/v1/Knowledge/" + knowledgeID, json: params,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteKnowledge removes a Knowledge source.
func (s *AssistantsV1Service) DeleteKnowledge(ctx context.Context, knowledgeID string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Knowledge/" + knowledgeID,
	}, nil)
}

// FetchKnowledgeStatus retrieves the read-only ingestion status for a
// Knowledge source.
func (s *AssistantsV1Service) FetchKnowledgeStatus(ctx context.Context, knowledgeID string) (*AssistantsV1KnowledgeStatus, error) {
	var out AssistantsV1KnowledgeStatus
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Knowledge/" + knowledgeID + "/Status",
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListKnowledgeChunks returns a page of ingested chunks for a Knowledge source.
func (s *AssistantsV1Service) ListKnowledgeChunks(ctx context.Context, knowledgeID string, params V1PageParams) (*AssistantsV1KnowledgeChunkList, error) {
	var out AssistantsV1KnowledgeChunkList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Knowledge/" + knowledgeID + "/Chunks", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAssistantKnowledge returns a page of Knowledge attached to a specific
// Assistant. Mirrors ListKnowledge with AssistantID set.
func (s *AssistantsV1Service) ListAssistantKnowledge(ctx context.Context, assistantID string, params V1PageParams) (*AssistantsV1KnowledgeList, error) {
	var out AssistantsV1KnowledgeList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Assistants/" + assistantID + "/Knowledge", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AttachKnowledgeToAssistant attaches existing Knowledge to an Assistant. 204 success.
func (s *AssistantsV1Service) AttachKnowledgeToAssistant(ctx context.Context, assistantID, knowledgeID string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Assistants/" + assistantID + "/Knowledge/" + knowledgeID,
	}, nil)
}

// DetachKnowledgeFromAssistant detaches Knowledge from an Assistant. 204 success.
func (s *AssistantsV1Service) DetachKnowledgeFromAssistant(ctx context.Context, assistantID, knowledgeID string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Assistants/" + assistantID + "/Knowledge/" + knowledgeID,
	}, nil)
}

// --- Sessions + Messages ---------------------------------------------------

// SendMessage posts a user message to an Assistant. The endpoint creates a
// new Session when SessionID is unset, or resumes the named one when set.
// The response's SessionID echoes the one used.
func (s *AssistantsV1Service) SendMessage(ctx context.Context, assistantID string, params *SendMessageRequest) (*AssistantsV1SendMessageResponse, error) {
	var out AssistantsV1SendMessageResponse
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Assistants/" + assistantID + "/Messages", json: params,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSessions returns a single page of Sessions for the authenticated account.
func (s *AssistantsV1Service) ListSessions(ctx context.Context, params V1PageParams) (*AssistantsV1SessionList, error) {
	var out AssistantsV1SessionList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Sessions", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchSession retrieves a Session by id.
func (s *AssistantsV1Service) FetchSession(ctx context.Context, sessionID string) (*AssistantsV1Session, error) {
	var out AssistantsV1Session
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Sessions/" + sessionID,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSessionMessages returns a single page of a Session's Messages.
func (s *AssistantsV1Service) ListSessionMessages(ctx context.Context, sessionID string, params V1PageParams) (*AssistantsV1MessageList, error) {
	var out AssistantsV1MessageList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Sessions/" + sessionID + "/Messages", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Feedback --------------------------------------------------------------

// ListAssistantFeedback returns a page of Feedback rows for an Assistant.
func (s *AssistantsV1Service) ListAssistantFeedback(ctx context.Context, assistantID string, params V1PageParams) (*AssistantsV1FeedbackList, error) {
	var out AssistantsV1FeedbackList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Assistants/" + assistantID + "/Feedbacks", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateFeedback creates a Feedback row scoped to an Assistant. SessionID
// is required; Score is a float in [0, 1].
func (s *AssistantsV1Service) CreateFeedback(ctx context.Context, assistantID string, params *CreateFeedbackRequest) (*AssistantsV1Feedback, error) {
	var out AssistantsV1Feedback
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Assistants/" + assistantID + "/Feedbacks", json: params,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Policies --------------------------------------------------------------

// ListPolicies returns a single page of Policies, optionally filtered by
// ToolID or KnowledgeID.
func (s *AssistantsV1Service) ListPolicies(ctx context.Context, params ListPoliciesParams) (*AssistantsV1PolicyList, error) {
	var out AssistantsV1PolicyList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Policies", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
