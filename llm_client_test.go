package core

import (
	"context"
	"testing"
)

type fakeLlmClient struct {
	model string
	reply string
	calls int
	err   error
}

func (f *fakeLlmClient) Model() string { return f.model }

func (f *fakeLlmClient) Generate(context.Context, string) (string, error) {
	f.calls++
	return f.reply, f.err
}

func TestReplayAbleLlmClientPromptMatch(t *testing.T) {
	client := NewReplayAbleLlmClient(ReplayOptions{
		Calls: []ReplayCall{{Prompt: "prompt", Output: "stored"}},
	})
	output, err := client.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if output != "stored" {
		t.Fatalf("output = %q", output)
	}
	if client.Model() != "replay" {
		t.Fatalf("model = %q", client.Model())
	}
}

func TestReplayAbleLlmClientStageFallback(t *testing.T) {
	client := NewReplayAbleLlmClient(ReplayOptions{
		Calls:   []ReplayCall{{Stage: "status_extract", Output: "stored"}},
		StageOf: func(string) string { return "status_extract" },
	})
	output, err := client.Generate(context.Background(), "different prompt")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if output != "stored" {
		t.Fatalf("output = %q", output)
	}
}

func TestReplayAbleLlmClientLiveStage(t *testing.T) {
	live := &fakeLlmClient{model: "live", reply: "live output"}
	client := NewReplayAbleLlmClient(ReplayOptions{
		Calls:      []ReplayCall{{Stage: "status_extract", Output: "stored"}},
		Live:       live,
		LiveStages: map[string]bool{"status_extract": true},
		StageOf:    func(string) string { return "status_extract" },
	})
	output, err := client.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if output != "live output" || live.calls != 1 {
		t.Fatalf("output = %q calls = %d", output, live.calls)
	}
	if client.Model() != "live" {
		t.Fatalf("model = %q", client.Model())
	}
}

func TestReplayAbleLlmClientMissingLive(t *testing.T) {
	client := NewReplayAbleLlmClient(ReplayOptions{
		LiveStages: map[string]bool{"status_extract": true},
		StageOf:    func(string) string { return "status_extract" },
	})
	if _, err := client.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected live client error")
	}
}

func TestReplayAbleLlmClientMissingOutput(t *testing.T) {
	client := NewReplayAbleLlmClient(ReplayOptions{})
	_, err := client.Generate(context.Background(), "prompt")
	if err == nil {
		t.Fatal("expected missing output error")
	}
	if err.Error() != "replay raw output 이 없습니다" {
		t.Fatalf("err = %v", err)
	}
}
