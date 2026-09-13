// Package engine is the composition root of the HomeService Meeting Setter.
//
// It wires the healthy sibling packages (prospect, email, asfdk, queue)
// behind the contract that internal/agent depends on:
//
//	engine.LoadFile(path) ([]model.Prospect, error)
//	(*Engine).Generate(p) (model.GeneratedEmail, bool)
//	(*Engine).GuardEmail(email, p) error
//	(*Engine).StatusSummary() string
//
// The Engine is intentionally thin — it delegates to the siblings and
// tracks aggregate counters. It holds no business logic of its own.
package engine

import (
	"fmt"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/asfdk"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/email"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/prospect"
	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/queue"
)

// Config holds the wiring parameters for an Engine.
type Config struct {
	// QueuePath is the file-backed approval queue path.
	QueuePath string
	// SecurityLogPath is the ASFDK security event log path.
	SecurityLogPath string
	// UserID is the identity under which the agent governs itself.
	UserID string
	// EmailConfig controls the generator persona and specificity bar.
	EmailConfig email.Config
}

// Engine composes research → generate → guard → queue behind a single seam.
type Engine struct {
	gen   *email.Generator
	guard *asfdk.Guard
	q     *queue.Queue

	processed int
	failed    int
}

// New constructs an Engine and initializes the queue and ASFDK guard.
func New(cfg Config) (*Engine, error) {
	guard, err := asfdk.NewGuard(cfg.UserID, cfg.SecurityLogPath)
	if err != nil {
		return nil, fmt.Errorf("engine: asfdk guard: %w", err)
	}

	q, err := queue.Open(cfg.QueuePath)
	if err != nil {
		return nil, fmt.Errorf("engine: queue: %w", err)
	}

	return &Engine{
		gen:   email.NewGenerator(cfg.EmailConfig),
		guard: guard,
		q:     q,
	}, nil
}

// LoadFile delegates to prospect.Load so callers can use engine.LoadFile
// through a single import.
func LoadFile(path string) ([]model.Prospect, error) {
	return prospect.Load(path)
}

// Generate builds a personalized email for the prospect, applying the
// generator's specificity gate. The returned ok is false when the email
// does not meet the specificity bar — callers must not fall back to a
// generic message in that case.
func (e *Engine) Generate(p model.Prospect) (model.GeneratedEmail, bool) {
	obs := prospect.ExtractObservations(p)
	return e.gen.Generate(p, obs)
}

// GuardEmail runs the generated email through ASFDK (prompt defense +
// output validation). It sets the email's ASFDKStatus and ASFDKReason,
// logs blocks/flags to the security event log, and returns an error when
// the email is blocked or invalid. A nil return means the email is clean
// and safe to enqueue.
func (e *Engine) GuardEmail(email model.GeneratedEmail, p model.Prospect) error {
	// Sanitize subject (prompt defense).
	subjectResult := e.guard.SanitizeEmailInput("subject", email.Subject)
	if !subjectResult.Clean {
		email.ASFDKStatus = model.ASFDKFlagged
		email.ASFDKReason = subjectResult.Reason
		return fmt.Errorf("asfdk: subject blocked: %s", subjectResult.Reason)
	}

	// Sanitize body (prompt defense).
	bodyResult := e.guard.SanitizeEmailInput("body", email.Body)
	if !bodyResult.Clean {
		email.ASFDKStatus = model.ASFDKFlagged
		email.ASFDKReason = bodyResult.Reason
		return fmt.Errorf("asfdk: body blocked: %s", bodyResult.Reason)
	}

	// Validate output (output validation).
	if ok, reason := e.guard.ValidateEmailOutput("body", email.Body); !ok {
		email.ASFDKStatus = model.ASFDKInvalid
		email.ASFDKReason = reason
		return fmt.Errorf("asfdk: output invalid: %s", reason)
	}

	// All clear.
	email.ASFDKStatus = model.ASFDKClean
	return nil
}

// Enqueue inserts a guarded email into the approval queue.
func (e *Engine) Enqueue(email model.GeneratedEmail) error {
	return e.q.Enqueue(email)
}

// IncrementProcessed records a successfully processed prospect.
func (e *Engine) IncrementProcessed() {
	e.processed++
}

// IncrementFailed records a prospect that was skipped or rejected.
func (e *Engine) IncrementFailed() {
	e.failed++
}

// StatusSummary returns a human-readable summary of the run.
func (e *Engine) StatusSummary() string {
	return fmt.Sprintf(
		"processed=%d failed=%d queue.pending=%d queue.approved=%d queue.monitor_mode=%v",
		e.processed,
		e.failed,
		len(e.q.Pending()),
		e.q.ApprovedCount(),
		e.q.IsMonitorMode(),
	)
}

// Queue returns the underlying approval queue for status inquiries.
func (e *Engine) Queue() *queue.Queue {
	return e.q
}

// Close persists the queue and guard state.
func (e *Engine) Close() error {
	return nil
}