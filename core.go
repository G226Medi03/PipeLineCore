// Package core defines the application pipeline and node model.
package core

import (
	"context"
	"fmt"
	"time"
)

type PipelineName string

type NodeName string

type NodeKind string

const (
	KindPrepare   NodeKind = "prepare"
	KindLLM       NodeKind = "llm"
	KindTransform NodeKind = "transform"
	KindResolve   NodeKind = "resolve"
	KindPersist   NodeKind = "persist"
	KindGraph     NodeKind = "graph"
	KindCheck     NodeKind = "check"
	KindFinalize  NodeKind = "finalize"
)

type NodeState string

const (
	NodePending   NodeState = "pending"
	NodeRunning   NodeState = "running"
	NodeSucceeded NodeState = "succeeded"
	NodeFailed    NodeState = "failed"
	NodeSkipped   NodeState = "skipped"
)

// Event is emitted whenever a node changes state.
type Event struct {
	Node      NodeName
	State     NodeState
	Error     error
	StartedAt time.Time
	EndedAt   time.Time
}

// Node is one executable stage in a Pipeline.
type Node struct {
	Name     NodeName
	Stage    string
	Kind     NodeKind
	Run      func(context.Context, *State) error
	Optional bool
}

// Pipeline is an ordered set of nodes. Handlers invoke a Pipeline for their
// complete request process.
type Pipeline struct {
	Name  PipelineName
	Nodes []Node
}

// State is the shared state passed between nodes.
type State struct {
	Values map[string]any
	Events []Event
}

// NewState creates an empty pipeline state.
func NewState() *State {
	return &State{Values: map[string]any{}}
}

// Set stores a value in the shared state.
func (s *State) Set(key string, value any) {
	if s.Values == nil {
		s.Values = map[string]any{}
	}
	s.Values[key] = value
}

// Get reads a value from the shared state.
func (s *State) Get(key string) (any, bool) {
	if s.Values == nil {
		return nil, false
	}
	value, ok := s.Values[key]
	return value, ok
}

// Validate reports structural pipeline errors before execution.
func (p Pipeline) Validate() error {
	seen := make(map[NodeName]struct{}, len(p.Nodes))
	for _, node := range p.Nodes {
		if node.Name == "" {
			return fmt.Errorf("pipeline %q has a node with an empty name", p.Name)
		}
		if _, exists := seen[node.Name]; exists {
			return fmt.Errorf("pipeline %q has duplicate node %q", p.Name, node.Name)
		}
		seen[node.Name] = struct{}{}
		if node.Run == nil {
			return fmt.Errorf("pipeline %q node %q has no run function", p.Name, node.Name)
		}
	}
	return nil
}

// Run executes every node in order and stops at the first non-optional error.
func (p Pipeline) Run(ctx context.Context, state *State) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if state == nil {
		state = NewState()
	}
	if err := p.Validate(); err != nil {
		return err
	}
	for _, node := range p.Nodes {
		if err := node.run(ctx, state); err != nil {
			if node.Optional {
				continue
			}
			return err
		}
	}
	return nil
}

// RunFrom executes nodes starting at the named node.
func (p Pipeline) RunFrom(ctx context.Context, state *State, from NodeName) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if state == nil {
		state = NewState()
	}
	start := -1
	for index, node := range p.Nodes {
		if node.Name == from {
			start = index
			break
		}
	}
	if start < 0 {
		return fmt.Errorf("pipeline %q has no node %q", p.Name, from)
	}
	return Pipeline{Name: p.Name, Nodes: p.Nodes[start:]}.Run(ctx, state)
}

func (n Node) run(ctx context.Context, state *State) error {
	started := time.Now()
	state.Events = append(state.Events, Event{
		Node: n.Name, State: NodeRunning, StartedAt: started,
	})
	if err := n.Run(ctx, state); err != nil {
		state.Events = append(state.Events, Event{
			Node: n.Name, State: NodeFailed, Error: err,
			StartedAt: started, EndedAt: time.Now(),
		})
		return fmt.Errorf("%s: %w", n.Name, err)
	}
	state.Events = append(state.Events, Event{
		Node: n.Name, State: NodeSucceeded,
		StartedAt: started, EndedAt: time.Now(),
	})
	return nil
}
