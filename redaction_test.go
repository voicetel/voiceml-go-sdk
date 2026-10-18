package voiceml_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

const secretKey = "sk-live-do-not-print-0123456789"

func TestSecretsNeverFormatted(t *testing.T) {
	opts := voiceml.ClientOptions{AccountSid: testAccountSid, APIKey: secretKey}
	tokenOpts := voiceml.ClientOptions{AccountSid: testAccountSid, AuthToken: secretKey}
	c, err := voiceml.NewClient(opts)
	if err != nil {
		t.Fatal(err)
	}
	wrapped := struct {
		Opts   voiceml.ClientOptions
		Client *voiceml.Client
	}{opts, c}

	var outputs []string
	validator := voiceml.NewRequestValidator(secretKey)
	rotated := voiceml.AccountWithKey{Sid: testAccountSid, APIKey: secretKey}
	for _, v := range []any{opts, &opts, tokenOpts, c, *c, wrapped, c.Calls, *c.Calls, c.ConversationsV1, *c.ConversationsV1, c.MessagingV1.Services, c.Diagnostics, *c.Diagnostics, validator, *validator, rotated, &rotated} {
		for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
			outputs = append(outputs, fmt.Sprintf(verb, v))
		}
	}
	for _, v := range []any{opts, tokenOpts, wrapped} {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		outputs = append(outputs, string(b))
	}
	var logBuf bytes.Buffer
	slog.New(slog.NewJSONHandler(&logBuf, nil)).Info("configured", "opts", opts)
	slog.New(slog.NewTextHandler(&logBuf, nil)).Info("configured", "opts", tokenOpts)
	outputs = append(outputs, logBuf.String())

	for _, out := range outputs {
		if strings.Contains(out, secretKey) {
			t.Fatalf("API key leaked: %s", out)
		}
	}
	if !strings.Contains(fmt.Sprint(opts), "[redacted]") || !strings.Contains(fmt.Sprint(opts), testAccountSid) {
		t.Fatalf("String() should show the AccountSid and a redaction marker: %s", opts)
	}
}

func TestAPIErrorDoesNotEchoCredentials(t *testing.T) {
	c, _, cleanup := newClient(t, []handlerStep{
		jsonStep(401, map[string]any{"code": 20003, "message": "Authentication Error", "status": 401}),
	}, func(o *voiceml.ClientOptions) { o.APIKey = secretKey })
	defer cleanup()
	_, err := c.Calls.Get(context.Background(), "CA1")
	if err == nil || strings.Contains(fmt.Sprintf("%+v", err), secretKey) {
		t.Fatalf("err = %v", err)
	}
}
