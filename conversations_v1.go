// Conversations v1 (conversations.twilio.com/v1) resources: Conversation +
// nested Messages / Participants / Webhooks / Receipts, account-scoped Roles,
// Users + UserConversations, push Credentials, the Configuration singleton +
// its Webhooks + Addresses sub-resources, ParticipantConversation listings,
// ConversationWithParticipants composite create, and Services.
//
// The /v1 namespace omits /Accounts/{AccountSid}; account is resolved from
// HTTP Basic auth. List responses share the VoiceV1Meta envelope declared in
// voice_v1.go. Dotted parameter names (e.g. "Timers.Inactive") are sent
// verbatim — url.Values preserves them through encoding.

package voiceml

import (
	"context"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Response models.
// ---------------------------------------------------------------------------

// ConversationsV1Conversation is a stateful messaging thread (CH...). State
// tracks lifecycle (initializing → active → inactive → closed). Attributes
// is server-stringified JSON.
type ConversationsV1Conversation struct {
	AccountSid          *string           `json:"account_sid"`
	ChatServiceSid      *string           `json:"chat_service_sid"`
	MessagingServiceSid *string           `json:"messaging_service_sid"`
	Sid                 *string           `json:"sid"`
	FriendlyName        *string           `json:"friendly_name"`
	UniqueName          *string           `json:"unique_name"`
	Attributes          *string           `json:"attributes"`
	State               string            `json:"state"`
	DateCreated         *string           `json:"date_created"`
	DateUpdated         *string           `json:"date_updated"`
	Timers              map[string]string `json:"timers,omitempty"`
	URL                 *string           `json:"url"`
	Links               map[string]string `json:"links,omitempty"`
	Bindings            map[string]string `json:"bindings,omitempty"`
}

// ConversationsV1ConversationList is the paginated /v1/Conversations response.
type ConversationsV1ConversationList struct {
	Conversations []ConversationsV1Conversation `json:"conversations"`
	Meta          VoiceV1Meta                   `json:"meta"`
}

// ConversationsV1ConversationMessage is a single message inside a
// Conversation (IM...). Index is monotonically increasing per conversation.
type ConversationsV1ConversationMessage struct {
	AccountSid      *string             `json:"account_sid"`
	ConversationSid *string             `json:"conversation_sid"`
	Sid             *string             `json:"sid"`
	Index           int                 `json:"index"`
	Author          *string             `json:"author"`
	Body            *string             `json:"body"`
	Media           []map[string]string `json:"media,omitempty"`
	Attributes      *string             `json:"attributes"`
	ParticipantSid  *string             `json:"participant_sid"`
	DateCreated     *string             `json:"date_created"`
	DateUpdated     *string             `json:"date_updated"`
	URL             *string             `json:"url"`
	Delivery        map[string]string   `json:"delivery,omitempty"`
	Links           map[string]string   `json:"links,omitempty"`
	ContentSid      *string             `json:"content_sid"`
}

// ConversationsV1ConversationMessageList is the paginated nested-message response.
type ConversationsV1ConversationMessageList struct {
	Messages []ConversationsV1ConversationMessage `json:"messages"`
	Meta     VoiceV1Meta                          `json:"meta"`
}

// ConversationsV1ConversationParticipant is a chat or SMS participant in a
// Conversation (MB...). Identity is set for chat; MessagingBinding is set
// for SMS/WhatsApp.
type ConversationsV1ConversationParticipant struct {
	AccountSid           *string           `json:"account_sid"`
	ConversationSid      *string           `json:"conversation_sid"`
	Sid                  *string           `json:"sid"`
	Identity             *string           `json:"identity"`
	Attributes           *string           `json:"attributes"`
	MessagingBinding     map[string]string `json:"messaging_binding,omitempty"`
	RoleSid              *string           `json:"role_sid"`
	DateCreated          *string           `json:"date_created"`
	DateUpdated          *string           `json:"date_updated"`
	URL                  *string           `json:"url"`
	LastReadMessageIndex *int              `json:"last_read_message_index"`
	LastReadTimestamp    *string           `json:"last_read_timestamp"`
}

// ConversationsV1ConversationParticipantList is the paginated nested-participant response.
type ConversationsV1ConversationParticipantList struct {
	Participants []ConversationsV1ConversationParticipant `json:"participants"`
	Meta         VoiceV1Meta                              `json:"meta"`
}

// ConversationsV1ConversationMessageReceipt is a per-channel delivery
// receipt for a Message (DY...). Read-only.
type ConversationsV1ConversationMessageReceipt struct {
	AccountSid        *string `json:"account_sid"`
	ConversationSid   *string `json:"conversation_sid"`
	Sid               *string `json:"sid"`
	MessageSid        *string `json:"message_sid"`
	ChannelMessageSid *string `json:"channel_message_sid"`
	ParticipantSid    *string `json:"participant_sid"`
	Status            string  `json:"status"`
	ErrorCode         int     `json:"error_code"`
	DateCreated       *string `json:"date_created"`
	DateUpdated       *string `json:"date_updated"`
	URL               *string `json:"url"`
}

// ConversationsV1ConversationMessageReceiptList is the paginated
// delivery-receipt response. The wire key is `delivery_receipts`.
type ConversationsV1ConversationMessageReceiptList struct {
	DeliveryReceipts []ConversationsV1ConversationMessageReceipt `json:"delivery_receipts"`
	Meta             VoiceV1Meta                                 `json:"meta"`
}

// ConversationsV1ConversationScopedWebhook is a webhook attached to a
// specific Conversation (WH...). Target is "webhook", "trigger", or "studio".
type ConversationsV1ConversationScopedWebhook struct {
	Sid             *string           `json:"sid"`
	AccountSid      *string           `json:"account_sid"`
	ConversationSid *string           `json:"conversation_sid"`
	Target          *string           `json:"target"`
	URL             *string           `json:"url"`
	Configuration   map[string]string `json:"configuration,omitempty"`
	DateCreated     *string           `json:"date_created"`
	DateUpdated     *string           `json:"date_updated"`
}

// ConversationsV1ConversationScopedWebhookList is the paginated
// scoped-webhook response.
type ConversationsV1ConversationScopedWebhookList struct {
	Webhooks []ConversationsV1ConversationScopedWebhook `json:"webhooks"`
	Meta     VoiceV1Meta                                `json:"meta"`
}

// ConversationsV1Role is a Conversations-scoped permission bundle (RL...).
type ConversationsV1Role struct {
	Sid            *string  `json:"sid"`
	AccountSid     *string  `json:"account_sid"`
	ChatServiceSid *string  `json:"chat_service_sid"`
	FriendlyName   *string  `json:"friendly_name"`
	Type           string   `json:"type"`
	Permissions    []string `json:"permissions,omitempty"`
	DateCreated    *string  `json:"date_created"`
	DateUpdated    *string  `json:"date_updated"`
	URL            *string  `json:"url"`
}

// ConversationsV1RoleList is the paginated /v1/Roles response.
type ConversationsV1RoleList struct {
	Roles []ConversationsV1Role `json:"roles"`
	Meta  VoiceV1Meta           `json:"meta"`
}

// ConversationsV1User is an end-user identity that can participate in
// conversations (US...).
type ConversationsV1User struct {
	Sid            *string           `json:"sid"`
	AccountSid     *string           `json:"account_sid"`
	ChatServiceSid *string           `json:"chat_service_sid"`
	RoleSid        *string           `json:"role_sid"`
	Identity       *string           `json:"identity"`
	FriendlyName   *string           `json:"friendly_name"`
	Attributes     *string           `json:"attributes"`
	IsOnline       *bool             `json:"is_online"`
	IsNotifiable   *bool             `json:"is_notifiable"`
	DateCreated    *string           `json:"date_created"`
	DateUpdated    *string           `json:"date_updated"`
	URL            *string           `json:"url"`
	Links          map[string]string `json:"links,omitempty"`
}

// ConversationsV1UserList is the paginated /v1/Users response.
type ConversationsV1UserList struct {
	Users []ConversationsV1User `json:"users"`
	Meta  VoiceV1Meta           `json:"meta"`
}

// ConversationsV1Credential is a push-notification credential (CR...).
// Type is one of "apn" / "gcm" / "fcm".
type ConversationsV1Credential struct {
	Sid          *string `json:"sid"`
	AccountSid   *string `json:"account_sid"`
	FriendlyName *string `json:"friendly_name"`
	Type         string  `json:"type"`
	Sandbox      *string `json:"sandbox"`
	DateCreated  *string `json:"date_created"`
	DateUpdated  *string `json:"date_updated"`
	URL          *string `json:"url"`
}

// ConversationsV1CredentialList is the paginated /v1/Credentials response.
type ConversationsV1CredentialList struct {
	Credentials []ConversationsV1Credential `json:"credentials"`
	Meta        VoiceV1Meta                 `json:"meta"`
}

// ConversationsV1Configuration is the account-global Conversations
// Configuration singleton. Defaults when unset; fetch + update only.
type ConversationsV1Configuration struct {
	AccountSid                 *string           `json:"account_sid"`
	DefaultChatServiceSid      *string           `json:"default_chat_service_sid"`
	DefaultMessagingServiceSid *string           `json:"default_messaging_service_sid"`
	DefaultInactiveTimer       *string           `json:"default_inactive_timer"`
	DefaultClosedTimer         *string           `json:"default_closed_timer"`
	URL                        *string           `json:"url"`
	Links                      map[string]string `json:"links,omitempty"`
}

// ConversationsV1ConfigurationWebhook is the account-global Conversation
// webhook config singleton. Fetch + update only.
type ConversationsV1ConfigurationWebhook struct {
	AccountSid     *string  `json:"account_sid"`
	Method         string   `json:"method"`
	Filters        []string `json:"filters,omitempty"`
	PreWebhookURL  *string  `json:"pre_webhook_url"`
	PostWebhookURL *string  `json:"post_webhook_url"`
	Target         string   `json:"target"`
	URL            *string  `json:"url"`
}

// ConversationsV1ConfigAddress is an account-level inbound address (IG...).
// Type names a messaging channel ("sms", "whatsapp", ...).
type ConversationsV1ConfigAddress struct {
	Sid            *string           `json:"sid"`
	AccountSid     *string           `json:"account_sid"`
	Type           *string           `json:"type"`
	Address        *string           `json:"address"`
	FriendlyName   *string           `json:"friendly_name"`
	AutoCreation   map[string]string `json:"auto_creation,omitempty"`
	DateCreated    *string           `json:"date_created"`
	DateUpdated    *string           `json:"date_updated"`
	URL            *string           `json:"url"`
	AddressCountry *string           `json:"address_country"`
}

// ConversationsV1ConfigAddressList is the paginated /v1/Configuration/Addresses response.
type ConversationsV1ConfigAddressList struct {
	Addresses []ConversationsV1ConfigAddress `json:"addresses"`
	Meta      VoiceV1Meta                    `json:"meta"`
}

// ConversationsV1ParticipantConversation describes one participant's
// membership in a conversation — the GET /v1/ParticipantConversations
// inverse-index row. Read-only (no sid of its own).
type ConversationsV1ParticipantConversation struct {
	AccountSid                  *string           `json:"account_sid"`
	ChatServiceSid              *string           `json:"chat_service_sid"`
	ParticipantSid              *string           `json:"participant_sid"`
	ParticipantUserSid          *string           `json:"participant_user_sid"`
	ParticipantIdentity         *string           `json:"participant_identity"`
	ParticipantMessagingBinding map[string]string `json:"participant_messaging_binding,omitempty"`
	ConversationSid             *string           `json:"conversation_sid"`
	ConversationUniqueName      *string           `json:"conversation_unique_name"`
	ConversationFriendlyName    *string           `json:"conversation_friendly_name"`
	ConversationAttributes      *string           `json:"conversation_attributes"`
	ConversationDateCreated     *string           `json:"conversation_date_created"`
	ConversationDateUpdated     *string           `json:"conversation_date_updated"`
	ConversationCreatedBy       *string           `json:"conversation_created_by"`
	ConversationState           string            `json:"conversation_state"`
	ConversationTimers          map[string]string `json:"conversation_timers,omitempty"`
	Links                       map[string]string `json:"links,omitempty"`
}

// ConversationsV1ParticipantConversationList is the paginated response.
type ConversationsV1ParticipantConversationList struct {
	Conversations []ConversationsV1ParticipantConversation `json:"conversations"`
	Meta          VoiceV1Meta                              `json:"meta"`
}

// ConversationsV1ConversationWithParticipants is the response shape of the
// /v1/ConversationWithParticipants composite-create endpoint. Mirrors
// ConversationsV1Conversation; declared separately so the spec field-by-
// field parity remains clear.
type ConversationsV1ConversationWithParticipants struct {
	AccountSid          *string           `json:"account_sid"`
	ChatServiceSid      *string           `json:"chat_service_sid"`
	MessagingServiceSid *string           `json:"messaging_service_sid"`
	Sid                 *string           `json:"sid"`
	FriendlyName        *string           `json:"friendly_name"`
	UniqueName          *string           `json:"unique_name"`
	Attributes          *string           `json:"attributes"`
	State               string            `json:"state"`
	DateCreated         *string           `json:"date_created"`
	DateUpdated         *string           `json:"date_updated"`
	Timers              map[string]string `json:"timers,omitempty"`
	Links               map[string]string `json:"links,omitempty"`
	Bindings            map[string]string `json:"bindings,omitempty"`
	URL                 *string           `json:"url"`
}

// ConversationsV1UserConversation is the per-user view of a Conversation
// membership — the GET /v1/Users/{Sid}/Conversations row.
type ConversationsV1UserConversation struct {
	AccountSid           *string           `json:"account_sid"`
	ChatServiceSid       *string           `json:"chat_service_sid"`
	ConversationSid      *string           `json:"conversation_sid"`
	UnreadMessagesCount  *int              `json:"unread_messages_count"`
	LastReadMessageIndex *int              `json:"last_read_message_index"`
	ParticipantSid       *string           `json:"participant_sid"`
	UserSid              *string           `json:"user_sid"`
	FriendlyName         *string           `json:"friendly_name"`
	ConversationState    string            `json:"conversation_state"`
	Timers               map[string]string `json:"timers,omitempty"`
	Attributes           *string           `json:"attributes"`
	DateCreated          *string           `json:"date_created"`
	DateUpdated          *string           `json:"date_updated"`
	CreatedBy            *string           `json:"created_by"`
	NotificationLevel    string            `json:"notification_level"`
	UniqueName           *string           `json:"unique_name"`
	URL                  *string           `json:"url"`
	Links                map[string]string `json:"links,omitempty"`
}

// ConversationsV1UserConversationList is the paginated response for a user's conversations.
type ConversationsV1UserConversationList struct {
	Conversations []ConversationsV1UserConversation `json:"conversations"`
	Meta          VoiceV1Meta                       `json:"meta"`
}

// ConversationsV1ChatService is an isolated Conversations realm (IS...).
// Named to avoid colliding with the ConversationsV1Service methods container;
// Twilio internally refers to this resource as the chat service, and the
// path parameter is ChatServiceSid.
type ConversationsV1ChatService struct {
	Sid          *string           `json:"sid"`
	AccountSid   *string           `json:"account_sid"`
	FriendlyName *string           `json:"friendly_name"`
	DateCreated  *string           `json:"date_created"`
	DateUpdated  *string           `json:"date_updated"`
	URL          *string           `json:"url"`
	Links        map[string]string `json:"links,omitempty"`
}

// ConversationsV1ChatServiceList is the paginated /v1/Services response.
type ConversationsV1ChatServiceList struct {
	Services []ConversationsV1ChatService `json:"services"`
	Meta     VoiceV1Meta                  `json:"meta"`
}

// ---------------------------------------------------------------------------
// Request params.
// ---------------------------------------------------------------------------

// addStrings appends each value under the same key (repeated form params).
// url.Values.Encode renders this as "Key=v1&Key=v2&...".
func addStrings(v url.Values, key string, vs []string) {
	for _, s := range vs {
		v.Add(key, s)
	}
}

// CreateConversationRequest is the body for POST /v1/Conversations. All
// fields optional; dotted keys (Timers.*, Bindings.Email.*) are emitted verbatim.
type CreateConversationRequest struct {
	FriendlyName        *string `form:"FriendlyName"`
	UniqueName          *string `form:"UniqueName"`
	MessagingServiceSid *string `form:"MessagingServiceSid"`
	Attributes          *string `form:"Attributes"`
	State               *string `form:"State"`
	TimersInactive      *string `form:"Timers.Inactive"`
	TimersClosed        *string `form:"Timers.Closed"`
	BindingsEmailAddr   *string `form:"Bindings.Email.Address"`
	BindingsEmailName   *string `form:"Bindings.Email.Name"`
}

func (p CreateConversationRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "UniqueName", p.UniqueName)
	setStr(v, "MessagingServiceSid", p.MessagingServiceSid)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "State", p.State)
	setStr(v, "Timers.Inactive", p.TimersInactive)
	setStr(v, "Timers.Closed", p.TimersClosed)
	setStr(v, "Bindings.Email.Address", p.BindingsEmailAddr)
	setStr(v, "Bindings.Email.Name", p.BindingsEmailName)
	return v
}

// UpdateConversationRequest is the body for POST /v1/Conversations/{ConversationSid}.
type UpdateConversationRequest struct {
	FriendlyName        *string `form:"FriendlyName"`
	UniqueName          *string `form:"UniqueName"`
	MessagingServiceSid *string `form:"MessagingServiceSid"`
	Attributes          *string `form:"Attributes"`
	State               *string `form:"State"`
	TimersInactive      *string `form:"Timers.Inactive"`
	TimersClosed        *string `form:"Timers.Closed"`
}

func (p UpdateConversationRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "UniqueName", p.UniqueName)
	setStr(v, "MessagingServiceSid", p.MessagingServiceSid)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "State", p.State)
	setStr(v, "Timers.Inactive", p.TimersInactive)
	setStr(v, "Timers.Closed", p.TimersClosed)
	return v
}

// ListConversationsParams is the query for GET /v1/Conversations.
type ListConversationsParams = V1PageParams

// CreateMessageRequest is the body for POST /v1/Conversations/{ConversationSid}/Messages.
type CreateMessageRequest struct {
	Author     *string `form:"Author"`
	Body       *string `form:"Body"`
	Attributes *string `form:"Attributes"`
	ContentSid *string `form:"ContentSid"`
}

func (p CreateMessageRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Author", p.Author)
	setStr(v, "Body", p.Body)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "ContentSid", p.ContentSid)
	return v
}

// UpdateMessageRequest is the body for POST /v1/Conversations/{ConversationSid}/Messages/{MessageSid}.
type UpdateMessageRequest struct {
	Author     *string `form:"Author"`
	Body       *string `form:"Body"`
	Attributes *string `form:"Attributes"`
}

func (p UpdateMessageRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Author", p.Author)
	setStr(v, "Body", p.Body)
	setStr(v, "Attributes", p.Attributes)
	return v
}

// CreateParticipantRequest is the body for POST /v1/Conversations/{ConversationSid}/Participants.
type CreateParticipantRequest struct {
	Identity                         *string `form:"Identity"`
	Attributes                       *string `form:"Attributes"`
	RoleSid                          *string `form:"RoleSid"`
	MessagingBindingAddress          *string `form:"MessagingBinding.Address"`
	MessagingBindingProxyAddress     *string `form:"MessagingBinding.ProxyAddress"`
	MessagingBindingProjectedAddress *string `form:"MessagingBinding.ProjectedAddress"`
}

func (p CreateParticipantRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Identity", p.Identity)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "RoleSid", p.RoleSid)
	setStr(v, "MessagingBinding.Address", p.MessagingBindingAddress)
	setStr(v, "MessagingBinding.ProxyAddress", p.MessagingBindingProxyAddress)
	setStr(v, "MessagingBinding.ProjectedAddress", p.MessagingBindingProjectedAddress)
	return v
}

// UpdateParticipantRequest is the body for POST /v1/Conversations/{ConversationSid}/Participants/{ParticipantSid}.
type UpdateParticipantRequest struct {
	Identity             *string `form:"Identity"`
	Attributes           *string `form:"Attributes"`
	RoleSid              *string `form:"RoleSid"`
	LastReadMessageIndex *int    `form:"LastReadMessageIndex"`
	LastReadTimestamp    *string `form:"LastReadTimestamp"`
}

func (p UpdateParticipantRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Identity", p.Identity)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "RoleSid", p.RoleSid)
	setInt(v, "LastReadMessageIndex", p.LastReadMessageIndex)
	setStr(v, "LastReadTimestamp", p.LastReadTimestamp)
	return v
}

// CreateScopedWebhookRequest is the body for POST
// /v1/Conversations/{ConversationSid}/Webhooks. Target is required.
type CreateScopedWebhookRequest struct {
	Target                   string  `form:"Target"`
	ConfigurationURL         *string `form:"Configuration.Url"`
	ConfigurationMethod      *string `form:"Configuration.Method"`
	ConfigurationFlowSid     *string `form:"Configuration.FlowSid"`
	ConfigurationReplayAfter *int    `form:"Configuration.ReplayAfter"`
}

func (p CreateScopedWebhookRequest) form() url.Values {
	v := url.Values{}
	v.Set("Target", p.Target)
	setStr(v, "Configuration.Url", p.ConfigurationURL)
	setStr(v, "Configuration.Method", p.ConfigurationMethod)
	setStr(v, "Configuration.FlowSid", p.ConfigurationFlowSid)
	setInt(v, "Configuration.ReplayAfter", p.ConfigurationReplayAfter)
	return v
}

// UpdateScopedWebhookRequest is the body for POST
// /v1/Conversations/{ConversationSid}/Webhooks/{WebhookSid}.
type UpdateScopedWebhookRequest struct {
	ConfigurationURL     *string `form:"Configuration.Url"`
	ConfigurationMethod  *string `form:"Configuration.Method"`
	ConfigurationFlowSid *string `form:"Configuration.FlowSid"`
}

func (p UpdateScopedWebhookRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Configuration.Url", p.ConfigurationURL)
	setStr(v, "Configuration.Method", p.ConfigurationMethod)
	setStr(v, "Configuration.FlowSid", p.ConfigurationFlowSid)
	return v
}

// CreateRoleRequest is the body for POST /v1/Roles. All three fields required.
// Permission is repeated form-encoded.
type CreateRoleRequest struct {
	FriendlyName string   `form:"FriendlyName"`
	Type         string   `form:"Type"`
	Permission   []string `form:"Permission"`
}

func (p CreateRoleRequest) form() url.Values {
	v := url.Values{}
	v.Set("FriendlyName", p.FriendlyName)
	v.Set("Type", p.Type)
	addStrings(v, "Permission", p.Permission)
	return v
}

// UpdateRoleRequest is the body for POST /v1/Roles/{Sid}.
type UpdateRoleRequest struct {
	Permission []string `form:"Permission"`
}

func (p UpdateRoleRequest) form() url.Values {
	v := url.Values{}
	addStrings(v, "Permission", p.Permission)
	return v
}

// CreateUserRequest is the body for POST /v1/Users. Identity is required.
type CreateUserRequest struct {
	Identity     string  `form:"Identity"`
	FriendlyName *string `form:"FriendlyName"`
	Attributes   *string `form:"Attributes"`
	RoleSid      *string `form:"RoleSid"`
}

func (p CreateUserRequest) form() url.Values {
	v := url.Values{}
	v.Set("Identity", p.Identity)
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "RoleSid", p.RoleSid)
	return v
}

// UpdateUserRequest is the body for POST /v1/Users/{Sid}.
type UpdateUserRequest struct {
	FriendlyName *string `form:"FriendlyName"`
	Attributes   *string `form:"Attributes"`
	RoleSid      *string `form:"RoleSid"`
}

func (p UpdateUserRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "RoleSid", p.RoleSid)
	return v
}

// UpdateUserConversationRequest is the body for POST
// /v1/Users/{Sid}/Conversations/{ConversationSid}.
type UpdateUserConversationRequest struct {
	NotificationLevel    *string `form:"NotificationLevel"`
	LastReadMessageIndex *int    `form:"LastReadMessageIndex"`
	LastReadTimestamp    *string `form:"LastReadTimestamp"`
}

func (p UpdateUserConversationRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "NotificationLevel", p.NotificationLevel)
	setInt(v, "LastReadMessageIndex", p.LastReadMessageIndex)
	setStr(v, "LastReadTimestamp", p.LastReadTimestamp)
	return v
}

// CreateCredentialRequest is the body for POST /v1/Credentials. Type is required.
type CreateCredentialRequest struct {
	Type         string  `form:"Type"`
	FriendlyName *string `form:"FriendlyName"`
	Certificate  *string `form:"Certificate"`
	PrivateKey   *string `form:"PrivateKey"`
	Sandbox      *bool   `form:"Sandbox"`
	APIKey       *string `form:"ApiKey"`
	Secret       *string `form:"Secret"`
}

func (p CreateCredentialRequest) form() url.Values {
	v := url.Values{}
	v.Set("Type", p.Type)
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "Certificate", p.Certificate)
	setStr(v, "PrivateKey", p.PrivateKey)
	setBool(v, "Sandbox", p.Sandbox)
	setStr(v, "ApiKey", p.APIKey)
	setStr(v, "Secret", p.Secret)
	return v
}

// UpdateCredentialRequest is the body for POST /v1/Credentials/{Sid}.
type UpdateCredentialRequest struct {
	Type         *string `form:"Type"`
	FriendlyName *string `form:"FriendlyName"`
	Certificate  *string `form:"Certificate"`
	PrivateKey   *string `form:"PrivateKey"`
	Sandbox      *bool   `form:"Sandbox"`
	APIKey       *string `form:"ApiKey"`
	Secret       *string `form:"Secret"`
}

func (p UpdateCredentialRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Type", p.Type)
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "Certificate", p.Certificate)
	setStr(v, "PrivateKey", p.PrivateKey)
	setBool(v, "Sandbox", p.Sandbox)
	setStr(v, "ApiKey", p.APIKey)
	setStr(v, "Secret", p.Secret)
	return v
}

// UpdateConfigurationRequest is the body for POST /v1/Configuration.
type UpdateConfigurationRequest struct {
	DefaultChatServiceSid      *string `form:"DefaultChatServiceSid"`
	DefaultMessagingServiceSid *string `form:"DefaultMessagingServiceSid"`
	DefaultInactiveTimer       *string `form:"DefaultInactiveTimer"`
	DefaultClosedTimer         *string `form:"DefaultClosedTimer"`
}

func (p UpdateConfigurationRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "DefaultChatServiceSid", p.DefaultChatServiceSid)
	setStr(v, "DefaultMessagingServiceSid", p.DefaultMessagingServiceSid)
	setStr(v, "DefaultInactiveTimer", p.DefaultInactiveTimer)
	setStr(v, "DefaultClosedTimer", p.DefaultClosedTimer)
	return v
}

// UpdateConfigurationWebhookRequest is the body for POST /v1/Configuration/Webhooks.
// Filters is repeated form-encoded.
type UpdateConfigurationWebhookRequest struct {
	Method         *string  `form:"Method"`
	Filters        []string `form:"Filters"`
	PreWebhookURL  *string  `form:"PreWebhookUrl"`
	PostWebhookURL *string  `form:"PostWebhookUrl"`
	Target         *string  `form:"Target"`
}

func (p UpdateConfigurationWebhookRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Method", p.Method)
	addStrings(v, "Filters", p.Filters)
	setStr(v, "PreWebhookUrl", p.PreWebhookURL)
	setStr(v, "PostWebhookUrl", p.PostWebhookURL)
	setStr(v, "Target", p.Target)
	return v
}

// CreateConfigAddressRequest is the body for POST /v1/Configuration/Addresses.
// Type and Address are required.
type CreateConfigAddressRequest struct {
	Type                   string  `form:"Type"`
	Address                string  `form:"Address"`
	FriendlyName           *string `form:"FriendlyName"`
	AutoCreationEnabled    *bool   `form:"AutoCreation.Enabled"`
	AutoCreationType       *string `form:"AutoCreation.Type"`
	AutoCreationWebhookURL *string `form:"AutoCreation.WebhookUrl"`
	AddressCountry         *string `form:"AddressCountry"`
}

func (p CreateConfigAddressRequest) form() url.Values {
	v := url.Values{}
	v.Set("Type", p.Type)
	v.Set("Address", p.Address)
	setStr(v, "FriendlyName", p.FriendlyName)
	setBool(v, "AutoCreation.Enabled", p.AutoCreationEnabled)
	setStr(v, "AutoCreation.Type", p.AutoCreationType)
	setStr(v, "AutoCreation.WebhookUrl", p.AutoCreationWebhookURL)
	setStr(v, "AddressCountry", p.AddressCountry)
	return v
}

// UpdateConfigAddressRequest is the body for POST /v1/Configuration/Addresses/{Sid}.
type UpdateConfigAddressRequest struct {
	FriendlyName           *string `form:"FriendlyName"`
	AutoCreationEnabled    *bool   `form:"AutoCreation.Enabled"`
	AutoCreationType       *string `form:"AutoCreation.Type"`
	AutoCreationWebhookURL *string `form:"AutoCreation.WebhookUrl"`
}

func (p UpdateConfigAddressRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setBool(v, "AutoCreation.Enabled", p.AutoCreationEnabled)
	setStr(v, "AutoCreation.Type", p.AutoCreationType)
	setStr(v, "AutoCreation.WebhookUrl", p.AutoCreationWebhookURL)
	return v
}

// ListParticipantConversationsParams is the query for GET /v1/ParticipantConversations.
// Either Identity or Address narrows the result; neither lists all
// participant-conversation rows for the account.
type ListParticipantConversationsParams struct {
	Identity *string
	Address  *string
	PageSize *int
}

func (p ListParticipantConversationsParams) query() url.Values {
	v := url.Values{}
	if p.Identity != nil {
		v.Set("Identity", *p.Identity)
	}
	if p.Address != nil {
		v.Set("Address", *p.Address)
	}
	if p.PageSize != nil {
		v.Set("PageSize", strconv.Itoa(*p.PageSize))
	}
	return v
}

// CreateConversationWithParticipantsRequest is the body for POST
// /v1/ConversationWithParticipants. Participant is repeated; each value is
// a JSON-encoded participant spec.
type CreateConversationWithParticipantsRequest struct {
	FriendlyName        *string  `form:"FriendlyName"`
	UniqueName          *string  `form:"UniqueName"`
	MessagingServiceSid *string  `form:"MessagingServiceSid"`
	Attributes          *string  `form:"Attributes"`
	State               *string  `form:"State"`
	TimersInactive      *string  `form:"Timers.Inactive"`
	TimersClosed        *string  `form:"Timers.Closed"`
	Participant         []string `form:"Participant"`
}

func (p CreateConversationWithParticipantsRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "UniqueName", p.UniqueName)
	setStr(v, "MessagingServiceSid", p.MessagingServiceSid)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "State", p.State)
	setStr(v, "Timers.Inactive", p.TimersInactive)
	setStr(v, "Timers.Closed", p.TimersClosed)
	addStrings(v, "Participant", p.Participant)
	return v
}

// CreateServiceRequest is the body for POST /v1/Services. FriendlyName required.
type CreateServiceRequest struct {
	FriendlyName string `form:"FriendlyName"`
}

func (p CreateServiceRequest) form() url.Values {
	v := url.Values{}
	v.Set("FriendlyName", p.FriendlyName)
	return v
}

// ---------------------------------------------------------------------------
// Service — flat-method facade for the entire Conversations v1 surface.
// ---------------------------------------------------------------------------

// ConversationsV1Service exposes the conversations.twilio.com/v1 endpoints.
// Reach it as c.ConversationsV1. Methods are named verb+resource so the
// surface stays flat for IDE discovery.
type ConversationsV1Service struct{ c *Client }

// --- Conversation ----------------------------------------------------------

// CreateConversation adds a Conversation.
func (s *ConversationsV1Service) CreateConversation(ctx context.Context, params CreateConversationRequest) (*ConversationsV1Conversation, error) {
	var out ConversationsV1Conversation
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Conversations", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListConversations returns a single page of Conversations.
func (s *ConversationsV1Service) ListConversations(ctx context.Context, params ListConversationsParams) (*ConversationsV1ConversationList, error) {
	var out ConversationsV1ConversationList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchConversation retrieves a Conversation by sid.
func (s *ConversationsV1Service) FetchConversation(ctx context.Context, conversationSid string) (*ConversationsV1Conversation, error) {
	var out ConversationsV1Conversation
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations/" + conversationSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateConversation mutates a Conversation in place.
func (s *ConversationsV1Service) UpdateConversation(ctx context.Context, conversationSid string, params UpdateConversationRequest) (*ConversationsV1Conversation, error) {
	var out ConversationsV1Conversation
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Conversations/" + conversationSid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteConversation removes a Conversation.
func (s *ConversationsV1Service) DeleteConversation(ctx context.Context, conversationSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Conversations/" + conversationSid,
	}, nil)
}

// --- ConversationMessage ---------------------------------------------------

// CreateMessage adds a Message to a Conversation.
func (s *ConversationsV1Service) CreateMessage(ctx context.Context, conversationSid string, params CreateMessageRequest) (*ConversationsV1ConversationMessage, error) {
	var out ConversationsV1ConversationMessage
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Conversations/" + conversationSid + "/Messages", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListMessages returns a single page of a Conversation's Messages.
func (s *ConversationsV1Service) ListMessages(ctx context.Context, conversationSid string, params V1PageParams) (*ConversationsV1ConversationMessageList, error) {
	var out ConversationsV1ConversationMessageList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations/" + conversationSid + "/Messages", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchMessage retrieves a Message by sid.
func (s *ConversationsV1Service) FetchMessage(ctx context.Context, conversationSid, messageSid string) (*ConversationsV1ConversationMessage, error) {
	var out ConversationsV1ConversationMessage
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations/" + conversationSid + "/Messages/" + messageSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMessage mutates a Message in place.
func (s *ConversationsV1Service) UpdateMessage(ctx context.Context, conversationSid, messageSid string, params UpdateMessageRequest) (*ConversationsV1ConversationMessage, error) {
	var out ConversationsV1ConversationMessage
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Conversations/" + conversationSid + "/Messages/" + messageSid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMessage removes a Message.
func (s *ConversationsV1Service) DeleteMessage(ctx context.Context, conversationSid, messageSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Conversations/" + conversationSid + "/Messages/" + messageSid,
	}, nil)
}

// --- Conversation Participants ---------------------------------------------

// CreateParticipant adds a Participant to a Conversation.
func (s *ConversationsV1Service) CreateParticipant(ctx context.Context, conversationSid string, params CreateParticipantRequest) (*ConversationsV1ConversationParticipant, error) {
	var out ConversationsV1ConversationParticipant
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Conversations/" + conversationSid + "/Participants", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListParticipants returns a single page of a Conversation's Participants.
func (s *ConversationsV1Service) ListParticipants(ctx context.Context, conversationSid string, params V1PageParams) (*ConversationsV1ConversationParticipantList, error) {
	var out ConversationsV1ConversationParticipantList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations/" + conversationSid + "/Participants", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchParticipant retrieves a Participant by sid.
func (s *ConversationsV1Service) FetchParticipant(ctx context.Context, conversationSid, participantSid string) (*ConversationsV1ConversationParticipant, error) {
	var out ConversationsV1ConversationParticipant
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations/" + conversationSid + "/Participants/" + participantSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateParticipant mutates a Participant in place.
func (s *ConversationsV1Service) UpdateParticipant(ctx context.Context, conversationSid, participantSid string, params UpdateParticipantRequest) (*ConversationsV1ConversationParticipant, error) {
	var out ConversationsV1ConversationParticipant
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Conversations/" + conversationSid + "/Participants/" + participantSid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteParticipant removes a Participant.
func (s *ConversationsV1Service) DeleteParticipant(ctx context.Context, conversationSid, participantSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Conversations/" + conversationSid + "/Participants/" + participantSid,
	}, nil)
}

// --- Message Receipts ------------------------------------------------------

// ListMessageReceipts returns a single page of a Message's delivery receipts.
func (s *ConversationsV1Service) ListMessageReceipts(ctx context.Context, conversationSid, messageSid string, params V1PageParams) (*ConversationsV1ConversationMessageReceiptList, error) {
	var out ConversationsV1ConversationMessageReceiptList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations/" + conversationSid + "/Messages/" + messageSid + "/Receipts", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchMessageReceipt retrieves a single delivery receipt by sid.
func (s *ConversationsV1Service) FetchMessageReceipt(ctx context.Context, conversationSid, messageSid, sid string) (*ConversationsV1ConversationMessageReceipt, error) {
	var out ConversationsV1ConversationMessageReceipt
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations/" + conversationSid + "/Messages/" + messageSid + "/Receipts/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Conversation Scoped Webhooks ------------------------------------------

// CreateScopedWebhook adds a webhook scoped to a Conversation.
func (s *ConversationsV1Service) CreateScopedWebhook(ctx context.Context, conversationSid string, params CreateScopedWebhookRequest) (*ConversationsV1ConversationScopedWebhook, error) {
	var out ConversationsV1ConversationScopedWebhook
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Conversations/" + conversationSid + "/Webhooks", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListScopedWebhooks returns a single page of a Conversation's scoped webhooks.
func (s *ConversationsV1Service) ListScopedWebhooks(ctx context.Context, conversationSid string, params V1PageParams) (*ConversationsV1ConversationScopedWebhookList, error) {
	var out ConversationsV1ConversationScopedWebhookList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations/" + conversationSid + "/Webhooks", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchScopedWebhook retrieves a scoped webhook by sid.
func (s *ConversationsV1Service) FetchScopedWebhook(ctx context.Context, conversationSid, webhookSid string) (*ConversationsV1ConversationScopedWebhook, error) {
	var out ConversationsV1ConversationScopedWebhook
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Conversations/" + conversationSid + "/Webhooks/" + webhookSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateScopedWebhook mutates a scoped webhook in place.
func (s *ConversationsV1Service) UpdateScopedWebhook(ctx context.Context, conversationSid, webhookSid string, params UpdateScopedWebhookRequest) (*ConversationsV1ConversationScopedWebhook, error) {
	var out ConversationsV1ConversationScopedWebhook
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Conversations/" + conversationSid + "/Webhooks/" + webhookSid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteScopedWebhook removes a scoped webhook.
func (s *ConversationsV1Service) DeleteScopedWebhook(ctx context.Context, conversationSid, webhookSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Conversations/" + conversationSid + "/Webhooks/" + webhookSid,
	}, nil)
}

// --- Roles -----------------------------------------------------------------

// CreateRole adds a Role.
func (s *ConversationsV1Service) CreateRole(ctx context.Context, params CreateRoleRequest) (*ConversationsV1Role, error) {
	var out ConversationsV1Role
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Roles", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRoles returns a single page of Roles.
func (s *ConversationsV1Service) ListRoles(ctx context.Context, params V1PageParams) (*ConversationsV1RoleList, error) {
	var out ConversationsV1RoleList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Roles", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchRole retrieves a Role by sid.
func (s *ConversationsV1Service) FetchRole(ctx context.Context, sid string) (*ConversationsV1Role, error) {
	var out ConversationsV1Role
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Roles/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateRole replaces a Role's permission list.
func (s *ConversationsV1Service) UpdateRole(ctx context.Context, sid string, params UpdateRoleRequest) (*ConversationsV1Role, error) {
	var out ConversationsV1Role
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Roles/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteRole removes a Role.
func (s *ConversationsV1Service) DeleteRole(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Roles/" + sid,
	}, nil)
}

// --- Users -----------------------------------------------------------------

// CreateUser adds a User.
func (s *ConversationsV1Service) CreateUser(ctx context.Context, params CreateUserRequest) (*ConversationsV1User, error) {
	var out ConversationsV1User
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Users", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListUsers returns a single page of Users.
func (s *ConversationsV1Service) ListUsers(ctx context.Context, params V1PageParams) (*ConversationsV1UserList, error) {
	var out ConversationsV1UserList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Users", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchUser retrieves a User by sid.
func (s *ConversationsV1Service) FetchUser(ctx context.Context, sid string) (*ConversationsV1User, error) {
	var out ConversationsV1User
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Users/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateUser mutates a User in place.
func (s *ConversationsV1Service) UpdateUser(ctx context.Context, sid string, params UpdateUserRequest) (*ConversationsV1User, error) {
	var out ConversationsV1User
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Users/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteUser removes a User.
func (s *ConversationsV1Service) DeleteUser(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Users/" + sid,
	}, nil)
}

// --- User's Conversations --------------------------------------------------

// ListUserConversations returns a single page of a User's conversations.
func (s *ConversationsV1Service) ListUserConversations(ctx context.Context, userSid string, params V1PageParams) (*ConversationsV1UserConversationList, error) {
	var out ConversationsV1UserConversationList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Users/" + userSid + "/Conversations", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchUserConversation retrieves a single User-Conversation membership.
func (s *ConversationsV1Service) FetchUserConversation(ctx context.Context, userSid, conversationSid string) (*ConversationsV1UserConversation, error) {
	var out ConversationsV1UserConversation
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Users/" + userSid + "/Conversations/" + conversationSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateUserConversation mutates per-user state (notification level, read marker).
func (s *ConversationsV1Service) UpdateUserConversation(ctx context.Context, userSid, conversationSid string, params UpdateUserConversationRequest) (*ConversationsV1UserConversation, error) {
	var out ConversationsV1UserConversation
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Users/" + userSid + "/Conversations/" + conversationSid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteUserConversation removes a User from a Conversation.
func (s *ConversationsV1Service) DeleteUserConversation(ctx context.Context, userSid, conversationSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Users/" + userSid + "/Conversations/" + conversationSid,
	}, nil)
}

// --- Push Credentials ------------------------------------------------------

// CreateCredential adds a push Credential.
func (s *ConversationsV1Service) CreateCredential(ctx context.Context, params CreateCredentialRequest) (*ConversationsV1Credential, error) {
	var out ConversationsV1Credential
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Credentials", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListCredentials returns a single page of push Credentials.
func (s *ConversationsV1Service) ListCredentials(ctx context.Context, params V1PageParams) (*ConversationsV1CredentialList, error) {
	var out ConversationsV1CredentialList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Credentials", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchCredential retrieves a push Credential by sid.
func (s *ConversationsV1Service) FetchCredential(ctx context.Context, sid string) (*ConversationsV1Credential, error) {
	var out ConversationsV1Credential
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Credentials/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateCredential mutates a push Credential in place.
func (s *ConversationsV1Service) UpdateCredential(ctx context.Context, sid string, params UpdateCredentialRequest) (*ConversationsV1Credential, error) {
	var out ConversationsV1Credential
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Credentials/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCredential removes a push Credential.
func (s *ConversationsV1Service) DeleteCredential(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Credentials/" + sid,
	}, nil)
}

// --- Configuration (singleton) ---------------------------------------------

// FetchConfiguration retrieves the account-global Conversations Configuration.
func (s *ConversationsV1Service) FetchConfiguration(ctx context.Context) (*ConversationsV1Configuration, error) {
	var out ConversationsV1Configuration
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Configuration",
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateConfiguration mutates the account-global Conversations Configuration.
func (s *ConversationsV1Service) UpdateConfiguration(ctx context.Context, params UpdateConfigurationRequest) (*ConversationsV1Configuration, error) {
	var out ConversationsV1Configuration
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Configuration", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchConfigurationWebhook retrieves the account-global webhook config.
func (s *ConversationsV1Service) FetchConfigurationWebhook(ctx context.Context) (*ConversationsV1ConfigurationWebhook, error) {
	var out ConversationsV1ConfigurationWebhook
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Configuration/Webhooks",
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateConfigurationWebhook mutates the account-global webhook config.
func (s *ConversationsV1Service) UpdateConfigurationWebhook(ctx context.Context, params UpdateConfigurationWebhookRequest) (*ConversationsV1ConfigurationWebhook, error) {
	var out ConversationsV1ConfigurationWebhook
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Configuration/Webhooks", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Configuration Addresses ----------------------------------------------

// CreateConfigAddress adds a Configuration Address.
func (s *ConversationsV1Service) CreateConfigAddress(ctx context.Context, params CreateConfigAddressRequest) (*ConversationsV1ConfigAddress, error) {
	var out ConversationsV1ConfigAddress
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Configuration/Addresses", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListConfigAddresses returns a single page of Configuration Addresses.
func (s *ConversationsV1Service) ListConfigAddresses(ctx context.Context, params V1PageParams) (*ConversationsV1ConfigAddressList, error) {
	var out ConversationsV1ConfigAddressList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Configuration/Addresses", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchConfigAddress retrieves a Configuration Address by sid.
func (s *ConversationsV1Service) FetchConfigAddress(ctx context.Context, sid string) (*ConversationsV1ConfigAddress, error) {
	var out ConversationsV1ConfigAddress
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Configuration/Addresses/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateConfigAddress mutates a Configuration Address in place.
func (s *ConversationsV1Service) UpdateConfigAddress(ctx context.Context, sid string, params UpdateConfigAddressRequest) (*ConversationsV1ConfigAddress, error) {
	var out ConversationsV1ConfigAddress
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Configuration/Addresses/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteConfigAddress removes a Configuration Address.
func (s *ConversationsV1Service) DeleteConfigAddress(ctx context.Context, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Configuration/Addresses/" + sid,
	}, nil)
}

// --- ParticipantConversations (inverse index, list-only) -------------------

// ListParticipantConversations returns the conversations a participant
// belongs to, optionally filtered by Identity or messaging-binding Address.
func (s *ConversationsV1Service) ListParticipantConversations(ctx context.Context, params ListParticipantConversationsParams) (*ConversationsV1ParticipantConversationList, error) {
	var out ConversationsV1ParticipantConversationList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/ParticipantConversations", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- ConversationWithParticipants (composite create) -----------------------

// CreateConversationWithParticipants adds a Conversation and its initial Participants in one call.
func (s *ConversationsV1Service) CreateConversationWithParticipants(ctx context.Context, params CreateConversationWithParticipantsRequest) (*ConversationsV1ConversationWithParticipants, error) {
	var out ConversationsV1ConversationWithParticipants
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/ConversationWithParticipants", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Services --------------------------------------------------------------

// CreateService adds a Conversations Service.
func (s *ConversationsV1Service) CreateService(ctx context.Context, params CreateServiceRequest) (*ConversationsV1ChatService, error) {
	var out ConversationsV1ChatService
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListServices returns a single page of Conversations Services.
func (s *ConversationsV1Service) ListServices(ctx context.Context, params V1PageParams) (*ConversationsV1ChatServiceList, error) {
	var out ConversationsV1ChatServiceList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchService retrieves a Conversations Service by sid.
func (s *ConversationsV1Service) FetchService(ctx context.Context, chatServiceSid string) (*ConversationsV1ChatService, error) {
	var out ConversationsV1ChatService
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteService removes a Conversations Service.
func (s *ConversationsV1Service) DeleteService(ctx context.Context, chatServiceSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Services/" + chatServiceSid,
	}, nil)
}
