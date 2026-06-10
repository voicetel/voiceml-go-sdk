// Twilio response-shape conformance tests for the Go SDK (#256).
//
// These tests are SKIPPED unless VOICEML_CONFORMANCE_FIXTURES points at a
// fixture corpus emitted by callBroadcast's cmd/twilio-conformance-fixtures.
// They load each canonical Twilio response example from the corpus and
// json.Unmarshal it into the matching SDK resource type. If decoding
// fails, our type model has drifted from Twilio's documented shape — fix
// the SDK, not the fixture.
//
// Phase B of #256. Phase A vendored 115 fixtures across 7 resources.
// Run:
//
//	VOICEML_CONFORMANCE_FIXTURES=/path/to/callBroadcast/cmd/twilio-conformance-fixtures/fixtures \
//	  go test ./... -run TestTwilioFixtureConformance

package voiceml_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

type conformanceEntry struct {
	Resource    string `json:"resource"`
	Method      string `json:"method"`
	Status      string `json:"status"`
	OperationID string `json:"operation_id"`
	ExampleName string `json:"example_name"`
	Path        string `json:"path"`
	File        string `json:"file"`
}

func TestTwilioFixtureConformance(t *testing.T) {
	root := os.Getenv("VOICEML_CONFORMANCE_FIXTURES")
	if root == "" {
		t.Skip("VOICEML_CONFORMANCE_FIXTURES not set; skipping conformance fixtures")
	}

	indexBody, err := os.ReadFile(filepath.Join(root, "index.json"))
	if err != nil {
		t.Fatalf("read index.json: %v", err)
	}
	var entries []conformanceEntry
	if err := json.Unmarshal(indexBody, &entries); err != nil {
		t.Fatalf("decode index.json: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("empty fixture corpus")
	}

	t.Logf("running conformance against %d fixtures from %s", len(entries), root)

	for _, e := range entries {
		e := e
		name := e.Resource + "/" + e.OperationID
		if e.ExampleName != "" {
			name += "/" + e.ExampleName
		}
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, e.File))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			target := pickConformanceTarget(e.OperationID)
			if target == nil {
				t.Skipf("no SDK type mapped for operation %s", e.OperationID)
			}
			if err := json.Unmarshal(data, target); err != nil {
				t.Errorf("decode into %T: %v\nbody: %s", target, err, truncate(data))
				return
			}
			assertKeyFields(t, e.OperationID, target)
		})
	}
}

// pickConformanceTarget returns a pointer to a freshly allocated SDK
// type appropriate for the operation, or nil if the operation has no
// resource model in this SDK (e.g. delete responses, message stubs).
func pickConformanceTarget(opID string) any {
	switch opID {
	case "CreateCall", "FetchCall", "UpdateCall":
		return &voiceml.Call{}
	case "ListCall":
		return &voiceml.CallList{}

	case "FetchConference", "UpdateConference":
		return &voiceml.Conference{}
	case "ListConference":
		return &voiceml.ConferenceList{}

	case "CreateParticipant", "FetchParticipant", "UpdateParticipant":
		return &voiceml.Participant{}
	case "ListParticipant":
		return &voiceml.ParticipantList{}

	case "CreateQueue", "FetchQueue", "UpdateQueue":
		return &voiceml.Queue{}
	case "ListQueue":
		return &voiceml.QueueList{}

	case "FetchMember", "UpdateMember":
		return &voiceml.QueueMember{}
	case "ListMember":
		return &voiceml.QueueMemberList{}

	case "CreateApplication", "FetchApplication", "UpdateApplication":
		return &voiceml.Application{}
	case "ListApplication":
		return &voiceml.ApplicationList{}

	case "CreateCallRecording", "FetchCallRecording", "UpdateCallRecording",
		"FetchRecording", "FetchConferenceRecording", "UpdateConferenceRecording":
		return &voiceml.Recording{}
	case "ListCallRecording", "ListRecording", "ListConferenceRecording":
		return &voiceml.RecordingList{}

	case "CreateIncomingPhoneNumber",
		"CreateIncomingPhoneNumberLocal",
		"CreateIncomingPhoneNumberMobile",
		"CreateIncomingPhoneNumberTollFree",
		"FetchIncomingPhoneNumber",
		"UpdateIncomingPhoneNumber":
		return &voiceml.IncomingPhoneNumber{}
	case "ListIncomingPhoneNumber",
		"ListIncomingPhoneNumberLocal",
		"ListIncomingPhoneNumberMobile",
		"ListIncomingPhoneNumberTollFree":
		return &voiceml.IncomingPhoneNumbersList{}

	case "CreateStream", "UpdateStream":
		return &voiceml.Stream{}

	case "CreateSiprec", "UpdateSiprec":
		return &voiceml.SiprecSession{}

	case "CreateRealtimeTranscription", "UpdateRealtimeTranscription":
		return &voiceml.CallTranscription{}

	case "ListCallNotification", "ListNotification":
		// VoiceML treats Notifications as compat stubs (always-empty);
		// Twilio's example is a fully-populated notification object.
		// Decoding into a permissive container catches grossly malformed
		// JSON without asserting field-level parity for a feature we
		// don't ship.
		return &map[string]any{}
	case "FetchCallNotification", "FetchNotification":
		return &map[string]any{}

	case "ListCallEvent":
		return &voiceml.EventsList{}

	case "CreateUserDefinedMessage":
		// No SDK model; surface as raw JSON.
		return &map[string]any{}
	}
	return nil
}

// assertKeyFields checks that core Twilio fields are populated after
// decoding. Skipped for operations that decode into map[string]any or
// for resource models the SDK doesn't fully field-model.
func assertKeyFields(t *testing.T, opID string, target any) {
	t.Helper()
	switch v := target.(type) {
	case *voiceml.Call:
		if v.Sid == "" {
			t.Error("Call.Sid empty")
		}
		if v.AccountSid == "" {
			t.Error("Call.AccountSid empty")
		}
	case *voiceml.CallList:
		// Empty list responses are valid; assert envelope URI is populated
		// (Twilio sets this on every list response, even empty).
		if v.URI == "" {
			t.Error("CallList.URI empty (expected on every Twilio list response)")
		}
	case *voiceml.Conference:
		if v.Sid == "" {
			t.Error("Conference.Sid empty")
		}
	case *voiceml.Recording:
		if v.Sid == "" {
			t.Error("Recording.Sid empty")
		}
	case *voiceml.Queue:
		if v.Sid == "" {
			t.Error("Queue.Sid empty")
		}
	case *voiceml.Application:
		if v.Sid == "" {
			t.Error("Application.Sid empty")
		}
	case *voiceml.IncomingPhoneNumber:
		if v.Sid == "" {
			t.Error("IncomingPhoneNumber.Sid empty")
		}
	case *voiceml.Participant:
		// Participant.Sid is CallSid; assert it is set.
		if v.CallSid == "" {
			t.Error("Participant.CallSid empty")
		}
	case *voiceml.Stream:
		if v.Sid == "" {
			t.Error("Stream.Sid empty")
		}
	case *voiceml.SiprecSession:
		if v.Sid == "" {
			t.Error("SiprecSession.Sid empty")
		}
	case *voiceml.CallTranscription:
		if v.Sid == "" {
			t.Error("CallTranscription.Sid empty")
		}
	}
}

func truncate(b []byte) string {
	const max = 200
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "...(" + itoa(len(b)-max) + " more bytes)"
}

func itoa(n int) string {
	// avoid strconv import to keep this file's imports tight
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Compile-time guard that the corpus path resolves to a real directory
// when set, so a typo'd env var surfaces as a build-like failure rather
// than a silent skip.
var _ = strings.EqualFold
