// Package sender delivers approved emails. In v1 the MockSender is used for
// local demos and tests; SMTPSender provides a real net/smtp transport. Every
// delivery attempt is recorded in the sent log.
package sender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
)

// Sender is the transport seam for email delivery.
type Sender interface {
	// Name is the provider identifier recorded on SendRecord (e.g. "mock").
	Name() string
	// Send delivers one email and returns the resulting record. A returned
	// record with Status == "error" counts as a failed attempt.
	Send(ctx context.Context, e model.GeneratedEmail, to string) (model.SendRecord, error)
}

// Log is a concurrency-safe, file-backed journal of send records.
type Log struct {
	mu   sync.Mutex
	path string
	recs []model.SendRecord
}

// OpenLog loads (or initializes) the send log at path.
func OpenLog(path string) (*Log, error) {
	l := &Log{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return l, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &l.recs); err != nil {
		return nil, fmt.Errorf("send log %s corrupt: %w", path, err)
	}
	return l, nil
}

// Record appends a delivery attempt and persists the log.
func (l *Log) Record(r model.SendRecord) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.recs = append(l.recs, r)
	data, err := json.MarshalIndent(l.recs, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(l.path, data, 0o644)
}

// Records returns all recorded attempts.
func (l *Log) Records() []model.SendRecord {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]model.SendRecord, len(l.recs))
	copy(out, l.recs)
	return out
}

// MockSender simulates a provider. Delivery is instantaneous and always
// succeeds unless SimulateFailure is set (used to exercise the error path).
type MockSender struct {
	SimulateFailure bool
}

// Name implements Sender.
func (m MockSender) Name() string { return "mock" }

// Send implements Sender.
func (m MockSender) Send(_ context.Context, e model.GeneratedEmail, to string) (model.SendRecord, error) {
	if m.SimulateFailure {
		return model.SendRecord{
			ID: fmt.Sprintf("mock_%d", time.Now().UnixNano()), EmailID: e.ID, ProspectID: e.ProspectID,
			To: to, Subject: e.Subject, Provider: m.Name(), Status: "error",
			Error: "simulated transport failure", SentAt: time.Now().UTC(),
		}, nil
	}
	return model.SendRecord{
		ID: fmt.Sprintf("mock_%d", time.Now().UnixNano()), EmailID: e.ID, ProspectID: e.ProspectID,
		To: to, Subject: e.Subject, Provider: m.Name(), Status: "sent",
		MessageID: fmt.Sprintf("<%d.mock@example>", time.Now().UnixNano()), SentAt: time.Now().UTC(),
	}, nil
}

// SMTPSender sends through a net/smtp server using STARTTLS auth.
type SMTPSender struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Name implements Sender.
func (s SMTPSender) Name() string { return "smtp" }

// Send implements Sender. The message uses no external templates: subject and
// body come straight from the approved email.
func (s SMTPSender) Send(_ context.Context, e model.GeneratedEmail, to string) (model.SendRecord, error) {
	addr := s.Host + ":" + s.Port
	msg := "From: " + s.From + "\n" +
		"To: " + to + "\n" +
		"Subject: " + e.Subject + "\n" +
		"MIME-Version: 1.0\n" +
		"Content-Type: text/plain; charset=UTF-8\n\n" +
		e.Body

	var auth smtp.Auth
	if s.Username != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}
	if err := smtp.SendMail(addr, auth, s.From, []string{to}, []byte(msg)); err != nil {
		return model.SendRecord{
			ID: fmt.Sprintf("smtp_%d", time.Now().UnixNano()), EmailID: e.ID, ProspectID: e.ProspectID,
			To: to, Subject: e.Subject, Provider: s.Name(), Status: "error",
			Error: err.Error(), SentAt: time.Now().UTC(),
		}, err
	}
	return model.SendRecord{
		ID: fmt.Sprintf("smtp_%d", time.Now().UnixNano()), EmailID: e.ID, ProspectID: e.ProspectID,
		To: to, Subject: e.Subject, Provider: s.Name(), Status: "sent",
		MessageID: fmt.Sprintf("<%d@smtp>", time.Now().UnixNano()), SentAt: time.Now().UTC(),
	}, nil
}

// ErrNoTransport is returned when a required transport parameter is missing.
var ErrNoTransport = errors.New("smtp: missing host/port/from for real send")