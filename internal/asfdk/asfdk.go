// Package asfdk wraps the NeuroLift ASFDK Go library so the agent has a
// single, auditable point of contact with the governance/safety layer.
//
// Every generated email passes through SanitizeInput (prompt defense) and
// ValidateOutput (output validation) before it is allowed into the approval
// queue, and everything the guard blocks or flags is appended to the ASFDK
// security event log. Actions that map onto foundation interaction types
// (status inquiries, optimization requests) are routed through
// NeuroLiftFoundation.ProcessInteraction with explicit channel provenance,
// which fails closed on unknown channels.
package asfdk

import (
	"fmt"
	"time"

	asf "github.com/NeuroLift-Technologies/asfdk-go"
)

// Guard is the agent's single seam into ASFDK.
type Guard struct {
	Foundation      *asf.NeuroLiftFoundation
	SecurityLogPath string
	UserID          string
}

// UserID is the identity under which the agent governs its own side of the
// session (the orchestrator/human principal, not the prospect).
const UserID = "agent:homeservice-meeting-setter"

// NewGuard constructs a unified-mode foundation (TOI/OTOI + Sleepwalker +
// RRT Advocate) and wires the security event log. userId names the operator
// that owns the guard; pass UserID for the agent's own governance actions.
func NewGuard(userID, securityLogPath string) (*Guard, error) {
	f, err := asf.NewNeuroLiftFoundation(asf.FoundationConfig{
		UserId: userID,
		Mode:   asf.ModeUnified,
	})
	if err != nil {
		return nil, fmt.Errorf("asfdk foundation: %w", err)
	}
	if securityLogPath == "" {
		securityLogPath = "data/security.jsonl"
	}
	return &Guard{Foundation: f, SecurityLogPath: securityLogPath, UserID: userID}, nil
}

// GuardResult is the outcome of guarding a piece of text.
type GuardResult struct {
	Clean  bool
	Risk   asf.RiskLevel
	Reason string
}

// SanitizeEmailInput runs one field of an email (subject or body) through
// asfdk.SanitizeInput. A non-clean result is logged as a security event and
// the caller must not use the content.
func (g *Guard) SanitizeEmailInput(field, input string) GuardResult {
	label := fmt.Sprintf("field=%q", field)
	defer func() {
		// Guard failures are always observable.
	}()
	res := asf.SanitizeInput(input, 0) // 0 => asfdk default max length
	if res.Clean {
		return GuardResult{Clean: true, Risk: res.RiskLevel}
	}
	res.Reason = fmt.Sprintf("%s: %s", label, res.Reason)
	g.LogBlock(res.Reason, res.RiskLevel)
	return GuardResult{Clean: false, Risk: res.RiskLevel, Reason: res.Reason}
}

// ValidateEmailOutput runs generated email text through asfdk.ValidateOutput
// with the text schema (non-empty, printable output). Failed validation is
// logged as a security event.
func (g *Guard) ValidateEmailOutput(field, output string) (bool, string) {
	res := asf.ValidateOutput(output, asf.OutputSchemaText)
	if !res.Valid {
		reason := fmt.Sprintf("field=%q: %s", field, res.Reason)
		_ = g.LogEvent(asf.SecurityValidationFailure, reason)
		return false, reason
	}
	return true, ""
}

// LogBlock records a sanitization block/flag in the security event log.
func (g *Guard) LogBlock(reason string, risk asf.RiskLevel) {
	eventType := asf.SecurityValidationFailure
	switch risk {
	case asf.RiskLevelHigh:
		eventType = asf.SecurityInjectionAttempt
	case asf.RiskLevelMedium:
		eventType = asf.SecurityLengthExceeded
	}
	_ = g.LogEvent(eventType, reason)
}

// LogEvent appends a timestamped security event to the log file.
func (g *Guard) LogEvent(eventType asf.SecurityEventType, details string) error {
	return asf.StoreSecurityEvent(g.SecurityLogPath, asf.NewSecurityEvent(eventType, g.UserID, details))
}

// ProcessInteraction routes an interaction through the foundation, always
// with explicit channel provenance. Unknown channels are rejected by the
// foundation (fail closed).
func (g *Guard) ProcessInteraction(interactionType asf.InteractionType, data map[string]any, channel asf.Channel, sessionID string) (asf.FoundationResponse, error) {
	interaction := asf.UserInteraction{
		UserId:          g.UserID,
		SessionId:       sessionID,
		InteractionType: interactionType,
		Data:            data,
		Channel:         channel,
		Timestamp:       time.Now(),
	}
	return g.Foundation.ProcessInteraction(interaction)
}

// StatusInquiry is a convenience for logging a governed status heartbeat
// through the foundation with explicit channel provenance.
func (g *Guard) StatusInquiry(sessionID, detail string) error {
	resp, err := g.ProcessInteraction(asf.InteractionStatusInquiry,
		map[string]any{"text": detail}, asf.ChannelSystem, sessionID)
	if err != nil {
		return fmt.Errorf("foundation status_inquiry: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("foundation status_inquiry failed: %v", resp.Content)
	}
	return nil
}

// Health returns the foundation health check.
func (g *Guard) Health() asf.HealthCheckResult {
	return g.Foundation.HealthCheck()
}

// ValidateTOI revalidates the foundation's resolved TOI document.
func (g *Guard) ValidateTOI() asf.TOIValidationResult {
	return g.Foundation.ValidateTOIDocument()
}