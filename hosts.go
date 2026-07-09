package voiceml

// Per-product host resolution for the VoiceML API.
//
// Twilio splits its products across dedicated subdomains (api.twilio.com,
// conversations.twilio.com, messaging.twilio.com, ...). VoiceML mirrors that
// shape on voicetel.com: the Conversations product answers on
// conversations.voicetel.com and the Messaging Service product on
// messaging.voicetel.com, while everything else stays on the default
// voiceml.voicetel.com host. Conversation Service and Messaging Service share
// the identical /v1/Services path shape, so the host is what disambiguates them
// on the wire.
//
// Given the configured base URL these helpers derive the two product hosts by
// swapping the leftmost "voiceml" label — but only for recognised
// *.voicetel.com hosts. For any other base URL (a self-hosted callBroadcast
// instance, a test stub, a regional override) the product hosts fall back to
// the configured host unchanged, so a single-host deployment keeps working. A
// caller who needs Messaging Service against a custom host points
// MessagingBaseURL (or ConversationsBaseURL) at their own subdomain explicitly.

import (
	"net/url"
	"strings"
)

// deriveProductHost swaps the "voiceml" label of a *.voicetel.com host for
// product. Returns baseURL unchanged when the host is not a
// voiceml.*.voicetel.com style host (e.g. a self-hosted instance), so
// single-host deployments keep working without special-casing.
func deriveProductHost(baseURL, product string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	host := u.Hostname()
	if host == "" || !strings.HasSuffix(host, ".voicetel.com") {
		return baseURL
	}
	labels := strings.Split(host, ".")
	idx := -1
	for i, label := range labels {
		if label == "voiceml" {
			idx = i
			break
		}
	}
	if idx == -1 {
		return baseURL
	}
	labels[idx] = product
	newHost := strings.Join(labels, ".")
	if port := u.Port(); port != "" {
		u.Host = newHost + ":" + port
	} else {
		u.Host = newHost
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// resolveProductBaseURLs returns the (default, messaging, conversations) base
// URLs. Explicit overrides win; otherwise each product host is derived from
// baseURL (see the package comment above). All three are returned without a
// trailing slash.
func resolveProductBaseURLs(baseURL, messagingBaseURL, conversationsBaseURL string) (defaultURL, messagingURL, conversationsURL string) {
	defaultURL = strings.TrimRight(baseURL, "/")
	if messagingBaseURL == "" {
		messagingBaseURL = deriveProductHost(defaultURL, "messaging")
	}
	if conversationsBaseURL == "" {
		conversationsBaseURL = deriveProductHost(defaultURL, "conversations")
	}
	messagingURL = strings.TrimRight(messagingBaseURL, "/")
	conversationsURL = strings.TrimRight(conversationsBaseURL, "/")
	return defaultURL, messagingURL, conversationsURL
}
