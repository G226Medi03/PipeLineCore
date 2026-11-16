package core

import (
	"context"
	"errors"
)

// LlmClient is the minimal text generation client used for replay.
type LlmClient interface {
	Model() string
	Generate(ctx context.Context, prompt string) (string, error)
}

// ReplayCall is one stored LLM call used to replay a prior run.
type ReplayCall struct {
	Stage  string
	Prompt string
	Output string
}

// ReplayOptions configures a replay-capable LLM client.
type ReplayOptions struct {
	Calls      []ReplayCall
	Live       LlmClient
	LiveStages map[string]bool
	StageOf    func(prompt string) string
}

// ReplayAbleLlmClient replays stored calls and delegates live stages.
type ReplayAbleLlmClient struct {
	outputs    map[string]string
	byStage    map[string]string
	live       LlmClient
	liveStages map[string]bool
	stageOf    func(prompt string) string
	model      string
}

// NewReplayAbleLlmClient creates a replay-capable LLM client.
func NewReplayAbleLlmClient(opts ReplayOptions) *ReplayAbleLlmClient {
	outputs := make(map[string]string, len(opts.Calls))
	byStage := make(map[string]string, len(opts.Calls))
	for _, call := range opts.Calls {
		if call.Prompt != "" {
			outputs[call.Prompt] = call.Output
		}
		if call.Stage != "" {
			byStage[call.Stage] = call.Output
		}
	}
	model := "replay"
	if opts.Live != nil {
		model = opts.Live.Model()
	}
	return &ReplayAbleLlmClient{
		outputs:    outputs,
		byStage:    byStage,
		live:       opts.Live,
		liveStages: opts.LiveStages,
		stageOf:    opts.StageOf,
		model:      model,
	}
}

// Model returns the live model name, or "replay" when no live client exists.
func (c *ReplayAbleLlmClient) Model() string {
	if c == nil {
		return ""
	}
	return c.model
}

// Generate replays a stored output, or calls the live client for a live stage.
func (c *ReplayAbleLlmClient) Generate(ctx context.Context, prompt string) (string, error) {
	if c == nil {
		return "", errors.New("replay client is not initialized")
	}
	stage := ""
	if c.stageOf != nil {
		stage = c.stageOf(prompt)
	}
	if stage != "" && c.liveStages[stage] {
		if c.live == nil {
			return "", errors.New("replay live client is required")
		}
		return c.live.Generate(ctx, prompt)
	}
	if output, ok := c.outputs[prompt]; ok {
		return output, nil
	}
	if stage != "" {
		if output, ok := c.byStage[stage]; ok {
			return output, nil
		}
	}
	return "", errors.New("replay raw output 이 없습니다")
}
