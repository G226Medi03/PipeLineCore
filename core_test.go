package core

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestPipelineRunOrderAndState(t *testing.T) {
	var order []string
	pipeline := Pipeline{
		Name: "test",
		Nodes: []Node{
			{
				Name: "first",
				Run: func(ctx context.Context, state *State) error {
					order = append(order, "first")
					state.Set("value", 41)
					return nil
				},
			},
			{
				Name: "second",
				Run: func(ctx context.Context, state *State) error {
					order = append(order, "second")
					value, ok := state.Get("value")
					if !ok {
						t.Fatal("value missing from state")
					}
					state.Set("value", value.(int)+1)
					return nil
				},
			},
		},
	}

	state := NewState()
	if err := pipeline.Run(context.Background(), state); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got, want := strings.Join(order, ","), "first,second"; got != want {
		t.Fatalf("order = %q, want %q", got, want)
	}
	value, ok := state.Get("value")
	if !ok || value.(int) != 42 {
		t.Fatalf("value = %v, ok=%v, want 42", value, ok)
	}
	if len(state.Events) != 4 {
		t.Fatalf("events = %d, want 4", len(state.Events))
	}
	if state.Events[0].State != NodeRunning || state.Events[1].State != NodeSucceeded {
		t.Fatalf("unexpected events: %#v", state.Events)
	}
}

func TestPipelineStopsOnRequiredError(t *testing.T) {
	wantErr := errors.New("boom")
	ran := false
	pipeline := Pipeline{
		Name: "test",
		Nodes: []Node{
			{Name: "fail", Run: func(context.Context, *State) error { return wantErr }},
			{Name: "never", Run: func(context.Context, *State) error { ran = true; return nil }},
		},
	}
	err := pipeline.Run(context.Background(), NewState())
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if ran {
		t.Fatal("node after required failure must not run")
	}
}

func TestPipelineContinuesOptionalError(t *testing.T) {
	ran := false
	pipeline := Pipeline{
		Name: "test",
		Nodes: []Node{
			{Name: "optional", Optional: true, Run: func(context.Context, *State) error { return errors.New("ignored") }},
			{Name: "next", Run: func(context.Context, *State) error { ran = true; return nil }},
		},
	}
	if err := pipeline.Run(context.Background(), NewState()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !ran {
		t.Fatal("node after optional failure must run")
	}
}

func TestPipelineRunFrom(t *testing.T) {
	var order []string
	pipeline := Pipeline{
		Name: "test",
		Nodes: []Node{
			{Name: "one", Run: func(context.Context, *State) error { order = append(order, "one"); return nil }},
			{Name: "two", Run: func(context.Context, *State) error { order = append(order, "two"); return nil }},
			{Name: "three", Run: func(context.Context, *State) error { order = append(order, "three"); return nil }},
		},
	}
	if err := pipeline.RunFrom(context.Background(), NewState(), "two"); err != nil {
		t.Fatalf("RunFrom: %v", err)
	}
	if got, want := strings.Join(order, ","), "two,three"; got != want {
		t.Fatalf("order = %q, want %q", got, want)
	}
}

func TestPipelineRunFromUnknown(t *testing.T) {
	pipeline := Pipeline{Name: "test"}
	if err := pipeline.RunFrom(context.Background(), NewState(), "missing"); err == nil {
		t.Fatal("RunFrom must fail for unknown node")
	}
}

func TestPipelineValidate(t *testing.T) {
	node := func(name NodeName) Node {
		return Node{Name: name, Run: func(context.Context, *State) error { return nil }}
	}
	tests := []struct {
		name     string
		pipeline Pipeline
	}{
		{name: "empty node name", pipeline: Pipeline{Nodes: []Node{node("")}}},
		{name: "duplicate node", pipeline: Pipeline{Nodes: []Node{node("same"), node("same")}}},
		{name: "nil run", pipeline: Pipeline{Nodes: []Node{{Name: "broken"}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.pipeline.Validate(); err == nil {
				t.Fatal("Validate must fail")
			}
		})
	}
	if err := (Pipeline{Nodes: []Node{node("one")}}).Validate(); err != nil {
		t.Fatalf("valid pipeline: %v", err)
	}
}
