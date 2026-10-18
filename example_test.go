package voiceml_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

func ExampleNewClient() {
	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
		APIKey:     os.Getenv("VOICEML_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	call, err := c.Calls.Create(context.Background(), voiceml.CreateCallParams{
		To:               "+14155550100",
		From:             "+14155550101",
		URL:              voiceml.String("https://example.com/twiml"),
		MachineDetection: voiceml.String("DetectMessageEnd"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(call.Sid, call.Status)
}

func ExampleCallsService_Iterate() {
	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
		APIKey:     os.Getenv("VOICEML_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	it := c.Calls.Iterate(context.Background(), voiceml.ListCallsParams{Status: "completed", PageSize: voiceml.Int(100)})
	for it.Next() {
		fmt.Println(it.Current().Sid)
	}
	if err := it.Err(); err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_NextPage() {
	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
		APIKey:     os.Getenv("VOICEML_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	page, err := c.Messages.List(ctx, voiceml.ListMessagesParams{})
	if err != nil {
		log.Fatal(err)
	}
	for {
		for _, m := range page.Messages {
			fmt.Println(m.Sid)
		}
		var next voiceml.MessageList
		more, err := c.NextPage(ctx, page, &next)
		if err != nil {
			log.Fatal(err)
		}
		if !more {
			break
		}
		page = &next
	}
}

func ExampleComputeSignature() {
	params := url.Values{"CallSid": {"CA0123456789abcdef0123456789abcdef"}, "CallStatus": {"completed"}}
	sig := voiceml.ComputeSignature("test-api-key-0123456789", "https://example.com/voice/status", params)
	fmt.Println(voiceml.ValidateRequest("test-api-key-0123456789", "https://example.com/voice/status", params, sig))
	fmt.Println(voiceml.ValidateRequest("test-api-key-0123456789", "https://example.com:443/voice/status", params, sig))
	fmt.Println(voiceml.ValidateRequest("wrong-key", "https://example.com/voice/status", params, sig))
	// Output:
	// true
	// true
	// false
}

func ExampleNewDiagnosticsService() {
	diag, err := voiceml.NewDiagnosticsService(voiceml.ClientOptions{})
	if err != nil {
		log.Fatal(err)
	}
	health, err := diag.Health(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(health.OK)
}

func ExampleAPIError() {
	c, err := voiceml.NewClient(voiceml.ClientOptions{
		AccountSid: os.Getenv("VOICEML_ACCOUNT_SID"),
		APIKey:     os.Getenv("VOICEML_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	_, err = c.Accounts.Update(context.Background(), voiceml.UpdateAccountParams{FriendlyName: voiceml.String("Support")})
	var apiErr *voiceml.APIError
	switch {
	case voiceml.IsConflict(err):
		// Concurrent update (code 20409): refetch and retry.
	case errors.As(err, &apiErr):
		fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message)
	}
}

func ExampleClientOptions_String() {
	opts := voiceml.ClientOptions{AccountSid: "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", APIKey: "not-printed"}
	fmt.Println(opts)
	// Output:
	// voiceml.ClientOptions{AccountSid: "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", APIKey: "[redacted]", AuthToken: "", BaseURL: "", MessagingBaseURL: "", ConversationsBaseURL: "", Timeout: 0s, UserAgent: ""}
}
