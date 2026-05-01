package openapisearch

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestExtractAndDecodeJSONBlock(t *testing.T) {
	tests := []string{
		`{"ok":true}`,
		"```json\n{\"ok\":true}\n```",
		"prefix {\"ok\":true} suffix",
	}
	for _, tt := range tests {
		var out struct {
			OK bool `json:"ok"`
		}
		if err := DecodeJSONBlock(tt, &out); err != nil {
			t.Fatalf("DecodeJSONBlock(%q): %v", tt, err)
		}
		if !out.OK {
			t.Fatalf("decoded = %#v", out)
		}
	}
	if _, err := ExtractJSONBlock("no json"); err == nil {
		t.Fatal("expected missing JSON error")
	}
}

func TestAppendLegacyJSONInstructionAndSnapshot(t *testing.T) {
	turns := []TranscriptTurn{{Role: "system", Content: "sys"}, {Role: "user", Content: "reply"}}
	got := AppendLegacyJSONInstruction(turns, "")
	if !strings.Contains(got[1].Content, "Return only JSON") {
		t.Fatalf("turns = %#v", got)
	}
	again := AppendLegacyJSONInstruction(got, "")
	if strings.Count(again[1].Content, "Return only JSON") != 1 {
		t.Fatalf("instruction duplicated: %#v", again)
	}
	if snapshot := RenderTranscriptSnapshot(again); !strings.Contains(snapshot, "## SYSTEM") || !strings.Contains(snapshot, "## USER") {
		t.Fatalf("snapshot = %q", snapshot)
	}
}

func TestCompleteJSONWithFallback(t *testing.T) {
	client := &testChatClient{structuredErr: errors.New("structured unavailable"), chatResponse: "```json\n{\"ok\":true}\n```"}
	var out struct {
		OK bool `json:"ok"`
	}
	result, err := CompleteJSONWithFallback(context.Background(), client, []TranscriptTurn{{Role: "user", Content: "reply"}}, json.RawMessage(`{"type":"object"}`), &out, JSONCompletionOptions{FallbackOnStructuredError: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Mode != JSONCompletionModeLegacy || !out.OK || !client.chatCalled || !client.structuredCalled {
		t.Fatalf("result=%#v out=%#v client=%#v", result, out, client)
	}
	if !strings.Contains(client.chatTranscript[0].Content, "Return only JSON") {
		t.Fatalf("fallback transcript = %#v", client.chatTranscript)
	}

	client = &testChatClient{structuredResponse: `{"ok":true}`}
	out.OK = false
	result, err = CompleteJSONWithFallback(context.Background(), client, []TranscriptTurn{{Role: "user", Content: "reply"}}, json.RawMessage(`{"type":"object"}`), &out, JSONCompletionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Mode != JSONCompletionModeStructured || !out.OK || client.chatCalled {
		t.Fatalf("structured result=%#v out=%#v client=%#v", result, out, client)
	}
}

func TestRawSchema(t *testing.T) {
	raw, err := RawSchema(map[string]string{"type": "object"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "object") {
		t.Fatalf("raw = %s", raw)
	}
}

type testChatClient struct {
	chatResponse       string
	structuredResponse string
	structuredErr      error
	chatCalled         bool
	structuredCalled   bool
	chatTranscript     []TranscriptTurn
}

func (client *testChatClient) Complete(_ context.Context, transcript []TranscriptTurn) (TranscriptTurn, error) {
	client.chatCalled = true
	client.chatTranscript = append([]TranscriptTurn(nil), transcript...)
	return TranscriptTurn{Role: "assistant", Content: client.chatResponse}, nil
}

func (client *testChatClient) CompleteStructured(_ context.Context, _ []TranscriptTurn, _ any, out any) error {
	client.structuredCalled = true
	if client.structuredErr != nil {
		return client.structuredErr
	}
	return json.Unmarshal([]byte(client.structuredResponse), out)
}
