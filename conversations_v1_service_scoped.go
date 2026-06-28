// Service-scoped Conversations v1 — every resource living under
// /v1/Services/{ChatServiceSid}/. Twilio mirrors the account-scoped surface
// inside a Chat Service realm, so most shapes are field-identical to their
// account-level counterparts (declared in conversations_v1.go) with an extra
// chat_service_sid field. The two configuration singletons (Notifications
// and WebhookConfiguration) are service-only — there is no account-scoped
// twin.
//
// Methods slot onto the existing *ConversationsV1Service receiver so the
// flat-namespace facade stays intact: c.ConversationsV1.CreateServiceX(ctx,
// chatServiceSid, ...).

package voiceml

import (
	"context"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Response models.
// ---------------------------------------------------------------------------

// ConversationsV1ServiceConversation is a Conversation living inside a Chat
// Service realm. Field-identical to ConversationsV1Conversation; declared
// separately so spec drift surfaces as a compile error.
type ConversationsV1ServiceConversation struct {
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

// ConversationsV1ServiceConversationList is the paginated response.
type ConversationsV1ServiceConversationList struct {
	Conversations []ConversationsV1ServiceConversation `json:"conversations"`
	Meta          VoiceV1Meta                          `json:"meta"`
}

// ConversationsV1ServiceConversationMessage mirrors the account-scoped
// message body with an added chat_service_sid.
type ConversationsV1ServiceConversationMessage struct {
	AccountSid      *string             `json:"account_sid"`
	ChatServiceSid  *string             `json:"chat_service_sid"`
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

// ConversationsV1ServiceConversationMessageList is the paginated response.
type ConversationsV1ServiceConversationMessageList struct {
	Messages []ConversationsV1ServiceConversationMessage `json:"messages"`
	Meta     VoiceV1Meta                                 `json:"meta"`
}

// ConversationsV1ServiceConversationParticipant mirrors the account-scoped
// participant with an added chat_service_sid.
type ConversationsV1ServiceConversationParticipant struct {
	AccountSid           *string           `json:"account_sid"`
	ChatServiceSid       *string           `json:"chat_service_sid"`
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

// ConversationsV1ServiceConversationParticipantList is the paginated response.
type ConversationsV1ServiceConversationParticipantList struct {
	Participants []ConversationsV1ServiceConversationParticipant `json:"participants"`
	Meta         VoiceV1Meta                                     `json:"meta"`
}

// ConversationsV1ServiceConversationMessageReceipt mirrors the account-scoped
// delivery receipt with an added chat_service_sid.
type ConversationsV1ServiceConversationMessageReceipt struct {
	AccountSid        *string `json:"account_sid"`
	ChatServiceSid    *string `json:"chat_service_sid"`
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

// ConversationsV1ServiceConversationMessageReceiptList is the paginated
// response. The wire key is `delivery_receipts`.
type ConversationsV1ServiceConversationMessageReceiptList struct {
	DeliveryReceipts []ConversationsV1ServiceConversationMessageReceipt `json:"delivery_receipts"`
	Meta             VoiceV1Meta                                        `json:"meta"`
}

// ConversationsV1ServiceConversationScopedWebhook mirrors the account-scoped
// scoped webhook with an added chat_service_sid.
type ConversationsV1ServiceConversationScopedWebhook struct {
	Sid             *string           `json:"sid"`
	AccountSid      *string           `json:"account_sid"`
	ChatServiceSid  *string           `json:"chat_service_sid"`
	ConversationSid *string           `json:"conversation_sid"`
	Target          *string           `json:"target"`
	URL             *string           `json:"url"`
	Configuration   map[string]string `json:"configuration,omitempty"`
	DateCreated     *string           `json:"date_created"`
	DateUpdated     *string           `json:"date_updated"`
}

// ConversationsV1ServiceConversationScopedWebhookList is the paginated response.
type ConversationsV1ServiceConversationScopedWebhookList struct {
	Webhooks []ConversationsV1ServiceConversationScopedWebhook `json:"webhooks"`
	Meta     VoiceV1Meta                                       `json:"meta"`
}

// ConversationsV1ServiceConversationWithParticipants is the response of the
// per-service composite-create endpoint. Mirrors ServiceConversation.
type ConversationsV1ServiceConversationWithParticipants struct {
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

// ConversationsV1ServiceParticipantConversation mirrors the account-level
// ParticipantConversation inverse-index row with an added chat_service_sid.
type ConversationsV1ServiceParticipantConversation struct {
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

// ConversationsV1ServiceParticipantConversationList is the paginated response.
type ConversationsV1ServiceParticipantConversationList struct {
	Conversations []ConversationsV1ServiceParticipantConversation `json:"conversations"`
	Meta          VoiceV1Meta                                     `json:"meta"`
}

// ConversationsV1ServiceUserConversation mirrors the account-level
// UserConversation per-user view with an added chat_service_sid.
type ConversationsV1ServiceUserConversation struct {
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

// ConversationsV1ServiceUserConversationList is the paginated response.
type ConversationsV1ServiceUserConversationList struct {
	Conversations []ConversationsV1ServiceUserConversation `json:"conversations"`
	Meta          VoiceV1Meta                              `json:"meta"`
}

// ConversationsV1ServiceRole mirrors the account-level Role with an added
// chat_service_sid.
type ConversationsV1ServiceRole struct {
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

// ConversationsV1ServiceRoleList is the paginated response.
type ConversationsV1ServiceRoleList struct {
	Roles []ConversationsV1ServiceRole `json:"roles"`
	Meta  VoiceV1Meta                  `json:"meta"`
}

// ConversationsV1ServiceUser mirrors the account-level User with an added
// chat_service_sid.
type ConversationsV1ServiceUser struct {
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

// ConversationsV1ServiceUserList is the paginated response.
type ConversationsV1ServiceUserList struct {
	Users []ConversationsV1ServiceUser `json:"users"`
	Meta  VoiceV1Meta                  `json:"meta"`
}

// ConversationsV1ServiceBinding is a push-notification Binding (BS...) for a
// given Chat Service. List + fetch + delete only.
type ConversationsV1ServiceBinding struct {
	Sid            *string  `json:"sid"`
	AccountSid     *string  `json:"account_sid"`
	ChatServiceSid *string  `json:"chat_service_sid"`
	CredentialSid  *string  `json:"credential_sid"`
	DateCreated    *string  `json:"date_created"`
	DateUpdated    *string  `json:"date_updated"`
	Endpoint       *string  `json:"endpoint"`
	Identity       *string  `json:"identity"`
	BindingType    string   `json:"binding_type"`
	MessageTypes   []string `json:"message_types,omitempty"`
	URL            *string  `json:"url"`
}

// ConversationsV1ServiceBindingList is the paginated response.
type ConversationsV1ServiceBindingList struct {
	Bindings []ConversationsV1ServiceBinding `json:"bindings"`
	Meta     VoiceV1Meta                     `json:"meta"`
}

// ConversationsV1ServiceConfiguration is the per-service Configuration
// singleton. Fetch + update only.
type ConversationsV1ServiceConfiguration struct {
	ChatServiceSid                    *string           `json:"chat_service_sid"`
	DefaultConversationCreatorRoleSid *string           `json:"default_conversation_creator_role_sid"`
	DefaultConversationRoleSid        *string           `json:"default_conversation_role_sid"`
	DefaultChatServiceRoleSid         *string           `json:"default_chat_service_role_sid"`
	URL                               *string           `json:"url"`
	Links                             map[string]string `json:"links,omitempty"`
	ReachabilityEnabled               *bool             `json:"reachability_enabled"`
}

// ConversationsV1ServiceNotification is the per-service push Notification
// configuration singleton. Fetch + update only.
type ConversationsV1ServiceNotification struct {
	AccountSid              *string                `json:"account_sid"`
	ChatServiceSid          *string                `json:"chat_service_sid"`
	NewMessage              map[string]interface{} `json:"new_message,omitempty"`
	AddedToConversation     map[string]interface{} `json:"added_to_conversation,omitempty"`
	RemovedFromConversation map[string]interface{} `json:"removed_from_conversation,omitempty"`
	LogEnabled              *bool                  `json:"log_enabled"`
	URL                     *string                `json:"url"`
}

// ConversationsV1ServiceWebhookConfiguration is the per-service Webhook
// configuration singleton. Fetch + update only.
type ConversationsV1ServiceWebhookConfiguration struct {
	AccountSid     *string  `json:"account_sid"`
	ChatServiceSid *string  `json:"chat_service_sid"`
	PreWebhookURL  *string  `json:"pre_webhook_url"`
	PostWebhookURL *string  `json:"post_webhook_url"`
	Filters        []string `json:"filters,omitempty"`
	Method         string   `json:"method"`
	URL            *string  `json:"url"`
}

// ---------------------------------------------------------------------------
// Request params.
// ---------------------------------------------------------------------------

// CreateServiceConversationRequest is the body for POST
// /v1/Services/{ChatServiceSid}/Conversations. Dotted Timers.* keys go
// through verbatim.
type CreateServiceConversationRequest struct {
	FriendlyName        *string `form:"FriendlyName"`
	UniqueName          *string `form:"UniqueName"`
	MessagingServiceSid *string `form:"MessagingServiceSid"`
	Attributes          *string `form:"Attributes"`
	State               *string `form:"State"`
	TimersInactive      *string `form:"Timers.Inactive"`
	TimersClosed        *string `form:"Timers.Closed"`
}

func (p CreateServiceConversationRequest) form() url.Values {
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

// UpdateServiceConversationRequest is the body for POST
// /v1/Services/{ChatServiceSid}/Conversations/{ConversationSid}.
type UpdateServiceConversationRequest struct {
	FriendlyName   *string `form:"FriendlyName"`
	UniqueName     *string `form:"UniqueName"`
	Attributes     *string `form:"Attributes"`
	State          *string `form:"State"`
	TimersInactive *string `form:"Timers.Inactive"`
	TimersClosed   *string `form:"Timers.Closed"`
}

func (p UpdateServiceConversationRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "UniqueName", p.UniqueName)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "State", p.State)
	setStr(v, "Timers.Inactive", p.TimersInactive)
	setStr(v, "Timers.Closed", p.TimersClosed)
	return v
}

// CreateServiceMessageRequest is the body for POST
// /v1/Services/{ChatServiceSid}/Conversations/{ConversationSid}/Messages.
type CreateServiceMessageRequest struct {
	Author     *string `form:"Author"`
	Body       *string `form:"Body"`
	Attributes *string `form:"Attributes"`
	ContentSid *string `form:"ContentSid"`
}

func (p CreateServiceMessageRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Author", p.Author)
	setStr(v, "Body", p.Body)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "ContentSid", p.ContentSid)
	return v
}

// UpdateServiceMessageRequest is the body for POST
// .../Messages/{MessageSid}.
type UpdateServiceMessageRequest struct {
	Author     *string `form:"Author"`
	Body       *string `form:"Body"`
	Attributes *string `form:"Attributes"`
}

func (p UpdateServiceMessageRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Author", p.Author)
	setStr(v, "Body", p.Body)
	setStr(v, "Attributes", p.Attributes)
	return v
}

// CreateServiceParticipantRequest is the body for POST .../Participants.
type CreateServiceParticipantRequest struct {
	Identity                         *string `form:"Identity"`
	Attributes                       *string `form:"Attributes"`
	RoleSid                          *string `form:"RoleSid"`
	MessagingBindingAddress          *string `form:"MessagingBinding.Address"`
	MessagingBindingProxyAddress     *string `form:"MessagingBinding.ProxyAddress"`
	MessagingBindingProjectedAddress *string `form:"MessagingBinding.ProjectedAddress"`
}

func (p CreateServiceParticipantRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Identity", p.Identity)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "RoleSid", p.RoleSid)
	setStr(v, "MessagingBinding.Address", p.MessagingBindingAddress)
	setStr(v, "MessagingBinding.ProxyAddress", p.MessagingBindingProxyAddress)
	setStr(v, "MessagingBinding.ProjectedAddress", p.MessagingBindingProjectedAddress)
	return v
}

// UpdateServiceParticipantRequest is the body for POST
// .../Participants/{ParticipantSid}. The service-scoped update accepts only
// Attributes and RoleSid (spec narrows the surface vs the account variant).
type UpdateServiceParticipantRequest struct {
	Attributes *string `form:"Attributes"`
	RoleSid    *string `form:"RoleSid"`
}

func (p UpdateServiceParticipantRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "RoleSid", p.RoleSid)
	return v
}

// CreateServiceScopedWebhookRequest is the body for POST .../Webhooks.
// Target is required.
type CreateServiceScopedWebhookRequest struct {
	Target               string  `form:"Target"`
	ConfigurationURL     *string `form:"Configuration.Url"`
	ConfigurationMethod  *string `form:"Configuration.Method"`
	ConfigurationFlowSid *string `form:"Configuration.FlowSid"`
}

func (p CreateServiceScopedWebhookRequest) form() url.Values {
	v := url.Values{}
	v.Set("Target", p.Target)
	setStr(v, "Configuration.Url", p.ConfigurationURL)
	setStr(v, "Configuration.Method", p.ConfigurationMethod)
	setStr(v, "Configuration.FlowSid", p.ConfigurationFlowSid)
	return v
}

// UpdateServiceScopedWebhookRequest is the body for POST
// .../Webhooks/{WebhookSid}.
type UpdateServiceScopedWebhookRequest struct {
	ConfigurationURL     *string `form:"Configuration.Url"`
	ConfigurationMethod  *string `form:"Configuration.Method"`
	ConfigurationFlowSid *string `form:"Configuration.FlowSid"`
}

func (p UpdateServiceScopedWebhookRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "Configuration.Url", p.ConfigurationURL)
	setStr(v, "Configuration.Method", p.ConfigurationMethod)
	setStr(v, "Configuration.FlowSid", p.ConfigurationFlowSid)
	return v
}

// CreateServiceRoleRequest is the body for POST
// /v1/Services/{ChatServiceSid}/Roles. All three fields required; Permission
// is repeated form-encoded.
type CreateServiceRoleRequest struct {
	FriendlyName string   `form:"FriendlyName"`
	Type         string   `form:"Type"`
	Permission   []string `form:"Permission"`
}

func (p CreateServiceRoleRequest) form() url.Values {
	v := url.Values{}
	v.Set("FriendlyName", p.FriendlyName)
	v.Set("Type", p.Type)
	addStrings(v, "Permission", p.Permission)
	return v
}

// UpdateServiceRoleRequest is the body for POST .../Roles/{Sid}.
type UpdateServiceRoleRequest struct {
	Permission []string `form:"Permission"`
}

func (p UpdateServiceRoleRequest) form() url.Values {
	v := url.Values{}
	addStrings(v, "Permission", p.Permission)
	return v
}

// CreateServiceUserRequest is the body for POST
// /v1/Services/{ChatServiceSid}/Users. Identity is required.
type CreateServiceUserRequest struct {
	Identity     string  `form:"Identity"`
	FriendlyName *string `form:"FriendlyName"`
	Attributes   *string `form:"Attributes"`
	RoleSid      *string `form:"RoleSid"`
}

func (p CreateServiceUserRequest) form() url.Values {
	v := url.Values{}
	v.Set("Identity", p.Identity)
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "RoleSid", p.RoleSid)
	return v
}

// UpdateServiceUserRequest is the body for POST .../Users/{Sid}.
type UpdateServiceUserRequest struct {
	FriendlyName *string `form:"FriendlyName"`
	Attributes   *string `form:"Attributes"`
	RoleSid      *string `form:"RoleSid"`
}

func (p UpdateServiceUserRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "FriendlyName", p.FriendlyName)
	setStr(v, "Attributes", p.Attributes)
	setStr(v, "RoleSid", p.RoleSid)
	return v
}

// CreateServiceConversationWithParticipantsRequest is the body for POST
// /v1/Services/{ChatServiceSid}/ConversationWithParticipants. Participant is
// repeated; each value is a JSON-encoded participant spec.
type CreateServiceConversationWithParticipantsRequest struct {
	FriendlyName        *string  `form:"FriendlyName"`
	UniqueName          *string  `form:"UniqueName"`
	MessagingServiceSid *string  `form:"MessagingServiceSid"`
	Attributes          *string  `form:"Attributes"`
	State               *string  `form:"State"`
	TimersInactive      *string  `form:"Timers.Inactive"`
	TimersClosed        *string  `form:"Timers.Closed"`
	Participant         []string `form:"Participant"`
}

func (p CreateServiceConversationWithParticipantsRequest) form() url.Values {
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

// ListServiceParticipantConversationsParams is the query for GET
// /v1/Services/{ChatServiceSid}/ParticipantConversations.
type ListServiceParticipantConversationsParams struct {
	Identity *string
	Address  *string
	PageSize *int
}

func (p ListServiceParticipantConversationsParams) query() url.Values {
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

// ListServiceBindingsParams is the query for GET
// /v1/Services/{ChatServiceSid}/Bindings.
type ListServiceBindingsParams struct {
	BindingType *string
	Identity    *string
	PageSize    *int
}

func (p ListServiceBindingsParams) query() url.Values {
	v := url.Values{}
	if p.BindingType != nil {
		v.Set("BindingType", *p.BindingType)
	}
	if p.Identity != nil {
		v.Set("Identity", *p.Identity)
	}
	if p.PageSize != nil {
		v.Set("PageSize", strconv.Itoa(*p.PageSize))
	}
	return v
}

// UpdateServiceConfigurationRequest is the body for POST
// /v1/Services/{ChatServiceSid}/Configuration.
type UpdateServiceConfigurationRequest struct {
	DefaultChatServiceRoleSid         *string `form:"DefaultChatServiceRoleSid"`
	DefaultConversationCreatorRoleSid *string `form:"DefaultConversationCreatorRoleSid"`
	DefaultConversationRoleSid        *string `form:"DefaultConversationRoleSid"`
	ReachabilityEnabled               *bool   `form:"ReachabilityEnabled"`
}

func (p UpdateServiceConfigurationRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "DefaultChatServiceRoleSid", p.DefaultChatServiceRoleSid)
	setStr(v, "DefaultConversationCreatorRoleSid", p.DefaultConversationCreatorRoleSid)
	setStr(v, "DefaultConversationRoleSid", p.DefaultConversationRoleSid)
	setBool(v, "ReachabilityEnabled", p.ReachabilityEnabled)
	return v
}

// UpdateServiceNotificationRequest is the body for POST
// /v1/Services/{ChatServiceSid}/Configuration/Notifications. Dotted keys
// (NewMessage.*, AddedToConversation.*, RemovedFromConversation.*,
// NewMessage.WithMedia.*) are emitted verbatim.
type UpdateServiceNotificationRequest struct {
	LogEnabled                      *bool   `form:"LogEnabled"`
	NewMessageEnabled               *bool   `form:"NewMessage.Enabled"`
	NewMessageTemplate              *string `form:"NewMessage.Template"`
	NewMessageSound                 *string `form:"NewMessage.Sound"`
	NewMessageBadgeCountEnabled     *bool   `form:"NewMessage.BadgeCountEnabled"`
	NewMessageWithMediaEnabled      *bool   `form:"NewMessage.WithMedia.Enabled"`
	NewMessageWithMediaTemplate     *string `form:"NewMessage.WithMedia.Template"`
	AddedToConversationEnabled      *bool   `form:"AddedToConversation.Enabled"`
	AddedToConversationTemplate     *string `form:"AddedToConversation.Template"`
	AddedToConversationSound        *string `form:"AddedToConversation.Sound"`
	RemovedFromConversationEnabled  *bool   `form:"RemovedFromConversation.Enabled"`
	RemovedFromConversationTemplate *string `form:"RemovedFromConversation.Template"`
	RemovedFromConversationSound    *string `form:"RemovedFromConversation.Sound"`
}

func (p UpdateServiceNotificationRequest) form() url.Values {
	v := url.Values{}
	setBool(v, "LogEnabled", p.LogEnabled)
	setBool(v, "NewMessage.Enabled", p.NewMessageEnabled)
	setStr(v, "NewMessage.Template", p.NewMessageTemplate)
	setStr(v, "NewMessage.Sound", p.NewMessageSound)
	setBool(v, "NewMessage.BadgeCountEnabled", p.NewMessageBadgeCountEnabled)
	setBool(v, "NewMessage.WithMedia.Enabled", p.NewMessageWithMediaEnabled)
	setStr(v, "NewMessage.WithMedia.Template", p.NewMessageWithMediaTemplate)
	setBool(v, "AddedToConversation.Enabled", p.AddedToConversationEnabled)
	setStr(v, "AddedToConversation.Template", p.AddedToConversationTemplate)
	setStr(v, "AddedToConversation.Sound", p.AddedToConversationSound)
	setBool(v, "RemovedFromConversation.Enabled", p.RemovedFromConversationEnabled)
	setStr(v, "RemovedFromConversation.Template", p.RemovedFromConversationTemplate)
	setStr(v, "RemovedFromConversation.Sound", p.RemovedFromConversationSound)
	return v
}

// UpdateServiceWebhookConfigurationRequest is the body for POST
// /v1/Services/{ChatServiceSid}/Configuration/Webhooks. Filters is repeated
// form-encoded.
type UpdateServiceWebhookConfigurationRequest struct {
	PreWebhookURL  *string  `form:"PreWebhookUrl"`
	PostWebhookURL *string  `form:"PostWebhookUrl"`
	Method         *string  `form:"Method"`
	Filters        []string `form:"Filters"`
}

func (p UpdateServiceWebhookConfigurationRequest) form() url.Values {
	v := url.Values{}
	setStr(v, "PreWebhookUrl", p.PreWebhookURL)
	setStr(v, "PostWebhookUrl", p.PostWebhookURL)
	setStr(v, "Method", p.Method)
	addStrings(v, "Filters", p.Filters)
	return v
}

// ---------------------------------------------------------------------------
// Methods on *ConversationsV1Service.
// ---------------------------------------------------------------------------

// --- ServiceConversation ---------------------------------------------------

// CreateServiceConversation adds a Conversation inside a Chat Service realm.
func (s *ConversationsV1Service) CreateServiceConversation(ctx context.Context, chatServiceSid string, params CreateServiceConversationRequest) (*ConversationsV1ServiceConversation, error) {
	var out ConversationsV1ServiceConversation
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Conversations", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListServiceConversations returns a single page of service-scoped Conversations.
func (s *ConversationsV1Service) ListServiceConversations(ctx context.Context, chatServiceSid string, params V1PageParams) (*ConversationsV1ServiceConversationList, error) {
	var out ConversationsV1ServiceConversationList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchServiceConversation retrieves a service-scoped Conversation by sid.
func (s *ConversationsV1Service) FetchServiceConversation(ctx context.Context, chatServiceSid, conversationSid string) (*ConversationsV1ServiceConversation, error) {
	var out ConversationsV1ServiceConversation
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateServiceConversation mutates a service-scoped Conversation in place.
func (s *ConversationsV1Service) UpdateServiceConversation(ctx context.Context, chatServiceSid, conversationSid string, params UpdateServiceConversationRequest) (*ConversationsV1ServiceConversation, error) {
	var out ConversationsV1ServiceConversation
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteServiceConversation removes a service-scoped Conversation.
func (s *ConversationsV1Service) DeleteServiceConversation(ctx context.Context, chatServiceSid, conversationSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid,
	}, nil)
}

// --- ServiceConversationMessage --------------------------------------------

// CreateServiceMessage adds a Message to a service-scoped Conversation.
func (s *ConversationsV1Service) CreateServiceMessage(ctx context.Context, chatServiceSid, conversationSid string, params CreateServiceMessageRequest) (*ConversationsV1ServiceConversationMessage, error) {
	var out ConversationsV1ServiceConversationMessage
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Messages", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListServiceMessages returns a single page of service-scoped Messages.
func (s *ConversationsV1Service) ListServiceMessages(ctx context.Context, chatServiceSid, conversationSid string, params V1PageParams) (*ConversationsV1ServiceConversationMessageList, error) {
	var out ConversationsV1ServiceConversationMessageList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Messages", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchServiceMessage retrieves a service-scoped Message by sid.
func (s *ConversationsV1Service) FetchServiceMessage(ctx context.Context, chatServiceSid, conversationSid, messageSid string) (*ConversationsV1ServiceConversationMessage, error) {
	var out ConversationsV1ServiceConversationMessage
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Messages/" + messageSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateServiceMessage mutates a service-scoped Message in place.
func (s *ConversationsV1Service) UpdateServiceMessage(ctx context.Context, chatServiceSid, conversationSid, messageSid string, params UpdateServiceMessageRequest) (*ConversationsV1ServiceConversationMessage, error) {
	var out ConversationsV1ServiceConversationMessage
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Messages/" + messageSid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteServiceMessage removes a service-scoped Message.
func (s *ConversationsV1Service) DeleteServiceMessage(ctx context.Context, chatServiceSid, conversationSid, messageSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Messages/" + messageSid,
	}, nil)
}

// --- ServiceConversationMessageReceipt -------------------------------------

// ListServiceMessageReceipts returns a single page of a service-scoped
// Message's delivery receipts.
func (s *ConversationsV1Service) ListServiceMessageReceipts(ctx context.Context, chatServiceSid, conversationSid, messageSid string, params V1PageParams) (*ConversationsV1ServiceConversationMessageReceiptList, error) {
	var out ConversationsV1ServiceConversationMessageReceiptList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Messages/" + messageSid + "/Receipts", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchServiceMessageReceipt retrieves a single service-scoped delivery
// receipt by sid.
func (s *ConversationsV1Service) FetchServiceMessageReceipt(ctx context.Context, chatServiceSid, conversationSid, messageSid, sid string) (*ConversationsV1ServiceConversationMessageReceipt, error) {
	var out ConversationsV1ServiceConversationMessageReceipt
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Messages/" + messageSid + "/Receipts/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- ServiceConversationParticipant ----------------------------------------

// CreateServiceParticipant adds a Participant to a service-scoped Conversation.
func (s *ConversationsV1Service) CreateServiceParticipant(ctx context.Context, chatServiceSid, conversationSid string, params CreateServiceParticipantRequest) (*ConversationsV1ServiceConversationParticipant, error) {
	var out ConversationsV1ServiceConversationParticipant
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Participants", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListServiceParticipants returns a single page of service-scoped Participants.
func (s *ConversationsV1Service) ListServiceParticipants(ctx context.Context, chatServiceSid, conversationSid string, params V1PageParams) (*ConversationsV1ServiceConversationParticipantList, error) {
	var out ConversationsV1ServiceConversationParticipantList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Participants", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchServiceParticipant retrieves a service-scoped Participant by sid.
func (s *ConversationsV1Service) FetchServiceParticipant(ctx context.Context, chatServiceSid, conversationSid, participantSid string) (*ConversationsV1ServiceConversationParticipant, error) {
	var out ConversationsV1ServiceConversationParticipant
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Participants/" + participantSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateServiceParticipant mutates a service-scoped Participant in place.
func (s *ConversationsV1Service) UpdateServiceParticipant(ctx context.Context, chatServiceSid, conversationSid, participantSid string, params UpdateServiceParticipantRequest) (*ConversationsV1ServiceConversationParticipant, error) {
	var out ConversationsV1ServiceConversationParticipant
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Participants/" + participantSid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteServiceParticipant removes a service-scoped Participant.
func (s *ConversationsV1Service) DeleteServiceParticipant(ctx context.Context, chatServiceSid, conversationSid, participantSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Participants/" + participantSid,
	}, nil)
}

// --- ServiceConversationScopedWebhook --------------------------------------

// CreateServiceScopedWebhook adds a webhook scoped to a service-scoped Conversation.
func (s *ConversationsV1Service) CreateServiceScopedWebhook(ctx context.Context, chatServiceSid, conversationSid string, params CreateServiceScopedWebhookRequest) (*ConversationsV1ServiceConversationScopedWebhook, error) {
	var out ConversationsV1ServiceConversationScopedWebhook
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Webhooks", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListServiceScopedWebhooks returns a single page of service-scoped Scoped Webhooks.
func (s *ConversationsV1Service) ListServiceScopedWebhooks(ctx context.Context, chatServiceSid, conversationSid string, params V1PageParams) (*ConversationsV1ServiceConversationScopedWebhookList, error) {
	var out ConversationsV1ServiceConversationScopedWebhookList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Webhooks", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchServiceScopedWebhook retrieves a service-scoped Scoped Webhook by sid.
func (s *ConversationsV1Service) FetchServiceScopedWebhook(ctx context.Context, chatServiceSid, conversationSid, webhookSid string) (*ConversationsV1ServiceConversationScopedWebhook, error) {
	var out ConversationsV1ServiceConversationScopedWebhook
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Webhooks/" + webhookSid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateServiceScopedWebhook mutates a service-scoped Scoped Webhook in place.
func (s *ConversationsV1Service) UpdateServiceScopedWebhook(ctx context.Context, chatServiceSid, conversationSid, webhookSid string, params UpdateServiceScopedWebhookRequest) (*ConversationsV1ServiceConversationScopedWebhook, error) {
	var out ConversationsV1ServiceConversationScopedWebhook
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Webhooks/" + webhookSid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteServiceScopedWebhook removes a service-scoped Scoped Webhook.
func (s *ConversationsV1Service) DeleteServiceScopedWebhook(ctx context.Context, chatServiceSid, conversationSid, webhookSid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Services/" + chatServiceSid + "/Conversations/" + conversationSid + "/Webhooks/" + webhookSid,
	}, nil)
}

// --- ServiceConversationWithParticipants -----------------------------------

// CreateServiceConversationWithParticipants adds a service-scoped
// Conversation and its initial Participants in one call.
func (s *ConversationsV1Service) CreateServiceConversationWithParticipants(ctx context.Context, chatServiceSid string, params CreateServiceConversationWithParticipantsRequest) (*ConversationsV1ServiceConversationWithParticipants, error) {
	var out ConversationsV1ServiceConversationWithParticipants
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/ConversationWithParticipants", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- ServiceParticipantConversation (list-only) ----------------------------

// ListServiceParticipantConversations returns the conversations a
// participant belongs to inside a Chat Service realm, optionally filtered
// by Identity or messaging-binding Address.
func (s *ConversationsV1Service) ListServiceParticipantConversations(ctx context.Context, chatServiceSid string, params ListServiceParticipantConversationsParams) (*ConversationsV1ServiceParticipantConversationList, error) {
	var out ConversationsV1ServiceParticipantConversationList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/ParticipantConversations", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- ServiceUserConversation (list-only, nested under /Users/{UserSid}) ----

// ListServiceUserConversations returns a single page of a User's
// conversations within a Chat Service realm.
func (s *ConversationsV1Service) ListServiceUserConversations(ctx context.Context, chatServiceSid, userSid string, params V1PageParams) (*ConversationsV1ServiceUserConversationList, error) {
	var out ConversationsV1ServiceUserConversationList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Users/" + userSid + "/Conversations", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- ServiceRole -----------------------------------------------------------

// CreateServiceRole adds a Role inside a Chat Service realm.
func (s *ConversationsV1Service) CreateServiceRole(ctx context.Context, chatServiceSid string, params CreateServiceRoleRequest) (*ConversationsV1ServiceRole, error) {
	var out ConversationsV1ServiceRole
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Roles", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListServiceRoles returns a single page of service-scoped Roles.
func (s *ConversationsV1Service) ListServiceRoles(ctx context.Context, chatServiceSid string, params V1PageParams) (*ConversationsV1ServiceRoleList, error) {
	var out ConversationsV1ServiceRoleList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Roles", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchServiceRole retrieves a service-scoped Role by sid.
func (s *ConversationsV1Service) FetchServiceRole(ctx context.Context, chatServiceSid, sid string) (*ConversationsV1ServiceRole, error) {
	var out ConversationsV1ServiceRole
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Roles/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateServiceRole replaces a service-scoped Role's permission list.
func (s *ConversationsV1Service) UpdateServiceRole(ctx context.Context, chatServiceSid, sid string, params UpdateServiceRoleRequest) (*ConversationsV1ServiceRole, error) {
	var out ConversationsV1ServiceRole
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Roles/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteServiceRole removes a service-scoped Role.
func (s *ConversationsV1Service) DeleteServiceRole(ctx context.Context, chatServiceSid, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Services/" + chatServiceSid + "/Roles/" + sid,
	}, nil)
}

// --- ServiceUser -----------------------------------------------------------

// CreateServiceUser adds a User inside a Chat Service realm.
func (s *ConversationsV1Service) CreateServiceUser(ctx context.Context, chatServiceSid string, params CreateServiceUserRequest) (*ConversationsV1ServiceUser, error) {
	var out ConversationsV1ServiceUser
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Users", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListServiceUsers returns a single page of service-scoped Users.
func (s *ConversationsV1Service) ListServiceUsers(ctx context.Context, chatServiceSid string, params V1PageParams) (*ConversationsV1ServiceUserList, error) {
	var out ConversationsV1ServiceUserList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Users", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchServiceUser retrieves a service-scoped User by sid.
func (s *ConversationsV1Service) FetchServiceUser(ctx context.Context, chatServiceSid, sid string) (*ConversationsV1ServiceUser, error) {
	var out ConversationsV1ServiceUser
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Users/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateServiceUser mutates a service-scoped User in place.
func (s *ConversationsV1Service) UpdateServiceUser(ctx context.Context, chatServiceSid, sid string, params UpdateServiceUserRequest) (*ConversationsV1ServiceUser, error) {
	var out ConversationsV1ServiceUser
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Users/" + sid, form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteServiceUser removes a service-scoped User.
func (s *ConversationsV1Service) DeleteServiceUser(ctx context.Context, chatServiceSid, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Services/" + chatServiceSid + "/Users/" + sid,
	}, nil)
}

// --- ServiceBinding --------------------------------------------------------

// ListServiceBindings returns a single page of push Bindings for a Chat
// Service realm. Optional BindingType and Identity filters.
func (s *ConversationsV1Service) ListServiceBindings(ctx context.Context, chatServiceSid string, params ListServiceBindingsParams) (*ConversationsV1ServiceBindingList, error) {
	var out ConversationsV1ServiceBindingList
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Bindings", query: params.query(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchServiceBinding retrieves a push Binding by sid.
func (s *ConversationsV1Service) FetchServiceBinding(ctx context.Context, chatServiceSid, sid string) (*ConversationsV1ServiceBinding, error) {
	var out ConversationsV1ServiceBinding
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Bindings/" + sid,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteServiceBinding removes a push Binding.
func (s *ConversationsV1Service) DeleteServiceBinding(ctx context.Context, chatServiceSid, sid string) error {
	return s.c.t.do(ctx, requestOpts{
		method: "DELETE", path: "/v1/Services/" + chatServiceSid + "/Bindings/" + sid,
	}, nil)
}

// --- ServiceConfiguration (singleton) --------------------------------------

// FetchServiceConfiguration retrieves the per-service Configuration.
func (s *ConversationsV1Service) FetchServiceConfiguration(ctx context.Context, chatServiceSid string) (*ConversationsV1ServiceConfiguration, error) {
	var out ConversationsV1ServiceConfiguration
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Configuration",
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateServiceConfiguration mutates the per-service Configuration.
func (s *ConversationsV1Service) UpdateServiceConfiguration(ctx context.Context, chatServiceSid string, params UpdateServiceConfigurationRequest) (*ConversationsV1ServiceConfiguration, error) {
	var out ConversationsV1ServiceConfiguration
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Configuration", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- ServiceNotification (singleton at /Configuration/Notifications) -------

// FetchServiceNotification retrieves the per-service push Notification
// configuration.
func (s *ConversationsV1Service) FetchServiceNotification(ctx context.Context, chatServiceSid string) (*ConversationsV1ServiceNotification, error) {
	var out ConversationsV1ServiceNotification
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Configuration/Notifications",
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateServiceNotification mutates the per-service push Notification
// configuration.
func (s *ConversationsV1Service) UpdateServiceNotification(ctx context.Context, chatServiceSid string, params UpdateServiceNotificationRequest) (*ConversationsV1ServiceNotification, error) {
	var out ConversationsV1ServiceNotification
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Configuration/Notifications", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- ServiceWebhookConfiguration (singleton at /Configuration/Webhooks) ----

// FetchServiceWebhookConfiguration retrieves the per-service Webhook
// configuration.
func (s *ConversationsV1Service) FetchServiceWebhookConfiguration(ctx context.Context, chatServiceSid string) (*ConversationsV1ServiceWebhookConfiguration, error) {
	var out ConversationsV1ServiceWebhookConfiguration
	if err := s.c.t.do(ctx, requestOpts{
		method: "GET", path: "/v1/Services/" + chatServiceSid + "/Configuration/Webhooks",
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateServiceWebhookConfiguration mutates the per-service Webhook
// configuration.
func (s *ConversationsV1Service) UpdateServiceWebhookConfiguration(ctx context.Context, chatServiceSid string, params UpdateServiceWebhookConfigurationRequest) (*ConversationsV1ServiceWebhookConfiguration, error) {
	var out ConversationsV1ServiceWebhookConfiguration
	if err := s.c.t.do(ctx, requestOpts{
		method: "POST", path: "/v1/Services/" + chatServiceSid + "/Configuration/Webhooks", form: params.form(),
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
