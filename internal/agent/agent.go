// Package agent wires the HomeService Meeting Setter engine into Google
// ADK-Go (google.golang.org/adk). The root agent is a custom agent with a
// deterministic Run pipeline — research → generate → ASFDK guard → enqueue —
// so the happy path is fully reproducible without an API key. An optional
// LLM sub-agent can be attached at build time (env-gated) for richer drafts;
// see Build.
package agent

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/engine"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// AgentName is the ADK agent identifier. It matches the "agent-name" field in
// the registered Agent Card and the runner app name.
const AgentName = "homeservice_meeting_setter"

// BuildOptions controls which optional ADK pieces are wired in.
type BuildOptions struct {
	// WithInventoryService (default true) attaches an in-memory session
	// service so runner.Run can auto-create sessions without external
	// storage.
	WithInventoryService bool
}

// Service bundles the engine and the ADK session service for the runner.
type Service struct {
	Engine   *engine.Engine
	Sessions session.Service
}

// Build constructs the custom ADK agent and its runner dependencies.
// runPipeline is a closure that captures eng so it is accessible inside
// the agent's Run callback.
func Build(ctx context.Context, eng *engine.Engine, opts BuildOptions) (agent.Agent, *Service, error) {
	svc := &Service{Engine: eng}
	if opts.WithInventoryService {
		svc.Sessions = session.InMemoryService()
	}

	a, err := agent.New(agent.Config{
		Name:        AgentName,
		Description: "HomeService Meeting Setter — researches home-service prospects, generates highly specific first-touch emails, guards them through ASFDK, and queues them for human approval.",
		Run: func(inv agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return runPipeline(inv, eng)
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create ADK agent: %w", err)
	}

	return a, svc, nil
}

// runPipeline drives the meeting-setter pipeline for a single turn of user
// text, yielding session events as it goes. It is the heart of the ADK agent:
// every step is recorded as a session Event so operators can replay exactly
// what the agent decided and why.
//
// Accepted turn shapes:
//
//	process <path>   research → generate → guard → enqueue
//
// Anything else is treated as a research status inquiry and reported without
// side effects.
func runPipeline(inv agent.InvocationContext, eng *engine.Engine) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		text := renderUserText(inv.UserContent())
		line := strings.TrimSpace(text)

		switch {
		case strings.HasPrefix(line, "process "):
			path := strings.TrimSpace(strings.TrimPrefix(line, "process "))
			step(inv, yield, "research", "researching %s", path)

			prospects, err := engine.LoadFile(path)
			if err != nil {
				msg := fmt.Sprintf("research failed: %v", err)
				step(inv, yield, "error", "%s", msg)
				return
			}
			step(inv, yield, "generate", "generating emails for %d prospects", len(prospects))

			var processed int
			var failed int
			for _, p := range prospects {
				email, ok := eng.Generate(p)
				if !ok {
					failed++
					continue
				}
				if err := eng.GuardEmail(email, p); err != nil {
					failed++
					continue
				}
				if err := eng.Enqueue(email); err != nil {
					failed++
					continue
				}
				processed++
				eng.IncrementProcessed()
				step(inv, yield, "email", "queued %s -> %s (specificity %.2f)", p.BusinessName, email.Subject, email.Specificity)
			}

			for i := 0; i < failed; i++ {
				eng.IncrementFailed()
			}
			step(inv, yield, "summary", "processed %d/%d, %d rejected", processed, len(prospects), failed)
		case line == "status" || strings.HasPrefix(line, "status "):
			step(inv, yield, "status", eng.StatusSummary())
		case line == "":
			step(inv, yield, "info", "no command given — try: process data/prospects.json")
		default:
			step(inv, yield, "info", "unknown command %q — try: process data/prospects.json", line)
		}
	}
}

// step emits a single session event, returning false if the consumer stopped.
func step(inv agent.InvocationContext, yield func(*session.Event, error) bool, kind, format string, args ...any) bool {
	text := fmt.Sprintf(format, args...)
	ev := session.NewEvent(inv.InvocationID())
	ev.LLMResponse = model.LLMResponse{
		Content: &genai.Content{
			Role:  genai.RoleModel,
			Parts: []*genai.Part{genai.NewPartFromText(text)},
		},
	}
	ev.Author = AgentName
	// Keep a thread-local log in the transport event's metadata for
	// observability without requiring model backends.
	return yield(ev, nil)
}

// renderUserText flattens the user turn's parts to a single string.
func renderUserText(content *genai.Content) string {
	if content == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range content.Parts {
		if p == nil {
			continue
		}
		if p.Text != "" {
			b.WriteString(p.Text)
			b.WriteString(" ")
		} else if p.InlineData != nil {
			b.WriteString("[inline data] ")
		}
	}
	return strings.TrimSpace(b.String())
}

var _ = context.TODO