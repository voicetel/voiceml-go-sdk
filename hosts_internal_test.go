// Internal unit tests for per-product host derivation (resolveProductBaseURLs).
// Kept in package voiceml so the unexported resolver can be exercised directly,
// mirroring the Python tests in tests/unit/test_v0_9_2.py.

package voiceml

import "testing"

func TestHostDerivationFromDefault(t *testing.T) {
	def, msg, conv := resolveProductBaseURLs("https://voiceml.voicetel.com", "", "")
	if def != "https://voiceml.voicetel.com" {
		t.Fatalf("default: %q", def)
	}
	if msg != "https://messaging.voicetel.com" {
		t.Fatalf("messaging: %q", msg)
	}
	if conv != "https://conversations.voicetel.com" {
		t.Fatalf("conversations: %q", conv)
	}
}

func TestHostDerivationRegional(t *testing.T) {
	def, msg, conv := resolveProductBaseURLs("https://east-1.us.voiceml.voicetel.com", "", "")
	if def != "https://east-1.us.voiceml.voicetel.com" {
		t.Fatalf("default: %q", def)
	}
	if msg != "https://east-1.us.messaging.voicetel.com" {
		t.Fatalf("messaging: %q", msg)
	}
	if conv != "https://east-1.us.conversations.voicetel.com" {
		t.Fatalf("conversations: %q", conv)
	}
}

func TestHostDerivationSelfHostedFallsBackToSingleHost(t *testing.T) {
	// A custom host has no "voiceml" label to swap — every product stays on it,
	// so a single-host deployment keeps working.
	def, msg, conv := resolveProductBaseURLs("https://pbx.acme.com", "", "")
	if def != "https://pbx.acme.com" || msg != "https://pbx.acme.com" || conv != "https://pbx.acme.com" {
		t.Fatalf("expected single-host fallback, got def=%q msg=%q conv=%q", def, msg, conv)
	}
}

func TestHostDerivationExplicitOverridesWin(t *testing.T) {
	def, msg, conv := resolveProductBaseURLs(
		"https://pbx.acme.com",
		"https://msg.acme.com",
		"https://conv.acme.com/",
	)
	if def != "https://pbx.acme.com" {
		t.Fatalf("default: %q", def)
	}
	if msg != "https://msg.acme.com" {
		t.Fatalf("messaging: %q", msg)
	}
	// Trailing slash is trimmed.
	if conv != "https://conv.acme.com" {
		t.Fatalf("conversations: %q", conv)
	}
}

func TestHostDerivationPreservesPort(t *testing.T) {
	def, msg, conv := resolveProductBaseURLs("https://voiceml.voicetel.com:8443", "", "")
	if def != "https://voiceml.voicetel.com:8443" {
		t.Fatalf("default: %q", def)
	}
	if msg != "https://messaging.voicetel.com:8443" {
		t.Fatalf("messaging: %q", msg)
	}
	if conv != "https://conversations.voicetel.com:8443" {
		t.Fatalf("conversations: %q", conv)
	}
}
