// Package queue implements the human approval queue for first-touch emails.
//
// The first ApprovalThreshold approved emails must be individually approved,
// rejected or edited by a human. Once a cumulative ApprovalThreshold emails
// have been approved the queue flips into monitoring mode: new emails are
// auto-approved (still logged, still governed) and the operator monitors
// rather than gates every send.
package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/NeuroLift-Technologies/homeservice-meeting-setter/internal/model"
)

// ApprovalThreshold is the cumulative number of approved first-touch emails
// after which human gating switches to monitoring mode.
const ApprovalThreshold = 50

// Queue is a file-backed approval queue. It is safe for concurrent use.
type Queue struct {
	mu      sync.Mutex
	path    string
	nextPos int
	approved int
	items   []model.ApprovalItem
}

type fileState struct {
	NextPos  int                    `json:"next_pos"`
	Approved int                    `json:"approved"`
	Items    []model.ApprovalItem   `json:"items"`
}

// Open loads (or initializes) a queue persisted at path.
func Open(path string) (*Queue, error) {
	q := &Queue{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := q.persist(); err != nil {
				return nil, err
			}
			return q, nil
		}
		return nil, err
	}
	var st fileState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("queue file %s corrupt: %w", path, err)
	}
	q.nextPos = st.NextPos
	q.approved = st.Approved
	q.items = st.Items
	return q, nil
}

// Enqueue inserts a generated email into the pending section of the queue.
func (q *Queue) Enqueue(e model.GeneratedEmail) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.nextPos++
	item := model.ApprovalItem{
		ID:       e.ID,
		Position: q.nextPos,
		Email:    e,
		Status:   model.ApprovalPending,
	}
	q.items = append(q.items, item)
	return q.persist()
}

// Approve marks a pending item approved unless auto-monitoring is active.
// It returns the running cumulative approved count.
func (q *Queue) Approve(id, approver string) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	item := q.find(id)
	if item == nil {
		return q.approved, fmt.Errorf("no queue item %q", id)
	}
	if item.Status == model.ApprovalApproved {
		return q.approved, fmt.Errorf("item %q already approved", id)
	}
	item.Status = model.ApprovalApproved
	item.ReviewedBy = approver
	now := time.Now().UTC()
	item.ReviewedAt = &now
	item.Email.Status = model.EmailApproved
	item.Email.ApprovedBy = approver
	item.Email.ApprovedAt = &now
	q.approved++
	return q.approved, q.persist()
}

// Reject marks a pending item rejected.
func (q *Queue) Reject(id, reason string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	item := q.find(id)
	if item == nil {
		return fmt.Errorf("no queue item %q", id)
	}
	if item.Status == model.ApprovalApproved {
		return fmt.Errorf("item %q already approved", id)
	}
	item.Status = model.ApprovalRejected
	item.ReviewedBy = "human"
	now := time.Now().UTC()
	item.ReviewedAt = &now
	item.RejectReason = reason
	item.Email.Status = model.EmailRejected
	return q.persist()
}

// Edit replaces the subject/body of a pending item's email and flags it as
// human-edited so it must pass re-review (and re-validation by the ASFDK
// guard at send time).
func (q *Queue) Edit(id, approver, subject, body string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	item := q.find(id)
	if item == nil {
		return fmt.Errorf("no queue item %q", id)
	}
	if item.Status != model.ApprovalPending && item.Status != model.ApprovalRejected {
		return fmt.Errorf("item %q is %s and cannot be edited; reject then regenerate if needed", id, item.Status)
	}
	if item.Email.OriginalBody == "" {
		item.Email.OriginalBody = item.Email.Body
	}
	item.Email.Subject = subject
	item.Email.Body = body
	item.Email.Edited = true
	item.ReviewedBy = approver
	if item.Status == model.ApprovalRejected {
		item.Status = model.ApprovalPending
		item.Email.Status = model.EmailPendingApproval
		item.RejectReason = ""
	}
	return q.persist()
}

// Pending returns the pending items in enqueue order.
func (q *Queue) Pending() []model.ApprovalItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]model.ApprovalItem, 0, len(q.items))
	for _, it := range q.items {
		if it.Status == model.ApprovalPending {
			out = append(out, it)
		}
	}
	return out
}

// ApprovedNotSent returns approved emails that have not been sent yet.
func (q *Queue) ApprovedNotSent() []model.GeneratedEmail {
	q.mu.Lock()
	defer q.mu.Unlock()
	var out []model.GeneratedEmail
	for _, it := range q.items {
		if it.Status == model.ApprovalApproved && it.Email.Status == model.EmailApproved {
			out = append(out, it.Email)
		}
	}
	return out
}

// MarkSent flips an email to sent and increments the send counter.
func (q *Queue) MarkSent(emailID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := range q.items {
		if q.items[i].Email.ID == emailID && q.items[i].Email.Status == model.EmailApproved {
			q.items[i].Email.Status = model.EmailSent
			q.items[i].Email.TimesSent++
			return q.persist()
		}
	}
	return fmt.Errorf("approved email %q not found for mark-sent", emailID)
}

// Items returns all queue items (for the status subcommand).
func (q *Queue) Items() []model.ApprovalItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]model.ApprovalItem, len(q.items))
	copy(out, q.items)
	return out
}

// ApprovedCount returns the cumulative count of human-approved emails.
func (q *Queue) ApprovedCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.approved
}

// IsMonitorMode reports whether the cumulative approval threshold has been
// reached, switching the system from gate-by-human to monitor-only.
func (q *Queue) IsMonitorMode() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.approved >= ApprovalThreshold
}

func (q *Queue) find(id string) *model.ApprovalItem {
	for i := range q.items {
		if q.items[i].ID == id {
			return &q.items[i]
		}
	}
	return nil
}

func (q *Queue) persist() error {
	if err := os.MkdirAll(filepath.Dir(q.path), 0o755); err != nil {
		return err
	}
	st := fileState{NextPos: q.nextPos, Approved: q.approved, Items: q.items}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := q.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, q.path)
}