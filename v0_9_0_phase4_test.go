// v0.9.0 Phase 4 surface tests — service-scoped Conversations v1 (the
// /v1/Services/{ChatServiceSid}/ resource family).
//
// Each case stands up an httptest server, asserts the request the SDK sent
// (method + path + form/query encoding), and feeds back a canned wire shape
// decoded into the response model. Decode coverage is checked in
// TestConversationsV1ServiceDecodes below.

package voiceml_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

// Phase 4 sids — separate constants so collisions with v0_9_0_test.go
// stay obvious.
const (
	svcBindSid = "BS" + "0123456789abcdef0123456789abcdef"
)

// shortcut: build the service-scoped base path so per-case literals stay readable.
func svcBase() string { return "/v1/Services/" + convChatSvc }

func TestConversationsV1ServicePaths(t *testing.T) {
	convPayload := map[string]any{
		"sid": convSid, "account_sid": testAccountSid, "chat_service_sid": convChatSvc,
		"state": "active", "attributes": "{}",
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	msgPayload := map[string]any{
		"sid": convMsgSid, "conversation_sid": convSid, "account_sid": testAccountSid,
		"chat_service_sid": convChatSvc, "index": 0, "attributes": "{}",
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	partPayload := map[string]any{
		"sid": convPartSid, "conversation_sid": convSid, "account_sid": testAccountSid,
		"chat_service_sid": convChatSvc, "attributes": "{}",
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	hookPayload := map[string]any{
		"sid": convHookSid, "conversation_sid": convSid, "account_sid": testAccountSid,
		"chat_service_sid": convChatSvc,
		"date_created":     "x", "date_updated": "x",
	}
	rcptPayload := map[string]any{
		"sid": convRcptSid, "conversation_sid": convSid, "account_sid": testAccountSid,
		"chat_service_sid": convChatSvc, "message_sid": convMsgSid,
		"status": "delivered", "error_code": 0,
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	rolePayload := map[string]any{
		"sid": convRoleSid, "account_sid": testAccountSid, "chat_service_sid": convChatSvc,
		"type": "conversation", "permissions": []string{"sendMessage"},
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	userPayload := map[string]any{
		"sid": convUserSid, "account_sid": testAccountSid, "chat_service_sid": convChatSvc,
		"attributes": "{}", "date_created": "x", "date_updated": "x", "url": "u",
	}
	bindPayload := map[string]any{
		"sid": svcBindSid, "account_sid": testAccountSid, "chat_service_sid": convChatSvc,
		"binding_type": "fcm",
	}
	cfgPayload := map[string]any{
		"chat_service_sid": convChatSvc, "url": "u",
	}
	notifPayload := map[string]any{
		"account_sid": testAccountSid, "chat_service_sid": convChatSvc, "url": "u",
	}
	whCfgPayload := map[string]any{
		"account_sid": testAccountSid, "chat_service_sid": convChatSvc,
		"method": "POST", "url": "u",
	}
	ucPayload := map[string]any{
		"account_sid": testAccountSid, "chat_service_sid": convChatSvc,
		"conversation_state": "active", "notification_level": "default",
		"date_created": "x", "date_updated": "x", "url": "u",
	}
	pcPayload := map[string]any{
		"account_sid": testAccountSid, "chat_service_sid": convChatSvc,
		"conversation_state":        "active",
		"conversation_date_created": "x", "conversation_date_updated": "x",
	}

	cases := []struct {
		name       string
		response   map[string]any
		listShape  map[string]any
		method     string
		path       string
		invoke     func(c *voiceml.Client) error
		assertBody func(t *testing.T, body, query url.Values)
	}{
		// --- ServiceConversation ------------------------------------------
		{
			name: "CreateServiceConversation", response: convPayload,
			method: "POST", path: svcBase() + "/Conversations",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateServiceConversation(context.Background(), convChatSvc, voiceml.CreateServiceConversationRequest{
					FriendlyName:   voiceml.String("Support"),
					State:          voiceml.String("active"),
					TimersInactive: voiceml.String("PT1H"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "Support" ||
					body.Get("State") != "active" ||
					body.Get("Timers.Inactive") != "PT1H" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListServiceConversations", listShape: map[string]any{"conversations": []any{}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/Conversations",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServiceConversations(context.Background(), convChatSvc, voiceml.V1PageParams{PageSize: voiceml.Int(25)})
				return err
			},
			assertBody: func(t *testing.T, _, q url.Values) {
				if q.Get("PageSize") != "25" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name: "FetchServiceConversation", response: convPayload,
			method: "GET", path: svcBase() + "/Conversations/" + convSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceConversation(context.Background(), convChatSvc, convSid)
				return err
			},
		},
		{
			name: "UpdateServiceConversation", response: convPayload,
			method: "POST", path: svcBase() + "/Conversations/" + convSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateServiceConversation(context.Background(), convChatSvc, convSid, voiceml.UpdateServiceConversationRequest{
					State: voiceml.String("closed"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("State") != "closed" || len(body) != 1 {
					t.Fatalf("partial body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteServiceConversation",
			method: "DELETE", path: svcBase() + "/Conversations/" + convSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteServiceConversation(context.Background(), convChatSvc, convSid)
			},
		},
		// --- ServiceConversationMessage -----------------------------------
		{
			name: "CreateServiceMessage", response: msgPayload,
			method: "POST", path: svcBase() + "/Conversations/" + convSid + "/Messages",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateServiceMessage(context.Background(), convChatSvc, convSid, voiceml.CreateServiceMessageRequest{
					Author: voiceml.String("alice"), Body: voiceml.String("hello"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Author") != "alice" || body.Get("Body") != "hello" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListServiceMessages", listShape: map[string]any{"messages": []any{}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/Conversations/" + convSid + "/Messages",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServiceMessages(context.Background(), convChatSvc, convSid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchServiceMessage", response: msgPayload,
			method: "GET", path: svcBase() + "/Conversations/" + convSid + "/Messages/" + convMsgSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceMessage(context.Background(), convChatSvc, convSid, convMsgSid)
				return err
			},
		},
		{
			name: "UpdateServiceMessage", response: msgPayload,
			method: "POST", path: svcBase() + "/Conversations/" + convSid + "/Messages/" + convMsgSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateServiceMessage(context.Background(), convChatSvc, convSid, convMsgSid, voiceml.UpdateServiceMessageRequest{
					Body: voiceml.String("edited"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Body") != "edited" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteServiceMessage",
			method: "DELETE", path: svcBase() + "/Conversations/" + convSid + "/Messages/" + convMsgSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteServiceMessage(context.Background(), convChatSvc, convSid, convMsgSid)
			},
		},
		// --- ServiceConversationMessageReceipt ----------------------------
		{
			name: "ListServiceMessageReceipts", listShape: map[string]any{"delivery_receipts": []any{}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/Conversations/" + convSid + "/Messages/" + convMsgSid + "/Receipts",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServiceMessageReceipts(context.Background(), convChatSvc, convSid, convMsgSid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchServiceMessageReceipt", response: rcptPayload,
			method: "GET", path: svcBase() + "/Conversations/" + convSid + "/Messages/" + convMsgSid + "/Receipts/" + convRcptSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceMessageReceipt(context.Background(), convChatSvc, convSid, convMsgSid, convRcptSid)
				return err
			},
		},
		// --- ServiceConversationParticipant -------------------------------
		{
			name: "CreateServiceParticipant", response: partPayload,
			method: "POST", path: svcBase() + "/Conversations/" + convSid + "/Participants",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateServiceParticipant(context.Background(), convChatSvc, convSid, voiceml.CreateServiceParticipantRequest{
					Identity:                voiceml.String("alice"),
					MessagingBindingAddress: voiceml.String("+15551234567"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Identity") != "alice" ||
					body.Get("MessagingBinding.Address") != "+15551234567" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListServiceParticipants", listShape: map[string]any{"participants": []any{}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/Conversations/" + convSid + "/Participants",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServiceParticipants(context.Background(), convChatSvc, convSid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchServiceParticipant", response: partPayload,
			method: "GET", path: svcBase() + "/Conversations/" + convSid + "/Participants/" + convPartSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceParticipant(context.Background(), convChatSvc, convSid, convPartSid)
				return err
			},
		},
		{
			name: "UpdateServiceParticipant", response: partPayload,
			method: "POST", path: svcBase() + "/Conversations/" + convSid + "/Participants/" + convPartSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateServiceParticipant(context.Background(), convChatSvc, convSid, convPartSid, voiceml.UpdateServiceParticipantRequest{
					RoleSid: voiceml.String(convRoleSid),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("RoleSid") != convRoleSid || len(body) != 1 {
					t.Fatalf("partial body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteServiceParticipant",
			method: "DELETE", path: svcBase() + "/Conversations/" + convSid + "/Participants/" + convPartSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteServiceParticipant(context.Background(), convChatSvc, convSid, convPartSid)
			},
		},
		// --- ServiceConversationScopedWebhook -----------------------------
		{
			name: "CreateServiceScopedWebhook", response: hookPayload,
			method: "POST", path: svcBase() + "/Conversations/" + convSid + "/Webhooks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateServiceScopedWebhook(context.Background(), convChatSvc, convSid, voiceml.CreateServiceScopedWebhookRequest{
					Target:              "webhook",
					ConfigurationURL:    voiceml.String("https://example.com/hook"),
					ConfigurationMethod: voiceml.String("POST"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Target") != "webhook" ||
					body.Get("Configuration.Url") != "https://example.com/hook" ||
					body.Get("Configuration.Method") != "POST" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListServiceScopedWebhooks", listShape: map[string]any{"webhooks": []any{}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/Conversations/" + convSid + "/Webhooks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServiceScopedWebhooks(context.Background(), convChatSvc, convSid, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchServiceScopedWebhook", response: hookPayload,
			method: "GET", path: svcBase() + "/Conversations/" + convSid + "/Webhooks/" + convHookSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceScopedWebhook(context.Background(), convChatSvc, convSid, convHookSid)
				return err
			},
		},
		{
			name: "UpdateServiceScopedWebhook", response: hookPayload,
			method: "POST", path: svcBase() + "/Conversations/" + convSid + "/Webhooks/" + convHookSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateServiceScopedWebhook(context.Background(), convChatSvc, convSid, convHookSid, voiceml.UpdateServiceScopedWebhookRequest{
					ConfigurationURL: voiceml.String("https://example.com/new"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Configuration.Url") != "https://example.com/new" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name:   "DeleteServiceScopedWebhook",
			method: "DELETE", path: svcBase() + "/Conversations/" + convSid + "/Webhooks/" + convHookSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteServiceScopedWebhook(context.Background(), convChatSvc, convSid, convHookSid)
			},
		},
		// --- ServiceConversationWithParticipants (create-only) ------------
		{
			name: "CreateServiceConversationWithParticipants", response: convPayload,
			method: "POST", path: svcBase() + "/ConversationWithParticipants",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateServiceConversationWithParticipants(context.Background(), convChatSvc, voiceml.CreateServiceConversationWithParticipantsRequest{
					FriendlyName: voiceml.String("team"),
					Participant: []string{
						`{"identity":"alice"}`,
						`{"messaging_binding":{"address":"+15551234567"}}`,
					},
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "team" {
					t.Fatalf("body: %+v", body)
				}
				parts := body["Participant"]
				if len(parts) != 2 || parts[0] != `{"identity":"alice"}` {
					t.Fatalf("repeated Participant: %+v", parts)
				}
			},
		},
		// --- ServiceParticipantConversation (list-only) -------------------
		{
			name: "ListServiceParticipantConversationsByIdentity", listShape: map[string]any{"conversations": []any{}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/ParticipantConversations",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServiceParticipantConversations(context.Background(), convChatSvc, voiceml.ListServiceParticipantConversationsParams{
					Identity: voiceml.String("alice"),
				})
				return err
			},
			assertBody: func(t *testing.T, _, q url.Values) {
				if q.Get("Identity") != "alice" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name: "ListServiceParticipantConversationsByAddress", listShape: map[string]any{"conversations": []map[string]any{pcPayload}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/ParticipantConversations",
			invoke: func(c *voiceml.Client) error {
				out, err := c.ConversationsV1.ListServiceParticipantConversations(context.Background(), convChatSvc, voiceml.ListServiceParticipantConversationsParams{
					Address:  voiceml.String("+15551234567"),
					PageSize: voiceml.Int(50),
				})
				if err != nil {
					return err
				}
				if len(out.Conversations) != 1 || out.Conversations[0].ConversationState != "active" {
					t.Fatalf("decoded participant-conv: %+v", out.Conversations)
				}
				return nil
			},
			assertBody: func(t *testing.T, _, q url.Values) {
				if q.Get("Address") != "+15551234567" || q.Get("PageSize") != "50" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		// --- ServiceUserConversation (list-only) --------------------------
		{
			name: "ListServiceUserConversations", listShape: map[string]any{"conversations": []map[string]any{ucPayload}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/Users/" + convUserSid + "/Conversations",
			invoke: func(c *voiceml.Client) error {
				out, err := c.ConversationsV1.ListServiceUserConversations(context.Background(), convChatSvc, convUserSid, voiceml.V1PageParams{})
				if err != nil {
					return err
				}
				if len(out.Conversations) != 1 || out.Conversations[0].NotificationLevel != "default" {
					t.Fatalf("decoded user-conv: %+v", out.Conversations)
				}
				return nil
			},
		},
		// --- ServiceRole --------------------------------------------------
		{
			name: "CreateServiceRole", response: rolePayload,
			method: "POST", path: svcBase() + "/Roles",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateServiceRole(context.Background(), convChatSvc, voiceml.CreateServiceRoleRequest{
					FriendlyName: "agent",
					Type:         "conversation",
					Permission:   []string{"sendMessage", "leaveConversation"},
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("FriendlyName") != "agent" || body.Get("Type") != "conversation" {
					t.Fatalf("body: %+v", body)
				}
				perms := body["Permission"]
				if len(perms) != 2 || perms[0] != "sendMessage" || perms[1] != "leaveConversation" {
					t.Fatalf("repeated Permission: %+v", perms)
				}
			},
		},
		{
			name: "ListServiceRoles", listShape: map[string]any{"roles": []any{}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/Roles",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServiceRoles(context.Background(), convChatSvc, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchServiceRole", response: rolePayload,
			method: "GET", path: svcBase() + "/Roles/" + convRoleSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceRole(context.Background(), convChatSvc, convRoleSid)
				return err
			},
		},
		{
			name: "UpdateServiceRole", response: rolePayload,
			method: "POST", path: svcBase() + "/Roles/" + convRoleSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateServiceRole(context.Background(), convChatSvc, convRoleSid, voiceml.UpdateServiceRoleRequest{
					Permission: []string{"sendMessage"},
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if perms := body["Permission"]; len(perms) != 1 || perms[0] != "sendMessage" {
					t.Fatalf("perms: %+v", perms)
				}
			},
		},
		{
			name:   "DeleteServiceRole",
			method: "DELETE", path: svcBase() + "/Roles/" + convRoleSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteServiceRole(context.Background(), convChatSvc, convRoleSid)
			},
		},
		// --- ServiceUser --------------------------------------------------
		{
			name: "CreateServiceUser", response: userPayload,
			method: "POST", path: svcBase() + "/Users",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.CreateServiceUser(context.Background(), convChatSvc, voiceml.CreateServiceUserRequest{
					Identity: "alice", FriendlyName: voiceml.String("Alice"),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("Identity") != "alice" || body.Get("FriendlyName") != "Alice" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		{
			name: "ListServiceUsers", listShape: map[string]any{"users": []any{}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/Users",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServiceUsers(context.Background(), convChatSvc, voiceml.V1PageParams{})
				return err
			},
		},
		{
			name: "FetchServiceUser", response: userPayload,
			method: "GET", path: svcBase() + "/Users/" + convUserSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceUser(context.Background(), convChatSvc, convUserSid)
				return err
			},
		},
		{
			name: "UpdateServiceUser", response: userPayload,
			method: "POST", path: svcBase() + "/Users/" + convUserSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateServiceUser(context.Background(), convChatSvc, convUserSid, voiceml.UpdateServiceUserRequest{
					FriendlyName: voiceml.String("Alice II"),
				})
				return err
			},
		},
		{
			name:   "DeleteServiceUser",
			method: "DELETE", path: svcBase() + "/Users/" + convUserSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteServiceUser(context.Background(), convChatSvc, convUserSid)
			},
		},
		// --- ServiceBinding -----------------------------------------------
		{
			name: "ListServiceBindings", listShape: map[string]any{"bindings": []any{}, "meta": map[string]any{}},
			method: "GET", path: svcBase() + "/Bindings",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.ListServiceBindings(context.Background(), convChatSvc, voiceml.ListServiceBindingsParams{
					BindingType: voiceml.String("fcm"),
					Identity:    voiceml.String("alice"),
					PageSize:    voiceml.Int(10),
				})
				return err
			},
			assertBody: func(t *testing.T, _, q url.Values) {
				if q.Get("BindingType") != "fcm" || q.Get("Identity") != "alice" || q.Get("PageSize") != "10" {
					t.Fatalf("query: %+v", q)
				}
			},
		},
		{
			name: "FetchServiceBinding", response: bindPayload,
			method: "GET", path: svcBase() + "/Bindings/" + svcBindSid,
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceBinding(context.Background(), convChatSvc, svcBindSid)
				return err
			},
		},
		{
			name:   "DeleteServiceBinding",
			method: "DELETE", path: svcBase() + "/Bindings/" + svcBindSid,
			invoke: func(c *voiceml.Client) error {
				return c.ConversationsV1.DeleteServiceBinding(context.Background(), convChatSvc, svcBindSid)
			},
		},
		// --- ServiceConfiguration -----------------------------------------
		{
			name: "FetchServiceConfiguration", response: cfgPayload,
			method: "GET", path: svcBase() + "/Configuration",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceConfiguration(context.Background(), convChatSvc)
				return err
			},
		},
		{
			name: "UpdateServiceConfiguration", response: cfgPayload,
			method: "POST", path: svcBase() + "/Configuration",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateServiceConfiguration(context.Background(), convChatSvc, voiceml.UpdateServiceConfigurationRequest{
					DefaultConversationCreatorRoleSid: voiceml.String(convRoleSid),
					ReachabilityEnabled:               voiceml.Bool(true),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("DefaultConversationCreatorRoleSid") != convRoleSid ||
					body.Get("ReachabilityEnabled") != "true" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		// --- ServiceNotification ------------------------------------------
		{
			name: "FetchServiceNotification", response: notifPayload,
			method: "GET", path: svcBase() + "/Configuration/Notifications",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceNotification(context.Background(), convChatSvc)
				return err
			},
		},
		{
			name: "UpdateServiceNotification", response: notifPayload,
			method: "POST", path: svcBase() + "/Configuration/Notifications",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateServiceNotification(context.Background(), convChatSvc, voiceml.UpdateServiceNotificationRequest{
					LogEnabled:                  voiceml.Bool(true),
					NewMessageEnabled:           voiceml.Bool(true),
					NewMessageTemplate:          voiceml.String("${PARTICIPANT}: ${MESSAGE}"),
					NewMessageBadgeCountEnabled: voiceml.Bool(false),
					NewMessageWithMediaEnabled:  voiceml.Bool(true),
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("LogEnabled") != "true" ||
					body.Get("NewMessage.Enabled") != "true" ||
					body.Get("NewMessage.Template") != "${PARTICIPANT}: ${MESSAGE}" ||
					body.Get("NewMessage.BadgeCountEnabled") != "false" ||
					body.Get("NewMessage.WithMedia.Enabled") != "true" {
					t.Fatalf("body: %+v", body)
				}
			},
		},
		// --- ServiceWebhookConfiguration ----------------------------------
		{
			name: "FetchServiceWebhookConfiguration", response: whCfgPayload,
			method: "GET", path: svcBase() + "/Configuration/Webhooks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.FetchServiceWebhookConfiguration(context.Background(), convChatSvc)
				return err
			},
		},
		{
			name: "UpdateServiceWebhookConfiguration", response: whCfgPayload,
			method: "POST", path: svcBase() + "/Configuration/Webhooks",
			invoke: func(c *voiceml.Client) error {
				_, err := c.ConversationsV1.UpdateServiceWebhookConfiguration(context.Background(), convChatSvc, voiceml.UpdateServiceWebhookConfigurationRequest{
					PreWebhookURL:  voiceml.String("https://example.com/pre"),
					PostWebhookURL: voiceml.String("https://example.com/post"),
					Method:         voiceml.String("POST"),
					Filters:        []string{"onMessageAdded", "onConversationAdded"},
				})
				return err
			},
			assertBody: func(t *testing.T, body, _ url.Values) {
				if body.Get("PreWebhookUrl") != "https://example.com/pre" ||
					body.Get("PostWebhookUrl") != "https://example.com/post" ||
					body.Get("Method") != "POST" {
					t.Fatalf("body: %+v", body)
				}
				filters := body["Filters"]
				if len(filters) != 2 || filters[0] != "onMessageAdded" || filters[1] != "onConversationAdded" {
					t.Fatalf("filters: %+v", filters)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var responseBody any
			switch {
			case tc.listShape != nil:
				responseBody = tc.listShape
			case tc.response != nil:
				responseBody = tc.response
			default:
				responseBody = nil
			}
			steps := []handlerStep{jsonStep(200, responseBody)}
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
			if strings.Contains(r.Path, testAccountSid) {
				t.Fatalf("path leaked account sid: %q", r.Path)
			}
			if tc.assertBody != nil {
				formBody, _ := url.ParseQuery(string(r.Body))
				query, _ := url.ParseQuery(r.Query)
				tc.assertBody(t, formBody, query)
			}
		})
	}
}

// Spot-check JSON decoding on a handful of the new service-scoped response
// models — the table-driven cases above mostly assert wire shape.
func TestConversationsV1ServiceDecodes(t *testing.T) {
	t.Run("ServiceConfiguration", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"chat_service_sid":                      convChatSvc,
			"default_chat_service_role_sid":         convRoleSid,
			"default_conversation_creator_role_sid": convRoleSid,
			"default_conversation_role_sid":         convRoleSid,
			"reachability_enabled":                  true,
			"url":                                   "https://example/v1/Services/" + convChatSvc + "/Configuration",
			"links":                                 map[string]string{"notifications": "https://example/v1/Services/" + convChatSvc + "/Configuration/Notifications"},
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		out, err := c.ConversationsV1.FetchServiceConfiguration(context.Background(), convChatSvc)
		if err != nil {
			t.Fatalf("FetchServiceConfiguration: %v", err)
		}
		if out.ChatServiceSid == nil || *out.ChatServiceSid != convChatSvc {
			t.Fatalf("chat_service_sid: %+v", out.ChatServiceSid)
		}
		if out.ReachabilityEnabled == nil || !*out.ReachabilityEnabled {
			t.Fatalf("reachability_enabled: %+v", out.ReachabilityEnabled)
		}
		if !strings.HasSuffix(out.Links["notifications"], "/Notifications") {
			t.Fatalf("links: %+v", out.Links)
		}
	})

	t.Run("ServiceBinding", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"sid": svcBindSid, "account_sid": testAccountSid, "chat_service_sid": convChatSvc,
			"credential_sid": convCredSid, "binding_type": "apn",
			"identity":      "alice",
			"endpoint":      "device-token",
			"message_types": []string{"new_message", "added_to_conversation"},
			"url":           "https://example/v1/Services/" + convChatSvc + "/Bindings/" + svcBindSid,
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		out, err := c.ConversationsV1.FetchServiceBinding(context.Background(), convChatSvc, svcBindSid)
		if err != nil {
			t.Fatalf("FetchServiceBinding: %v", err)
		}
		if out.BindingType != "apn" {
			t.Fatalf("binding_type: %q", out.BindingType)
		}
		if len(out.MessageTypes) != 2 || out.MessageTypes[0] != "new_message" {
			t.Fatalf("message_types: %+v", out.MessageTypes)
		}
	})

	t.Run("ServiceNotification", func(t *testing.T) {
		steps := []handlerStep{jsonStep(200, map[string]any{
			"account_sid":      testAccountSid,
			"chat_service_sid": convChatSvc,
			"log_enabled":      true,
			"new_message":      map[string]any{"enabled": true, "template": "${MESSAGE}"},
			"url":              "https://example/v1/Services/" + convChatSvc + "/Configuration/Notifications",
		})}
		c, _, done := newClient(t, steps, nil)
		defer done()
		out, err := c.ConversationsV1.FetchServiceNotification(context.Background(), convChatSvc)
		if err != nil {
			t.Fatalf("FetchServiceNotification: %v", err)
		}
		if out.LogEnabled == nil || !*out.LogEnabled {
			t.Fatalf("log_enabled: %+v", out.LogEnabled)
		}
		if got, ok := out.NewMessage["template"].(string); !ok || got != "${MESSAGE}" {
			t.Fatalf("new_message.template: %+v", out.NewMessage)
		}
	})
}
